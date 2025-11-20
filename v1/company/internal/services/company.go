package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	pb "github.com/unwelcome/FrameWorkTask1/v1/company/api"
	postgresDB "github.com/unwelcome/FrameWorkTask1/v1/company/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/v1/company/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/v1/company/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/v1/company/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const JOIN_CODE_LENGTH = 6

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
	return &pb.HealthResponse{Health: "healthy"}, nil
}

// CreateCompany Создает компанию
func (s *CompanyService) CreateCompany(ctx context.Context, req *pb.CreateCompanyRequest) (*pb.CreateCompanyResponse, error) {
	// Создаем uuid компании
	companyUUID := uuid.New().String()

	// Создаем компанию
	createErr := s.db.Company.CreateCompany(ctx, &entities.CreateCompany{
		CompanyUUID: companyUUID,
		Title:       req.GetTitle(),
		CreatedBy:   req.GetUserUuid(),
	})
	err := Error.HandleError(createErr, req.GetOperationId(), "create company")
	if err != nil {
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
	// Валидируем offset
	offset := req.GetOffset()
	if offset < 0 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get companies").Err(fmt.Errorf("invalid offset")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid offset")
	}

	// Валидируем count
	count := req.GetCount()
	if count < 0 || count > 100 {
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
	// Проверка роли пользователя
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "update title", req.GetCompanyUuid(), req.GetUserUuid(), []string{"chief"})
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
	// Проверка роли пользователя
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "update status", req.GetCompanyUuid(), req.GetUserUuid(), []string{"chief"})
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
	// Проверка роли пользователя
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "delete company", req.GetCompanyUuid(), req.GetUserUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Обновляем статус компании
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
	// Проверка роли пользователя
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "delete company", req.GetCompanyUuid(), req.GetUserUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Валидация времени жизни кода
	ttl := req.GetCodeTtl()
	if ttl < 0 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "create join code").Err(fmt.Errorf("invalid code ttl")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid code ttl")
	}
	joinCodeTTL := time.Second * time.Duration(ttl)

	// Создаем код добавления
	joinCode := generateJoinCode(JOIN_CODE_LENGTH)
	loopCount := 0
	for ; true; loopCount++ {
		// Проверяем, что такого кода еще нет
		getErr := s.cache.Company.CheckJoinCodeBelongToCompany(ctx, req.GetCompanyUuid(), joinCode)

		// Код еще не зарегистрирован -> выходим из цикла
		if getErr.Code == -1 {
			break
		}

		// Защита от бесконечного цикла
		if loopCount == 10 {
			break
		}
	}

	// Если сработала защита, значит код не был подобран -> ошибка
	if loopCount == 10 {
		log.Warn().Str("id", req.GetOperationId()).Str("method", "create join code").Err(fmt.Errorf("loop break")).Msg("error")
		return nil, status.Errorf(codes.Internal, "failed to create join code")
	}

	// Сохраняем код добавления
	saveErr := s.cache.Company.CreateCompanyJoinCode(ctx, req.GetCompanyUuid(), joinCode, joinCodeTTL)
	err = Error.HandleError(&saveErr, req.GetOperationId(), "create join code")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "create join code").Msg("success")
	return &pb.CreateCompanyJoinCodeResponse{JoinCode: joinCode}, nil
}

// GetCompanyJoinCodes Возвращает все активные коды для добавления к компании
func (s *CompanyService) GetCompanyJoinCodes(ctx context.Context, req *pb.GetCompanyJoinCodesRequest) (*pb.GetCompanyJoinCodesResponse, error) {
	// Проверка роли пользователя
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "get company codes", req.GetCompanyUuid(), req.GetUserUuid(), []string{"chief"})
	if err != nil {
		return nil, err
	}

	// Получаем все коды добавления компании
	codes, getCodesErr := s.cache.Company.GetCompanyJoinCodes(ctx, req.GetCompanyUuid())
	err = Error.HandleError(&getCodesErr, req.GetOperationId(), "get company codes")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get company codes").Msg("success")
	return &pb.GetCompanyJoinCodesResponse{Codes: codes}, nil
}

