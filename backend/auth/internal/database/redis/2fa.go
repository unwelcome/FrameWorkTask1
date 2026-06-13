package redisDB

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
)

const (
	twoFACodeTTL          = 15 * time.Minute
	twoFAEmailCooldownTTL = 60 * time.Second
	twoFAEmailDailyTTL    = 24 * time.Hour
)

type TwoFARepository interface {
	Save2FAData(ctx context.Context, dto entities.Save2FADataDTO) Error.CodeError
	Get2FAData(ctx context.Context, dto entities.Get2FADataDTO) (*entities.TwoFAData, Error.CodeError)
	Delete2FAData(ctx context.Context, dto entities.Delete2FADataDTO) Error.CodeError
	Incr2FAAttempts(ctx context.Context, dto entities.Incr2FAAttemptsDTO) (int64, Error.CodeError)
	// Acquire2FAEmailCooldown устанавливает cooldown-ключ (SetNX). Возвращает true, если разрешено отправить письмо.
	Acquire2FAEmailCooldown(ctx context.Context, dto entities.Acquire2FAEmailCooldownDTO) (bool, Error.CodeError)
	// Incr2FAEmailDailyCount увеличивает суточный счётчик отправок 2FA писем и возвращает новое значение.
	Incr2FAEmailDailyCount(ctx context.Context, dto entities.Incr2FAEmailDailyCountDTO) (int64, Error.CodeError)
}

type twoFARepository struct {
	redis  *redis.Client
	prefix string
}

func NewTwoFARepository(redis *redis.Client, prefix string) TwoFARepository {
	return &twoFARepository{redis: redis, prefix: prefix}
}

// Save2FAData Сохраняет 2FA данные по uuid сессии авторизации
func (r *twoFARepository) Save2FAData(ctx context.Context, dto entities.Save2FADataDTO) Error.CodeError {
	body, err := json.Marshal(entities.TwoFAData{
		UserUUID:  dto.UserUUID,
		Email:     dto.Email,
		FirstName: dto.FirstName,
		Code:      dto.Code,
	})
	if err != nil {
		return Error.Internal(err)
	}

	pipe := r.redis.Pipeline()
	pipe.Set(ctx, r.get2FADataKey(dto.SessionUUID), body, twoFACodeTTL)
	pipe.Set(ctx, r.get2FAAttemptsKey(dto.SessionUUID), 0, twoFACodeTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// Get2FAData Получает 2FA данные по uuid сессии авторизации
func (r *twoFARepository) Get2FAData(ctx context.Context, dto entities.Get2FADataDTO) (*entities.TwoFAData, Error.CodeError) {
	body, err := r.redis.Get(ctx, r.get2FADataKey(dto.SessionUUID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, Error.Public(codes.NotFound, "2FA code not found")
		}
		return nil, Error.Internal(err)
	}

	data := &entities.TwoFAData{}
	if err := json.Unmarshal([]byte(body), data); err != nil {
		return nil, Error.Internal(err)
	}

	return data, Error.CodeError{}
}

// Delete2FAData Удаляет 2FA данные по uuid сессии авторизации
func (r *twoFARepository) Delete2FAData(ctx context.Context, dto entities.Delete2FADataDTO) Error.CodeError {
	pipe := r.redis.Pipeline()
	pipe.Del(ctx, r.get2FADataKey(dto.SessionUUID))
	pipe.Del(ctx, r.get2FAAttemptsKey(dto.SessionUUID))

	if _, err := pipe.Exec(ctx); err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// Incr2FAAttempts Увеличивает счетчик попыток 2FA авторизации и возвращает текущее значение
func (r *twoFARepository) Incr2FAAttempts(ctx context.Context, dto entities.Incr2FAAttemptsDTO) (int64, Error.CodeError) {
	count, err := r.redis.Incr(ctx, r.get2FAAttemptsKey(dto.SessionUUID)).Result()
	if err != nil {
		return 0, Error.Internal(err)
	}
	return count, Error.CodeError{}
}

// Acquire2FAEmailCooldown Устанавливает cooldown-ключ. Возвращает true, если ключ создан (разрешено отправить),
// false — если ключ уже существует (cooldown ещё активен).
func (r *twoFARepository) Acquire2FAEmailCooldown(ctx context.Context, dto entities.Acquire2FAEmailCooldownDTO) (bool, Error.CodeError) {
	set, err := r.redis.SetNX(ctx, r.get2FAEmailCooldownKey(dto.UserUUID), 1, twoFAEmailCooldownTTL).Result()
	if err != nil {
		return false, Error.Internal(err)
	}
	return set, Error.CodeError{}
}

// Incr2FAEmailDailyCount Увеличивает суточный счётчик отправок 2FA писем
func (r *twoFARepository) Incr2FAEmailDailyCount(ctx context.Context, dto entities.Incr2FAEmailDailyCountDTO) (int64, Error.CodeError) {
	key := r.get2FAEmailDailyKey(dto.UserUUID)
	count, err := r.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, Error.Internal(err)
	}
	if count == 1 {
		r.redis.Expire(ctx, key, twoFAEmailDailyTTL)
	}
	return count, Error.CodeError{}
}

// ─── Вспомогательные функции ──────────────────────────────────────────────────

func (r *twoFARepository) get2FADataKey(sessionUUID string) string {
	return fmt.Sprintf("%s:2fa:%s:data", r.prefix, sessionUUID)
}

func (r *twoFARepository) get2FAAttemptsKey(sessionUUID string) string {
	return fmt.Sprintf("%s:2fa:%s:attempts", r.prefix, sessionUUID)
}

func (r *twoFARepository) get2FAEmailCooldownKey(userUUID string) string {
	return fmt.Sprintf("%s:2fa:%s:email:cooldown", r.prefix, userUUID)
}

func (r *twoFARepository) get2FAEmailDailyKey(userUUID string) string {
	return fmt.Sprintf("%s:2fa:%s:email:daily", r.prefix, userUUID)
}
