package entities

import "time"

type ConsumeResetTokenDTO struct {
	TokenID string
	TTL     time.Duration
}

type AcquireRecoveryEmailCooldownDTO struct {
	UserUUID string
}

type IncrRecoveryEmailDailyCountDTO struct {
	UserUUID string
}
