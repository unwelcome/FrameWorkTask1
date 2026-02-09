package entities

type FixLog struct {
	ID              int    `db:"id"`
	ApplicationUUID string `db:"application_uuid"`
	Text            string `db:"text"`
	CreatedAt       string `db:"created_at"`
	CreatedBy       string `db:"created_by"`
}

type CreateFixLogDTO struct {
	ApplicationUUID string
	Text            string
	CreatedBy       string
}

type GetApplicationFixLogsDTO struct {
	ApplicationUUID string
}
