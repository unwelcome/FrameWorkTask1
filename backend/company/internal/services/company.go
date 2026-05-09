package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	pb "github.com/unwelcome/FrameWorkTask1/backend/company/api/generated"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/company/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/backend/company/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/backend/company/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const JoinCodeLength = 6

var AllRoles = []string{"chief", "analytic", "manager", "engineer", "inspector", "unemployed"}

type CompanyService struct {
	db    *postgresDB.DatabaseRepository
	cache *redisDB.CacheRepository
	pb.UnimplementedCompanyServiceServer
}

func NewCompanyService(db *postgresDB.DatabaseRepository, cache *redisDB.CacheRepository) *CompanyService {
	return &CompanyService{
		db:    db,
		cache: cache,
	}
}

// Health Проверка состояния сервиса
func (s *CompanyService) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "health").Msg("success")
	return &pb.HealthResponse{
		Service:  "healthy",
		Postgres: pingStatus(s.db.Ping(ctx)),
		Redis:    pingStatus(s.cache.Ping(ctx)),
		Minio:    "not implemented",
		Mongo:    "not implemented",
	}, nil
}

// CreateCompany Создает компанию
func (s *CompanyService) CreateCompany(ctx context.Context, req *pb.CreateCompanyRequest) (*pb.CreateCompanyResponse, error) {
	// Проверка title
	title := req.GetTitle()
	if strings.TrimSpace(title) == "" || len([]rune(title)) >= 255 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "create company").Err(fmt.Errorf("invalid title")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid title")
	}

	// Создаем uuid компании
	companyUUID := uuid.New().String()

	// Создаем компанию
	createErr := s.db.Company.CreateCompany(ctx, &entities.CreateCompany{
		CompanyUUID: companyUUID,
		Title:       req.GetTitle(),
		CreatedBy:   req.GetInitiatorUuid(),
	})
	err := Error.HandleError(createErr, req.GetOperationId(), "create company")
	if err != nil {
		return nil, err
	}

	// Добавляем руководителя в компанию
	createChiefErr := s.db.Company.JoinCompany(ctx, companyUUID, req.GetInitiatorUuid())
	err = Error.HandleError(createChiefErr, req.GetOperationId(), "create company")
	if err != nil {
		// Не удалось добавить руководителя в компанию -> удаляем компанию
		_ = s.db.Company.DeleteCompany(ctx, companyUUID)
		return nil, err
	}

	// Устанавливаем роль руководителя
	setRoleErr := s.db.Company.SetCompanyEmployeeRole(ctx, companyUUID, req.GetInitiatorUuid(), "chief")
	err = Error.HandleError(setRoleErr, req.GetOperationId(), "create company")
	if err != nil {
		// Не удалось установить руководителю роль "chief" -> удаляем компанию
		_ = s.db.Company.DeleteCompany(ctx, companyUUID)
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "create company").Msg("success")
	return &pb.CreateCompanyResponse{CompanyUuid: companyUUID}, nil
}

// GetCompany Возвращает всю информацию о компании
func (s *CompanyService) GetCompany(ctx context.Context, req *pb.GetCompanyRequest) (*pb.GetCompanyResponse, error) {
	// Получаем данные о компании
	companyInfo, getErr := s.db.Company.GetCompany(ctx, req.GetCompanyUuid())
	err := Error.HandleError(getErr, req.GetOperationId(), "get company")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get company").Msg("success")
	return &pb.GetCompanyResponse{
		CompanyUuid: companyInfo.CompanyUUID,
		Title:       companyInfo.Title,
		Status:      companyInfo.Status,
	}, nil
}

// GetCompanies Возвращает список всех компаний (count, offset)
func (s *CompanyService) GetCompanies(ctx context.Context, req *pb.GetCompaniesRequest) (*pb.GetCompaniesResponse, error) {
	// Валидация offset
	offset := req.GetOffset()
	if offset < 0 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get companies").Err(fmt.Errorf("invalid offset")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid offset")
	}

	// Валидация count
	count := req.GetCount()
	if count <= 0 || count > 100 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get companies").Err(fmt.Errorf("invalid count")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid count (1..100)")
	}

	// Получаем массив компаний длинной count со сдвигом offset
	companies, getErr := s.db.Company.GetCompanies(ctx, offset, count)
	err := Error.HandleError(getErr, req.GetOperationId(), "get companies")
	if err != nil {
		return nil, err
	}

	// Маппинг ответа
	resCompanies := make([]*pb.Company, 0)
	for _, company := range companies {
		resCompanies = append(resCompanies, &pb.Company{
			CompanyUuid: company.CompanyUUID,
			Title:       company.Title,
			Status:      company.Status,
		})
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get companies").Msg("success")
	return &pb.GetCompaniesResponse{Companies: resCompanies}, nil
}

