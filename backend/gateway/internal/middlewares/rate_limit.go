package middlewares

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	Error "github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/errors"
)

// rateLimitRedisTimeout ограничивает один вызов к Redis в rate-limiter-е
const rateLimitRedisTimeout = 100 * time.Millisecond

// tokenBucketScript реализует алгоритм token bucket целиком на стороне Redis,
// чтобы чтение состояния, пополнение, списание токена и установка TTL
// выполнялись атомарно (без гонок между параллельными запросами одного ключа).
//
// Состояние хранится в HASH из двух полей:
//   - tokens — текущее (возможно дробное) число токенов;
//   - ts     — момент последнего пересчёта, unix-время в миллисекундах.
//
// Аргументы:
//
//	KEYS[1] — ключ корзины
//	ARGV[1] — capacity (размер корзины / максимальный burst)
//	ARGV[2] — refill rate (токенов в секунду)
//	ARGV[3] — now (unix-время в миллисекундах)
//	ARGV[4] — requested (сколько токенов списываем, обычно 1)
//
// Возвращает массив {allowed, retry_after}:
//   - allowed = 1, если запрос разрешён, иначе 0;
//   - retry_after — через сколько секунд накопится нужное число токенов (при отказе).
var tokenBucketScript = redis.NewScript(`
local capacity    = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now         = tonumber(ARGV[3])
local requested   = tonumber(ARGV[4])

local data   = redis.call('HMGET', KEYS[1], 'tokens', 'ts')
local tokens = tonumber(data[1])
local last   = tonumber(data[2])

-- Новая корзина — заполнена до краёв.
if tokens == nil then
    tokens = capacity
    last   = now
end

-- Пополняем пропорционально прошедшему времени, но не выше capacity.
local elapsed = now - last
if elapsed < 0 then
    elapsed = 0
end
tokens = math.min(capacity, tokens + (elapsed / 1000.0) * refill_rate)

local allowed = 0
if tokens >= requested then
    tokens  = tokens - requested
    allowed = 1
end

-- tokens пишем строкой, чтобы не потерять дробную часть при сериализации.
redis.call('HSET', KEYS[1], 'tokens', tostring(tokens), 'ts', now)

-- TTL = время полного наполнения пустой корзины + запас.
-- Простаивающие корзины удаляются; вернувшийся клиент получит полную корзину.
local ttl = math.ceil(capacity / refill_rate) + 1
redis.call('EXPIRE', KEYS[1], ttl)

local retry_after = 0
if allowed == 0 then
    retry_after = math.ceil((requested - tokens) / refill_rate)
end

return {allowed, retry_after}
`)

// RateLimiterConfig конфигурация одного тира token bucket rate-limiter-а.
type RateLimiterConfig struct {
	Capacity     int     // размер корзины (максимальный burst)
	RefillPerSec float64 // скорость пополнения, токенов в секунду
	// KeyFn возвращает строку, идентифицирующую клиента (IP или userUUID).
	KeyFn func(*fiber.Ctx) string
}

// NewRateLimiter возвращает Fiber middleware для ограничения частоты запросов
// по алгоритму token bucket. tierPrefix используется как часть Redis-ключа
// (например "password", "code", "user").
// При недоступности Redis middleware пропускает запрос (fail-open), чтобы не
// блокировать легитимных пользователей из-за инфраструктурных проблем.
func NewRateLimiter(client *redis.Client, prefix, tierPrefix string, cfg RateLimiterConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := fmt.Sprintf("%s:rl:%s:%s", prefix, tierPrefix, cfg.KeyFn(c))
		ctx, cancel := context.WithTimeout(context.Background(), rateLimitRedisTimeout)
		defer cancel()

		res, err := tokenBucketScript.Run(
			ctx, client, []string{key},
			cfg.Capacity, cfg.RefillPerSec, time.Now().UnixMilli(), 1,
		).Int64Slice()
		if err != nil || len(res) < 2 {
			// fail-open: при сбое или таймауте Redis пропускаем запрос, чтобы
			// инфраструктурная проблема не блокировала легитимный трафик.
			// Логируем на Warn — защита временно отключена, это должно быть видно.
			if err != nil {
				log.Warn().Err(err).Str("tier", tierPrefix).Msg("rate limiter: redis unavailable, failing open")
			}
			return c.Next()
		}

		if res[0] == 0 {
			c.Set(fiber.HeaderRetryAfter, fmt.Sprintf("%d", res[1]))
			return c.Status(fiber.StatusTooManyRequests).JSON(Error.HttpError{Code: fiber.StatusTooManyRequests, Message: "too many requests"})
		}

		return c.Next()
	}
}
