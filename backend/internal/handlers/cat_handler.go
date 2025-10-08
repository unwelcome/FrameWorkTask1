package handlers

import (
	"context"
	"github.com/unwelcome/iqjtest/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/unwelcome/iqjtest/internal/entities"
	"github.com/unwelcome/iqjtest/internal/services"
)

type CatHandler interface {
	CreateCat(c *fiber.Ctx) error
	GetCatByID(c *fiber.Ctx) error
	GetAllCats(c *fiber.Ctx) error
	UpdateCatName(c *fiber.Ctx) error
	UpdateCatAge(c *fiber.Ctx) error
	UpdateCatDescription(c *fiber.Ctx) error
	UpdateCat(c *fiber.Ctx) error
	DeleteCat(c *fiber.Ctx) error
}

type catHandlerImpl struct {
	catService         services.CatService
	requestTimeout     time.Duration
	fileRequestTimeout time.Duration
}

func NewCatHandler(catService services.CatService, requestTimeout, fileRequestTimeout time.Duration) CatHandler {
	return &catHandlerImpl{catService: catService, requestTimeout: requestTimeout, fileRequestTimeout: fileRequestTimeout}
}

// CreateCat
// @Summary Создание кота
// @Description Создает нового кота
// @Tags cat
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param name formData string true "Кличка кота"
// @Param age formData integer true "Возраст кота"
// @Param description formData string true "Описание кота"
// @Param files formData []file true "Файлы изображений"
// @Success 201 {object} entities.CatCreateResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/cat/create [post]
func (h *catHandlerImpl) CreateCat(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.fileRequestTimeout)
	defer cancel()

	// Парсим текстовые поля из formData
	fields := &entities.CatCreateRequestFields{}
	if err := c.BodyParser(fields); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing formData fields: " + err.Error()})
	}

	// Получаем файлы из multipart/formData
	files, err := utils.GetFilesFromFormData(c, "files", 20)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Создаем тело запроса
	createCatRequest := &entities.CatCreateRequestWithPhotos{
		Fields: fields,
		Photos: files,
	}

	userID := c.Locals("userID").(int)

	// Создаем кота
	createCatResponse, err := h.catService.CreateCat(ctx, userID, createCatRequest)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(createCatResponse)
}

// GetCatByID
// @Summary Получение кота по ID
// @Description Получение кота по ID
// @Tags cat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Cat ID"
// @Success 200 {object} entities.CatWithPhotos
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/cat/id/{id} [get]
func (h *catHandlerImpl) GetCatByID(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Получаем ID кота из параметров
	catID, err := utils.ValidateIntParams(c, "id", 1, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Получаем кота по ID
	cat, err := h.catService.GetCatByID(ctx, catID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(cat)
}

// GetAllCats
// @Summary Получение всех котов
// @Description Получение всех котов
// @Tags cat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} []entities.CatWithPrimePhoto
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/cat/all [get]
func (h *catHandlerImpl) GetAllCats(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Получаем всех котов
	cats, err := h.catService.GetAllCats(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(cats)
}

// UpdateCat
// @Summary Обновление клички, возраста и описания кота
// @Description Обновление клички, возраста и описания кота
// @Tags cat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Cat ID"
// @Param cat body entities.CatUpdateRequest true "Данные кота"
// @Success 200 {object} entities.CatUpdateResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/cat/mw/{id} [put]
func (h *catHandlerImpl) UpdateCat(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Парсим тело запроса в структуру
	catUpdateRequest := &entities.CatUpdateRequest{}
	if err := c.BodyParser(&catUpdateRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	catID := c.Locals("catID").(int)

	// Обновляем все данные кота
	catUpdateResponse, err := h.catService.UpdateCat(ctx, catID, catUpdateRequest)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(catUpdateResponse)
}

// UpdateCatName
// @Summary Обновление клички кота
// @Description Обновление клички кота
// @Tags cat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Cat ID"
// @Param cat body entities.CatUpdateNameRequest true "Данные кота"
// @Success 200 {object} entities.CatUpdateNameResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/cat/mw/{id}/name [patch]
func (h *catHandlerImpl) UpdateCatName(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Парсим тело запроса в структуру
	catUpdateNameRequest := &entities.CatUpdateNameRequest{}
	if err := c.BodyParser(&catUpdateNameRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	catID := c.Locals("catID").(int)

	// Обновляем кличку кота
	catUpdateNameResponse, err := h.catService.UpdateCatName(ctx, catID, catUpdateNameRequest)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(catUpdateNameResponse)
}

// UpdateCatAge
// @Summary Обновление возраста кота
// @Description Обновление возраста кота
// @Tags cat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Cat ID"
// @Param cat body entities.CatUpdateAgeRequest true "Данные кота"
// @Success 200 {object} entities.CatUpdateAgeResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/cat/mw/{id}/age [patch]
func (h *catHandlerImpl) UpdateCatAge(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Парсим тело запроса в структуру
	catUpdateAgeRequest := &entities.CatUpdateAgeRequest{}
	if err := c.BodyParser(&catUpdateAgeRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	catID := c.Locals("catID").(int)

	// Обновляем возраст кота
	catUpdateAgeResponse, err := h.catService.UpdateCatAge(ctx, catID, catUpdateAgeRequest)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(catUpdateAgeResponse)
}

// UpdateCatDescription
// @Summary Обновление описания кота
// @Description Обновление описания кота
// @Tags cat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Cat ID"
// @Param cat body entities.CatUpdateDescriptionRequest true "Данные кота"
// @Success 200 {object} entities.CatUpdateDescriptionResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/cat/mw/{id}/description [patch]
func (h *catHandlerImpl) UpdateCatDescription(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Парсим тело запроса в структуру
	catUpdateDescriptionRequest := &entities.CatUpdateDescriptionRequest{}
	if err := c.BodyParser(&catUpdateDescriptionRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	catID := c.Locals("catID").(int)

	// Обновляем описание кота
	catUpdateDescriptionResponse, err := h.catService.UpdateCatDescription(ctx, catID, catUpdateDescriptionRequest)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(catUpdateDescriptionResponse)
}

// DeleteCat
// @Summary Удаление кота
// @Description Удаление кота
// @Tags cat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Cat ID"
// @Success 200 {object} string
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/cat/mw/{id} [delete]
func (h *catHandlerImpl) DeleteCat(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	catID := c.Locals("catID").(int)

	// Удаляем кота
	err := h.catService.DeleteCat(ctx, catID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).SendString("Successfully deleted cat")
}
