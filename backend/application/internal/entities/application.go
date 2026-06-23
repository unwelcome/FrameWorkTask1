package entities

type Application struct {
	ApplicationUUID string `db:"uuid" json:"uuid"`
	CompanyUUID     string `db:"company_uuid" json:"company_uuid"`
	DepartmentUUID  string `db:"department_uuid" json:"department_uuid"`
	Version         int64  `db:"version" json:"version"`
	Title           string `db:"title" json:"title"`
	Description     string `db:"description" json:"description"`
	Status          string `db:"status" json:"status"`
	RevisionCount   int64  `db:"revision_count" json:"revision_count"`
	CreatedAt       string `db:"created_at" json:"created_at"`
	CreatedBy       string `db:"created_by" json:"created_by"`
	UpdatedAt       string `db:"updated_at" json:"updated_at"`
	UpdatedBy       string `db:"updated_by" json:"updated_by"`
	ManagedBy       string `db:"managed_by" json:"managed_by"`
	ExecutedBy      string `db:"executed_by" json:"executed_by"`
	InspectedBy     string `db:"inspected_by" json:"inspected_by"`
	ClosedAt        string `db:"closed_at" json:"closed_at"`
	DeletedAt       string `db:"deleted_at" json:"deleted_at"`
	DeletedBy       string `db:"deleted_by" json:"deleted_by"`
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
	Count            int64
	Offset           int64
	IsDeleted        bool
}

type UpdateApplicationStatusDTO struct {
	ApplicationUUID string
	InitiatorUUID   string
	Status          string
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

type GetApplicationHistoryDTO struct {
	ApplicationUUID string
	Offset          int64
	Count           int64
}
