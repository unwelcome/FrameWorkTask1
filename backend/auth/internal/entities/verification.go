package entities

import "time"

type AddToVerificationTokenBlacklistDTO struct {
	TokenID string
	TTL     time.Duration
}

type IsVerificationTokenBlacklistedDTO struct {
	TokenID string
}
