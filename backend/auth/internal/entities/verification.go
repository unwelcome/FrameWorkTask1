package entities

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

type CheckResendCooldownDTO struct {
	UserUUID string
}

type IncrResendDailyCountDTO struct {
	UserUUID string
}
