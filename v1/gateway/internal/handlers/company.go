package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	company_proto "github.com/unwelcome/FrameWorkTask1/v1/gateway/api/company"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/errors"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/pkg/utils"
)

type CompanyHandler interface {
	CreateCompany(c *fiber.Ctx) error
	GetCompany(c *fiber.Ctx) error
	GetCompanies(c *fiber.Ctx) error
	UpdateCompanyTitle(c *fiber.Ctx) error
	UpdateCompanyStatus(c *fiber.Ctx) error
	DeleteCompany(c *fiber.Ctx) error
	CreateCompanyJoinCode(c *fiber.Ctx) error
	GetCompanyJoinCodes(c *fiber.Ctx) error
	DeleteCompanyJoinCode(c *fiber.Ctx) error
	JoinCompany(c *fiber.Ctx) error
	GetCompanyEmployee(c *fiber.Ctx) error
	GetCompanyEmployeesSummary(c *fiber.Ctx) error
	RemoveCompanyEmployee(c *fiber.Ctx) error
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
//	@Summary			Create company
//	@Description		Create new company (user that created company became "chief" automatically)
//	@Tags				Company
//	@Accept				json
//	@Produce			json
//	@Security			ApiKeyAuth
//	@Param				data body entities.CreateCompanyRequest true "Данные компании"
//	@Success			201  {object}  entities.CreateCompanyResponse
//	@Failure			400  {object}  Error.HttpError
//	@Failure			401  {object}  Error.HttpError
//	@Failure			500  {object}  Error.HttpError
//	@Router				/auth/company/create [post]
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

// GetCompany
//
//	@Summary			Get company info
//	@Description	Get company info by company uuid
//	@Tags					Company
//	@Accept				json
//	@Produce			json
//	@Security			ApiKeyAuth
//	@Param 				company_uuid path string true "Company UUID"
//	@Success			200  {object}  entities.GetCompanyResponse
//	@Failure			400  {object}  Error.HttpError
//	@Failure			401  {object}  Error.HttpError
//	@Failure			404  {object}  Error.HttpError
//	@Failure			500  {object}  Error.HttpError
//	@Router				/auth/company/{company_uuid} [get]
func (h *companyHandler) GetCompany(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Получаем CompanyUUID из параметров
	httpReq := &entities.GetCompanyRequest{
		CompanyUUID: c.Params("company_uuid", ""),
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &company_proto.GetCompanyRequest{
		OperationId: operationID,
		UserUuid:    utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid: httpReq.CompanyUUID,
	}
	// Запрос в company сервис
	res, err := h.CompanyServiceClient.GetCompany(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.GetCompanyResponse{
		CompanyUUID: res.GetCompanyUuid(),
		Title:       res.GetTitle(),
		Status:      res.GetStatus(),
	}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// GetCompanies
//
//	@Summary			Get companies list
//	@Description		Get companies list
//	@Tags				Company
//	@Produce			json
//	@Security			ApiKeyAuth
//	@Param				offset	query int false "Offset"	default(0)
//	@Param				count	query int false "Count"		default(10)
//	@Success			200  {object}  entities.GetCompaniesResponse
//	@Failure			400  {object}  Error.HttpError
//	@Failure			401  {object}  Error.HttpError
//	@Failure			500  {object}  Error.HttpError
//	@Router				/auth/company/list [get]
func (h *companyHandler) GetCompanies(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	fmt.Println("1")
	// Парсинг тела запроса
	httpReq := &entities.GetCompaniesRequest{}
	if err := c.QueryParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	fmt.Println("2")
	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	fmt.Println("3")
	// Формируем тело запроса
	req := &company_proto.GetCompaniesRequest{
		OperationId: operationID,
		Offset:      httpReq.Offset,
		Count:       httpReq.Count,
	}

	fmt.Println("4")
	// Запрос в company сервис
	res, err := h.CompanyServiceClient.GetCompanies(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	fmt.Println("5")
	// Маппинг ответа
	companies := make([]*entities.GetCompanyResponse, 0)

	for _, company := range res.GetCompanies() {
		companies = append(companies, &entities.GetCompanyResponse{
			CompanyUUID: company.CompanyUuid,
			Title:       company.Title,
			Status:      company.Status,
		})
	}

	fmt.Println("6")
	// Формируем тело ответа
	httpRes := &entities.GetCompaniesResponse{
		Companies: companies,
	}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// UpdateCompanyTitle
//
//	@Summary			Update company title
//	@Description	Update company title by company uuid (chief only)
//	@Tags					Company
//	@Accept				json
//	@Produce			json
//	@Security			ApiKeyAuth
//	@Param 				company_uuid path string true "Company UUID"
//	@Param				data body entities.UpdateCompanyTitleRequest true "Параметры запроса"
//	@Success			200  {object}  entities.UpdateCompanyTitleResponse
//	@Failure			400  {object}  Error.HttpError
//	@Failure			401  {object}  Error.HttpError
//	@Failure			403  {object}  Error.HttpError
//	@Failure			404  {object}  Error.HttpError
//	@Failure			500  {object}  Error.HttpError
//	@Router				/auth/company/{company_uuid}/title [patch]
func (h *companyHandler) UpdateCompanyTitle(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Парсинг тела запроса
	httpReq := &entities.UpdateCompanyTitleRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	httpReqFull := &entities.UpdateCompanyTitleRequestFull{
		CompanyUUID: c.Params("company_uuid", ""),
		Title:       httpReq.Title,
	}

	// Валидация
	err := httpReqFull.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &company_proto.UpdateCompanyTitleRequest{
		OperationId: operationID,
		UserUuid:    utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid: httpReqFull.CompanyUUID,
		Title:       httpReqFull.Title,
	}

	// Запрос в company сервис
	_, err = h.CompanyServiceClient.UpdateCompanyTitle(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.UpdateCompanyTitleResponse{}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// UpdateCompanyStatus
//
//	@Summary			Update company status
//	@Description	Update company status by company uuid (chief only). Available statuses: "unemployed", "engineer", "manager", "analytic", "chief"
//	@Tags					Company
//	@Accept				json
//	@Produce			json
//	@Security			ApiKeyAuth
//	@Param 				company_uuid path string true "Company UUID"
//	@Param				data body entities.UpdateCompanyStatusRequest true "Параметры запроса"
//	@Success			200  {object}  entities.UpdateCompanyStatusResponse
//	@Failure			400  {object}  Error.HttpError
//	@Failure			401  {object}  Error.HttpError
//	@Failure			403  {object}  Error.HttpError
//	@Failure			404  {object}  Error.HttpError
//	@Failure			500  {object}  Error.HttpError
//	@Router				/auth/company/{company_uuid}/status [patch]
func (h *companyHandler) UpdateCompanyStatus(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Парсинг тела запроса
	httpReq := &entities.UpdateCompanyStatusRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	httpReqFull := &entities.UpdateCompanyStatusRequestFull{
		CompanyUUID: c.Params("company_uuid", ""),
		Status:      httpReq.Status,
	}

	// Валидация
	err := httpReqFull.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &company_proto.UpdateCompanyStatusRequest{
		OperationId: operationID,
		UserUuid:    utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid: httpReqFull.CompanyUUID,
		Status:      httpReqFull.Status,
	}

	// Запрос в company сервис
	_, err = h.CompanyServiceClient.UpdateCompanyStatus(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.UpdateCompanyStatusResponse{}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// DeleteCompany
//
//	@Summary			Delete company
//	@Description	Delete company by company uuid (chief only)
//	@Tags					Company
//	@Accept				json
//	@Produce			json
//	@Security			ApiKeyAuth
//	@Param 				company_uuid path string true "Company UUID"
//	@Success			200  {object}  entities.DeleteCompanyResponse
//	@Failure			400  {object}  Error.HttpError
//	@Failure			401  {object}  Error.HttpError
//	@Failure			403  {object}  Error.HttpError
//	@Failure			404  {object}  Error.HttpError
//	@Failure			500  {object}  Error.HttpError
//	@Router				/auth/company/{company_uuid} [delete]
func (h *companyHandler) DeleteCompany(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.DeleteCompanyRequest{
		CompanyUUID: c.Params("company_uuid", ""),
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &company_proto.DeleteCompanyRequest{
		OperationId: operationID,
		UserUuid:    utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid: httpReq.CompanyUUID,
	}

	// Запрос в company сервис
	_, err = h.CompanyServiceClient.DeleteCompany(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.DeleteCompanyResponse{}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// CreateCompanyJoinCode
//
//	@Summary			Create join company code
//	@Description	Create code for users to join company and became it's employee (chief only)
//	@Tags					Company
//	@Accept				json
//	@Produce			json
//	@Security			ApiKeyAuth
//	@Param 				company_uuid path string true "Company UUID"
//	@Success			201  {object}  entities.CreateCompanyJoinCodeResponse
//	@Failure			400  {object}  Error.HttpError
//	@Failure			401  {object}  Error.HttpError
//	@Failure			403  {object}  Error.HttpError
//	@Failure			404  {object}  Error.HttpError
//	@Failure			500  {object}  Error.HttpError
//	@Router				/auth/company/{company_uuid}/code [post]
func (h *companyHandler) CreateCompanyJoinCode(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.CreateCompanyJoinCodeRequest{
		CompanyUUID: c.Params("company_uuid", ""),
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &company_proto.CreateCompanyJoinCodeRequest{
		OperationId: operationID,
		UserUuid:    utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid: httpReq.CompanyUUID,
	}

	// Запрос в company сервис
	res, err := h.CompanyServiceClient.CreateCompanyJoinCode(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.CreateCompanyJoinCodeResponse{
		Code: res.GetJoinCode(),
	}

	return c.Status(fiber.StatusCreated).JSON(httpRes)
}

// GetCompanyJoinCodes
//
//	@Summary			Get all company join codes
//	@Description	Get all company join codes (chief only)
//	@Tags					Company
//	@Accept				json
//	@Produce			json
//	@Security			ApiKeyAuth
//	@Param 				company_uuid path string true "Company UUID"
//	@Success			200  {object}  entities.GetCompanyJoinCodesResponse
//	@Failure			400  {object}  Error.HttpError
//	@Failure			401  {object}  Error.HttpError
//	@Failure			403  {object}  Error.HttpError
//	@Failure			404  {object}  Error.HttpError
//	@Failure			500  {object}  Error.HttpError
//	@Router				/auth/company/{company_uuid}/codes [get]
func (h *companyHandler) GetCompanyJoinCodes(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.CreateCompanyJoinCodeRequest{
		CompanyUUID: c.Params("company_uuid", ""),
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &company_proto.GetCompanyJoinCodesRequest{
		OperationId: operationID,
		UserUuid:    utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid: httpReq.CompanyUUID,
	}

	// Запрос в company сервис
	res, err := h.CompanyServiceClient.GetCompanyJoinCodes(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.GetCompanyJoinCodesResponse{
		Codes: res.GetCodes(),
	}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// DeleteCompanyJoinCode
//
//	@Summary			Delete company join code
//	@Description	Delete company join code (chief only)
//	@Tags					Company
//	@Accept				json
//	@Produce			json
//	@Security			ApiKeyAuth
//	@Param 				company_uuid path string true "Company UUID"
//	@Param				data body entities.DeleteCompanyJoinCodeRequest true "Параметры запроса"
//	@Success			200  {object}  entities.DeleteCompanyJoinCodeResponse
//	@Failure			400  {object}  Error.HttpError
//	@Failure			401  {object}  Error.HttpError
//	@Failure			403  {object}  Error.HttpError
//	@Failure			404  {object}  Error.HttpError
//	@Failure			500  {object}  Error.HttpError
//	@Router				/auth/company/{company_uuid}/code [delete]
func (h *companyHandler) DeleteCompanyJoinCode(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Парсинг тела запроса
	httpReq := &entities.DeleteCompanyJoinCodeRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	httpReqFull := &entities.DeleteCompanyJoinCodeRequestFull{
		CompanyUUID: c.Params("company_uuid", ""),
		Code:        httpReq.Code,
	}

	// Валидация
	err := httpReqFull.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &company_proto.DeleteCompanyJoinCodeRequest{
		OperationId: operationID,
		UserUuid:    utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid: httpReqFull.CompanyUUID,
		Code:        httpReqFull.Code,
	}

	// Запрос в company сервис
	_, err = h.CompanyServiceClient.DeleteCompanyJoinCode(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.DeleteCompanyJoinCodeResponse{}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// JoinCompany
//
//	@Summary			Join company
//	@Description	Join company by join code
//	@Tags					Employee
//	@Accept				json
//	@Produce			json
//	@Security			ApiKeyAuth
//	@Param				data body entities.JoinCompanyRequest true "Параметры запроса"
//	@Success			200  {object}  entities.JoinCompanyResponse
//	@Failure			400  {object}  Error.HttpError
//	@Failure			401  {object}  Error.HttpError
//	@Failure			404  {object}  Error.HttpError
//	@Failure			500  {object}  Error.HttpError
//	@Router				/auth/company/join [post]
func (h *companyHandler) JoinCompany(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Парсинг тела запроса
	httpReq := &entities.JoinCompanyRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &company_proto.JoinCompanyRequest{
		OperationId: operationID,
		UserUuid:    utils.GetLocal[string](c, h.userUUIDKey),
		JoinCode:    httpReq.Code,
	}

	// Запрос в company сервис
	res, err := h.CompanyServiceClient.JoinCompany(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.JoinCompanyResponse{
		CompanyUUID: res.GetCompanyUuid(),
		Role:        res.GetRole(),
	}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// GetCompanyEmployee
//
//	@Summary			Get employee info
//	@Description	Get employee info by his uuid
//	@Tags					Employee
//	@Accept				json
//	@Produce			json
//	@Security			ApiKeyAuth
//	@Param 				company_uuid path string true "Company UUID"
//	@Param 				employee_uuid path string true "Employee UUID"
//	@Success			200  {object}  entities.GetCompanyEmployeeResponse
//	@Failure			400  {object}  Error.HttpError
//	@Failure			401  {object}  Error.HttpError
//	@Failure			403  {object}  Error.HttpError
//	@Failure			404  {object}  Error.HttpError
//	@Failure			500  {object}  Error.HttpError
//	@Router				/auth/company/{company_uuid}/employee/{employee_uuid}/info [get]
func (h *companyHandler) GetCompanyEmployee(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.GetCompanyEmployeeRequest{
		TargetUUID:  c.Params("employee_uuid", ""),
		CompanyUUID: c.Params("company_uuid", ""),
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &company_proto.GetCompanyEmployeeRequest{
		OperationId:   operationID,
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		TargetUuid:    httpReq.TargetUUID,
		CompanyUuid:   httpReq.CompanyUUID,
	}

	// Запрос в company сервис
	res, err := h.CompanyServiceClient.GetCompanyEmployee(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.GetCompanyEmployeeResponse{
		Role:     res.GetRole(),
		JoinedAt: res.GetJoinedAt(),
	}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// GetCompanyEmployeesSummary
//
//	@Summary			Get company employees summary
//	@Description	Get company employees summary count by each role
//	@Tags					Employee
//	@Accept				json
//	@Produce			json
//	@Security			ApiKeyAuth
//	@Param 				company_uuid path string true "Company UUID"
//	@Success			200  {object}  entities.GetCompanyEmployeesSummaryResponse
//	@Failure			400  {object}  Error.HttpError
//	@Failure			401  {object}  Error.HttpError
//	@Failure			403  {object}  Error.HttpError
//	@Failure			404  {object}  Error.HttpError
//	@Failure			500  {object}  Error.HttpError
//	@Router				/auth/company/{company_uuid}/employees/summary [get]
func (h *companyHandler) GetCompanyEmployeesSummary(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.GetCompanyEmployeesSummaryRequest{
		CompanyUUID: c.Params("company_uuid", ""),
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &company_proto.GetCompanyEmployeesSummaryRequest{
		OperationId: operationID,
		UserUuid:    utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid: httpReq.CompanyUUID,
	}

	// Запрос в company сервис
	res, err := h.CompanyServiceClient.GetCompanyEmployeesSummary(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.GetCompanyEmployeesSummaryResponse{
		ChiefCount:     res.GetChiefCount(),
		AnalyticsCount: res.GetAnalyticsCount(),
		ManagerCount:   res.GetManagerCount(),
		EngineerCount:  res.GetEngineerCount(),
		UnemployedCoun: res.GetUnemployedCount(),
	}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// RemoveCompanyEmployee
//
//	@Summary			Remove company employee
//	@Description	Remove company employee by his uuid
//	@Tags					Employee
//	@Accept				json
//	@Produce			json
//	@Security			ApiKeyAuth
//	@Param 				company_uuid path string true "Company UUID"
//	@Param 				employee_uuid path string true "Employee UUID"
//	@Success			200  {object}  entities.RemoveCompanyEmployeeResponse
//	@Failure			400  {object}  Error.HttpError
//	@Failure			401  {object}  Error.HttpError
//	@Failure			403  {object}  Error.HttpError
//	@Failure			404  {object}  Error.HttpError
//	@Failure			500  {object}  Error.HttpError
//	@Router				/auth/company/{company_uuid}/employee/{employee_uuid} [delete]
func (h *companyHandler) RemoveCompanyEmployee(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.RemoveCompanyEmployeeRequest{
		CompanyUUID: c.Params("company_uuid", ""),
		TargetUUID:  c.Params("target_uuid", ""),
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &company_proto.RemoveCompanyEmployeeRequest{
		OperationId:   operationID,
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		TargetUuid:    httpReq.TargetUUID,
		CompanyUuid:   httpReq.CompanyUUID,
	}

	// Запрос в company сервис
	_, err = h.CompanyServiceClient.RemoveCompanyEmployee(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.RemoveCompanyEmployeeResponse{}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}
