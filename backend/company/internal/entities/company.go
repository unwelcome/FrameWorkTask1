package entities

type Company struct {
	CompanyUUID string `db:"uuid"`
	Title       string `db:"title"`
	Status      string `db:"status"`
	CreatedAt   string `db:"created_at"`
	CreatedBy   string `db:"created_by"`
}

type CreateCompany struct {
	CompanyUUID string `db:"uuid"`
	Title       string `db:"title"`
	CreatedBy   string `db:"created_by"`
}

type GetCompanies struct {
	CompanyUUID string `db:"uuid"`
	Title       string `db:"title"`
	Status      string `db:"status"`
}
