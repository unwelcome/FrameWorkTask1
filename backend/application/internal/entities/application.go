package entities

type Application struct {
	ApplicationUUID     string `db:"uuid"`
	CompanyUUID         string `db:"company_uuid"`
	Version             int    `db:"version"`
	Title               string `db:"title"`
	Description         string `db:"description"`
	Status              string `db:"status"`
	CreatedAt           string `db:"created_at"`
	CreatedBy           string `db:"created_by"`
	ClosedAt            string `db:"closed_at"`
	ResponsibleManager  string `db:"managed_by"`
	ResponsibleEngineer string `db:"executed_by"`
}

type CreateApplicationDTO struct {
	ApplicationUUID string
	CompanyUUID     string
	Title           string
	Description     string
	CreatedBy       string
}

type GetApplicationDTO struct {
	ApplicationUUID string
}

type GetApplicationsDTO struct {
	CompanyUUID string
	Status      string
	Count       int
	Offset      int
}

type UpdateApplicationStatusDTO struct {
	ApplicationUUID string
	Status          string
	InitiatorUUID   string
}

type AssignApplicationToEmployeeDTO struct {
	ApplicationUUID string
	InitiatorUUID   string
	TargetUUID      string
}

type UpdateApplicationDataDTO struct {
	ApplicationUUID string
	Title           *string
	Desctiption     *string // typo preserved from repo
}

type DeleteApplicationDTO struct {
	ApplicationUUID string
}
