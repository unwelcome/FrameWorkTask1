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
	getCompanyEmployees        func(ctx context.Context, companyUUID, departmentUUID, role string, offset, count int64) ([]*entities.Employee, Error.CodeError)
	getCompanyEmployeesSummary func(ctx context.Context, companyUUID, departmentUUID string) (*entities.EmployeesSummary, Error.CodeError)
	setCompanyEmployeeRole     func(ctx context.Context, companyUUID, userUUID, role string) Error.CodeError
	removeCompanyEmployee      func(ctx context.Context, companyUUID, userUUID string) Error.CodeError
	createDepartment           func(ctx context.Context, dto *entities.CreateDepartment) Error.CodeError
	addEmployeeToDepartment    func(ctx context.Context, departmentUUID, companyUUID, targetUUID string) Error.CodeError
	getDepartment              func(ctx context.Context, departmentUUID string) (*entities.Department, Error.CodeError)
	getCompanyDepartments      func(ctx context.Context, companyUUID string, offset, count int64) ([]*entities.Department, Error.CodeError)
	updateDepartmentTitle      func(ctx context.Context, dto *entities.UpdateDepartment) Error.CodeError
	deleteDepartment           func(ctx context.Context, departmentUUID string) Error.CodeError
	removeEmployeeFromDepartment func(ctx context.Context, companyUUID, targetUUID string) Error.CodeError
	getUserCompanies             func(ctx context.Context, userUUID string) ([]*entities.GetCompanies, Error.CodeError)
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
func (m *mockPGCompanyRepo) GetCompanyEmployees(ctx context.Context, companyUUID, departmentUUID, role string, offset, count int64) ([]*entities.Employee, Error.CodeError) {
	return m.getCompanyEmployees(ctx, companyUUID, departmentUUID, role, offset, count)
}
func (m *mockPGCompanyRepo) GetCompanyEmployeesSummary(ctx context.Context, companyUUID, departmentUUID string) (*entities.EmployeesSummary, Error.CodeError) {
	return m.getCompanyEmployeesSummary(ctx, companyUUID, departmentUUID)
}
func (m *mockPGCompanyRepo) SetCompanyEmployeeRole(ctx context.Context, companyUUID, userUUID, role string) Error.CodeError {
	return m.setCompanyEmployeeRole(ctx, companyUUID, userUUID, role)
}
func (m *mockPGCompanyRepo) RemoveCompanyEmployee(ctx context.Context, companyUUID, userUUID string) Error.CodeError {
	return m.removeCompanyEmployee(ctx, companyUUID, userUUID)
}
func (m *mockPGCompanyRepo) CreateDepartment(ctx context.Context, dto *entities.CreateDepartment) Error.CodeError {
	return m.createDepartment(ctx, dto)
}
func (m *mockPGCompanyRepo) AddEmployeeToDepartment(ctx context.Context, departmentUUID, companyUUID, targetUUID string) Error.CodeError {
	return m.addEmployeeToDepartment(ctx, departmentUUID, companyUUID, targetUUID)
}
func (m *mockPGCompanyRepo) GetDepartment(ctx context.Context, departmentUUID string) (*entities.Department, Error.CodeError) {
	return m.getDepartment(ctx, departmentUUID)
}
func (m *mockPGCompanyRepo) GetCompanyDepartments(ctx context.Context, companyUUID string, offset, count int64) ([]*entities.Department, Error.CodeError) {
	return m.getCompanyDepartments(ctx, companyUUID, offset, count)
}
func (m *mockPGCompanyRepo) UpdateDepartmentTitle(ctx context.Context, dto *entities.UpdateDepartment) Error.CodeError {
	return m.updateDepartmentTitle(ctx, dto)
}
func (m *mockPGCompanyRepo) DeleteDepartment(ctx context.Context, departmentUUID string) Error.CodeError {
	return m.deleteDepartment(ctx, departmentUUID)
}
func (m *mockPGCompanyRepo) RemoveEmployeeFromDepartment(ctx context.Context, companyUUID, targetUUID string) Error.CodeError {
	return m.removeEmployeeFromDepartment(ctx, companyUUID, targetUUID)
}
func (m *mockPGCompanyRepo) GetUserCompanies(ctx context.Context, userUUID string) ([]*entities.GetCompanies, Error.CodeError) {
	return m.getUserCompanies(ctx, userUUID)
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

func newTestService(pgRepo postgresDB.CompanyRepository, redisRepo redisDB.CompanyRepository) *CompanyService {
	db := &postgresDB.DatabaseRepository{Company: pgRepo}
	cache := &redisDB.CacheRepository{Company: redisRepo}
	return NewCompanyService(db, cache)
}

func emptyPGRepo() *mockPGCompanyRepo { return &mockPGCompanyRepo{} }

func emptyRedisRepo() *mockRedisCompanyRepo { return &mockRedisCompanyRepo{} }

func ok() Error.CodeError { return Error.CodeError{} }

func notFound() Error.CodeError {
	return Error.Public(codes.NotFound, "not found")
}

func internalErr() Error.CodeError {
	return Error.Internal(fmt.Errorf("db error"))
}

func companyEntity() *entities.Company {
	return &entities.Company{Status: "open", Title: "Test Co"}
}

func chiefEmployee() *entities.Employee {
	return &entities.Employee{Role: "chief"}
}

func departmentEntity() *entities.Department {
	return &entities.Department{UUID: deptID, CompanyUUID: companyID, Title: "Test Dept"}
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

// pgRepoWithChiefAndDept настраивает pg-мок для методов департаментов:
// getDepartment возвращает тестовый департамент, а initiator всегда chief
func pgRepoWithChiefAndDept() *mockPGCompanyRepo {
	pg := pgRepoWithChief()
	pg.getDepartment = func(_ context.Context, _ string) (*entities.Department, Error.CodeError) {
		return departmentEntity(), ok()
	}
	return pg
}
