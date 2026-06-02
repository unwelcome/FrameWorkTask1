package entities

type Employee struct {
	CompanyUUID    string `db:"company_uuid"`
	UserUUID       string `db:"user_uuid"`
	DepartmentUUID string `db:"department_uuid"`
	Role           string `db:"role"`
	JoinedAt       string `db:"joined_at"`
}

type EmployeesSummary struct {
	CompanyUUID     string `db:"company_uuid"`
	ChiefCount      int64
	AnalyticCount   int64
	ManagerCount    int64
	EngineerCount   int64
	InspectorCount  int64
	UnemployedCount int64
}

type JoinCompanyDTO struct {
	CompanyUUID string
	UserUUID    string
}

type GetCompanyEmployeeDTO struct {
	CompanyUUID string
	UserUUID    string
}

type GetCompanyEmployeesDTO struct {
	CompanyUUID    string
	DepartmentUUID string
	Role           string
	Offset         int64
	Count          int64
}

type GetCompanyEmployeesSummaryDTO struct {
	CompanyUUID    string
	DepartmentUUID string
}

type SetCompanyEmployeeRoleDTO struct {
	CompanyUUID string
	UserUUID    string
	Role        string
}

type RemoveCompanyEmployeeDTO struct {
	CompanyUUID string
	UserUUID    string
}
