package redisDB

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
)

// tokenIndexTTL is the tombstone lifetime for a rotated refresh token hash in the token index.
// Keeps reuse detection active for 1 h after rotation.
// tokenIndexTombstoneTTL TTL для
const tokenIndexTombstoneTTL = time.Hour

var errSessionNotFound = errors.New("session not found")

type AuthRepository interface {
	SaveSession(ctx context.Context, dto entities.SaveSessionDTO) Error.CodeError
	GetAllSessions(ctx context.Context, dto entities.GetAllSessionsDTO) ([]entities.SessionEntry, Error.CodeError)
	CheckSessionExists(ctx context.Context, dto entities.CheckSessionExistsDTO) Error.CodeError
	RevokeSession(ctx context.Context, dto entities.RevokeSessionDTO) Error.CodeError
	RevokeAllSessions(ctx context.Context, dto entities.RevokeAllSessionsDTO) Error.CodeError
	RefreshToken(ctx context.Context, dto entities.RefreshTokenDTO) Error.CodeError
}

type authRepository struct {
	redis           *redis.Client
	refreshTokenTTL time.Duration
	prefix          string
}

func NewAuthRepository(redis *redis.Client, refreshTokenTTL time.Duration, prefix string) AuthRepository {
	return &authRepository{redis: redis, refreshTokenTTL: refreshTokenTTL, prefix: prefix}
}

// SaveSession сохраняет новую сессию, индексированную по sessionUUID.
// Поле current_hash хранит хеш текущего refresh токена для детектирования повторного использования.
func (r *authRepository) SaveSession(ctx context.Context, dto entities.SaveSessionDTO) Error.CodeError {
	sessionKey := r.getSessionKey(dto.SessionUUID)
	tokenIndexKey := r.getTokenIndexKey(dto.HashedToken)
	userSessionsKey := r.getUserSessionsKey(dto.UserUUID)

	fields := dto.Session.ToMap()
	fields["current_hash"] = dto.HashedToken

	pipeline := r.redis.Pipeline()
	pipeline.HSet(ctx, sessionKey, fields)
	pipeline.Expire(ctx, sessionKey, r.refreshTokenTTL)
	pipeline.Set(ctx, tokenIndexKey, dto.SessionUUID, r.refreshTokenTTL)
	pipeline.SAdd(ctx, userSessionsKey, dto.SessionUUID)

	_, err := pipeline.Exec(ctx)
	if err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// GetAllSessions возвращает все активные сессии пользователя
func (r *authRepository) GetAllSessions(ctx context.Context, dto entities.GetAllSessionsDTO) ([]entities.SessionEntry, Error.CodeError) {
	userSessionsKey := r.getUserSessionsKey(dto.UserUUID)

	sessionUUIDs, err := r.redis.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, Error.Internal(err)
	}

	if len(sessionUUIDs) == 0 {
		return []entities.SessionEntry{}, Error.CodeError{}
	}

	pipe := r.redis.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(sessionUUIDs))
	for i, sid := range sessionUUIDs {
		cmds[i] = pipe.HGetAll(ctx, r.getSessionKey(sid))
	}
	if _, err = pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, Error.Internal(err)
	}

	activeEntries := make([]entities.SessionEntry, 0, len(sessionUUIDs))
	expiredUUIDs := make([]interface{}, 0)

	for i, sid := range sessionUUIDs {
		fields, cmdErr := cmds[i].Result()
		if cmdErr != nil || len(fields) == 0 {
			expiredUUIDs = append(expiredUUIDs, sid)
			continue
		}

		session := &entities.SessionInfo{}
		session.FromMap(fields)

		activeEntries = append(activeEntries, entities.SessionEntry{
			SessionUUID: sid,
			Session:     session,
		})
	}

	if len(expiredUUIDs) > 0 {
		_ = r.redis.SRem(ctx, userSessionsKey, expiredUUIDs...).Err()
	}

	return activeEntries, Error.CodeError{}
}

// CheckSessionExists проверяет наличие сессии
func (r *authRepository) CheckSessionExists(ctx context.Context, dto entities.CheckSessionExistsDTO) Error.CodeError {
	tokenIndexKey := r.getTokenIndexKey(dto.HashedToken)

	sessionUUID, err := r.redis.Get(ctx, tokenIndexKey).Result()
	if errors.Is(err, redis.Nil) || sessionUUID == "" {
		return Error.Public(codes.NotFound, "session not found")
	}
	if err != nil {
		return Error.Internal(err)
	}

	// Verify the session HSET still exists (could have been revoked concurrently)
	exists, err := r.redis.Exists(ctx, r.getSessionKey(sessionUUID)).Result()
	if err != nil {
		return Error.Internal(err)
	}
	if exists == 0 {
		_ = r.redis.Del(ctx, tokenIndexKey).Err()
		return Error.Public(codes.NotFound, "session not found")
	}

	return Error.CodeError{}
}

