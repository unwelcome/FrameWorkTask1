package entities

type Department struct {
	UUID        string `db:"uuid"`
	CompanyUUID string `db:"company_uuid"`
	Title       string `db:"title"`
	CreatedAt   string `db:"created_at"`
	CreatedBy   string `db:"created_by"`
}

type CreateDepartment struct {
	UUID        string `db:"uuid"`
	CompanyUUID string `db:"company_uuid"`
	Title       string `db:"title"`
	CreatedBy   string `db:"created_by"`
}

type UpdateDepartment struct {
	UUID  string `db:"uuid"`
	Title string `db:"title"`
}

type AddEmployeeToDepartmentDTO struct {
	DepartmentUUID string
	CompanyUUID    string
	TargetUUID     string
}

type GetDepartmentDTO struct {
	DepartmentUUID string
}

type GetCompanyDepartmentsDTO struct {
	CompanyUUID string
	Offset      int64
	Count       int64
}

type DeleteDepartmentDTO struct {
	DepartmentUUID string
}

type RemoveEmployeeFromDepartmentDTO struct {
	CompanyUUID string
	TargetUUID  string
}
