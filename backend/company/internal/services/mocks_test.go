package services

import (
	"context"
	"fmt"
	"time"

	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/company/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/backend/company/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/backend/company/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
)

// ─── Mock: Postgres CompanyRepository ────────────────────────────────────────

type mockPGCompanyRepo struct {
	createCompany              func(ctx context.Context, dto *entities.CreateCompany) Error.CodeError
	getCompany                 func(ctx context.Context, companyUUID string) (*entities.Company, Error.CodeError)
	getCompanies               func(ctx context.Context, offset, count int64) ([]*entities.GetCompanies, Error.CodeError)
	updateCompanyTitle         func(ctx context.Context, companyUUID, title string) Error.CodeError
	updateCompanyStatus        func(ctx context.Context, companyUUID, status string) Error.CodeError
	deleteCompany              func(ctx context.Context, companyUUID string) Error.CodeError
	joinCompany                func(ctx context.Context, companyUUID, userUUID string) Error.CodeError
	getCompanyEmployee         func(ctx context.Context, companyUUID, userUUID string) (*entities.Employee, Error.CodeError)
	getCompanyEmployees        func(ctx context.Context, companyUUID string, offset, count int64) ([]*entities.Employee, Error.CodeError)
	getCompanyEmployeesByRole  func(ctx context.Context, companyUUID, role string, offset, count int64) ([]*entities.Employee, Error.CodeError)
	getCompanyEmployeesSummary func(ctx context.Context, companyUUID string) (*entities.EmployeesSummary, Error.CodeError)
	setCompanyEmployeeRole     func(ctx context.Context, companyUUID, userUUID, role string) Error.CodeError
	removeCompanyEmployee      func(ctx context.Context, companyUUID, userUUID string) Error.CodeError
}

func (m *mockPGCompanyRepo) CreateCompany(ctx context.Context, dto *entities.CreateCompany) Error.CodeError {
	return m.createCompany(ctx, dto)
}
func (m *mockPGCompanyRepo) GetCompany(ctx context.Context, companyUUID string) (*entities.Company, Error.CodeError) {
	return m.getCompany(ctx, companyUUID)
}
func (m *mockPGCompanyRepo) GetCompanies(ctx context.Context, offset, count int64) ([]*entities.GetCompanies, Error.CodeError) {
	return m.getCompanies(ctx, offset, count)
}
func (m *mockPGCompanyRepo) UpdateCompanyTitle(ctx context.Context, companyUUID, title string) Error.CodeError {
	return m.updateCompanyTitle(ctx, companyUUID, title)
}
func (m *mockPGCompanyRepo) UpdateCompanyStatus(ctx context.Context, companyUUID, status string) Error.CodeError {
	return m.updateCompanyStatus(ctx, companyUUID, status)
}
func (m *mockPGCompanyRepo) DeleteCompany(ctx context.Context, companyUUID string) Error.CodeError {
	return m.deleteCompany(ctx, companyUUID)
}
func (m *mockPGCompanyRepo) JoinCompany(ctx context.Context, companyUUID, userUUID string) Error.CodeError {
	return m.joinCompany(ctx, companyUUID, userUUID)
}
func (m *mockPGCompanyRepo) GetCompanyEmployee(ctx context.Context, companyUUID, userUUID string) (*entities.Employee, Error.CodeError) {
	return m.getCompanyEmployee(ctx, companyUUID, userUUID)
}
func (m *mockPGCompanyRepo) GetCompanyEmployees(ctx context.Context, companyUUID string, offset, count int64) ([]*entities.Employee, Error.CodeError) {
	return m.getCompanyEmployees(ctx, companyUUID, offset, count)
}
func (m *mockPGCompanyRepo) GetCompanyEmployeesByRole(ctx context.Context, companyUUID, role string, offset, count int64) ([]*entities.Employee, Error.CodeError) {
	return m.getCompanyEmployeesByRole(ctx, companyUUID, role, offset, count)
}
func (m *mockPGCompanyRepo) GetCompanyEmployeesSummary(ctx context.Context, companyUUID string) (*entities.EmployeesSummary, Error.CodeError) {
	return m.getCompanyEmployeesSummary(ctx, companyUUID)
}
func (m *mockPGCompanyRepo) SetCompanyEmployeeRole(ctx context.Context, companyUUID, userUUID, role string) Error.CodeError {
	return m.setCompanyEmployeeRole(ctx, companyUUID, userUUID, role)
}
func (m *mockPGCompanyRepo) RemoveCompanyEmployee(ctx context.Context, companyUUID, userUUID string) Error.CodeError {
	return m.removeCompanyEmployee(ctx, companyUUID, userUUID)
}