// UpdateCompanyTitle Обновляет название компании
func (s *CompanyService) UpdateCompanyTitle(ctx context.Context, req *pb.UpdateCompanyTitleRequest) (*emptypb.Empty, error) {
	// Проверка роли инициатора
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "update title", req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Обновляем название компании
	updateErr := s.db.Company.UpdateCompanyTitle(ctx, req.GetCompanyUuid(), req.GetTitle())
	err = Error.HandleError(updateErr, req.GetOperationId(), "update title")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "update title").Msg("success")
	return &emptypb.Empty{}, nil
}

// UpdateCompanyStatus Обновляет статус компании (open | close)
func (s *CompanyService) UpdateCompanyStatus(ctx context.Context, req *pb.UpdateCompanyStatusRequest) (*emptypb.Empty, error) {
	// Проверка роли инициатора
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "update status", req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Обновляем статус компании
	updateErr := s.db.Company.UpdateCompanyStatus(ctx, req.GetCompanyUuid(), req.GetStatus())
	err = Error.HandleError(updateErr, req.GetOperationId(), "update status")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "update status").Msg("success")
	return &emptypb.Empty{}, nil
}

// DeleteCompany Удаляет компанию
func (s *CompanyService) DeleteCompany(ctx context.Context, req *pb.DeleteCompanyRequest) (*emptypb.Empty, error) {
	// Проверка роли инициатора
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "delete company", req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Удаляем компанию
	deleteErr := s.db.Company.DeleteCompany(ctx, req.GetCompanyUuid())
	err = Error.HandleError(deleteErr, req.GetOperationId(), "delete company")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "delete company").Msg("success")
	return &emptypb.Empty{}, nil
}

// CreateCompanyJoinCode Создает код для добавления в компанию
func (s *CompanyService) CreateCompanyJoinCode(ctx context.Context, req *pb.CreateCompanyJoinCodeRequest) (*pb.CreateCompanyJoinCodeResponse, error) {
	// Проверка роли инициатора
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "create join code", req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Валидация времени жизни кода: мин - 60 сек / макс - 7 дней
	ttl := req.GetCodeTtl()
	if ttl < 60 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "create join code").Err(fmt.Errorf("too short code ttl")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid code ttl (min 60s)")
	}
	if ttl > 60*60*24*7 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "create join code").Err(fmt.Errorf("too long code ttl")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid code ttl (max 7 days)")
	}
	joinCodeTTL := time.Second * time.Duration(ttl)

	// Создаем уникальный код добавления (до 10 попыток)
	var joinCode string
	found := false
	for i := 0; i < 10; i++ {
		code, genErr := generateJoinCode(JoinCodeLength)
		if genErr != nil {
			log.Error().Str("id", req.GetOperationId()).Str("method", "create join code").Err(genErr).Msg("error")
			return nil, status.Error(codes.Internal, "internal error")
		}

		// Проверяем, что такого кода ещё нет
		checkErr := s.cache.Company.CheckJoinCodeExists(ctx, code)
		if checkErr.Code == int(codes.NotFound) {
			joinCode = code
			found = true
			break
		}
	}

	if !found {
		log.Warn().Str("id", req.GetOperationId()).Str("method", "create join code").Msg("failed to generate unique join code")
		return nil, status.Error(codes.Internal, "failed to create join code")
	}

	// Сохраняем код добавления
	saveErr := s.cache.Company.CreateCompanyJoinCode(ctx, req.GetCompanyUuid(), joinCode, joinCodeTTL)
	err = Error.HandleError(saveErr, req.GetOperationId(), "create join code")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "create join code").Msg("success")
	return &pb.CreateCompanyJoinCodeResponse{JoinCode: joinCode}, nil
}

