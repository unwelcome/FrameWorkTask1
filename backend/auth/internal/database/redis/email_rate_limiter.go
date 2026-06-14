package redisDB

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
)

const (
	emailCooldownTTL = 180 * time.Second
	emailDailyTTL    = 24 * time.Hour
)

// emailRateLimiter implements per-user email send rate limiting via Redis
type emailRateLimiter struct {
	redis     *redis.Client
	keyPrefix string
}

func newEmailRateLimiter(rdb *redis.Client, keyPrefix string) emailRateLimiter {
	return emailRateLimiter{redis: rdb, keyPrefix: keyPrefix}
}

// acquireCooldown sets the cooldown key via SetNX.
// Returns true when the key was created (send allowed); false when cooldown is still active.
func (l *emailRateLimiter) acquireCooldown(ctx context.Context, userUUID string) (bool, Error.CodeError) {
	set, err := l.redis.SetNX(ctx, l.cooldownKey(userUUID), 1, emailCooldownTTL).Result()
	if err != nil {
		return false, Error.Internal(err)
	}
	return set, Error.CodeError{}
}

// incrDailyCount increments the 24-hour send counter and returns the new value.
// The TTL is set atomically on the first increment.
func (l *emailRateLimiter) incrDailyCount(ctx context.Context, userUUID string) (int64, Error.CodeError) {
	key := l.dailyKey(userUUID)
	count, err := l.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, Error.Internal(err)
	}
	if count == 1 {
		if err := l.redis.Expire(ctx, key, emailDailyTTL).Err(); err != nil {
			return 0, Error.Internal(err)
		}
	}
	return count, Error.CodeError{}
}

func (l *emailRateLimiter) cooldownKey(userUUID string) string {
	return fmt.Sprintf("%s:%s:email:cooldown", l.keyPrefix, userUUID)
}

func (l *emailRateLimiter) dailyKey(userUUID string) string {
	return fmt.Sprintf("%s:%s:email:daily", l.keyPrefix, userUUID)
}
