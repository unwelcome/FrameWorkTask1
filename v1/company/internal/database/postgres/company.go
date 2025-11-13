package postgresDB

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/unwelcome/FrameWorkTask1/v1/company/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/v1/company/pkg/errors"
	"google.golang.org/grpc/codes"
)

type CompanyRepository interface {
	CreateCompany(ctx context.Context, dto *entities.CreateCompany) *Error.CodeError
	GetCompany(ctx context.Context, company_uuid string) (*entities.Company, *Error.CodeError)
	GetCompanies(ctx context.Context, offset, count int) ([]*entities.GetCompanies, *Error.CodeError)
	UpdateCompanyTitle(ctx context.Context, company_uuid, title string) *Error.CodeError
	UpdateCompanyStatus(ctx context.Context, company_uuid, status string) *Error.CodeError
	DeleteCompany(ctx context.Context, companu_uuid string) *Error.CodeError
	JoinCompany(ctx context.Context, company_uuid, user_uuid string) *Error.CodeError
	GetCompanyEmployee(ctx context.Context, company_uuid, user_uuid string) (*entities.Employee, *Error.CodeError)
	GetCompanyEmployeesSummary(ctx context.Context, company_uuid string) (*entities.EmployeesSummary, *Error.CodeError)
	RemoveCompanyEmployee(ctx context.Context, company_uuid, user_uuid string) *Error.CodeError
}

type companyRepository struct {
	db *sql.DB
}

func NewCompanyRepository(db *sql.DB) CompanyRepository {
	return &companyRepository{db: db}
}

// CreateCompany Создание компании
func (r *companyRepository) CreateCompany(ctx context.Context, dto *entities.CreateCompany) *Error.CodeError {
	query := `INSERT INTO companies (uuid, title, created_by) VALUES ($1, $2, $3);`

	_, err := r.db.ExecContext(ctx, query, dto.CompanyUUID, dto.Title, dto.CreatedBy)
	if err != nil {
		return &Error.CodeError{Code: 0, Err: err}
	}
	return &Error.CodeError{Code: -1, Err: nil}
}

// GetCompany Получение данных о компании по uuid
func (r *companyRepository) GetCompany(ctx context.Context, company_uuid string) (*entities.Company, *Error.CodeError) {
	query := `SELECT title, status, created_by, created_at FROM companies WHERE uuid = $1;`

	company := &entities.Company{
		CompanyUUID: company_uuid,
	}

	err := r.db.QueryRowContext(ctx, query, company_uuid).Scan(&company.Title, &company.Status, &company.CreatedBy, &company.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("company not found")}
		}
		return nil, &Error.CodeError{Code: 0, Err: err}
	}
	return company, &Error.CodeError{Code: -1, Err: nil}
}

// GetCompanies Получение списка компаний размера count со сдвигом offset
func (r *companyRepository) GetCompanies(ctx context.Context, offset, count int) ([]*entities.GetCompanies, *Error.CodeError) {
	query := `SELECT uuid, title, status FROM companies OFFSET $1 LIMIT $2;`

	// Получение компаний
	res, err := r.db.QueryContext(ctx, query, offset, count)
	if err != nil {
		return nil, &Error.CodeError{Code: 0, Err: err}
	}
	defer res.Close()

	// Маппинг ответа в структуру
	companies := make([]*entities.GetCompanies, 0)
	for res.Next() {
		company := &entities.GetCompanies{}
		err := res.Scan(&company.CompanyUUID, &company.Title, &company.Status)
		if err != nil {
			return nil, &Error.CodeError{Code: 0, Err: nil}
		}

		companies = append(companies, company)
	}

	return companies, &Error.CodeError{Code: -1, Err: nil}
}

// UpdateCompanyTitle Обновление названия компании
func (r *companyRepository) UpdateCompanyTitle(ctx context.Context, company_uuid, title string) *Error.CodeError {
	query := `UPDATE companies SET title = $2 WHERE uuid = $1;`

	// Обновление названия компании
	res, err := r.db.ExecContext(ctx, query, company_uuid, title)
	if err != nil {
		return &Error.CodeError{Code: 0, Err: err}
	}

	// Проверка выполнения запроса
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return &Error.CodeError{Code: 0, Err: err}
	}

	if rowsAffected == 0 {
		return &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("company not found")}
	}
	return &Error.CodeError{Code: -1, Err: nil}
}

