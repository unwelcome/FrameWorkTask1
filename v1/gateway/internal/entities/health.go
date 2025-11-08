package entities

type Health struct {
	Gateway     string `json:"gateway"`
	Auth        string `json:"auth_service"`
	Application string `json:"application_service"`
}
