package services

import (
	"context"
	"fmt"

	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/company/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/backend/company/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/backend/company/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
)

// ─── Mock: Postgres CompanyRepository ────────────────────────────────────────

type mockPGCompanyRepo struct {
	createCompany              func(ctx context.Context, dto entities.CreateCompany) Error.CodeError
	getCompany                 func(ctx context.Context, dto entities.GetCompanyDTO) (*entities.Company, Error.CodeError)
	getCompanies               func(ctx context.Context, dto entities.GetCompaniesDTO) ([]*entities.GetCompanies, Error.CodeError)
	updateCompanyTitle         func(ctx context.Context, dto entities.UpdateCompanyTitleDTO) Error.CodeError
	updateCompanyStatus        func(ctx context.Context, dto entities.UpdateCompanyStatusDTO) Error.CodeError
	deleteCompany              func(ctx context.Context, dto entities.DeleteCompanyDTO) Error.CodeError
	joinCompany                func(ctx context.Context, dto entities.JoinCompanyDTO) Error.CodeError
	getCompanyEmployee         func(ctx context.Context, dto entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError)
	getCompanyEmployees        func(ctx context.Context, dto entities.GetCompanyEmployeesDTO) ([]*entities.Employee, Error.CodeError)
	getCompanyEmployeesSummary func(ctx context.Context, dto entities.GetCompanyEmployeesSummaryDTO) (*entities.EmployeesSummary, Error.CodeError)
	setCompanyEmployeeRole     func(ctx context.Context, dto entities.SetCompanyEmployeeRoleDTO) Error.CodeError
	removeCompanyEmployee      func(ctx context.Context, dto entities.RemoveCompanyEmployeeDTO) Error.CodeError
	createDepartment           func(ctx context.Context, dto entities.CreateDepartment) Error.CodeError
	addEmployeeToDepartment    func(ctx context.Context, dto entities.AddEmployeeToDepartmentDTO) Error.CodeError
	getDepartment              func(ctx context.Context, dto entities.GetDepartmentDTO) (*entities.Department, Error.CodeError)
	getCompanyDepartments      func(ctx context.Context, dto entities.GetCompanyDepartmentsDTO) ([]*entities.Department, Error.CodeError)
	updateDepartmentTitle      func(ctx context.Context, dto *entities.UpdateDepartment) Error.CodeError
	deleteDepartment           func(ctx context.Context, dto entities.DeleteDepartmentDTO) Error.CodeError
	removeEmployeeFromDepartment func(ctx context.Context, dto entities.RemoveEmployeeFromDepartmentDTO) Error.CodeError
	getUserCompanies             func(ctx context.Context, dto entities.GetUserCompaniesDTO) ([]*entities.GetCompanies, Error.CodeError)
	checkColleagues              func(ctx context.Context, dto entities.CheckColleaguesDTO) (bool, Error.CodeError)
}

func (m *mockPGCompanyRepo) CreateCompany(ctx context.Context, dto entities.CreateCompany) Error.CodeError {
	return m.createCompany(ctx, dto)
}
func (m *mockPGCompanyRepo) GetCompany(ctx context.Context, dto entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
	return m.getCompany(ctx, dto)
}
func (m *mockPGCompanyRepo) GetCompanies(ctx context.Context, dto entities.GetCompaniesDTO) ([]*entities.GetCompanies, Error.CodeError) {
	return m.getCompanies(ctx, dto)
}
func (m *mockPGCompanyRepo) UpdateCompanyTitle(ctx context.Context, dto entities.UpdateCompanyTitleDTO) Error.CodeError {
	return m.updateCompanyTitle(ctx, dto)
}
func (m *mockPGCompanyRepo) UpdateCompanyStatus(ctx context.Context, dto entities.UpdateCompanyStatusDTO) Error.CodeError {
	return m.updateCompanyStatus(ctx, dto)
}
func (m *mockPGCompanyRepo) DeleteCompany(ctx context.Context, dto entities.DeleteCompanyDTO) Error.CodeError {
	return m.deleteCompany(ctx, dto)
}
func (m *mockPGCompanyRepo) JoinCompany(ctx context.Context, dto entities.JoinCompanyDTO) Error.CodeError {
	return m.joinCompany(ctx, dto)
}
func (m *mockPGCompanyRepo) GetCompanyEmployee(ctx context.Context, dto entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
	return m.getCompanyEmployee(ctx, dto)
}
func (m *mockPGCompanyRepo) GetCompanyEmployees(ctx context.Context, dto entities.GetCompanyEmployeesDTO) ([]*entities.Employee, Error.CodeError) {
	return m.getCompanyEmployees(ctx, dto)
}
func (m *mockPGCompanyRepo) GetCompanyEmployeesSummary(ctx context.Context, dto entities.GetCompanyEmployeesSummaryDTO) (*entities.EmployeesSummary, Error.CodeError) {
	return m.getCompanyEmployeesSummary(ctx, dto)
}
func (m *mockPGCompanyRepo) SetCompanyEmployeeRole(ctx context.Context, dto entities.SetCompanyEmployeeRoleDTO) Error.CodeError {
	return m.setCompanyEmployeeRole(ctx, dto)
}
func (m *mockPGCompanyRepo) RemoveCompanyEmployee(ctx context.Context, dto entities.RemoveCompanyEmployeeDTO) Error.CodeError {
	return m.removeCompanyEmployee(ctx, dto)
}
func (m *mockPGCompanyRepo) CreateDepartment(ctx context.Context, dto entities.CreateDepartment) Error.CodeError {
	return m.createDepartment(ctx, dto)
}
func (m *mockPGCompanyRepo) AddEmployeeToDepartment(ctx context.Context, dto entities.AddEmployeeToDepartmentDTO) Error.CodeError {
	return m.addEmployeeToDepartment(ctx, dto)
}
func (m *mockPGCompanyRepo) GetDepartment(ctx context.Context, dto entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
	return m.getDepartment(ctx, dto)
}
func (m *mockPGCompanyRepo) GetCompanyDepartments(ctx context.Context, dto entities.GetCompanyDepartmentsDTO) ([]*entities.Department, Error.CodeError) {
	return m.getCompanyDepartments(ctx, dto)
}
func (m *mockPGCompanyRepo) UpdateDepartmentTitle(ctx context.Context, dto *entities.UpdateDepartment) Error.CodeError {
	return m.updateDepartmentTitle(ctx, dto)
}
func (m *mockPGCompanyRepo) DeleteDepartment(ctx context.Context, dto entities.DeleteDepartmentDTO) Error.CodeError {
	return m.deleteDepartment(ctx, dto)
}
func (m *mockPGCompanyRepo) RemoveEmployeeFromDepartment(ctx context.Context, dto entities.RemoveEmployeeFromDepartmentDTO) Error.CodeError {
	return m.removeEmployeeFromDepartment(ctx, dto)
}
func (m *mockPGCompanyRepo) GetUserCompanies(ctx context.Context, dto entities.GetUserCompaniesDTO) ([]*entities.GetCompanies, Error.CodeError) {
	return m.getUserCompanies(ctx, dto)
}
func (m *mockPGCompanyRepo) CheckColleagues(ctx context.Context, dto entities.CheckColleaguesDTO) (bool, Error.CodeError) {
	return m.checkColleagues(ctx, dto)
}

