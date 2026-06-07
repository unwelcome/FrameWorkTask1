package entities

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
