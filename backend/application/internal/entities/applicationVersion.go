package entities

type ApplicationVersion struct {
	ApplicationUUID     string `bson:"application_uuid"`
	Version             int    `bson:"version"`
	CompanyUUID         string `bson:"company_uuid"`
	DepartmentUUID      string `bson:"department_uuid"`
	Title               string `bson:"title"`
	Description         string `bson:"description"`
	Status              string `bson:"status"`
	ResponsibleManager  string `bson:"responsible_manager"`
	ResponsibleEngineer string `bson:"responsible_engineer"`
	CreatedAt           string `bson:"created_at"`
	CreatedBy           string `bson:"created_by"`
	ClosedAt            string `bson:"closed_at"`
	SavedAt             string `bson:"saved_at"`
}
