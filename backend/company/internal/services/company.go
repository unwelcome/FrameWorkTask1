package services

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	"github.com/google/uuid"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/company/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/backend/company/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/backend/company/internal/entities"
	pb "github.com/unwelcome/FrameWorkTask1/backend/contracts/company/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/helpers"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/validate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const JoinCodeLength = 6
const JoinCodeCreateTries = 10

var AllStatuses = []string{"open", "close"}
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
func (s *CompanyService) Health(ctx context.Context, _ *emptypb.Empty) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Service:  "healthy",
		Postgres: helpers.PingStatus(s.db.Ping(ctx)),
		Redis:    helpers.PingStatus(s.cache.Ping(ctx)),
		Minio:    "not implemented",
		Mongo:    "not implemented",
	}, nil
}

// CreateCompany Создает компанию
func (s *CompanyService) CreateCompany(ctx context.Context, req *pb.CreateCompanyRequest) (*pb.CreateCompanyResponse, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.CompanyTitle(req.GetTitle()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company title")
	}

	companyUUID := uuid.Must(uuid.NewV7()).String()

	if err := s.db.Company.CreateCompany(ctx, &entities.CreateCompany{
		CompanyUUID: companyUUID,
		Title:       req.GetTitle(),
		CreatedBy:   req.GetInitiatorUuid(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &pb.CreateCompanyResponse{CompanyUuid: companyUUID}, nil
}

// GetCompany Возвращает всю информацию о компании
func (s *CompanyService) GetCompany(ctx context.Context, req *pb.GetCompanyRequest) (*pb.GetCompanyResponse, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}

	companyInfo, getErr := s.db.Company.GetCompany(ctx, req.GetCompanyUuid())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	return &pb.GetCompanyResponse{
		CompanyUuid: companyInfo.CompanyUUID,
		Title:       companyInfo.Title,
		Status:      companyInfo.Status,
	}, nil
}

// GetCompanies Возвращает список всех компаний (count, offset)
func (s *CompanyService) GetCompanies(ctx context.Context, req *pb.GetCompaniesRequest) (*pb.GetCompaniesResponse, error) {
	offset := req.GetOffset()
	if offset < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid offset")
	}
	count := req.GetCount()
	if count <= 0 || count > 100 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid count (1..100)")
	}

	companies, getErr := s.db.Company.GetCompanies(ctx, offset, count)
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	resCompanies := make([]*pb.Company, 0)
	for _, company := range companies {
		resCompanies = append(resCompanies, &pb.Company{
			CompanyUuid: company.CompanyUUID,
			Title:       company.Title,
			Status:      company.Status,
		})
	}

	return &pb.GetCompaniesResponse{Companies: resCompanies}, nil
}

// GetUserCompanies Возвращает список компаний, в которых состоит пользователь
func (s *CompanyService) GetUserCompanies(ctx context.Context, req *pb.GetUserCompaniesRequest) (*pb.GetUserCompaniesResponse, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}

	companies, getErr := s.db.Company.GetUserCompanies(ctx, req.GetInitiatorUuid())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	resCompanies := make([]*pb.Company, 0)
	for _, company := range companies {
		resCompanies = append(resCompanies, &pb.Company{
			CompanyUuid: company.CompanyUUID,
			Title:       company.Title,
			Status:      company.Status,
		})
	}

	return &pb.GetUserCompaniesResponse{Companies: resCompanies}, nil
}

