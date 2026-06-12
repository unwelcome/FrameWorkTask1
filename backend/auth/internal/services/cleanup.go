package services

import (
	"context"
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

// hoursUntilAnonymization возвращает точное количество полных часов до анонимизации аккаунта
func hoursUntilAnonymization(deletedAt time.Time) int {
	retentionDeadline := deletedAt.Add(accountDeletionRetention)
	anonymizationTime := nextCleanupTimeAfter(retentionDeadline)
	remaining := time.Until(anonymizationTime)
	if remaining <= 0 {
		return 0
	}
	return int(remaining.Hours()) // floor: 59m → 0, 61m → 1
}
