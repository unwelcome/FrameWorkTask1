package entities

type Application struct {
	ApplicationUUID     string `db:"uuid"`
	Title               string `db:"title"`
	Description         string `db:"description"`
	Status              string `db:"status"`
	CreatedAt           string `db:"created_at"`
	CreatedBy           string `db:"created_by"`
	ClosedAt            string `db:"closed_at"`
	ResponsibleManager  string `db:"managed_by"`
	ResponsibleEngineer string `db:"executed_by"`
}

type CreateApplication struct {
	ApplicationUUID string `db:"uuid"`
	Title           string `db:"title"`
	Description     string `db:"description"`
	CreatedBy       string `db:"created_by"`
}
