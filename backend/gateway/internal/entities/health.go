package entities

type HealthResponse struct {
	Gateway     string `json:"gateway"`
	Auth        string `json:"auth_service"`
	Company     string `json:"company_service"`
	Application string `json:"application_service"`
}
