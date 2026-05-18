package redisDB

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
)

type CompanyRepository interface {
	CreateCompanyJoinCode(ctx context.Context, companyUUID string, code string, tokenTTL time.Duration) Error.CodeError
	CheckJoinCodeExists(ctx context.Context, code string) Error.CodeError
	CheckJoinCodeBelongToCompany(ctx context.Context, companyUUID string, code string) Error.CodeError
	GetCompanyJoinCodes(ctx context.Context, companyUUID string) ([]string, Error.CodeError)
	GetCompanyByJoinCode(ctx context.Context, code string) (string, Error.CodeError)
	DeleteCompanyJoinCode(ctx context.Context, companyUUID string, code string) Error.CodeError
}

type companyRepository struct {
	redis  *redis.Client
	prefix string
}

func NewCompanyRepository(redis *redis.Client, prefix string) CompanyRepository {
	return &companyRepository{redis: redis, prefix: prefix}
}

// CreateCompanyJoinCode Создает новый код для вступления сотрудника в компанию
func (r *companyRepository) CreateCompanyJoinCode(ctx context.Context, companyUUID string, code string, tokenTTL time.Duration) Error.CodeError {
	// Начинаем транзакцию
	pipeline := r.redis.Pipeline()

	// Сохраняем код
	pipeline.Set(ctx, r.getCodeKey(code), companyUUID, tokenTTL)

	// Добавляем код в коды компании
	pipeline.SAdd(ctx, r.getCompanyCodesKey(companyUUID), code)

	// Завершаем транзакцию
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// CheckJoinCodeExists Проверяет, что код для вступления существует
func (r *companyRepository) CheckJoinCodeExists(ctx context.Context, code string) Error.CodeError {
	// Получение токена
	exist, err := r.redis.Exists(ctx, r.getCodeKey(code)).Result()
	if err != nil {
		return Error.Internal(err)
	}

	// Токен не существует (возможно истек)
	if exist == 0 {
		return Error.Public(codes.NotFound, "code not found")
	}

	return Error.CodeError{}
}

// CheckJoinCodeBelongToCompany Проверяет, что код для вступления принадлежит конкретной компании
func (r *companyRepository) CheckJoinCodeBelongToCompany(ctx context.Context, companyUUID string, code string) Error.CodeError {
	// Проверка наличия кода у компании
	exist, err := r.redis.SIsMember(ctx, r.getCompanyCodesKey(companyUUID), code).Result()
	if err != nil {
		return Error.Internal(err)
	}

	if !exist {
		return Error.Public(codes.PermissionDenied, "code not belong to company")
	}

	return Error.CodeError{}
}

// GetCompanyJoinCodes Получение всех кодов компании для вступления
func (r *companyRepository) GetCompanyJoinCodes(ctx context.Context, companyUUID string) ([]string, Error.CodeError) {
	// Получаем коды компании
	companyCodes, err := r.redis.SMembers(ctx, r.getCompanyCodesKey(companyUUID)).Result()
	if err != nil {
		return nil, Error.Internal(err)
	}

	validCodes := make([]string, 0)
	invalidCodes := make([]string, 0)

	// Проверяем существование кодов
	for _, code := range companyCodes {
		existErr := r.CheckJoinCodeExists(ctx, code)

		if existErr.Code != 0 { // Код не найден -> добавляем в массив на удаление
			invalidCodes = append(invalidCodes, code)
		} else { // Код найден -> записываем в ответ
			validCodes = append(validCodes, code)
		}
	}

	// Удаляем истекшие коды
	_ = r.redis.SRem(ctx, r.getCompanyCodesKey(companyUUID), invalidCodes).Err()

	return validCodes, Error.CodeError{}
}

// GetCompanyByJoinCode Возвращает uuid компании, которой принадлежит данный код добавления
func (r *companyRepository) GetCompanyByJoinCode(ctx context.Context, code string) (string, Error.CodeError) {
	// Получаем uuid компании по коду
	companyUUID, err := r.redis.Get(ctx, r.getCodeKey(code)).Result()
	if err != nil {
		return "", Error.Internal(err)
	}

	return companyUUID, Error.CodeError{}
}

// DeleteCompanyJoinCode Удаляет код для вступления
func (r *companyRepository) DeleteCompanyJoinCode(ctx context.Context, companyUUID string, code string) Error.CodeError {
	// Удаляем код из кодов компании
	rmCount, err := r.redis.SRem(ctx, r.getCompanyCodesKey(companyUUID), code).Result()
	if err != nil {
		return Error.Internal(err)
	}
	if rmCount == 0 {
		return Error.Public(codes.PermissionDenied, "code not belong to company")
	}

	// Удаляем сам код только если он успешно удалился из кодов компании
	err = r.redis.Del(ctx, r.getCodeKey(code)).Err()
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
