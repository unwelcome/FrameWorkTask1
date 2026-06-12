package entities

type VerificationEmailMsg struct {
	UserUUID  string `json:"user_uuid"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	Code      string `json:"code"`
}

type RecoveryEmailMsg struct {
	UserUUID  string `json:"user_uuid"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	Code      string `json:"code"`
}

type TwoFAEmailMsg struct {
	UserUUID  string `json:"user_uuid"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	Code      string `json:"code"`
}

type PasswordChangedEmailMsg struct {
	UserUUID  string `json:"user_uuid"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
}

type RegistrationAttemptEmailMsg struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
}