// ─── Mock: Redis CompanyRepository ───────────────────────────────────────────

type mockRedisCompanyRepo struct {
	createCompanyJoinCode        func(ctx context.Context, dto entities.CreateCompanyJoinCodeDTO) Error.CodeError
	checkJoinCodeExists          func(ctx context.Context, dto entities.CheckJoinCodeExistsDTO) Error.CodeError
	checkJoinCodeBelongToCompany func(ctx context.Context, dto entities.CheckJoinCodeBelongToCompanyDTO) Error.CodeError
	getCompanyJoinCodes          func(ctx context.Context, dto entities.GetCompanyJoinCodesDTO) ([]string, Error.CodeError)
	getCompanyByJoinCode         func(ctx context.Context, dto entities.GetCompanyByJoinCodeDTO) (string, Error.CodeError)
	deleteCompanyJoinCode        func(ctx context.Context, dto entities.DeleteCompanyJoinCodeDTO) Error.CodeError
}

func (m *mockRedisCompanyRepo) CreateCompanyJoinCode(ctx context.Context, dto entities.CreateCompanyJoinCodeDTO) Error.CodeError {
	return m.createCompanyJoinCode(ctx, dto)
}
func (m *mockRedisCompanyRepo) CheckJoinCodeExists(ctx context.Context, dto entities.CheckJoinCodeExistsDTO) Error.CodeError {
	return m.checkJoinCodeExists(ctx, dto)
}
func (m *mockRedisCompanyRepo) CheckJoinCodeBelongToCompany(ctx context.Context, dto entities.CheckJoinCodeBelongToCompanyDTO) Error.CodeError {
	return m.checkJoinCodeBelongToCompany(ctx, dto)
}
func (m *mockRedisCompanyRepo) GetCompanyJoinCodes(ctx context.Context, dto entities.GetCompanyJoinCodesDTO) ([]string, Error.CodeError) {
	return m.getCompanyJoinCodes(ctx, dto)
}
func (m *mockRedisCompanyRepo) GetCompanyByJoinCode(ctx context.Context, dto entities.GetCompanyByJoinCodeDTO) (string, Error.CodeError) {
	return m.getCompanyByJoinCode(ctx, dto)
}
func (m *mockRedisCompanyRepo) DeleteCompanyJoinCode(ctx context.Context, dto entities.DeleteCompanyJoinCodeDTO) Error.CodeError {
	return m.deleteCompanyJoinCode(ctx, dto)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func newTestService(pgRepo postgresDB.CompanyRepository, redisRepo redisDB.CompanyRepository) *CompanyService {
	db := &postgresDB.DatabaseRepository{Company: pgRepo}
	cache := &redisDB.CacheRepository{Company: redisRepo}
	return NewCompanyService(db, cache)
}

func emptyPGRepo() *mockPGCompanyRepo {
	return &mockPGCompanyRepo{
		checkColleagues: func(_ context.Context, _ entities.CheckColleaguesDTO) (bool, Error.CodeError) {
			return false, Error.CodeError{}
		},
	}
}

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
	pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
		return companyEntity(), ok()
	}
	pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
		return chiefEmployee(), ok()
	}
	return pg
}

// pgRepoWithChiefAndDept настраивает pg-мок для методов департаментов:
// getDepartment возвращает тестовый департамент, а initiator всегда chief
func pgRepoWithChiefAndDept() *mockPGCompanyRepo {
	pg := pgRepoWithChief()
	pg.getDepartment = func(_ context.Context, _ entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
		return departmentEntity(), ok()
	}
	return pg
}

