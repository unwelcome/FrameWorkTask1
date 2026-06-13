package entities

import "time"

type SaveSessionDTO struct {
	UserUUID    string
	HashedToken string
	Session     *SessionInfo
}

type GetAllSessionsDTO struct {
	UserUUID string
}

// SessionEntry — одна активная сессия с данными и хешем refresh токена
type SessionEntry struct {
	TokenHash string
	Session   *SessionInfo
}

type CheckSessionExistsDTO struct {
	UserUUID string
	RawToken string
}

type RevokeSessionDTO struct {
	UserUUID  string
	TokenHash string
}

type RevokeAllSessionsDTO struct {
	UserUUID string
}

type RefreshTokenDTO struct {
	UserUUID     string
	OldHashToken string
	NewHashToken string
	LastIP       string    // IP последнего refresh — обновляется в хеше
	LastActiveAt time.Time // Время последнего refresh — обновляется в хеше
}
