package services

import (
	"context"
	"fmt"
	"github.com/unwelcome/iqjtest/pkg/utils"
	"mime/multipart"

	"github.com/unwelcome/iqjtest/internal/entities"
	"github.com/unwelcome/iqjtest/internal/repositories"
)

type CatPhotoService interface {
	AddCatPhoto(ctx context.Context, catID int, photos []*multipart.FileHeader) *entities.CatPhotoUploadResponse
	GetCatPhotoByID(ctx context.Context, photoID int) (*entities.CatPhoto, error)
	GetAllCatPhotos(ctx context.Context, catID int) ([]*entities.CatPhotoUrl, error)
	SetCatPhotoPrimary(ctx context.Context, catID int, photoID int) (*entities.CatPhotoSetPrimaryResponse, error)
	DeleteCatPhoto(ctx context.Context, catID, photoID int) error
	DeleteAllCatPhotos(ctx context.Context, catID int) error
}

type catPhotoServiceImpl struct {
	catPhotoRepository repositories.CatPhotoRepository
}

func NewCatPhotoService(catPhotoRepository repositories.CatPhotoRepository) CatPhotoService {
	return &catPhotoServiceImpl{catPhotoRepository: catPhotoRepository}
}

func (s *catPhotoServiceImpl) AddCatPhoto(ctx context.Context, catID int, photos []*multipart.FileHeader) *entities.CatPhotoUploadResponse {

	// Создаем массив загруженных фото и массив с ошибками загрузки
	var uploadedPhotos []*entities.CatPhotoUploadSuccess
	var errors []*entities.CatPhotoUploadError

	// Проходимся по каждому фото
	for _, file := range photos {

		// Проверяем размер файла
		err := utils.CheckFileSize(file, 50*1024*1024)
		if err != nil {
			errors = append(errors, &entities.CatPhotoUploadError{
				FileName: file.Filename,
				Error:    err.Error(),
			})
			continue
		}

		// Проверяем тип файла
		if !utils.IsImageFile(file) {
			errors = append(errors, &entities.CatPhotoUploadError{
				FileName: file.Filename,
				Error:    "file must be an image file (jpg, png, webp)",
			})
			continue
		}

		// Открываем файл
		fileReader, err := file.Open()
		if err != nil {
			errors = append(errors, &entities.CatPhotoUploadError{
				FileName: file.Filename,
				Error:    "failed to open file",
			})
			continue
		}
		defer fileReader.Close()

		// Загружаем фото
		success, err := s.catPhotoRepository.AddCatPhoto(ctx, catID, &entities.CatPhotoUploadRequest{
			File:     fileReader,
			FileName: file.Filename,
			FileSize: file.Size,
			MimeType: file.Header.Get("Content-Type"),
		})
		if err != nil {
			errors = append(errors, &entities.CatPhotoUploadError{
				FileName: file.Filename,
				Error:    fmt.Sprintf("add photo error: %w", err),
			})
			continue
		}
		uploadedPhotos = append(uploadedPhotos, success)
	}

	// Создаем отчет о загрузке фото
	catPhotoUploadResponse := &entities.CatPhotoUploadResponse{
		Message:        fmt.Sprintf("Uploaded %d out of %d photos", len(uploadedPhotos), len(photos)),
		UploadedCount:  len(uploadedPhotos),
		FailedCount:    len(errors),
		UploadedPhotos: uploadedPhotos,
		Errors:         errors,
	}

	return catPhotoUploadResponse
}

func (s *catPhotoServiceImpl) GetCatPhotoByID(ctx context.Context, photoID int) (*entities.CatPhoto, error) {

	// Получаем фото по ID
	catPhoto, err := s.catPhotoRepository.GetCatPhotoByID(ctx, photoID)
	if err != nil {
		return nil, fmt.Errorf("get cat photo by id error: %w", err)
	}

	return catPhoto, nil
}

func (s *catPhotoServiceImpl) GetAllCatPhotos(ctx context.Context, catID int) ([]*entities.CatPhotoUrl, error) {

	// Получаем все фото кота
	catPhotosUrl, err := s.catPhotoRepository.GetAllCatPhotos(ctx, catID)
	if err != nil {
		return nil, fmt.Errorf("get all cat photo error: %w", err)
	}

	return catPhotosUrl, nil
}

func (s *catPhotoServiceImpl) SetCatPhotoPrimary(ctx context.Context, catID int, photoID int) (*entities.CatPhotoSetPrimaryResponse, error) {

	// Устанавливаем главное фото кота
	err := s.catPhotoRepository.SetCatPhotoPrimary(ctx, catID, photoID)
	if err != nil {
		return nil, fmt.Errorf("set cat photo primary error: %w", err)
	}

	return &entities.CatPhotoSetPrimaryResponse{ID: photoID}, nil
}

func (s *catPhotoServiceImpl) DeleteCatPhoto(ctx context.Context, catID, photoID int) error {

	// Получаем информацию о фото
	catPhoto, err := s.catPhotoRepository.GetCatPhotoByID(ctx, photoID)
	if err != nil {
		return fmt.Errorf("delete cat photo error: %w", err)
	}

	// Проверяем, что фото принадлежит коту
	if catPhoto.CatID != catID {
		return fmt.Errorf("delete cat photo error: photo %d doesn't belong to cat %d", photoID, catID)
	}

	// Удаляем фото
	err = s.catPhotoRepository.DeleteCatPhoto(ctx, photoID)
	if err != nil {
		return fmt.Errorf("delete cat photo error: %w", err)
	}

	return nil
}

func (s *catPhotoServiceImpl) DeleteAllCatPhotos(ctx context.Context, catID int) error {

	// Удаляем все фото кота
	err := s.catPhotoRepository.DeleteAllCatPhotos(ctx, catID)
	if err != nil {
		return fmt.Errorf("delete all cat photos error: %w", err)
	}

	return nil
}
