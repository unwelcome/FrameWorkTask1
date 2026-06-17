package entities

import "time"

type ConsumeVerificationTokenDTO struct {
	TokenID string
	TTL     time.Duration
}

type AcquireVerificationEmailCooldownDTO struct {
	UserUUID string
}

type IncrVerificationEmailDailyCountDTO struct {
	UserUUID string
}
