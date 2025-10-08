package entities

import "mime/multipart"

type Cat struct {
	ID          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Age         int    `json:"age" db:"age"`
	Description string `json:"description" db:"description"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	CreatedBy   int    `json:"created_by" db:"created_by"`
}

type CatWithPhotos struct {
	ID          int            `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`
	Age         int            `json:"age" db:"age"`
	Description string         `json:"description" db:"description"`
	CreatedAt   string         `json:"created_at" db:"created_at"`
	CreatedBy   int            `json:"created_by" db:"created_by"`
	Photos      []*CatPhotoUrl `json:"photos"`
}

type CatWithPrimePhoto struct {
	ID      int     `json:"id" db:"id"`
	Name    string  `json:"name" db:"name"`
	Age     int     `json:"age" db:"age"`
	PhotoID *int    `json:"photo_id" db:"photo_id"`
	Url     *string `json:"url" db:"url"`
}

type CatCreateRequestWithPhotos struct {
	Fields *CatCreateRequestFields
	Photos []*multipart.FileHeader
}

type CatCreateRequestFields struct {
	Name        string `form:"name" json:"name" db:"name"`
	Age         int    `form:"age" json:"age" db:"age"`
	Description string `form:"description" json:"description" db:"description"`
}

type CatCreateResponse struct {
	ID    int                     `json:"id" db:"id"`
	Photo *CatPhotoUploadResponse `json:"photo"`
}

type CatUpdateRequest struct {
	Name        string `json:"name" db:"name"`
	Age         int    `json:"age" db:"age"`
	Description string `json:"description" db:"description"`
}

type CatUpdateResponse struct {
	ID          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Age         int    `json:"age" db:"age"`
	Description string `json:"description" db:"description"`
}

type CatUpdateNameRequest struct {
	Name string `json:"name" db:"name"`
}

type CatUpdateNameResponse struct {
	ID int `json:"id" db:"id"`
}

type CatUpdateAgeRequest struct {
	Age int `json:"age" db:"age"`
}

type CatUpdateAgeResponse struct {
	ID int `json:"id" db:"id"`
}

type CatUpdateDescriptionRequest struct {
	Description string `json:"description" db:"description"`
}

type CatUpdateDescriptionResponse struct {
	ID int `json:"id" db:"id"`
}
