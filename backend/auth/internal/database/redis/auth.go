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

var errRefreshTokenNotFound = errors.New("refresh token not found")

type AuthRepository interface {
	SaveRefreshToken(ctx context.Context, dto entities.SaveRefreshTokenDTO) Error.CodeError
	GetAllRefreshTokens(ctx context.Context, dto entities.GetAllRefreshTokensDTO) ([]entities.RefreshTokenEntry, Error.CodeError)
	CheckRefreshTokenExists(ctx context.Context, dto entities.CheckRefreshTokenExistsDTO) Error.CodeError
	RevokeRefreshToken(ctx context.Context, dto entities.RevokeRefreshTokenDTO) Error.CodeError
	RevokeAllRefreshTokens(ctx context.Context, dto entities.RevokeAllRefreshTokensDTO) Error.CodeError
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

// SaveRefreshToken Сохраняет хеш refresh токена вместе с данными сессии.
func (r *authRepository) SaveRefreshToken(ctx context.Context, dto entities.SaveRefreshTokenDTO) Error.CodeError {
	userTokensKey := r.getUserTokensKey(dto.UserUUID)
	tokenKey := r.getRefreshTokenKey(dto.HashedToken)

	pipeline := r.redis.Pipeline()
	pipeline.HSet(ctx, tokenKey, sessionToFields(dto.Session))
	pipeline.Expire(ctx, tokenKey, r.refreshTokenTTL)
	pipeline.SAdd(ctx, userTokensKey, dto.HashedToken)

	_, err := pipeline.Exec(ctx)
	if err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// GetAllRefreshTokens Возвращает все активные токены пользователя с данными сессий.
func (r *authRepository) GetAllRefreshTokens(ctx context.Context, dto entities.GetAllRefreshTokensDTO) ([]entities.RefreshTokenEntry, Error.CodeError) {
	userTokensKey := r.getUserTokensKey(dto.UserUUID)

	hashedTokens, err := r.redis.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return nil, Error.CodeError{Code: int(codes.Internal), Err: err}
	}

	if len(hashedTokens) == 0 {
		return []entities.RefreshTokenEntry{}, Error.CodeError{}
	}

	// Пайплайн: HGETALL для каждого токена за один сетевой round-trip
	pipe := r.redis.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(hashedTokens))
	for i, hash := range hashedTokens {
		cmds[i] = pipe.HGetAll(ctx, r.getRefreshTokenKey(hash))
	}
	if _, err = pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, Error.Internal(err)
	}

	activeEntries := make([]entities.RefreshTokenEntry, 0, len(hashedTokens))
	expiredHashes := make([]interface{}, 0)

	for i, hash := range hashedTokens {
		fields, cmdErr := cmds[i].Result()
		if cmdErr != nil || len(fields) == 0 {
			// Ключ истёк или не существует — чистим из сета
			expiredHashes = append(expiredHashes, hash)
			continue
		}
		activeEntries = append(activeEntries, entities.RefreshTokenEntry{
			TokenHash: hash,
			Session:   sessionFromFields(fields),
		})
	}

	if len(expiredHashes) > 0 {
		_ = r.redis.SRem(ctx, userTokensKey, expiredHashes...).Err()
	}

	return activeEntries, Error.CodeError{}
}

// CheckRefreshTokenExists Проверяет существование refresh токена по хешу.
func (r *authRepository) CheckRefreshTokenExists(ctx context.Context, dto entities.CheckRefreshTokenExistsDTO) Error.CodeError {
	hash := utils.HashToken(dto.RawToken)
	tokenKey := r.getRefreshTokenKey(hash)

	exist, err := r.redis.Exists(ctx, tokenKey).Result()
	if err != nil {
		return Error.Internal(err)
	}

	if exist == 0 {
		userTokensKey := r.getUserTokensKey(dto.UserUUID)
		_ = r.redis.SRem(ctx, userTokensKey, hash).Err()
		return Error.Public(codes.NotFound, "refresh token not found")
	}

	return Error.CodeError{}
}

// RevokeRefreshToken Отзывает конкретный refresh токен, проверяя принадлежность пользователю.
func (r *authRepository) RevokeRefreshToken(ctx context.Context, dto entities.RevokeRefreshTokenDTO) Error.CodeError {
	userTokensKey := r.getUserTokensKey(dto.UserUUID)
	tokenKey := r.getRefreshTokenKey(dto.TokenHash)

	isMember, err := r.redis.SIsMember(ctx, userTokensKey, dto.TokenHash).Result()
	if err != nil {
		return Error.Internal(err)
	}
	if !isMember {
		return Error.Public(codes.NotFound, "refresh token not found")
	}

	count, err := r.redis.Del(ctx, tokenKey).Result()
	if err != nil {
		return Error.Internal(err)
	}
	if count == 0 {
		// Токен уже истёк, но ещё числился в сете — чистим
		_ = r.redis.SRem(ctx, userTokensKey, dto.TokenHash).Err()
		return Error.Public(codes.NotFound, "refresh token not found")
	}

	if err := r.redis.SRem(ctx, userTokensKey, dto.TokenHash).Err(); err != nil {
		return Error.Internal(err)
	}

	return Error.CodeError{}
}