// UpdateCompanyTitle Обновляет название компании
func (s *CompanyService) UpdateCompanyTitle(ctx context.Context, req *pb.UpdateCompanyTitleRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	if err := validate.CompanyTitle(req.GetTitle()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company title")
	}

	if err := s.checkEmployeeRole(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"}); err != nil {
		return nil, err
	}

	if err := s.db.Company.UpdateCompanyTitle(ctx, req.GetCompanyUuid(), req.GetTitle()).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// UpdateCompanyStatus Обновляет статус компании (open | close)
func (s *CompanyService) UpdateCompanyStatus(ctx context.Context, req *pb.UpdateCompanyStatusRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	if !helpers.Contains(AllStatuses, req.GetStatus()) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company status")
	}

	if err := s.checkEmployeeRole(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"}); err != nil {
		return nil, err
	}

	if err := s.db.Company.UpdateCompanyStatus(ctx, req.GetCompanyUuid(), req.GetStatus()).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// DeleteCompany Удаляет компанию
func (s *CompanyService) DeleteCompany(ctx context.Context, req *pb.DeleteCompanyRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}

	if err := s.checkEmployeeRole(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"}); err != nil {
		return nil, err
	}

	if err := s.db.Company.DeleteCompany(ctx, req.GetCompanyUuid()).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// CreateCompanyJoinCode Создает код для добавления в компанию
func (s *CompanyService) CreateCompanyJoinCode(ctx context.Context, req *pb.CreateCompanyJoinCodeRequest) (*pb.CreateCompanyJoinCodeResponse, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	// Валидация времени жизни кода: мин - 60 сек / макс - 7 дней
	ttl := req.GetCodeTtl()
	if ttl < 60 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid code ttl (min 60s)")
	}
	if ttl > 60*60*24*7 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid code ttl (max 7 days)")
	}
	joinCodeTTL := time.Second * time.Duration(ttl)

	if err := s.checkEmployeeRole(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"}); err != nil {
		return nil, err
	}

	// Создаем уникальный код добавления
	var joinCode string
	found := false
	for i := 0; i < JoinCodeCreateTries; i++ {
		code, genErr := generateJoinCode()
		if genErr != nil {
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
		return nil, status.Error(codes.Internal, "failed to create join code")
	}

	if err := s.cache.Company.CreateCompanyJoinCode(ctx, req.GetCompanyUuid(), joinCode, joinCodeTTL).GRPCError(); err != nil {
		return nil, err
	}

	return &pb.CreateCompanyJoinCodeResponse{JoinCode: joinCode}, nil
}

// GetCompanyJoinCodes Возвращает все активные коды для добавления к компании
func (s *CompanyService) GetCompanyJoinCodes(ctx context.Context, req *pb.GetCompanyJoinCodesRequest) (*pb.GetCompanyJoinCodesResponse, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}

	if err := s.checkEmployeeRole(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"}); err != nil {
		return nil, err
	}

	companyCodes, getCodesErr := s.cache.Company.GetCompanyJoinCodes(ctx, req.GetCompanyUuid())
	if err := getCodesErr.GRPCError(); err != nil {
		return nil, err
	}

	return &pb.GetCompanyJoinCodesResponse{Codes: companyCodes}, nil
}

