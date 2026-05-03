package entities

type Employee struct {
	CompanyUUID string `db:"company_uuid"`
	UserUUID    string `db:"user_uuid"`
	Role        string `db:"role"`
	JoinedAt    string `db:"joined_at"`
}

type EmployeesSummary struct {
	CompanyUUID     string `db:"company_uuid"`
	ChiefCount      int64
	AnalyticCount   int64
	ManagerCount    int64
	EngineerCount   int64
	UnemployedCount int64
}
