package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/company/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/errors"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/utils"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/interceptors"
	"google.golang.org/grpc/metadata"
)

type CompanyHandler interface {
	CreateCompany(c *fiber.Ctx) error
	GetCompany(c *fiber.Ctx) error
	GetCompanies(c *fiber.Ctx) error
	GetUserCompanies(c *fiber.Ctx) error
	UpdateCompanyTitle(c *fiber.Ctx) error
	UpdateCompanyStatus(c *fiber.Ctx) error
	DeleteCompany(c *fiber.Ctx) error
	CreateCompanyJoinCode(c *fiber.Ctx) error
	GetCompanyJoinCodes(c *fiber.Ctx) error
	DeleteCompanyJoinCode(c *fiber.Ctx) error
	JoinCompany(c *fiber.Ctx) error
	GetCompanyEmployee(c *fiber.Ctx) error
	GetCompanyEmployees(c *fiber.Ctx) error
	GetCompanyEmployeesSummary(c *fiber.Ctx) error
	UpdateEmployeeRole(c *fiber.Ctx) error
	RemoveCompanyEmployee(c *fiber.Ctx) error
	CreateDepartment(c *fiber.Ctx) error
	GetDepartment(c *fiber.Ctx) error
	GetCompanyDepartments(c *fiber.Ctx) error
	UpdateDepartmentTitle(c *fiber.Ctx) error
	DeleteDepartment(c *fiber.Ctx) error
	AddEmployeeToDepartment(c *fiber.Ctx) error
	RemoveEmployeeFromDepartment(c *fiber.Ctx) error
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
//	@Summary		Create company
//	@Description	Create new company (user that created company became "chief" automatically)
//	@Tags			Company
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			data	body		entities.CreateCompanyRequest	true	"Данные компании"
//	@Success		201		{object}	entities.CreateCompanyResponse
//	@Failure		400		{object}	Error.HttpError
//	@Failure		401		{object}	Error.HttpError
//	@Failure		500		{object}	Error.HttpError
//	@Router			/auth/company/create [post]
func (h *companyHandler) CreateCompany(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.CreateCompanyRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	req := &company_proto.CreateCompanyRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		Title:         httpReq.Title,
	}

	res, err := h.CompanyServiceClient.CreateCompany(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusCreated).JSON(&entities.CreateCompanyResponse{
		CompanyUUID: res.GetCompanyUuid(),
	})
}