// RevokeSession отзывает конкретную сессию по SessionUUID, проверяя принадлежность пользователю
func (r *authRepository) RevokeSession(ctx context.Context, dto entities.RevokeSessionDTO) Error.CodeError {
	userSessionsKey := r.getUserSessionsKey(dto.UserUUID)
	sessionKey := r.getSessionKey(dto.SessionUUID)

	isMember, err := r.redis.SIsMember(ctx, userSessionsKey, dto.SessionUUID).Result()
	if err != nil {
		return Error.Internal(err)
	}
	if !isMember {
		return Error.Public(codes.NotFound, "session not found")
	}

	// Fetch current_hash to clean up the token index
	currentHash, err := r.redis.HGet(ctx, sessionKey, "current_hash").Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return Error.Internal(err)
	}

	count, err := r.redis.Del(ctx, sessionKey).Result()
	if err != nil {
		return Error.Internal(err)
	}
	if count == 0 {
		_ = r.redis.SRem(ctx, userSessionsKey, dto.SessionUUID).Err()
		return Error.Public(codes.NotFound, "session not found")
	}

	pipeline := r.redis.Pipeline()
	pipeline.SRem(ctx, userSessionsKey, dto.SessionUUID)
	if currentHash != "" {
		pipeline.Del(ctx, r.getTokenIndexKey(currentHash))
	}
	if _, err := pipeline.Exec(ctx); err != nil {
		return Error.Internal(err)
	}

	return Error.CodeError{}
}

// RevokeAllSessions отзывает все сессии пользователя
func (r *authRepository) RevokeAllSessions(ctx context.Context, dto entities.RevokeAllSessionsDTO) Error.CodeError {
	userSessionsKey := r.getUserSessionsKey(dto.UserUUID)

	sessionUUIDs, err := r.redis.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return Error.Internal(err)
	}

	if len(sessionUUIDs) == 0 {
		return Error.Public(codes.NotFound, "sessions not found")
	}

	pipeline := r.redis.Pipeline()
	for _, sid := range sessionUUIDs {
		pipeline.Del(ctx, r.getSessionKey(sid))
	}
	pipeline.Del(ctx, userSessionsKey)
	if _, err := pipeline.Exec(ctx); err != nil {
		return Error.Internal(err)
	}

	return Error.CodeError{}
}

// RefreshToken атомарно ротирует refresh токен.
// При обнаружении повторного использования (current_hash не совпадает) отзывает сессию и возвращает entities.ErrTokenReuse
func (r *authRepository) RefreshToken(ctx context.Context, dto entities.RefreshTokenDTO) Error.CodeError {
	tokenIndexKey := r.getTokenIndexKey(dto.OldHashToken)

	// Получаем sessionUUID из индекса токенов
	sessionUUID, err := r.redis.Get(ctx, tokenIndexKey).Result()
	if errors.Is(err, redis.Nil) || sessionUUID == "" {
		return Error.Public(codes.NotFound, "session not found")
	}
	if err != nil {
		return Error.Internal(err)
	}

	sessionKey := r.getSessionKey(sessionUUID)
	userSessionsKey := r.getUserSessionsKey(dto.UserUUID)
	newTokenIndexKey := r.getTokenIndexKey(dto.NewHashToken)

	watchErr := r.redis.Watch(ctx, func(tx *redis.Tx) error {
		fields, err := tx.HGetAll(ctx, sessionKey).Result()
		if err != nil {
			return err
		}
		if len(fields) == 0 {
			return errSessionNotFound
		}

		// Детектируем повторное использование токена
		if fields["current_hash"] != dto.OldHashToken {
			// Токен уже был ротирован — атакующий пытается использовать старый токен.
			// Отзываем скомпрометированную сессию.
			_, _ = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Del(ctx, sessionKey)
				pipe.SRem(ctx, userSessionsKey, sessionUUID)
				pipe.Del(ctx, tokenIndexKey)
				return nil
			})
			return entities.ErrTokenReuse
		}

		// Токен актуален — обновляем сессию
		fields["last_ip"] = dto.LastIP
		fields["last_active"] = strconv.FormatInt(dto.LastActiveAt.Unix(), 10)
		fields["current_hash"] = dto.NewHashToken

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HSet(ctx, sessionKey, fields)
			pipe.Expire(ctx, sessionKey, r.refreshTokenTTL)
			pipe.Set(ctx, newTokenIndexKey, sessionUUID, r.refreshTokenTTL)
			// Оставляем старый ключ как tombstone на 1 час для детектирования повторного использования
			pipe.Expire(ctx, tokenIndexKey, tokenIndexTombstoneTTL)
			return nil
		})
		return err
	}, sessionKey)

	if watchErr != nil {
		if errors.Is(watchErr, errSessionNotFound) {
			return Error.Public(codes.NotFound, "session not found")
		}
		if errors.Is(watchErr, entities.ErrTokenReuse) {
			return Error.CodeError{Code: codes.Unauthenticated, Err: entities.ErrTokenReuse}
		}
		return Error.Internal(watchErr)
	}
	return Error.CodeError{}
}

// ── Вспомогательные функции ───────────────────────────────────────────────────

func (r *authRepository) getSessionKey(sessionUUID string) string {
	return fmt.Sprintf("%s:session:%s", r.prefix, sessionUUID)
}

func (r *authRepository) getTokenIndexKey(hash string) string {
	return fmt.Sprintf("%s:token:%s", r.prefix, hash)
}

func (r *authRepository) getUserSessionsKey(userUUID string) string {
	return fmt.Sprintf("%s:user:%s:sessions", r.prefix, userUUID)
}
