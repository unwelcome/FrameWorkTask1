package entities

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type TokenPairWithUserUUID struct {
	UserUUID     string `db:"uuid"`
	AccessToken  string
	RefreshToken string
}
