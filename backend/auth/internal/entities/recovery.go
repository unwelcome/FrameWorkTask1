package entities

import "time"

type AddToResetTokenBlacklistDTO struct {
	TokenID string
	TTL     time.Duration
}

type IsResetTokenBlacklistedDTO struct {
	TokenID string
}
