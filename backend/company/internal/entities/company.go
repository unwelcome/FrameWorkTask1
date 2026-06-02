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

type GetCompanyDTO struct {
	CompanyUUID string
}

type GetCompaniesDTO struct {
	Offset int64
	Count  int64
}

type GetUserCompaniesDTO struct {
	UserUUID string
}

type UpdateCompanyTitleDTO struct {
	CompanyUUID string
	Title       string
}

type UpdateCompanyStatusDTO struct {
	CompanyUUID string
	Status      string
}

type DeleteCompanyDTO struct {
	CompanyUUID string
}
