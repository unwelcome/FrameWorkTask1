package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	application_proto "github.com/unwelcome/FrameWorkTask1/backend/application/api/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/errors"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/utils"
)

type ApplicationHandler interface {
	CreateApplication(c *fiber.Ctx) error
	GetApplication(c *fiber.Ctx) error
	GetApplications(c *fiber.Ctx) error
	GetCompanyApplicationStatistic(c *fiber.Ctx) error
	GetEmployeeApplicationStatistic(c *fiber.Ctx) error
	UpdateApplicationStatus(c *fiber.Ctx) error
	AssignApplicationToEmployee(c *fiber.Ctx) error
	AddApplicationFixLog(c *fiber.Ctx) error
	DeleteApplication(c *fiber.Ctx) error
}

type applicationHandler struct {
	ApplicationServiceClient application_proto.ApplicationServiceClient
	operationIDKey           string
	userUUIDKey              string
}

func NewApplicationHandler(ApplicationServiceClient application_proto.ApplicationServiceClient, operationIDKey string, userUUIDKey string) ApplicationHandler {
	return &applicationHandler{
		ApplicationServiceClient: ApplicationServiceClient,
		operationIDKey:           operationIDKey,
		userUUIDKey:              userUUIDKey,
	}
}

// CreateApplication
//
//	@Summary		Create application
//	@Description	Create new application (only users with role "inspector" can create applications)
//	@Tags			Application
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			data	body		entities.CreateApplicationRequest	true	"Данные заявки"
//	@Success		201		{object}	entities.CreateApplicationResponse
//	@Failure		400		{object}	Error.HttpError
//	@Failure		401		{object}	Error.HttpError
//	@Failure		403		{object}	Error.HttpError
//	@Failure		500		{object}	Error.HttpError
//	@Router			/auth/application/create [post]
func (h *applicationHandler) CreateApplication(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Парсинг тела запроса
	httpReq := &entities.CreateApplicationRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &application_proto.CreateApplicationRequest{
		OperationId:   operationID,
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:   httpReq.CompanyUUID,
		Title:         httpReq.Title,
		Description:   httpReq.Description,
	}
	// Запрос в application сервис
	res, err := h.ApplicationServiceClient.CreateApplication(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.CreateApplicationResponse{
		ApplicationUUID: res.GetApplicationUuid(),
	}

	return c.Status(fiber.StatusCreated).JSON(httpRes)
}

// GetApplication
//
//	@Summary		Get application info
//	@Description	Get full application info by application uuid (unemployed employees are not allowed)
//	@Tags			Application
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			application_uuid	path		string	true	"Application UUID"
//	@Success		200					{object}	entities.GetApplicationResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/application/{application_uuid} [get]
func (h *applicationHandler) GetApplication(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.GetApplicationRequest{
		ApplicationUUID: c.Params("application_uuid", ""),
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &application_proto.GetApplicationRequest{
		OperationId:     operationID,
		InitiatorUuid:   utils.GetLocal[string](c, h.userUUIDKey),
		ApplicationUuid: httpReq.ApplicationUUID,
	}

	// Запрос в application сервис
	res, err := h.ApplicationServiceClient.GetApplication(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Маппинг fix log-ов
	app := res.GetApplication()
	fixLogs := make([]*entities.FixLogResponse, 0, len(app.GetFixLogs()))
	for _, fl := range app.GetFixLogs() {
		fixLogs = append(fixLogs, &entities.FixLogResponse{
			Text:      fl.GetText(),
			CreatedAt: fl.GetCreatedAt(),
			CreatedBy: fl.GetCreatedBy(),
		})
	}

	// Формируем тело ответа
	httpRes := &entities.GetApplicationResponse{
		Application: &entities.ApplicationResponse{
			ApplicationUUID:     app.GetApplicationUuid(),
			CompanyUUID:         app.GetCompanyUuid(),
			Title:               app.GetTitle(),
			Description:         app.GetDescription(),
			Status:              app.GetStatus(),
			ResponsibleManager:  app.GetResponsibleManager(),
			ResponsibleEngineer: app.GetResponsibleEngineer(),
			CreatedAt:           app.GetCreatedAt(),
			CreatedBy:           app.GetCreatedBy(),
			ClosedAt:            app.GetClosedAt(),
			FixLogs:             fixLogs,
		},
	}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// GetApplications
//
//	@Summary		Get applications list
//	@Description	Get paginated list of company applications (unemployed employees are not allowed)
//	@Tags			Application
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string	true	"Company UUID"
//	@Param			count			query		int		false	"Count"		default(10)
//	@Param			offset			query		int		false	"Offset"	default(0)
//	@Success		200				{object}	entities.GetApplicationsResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/applications/list [get]
func (h *applicationHandler) GetApplications(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Парсинг query параметров
	httpReq := &entities.GetApplicationsRequest{}
	if err := c.QueryParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Получаем company_uuid из пути
	httpReq.CompanyUUID = c.Params("company_uuid", "")

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &application_proto.GetApplicationsRequest{
		OperationId:   operationID,
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:   httpReq.CompanyUUID,
		Count:         httpReq.Count,
		Offset:        httpReq.Offset,
	}

	// Запрос в application сервис
	res, err := h.ApplicationServiceClient.GetApplications(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Маппинг ответа
	applications := make([]*entities.ApplicationResponse, 0, len(res.GetApplications()))
	for _, app := range res.GetApplications() {
		applications = append(applications, &entities.ApplicationResponse{
			ApplicationUUID:     app.GetApplicationUuid(),
			CompanyUUID:         app.GetCompanyUuid(),
			Title:               app.GetTitle(),
			Description:         app.GetDescription(),
			Status:              app.GetStatus(),
			ResponsibleManager:  app.GetResponsibleManager(),
			ResponsibleEngineer: app.GetResponsibleEngineer(),
			CreatedAt:           app.GetCreatedAt(),
			CreatedBy:           app.GetCreatedBy(),
			ClosedAt:            app.GetClosedAt(),
		})
	}

	// Формируем тело ответа
	httpRes := &entities.GetApplicationsResponse{
		Applications: applications,
	}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// GetCompanyApplicationStatistic
//
//	@Summary		Get company application statistic
//	@Description	Get application count by status for a company (analytic and chief only)
//	@Tags			Application
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string	true	"Company UUID"
//	@Success		200				{object}	entities.GetCompanyApplicationStatisticResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/applications/statistic [get]
func (h *applicationHandler) GetCompanyApplicationStatistic(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.GetCompanyApplicationStatisticRequest{
		CompanyUUID: c.Params("company_uuid", ""),
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &application_proto.GetCompanyApplicationStatisticRequest{
		OperationId:   operationID,
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:   httpReq.CompanyUUID,
	}

	// Запрос в application сервис
	res, err := h.ApplicationServiceClient.GetCompanyApplicationStatistic(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.GetCompanyApplicationStatisticResponse{
		Created:          res.GetCreated(),
		Assigned:         res.GetAssigned(),
		InProgress:       res.GetInProgress(),
		OnHold:           res.GetOnHold(),
		AwaitingApproval: res.GetAwaitingApproval(),
		Completed:        res.GetCompleted(),
		Cancelled:        res.GetCancelled(),
		Failed:           res.GetFailed(),
		Archived:         res.GetArchived(),
	}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// GetEmployeeApplicationStatistic
//
//	@Summary		Get employee application statistic
//	@Description	Get application count by status for a specific employee (unemployed employees are not allowed)
//	@Tags			Application
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string	true	"Company UUID"
//	@Param			employee_uuid	path		string	true	"Employee UUID"
//	@Success		200				{object}	entities.GetEmployeeApplicationStatisticResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/employee/{employee_uuid}/applications/statistic [get]
func (h *applicationHandler) GetEmployeeApplicationStatistic(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.GetEmployeeApplicationStatisticRequest{
		CompanyUUID: c.Params("company_uuid", ""),
		TargetUUID:  c.Params("employee_uuid", ""),
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &application_proto.GetEmployeeApplicationStatisticRequest{
		OperationId:   operationID,
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:   httpReq.CompanyUUID,
		TargetUuid:    httpReq.TargetUUID,
	}

	// Запрос в application сервис
	res, err := h.ApplicationServiceClient.GetEmployeeApplicationStatistic(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.GetEmployeeApplicationStatisticResponse{
		Created:          res.GetCreated(),
		Assigned:         res.GetAssigned(),
		InProgress:       res.GetInProgress(),
		OnHold:           res.GetOnHold(),
		AwaitingApproval: res.GetAwaitingApproval(),
		Completed:        res.GetCompleted(),
		Cancelled:        res.GetCancelled(),
		Failed:           res.GetFailed(),
		Archived:         res.GetArchived(),
	}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// UpdateApplicationStatus
//
//	@Summary		Update application status
//	@Description	Update application status. Engineer: in_progress, on_hold, awaiting_approval. Manager: completed, cancelled, failed.
//	@Tags			Application
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			application_uuid	path		string									true	"Application UUID"
//	@Param			data				body		entities.UpdateApplicationStatusRequest	true	"Параметры запроса"
//	@Success		200					{object}	entities.UpdateApplicationStatusResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/application/{application_uuid}/status [patch]
func (h *applicationHandler) UpdateApplicationStatus(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Парсинг тела запроса
	httpReq := &entities.UpdateApplicationStatusRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	httpReqFull := &entities.UpdateApplicationStatusRequestFull{
		ApplicationUUID: c.Params("application_uuid", ""),
		Status:          httpReq.Status,
	}

	// Валидация
	err := httpReqFull.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &application_proto.UpdateApplicationStatusRequest{
		OperationId:     operationID,
		InitiatorUuid:   utils.GetLocal[string](c, h.userUUIDKey),
		ApplicationUuid: httpReqFull.ApplicationUUID,
		Status:          httpReqFull.Status,
	}

	// Запрос в application сервис
	_, err = h.ApplicationServiceClient.UpdateApplicationStatus(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.UpdateApplicationStatusResponse{}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// AssignApplicationToEmployee
//
//	@Summary		Assign application to employee
//	@Description	Assign application to an engineer (manager only)
//	@Tags			Application
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			application_uuid	path		string										true	"Application UUID"
//	@Param			data				body		entities.AssignApplicationToEmployeeRequest	true	"Параметры запроса"
//	@Success		200					{object}	entities.AssignApplicationToEmployeeResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/application/{application_uuid}/assign [patch]
func (h *applicationHandler) AssignApplicationToEmployee(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Парсинг тела запроса
	httpReq := &entities.AssignApplicationToEmployeeRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	httpReqFull := &entities.AssignApplicationToEmployeeRequestFull{
		ApplicationUUID: c.Params("application_uuid", ""),
		TargetUUID:      httpReq.TargetUUID,
	}

	// Валидация
	err := httpReqFull.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &application_proto.AssignApplicationToEmployeeRequest{
		OperationId:     operationID,
		InitiatorUuid:   utils.GetLocal[string](c, h.userUUIDKey),
		ApplicationUuid: httpReqFull.ApplicationUUID,
		TargetUuid:      httpReqFull.TargetUUID,
	}

	// Запрос в application сервис
	_, err = h.ApplicationServiceClient.AssignApplicationToEmployee(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.AssignApplicationToEmployeeResponse{}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// AddApplicationFixLog
//
//	@Summary		Add application fix log
//	@Description	Add a fix log entry to the application (responsible engineer only)
//	@Tags			Application
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			application_uuid	path		string								true	"Application UUID"
//	@Param			data				body		entities.AddApplicationFixLogRequest	true	"Параметры запроса"
//	@Success		201					{object}	entities.AddApplicationFixLogResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/application/{application_uuid}/fix-log [post]
func (h *applicationHandler) AddApplicationFixLog(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Парсинг тела запроса
	httpReq := &entities.AddApplicationFixLogRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	httpReqFull := &entities.AddApplicationFixLogRequestFull{
		ApplicationUUID: c.Params("application_uuid", ""),
		LogText:         httpReq.LogText,
	}

	// Валидация
	err := httpReqFull.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &application_proto.AddApplicationFixLogRequest{
		OperationId:     operationID,
		InitiatorUuid:   utils.GetLocal[string](c, h.userUUIDKey),
		ApplicationUuid: httpReqFull.ApplicationUUID,
		LogText:         httpReqFull.LogText,
	}

	// Запрос в application сервис
	_, err = h.ApplicationServiceClient.AddApplicationFixLog(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.AddApplicationFixLogResponse{}

	return c.Status(fiber.StatusCreated).JSON(httpRes)
}

// DeleteApplication
//
//	@Summary		Delete application
//	@Description	Soft-delete an application (inspector only, status must be "created")
//	@Tags			Application
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			application_uuid	path		string	true	"Application UUID"
//	@Success		200					{object}	entities.DeleteApplicationResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/application/{application_uuid} [delete]
func (h *applicationHandler) DeleteApplication(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.DeleteApplicationRequest{
		ApplicationUUID: c.Params("application_uuid", ""),
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &application_proto.DeleteApplicationRequest{
		OperationId:     operationID,
		InitiatorUuid:   utils.GetLocal[string](c, h.userUUIDKey),
		ApplicationUuid: httpReq.ApplicationUUID,
	}

	// Запрос в application сервис
	_, err = h.ApplicationServiceClient.DeleteApplication(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.DeleteApplicationResponse{}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}