// GetCompanyJoinCodes Возвращает все активные коды для добавления к компании
func (s *CompanyService) GetCompanyJoinCodes(ctx context.Context, req *pb.GetCompanyJoinCodesRequest) (*pb.GetCompanyJoinCodesResponse, error) {
	// Проверка роли инициатора
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "get company codes", req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Получаем все коды добавления компании
	companyCodes, getCodesErr := s.cache.Company.GetCompanyJoinCodes(ctx, req.GetCompanyUuid())
	err = Error.HandleError(getCodesErr, req.GetOperationId(), "get company codes")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get company codes").Msg("success")
	return &pb.GetCompanyJoinCodesResponse{Codes: companyCodes}, nil
}

// DeleteCompanyJoinCode Удаляет код добавления в компанию
func (s *CompanyService) DeleteCompanyJoinCode(ctx context.Context, req *pb.DeleteCompanyJoinCodeRequest) (*emptypb.Empty, error) {
	// Проверка роли инициатора
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "delete join code", req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Проверяем, что код существует
	existErr := s.cache.Company.CheckJoinCodeExists(ctx, req.GetCode())
	if existErr.Code != -1 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "delete join code").Err(fmt.Errorf("join code not found")).Msg("error")
		return nil, status.Error(codes.NotFound, "join code not found")
	}

	// Проверяем, что код принадлежит компании
	belongErr := s.cache.Company.CheckJoinCodeBelongToCompany(ctx, req.GetCompanyUuid(), req.GetCode())
	if belongErr.Code != -1 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "delete join code").Err(fmt.Errorf("join code not belong to company")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "join code not belong to this company")
	}

	// Удаляем код добавления
	deleteErr := s.cache.Company.DeleteCompanyJoinCode(ctx, req.GetCompanyUuid(), req.GetCode())
	err = Error.HandleError(deleteErr, req.GetOperationId(), "delete join code")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "delete join code").Msg("success")
	return &emptypb.Empty{}, nil
}

// JoinCompany Добавляет пользователя в компанию
func (s *CompanyService) JoinCompany(ctx context.Context, req *pb.JoinCompanyRequest) (*pb.JoinCompanyResponse, error) {
	// Проверяем, что код существует
	existErr := s.cache.Company.CheckJoinCodeExists(ctx, req.GetJoinCode())
	if existErr.Code != -1 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "join company").Err(fmt.Errorf("join code not found")).Msg("error")
		return nil, status.Error(codes.NotFound, "join code not found")
	}

	// Получаем uuid компании по коду добавления
	companyUUID, getErr := s.cache.Company.GetCompanyByJoinCode(ctx, req.GetJoinCode())
	err := Error.HandleError(getErr, req.GetOperationId(), "join company")
	if err != nil {
		return nil, err
	}

	// Проверяем, что пользователь еще не состоит в компании
	_, getEmployeeErr := s.db.Company.GetCompanyEmployee(ctx, companyUUID, req.GetInitiatorUuid())
	if getEmployeeErr.Code != int(codes.NotFound) {
		// Нет ошибки -> пользователь уже в компании
		if getEmployeeErr.Code == -1 {
			log.Info().Str("id", req.GetOperationId()).Str("method", "join company").Err(fmt.Errorf("user already in company")).Msg("error")
			return nil, status.Error(codes.AlreadyExists, "user already in company")
		}
		// Внутренняя ошибка репозитория
		err = Error.HandleError(getEmployeeErr, req.GetOperationId(), "join company")
		if err != nil {
			return nil, err
		}
	}

	// Получаем информацию о компании (статус должен быть open)
	companyInfo, getCompanyErr := s.db.Company.GetCompany(ctx, companyUUID)
	err = Error.HandleError(getCompanyErr, req.GetOperationId(), "join company")
	if err != nil {
		return nil, err
	}

	if companyInfo.Status != "open" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "join company").Err(fmt.Errorf("company is closed")).Msg("error")
		return nil, status.Error(codes.Canceled, "company is closed")
	}

	// Добавление пользователя в компанию
	addErr := s.db.Company.JoinCompany(ctx, companyUUID, req.GetInitiatorUuid())
	err = Error.HandleError(addErr, req.GetOperationId(), "join company")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "join company").Msg("success")
	return &pb.JoinCompanyResponse{CompanyUuid: companyUUID, Role: "unemployed"}, nil
}

