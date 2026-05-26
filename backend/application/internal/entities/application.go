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
	CompanyUUID      string
	DepartmentUUID   string
	CreatedBy        string // Если указано - созданные заявки инспектора
	ManagedBy        string // Если указано - личные заявки менеджера
	ExecutedBy       string // Если указано - личные заявки инженера
	InspectedBy      string // Если указано - личные заявки инспектора
	ExecutedByIsNull bool   // При запросе заявок из пула менеджеров включаем заявки с on_revision и executed_by = null
	Statuses         []string
	Count            int
	Offset           int
	IsDeleted        bool
}

type UpdateApplicationStatusDTO struct {
	ApplicationUUID string
	InitiatorUUID   string
	Status          string
	DropManagedBy   bool
	DropExecutedBy  bool
}

type AssignApplicationDTO struct {
	ApplicationUUID string
	InitiatorUUID   string
	TargetUUID      string
}

type RedirectApplicationDTO struct {
	ApplicationUUID      string
	InitiatorUUID        string
	TargetDepartmentUUID string
	FixLogText           string
}

type RecallApplicationDTO struct {
	ApplicationUUID string
	InitiatorUUID   string
	FixLogText      string
}

type TakeApplicationToVerificationDTO struct {
	ApplicationUUID string
	InitiatorUUID   string
}

type ReleaseApplicationVerificationDTO struct {
	ApplicationUUID string
	InitiatorUUID   string
	FixLogText      string
}

type DeleteApplicationDTO struct {
	ApplicationUUID string
	DeletedBy       string
	FixLogText      string
}