// UpdateCompanyStatus Обновление статуса компании (open | close)
func (r *companyRepository) UpdateCompanyStatus(ctx context.Context, company_uuid, status string) *Error.CodeError {
	query := `UPDATE companies SET status = $2 WHERE uuid = $1;`

	res, err := r.db.ExecContext(ctx, query, company_uuid, status)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Неверное значение enum
			if pqErr.Code == "22P02" {
				return &Error.CodeError{Code: int(codes.InvalidArgument), Err: fmt.Errorf("invalid status value")}
			}
		}
		return &Error.CodeError{Code: 0, Err: err}
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		return &Error.CodeError{Code: 0, Err: err}
	}

	if affectedRows == 0 {
		return &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("company not found")}
	}
	return &Error.CodeError{Code: -1, Err: nil}
}

// DeleteCompany Удаление компании
func (r *companyRepository) DeleteCompany(ctx context.Context, companu_uuid string) *Error.CodeError {
	query := `DELETE FROM companies WHERE uuid = $1;`

	res, err := r.db.ExecContext(ctx, query, companu_uuid)
	if err != nil {
		return &Error.CodeError{Code: 0, Err: err}
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		return &Error.CodeError{Code: 0, Err: err}
	}

	if affectedRows == 0 {
		return &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("company not found")}
	}
	return &Error.CodeError{Code: -1, Err: nil}
}

// JoinCompany Добавление пользователя в список сотрудников компании
func (r *companyRepository) JoinCompany(ctx context.Context, company_uuid, user_uuid string) *Error.CodeError {
	query := `INSERT INTO employees (company_uuid, user_uuid) VALUES ($1, $2);`

	_, err := r.db.ExecContext(ctx, query, company_uuid, user_uuid)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return &Error.CodeError{Code: int(codes.AlreadyExists), Err: fmt.Errorf("user already in company")}
			}
			if pqErr.Code == "23503" { // foreign_key_violation
				return &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("company not found")}
			}
		}
		return &Error.CodeError{Code: 0, Err: err}
	}
	return &Error.CodeError{Code: -1, Err: nil}
}

// GetCompanyEmployee Получение данных о сотруднике в компании (ошибка если сотрудника нет)
func (r *companyRepository) GetCompanyEmployee(ctx context.Context, company_uuid, user_uuid string) (*entities.Employee, *Error.CodeError) {
	query := `SELECT role, joined_at FROM employees WHERE company_uuid = $1 AND user_uuid = $2;`

	employee := &entities.Employee{
		CompanyUUID: company_uuid,
		UserUUID:    user_uuid,
	}

	err := r.db.QueryRowContext(ctx, query, company_uuid, user_uuid).Scan(&employee.Role, &employee.JoinedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("user not in company")}
		}
		return nil, &Error.CodeError{Code: 0, Err: err}
	}
	return employee, &Error.CodeError{Code: -1, Err: nil}
}

// GetCompanyEmployeesSummary Получение кол-ва сотрудников по ролям в компании
func (r *companyRepository) GetCompanyEmployeesSummary(ctx context.Context, company_uuid string) (*entities.EmployeesSummary, *Error.CodeError) {
	query := `SELECT 
    COUNT(CASE WHEN role = 'unemployed' THEN 1 END) as unemployed_count,
    COUNT(CASE WHEN role = 'engineer' THEN 1 END) as engineer_count,
    COUNT(CASE WHEN role = 'manager' THEN 1 END) as manager_count,
    COUNT(CASE WHEN role = 'analytic' THEN 1 END) as analytic_count,
    COUNT(CASE WHEN role = 'chief' THEN 1 END) as chief_count
	FROM employees 
	WHERE company_uuid = $1;`

	employeeSummary := &entities.EmployeesSummary{
		CompanyUUID: company_uuid,
	}

	err := r.db.QueryRowContext(ctx, query, company_uuid).
		Scan(
			&employeeSummary.UnemployedCount,
			&employeeSummary.EngineerCount,
			&employeeSummary.ManagerCount,
			&employeeSummary.AnalyticCount,
			&employeeSummary.ChiefCount)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("company not found")}
		}
		return nil, &Error.CodeError{Code: 0, Err: err}
	}
	return employeeSummary, &Error.CodeError{Code: -1, Err: nil}
}

// RemoveCompanyEmployee Удаление пользователя из списка сотрудников компании
func (r *companyRepository) RemoveCompanyEmployee(ctx context.Context, company_uuid, user_uuid string) *Error.CodeError {
	query := `DELETE FROM employees WHERE company_uuid = $1 AND user_uuid = $2;`

	res, err := r.db.ExecContext(ctx, query, company_uuid, user_uuid)
	if err != nil {
		return &Error.CodeError{Code: 0, Err: err}
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return &Error.CodeError{Code: 0, Err: err}
	}

	if rowsAffected == 0 {
		return &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("user not in company")}
	}
	return &Error.CodeError{Code: -1, Err: nil}
}