// GetCompany
//
//	@Summary		Get company info
//	@Description	Get company info by company uuid
//	@Tags			Company
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string	true	"Company UUID"
//	@Success		200				{object}	entities.GetCompanyResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid} [get]
func (h *companyHandler) GetCompany(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.GetCompanyRequest{
		CompanyUUID: c.Params("company_uuid", ""),
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	res, err := h.CompanyServiceClient.GetCompany(ctx, &company_proto.GetCompanyRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:   httpReq.CompanyUUID,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.GetCompanyResponse{
		CompanyUUID: res.GetCompanyUuid(),
		Title:       res.GetTitle(),
		Status:      res.GetStatus(),
	})
}

// GetCompanies
//
//	@Summary		Get companies list
//	@Description	Get all companies list
//	@Tags			Company
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			offset	query		int	false	"Offset"	default(0)
//	@Param			count	query		int	false	"Count"		default(10)
//	@Success		200		{object}	entities.GetCompaniesResponse
//	@Failure		400		{object}	Error.HttpError
//	@Failure		401		{object}	Error.HttpError
//	@Failure		500		{object}	Error.HttpError
//	@Router			/auth/company/list [get]
func (h *companyHandler) GetCompanies(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.GetCompaniesRequest{}
	if err := c.QueryParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	res, err := h.CompanyServiceClient.GetCompanies(ctx, &company_proto.GetCompaniesRequest{
		Offset: httpReq.Offset,
		Count:  httpReq.Count,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	companies := make([]*entities.GetCompanyResponse, 0, len(res.GetCompanies()))
	for _, company := range res.GetCompanies() {
		companies = append(companies, &entities.GetCompanyResponse{
			CompanyUUID: company.GetCompanyUuid(),
			Title:       company.GetTitle(),
			Status:      company.GetStatus(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(&entities.GetCompaniesResponse{Companies: companies})
}

// GetUserCompanies
//
//	@Summary		Get user companies
//	@Description	Get list of companies where the current user is an employee
//	@Tags			Company
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	entities.GetUserCompaniesResponse
//	@Failure		401	{object}	Error.HttpError
//	@Failure		500	{object}	Error.HttpError
//	@Router			/auth/company/my [get]
func (h *companyHandler) GetUserCompanies(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	res, err := h.CompanyServiceClient.GetUserCompanies(ctx, &company_proto.GetUserCompaniesRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	companies := make([]*entities.GetCompanyResponse, 0, len(res.GetCompanies()))
	for _, company := range res.GetCompanies() {
		companies = append(companies, &entities.GetCompanyResponse{
			CompanyUUID: company.GetCompanyUuid(),
			Title:       company.GetTitle(),
			Status:      company.GetStatus(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(&entities.GetUserCompaniesResponse{Companies: companies})
}

// UpdateCompanyTitle
//
//	@Summary		Update company title
//	@Description	Update company title by company uuid (chief only)
//	@Tags			Company
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string								true	"Company UUID"
//	@Param			data			body		entities.UpdateCompanyTitleRequest	true	"Параметры запроса"
//	@Success		200				{object}	entities.UpdateCompanyTitleResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/title [patch]
func (h *companyHandler) UpdateCompanyTitle(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.UpdateCompanyTitleRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}
	httpReq.CompanyUUID = c.Params("company_uuid", "")

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.CompanyServiceClient.UpdateCompanyTitle(ctx, &company_proto.UpdateCompanyTitleRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:   httpReq.CompanyUUID,
		Title:         httpReq.Title,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.UpdateCompanyTitleResponse{})
}

// UpdateCompanyStatus
//
//	@Summary		Update company status
//	@Description	Update company status by company uuid (chief only). Available statuses: "open", "close"
//	@Tags			Company
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string								true	"Company UUID"
//	@Param			data			body		entities.UpdateCompanyStatusRequest	true	"Параметры запроса"
//	@Success		200				{object}	entities.UpdateCompanyStatusResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/status [patch]
func (h *companyHandler) UpdateCompanyStatus(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.UpdateCompanyStatusRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}
	httpReq.CompanyUUID = c.Params("company_uuid", "")

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.CompanyServiceClient.UpdateCompanyStatus(ctx, &company_proto.UpdateCompanyStatusRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:   httpReq.CompanyUUID,
		Status:        httpReq.Status,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.UpdateCompanyStatusResponse{})
}

// DeleteCompany
//
//	@Summary		Delete company
//	@Description	Delete company by company uuid (chief only)
//	@Tags			Company
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string	true	"Company UUID"
//	@Success		200				{object}	entities.DeleteCompanyResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid} [delete]
func (h *companyHandler) DeleteCompany(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.DeleteCompanyRequest{
		CompanyUUID: c.Params("company_uuid", ""),
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.CompanyServiceClient.DeleteCompany(ctx, &company_proto.DeleteCompanyRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:   httpReq.CompanyUUID,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.DeleteCompanyResponse{})
}

// CreateCompanyJoinCode
//
//	@Summary		Create join company code
//	@Description	Create code for users to join company (chief only). Min: 60s, Max: 604800s (1 week)
//	@Tags			Company
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string									true	"Company UUID"
//	@Param			data			body		entities.CreateCompanyJoinCodeRequest	true	"Параметры запроса"
//	@Success		201				{object}	entities.CreateCompanyJoinCodeResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/code [post]
func (h *companyHandler) CreateCompanyJoinCode(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.CreateCompanyJoinCodeRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}
	httpReq.CompanyUUID = c.Params("company_uuid", "")

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	res, err := h.CompanyServiceClient.CreateCompanyJoinCode(ctx, &company_proto.CreateCompanyJoinCodeRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:   httpReq.CompanyUUID,
		CodeTtl:       httpReq.CodeTTL,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusCreated).JSON(&entities.CreateCompanyJoinCodeResponse{
		Code: res.GetJoinCode(),
	})
}

// GetCompanyJoinCodes
//
//	@Summary		Get all company join codes
//	@Description	Get all company join codes (chief only)
//	@Tags			Company
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string	true	"Company UUID"
//	@Success		200				{object}	entities.GetCompanyJoinCodesResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/codes [get]
func (h *companyHandler) GetCompanyJoinCodes(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.GetCompanyJoinCodesRequest{
		CompanyUUID: c.Params("company_uuid", ""),
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	res, err := h.CompanyServiceClient.GetCompanyJoinCodes(ctx, &company_proto.GetCompanyJoinCodesRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:   httpReq.CompanyUUID,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.GetCompanyJoinCodesResponse{
		Codes: res.GetCodes(),
	})
}

// DeleteCompanyJoinCode
//
//	@Summary		Delete company join code
//	@Description	Delete company join code (chief only)
//	@Tags			Company
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string								true	"Company UUID"
//	@Param			data			body		entities.DeleteCompanyJoinCodeRequest	true	"Параметры запроса"
//	@Success		200				{object}	entities.DeleteCompanyJoinCodeResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/code [delete]
func (h *companyHandler) DeleteCompanyJoinCode(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.DeleteCompanyJoinCodeRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}
	httpReq.CompanyUUID = c.Params("company_uuid", "")

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.CompanyServiceClient.DeleteCompanyJoinCode(ctx, &company_proto.DeleteCompanyJoinCodeRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:   httpReq.CompanyUUID,
		Code:          httpReq.Code,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.DeleteCompanyJoinCodeResponse{})
}

// JoinCompany
//
//	@Summary		Join company
//	@Description	Join company by join code
//	@Tags			Employee
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			data	body		entities.JoinCompanyRequest	true	"Параметры запроса"
//	@Success		200		{object}	entities.JoinCompanyResponse
//	@Failure		400		{object}	Error.HttpError
//	@Failure		401		{object}	Error.HttpError
//	@Failure		404		{object}	Error.HttpError
//	@Failure		500		{object}	Error.HttpError
//	@Router			/auth/company/join [post]
func (h *companyHandler) JoinCompany(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.JoinCompanyRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	res, err := h.CompanyServiceClient.JoinCompany(ctx, &company_proto.JoinCompanyRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		JoinCode:      httpReq.Code,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.JoinCompanyResponse{
		CompanyUUID: res.GetCompanyUuid(),
		Role:        res.GetRole(),
	})
}

// GetCompanyEmployee
//
//	@Summary		Get employee info
//	@Description	Get employee info by his uuid
//	@Tags			Employee
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string	true	"Company UUID"
//	@Param			employee_uuid	path		string	true	"Employee UUID"
//	@Success		200				{object}	entities.GetCompanyEmployeeResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/employee/{employee_uuid}/info [get]
func (h *companyHandler) GetCompanyEmployee(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.GetCompanyEmployeeRequest{
		TargetUUID:  c.Params("employee_uuid", ""),
		CompanyUUID: c.Params("company_uuid", ""),
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	res, err := h.CompanyServiceClient.GetCompanyEmployee(ctx, &company_proto.GetCompanyEmployeeRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		TargetUuid:    httpReq.TargetUUID,
		CompanyUuid:   httpReq.CompanyUUID,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.GetCompanyEmployeeResponse{
		UserUUID:       httpReq.TargetUUID,
		Role:           res.GetRole(),
		DepartmentUUID: res.GetDepartmentUuid(),
		JoinedAt:       res.GetJoinedAt(),
	})
}

// GetCompanyEmployees
//
//	@Summary		Get company employees
//	@Description	Get company employees filtered by role and/or department
//	@Tags			Employee
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string	true	"Company UUID"
//	@Param			department_uuid	query		string	false	"Department UUID"
//	@Param			role			query		string	false	"Role"
//	@Param			offset			query		int		false	"Offset"	default(0)
//	@Param			count			query		int		false	"Count"		default(10)
//	@Success		200				{object}	entities.GetCompanyEmployeesResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/employees/list [get]
func (h *companyHandler) GetCompanyEmployees(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.GetCompanyEmployeesRequest{}
	if err := c.QueryParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}
	httpReq.CompanyUUID = c.Params("company_uuid", "")

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	res, err := h.CompanyServiceClient.GetCompanyEmployees(ctx, &company_proto.GetCompanyEmployeesRequest{
		InitiatorUuid:  utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:    httpReq.CompanyUUID,
		DepartmentUuid: httpReq.DepartmentUUID,
		Role:           httpReq.Role,
		Count:          httpReq.Count,
		Offset:         httpReq.Offset,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	employees := make([]*entities.GetCompanyEmployeeResponse, 0, len(res.GetEmployees()))
	for _, employee := range res.GetEmployees() {
		employees = append(employees, &entities.GetCompanyEmployeeResponse{
			UserUUID:       employee.GetUserUuid(),
			Role:           employee.GetRole(),
			DepartmentUUID: employee.GetDepartmentUuid(),
			JoinedAt:       employee.GetJoinedAt(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(&entities.GetCompanyEmployeesResponse{Employees: employees})
}

// GetCompanyEmployeesSummary
//
//	@Summary		Get company employees summary
//	@Description	Get company employees count by each role
//	@Tags			Employee
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string	true	"Company UUID"
//	@Param			department_uuid	query		string	false	"Department UUID"
//	@Success		200				{object}	entities.GetCompanyEmployeesSummaryResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/employees/summary [get]
func (h *companyHandler) GetCompanyEmployeesSummary(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.GetCompanyEmployeesSummaryRequest{}
	if err := c.QueryParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}
	httpReq.CompanyUUID = c.Params("company_uuid", "")

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	res, err := h.CompanyServiceClient.GetCompanyEmployeesSummary(ctx, &company_proto.GetCompanyEmployeesSummaryRequest{
		InitiatorUuid:  utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:    httpReq.CompanyUUID,
		DepartmentUuid: httpReq.DepartmentUUID,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.GetCompanyEmployeesSummaryResponse{
		ChiefCount:      res.GetChiefCount(),
		AnalyticsCount:  res.GetAnalyticsCount(),
		ManagerCount:    res.GetManagerCount(),
		EngineerCount:   res.GetEngineerCount(),
		InspectorCount:  res.GetInspectorCount(),
		UnemployedCount: res.GetUnemployedCount(),
	})
}

// UpdateEmployeeRole
//
//	@Summary		Update employee role
//	@Description	Update employee role (chief only). Available roles: "unemployed", "engineer", "manager", "analytic", "inspector", "chief"
//	@Tags			Employee
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string								true	"Company UUID"
//	@Param			employee_uuid	path		string								true	"Employee UUID"
//	@Param			data			body		entities.UpdateEmployeeRoleRequest	true	"Параметры запроса"
//	@Success		200				{object}	entities.UpdateEmployeeRoleResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/employee/{employee_uuid}/role [patch]
func (h *companyHandler) UpdateEmployeeRole(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.UpdateEmployeeRoleRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}
	httpReq.CompanyUUID = c.Params("company_uuid", "")
	httpReq.TargetUUID = c.Params("employee_uuid", "")

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.CompanyServiceClient.UpdateEmployeeRole(ctx, &company_proto.UpdateEmployeeRoleRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		TargetUuid:    httpReq.TargetUUID,
		CompanyUuid:   httpReq.CompanyUUID,
		Role:          httpReq.Role,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.UpdateEmployeeRoleResponse{})
}

// RemoveCompanyEmployee
//
//	@Summary		Remove company employee
//	@Description	Remove company employee by his uuid
//	@Tags			Employee
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string	true	"Company UUID"
//	@Param			employee_uuid	path		string	true	"Employee UUID"
//	@Success		200				{object}	entities.RemoveCompanyEmployeeResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/employee/{employee_uuid} [delete]
func (h *companyHandler) RemoveCompanyEmployee(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.RemoveCompanyEmployeeRequest{
		CompanyUUID: c.Params("company_uuid", ""),
		TargetUUID:  c.Params("employee_uuid", ""),
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.CompanyServiceClient.RemoveCompanyEmployee(ctx, &company_proto.RemoveCompanyEmployeeRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		TargetUuid:    httpReq.TargetUUID,
		CompanyUuid:   httpReq.CompanyUUID,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.RemoveCompanyEmployeeResponse{})
}

// CreateDepartment
//
//	@Summary		Create department
//	@Description	Create new department in a company (chief only)
//	@Tags			Department
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string								true	"Company UUID"
//	@Param			data			body		entities.CreateDepartmentRequest	true	"Параметры запроса"
//	@Success		201				{object}	entities.CreateDepartmentResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/department [post]
func (h *companyHandler) CreateDepartment(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.CreateDepartmentRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}
	httpReq.CompanyUUID = c.Params("company_uuid", "")

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	res, err := h.CompanyServiceClient.CreateDepartment(ctx, &company_proto.CreateDepartmentRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:   httpReq.CompanyUUID,
		Title:         httpReq.Title,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusCreated).JSON(&entities.CreateDepartmentResponse{
		DepartmentUUID: res.GetDepartmentUuid(),
	})
}

// GetDepartment
//
//	@Summary		Get department info
//	@Description	Get department info by uuid
//	@Tags			Department
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid		path		string	true	"Company UUID"
//	@Param			department_uuid		path		string	true	"Department UUID"
//	@Success		200					{object}	entities.GetDepartmentResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/department/{department_uuid} [get]
func (h *companyHandler) GetDepartment(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.GetDepartmentRequest{
		CompanyUUID:    c.Params("company_uuid", ""),
		DepartmentUUID: c.Params("department_uuid", ""),
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	res, err := h.CompanyServiceClient.GetDepartment(ctx, &company_proto.GetDepartmentRequest{
		InitiatorUuid:  utils.GetLocal[string](c, h.userUUIDKey),
		DepartmentUuid: httpReq.DepartmentUUID,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.GetDepartmentResponse{
		DepartmentUUID: res.GetDepartmentUuid(),
		CompanyUUID:    res.GetCompanyUuid(),
		Title:          res.GetTitle(),
		CreatedAt:      res.GetCreatedAt(),
		CreatedBy:      res.GetCreatedBy(),
	})
}

// GetCompanyDepartments
//
//	@Summary		Get company departments list
//	@Description	Get list of all departments in a company
//	@Tags			Department
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid	path		string	true	"Company UUID"
//	@Param			offset			query		int		false	"Offset"	default(0)
//	@Param			count			query		int		false	"Count"		default(10)
//	@Success		200				{object}	entities.GetCompanyDepartmentsResponse
//	@Failure		400				{object}	Error.HttpError
//	@Failure		401				{object}	Error.HttpError
//	@Failure		403				{object}	Error.HttpError
//	@Failure		404				{object}	Error.HttpError
//	@Failure		500				{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/departments/list [get]
func (h *companyHandler) GetCompanyDepartments(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.GetCompanyDepartmentsRequest{}
	if err := c.QueryParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}
	httpReq.CompanyUUID = c.Params("company_uuid", "")

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	res, err := h.CompanyServiceClient.GetCompanyDepartments(ctx, &company_proto.GetCompanyDepartmentsRequest{
		InitiatorUuid: utils.GetLocal[string](c, h.userUUIDKey),
		CompanyUuid:   httpReq.CompanyUUID,
		Offset:        httpReq.Offset,
		Count:         httpReq.Count,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	departments := make([]*entities.DepartmentListItem, 0, len(res.GetDepartments()))
	for _, dept := range res.GetDepartments() {
		departments = append(departments, &entities.DepartmentListItem{
			DepartmentUUID: dept.GetDepartmentUuid(),
			Title:          dept.GetTitle(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(&entities.GetCompanyDepartmentsResponse{Departments: departments})
}

// UpdateDepartmentTitle
//
//	@Summary		Update department title
//	@Description	Update department title (chief only)
//	@Tags			Department
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid		path		string									true	"Company UUID"
//	@Param			department_uuid		path		string									true	"Department UUID"
//	@Param			data				body		entities.UpdateDepartmentTitleRequest	true	"Параметры запроса"
//	@Success		200					{object}	entities.UpdateDepartmentTitleResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/department/{department_uuid}/title [patch]
func (h *companyHandler) UpdateDepartmentTitle(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.UpdateDepartmentTitleRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}
	httpReq.CompanyUUID = c.Params("company_uuid", "")
	httpReq.DepartmentUUID = c.Params("department_uuid", "")

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.CompanyServiceClient.UpdateDepartmentTitle(ctx, &company_proto.UpdateDepartmentTitleRequest{
		InitiatorUuid:  utils.GetLocal[string](c, h.userUUIDKey),
		DepartmentUuid: httpReq.DepartmentUUID,
		Title:          httpReq.Title,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.UpdateDepartmentTitleResponse{})
}

// DeleteDepartment
//
//	@Summary		Delete department
//	@Description	Delete department by uuid (chief only)
//	@Tags			Department
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid		path		string	true	"Company UUID"
//	@Param			department_uuid		path		string	true	"Department UUID"
//	@Success		200					{object}	entities.DeleteDepartmentResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/department/{department_uuid} [delete]
func (h *companyHandler) DeleteDepartment(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.DeleteDepartmentRequest{
		CompanyUUID:    c.Params("company_uuid", ""),
		DepartmentUUID: c.Params("department_uuid", ""),
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.CompanyServiceClient.DeleteDepartment(ctx, &company_proto.DeleteDepartmentRequest{
		InitiatorUuid:  utils.GetLocal[string](c, h.userUUIDKey),
		DepartmentUuid: httpReq.DepartmentUUID,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.DeleteDepartmentResponse{})
}

// AddEmployeeToDepartment
//
//	@Summary		Add employee to department
//	@Description	Add an employee to a department
//	@Tags			Department
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid		path		string	true	"Company UUID"
//	@Param			department_uuid		path		string	true	"Department UUID"
//	@Param			employee_uuid		path		string	true	"Employee UUID"
//	@Success		200					{object}	entities.AddEmployeeToDepartmentResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/department/{department_uuid}/employee/{employee_uuid} [post]
func (h *companyHandler) AddEmployeeToDepartment(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.AddEmployeeToDepartmentRequest{
		CompanyUUID:    c.Params("company_uuid", ""),
		DepartmentUUID: c.Params("department_uuid", ""),
		TargetUUID:     c.Params("employee_uuid", ""),
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.CompanyServiceClient.AddEmployeeToDepartment(ctx, &company_proto.AddEmployeeToDepartmentRequest{
		InitiatorUuid:  utils.GetLocal[string](c, h.userUUIDKey),
		DepartmentUuid: httpReq.DepartmentUUID,
		TargetUuid:     httpReq.TargetUUID,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.AddEmployeeToDepartmentResponse{})
}

// RemoveEmployeeFromDepartment
//
//	@Summary		Remove employee from department
//	@Description	Remove an employee from a department
//	@Tags			Department
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			company_uuid		path		string	true	"Company UUID"
//	@Param			department_uuid		path		string	true	"Department UUID"
//	@Param			employee_uuid		path		string	true	"Employee UUID"
//	@Success		200					{object}	entities.RemoveEmployeeFromDepartmentResponse
//	@Failure		400					{object}	Error.HttpError
//	@Failure		401					{object}	Error.HttpError
//	@Failure		403					{object}	Error.HttpError
//	@Failure		404					{object}	Error.HttpError
//	@Failure		500					{object}	Error.HttpError
//	@Router			/auth/company/{company_uuid}/department/{department_uuid}/employee/{employee_uuid} [delete]
func (h *companyHandler) RemoveEmployeeFromDepartment(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.RemoveEmployeeFromDepartmentRequest{
		CompanyUUID:    c.Params("company_uuid", ""),
		DepartmentUUID: c.Params("department_uuid", ""),
		TargetUUID:     c.Params("employee_uuid", ""),
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.CompanyServiceClient.RemoveEmployeeFromDepartment(ctx, &company_proto.RemoveEmployeeFromDepartmentRequest{
		InitiatorUuid:  utils.GetLocal[string](c, h.userUUIDKey),
		DepartmentUuid: httpReq.DepartmentUUID,
		TargetUuid:     httpReq.TargetUUID,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.RemoveEmployeeFromDepartmentResponse{})
}
