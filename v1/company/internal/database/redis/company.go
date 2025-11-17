package redisDB

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	Error "github.com/unwelcome/FrameWorkTask1/v1/company/pkg/errors"
	"google.golang.org/grpc/codes"
)

type CompanyRepository interface {
	CreateCompanyJoinCode(ctx context.Context, companyUUID string, code string, tokenTTL time.Duration) Error.CodeError
	CheckJoinCodeExists(ctx context.Context, code string) Error.CodeError
	CheckJoinCodeBelongToCompany(ctx context.Context, companyUUID string, code string) Error.CodeError
	GetCompanyJoinCodes(ctx context.Context, companyUUID string) ([]string, Error.CodeError)
	DeleteCompanyJoinCode(ctx context.Context, companyUUID string, code string) Error.CodeError
}

type companyRepository struct {
	redis *redis.Client
}

func NewCompanyRepository(redis *redis.Client) CompanyRepository {
	return &companyRepository{redis: redis}
}

// CreateCompanyJoinCode Создает новый код для вступления сотрудника в компанию
func (r *companyRepository) CreateCompanyJoinCode(ctx context.Context, companyUUID string, code string, tokenTTL time.Duration) Error.CodeError {
	// Начинаем транзакцию
	pipeline := r.redis.Pipeline()

	// Сохраняем код
	pipeline.Set(ctx, getCodeKey(code), 1, tokenTTL)

	// Добавляем код в коды компании
	pipeline.SAdd(ctx, getCompanyCodesKey(companyUUID), code)

	// Завершаем транзакцию
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	return Error.CodeError{Code: -1, Err: nil}
}

// CheckJoinCodeExists Проверяет, что код для вступления существует
func (r *companyRepository) CheckJoinCodeExists(ctx context.Context, code string) Error.CodeError {
	// Получение токена
	exist, err := r.redis.Exists(ctx, getCodeKey(code)).Result()
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	// Токен не существует (возможно истек)
	if exist == 0 {
		return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("code not found")}
	}

	return Error.CodeError{Code: -1, Err: nil}
}

// CheckJoinCodeBelongToCompany Проверяет, что код для вступления принадлежит конкретной компании
func (r *companyRepository) CheckJoinCodeBelongToCompany(ctx context.Context, companyUUID string, code string) Error.CodeError {
	// Проверка наличия кода у компании
	exist, err := r.redis.SIsMember(ctx, getCompanyCodesKey(companyUUID), code).Result()
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	if !exist {
		return Error.CodeError{Code: int(codes.PermissionDenied), Err: fmt.Errorf("code not belong to company")}
	}

	return Error.CodeError{Code: -1, Err: nil}
}

// GetCompanyJoinCodes Получение всех кодов компании для вступления
func (r *companyRepository) GetCompanyJoinCodes(ctx context.Context, companyUUID string) ([]string, Error.CodeError) {
	// Получаем коды компании
	companyCodes, err := r.redis.SMembers(ctx, getCompanyCodesKey(companyUUID)).Result()
	if err != nil {
		return nil, Error.CodeError{Code: 0, Err: err}
	}

	validCodes := make([]string, 0)
	invalidCodes := make([]string, 0)

	// Проверяем существование кодов
	for _, code := range companyCodes {
		existErr := r.CheckJoinCodeExists(ctx, code)

		if existErr.Code != -1 { // Код не найден -> добавляем в массив на удаление
			invalidCodes = append(invalidCodes, code)
		} else { // Код найден -> записываем в ответ
			validCodes = append(validCodes, code)
		}
	}

	// Удаляем истекшие коды
	_ = r.redis.SRem(ctx, getCompanyCodesKey(companyUUID), invalidCodes).Err()

	return validCodes, Error.CodeError{Code: -1, Err: nil}
}

// DeleteCompanyJoinCode Удаляет код для вступления
func (r *companyRepository) DeleteCompanyJoinCode(ctx context.Context, companyUUID string, code string) Error.CodeError {
	// Удаляем код из кодов компании
	rmCount, err := r.redis.SRem(ctx, getCompanyCodesKey(companyUUID), code).Result()
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	if rmCount == 0 {
		return Error.CodeError{Code: int(codes.PermissionDenied), Err: fmt.Errorf("code not belong to company")}
	}

	// Удаляем сам код только если он успешно удалился из кодов компании
	err = r.redis.Del(ctx, getCodeKey(code)).Err()
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	return Error.CodeError{Code: -1, Err: nil}
}

// getCompanyCodesKey Возвращает ключ для получения всех кодов для вступления в компанию
func getCompanyCodesKey(companyUUID string) string {
	return fmt.Sprintf("company:%s:codes", companyUUID)
}

// getCodeKey Возвращает ключ для получения информации о коде для вступления в компанию
func getCodeKey(code string) string {
	return fmt.Sprintf("code:%s", code)
}
