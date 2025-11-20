package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	company_proto "github.com/unwelcome/FrameWorkTask1/v1/gateway/api/company"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/errors"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/pkg/utils"
)

type CompanyHandler interface {
	CreateCompany(c *fiber.Ctx) error
}

type companyHandler struct {
	CompanyServiceClient company_proto.CompanyServiceClient
	operationIDKey       string
	userUUIDKey          string
}

func NewCompanyHandler(CompanyServiceClient company_proto.CompanyServiceClient, operationIDKey string, userUUIDKey string) CompanyHandler {
	return &companyHandler{CompanyServiceClient: CompanyServiceClient, operationIDKey: operationIDKey, userUUIDKey: userUUIDKey}
}

// CreateCompany
//
//	@Summary      Create company
//	@Description  Create new company (user that created company became "chief" automatically)
//	@Tags         Company
//	@Accept 	  json
//	@Produce 	  json
//	@Security 	  ApiKeyAuth
//	@Param 		  data body entities.CreateCompanyRequest true "Данные компании"
//	@Success      201  {object}  entities.CreateCompanyResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      401  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /auth/company/create [post]
func (h *companyHandler) CreateCompany(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Парсинг тела запроса
	httpReq := &entities.CreateCompanyRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &company_proto.CreateCompanyRequest{
		OperationId: operationID,
		UserUuid:    utils.GetLocal[string](c, h.userUUIDKey),
		Title:       httpReq.Title,
	}
	// Запрос в company сервис
	res, err := h.CompanyServiceClient.CreateCompany(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.CreateCompanyResponse{
		CompanyUUID: res.GetCompanyUuid(),
	}

	return c.Status(fiber.StatusCreated).JSON(httpRes)
}
