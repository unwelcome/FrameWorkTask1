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

type SaveVerificationCodeDTO struct {
	UserUUID string
	Code     string
}

type GetVerificationCodeDTO struct {
	UserUUID string
}

type DeleteVerificationCodeDTO struct {
	UserUUID string
}

type IncrVerificationAttemptsDTO struct {
	UserUUID string
}

type SaveRecoveryCodeDTO struct {
	UserUUID string
	Code     string
}

type GetRecoveryCodeDTO struct {
	UserUUID string
}

type DeleteRecoveryCodeDTO struct {
	UserUUID string
}

type IncrRecoveryAttemptsDTO struct {
	UserUUID string
}
