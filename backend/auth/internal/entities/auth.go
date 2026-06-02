package entities

type SaveRefreshTokenDTO struct {
	UserUUID string
	RawToken string
}

type GetAllRefreshTokensDTO struct {
	UserUUID string
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
	UserUUID    string
	OldRawToken string
	NewRawToken string
}
