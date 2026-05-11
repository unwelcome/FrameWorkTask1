package entities

type Application struct {
	ApplicationUUID string `db:"uuid"`
	CompanyUUID     string `db:"company_uuid"`
	DepartmentUUID  string `db:"department_uuid"`
	Version         int    `db:"version"`
	Title           string `db:"title"`
	Description     string `db:"description"`
	Status          string `db:"status"`
	RevisionCount   int    `db:"revision_count"`
	CreatedAt       string `db:"created_at"`
	CreatedBy       string `db:"created_by"`
	UpdatedAt       string `db:"updated_at"`
	UpdatedBy       string `db:"updated_by"`
	ManagedBy       string `db:"managed_by"`
	ExecutedBy      string `db:"executed_by"`
	InspectedBy     string `db:"inspected_by"`
	ClosedAt        string `db:"closed_at"`
	DeletedAt       string `db:"deleted_at"`
	DeletedBy       string `db:"deleted_by"`
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
	IsDeleted      bool
}

type UpdateApplicationStatusDTO struct {
	ApplicationUUID string
	InitiatorUUID   string
	Status          string
}

type AssignApplicationToEmployeeDTO struct {
	ApplicationUUID string
	InitiatorUUID   string
	TargetUUID      string
}

type RedirectApplicationDTO struct {
	ApplicationUUID      string
	InitiatorUUID        string
	TargetDepartmentUUID string
}

type RecallApplicationDTO struct {
	ApplicationUUID string
	InitiatorUUID   string
}

type TakeApplicationToVerificationDTO struct {
	ApplicationUUID string
	InitiatorUUID   string
}

type ReleaseApplicationVerificationDTO struct {
	ApplicationUUID string
	InitiatorUUID   string
}

type DeleteApplicationDTO struct {
	ApplicationUUID string
	DeletedBy       string
}
