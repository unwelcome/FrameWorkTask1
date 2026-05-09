package entities

type Application struct {
	ApplicationUUID     string `db:"uuid"`
	CompanyUUID         string `db:"company_uuid"`
	DepartmentUUID      string `db:"department_uuid"`
	Version             int    `db:"version"`
	Title               string `db:"title"`
	Description         string `db:"description"`
	Status              string `db:"status"`
	CreatedAt           string `db:"created_at"`
	CreatedBy           string `db:"created_by"`
	ClosedAt            string `db:"closed_at"`
	ResponsibleManager  string `db:"managed_by"`
	ResponsibleEngineer string `db:"executed_by"`
	DeletedBy           string `db:"deleted_by"`
}

type CreateApplicationDTO struct {
	ApplicationUUID string
	CompanyUUID     string
	DepartmentUUID  string
	Title           string
	Description     string
	CreatedBy       string
}

type GetApplicationDTO struct {
	ApplicationUUID string
}

type GetApplicationsDTO struct {
	CompanyUUID    string
	DepartmentUUID string
	Status         string
	Count          int
	Offset         int
}

type UpdateApplicationStatusDTO struct {
	ApplicationUUID      string
	Status               string
	InitiatorUUID        string
	TargetDepartmentUUID string
}

type AssignApplicationToEmployeeDTO struct {
	ApplicationUUID string
	InitiatorUUID   string
	TargetUUID      string
}

type DeleteApplicationDTO struct {
	ApplicationUUID string
	DeletedBy       string
}
