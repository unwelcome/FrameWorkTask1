package handlers

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/unwelcome/iqjtest/internal/services"
	"github.com/unwelcome/iqjtest/pkg/utils"
	"time"
)

type CatPhotoHandler interface {
	AddCatPhotos(c *fiber.Ctx) error
	GetCatPhotoByID(c *fiber.Ctx) error
	SetCatPhotoPrimary(c *fiber.Ctx) error
	DeleteCatPhoto(c *fiber.Ctx) error
}

type catPhotoHandlerImpl struct {
	catPhotoService    services.CatPhotoService
	requestTimeout     time.Duration
	fileRequestTimeout time.Duration
}

func NewCatPhotoHandler(catPhotoService services.CatPhotoService, requestTimeout, fileRequestTimeout time.Duration) CatPhotoHandler {
	return &catPhotoHandlerImpl{catPhotoService: catPhotoService, requestTimeout: requestTimeout, fileRequestTimeout: fileRequestTimeout}
}

// AddCatPhotos
// @Summary Загрузить фото кота
// @Description Загружает фотографию для указанного кота (не более 20 файлов)
// @Tags cat-photo
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Cat ID"
// @Param files formData []file true "Файлы изображений"
// @Success 201 {object} entities.CatPhotoUploadResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.CatPhotoUploadResponse
// @Router /auth/cat/mw/{id}/photo/add [post]
func (h *catPhotoHandlerImpl) AddCatPhotos(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.fileRequestTimeout)
	defer cancel()

	// Получаем файлы из multipart/formData
	files, err := utils.GetFilesFromFormData(c, "files", 20)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	catID := c.Locals("catID").(int)

	// Загружаем фото
	catPhotoUploadResponse := h.catPhotoService.AddCatPhoto(ctx, catID, files)

	// Отправляем результат
	if catPhotoUploadResponse.UploadedCount > 0 {
		return c.Status(fiber.StatusCreated).JSON(catPhotoUploadResponse)
	}
	return c.Status(fiber.StatusInternalServerError).JSON(catPhotoUploadResponse)
}

// GetCatPhotoByID
// @Summary Получение фото кота
// @Description Получение всей информации о фото кота
// @Tags cat-photo
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param photoID path int true "Photo ID"
// @Success 200 {object} entities.CatPhoto
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/cat/photo/{photoID} [get]
func (h *catPhotoHandlerImpl) GetCatPhotoByID(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Получаем ID фото из параметров
	photoID, err := utils.ValidateIntParams(c, "photoID", 1, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Получаем информацию о фото
	catPhoto, err := h.catPhotoService.GetCatPhotoByID(ctx, photoID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(catPhoto)
}

// SetCatPhotoPrimary
// @Summary Выбор основного фото кота
// @Description Выбор основного фото кота
// @Tags cat-photo
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Cat ID"
// @Param photoID path int true "Photo ID"
// @Success 200 {object} entities.CatPhotoSetPrimaryResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/cat/mw/{id}/photo/{photoID}/primary [patch]
func (h *catPhotoHandlerImpl) SetCatPhotoPrimary(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Получаем ID фото из параметров
	photoID, err := utils.ValidateIntParams(c, "photoID", 1, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	catID := c.Locals("catID").(int)

	// Устанавливаем главное фото
	res, err := h.catPhotoService.SetCatPhotoPrimary(ctx, catID, photoID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

// DeleteCatPhoto
// @Summary Удаление фото кота по ID
// @Description Удаление фото кота по ID
// @Tags cat-photo
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Cat ID"
// @Param photoID path int true "Photo ID"
// @Success 200 {object} string
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/cat/mw/{id}/photo/{photoID} [delete]
func (h *catPhotoHandlerImpl) DeleteCatPhoto(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Получаем ID фото из параметров
	photoID, err := utils.ValidateIntParams(c, "photoID", 1, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	catID := c.Locals("catID").(int)

	// Удаляем фото
	err = h.catPhotoService.DeleteCatPhoto(ctx, catID, photoID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).SendString("successfully deleted photo")
}
