package entities

type Save2FADataDTO struct {
	SessionUUID string
	UserUUID    string
	Email       string
	FirstName   string
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

type Acquire2FAEmailCooldownDTO struct {
	UserUUID string
}

type Incr2FAEmailDailyCountDTO struct {
	UserUUID string
}

type TwoFAData struct {
	UserUUID  string `json:"user_uuid"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	Code      string `json:"code"`
}