// DeleteCompanyJoinCode Удаляет код добавления в компанию
func (s *CompanyService) DeleteCompanyJoinCode(ctx context.Context, req *pb.DeleteCompanyJoinCodeRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	if err := validate.CompanyJoinCode(req.GetCode()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company join code")
	}

	if err := s.checkEmployeeRole(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"}); err != nil {
		return nil, err
	}

	// Проверяем, что код существует
	if existErr := s.cache.Company.CheckJoinCodeExists(ctx, req.GetCode()); existErr.Code != 0 {
		return nil, status.Error(codes.NotFound, "join code not found")
	}

	// Проверяем, что код принадлежит компании
	if belongErr := s.cache.Company.CheckJoinCodeBelongToCompany(ctx, req.GetCompanyUuid(), req.GetCode()); belongErr.Code != 0 {
		return nil, status.Error(codes.PermissionDenied, "join code not belong to this company")
	}

	if err := s.cache.Company.DeleteCompanyJoinCode(ctx, req.GetCompanyUuid(), req.GetCode()).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// JoinCompany Добавляет пользователя в компанию
func (s *CompanyService) JoinCompany(ctx context.Context, req *pb.JoinCompanyRequest) (*pb.JoinCompanyResponse, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.CompanyJoinCode(req.GetJoinCode()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company join code")
	}

	// Проверяем, что код существует
	if existErr := s.cache.Company.CheckJoinCodeExists(ctx, req.GetJoinCode()); existErr.Code != 0 {
		return nil, status.Error(codes.NotFound, "join code not found")
	}

	// Получаем uuid компании по коду добавления
	companyUUID, getErr := s.cache.Company.GetCompanyByJoinCode(ctx, req.GetJoinCode())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	// Проверяем, что пользователь еще не состоит в компании
	_, getEmployeeErr := s.db.Company.GetCompanyEmployee(ctx, companyUUID, req.GetInitiatorUuid())
	if getEmployeeErr.Code == 0 {
		return nil, status.Error(codes.AlreadyExists, "user already in company")
	}
	if getEmployeeErr.Code != int(codes.NotFound) {
		if err := getEmployeeErr.GRPCError(); err != nil {
			return nil, err
		}
	}

	// Получаем информацию о компании (статус должен быть open)
	companyInfo, getCompanyErr := s.db.Company.GetCompany(ctx, companyUUID)
	if err := getCompanyErr.GRPCError(); err != nil {
		return nil, err
	}

	if companyInfo.Status != "open" {
		return nil, status.Error(codes.PermissionDenied, "company is closed")
	}

	if err := s.db.Company.JoinCompany(ctx, companyUUID, req.GetInitiatorUuid()).GRPCError(); err != nil {
		return nil, err
	}

	return &pb.JoinCompanyResponse{CompanyUuid: companyUUID, Role: "unemployed"}, nil
}

// GetCompanyEmployee Возвращает роль сотрудника в компании, иначе возвращает ошибку
func (s *CompanyService) GetCompanyEmployee(ctx context.Context, req *pb.GetCompanyEmployeeRequest) (*pb.GetCompanyEmployeeResponse, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	if err := validate.UUID(req.GetTargetUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target uuid")
	}

	if err := s.checkEmployeeRole(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), AllRoles); err != nil {
		return nil, err
	}

	employeeInfo, getErr := s.db.Company.GetCompanyEmployee(ctx, req.GetCompanyUuid(), req.GetTargetUuid())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	return &pb.GetCompanyEmployeeResponse{
		Role:           employeeInfo.Role,
		DepartmentUuid: employeeInfo.DepartmentUUID,
		JoinedAt:       employeeInfo.JoinedAt,
	}, nil
}

// GetCompanyEmployees Возвращает список сотрудников компании с фильтрацией (count, offset, role)
func (s *CompanyService) GetCompanyEmployees(ctx context.Context, req *pb.GetCompanyEmployeesRequest) (*pb.GetCompanyEmployeesResponse, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	if err := validate.UUID(req.GetDepartmentUuid()); err != nil && req.GetDepartmentUuid() != "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid department uuid")
	}
	if req.GetRole() != "" && !helpers.Contains(AllRoles, req.GetRole()) {
		return nil, status.Error(codes.InvalidArgument, "incorrect role")
	}
	if req.GetCount() <= 0 || req.GetCount() > 100 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid count (1..100)")
	}
	if req.GetOffset() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid offset")
	}

	if err := s.checkEmployeeRole(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), AllRoles); err != nil {
		return nil, err
	}

	employees, getErr := s.db.Company.GetCompanyEmployees(ctx, req.GetCompanyUuid(), req.GetDepartmentUuid(), req.GetRole(), req.GetOffset(), req.GetCount())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	resEmployees := make([]*pb.Employee, 0)
	for _, employee := range employees {
		resEmployees = append(resEmployees, &pb.Employee{
			UserUuid:       employee.UserUUID,
			Role:           employee.Role,
			DepartmentUuid: employee.DepartmentUUID,
			JoinedAt:       employee.JoinedAt,
		})
	}

	return &pb.GetCompanyEmployeesResponse{Employees: resEmployees}, nil
}

// GetCompanyEmployeesSummary Возвращает кол-во сотрудников компании по ролям
func (s *CompanyService) GetCompanyEmployeesSummary(ctx context.Context, req *pb.GetCompanyEmployeesSummaryRequest) (*pb.GetCompanyEmployeesSummaryResponse, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	if err := validate.UUID(req.GetDepartmentUuid()); err != nil && req.GetDepartmentUuid() != "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid department uuid")
	}

	if err := s.checkEmployeeRole(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), AllRoles); err != nil {
		return nil, err
	}

	employeesInfo, getErr := s.db.Company.GetCompanyEmployeesSummary(ctx, req.GetCompanyUuid(), req.GetDepartmentUuid())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

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
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	if err := validate.UUID(req.GetTargetUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target uuid")
	}
	if req.GetInitiatorUuid() == req.GetTargetUuid() {
		return nil, status.Error(codes.InvalidArgument, "cannot change your own role")
	}

	if err := s.checkEmployeeRole(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"}); err != nil {
		return nil, err
	}

	// Проверяем наличие сотрудника
	if _, checkErr := s.db.Company.GetCompanyEmployee(ctx, req.GetCompanyUuid(), req.GetTargetUuid()); checkErr.Code != 0 {
		if err := checkErr.GRPCError(); err != nil {
			return nil, err
		}
	}

	if err := s.db.Company.SetCompanyEmployeeRole(ctx, req.GetCompanyUuid(), req.GetTargetUuid(), req.GetRole()).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RemoveCompanyEmployee Удаляет сотрудника из компании
