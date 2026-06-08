package entities

import "time"

type SaveRefreshTokenDTO struct {
	UserUUID    string
	HashedToken string
	Session     SessionInfo
}

type GetAllRefreshTokensDTO struct {
	UserUUID string
}

// RefreshTokenEntry — один активный токен с данными сессии.
type RefreshTokenEntry struct {
	TokenHash string
	Session   SessionInfo
}

type CheckRefreshTokenExistsDTO struct {
	UserUUID string
	RawToken string
}

type RevokeRefreshTokenDTO struct {
	UserUUID  string
	TokenHash string
}

type RevokeAllRefreshTokensDTO struct {
	UserUUID string
}

type RefreshTokenDTO struct {
	UserUUID     string
	OldHashToken string
	NewHashToken string
	LastIP       string    // IP последнего refresh — обновляется в хеше
	LastActiveAt time.Time // время последнего refresh — обновляется в хеше
}
