package entities

type FixLog struct {
	ID              int    `db:"id"`
	ApplicationUUID string `db:"application_uuid"`
	Text            string `db:"text"`
	CreatedAt       string `db:"created_at"`
	CreatedBy       string `db:"created_by"`
}

type CreateFixLog struct {
	ApplicationUUID string `db:"application_uuid"`
	Text            string `db:"text"`
	CreatedBy       string `db:"created_by"`
}
