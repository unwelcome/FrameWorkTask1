package entities

type ServiceHealth struct {
	Service  string `json:"service"`
	Postgres string `json:"postgres,omitempty"`
	Redis    string `json:"redis,omitempty"`
	Minio    string `json:"minio,omitempty"`
	Mongo    string `json:"mongo,omitempty"`
}

type HealthResponse struct {
	Gateway     string        `json:"gateway"`
	Auth        ServiceHealth `json:"auth_service"`
	Company     ServiceHealth `json:"company_service"`
	Application ServiceHealth `json:"application_service"`
}
