package services

import (
	"context"
	"fmt"
	"github.com/unwelcome/iqjtest/internal/entities"
	"github.com/unwelcome/iqjtest/internal/repositories"
)

type CatService interface {
	CreateCat(ctx context.Context, userID int, catCreateRequest *entities.CatCreateRequestWithPhotos) (*entities.CatCreateResponse, error)
	GetCatByID(ctx context.Context, catID int) (*entities.CatWithPhotos, error)
	GetAllCats(ctx context.Context) ([]*entities.CatWithPrimePhoto, error)
	CheckOwnershipRight(ctx context.Context, userID, catID int) (bool, error)
	UpdateCatName(ctx context.Context, catID int, catUpdateNameRequest *entities.CatUpdateNameRequest) (*entities.CatUpdateNameResponse, error)
	UpdateCatAge(ctx context.Context, catID int, catUpdateAgeRequest *entities.CatUpdateAgeRequest) (*entities.CatUpdateAgeResponse, error)
	UpdateCatDescription(ctx context.Context, catID int, catUpdateDescriptionRequest *entities.CatUpdateDescriptionRequest) (*entities.CatUpdateDescriptionResponse, error)
	UpdateCat(ctx context.Context, catID int, catUpdateRequest *entities.CatUpdateRequest) (*entities.CatUpdateResponse, error)
	DeleteCat(ctx context.Context, catID int) error
}

type catServiceImpl struct {
	catRepository   repositories.CatRepository
	catPhotoService CatPhotoService
}

func NewCatService(catRepository repositories.CatRepository, catPhotoService CatPhotoService) CatService {
	return &catServiceImpl{catRepository: catRepository, catPhotoService: catPhotoService}
}

func (s *catServiceImpl) CreateCat(ctx context.Context, userID int, catCreateRequest *entities.CatCreateRequestWithPhotos) (*entities.CatCreateResponse, error) {
	// Создаем кота
	cat := &entities.Cat{
		Name:        catCreateRequest.Fields.Name,
		Age:         catCreateRequest.Fields.Age,
		Description: catCreateRequest.Fields.Description,
	}

	// Добавляем кота в бд и получаем его ID
	err := s.catRepository.CreateCat(ctx, userID, cat)
	if err != nil {
		return nil, fmt.Errorf("create cat error: %s", err.Error())
	}

	// Добавляем фото кота в S3
	catPhotoUploadResponse := s.catPhotoService.AddCatPhoto(ctx, cat.ID, catCreateRequest.Photos)

	return &entities.CatCreateResponse{ID: cat.ID, Photo: catPhotoUploadResponse}, nil
}

func (s *catServiceImpl) GetCatByID(ctx context.Context, catID int) (*entities.CatWithPhotos, error) {
	// Получаем данные кота
	cat, err := s.catRepository.GetCatByID(ctx, catID)
	if err != nil {
		return nil, fmt.Errorf("get cat by id error: %w", err)
	}

	// Получаем все фото кота
	catPhotos, err := s.catPhotoService.GetAllCatPhotos(ctx, catID)
	if err != nil {
		return nil, err
	}

	// Подготавливаем тело ответа
	catWithPhotos := &entities.CatWithPhotos{
		ID:          catID,
		Name:        cat.Name,
		Age:         cat.Age,
		Description: cat.Description,
		CreatedBy:   cat.CreatedBy,
		CreatedAt:   cat.CreatedAt,
		Photos:      catPhotos,
	}

	return catWithPhotos, nil
}

func (s *catServiceImpl) GetAllCats(ctx context.Context) ([]*entities.CatWithPrimePhoto, error) {

	// Получаем всех котов
	cats, err := s.catRepository.GetAllCats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all cats error: %s", err.Error())
	}

	return cats, nil
}

func (s *catServiceImpl) CheckOwnershipRight(ctx context.Context, userID, catID int) (bool, error) {

	// Получаем кота по ID
	cat, err := s.catRepository.GetCatByID(ctx, catID)
	if err != nil {
		return false, fmt.Errorf("get cat info error: %s", err.Error())
	}

	// Проверяем владельца
	return cat.CreatedBy == userID, nil
}

func (s *catServiceImpl) UpdateCatName(ctx context.Context, catID int, catUpdateNameRequest *entities.CatUpdateNameRequest) (*entities.CatUpdateNameResponse, error) {

	// Обновляем кличку кота
	err := s.catRepository.UpdateCatName(ctx, catID, catUpdateNameRequest.Name)
	if err != nil {
		return nil, fmt.Errorf("update cat name error: %s", err.Error())
	}

	return &entities.CatUpdateNameResponse{ID: catID}, nil
}

func (s *catServiceImpl) UpdateCatAge(ctx context.Context, catID int, catUpdateAgeRequest *entities.CatUpdateAgeRequest) (*entities.CatUpdateAgeResponse, error) {

	// Обновляем возраст кота
	err := s.catRepository.UpdateCatAge(ctx, catID, catUpdateAgeRequest.Age)
	if err != nil {
		return nil, fmt.Errorf("update cat age error: %s", err.Error())
	}

	return &entities.CatUpdateAgeResponse{ID: catID}, nil
}

func (s *catServiceImpl) UpdateCatDescription(ctx context.Context, catID int, catUpdateDescriptionRequest *entities.CatUpdateDescriptionRequest) (*entities.CatUpdateDescriptionResponse, error) {

	// Обновляем описание кота
	err := s.catRepository.UpdateCatDescription(ctx, catID, catUpdateDescriptionRequest.Description)
	if err != nil {
		return nil, fmt.Errorf("update cat description error: %s", err.Error())
	}

	return &entities.CatUpdateDescriptionResponse{ID: catID}, nil
}

func (s *catServiceImpl) UpdateCat(ctx context.Context, catID int, catUpdateRequest *entities.CatUpdateRequest) (*entities.CatUpdateResponse, error) {

	// Обновляем все поля кота
	err := s.catRepository.UpdateCat(ctx, catID, catUpdateRequest)
	if err != nil {
		return nil, fmt.Errorf("update cat error: %s", err.Error())
	}

	return &entities.CatUpdateResponse{
		ID:          catID,
		Name:        catUpdateRequest.Name,
		Age:         catUpdateRequest.Age,
		Description: catUpdateRequest.Description,
	}, nil
}

func (s *catServiceImpl) DeleteCat(ctx context.Context, catID int) error {

	// Удаляем все фото кота
	err := s.catPhotoService.DeleteAllCatPhotos(ctx, catID)
	if err != nil {
		return err
	}

	// Удаляем кота
	err = s.catRepository.DeleteCat(ctx, catID)
	if err != nil {
		return fmt.Errorf("delete cat error: %s", err.Error())
	}

	return nil
}