// GetCompanyEmployee Возвращает роль сотрудника в компании, иначе возвращает ошибку
func (s *CompanyService) GetCompanyEmployee(ctx context.Context, req *pb.GetCompanyEmployeeRequest) (*pb.GetCompanyEmployeeResponse, error) {
	// Проверка роли инициатора
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "get company employee", req.GetCompanyUuid(), req.GetInitiatorUuid(), AllRoles)
	if err != nil {
		return nil, err
	}

	// Получаем данные сотрудника
	employeeInfo, getErr := s.db.Company.GetCompanyEmployee(ctx, req.GetCompanyUuid(), req.GetTargetUuid())
	err = Error.HandleError(getErr, req.GetOperationId(), "get company employee")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get company employee").Msg("success")
	return &pb.GetCompanyEmployeeResponse{
		Role:           employeeInfo.Role,
		DepartmentUuid: employeeInfo.DepartmentUUID,
		JoinedAt:       employeeInfo.JoinedAt,
	}, nil
}

// GetCompanyEmployees Возвращает список сотрудников компании с фильтрацией (count, offset, role)
func (s *CompanyService) GetCompanyEmployees(ctx context.Context, req *pb.GetCompanyEmployeesRequest) (*pb.GetCompanyEmployeesResponse, error) {
	// Проверка роли инициатора
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "get company employees", req.GetCompanyUuid(), req.GetInitiatorUuid(), AllRoles)
	if err != nil {
		return nil, err
	}

	// Валидация role
	role := req.GetRole()
	if role != "" && !checkArrayContain(AllRoles, role) {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get company employees").Err(fmt.Errorf("invalid role")).Msg("error")
		return nil, status.Error(codes.InvalidArgument, "incorrect role")
	}

	// Валидация count
	count := req.GetCount()
	if count <= 0 || count > 100 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get company employees").Err(fmt.Errorf("invalid count")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid count (1..100)")
	}

	// Валидация offset
	offset := req.GetOffset()
	if offset < 0 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get company employees").Err(fmt.Errorf("invalid offset")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid offset")
	}

	// Получаем сотрудников
	employees, getErr := s.db.Company.GetCompanyEmployees(ctx, req.GetCompanyUuid(), req.GetDepartmentUuid(), role, offset, count)
	err = Error.HandleError(getErr, req.GetOperationId(), "get company employees")
	if err != nil {
		return nil, err
	}

	// Маппинг ответа
	resEmployees := make([]*pb.Employee, 0)
	for _, employee := range employees {
		resEmployees = append(resEmployees, &pb.Employee{
			UserUuid:       employee.UserUUID,
			Role:           employee.Role,
			DepartmentUuid: employee.DepartmentUUID,
			JoinedAt:       employee.JoinedAt,
		})
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get company employees").Msg("success")
	return &pb.GetCompanyEmployeesResponse{Employees: resEmployees}, nil
}

// GetCompanyEmployeesSummary Возвращает кол-во сотрудников компании по ролям
func (s *CompanyService) GetCompanyEmployeesSummary(ctx context.Context, req *pb.GetCompanyEmployeesSummaryRequest) (*pb.GetCompanyEmployeesSummaryResponse, error) {
	// Проверка роли инициатора
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "get company employees summary", req.GetCompanyUuid(), req.GetInitiatorUuid(), AllRoles)
	if err != nil {
		return nil, err
	}

	// Получаем данные сотрудника
	employeesInfo, getErr := s.db.Company.GetCompanyEmployeesSummary(ctx, req.GetCompanyUuid(), req.GetDepartmentUuid())
	err = Error.HandleError(getErr, req.GetOperationId(), "get company employees summary")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get company employees summary").Msg("success")
	return &pb.GetCompanyEmployeesSummaryResponse{
		ChiefCount:      employeesInfo.ChiefCount,
		AnalyticsCount:  employeesInfo.AnalyticCount,
		ManagerCount:    employeesInfo.ManagerCount,
		EngineerCount:   employeesInfo.EngineerCount,
		InspectorCount:  employeesInfo.InspectorCount,
		UnemployedCount: employeesInfo.UnemployedCount,
	}, nil
}