// RevokeAllRefreshTokens Отзывает все refresh токены пользователя.
func (r *authRepository) RevokeAllRefreshTokens(ctx context.Context, dto entities.RevokeAllRefreshTokensDTO) Error.CodeError {
	userTokensKey := r.getUserTokensKey(dto.UserUUID)

	hashedTokens, err := r.redis.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return Error.Internal(err)
	}

	if len(hashedTokens) == 0 {
		return Error.Public(codes.NotFound, "refresh tokens not found")
	}

	pipeline := r.redis.Pipeline()
	for _, hash := range hashedTokens {
		pipeline.Del(ctx, r.getRefreshTokenKey(hash))
	}
	pipeline.Del(ctx, userTokensKey)
	if _, err := pipeline.Exec(ctx); err != nil {
		return Error.Internal(err)
	}

	return Error.CodeError{}
}

// RefreshToken Атомарно заменяет старый хеш на новый,
// копируя данные сессии и обновляя изменяемые поля (LastIP, LastActiveAt).
func (r *authRepository) RefreshToken(ctx context.Context, dto entities.RefreshTokenDTO) Error.CodeError {
	userTokensKey := r.getUserTokensKey(dto.UserUUID)
	oldTokenKey := r.getRefreshTokenKey(dto.OldHashToken)
	newTokenKey := r.getRefreshTokenKey(dto.NewHashToken)

	err := r.redis.Watch(ctx, func(tx *redis.Tx) error {
		// Читаем все поля старого токена в рамках Watch-транзакции
		fields, err := tx.HGetAll(ctx, oldTokenKey).Result()
		if err != nil {
			return err
		}
		if len(fields) == 0 {
			return errRefreshTokenNotFound
		}

		// Обновляем изменяемые поля
		fields["last_ip"] = dto.LastIP
		fields["last_active"] = strconv.FormatInt(dto.LastActiveAt.Unix(), 10)

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HSet(ctx, newTokenKey, fields)
			pipe.Expire(ctx, newTokenKey, r.refreshTokenTTL)
			pipe.SAdd(ctx, userTokensKey, dto.NewHashToken)
			pipe.Del(ctx, oldTokenKey)
			pipe.SRem(ctx, userTokensKey, dto.OldHashToken)
			return nil
		})
		return err
	}, oldTokenKey)

	if err != nil {
		if errors.Is(err, errRefreshTokenNotFound) {
			return Error.Public(codes.NotFound, "refresh token not found")
		}
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// ── Вспомогательные функции ───────────────────────────────────────────────────

func (r *authRepository) getRefreshTokenKey(hash string) string {
	return fmt.Sprintf("%s:token:%s", r.prefix, hash)
}

func (r *authRepository) getUserTokensKey(userUUID string) string {
	return fmt.Sprintf("%s:user:%s:tokens", r.prefix, userUUID)
}

// sessionToFields конвертирует SessionInfo в плоский map для HSET.
func sessionToFields(s entities.SessionInfo) map[string]interface{} {
	return map[string]interface{}{
		"ip":           s.IP,
		"last_ip":      s.LastIP,
		"isp":          s.ISP,
		"country_code": s.CountryCode,
		"country_name": s.CountryName,
		"city":         s.City,
		"timezone":     s.Timezone,
		"device_type":  s.DeviceType,
		"os":           s.OS,
		"os_version":   s.OSVersion,
		"browser":      s.Browser,
		"browser_ver":  s.BrowserVersion,
		"ua":           s.UserAgentRaw,
		"created_at":   strconv.FormatInt(s.CreatedAt.Unix(), 10),
		"last_active":  strconv.FormatInt(s.LastActiveAt.Unix(), 10),
	}
}

// sessionFromFields восстанавливает SessionInfo из результата HGETALL.
func sessionFromFields(fields map[string]string) entities.SessionInfo {
	s := entities.SessionInfo{
		IP:             fields["ip"],
		LastIP:         fields["last_ip"],
		ISP:            fields["isp"],
		CountryCode:    fields["country_code"],
		CountryName:    fields["country_name"],
		City:           fields["city"],
		Timezone:       fields["timezone"],
		DeviceType:     fields["device_type"],
		OS:             fields["os"],
		OSVersion:      fields["os_version"],
		Browser:        fields["browser"],
		BrowserVersion: fields["browser_ver"],
		UserAgentRaw:   fields["ua"],
	}
	if ts, err := strconv.ParseInt(fields["created_at"], 10, 64); err == nil {
		s.CreatedAt = time.Unix(ts, 0)
	}
	if ts, err := strconv.ParseInt(fields["last_active"], 10, 64); err == nil {
		s.LastActiveAt = time.Unix(ts, 0)
	}
	return s
}
