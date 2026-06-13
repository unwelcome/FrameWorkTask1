package redisDB

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/pkg/utils"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
)

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

// SaveSession Сохраняет хеш refresh токена вместе с данными сессии.
func (r *authRepository) SaveSession(ctx context.Context, dto entities.SaveSessionDTO) Error.CodeError {
	userSessionsKey := r.getUserSessionsKey(dto.UserUUID)
	sessionKey := r.getSessionKey(dto.HashedToken)

	pipeline := r.redis.Pipeline()
	pipeline.HSet(ctx, sessionKey, dto.Session.ToMap())
	pipeline.Expire(ctx, sessionKey, r.refreshTokenTTL)
	pipeline.SAdd(ctx, userSessionsKey, dto.HashedToken)

	_, err := pipeline.Exec(ctx)
	if err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// GetAllSessions Возвращает все активные сессии пользователя.
func (r *authRepository) GetAllSessions(ctx context.Context, dto entities.GetAllSessionsDTO) ([]entities.SessionEntry, Error.CodeError) {
	userSessionsKey := r.getUserSessionsKey(dto.UserUUID)

	hashedTokens, err := r.redis.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, Error.CodeError{Code: int(codes.Internal), Err: err}
	}

	if len(hashedTokens) == 0 {
		return []entities.SessionEntry{}, Error.CodeError{}
	}

	// Пайплайн: HGETALL для каждой сессии за один сетевой round-trip
	pipe := r.redis.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(hashedTokens))
	for i, hash := range hashedTokens {
		cmds[i] = pipe.HGetAll(ctx, r.getSessionKey(hash))
	}
	if _, err = pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, Error.Internal(err)
	}

	activeEntries := make([]entities.SessionEntry, 0, len(hashedTokens))
	expiredHashes := make([]interface{}, 0)

	for i, hash := range hashedTokens {
		fields, cmdErr := cmds[i].Result()
		if cmdErr != nil || len(fields) == 0 {
			// Ключ истёк или не существует — чистим из сета
			expiredHashes = append(expiredHashes, hash)
			continue
		}

		session := &entities.SessionInfo{}
		session.FromMap(fields)

		activeEntries = append(activeEntries, entities.SessionEntry{
			TokenHash: hash,
			Session:   session,
		})
	}

	if len(expiredHashes) > 0 {
		_ = r.redis.SRem(ctx, userSessionsKey, expiredHashes...).Err()
	}

	return activeEntries, Error.CodeError{}
}

// CheckSessionExists Проверяет существование сессии по хешу refresh токена.
func (r *authRepository) CheckSessionExists(ctx context.Context, dto entities.CheckSessionExistsDTO) Error.CodeError {
	hash := utils.HashToken(dto.RawToken)
	sessionKey := r.getSessionKey(hash)

	exist, err := r.redis.Exists(ctx, sessionKey).Result()
	if err != nil {
		return Error.Internal(err)
	}

	if exist == 0 {
		userSessionsKey := r.getUserSessionsKey(dto.UserUUID)
		_ = r.redis.SRem(ctx, userSessionsKey, hash).Err()
		return Error.Public(codes.NotFound, "session not found")
	}

	return Error.CodeError{}
}

// RevokeSession Отзывает конкретную сессию, проверяя принадлежность пользователю.
func (r *authRepository) RevokeSession(ctx context.Context, dto entities.RevokeSessionDTO) Error.CodeError {
	userSessionsKey := r.getUserSessionsKey(dto.UserUUID)
	sessionKey := r.getSessionKey(dto.TokenHash)

	isMember, err := r.redis.SIsMember(ctx, userSessionsKey, dto.TokenHash).Result()
	if err != nil {
		return Error.Internal(err)
	}
	if !isMember {
		return Error.Public(codes.NotFound, "session not found")
	}

	count, err := r.redis.Del(ctx, sessionKey).Result()
	if err != nil {
		return Error.Internal(err)
	}
	if count == 0 {
		// Сессия уже истекла, но ещё числилась в сете — чистим
		_ = r.redis.SRem(ctx, userSessionsKey, dto.TokenHash).Err()
		return Error.Public(codes.NotFound, "session not found")
	}

	if err := r.redis.SRem(ctx, userSessionsKey, dto.TokenHash).Err(); err != nil {
		return Error.Internal(err)
	}

	return Error.CodeError{}
}

// RevokeAllSessions Отзывает все сессии пользователя.
func (r *authRepository) RevokeAllSessions(ctx context.Context, dto entities.RevokeAllSessionsDTO) Error.CodeError {
	userSessionsKey := r.getUserSessionsKey(dto.UserUUID)

	hashedTokens, err := r.redis.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return Error.Internal(err)
	}

	if len(hashedTokens) == 0 {
		return Error.Public(codes.NotFound, "sessions not found")
	}

	pipeline := r.redis.Pipeline()
	for _, hash := range hashedTokens {
		pipeline.Del(ctx, r.getSessionKey(hash))
	}
	pipeline.Del(ctx, userSessionsKey)
	if _, err := pipeline.Exec(ctx); err != nil {
		return Error.Internal(err)
	}

	return Error.CodeError{}
}

// RefreshToken Атомарно заменяет старую сессию на новую,
// копируя данные и обновляя изменяемые поля (LastIP, LastActiveAt).
func (r *authRepository) RefreshToken(ctx context.Context, dto entities.RefreshTokenDTO) Error.CodeError {
	userSessionsKey := r.getUserSessionsKey(dto.UserUUID)
	oldSessionKey := r.getSessionKey(dto.OldHashToken)
	newSessionKey := r.getSessionKey(dto.NewHashToken)

	err := r.redis.Watch(ctx, func(tx *redis.Tx) error {
		// Читаем все поля старой сессии в рамках Watch-транзакции
		fields, err := tx.HGetAll(ctx, oldSessionKey).Result()
		if err != nil {
			return err
		}
		if len(fields) == 0 {
			return errSessionNotFound
		}

		// Обновляем изменяемые поля
		fields["last_ip"] = dto.LastIP
		fields["last_active"] = strconv.FormatInt(dto.LastActiveAt.Unix(), 10)

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HSet(ctx, newSessionKey, fields)
			pipe.Expire(ctx, newSessionKey, r.refreshTokenTTL)
			pipe.SAdd(ctx, userSessionsKey, dto.NewHashToken)
			pipe.Del(ctx, oldSessionKey)
			pipe.SRem(ctx, userSessionsKey, dto.OldHashToken)
			return nil
		})
		return err
	}, oldSessionKey)

	if err != nil {
		if errors.Is(err, errSessionNotFound) {
			return Error.Public(codes.NotFound, "session not found")
		}
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// ── Вспомогательные функции ───────────────────────────────────────────────────

func (r *authRepository) getSessionKey(hash string) string {
	return fmt.Sprintf("%s:session:%s", r.prefix, hash)
}

func (r *authRepository) getUserSessionsKey(userUUID string) string {
	return fmt.Sprintf("%s:user:%s:sessions", r.prefix, userUUID)
}