// UpdateEmployeeRole Обновляет роль сотрудника компании
func (s *CompanyService) UpdateEmployeeRole(ctx context.Context, req *pb.UpdateEmployeeRoleRequest) (*emptypb.Empty, error) {
	// Нельзя изменять свою роль
	if req.GetInitiatorUuid() == req.GetTargetUuid() {
		log.Info().Str("id", req.GetOperationId()).Str("method", "update employee role").Err(fmt.Errorf("cannot change your own role")).Msg("error")
		return nil, status.Error(codes.InvalidArgument, "cannot change your own role")
	}

	// Проверяем роль инициатора
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "update employee role", req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Проверяем наличие сотрудника
	_, checkErr := s.db.Company.GetCompanyEmployee(ctx, req.GetCompanyUuid(), req.GetTargetUuid())
	err = Error.HandleError(checkErr, req.GetOperationId(), "update employee role")
	if err != nil {
		return nil, err
	}

	// Обновляем роль сотрудника
	updateErr := s.db.Company.SetCompanyEmployeeRole(ctx, req.GetCompanyUuid(), req.GetTargetUuid(), req.GetRole())
	err = Error.HandleError(updateErr, req.GetOperationId(), "update employee role")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "update employee role").Msg("success")
	return &emptypb.Empty{}, nil
}

// RemoveCompanyEmployee Удаляет сотрудника из компании
func (s *CompanyService) RemoveCompanyEmployee(ctx context.Context, req *pb.RemoveCompanyEmployeeRequest) (*emptypb.Empty, error) {
	// Нельзя удалить себя из компании
	if req.GetInitiatorUuid() == req.GetTargetUuid() {
		log.Info().Str("id", req.GetOperationId()).Str("method", "remove company employee").Err(fmt.Errorf("cannot remove yourself")).Msg("error")
		return nil, status.Error(codes.InvalidArgument, "cannot remove yourself from company")
	}

	// Проверка роли инициатора
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "remove company employee", req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Удаление сотрудника
	deleteErr := s.db.Company.RemoveCompanyEmployee(ctx, req.GetCompanyUuid(), req.GetTargetUuid())
	err = Error.HandleError(deleteErr, req.GetOperationId(), "remove company employee")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "remove company employee").Msg("success")
	return &emptypb.Empty{}, nil
}

// CreateDepartment Создание департамента
func (s *CompanyService) CreateDepartment(ctx context.Context, req *pb.CreateDepartmentRequest) (*pb.CreateDepartmentResponse, error) {
	// Проверяем роль инициатора
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "create department", req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Создаем uuid департамента
	departmentUUID := uuid.New().String()

	// Создаем департамент
	createErr := s.db.Company.CreateDepartment(ctx, &entities.CreateDepartment{
		UUID:        departmentUUID,
		CompanyUUID: req.GetCompanyUuid(),
		Title:       req.GetTitle(),
		CreatedBy:   req.GetInitiatorUuid(),
	})
	err = Error.HandleError(createErr, req.GetOperationId(), "create department")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "create department").Msg("success")
	return &pb.CreateDepartmentResponse{DepartmentUuid: departmentUUID}, nil
}

// AddEmployeeToDepartment Добавление сотрудника в департамент
func (s *CompanyService) AddEmployeeToDepartment(ctx context.Context, req *pb.AddEmployeeToDepartmentRequest) (*emptypb.Empty, error) {
	// Получаем данные департамента
	department, getErr := s.db.Company.GetDepartment(ctx, req.GetDepartmentUuid())
	err := Error.HandleError(getErr, req.GetOperationId(), "add employee to department")
	if err != nil {
		return nil, err
	}

	// Проверяем роль инициатора
	err = s.checkEmployeeRole(ctx, req.GetOperationId(), "add employee to department", department.CompanyUUID, req.GetInitiatorUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Проверяем принадлежность сотрудника к организации
	err = s.checkEmployeeRole(ctx, req.GetOperationId(), "add employee to department", department.CompanyUUID, req.GetTargetUuid(), AllRoles)
	if err != nil {
		return nil, err
	}

	// Добавляем сотрудника в департамент
	addErr := s.db.Company.AddEmployeeToDepartment(ctx, req.GetDepartmentUuid(), department.CompanyUUID, req.GetTargetUuid())
	err = Error.HandleError(addErr, req.GetOperationId(), "add employee to department")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "add employee to department").Msg("success")
	return &emptypb.Empty{}, nil
}

