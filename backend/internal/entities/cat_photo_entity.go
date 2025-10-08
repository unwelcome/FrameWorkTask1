package entities

import "io"

type CatPhoto struct {
	ID        int    `json:"id" db:"id"`
	Url       string `json:"url" db:"url"`
	CatID     int    `json:"cat_id" db:"cat_id"`
	FileName  string `json:"file_name" db:"file_name"`
	FileSize  int    `json:"file_size" db:"file_size"`
	MimeType  string `json:"mime_type" db:"mime_type"`
	IsPrimary bool   `json:"is_primary" db:"is_primary"`
	CreatedAt string `json:"created_at" db:"created_at"`
}

type CatPhotoUrl struct {
	ID        int    `json:"id" db:"id"`
	Url       string `json:"url" db:"url"`
	IsPrimary bool   `json:"is_primary" db:"is_primary"`
}

type CatPhotoUploadRequest struct {
	File     io.Reader
	FileSize int64  `json:"file_size" db:"file_size"`
	FileName string `json:"file_name" db:"file_name"`
	MimeType string `json:"mime_type" db:"mime_type"`
}

type CatPhotoUploadResponse struct {
	Message        string                   `json:"message"`
	UploadedCount  int                      `json:"uploaded_count"`
	FailedCount    int                      `json:"failed_count"`
	UploadedPhotos []*CatPhotoUploadSuccess `json:"uploaded_photos"`
	Errors         []*CatPhotoUploadError   `json:"errors"`
}

type CatPhotoUploadSuccess struct {
	ID       int    `json:"id" db:"id"`
	Url      string `json:"url" db:"url"`
	FileName string `json:"file_name" db:"file_name"`
}
type CatPhotoUploadError struct {
	FileName string `json:"file_name" db:"file_name"`
	Error    string `json:"error" db:"error"`
}

type CatPhotoSetPrimaryResponse struct {
	ID int `json:"id" db:"id"`
}