// DeleteCompanyJoinCode Удаляет код добавления в компанию
func (s *CompanyService) DeleteCompanyJoinCode(ctx context.Context, req *pb.DeleteCompanyJoinCodeRequest) (*emptypb.Empty, error) {
	// Проверка роли пользователя
	err := s.checkEmployeeRole(ctx, req.GetOperationId(), "delete join code", req.GetCompanyUuid(), req.GetUserUuid(), []string{"chief"})
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
		return nil, status.Error(codes.PermissionDenied, "join code not belong to company")
	}

	// Удаляем код добавления
	deleteErr := s.cache.Company.DeleteCompanyJoinCode(ctx, req.GetCompanyUuid(), req.GetCode())
	err = Error.HandleError(&deleteErr, req.GetOperationId(), "delete join code")
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
	err := Error.HandleError(&getErr, req.GetOperationId(), "join company")
	if err != nil {
		return nil, err
	}

	// Проверяем, что пользователь еще не состоит в компании
	_, getEmployeeErr := s.db.Company.GetCompanyEmployee(ctx, companyUUID, req.GetUserUuid())
	if getEmployeeErr == nil || getEmployeeErr.Code != int(codes.NotFound) { // ошибка во всех случаях, если ошибка не NotFound (т.е. пользователь не в компании)
		// Если нет ошибки -> Значит пользователь уже в компании -> ошибка
		if getEmployeeErr.Code == -1 {
			return nil, status.Error(codes.AlreadyExists, "user already in company")
		}

		// Иначе возвращаем существующую ошибку
		err = Error.HandleError(getEmployeeErr, req.GetOperationId(), "join company")
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

	// Добавлем пользователя в компанию
	addErr := s.db.Company.JoinCompany(ctx, companyUUID, req.GetUserUuid())
	err = Error.HandleError(addErr, req.GetOperationId(), "join company")
	if err != nil {
		return nil, err
	}

	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// GetCompanyEmployee Возвращает роль сотрудника в компании, иначе возвращает ошибку
func (s *CompanyService) GetCompanyEmployee(ctx context.Context, req *pb.GetCompanyEmployeeRequest) (*pb.GetCompanyEmployeeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// GetCompanyEmployeesSummary Возвращает кол-во сотрудников компании по ролям
func (s *CompanyService) GetCompanyEmployeesSummary(ctx context.Context, req *pb.GetCompanyEmployeesSummaryRequest) (*pb.GetCompanyEmployeesSummaryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// RemoveCompanyEmployee Удаляет сотрудника из компании
func (s *CompanyService) RemoveCompanyEmployee(ctx context.Context, req *pb.RemoveCompanyEmployeeRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// generateJoinCode Генерирует рандомную строку цифр длинной length
func generateJoinCode(length int) string {
	digits := make([]byte, length)

	for i := 0; i < length; i++ {
		digits[i] = byte('0' + rand.Intn(10))
	}

	return string(digits)
}

// checkEmployeeRole Проверяет роль пользователя
func (s *CompanyService) checkEmployeeRole(ctx context.Context, operationUUID, methodName, companyUUID, userUUID string, requiredRoles []string) error {
	// Получаем роль пользователя в компании
	employee, getErr := s.db.Company.GetCompanyEmployee(ctx, companyUUID, userUUID)
	err := Error.HandleError(getErr, operationUUID, methodName)
	if err != nil {
		return err
	}

	// Проверяем роль (только chief может удалять коды добавления)
	if !contain(requiredRoles, employee.Role) {
		log.Info().Str("id", operationUUID).Str("method", methodName).Err(fmt.Errorf("not enought rights")).Msg("error")
		return status.Errorf(codes.PermissionDenied, "not enough rights")
	}

	return nil
}

// contain Функция проверка наличия строки в массиве строк
func contain(arr []string, target string) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}
	return false
}
