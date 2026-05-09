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