// GetDepartment Получение департамента по uuid
func (s *CompanyService) GetDepartment(ctx context.Context, req *pb.GetDepartmentRequest) (*pb.GetDepartmentResponse, error) {
	// Получение департамента
	department, getErr := s.db.Company.GetDepartment(ctx, req.GetDepartmentUuid())
	err := Error.HandleError(getErr, req.GetOperationId(), "get department")
	if err != nil {
		return nil, err
	}

	// Получение роли инициатора
	err = s.checkEmployeeRole(ctx, req.GetOperationId(), "get department", department.CompanyUUID, req.GetInitiatorUuid(), AllRoles)
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get department").Msg("success")
	return &pb.GetDepartmentResponse{
		DepartmentUuid: req.GetDepartmentUuid(),
		CompanyUuid:    department.CompanyUUID,
		Title:          department.Title,
		CreatedAt:      department.CreatedAt,
		CreatedBy:      department.CreatedBy,
	}, nil
}

// GetCompanyDepartments Получение списка департаментов компании с фильтрацией (offset, count)
func (s *CompanyService) GetCompanyDepartments(ctx context.Context, req *pb.GetCompanyDepartmentsRequest) (*pb.GetCompanyDepartmentsResponse, error) {
	// Проверка роли инициатора
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "get departments", req.GetCompanyUuid(), req.GetInitiatorUuid(), AllRoles)
	if err != nil {
		return nil, err
	}

	// Валидация offset
	offset := req.GetOffset()
	if offset < 0 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get departments").Err(fmt.Errorf("invalid offset")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid offset")
	}

	// Валидация count
	count := req.GetCount()
	if count <= 0 || count > 100 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get departments").Err(fmt.Errorf("invalid count")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid count (1..100)")
	}

	// Получение списка департаментов компании
	departments, getErr := s.db.Company.GetCompanyDepartments(ctx, req.GetCompanyUuid(), offset, count)
	err = Error.HandleError(getErr, req.GetOperationId(), "get departments")
	if err != nil {
		return nil, err
	}

	res := make([]*pb.Department, 0)
	for _, department := range departments {
		res = append(res, &pb.Department{
			DepartmentUuid: department.UUID,
			Title:          department.Title,
		})
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get departments").Msg("success")
	return &pb.GetCompanyDepartmentsResponse{Departments: res}, nil
}

// UpdateDepartmentTitle Обновление названия департамента
func (s *CompanyService) UpdateDepartmentTitle(ctx context.Context, req *pb.UpdateDepartmentTitleRequest) (*emptypb.Empty, error) {
	// Проверка title
	title := req.GetTitle()
	if strings.TrimSpace(title) == "" || len([]rune(title)) >= 255 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "update department title").Err(fmt.Errorf("invalid title")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid title")
	}

	// Получение департамента
	department, getErr := s.db.Company.GetDepartment(ctx, req.GetDepartmentUuid())
	err := Error.HandleError(getErr, req.GetOperationId(), "update department title")
	if err != nil {
		return nil, err
	}

	// Проверка роли инициатора
	err = s.checkEmployeeRole(ctx, req.GetOperationId(), "update department title", department.CompanyUUID, req.GetInitiatorUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Обновление title
	updateErr := s.db.Company.UpdateDepartmentTitle(ctx, &entities.UpdateDepartment{
		UUID:  req.GetDepartmentUuid(),
		Title: title,
	})
	err = Error.HandleError(updateErr, req.GetOperationId(), "update department title")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "update department title").Msg("success")
	return &emptypb.Empty{}, nil
}

// DeleteDepartment Удаление департамента
func (s *CompanyService) DeleteDepartment(ctx context.Context, req *pb.DeleteDepartmentRequest) (*emptypb.Empty, error) {
	// Получение департамента
	department, getErr := s.db.Company.GetDepartment(ctx, req.GetDepartmentUuid())
	err := Error.HandleError(getErr, req.GetOperationId(), "delete department")
	if err != nil {
		return nil, err
	}

	// Получение роли инициатора
	err = s.checkEmployeeRole(ctx, req.GetOperationId(), "delete department", department.CompanyUUID, req.GetInitiatorUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Удаление департамента
	deleteErr := s.db.Company.DeleteDepartment(ctx, req.GetDepartmentUuid())
	err = Error.HandleError(deleteErr, req.GetOperationId(), "delete department")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "delete department").Msg("success")
	return &emptypb.Empty{}, nil
}