// ─── Mock: Redis CompanyRepository ───────────────────────────────────────────

type mockRedisCompanyRepo struct {
	createCompanyJoinCode        func(ctx context.Context, companyUUID, code string, ttl time.Duration) Error.CodeError
	checkJoinCodeExists          func(ctx context.Context, code string) Error.CodeError
	checkJoinCodeBelongToCompany func(ctx context.Context, companyUUID, code string) Error.CodeError
	getCompanyJoinCodes          func(ctx context.Context, companyUUID string) ([]string, Error.CodeError)
	getCompanyByJoinCode         func(ctx context.Context, code string) (string, Error.CodeError)
	deleteCompanyJoinCode        func(ctx context.Context, companyUUID, code string) Error.CodeError
}

func (m *mockRedisCompanyRepo) CreateCompanyJoinCode(ctx context.Context, companyUUID, code string, ttl time.Duration) Error.CodeError {
	return m.createCompanyJoinCode(ctx, companyUUID, code, ttl)
}
func (m *mockRedisCompanyRepo) CheckJoinCodeExists(ctx context.Context, code string) Error.CodeError {
	return m.checkJoinCodeExists(ctx, code)
}
func (m *mockRedisCompanyRepo) CheckJoinCodeBelongToCompany(ctx context.Context, companyUUID, code string) Error.CodeError {
	return m.checkJoinCodeBelongToCompany(ctx, companyUUID, code)
}
func (m *mockRedisCompanyRepo) GetCompanyJoinCodes(ctx context.Context, companyUUID string) ([]string, Error.CodeError) {
	return m.getCompanyJoinCodes(ctx, companyUUID)
}
func (m *mockRedisCompanyRepo) GetCompanyByJoinCode(ctx context.Context, code string) (string, Error.CodeError) {
	return m.getCompanyByJoinCode(ctx, code)
}
func (m *mockRedisCompanyRepo) DeleteCompanyJoinCode(ctx context.Context, companyUUID, code string) Error.CodeError {
	return m.deleteCompanyJoinCode(ctx, companyUUID, code)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// newTestService создаёт CompanyService с подменёнными зависимостями
func newTestService(pgRepo postgresDB.CompanyRepository, redisRepo redisDB.CompanyRepository) *CompanyService {
	db := &postgresDB.DatabaseRepository{Company: pgRepo}
	cache := &redisDB.CacheRepository{Company: redisRepo}
	return NewCompanyService(db, cache)
}

// emptyPGRepo — заглушка для тестов, где PostgresRepository не должен вызываться
func emptyPGRepo() *mockPGCompanyRepo { return &mockPGCompanyRepo{} }

// emptyRedisRepo — заглушка для тестов, где RedisRepository не должен вызываться
func emptyRedisRepo() *mockRedisCompanyRepo { return &mockRedisCompanyRepo{} }

// ok возвращает CodeError без ошибки (успех)
func ok() Error.CodeError { return Error.CodeError{Code: -1} }

// notFound возвращает CodeError с кодом NotFound
func notFound() Error.CodeError {
	return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("not found")}
}

// internalErr возвращает CodeError с кодом Internal (0)
func internalErr() Error.CodeError {
	return Error.CodeError{Code: 0, Err: fmt.Errorf("db error")}
}

// companyEntity возвращает тестовую Company со статусом "open"
func companyEntity() *entities.Company {
	return &entities.Company{Status: "open", Title: "Test Co"}
}

// chiefEmployee возвращает тестового Employee с ролью "chief"
func chiefEmployee() *entities.Employee {
	return &entities.Employee{Role: "chief"}
}

// pgRepoWithChief настраивает pg-мок так, чтобы checkEmployeeRole проходил для роли "chief"
func pgRepoWithChief() *mockPGCompanyRepo {
	pg := emptyPGRepo()
	pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
		return companyEntity(), ok()
	}
	pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
		return chiefEmployee(), ok()
	}
	return pg
}
