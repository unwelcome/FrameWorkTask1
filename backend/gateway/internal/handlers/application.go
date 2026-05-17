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
	UpdateApplicationStatus(c *fiber.Ctx) error
	AssignApplication(c *fiber.Ctx) error
	RedirectApplication(c *fiber.Ctx) error
	RecallApplication(c *fiber.Ctx) error
	TakeApplicationToVerification(c *fiber.Ctx) error
	ReleaseApplicationVerification(c *fiber.Ctx) error
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
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Валидация
	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &application_proto.CreateApplicationRequest{
		OperationId:   operationID,
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:   httpReq.CompanyUUID,
		ApplicationData: &application_proto.ApplicationData{
			Title:       httpReq.Title,
			Description: httpReq.Description,
		},
	}

	// Запрос в application сервис
	res, err := h.ApplicationServiceClient.CreateApplication(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusCreated).JSON(&entities.CreateApplicationResponse{
		ApplicationUUID: res.GetApplicationUuid(),
	})
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

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	res, err := h.ApplicationServiceClient.GetApplication(ctx, &application_proto.GetApplicationRequest{
		OperationId:     operationID,
		InitiatorUuid:   utils.GetLocal[string](c, h.userUUIDKey),
		ApplicationUuid: httpReq.ApplicationUUID,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Маппинг fix log-ов
	app := res.GetApplication()
	fixLogs := make([]*entities.FixLogResponse, 0, len(app.GetFixLogs()))
	for _, fl := range app.GetFixLogs() {
		fixLogs = append(fixLogs, &entities.FixLogResponse{
			UUID:      fl.GetUuid(),
			Text:      fl.GetText(),
			CreatedAt: fl.GetCreatedAt(),
			CreatedBy: fl.GetCreatedBy(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(&entities.GetApplicationResponse{
		Application: &entities.ApplicationResponse{
			ApplicationUUID: app.GetApplicationUuid(),
			CompanyUUID:     app.GetCompanyUuid(),
			DepartmentUUID:  app.GetDepartmentUuid(),
			Version:         app.GetVersion(),
			Title:           app.GetTitle(),
			Description:     app.GetDescription(),
			Status:          app.GetStatus(),
			RevisionCount:   app.GetRevisionCount(),
			CreatedAt:       app.GetCreatedAt(),
			CreatedBy:       app.GetCreatedBy(),
			UpdatedAt:       app.GetUpdatedAt(),
			UpdatedBy:       app.GetUpdatedBy(),
			ManagedBy:       app.GetManagedBy(),
			ExecutedBy:      app.GetExecutedBy(),
			InspectedBy:     app.GetInspectedBy(),
			ClosedAt:        app.GetClosedAt(),
			DeletedAt:       app.GetDeletedAt(),
			DeletedBy:       app.GetDeletedBy(),
			FixLogs:         fixLogs,
		},
	})
}

// GetApplications
//
//	@Summary		Get applications list
//	@Description	Get paginated list of company applications (unemployed employees are not allowed)
//	@Tags			Application
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid		path		string		true	"Company UUID"
//	@Param			department_uuid		query		string		false	"Department UUID (chief/analytic only)"
//	@Param			statuses			query		[]string	false	"Filter by statuses"
//	@Param			count				query		int			false	"Count"			default(10)
//	@Param			offset				query		int			false	"Offset"		default(0)
//	@Param			is_deleted			query		bool		false	"Include deleted"
//	@Param			from_pool			query		bool		false	"Pool view (inspector/manager)"
//	@Success		200					{object}	entities.GetApplicationsResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/applications/list [get]
func (h *applicationHandler) GetApplications(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.GetApplicationsRequest{}
	if err := c.QueryParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}
	httpReq.CompanyUUID = c.Params("company_uuid", "")

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	res, err := h.ApplicationServiceClient.GetApplications(ctx, &application_proto.GetApplicationsRequest{
		OperationId:    operationID,
		InitiatorUuid:  utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:    httpReq.CompanyUUID,
		DepartmentUuid: httpReq.DepartmentUUID,
		Statuses:       httpReq.Statuses,
		Count:          httpReq.Count,
		Offset:         httpReq.Offset,
		IsDeleted:      httpReq.IsDeleted,
		FromPool:       httpReq.FromPool,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	items := make([]*entities.ApplicationListItem, 0, len(res.GetApplications()))
	for _, app := range res.GetApplications() {
		items = append(items, &entities.ApplicationListItem{
			ApplicationUUID: app.GetApplicationUuid(),
			Title:           app.GetTitle(),
			Status:          app.GetStatus(),
			CreatedAt:       app.GetCreatedAt(),
			UpdatedAt:       app.GetUpdatedAt(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(&entities.GetApplicationsResponse{Applications: items})
}

// UpdateApplicationStatus
//
//	@Summary		Update application status
//	@Description	Update application status. Inspector: completed, failed, on_revision. Manager: rejected. Engineer: in_progress, on_hold, pending_verification.
//	@Tags			Application
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			application_uuid	path		string									true	"Application UUID"
//	@Param			data				body		entities.UpdateApplicationStatusRequest	true	"Новый статус"
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

	httpReq := &entities.UpdateApplicationStatusRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	httpReqFull := &entities.UpdateApplicationStatusRequestFull{
		ApplicationUUID: c.Params("application_uuid", ""),
		Status:          httpReq.Status,
	}

	if err := httpReqFull.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.ApplicationServiceClient.UpdateApplicationStatus(ctx, &application_proto.UpdateApplicationStatusRequest{
		OperationId:     operationID,
		InitiatorUuid:   utils.GetLocal[string](c, h.userUUIDKey),
		ApplicationUuid: httpReqFull.ApplicationUUID,
		Status:          httpReqFull.Status,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.UpdateApplicationStatusResponse{})
}

// AssignApplication
//
//	@Summary		Assign application to employee
//	@Description	Assign application to an engineer (manager only)
//	@Tags			Application
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			application_uuid	path		string								true	"Application UUID"
//	@Param			data				body		entities.AssignApplicationRequest	true	"UUID инженера"
//	@Success		200					{object}	entities.AssignApplicationResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/application/{application_uuid}/assign [patch]
func (h *applicationHandler) AssignApplication(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.AssignApplicationRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	httpReqFull := &entities.AssignApplicationRequestFull{
		ApplicationUUID: c.Params("application_uuid", ""),
		TargetUUID:      httpReq.TargetUUID,
	}

	if err := httpReqFull.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.ApplicationServiceClient.AssignApplication(ctx, &application_proto.AssignApplicationRequest{
		OperationId:     operationID,
		InitiatorUuid:   utils.GetLocal[string](c, h.userUUIDKey),
		ApplicationUuid: httpReqFull.ApplicationUUID,
		TargetUuid:      httpReqFull.TargetUUID,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.AssignApplicationResponse{})
}

// RedirectApplication
//
//	@Summary		Redirect application to another department
//	@Description	Transfer application to a different department (manager only)
//	@Tags			Application
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			application_uuid	path		string									true	"Application UUID"
//	@Param			data				body		entities.RedirectApplicationRequest		true	"UUID целевого департамента и причина"
//	@Success		200					{object}	entities.RedirectApplicationResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/application/{application_uuid}/redirect [patch]
func (h *applicationHandler) RedirectApplication(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.RedirectApplicationRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	httpReqFull := &entities.RedirectApplicationRequestFull{
		ApplicationUUID:      c.Params("application_uuid", ""),
		TargetDepartmentUUID: httpReq.TargetDepartmentUUID,
		Message:              httpReq.Message,
	}

	if err := httpReqFull.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.ApplicationServiceClient.RedirectApplication(ctx, &application_proto.RedirectApplicationRequest{
		OperationId:          operationID,
		InitiatorUuid:        utils.GetLocal[string](c, h.userUUIDKey),
		ApplicationUuid:      httpReqFull.ApplicationUUID,
		TargetDepartmentUuid: httpReqFull.TargetDepartmentUUID,
		Message:              httpReqFull.Message,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.RedirectApplicationResponse{})
}

// RecallApplication
//
//	@Summary		Recall application from engineer
//	@Description	Return application from engineer back to the pool (responsible manager only)
//	@Tags			Application
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			application_uuid	path		string								true	"Application UUID"
//	@Param			data				body		entities.RecallApplicationRequest	true	"Причина отзыва"
//	@Success		200					{object}	entities.RecallApplicationResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/application/{application_uuid}/recall [patch]
func (h *applicationHandler) RecallApplication(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.RecallApplicationRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	httpReqFull := &entities.RecallApplicationRequestFull{
		ApplicationUUID: c.Params("application_uuid", ""),
		Message:         httpReq.Message,
	}

	if err := httpReqFull.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.ApplicationServiceClient.RecallApplication(ctx, &application_proto.RecallApplicationRequest{
		OperationId:     operationID,
		InitiatorUuid:   utils.GetLocal[string](c, h.userUUIDKey),
		ApplicationUuid: httpReqFull.ApplicationUUID,
		Message:         httpReqFull.Message,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.RecallApplicationResponse{})
}

// TakeApplicationToVerification
//
//	@Summary		Take application to verification
//	@Description	Inspector takes a pending_verification application for review
//	@Tags			Application
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			application_uuid	path		string	true	"Application UUID"
//	@Success		200					{object}	entities.TakeApplicationToVerificationResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/application/{application_uuid}/take-verification [patch]
func (h *applicationHandler) TakeApplicationToVerification(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.TakeApplicationToVerificationRequest{
		ApplicationUUID: c.Params("application_uuid", ""),
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.ApplicationServiceClient.TakeApplicationToVerification(ctx, &application_proto.TakeApplicationToVerificationRequest{
		OperationId:     operationID,
		InitiatorUuid:   utils.GetLocal[string](c, h.userUUIDKey),
		ApplicationUuid: httpReq.ApplicationUUID,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.TakeApplicationToVerificationResponse{})
}

// ReleaseApplicationVerification
//
//	@Summary		Release application verification
//	@Description	Inspector releases a previously taken application back to pending_verification
//	@Tags			Application
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			application_uuid	path		string											true	"Application UUID"
//	@Param			data				body		entities.ReleaseApplicationVerificationRequest	true	"Причина снятия"
//	@Success		200					{object}	entities.ReleaseApplicationVerificationResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/application/{application_uuid}/release-verification [patch]
func (h *applicationHandler) ReleaseApplicationVerification(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpReq := &entities.ReleaseApplicationVerificationRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	httpReqFull := &entities.ReleaseApplicationVerificationRequestFull{
		ApplicationUUID: c.Params("application_uuid", ""),
		Message:         httpReq.Message,
	}

	if err := httpReqFull.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.ApplicationServiceClient.ReleaseApplicationVerification(ctx, &application_proto.ReleaseApplicationVerificationRequest{
		OperationId:     operationID,
		InitiatorUuid:   utils.GetLocal[string](c, h.userUUIDKey),
		ApplicationUuid: httpReqFull.ApplicationUUID,
		Message:         httpReqFull.Message,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.ReleaseApplicationVerificationResponse{})
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
//	@Param			data				body		entities.AddApplicationFixLogRequest	true	"Текст записи"
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

	httpReq := &entities.AddApplicationFixLogRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	httpReqFull := &entities.AddApplicationFixLogRequestFull{
		ApplicationUUID: c.Params("application_uuid", ""),
		Message:         httpReq.Message,
	}

	if err := httpReqFull.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.ApplicationServiceClient.AddApplicationFixLog(ctx, &application_proto.AddApplicationFixLogRequest{
		OperationId:     operationID,
		InitiatorUuid:   utils.GetLocal[string](c, h.userUUIDKey),
		ApplicationUuid: httpReqFull.ApplicationUUID,
		Message:         httpReqFull.Message,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusCreated).JSON(&entities.AddApplicationFixLogResponse{})
}

// DeleteApplication
//
//	@Summary		Delete application
//	@Description	Soft-delete an application (inspector/creator only, status must be "created")
//	@Tags			Application
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			application_uuid	path		string								true	"Application UUID"
//	@Param			data				body		entities.DeleteApplicationRequest	false	"Опциональное сообщение"
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

	httpReq := &entities.DeleteApplicationRequest{}
	// body необязателен — игнорируем ошибку парсинга
	_ = c.BodyParser(httpReq)

	httpReqFull := &entities.DeleteApplicationRequestFull{
		ApplicationUUID: c.Params("application_uuid", ""),
		Message:         httpReq.Message,
	}

	if err := httpReqFull.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.ApplicationServiceClient.DeleteApplication(ctx, &application_proto.DeleteApplicationRequest{
		OperationId:     operationID,
		InitiatorUuid:   utils.GetLocal[string](c, h.userUUIDKey),
		ApplicationUuid: httpReqFull.ApplicationUUID,
		Message:         httpReqFull.Message,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.DeleteApplicationResponse{})
}
