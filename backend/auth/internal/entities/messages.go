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
	Token     string `json:"token"`
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

type PasswordResetEmailMsg struct {
	UserUUID  string `json:"user_uuid"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
}

type RegistrationAttemptEmailMsg struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
}

type LoginNotificationEmailMsg struct {
	UserUUID  string `json:"user_uuid"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	IP        string `json:"ip"`
	Browser   string `json:"browser"`
	OS        string `json:"os"`
	LoginAt   int64  `json:"login_at"`
}