func (s *CompanyService) RemoveCompanyEmployee(ctx context.Context, req *pb.RemoveCompanyEmployeeRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	if err := validate.UUID(req.GetTargetUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target uuid")
	}
	if req.GetInitiatorUuid() == req.GetTargetUuid() {
		return nil, status.Error(codes.InvalidArgument, "cannot remove yourself from company")
	}

	if err := s.checkEmployeeRole(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"}); err != nil {
		return nil, err
	}

	if err := s.db.Company.RemoveCompanyEmployee(ctx, req.GetCompanyUuid(), req.GetTargetUuid()).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// CreateDepartment Создание департамента
func (s *CompanyService) CreateDepartment(ctx context.Context, req *pb.CreateDepartmentRequest) (*pb.CreateDepartmentResponse, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	if err := validate.DepartmentTitle(req.GetTitle()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid department title")
	}

	if err := s.checkEmployeeRole(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), []string{"chief"}); err != nil {
		return nil, err
	}

	departmentUUID := uuid.Must(uuid.NewV7()).String()

	if err := s.db.Company.CreateDepartment(ctx, &entities.CreateDepartment{
		UUID:        departmentUUID,
		CompanyUUID: req.GetCompanyUuid(),
		Title:       req.GetTitle(),
		CreatedBy:   req.GetInitiatorUuid(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &pb.CreateDepartmentResponse{DepartmentUuid: departmentUUID}, nil
}

// AddEmployeeToDepartment Добавление сотрудника в департамент
func (s *CompanyService) AddEmployeeToDepartment(ctx context.Context, req *pb.AddEmployeeToDepartmentRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetDepartmentUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid department uuid")
	}
	if err := validate.UUID(req.GetTargetUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target uuid")
	}

	department, getErr := s.db.Company.GetDepartment(ctx, req.GetDepartmentUuid())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	if err := s.checkEmployeeRole(ctx, department.CompanyUUID, req.GetInitiatorUuid(), []string{"chief"}); err != nil {
		return nil, err
	}

	// Проверяем принадлежность сотрудника к организации и получаем его данные
	target, getTargetErr := s.db.Company.GetCompanyEmployee(ctx, department.CompanyUUID, req.GetTargetUuid())
	if getTargetErr.Code == int(codes.NotFound) {
		return nil, status.Errorf(codes.PermissionDenied, "employee not found in company")
	}
	if err := getTargetErr.GRPCError(); err != nil {
		return nil, err
	}

	if target.DepartmentUUID == req.GetDepartmentUuid() {
		return nil, status.Errorf(codes.AlreadyExists, "employee is already in this department")
	}

	if err := s.db.Company.AddEmployeeToDepartment(ctx, req.GetDepartmentUuid(), department.CompanyUUID, req.GetTargetUuid()).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// GetDepartment Получение департамента по uuid
func (s *CompanyService) GetDepartment(ctx context.Context, req *pb.GetDepartmentRequest) (*pb.GetDepartmentResponse, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetDepartmentUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid department uuid")
	}

	department, getErr := s.db.Company.GetDepartment(ctx, req.GetDepartmentUuid())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	if err := s.checkEmployeeRole(ctx, department.CompanyUUID, req.GetInitiatorUuid(), AllRoles); err != nil {
		return nil, err
	}

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
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	if req.GetOffset() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid offset")
	}
	if req.GetCount() <= 0 || req.GetCount() > 100 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid count (1..100)")
	}

	if err := s.checkEmployeeRole(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), AllRoles); err != nil {
		return nil, err
	}

	departments, getErr := s.db.Company.GetCompanyDepartments(ctx, req.GetCompanyUuid(), req.GetOffset(), req.GetCount())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	res := make([]*pb.Department, 0)
	for _, department := range departments {
		res = append(res, &pb.Department{
			DepartmentUuid: department.UUID,
			Title:          department.Title,
		})
	}

	return &pb.GetCompanyDepartmentsResponse{Departments: res}, nil
}

