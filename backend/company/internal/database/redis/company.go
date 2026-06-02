package redisDB

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/unwelcome/FrameWorkTask1/backend/company/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
)

type CompanyRepository interface {
	CreateCompanyJoinCode(ctx context.Context, dto entities.CreateCompanyJoinCodeDTO) Error.CodeError
	CheckJoinCodeExists(ctx context.Context, dto entities.CheckJoinCodeExistsDTO) Error.CodeError
	CheckJoinCodeBelongToCompany(ctx context.Context, dto entities.CheckJoinCodeBelongToCompanyDTO) Error.CodeError
	GetCompanyJoinCodes(ctx context.Context, dto entities.GetCompanyJoinCodesDTO) ([]string, Error.CodeError)
	GetCompanyByJoinCode(ctx context.Context, dto entities.GetCompanyByJoinCodeDTO) (string, Error.CodeError)
	DeleteCompanyJoinCode(ctx context.Context, dto entities.DeleteCompanyJoinCodeDTO) Error.CodeError
}

type companyRepository struct {
	redis  *redis.Client
	prefix string
}

func NewCompanyRepository(redis *redis.Client, prefix string) CompanyRepository {
	return &companyRepository{redis: redis, prefix: prefix}
}

// CreateCompanyJoinCode Создает новый код для вступления сотрудника в компанию
func (r *companyRepository) CreateCompanyJoinCode(ctx context.Context, dto entities.CreateCompanyJoinCodeDTO) Error.CodeError {
	pipeline := r.redis.Pipeline()

	pipeline.Set(ctx, r.getCodeKey(dto.Code), dto.CompanyUUID, dto.TTL)
	pipeline.SAdd(ctx, r.getCompanyCodesKey(dto.CompanyUUID), dto.Code)

	_, err := pipeline.Exec(ctx)
	if err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// CheckJoinCodeExists Проверяет, что код для вступления существует
func (r *companyRepository) CheckJoinCodeExists(ctx context.Context, dto entities.CheckJoinCodeExistsDTO) Error.CodeError {
	exist, err := r.redis.Exists(ctx, r.getCodeKey(dto.Code)).Result()
	if err != nil {
		return Error.Internal(err)
	}

	if exist == 0 {
		return Error.Public(codes.NotFound, "code not found")
	}

	return Error.CodeError{}
}

// CheckJoinCodeBelongToCompany Проверяет, что код для вступления принадлежит конкретной компании
func (r *companyRepository) CheckJoinCodeBelongToCompany(ctx context.Context, dto entities.CheckJoinCodeBelongToCompanyDTO) Error.CodeError {
	exist, err := r.redis.SIsMember(ctx, r.getCompanyCodesKey(dto.CompanyUUID), dto.Code).Result()
	if err != nil {
		return Error.Internal(err)
	}

	if !exist {
		return Error.Public(codes.PermissionDenied, "code not belong to company")
	}

	return Error.CodeError{}
}

// GetCompanyJoinCodes Получение всех кодов компании для вступления
func (r *companyRepository) GetCompanyJoinCodes(ctx context.Context, dto entities.GetCompanyJoinCodesDTO) ([]string, Error.CodeError) {
	companyCodes, err := r.redis.SMembers(ctx, r.getCompanyCodesKey(dto.CompanyUUID)).Result()
	if err != nil {
		return nil, Error.Internal(err)
	}

	validCodes := make([]string, 0)
	invalidCodes := make([]string, 0)

	for _, code := range companyCodes {
		existErr := r.CheckJoinCodeExists(ctx, entities.CheckJoinCodeExistsDTO{Code: code})

		if existErr.Code != 0 {
			invalidCodes = append(invalidCodes, code)
		} else {
			validCodes = append(validCodes, code)
		}
	}

	if len(invalidCodes) > 0 {
		members := make([]interface{}, len(invalidCodes))
		for i, c := range invalidCodes {
			members[i] = c
		}
		_ = r.redis.SRem(ctx, r.getCompanyCodesKey(dto.CompanyUUID), members...).Err()
	}

	return validCodes, Error.CodeError{}
}

// GetCompanyByJoinCode Возвращает uuid компании, которой принадлежит данный код добавления
func (r *companyRepository) GetCompanyByJoinCode(ctx context.Context, dto entities.GetCompanyByJoinCodeDTO) (string, Error.CodeError) {
	companyUUID, err := r.redis.Get(ctx, r.getCodeKey(dto.Code)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", Error.Public(codes.NotFound, "join code not found")
		}
		return "", Error.Internal(err)
	}

	return companyUUID, Error.CodeError{}
}

// DeleteCompanyJoinCode Удаляет код для вступления
func (r *companyRepository) DeleteCompanyJoinCode(ctx context.Context, dto entities.DeleteCompanyJoinCodeDTO) Error.CodeError {
	rmCount, err := r.redis.SRem(ctx, r.getCompanyCodesKey(dto.CompanyUUID), dto.Code).Result()
	if err != nil {
		return Error.Internal(err)
	}
	if rmCount == 0 {
		return Error.Public(codes.PermissionDenied, "code not belong to company")
	}

	err = r.redis.Del(ctx, r.getCodeKey(dto.Code)).Err()
	if err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// getCompanyCodesKey Возвращает ключ для получения всех кодов для вступления в компанию
func (r *companyRepository) getCompanyCodesKey(companyUUID string) string {
	return fmt.Sprintf("%s:company:%s:codes", r.prefix, companyUUID)
}

// getCodeKey Возвращает ключ для получения информации о коде для вступления в компанию
func (r *companyRepository) getCodeKey(code string) string {
	return fmt.Sprintf("%s:code:%s", r.prefix, code)
}