// RemoveEmployeeFromDepartment Удаление сотрудника из департамента
func (s *CompanyService) RemoveEmployeeFromDepartment(ctx context.Context, req *pb.RemoveEmployeeFromDepartmentRequest) (*emptypb.Empty, error) {
	// Получаем данные департамента
	department, getErr := s.db.Company.GetDepartment(ctx, req.GetDepartmentUuid())
	err := Error.HandleError(getErr, req.GetOperationId(), "remove employee from department")
	if err != nil {
		return nil, err
	}

	// Получаем роль инициатора
	err = s.checkEmployeeRole(ctx, req.GetOperationId(), "remove employee from department", department.CompanyUUID, req.GetInitiatorUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Получаем данные сотрудника (сравниваем departmentUUID)
	target, getErr := s.db.Company.GetCompanyEmployee(ctx, department.CompanyUUID, req.GetTargetUuid())
	err = Error.HandleError(getErr, req.GetOperationId(), "remove employee from department")
	if err != nil {
		return nil, err
	}

	if target.DepartmentUUID != req.GetDepartmentUuid() {
		log.Info().Str("id", req.GetOperationId()).Str("method", "remove employee from department").Err(fmt.Errorf("user not in this department")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "user not in this department")
	}

	// Убираем сотрудника из департамента
	removeErr := s.db.Company.RemoveEmployeeFromDepartment(ctx, department.CompanyUUID, req.GetTargetUuid())
	err = Error.HandleError(removeErr, req.GetOperationId(), "remove employee from department")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "remove employee from department").Msg("success")
	return &emptypb.Empty{}, nil
}

// ДОП ФУНКЦИИ

// pingStatus Обертка для ошибок пинга бд
func pingStatus(err error) string {
	if err != nil {
		return "not connected"
	}
	return "connected"
}

// generateJoinCode Генерирует криптографически случайную строку цифр длиной length
func generateJoinCode(length int) (string, error) {
	digits := make([]byte, length)

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		digits[i] = byte('0' + n.Int64())
	}

	return string(digits), nil
}

// checkEmployeeRole Проверяет роль пользователя
func (s *CompanyService) checkEmployeeRole(ctx context.Context, operationUUID, methodName, companyUUID, userUUID string, requiredRoles []string) error {
	// Проверяем существование компании
	_, getCompanyErr := s.db.Company.GetCompany(ctx, companyUUID)
	if getCompanyErr.Code == int(codes.NotFound) {
		log.Info().Str("id", operationUUID).Str("method", methodName).Err(fmt.Errorf("company not found")).Msg("error")
		return status.Error(codes.NotFound, "company not found")
	}
	// Обработка внутренней ошибки
	err := Error.HandleError(getCompanyErr, operationUUID, methodName)
	if err != nil {
		log.Info().Str("id", operationUUID).Str("method", methodName).Err(fmt.Errorf("failed get company: %w", err)).Msg("error")
		return status.Error(codes.Internal, "internal error")
	}

	// Получаем роль пользователя в компании
	employee, getErr := s.db.Company.GetCompanyEmployee(ctx, companyUUID, userUUID)
	if getErr.Code == int(codes.NotFound) { // Пользователь не состоит в компании
		log.Info().Str("id", operationUUID).Str("method", methodName).Err(fmt.Errorf("user not belong to company")).Msg("error")
		return status.Errorf(codes.PermissionDenied, "access denied")
	}
	// Обработка внутренней ошибки
	err = Error.HandleError(getErr, operationUUID, methodName)
	if err != nil {
		log.Info().Str("id", operationUUID).Str("method", methodName).Err(fmt.Errorf("failed get employee role: %w", err)).Msg("error")
		return status.Error(codes.Internal, "internal error")
	}

	// Проверяем роль сотрудника
	if !checkArrayContain(requiredRoles, employee.Role) {
		log.Info().Str("id", operationUUID).Str("method", methodName).Err(fmt.Errorf("not enought rights")).Msg("error")
		return status.Errorf(codes.PermissionDenied, "not enough rights")
	}

	return nil
}

// checkArrayContain Проверяет наличие строки в массиве строк
func checkArrayContain(arr []string, target string) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}
	return false
}
