package entities

type FixLog struct {
	UUID            string `db:"uuid"`
	ApplicationUUID string `db:"application_uuid"`
	Text            string `db:"text"`
	CreatedAt       string `db:"created_at"`
	CreatedBy       string `db:"created_by"`
}

type AddFixLogDTO struct {
	ApplicationUUID string
	Text            string
	CreatedBy       string
}

type GetApplicationFixLogsDTO struct {
	ApplicationUUID string
}