// UpdateDepartmentTitle Обновление названия департамента
func (s *CompanyService) UpdateDepartmentTitle(ctx context.Context, req *pb.UpdateDepartmentTitleRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetDepartmentUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid department uuid")
	}
	if err := validate.DepartmentTitle(req.GetTitle()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid department title")
	}

	department, getErr := s.db.Company.GetDepartment(ctx, req.GetDepartmentUuid())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	if err := s.checkEmployeeRole(ctx, department.CompanyUUID, req.GetInitiatorUuid(), []string{"chief"}); err != nil {
		return nil, err
	}

	if err := s.db.Company.UpdateDepartmentTitle(ctx, &entities.UpdateDepartment{
		UUID:  req.GetDepartmentUuid(),
		Title: req.GetTitle(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// DeleteDepartment Удаление департамента
func (s *CompanyService) DeleteDepartment(ctx context.Context, req *pb.DeleteDepartmentRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetDepartmentUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid department uuid")
	}

	department, getErr := s.db.Company.GetDepartment(ctx, req.GetDepartmentUuid())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	if err := s.checkEmployeeRole(ctx, department.CompanyUUID, req.GetInitiatorUuid(), []string{"chief"}); err != nil {
		return nil, err
	}

	if err := s.db.Company.DeleteDepartment(ctx, req.GetDepartmentUuid()).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RemoveEmployeeFromDepartment Удаление сотрудника из департамента
func (s *CompanyService) RemoveEmployeeFromDepartment(ctx context.Context, req *pb.RemoveEmployeeFromDepartmentRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetDepartmentUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid department uuid")
	}
	if err := validate.UUID(req.GetTargetUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target uuid")
	}

	department, getErr := s.db.Company.GetDepartment(ctx, req.GetDepartmentUuid())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	if err := s.checkEmployeeRole(ctx, department.CompanyUUID, req.GetInitiatorUuid(), []string{"chief"}); err != nil {
		return nil, err
	}

	target, getTargetErr := s.db.Company.GetCompanyEmployee(ctx, department.CompanyUUID, req.GetTargetUuid())
	if err := getTargetErr.GRPCError(); err != nil {
		return nil, err
	}

	if target.DepartmentUUID != req.GetDepartmentUuid() {
		return nil, status.Errorf(codes.InvalidArgument, "user not in this department")
	}

	if err := s.db.Company.RemoveEmployeeFromDepartment(ctx, department.CompanyUUID, req.GetTargetUuid()).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ДОП ФУНКЦИИ

// generateJoinCode Генерирует криптографически случайную строку цифр длиной JoinCodeLength
func generateJoinCode() (string, error) {
	digits := make([]byte, JoinCodeLength)

	for i := 0; i < JoinCodeLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		digits[i] = byte('0' + n.Int64())
	}

	return string(digits), nil
}

// checkEmployeeRole Проверяет роль пользователя в компании
func (s *CompanyService) checkEmployeeRole(ctx context.Context, companyUUID, userUUID string, requiredRoles []string) error {
	// Проверяем существование компании
	_, getCompanyErr := s.db.Company.GetCompany(ctx, companyUUID)
	if getCompanyErr.Code == int(codes.NotFound) {
		return status.Error(codes.NotFound, "company not found")
	}
	if err := getCompanyErr.GRPCError(); err != nil {
		return err
	}

	// Получаем роль пользователя в компании
	employee, getErr := s.db.Company.GetCompanyEmployee(ctx, companyUUID, userUUID)
	if getErr.Code == int(codes.NotFound) {
		return status.Errorf(codes.PermissionDenied, "access denied")
	}
	if err := getErr.GRPCError(); err != nil {
		return err
	}

	if !helpers.Contains(requiredRoles, employee.Role) {
		return status.Errorf(codes.PermissionDenied, "not enough rights")
	}

	return nil
}
