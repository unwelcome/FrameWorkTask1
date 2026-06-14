package entities

import (
	"errors"
	"time"
)

var ErrTokenReuse = errors.New("refresh token reuse detected")

type SaveSessionDTO struct {
	UserUUID    string
	SessionUUID string
	HashedToken string
	Session     *SessionInfo
}

type GetAllSessionsDTO struct {
	UserUUID string
}

// SessionEntry — одна активная сессия с данными и UUID сессии
type SessionEntry struct {
	SessionUUID string
	Session     *SessionInfo
}

type CheckSessionExistsDTO struct {
	UserUUID    string
	HashedToken string
}

type RevokeSessionDTO struct {
	UserUUID    string
	SessionUUID string
}

type RevokeAllSessionsDTO struct {
	UserUUID string
}

type RefreshTokenDTO struct {
	UserUUID     string
	OldHashToken string
	NewHashToken string
	LastIP       string
	LastActiveAt time.Time
}
