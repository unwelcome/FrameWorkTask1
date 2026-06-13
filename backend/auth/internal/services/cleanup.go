package services

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/postgres"
)

// cleanupHour — час запуска анонимизации (UTC).
// 00:00 UTC = 03:00 MSK
const (
	cleanupHour   = 0
	cleanupPeriod = 24 * time.Hour
)

// StartCleanupWorker запускает фоновую горутину, которая ежедневно в cleanupHour:00 UTC
// анонимизирует персональные данные пользователей, удалённых более accountDeletionRetention назад.
// Останавливается при отмене ctx (graceful shutdown).
func StartCleanupWorker(ctx context.Context, db *postgresDB.DatabaseRepository) {
	next := nextCleanupTimeAfter(time.Now())
	log.Info().Time("next_run", next).Msg("cleanup worker scheduled")

	timer := time.NewTimer(time.Until(next))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			runAnonymize(ctx, db)
			next = nextCleanupTimeAfter(time.Now())
			timer.Reset(time.Until(next))
			log.Info().Time("next_run", next).Msg("cleanup")
		case <-ctx.Done():
			log.Info().Msg("cleanup worker stopped")
			return
		}
	}
}

// nextCleanupTimeAfter возвращает первый момент cleanupHour:00:00 UTC строго после t
func nextCleanupTimeAfter(t time.Time) time.Time {
	t = t.UTC()
	candidate := time.Date(t.Year(), t.Month(), t.Day(), cleanupHour, 0, 0, 0, time.UTC)
	if !candidate.After(t) {
		candidate = candidate.Add(cleanupPeriod)
	}
	return candidate
}

// runAnonymize Анонимизация удаленных аккаунтов пользователей
func runAnonymize(ctx context.Context, db *postgresDB.DatabaseRepository) {
	before := time.Now().Add(-accountDeletionRetention)
	count, err := db.User.AnonymizeExpiredUsers(ctx, before)
	if err != nil {
		log.Error().Err(err).Msg("cleanup: failed to anonymize expired deleted users")
		return
	}
	log.Info().Int64("anonymized", count).Msg("cleanup")
}

// deletedAccountMessage формирует сообщение об ошибке для удалённого аккаунта.
// Показывает оставшееся время до анонимизации с точностью до минут:
//   - ≥ 1 час  → "you have N hours to restore it"
//   - ≥ 1 мин  → "you have N minutes to restore it"
//   - < 1 мин  → "account is deleted" (cleanup уже фактически наступил)
func deletedAccountMessage(deletedAt time.Time) string {
	anonymizationTime := nextCleanupTimeAfter(deletedAt.Add(accountDeletionRetention))
	remaining := time.Until(anonymizationTime)
	switch {
	case remaining >= time.Hour:
		return fmt.Sprintf("account is deleted, you have %d hours to restore it", int(remaining.Hours()))
	case remaining >= time.Minute:
		return fmt.Sprintf("account is deleted, you have %d minutes to restore it", int(remaining.Minutes()))
	default:
		return "account is deleted"
	}
}
