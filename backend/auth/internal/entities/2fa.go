package entities

type Save2FADataDTO struct {
	SessionUUID string
	UserUUID    string
	Code        string
}

type Get2FADataDTO struct {
	SessionUUID string
}

type Delete2FADataDTO struct {
	SessionUUID string
}

type Incr2FAAttemptsDTO struct {
	SessionUUID string
}

type TwoFAData struct {
	UserUUID string `json:"user_uuid"`
	Code     string `json:"code"`
}
