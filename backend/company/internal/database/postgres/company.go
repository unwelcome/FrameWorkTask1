package postgresDB

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/unwelcome/FrameWorkTask1/backend/company/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
)

type CompanyRepository interface {
	CreateCompany(ctx context.Context, dto *entities.CreateCompany) Error.CodeError
	GetCompany(ctx context.Context, companyUUID string) (*entities.Company, Error.CodeError)
	GetCompanies(ctx context.Context, offset, count int64) ([]*entities.GetCompanies, Error.CodeError)
	GetUserCompanies(ctx context.Context, userUUID string) ([]*entities.GetCompanies, Error.CodeError)
	UpdateCompanyTitle(ctx context.Context, companyUUID, title string) Error.CodeError
	UpdateCompanyStatus(ctx context.Context, companyUUID, status string) Error.CodeError
	DeleteCompany(ctx context.Context, companyUUID string) Error.CodeError
	JoinCompany(ctx context.Context, companyUUID, userUUID string) Error.CodeError
	GetCompanyEmployee(ctx context.Context, companyUUID, userUUID string) (*entities.Employee, Error.CodeError)
	GetCompanyEmployees(ctx context.Context, companyUUID, departmentUUID, role string, offset, count int64) ([]*entities.Employee, Error.CodeError)
	GetCompanyEmployeesSummary(ctx context.Context, companyUUID, departmentUUID string) (*entities.EmployeesSummary, Error.CodeError)
	SetCompanyEmployeeRole(ctx context.Context, companyUUID string, userUUID string, role string) Error.CodeError
	RemoveCompanyEmployee(ctx context.Context, companyUUID, userUUID string) Error.CodeError
	CreateDepartment(ctx context.Context, dto *entities.CreateDepartment) Error.CodeError
	AddEmployeeToDepartment(ctx context.Context, departmentUUID, companyUUID, targetUUID string) Error.CodeError
	GetDepartment(ctx context.Context, departmentUUID string) (*entities.Department, Error.CodeError)
	GetCompanyDepartments(ctx context.Context, companyUUID string, offset, count int64) ([]*entities.Department, Error.CodeError)
	UpdateDepartmentTitle(ctx context.Context, dto *entities.UpdateDepartment) Error.CodeError
	DeleteDepartment(ctx context.Context, departmentUUID string) Error.CodeError
	RemoveEmployeeFromDepartment(ctx context.Context, companyUUID, targetUUID string) Error.CodeError
}

type companyRepository struct {
	db *sql.DB
}

func NewCompanyRepository(db *sql.DB) CompanyRepository {
	return &companyRepository{db: db}
}

// CreateCompany Создание компании
func (r *companyRepository) CreateCompany(ctx context.Context, dto *entities.CreateCompany) Error.CodeError {
	query := `INSERT INTO companies (uuid, title, created_by) VALUES ($1, $2, $3);`

	_, err := r.db.ExecContext(ctx, query, dto.CompanyUUID, dto.Title, dto.CreatedBy)
	if err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// GetCompany Получение данных о компании по uuid
func (r *companyRepository) GetCompany(ctx context.Context, companyUUID string) (*entities.Company, Error.CodeError) {
	query := `SELECT title, status, created_by, created_at FROM companies WHERE uuid = $1;`

	company := &entities.Company{
		CompanyUUID: companyUUID,
	}

	err := r.db.QueryRowContext(ctx, query, companyUUID).Scan(&company.Title, &company.Status, &company.CreatedBy, &company.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, Error.Public(codes.NotFound, "company not found")
		}
		return nil, Error.Internal(err)
	}
	return company, Error.CodeError{}
}

// GetCompanies Получение списка компаний размера count со сдвигом offset
func (r *companyRepository) GetCompanies(ctx context.Context, offset, count int64) ([]*entities.GetCompanies, Error.CodeError) {
	query := `SELECT uuid, title, status FROM companies ORDER BY created_at DESC, uuid OFFSET $1 LIMIT $2;`

	// Получение компаний
	res, err := r.db.QueryContext(ctx, query, offset, count)
	if err != nil {
		return nil, Error.Internal(err)
	}
	defer res.Close()

	// Маппинг ответа в структуру
	companies := make([]*entities.GetCompanies, 0)
	for res.Next() {
		company := &entities.GetCompanies{}
		err = res.Scan(&company.CompanyUUID, &company.Title, &company.Status)
		if err != nil {
			return nil, Error.Internal(err)
		}

		companies = append(companies, company)
	}

	if err = res.Err(); err != nil {
		return nil, Error.Internal(err)
	}

	return companies, Error.CodeError{}
}

// GetUserCompanies Получение списка компаний, в которых состоит пользователь
func (r *companyRepository) GetUserCompanies(ctx context.Context, userUUID string) ([]*entities.GetCompanies, Error.CodeError) {
	query := `SELECT c.uuid, c.title, c.status FROM companies c JOIN employees e ON c.uuid = e.company_uuid WHERE e.user_uuid = $1 ORDER BY c.created_at DESC;`

	res, err := r.db.QueryContext(ctx, query, userUUID)
	if err != nil {
		return nil, Error.Internal(err)
	}
	defer res.Close()

	companies := make([]*entities.GetCompanies, 0)
	for res.Next() {
		company := &entities.GetCompanies{}
		err = res.Scan(&company.CompanyUUID, &company.Title, &company.Status)
		if err != nil {
			return nil, Error.Internal(err)
		}
		companies = append(companies, company)
	}

	if err = res.Err(); err != nil {
		return nil, Error.Internal(err)
	}

	return companies, Error.CodeError{}
}

// UpdateCompanyTitle Обновление названия компании
func (r *companyRepository) UpdateCompanyTitle(ctx context.Context, companyUUID, title string) Error.CodeError {
	query := `UPDATE companies SET title = $2 WHERE uuid = $1;`

	// Обновление названия компании
	res, err := r.db.ExecContext(ctx, query, companyUUID, title)
	if err != nil {
		return Error.Internal(err)
	}

	// Проверка выполнения запроса
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return Error.Internal(err)
	}

	if rowsAffected == 0 {
		return Error.Public(codes.NotFound, "company not found")
	}
	return Error.CodeError{}
}

// UpdateCompanyStatus Обновление статуса компании (open | close)
func (r *companyRepository) UpdateCompanyStatus(ctx context.Context, companyUUID, status string) Error.CodeError {
	query := `UPDATE companies SET status = $2 WHERE uuid = $1;`

	res, err := r.db.ExecContext(ctx, query, companyUUID, status)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Неверное значение enum
			if pqErr.Code == "22P02" {
				return Error.Public(codes.InvalidArgument, "invalid status value")
			}
		}
		return Error.Internal(err)
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		return Error.Internal(err)
	}

	if affectedRows == 0 {
		return Error.Public(codes.NotFound, "company not found")
	}
	return Error.CodeError{}
}

// DeleteCompany Удаление компании
func (r *companyRepository) DeleteCompany(ctx context.Context, companyUUID string) Error.CodeError {
	query := `DELETE FROM companies WHERE uuid = $1;`

	res, err := r.db.ExecContext(ctx, query, companyUUID)
	if err != nil {
		return Error.Internal(err)
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		return Error.Internal(err)
	}

	if affectedRows == 0 {
		return Error.Public(codes.NotFound, "company not found")
	}
	return Error.CodeError{}
}

// JoinCompany Добавление пользователя в список сотрудников компании
func (r *companyRepository) JoinCompany(ctx context.Context, companyUUID, userUUID string) Error.CodeError {
	query := `INSERT INTO employees (company_uuid, user_uuid) VALUES ($1, $2);`

	_, err := r.db.ExecContext(ctx, query, companyUUID, userUUID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return Error.Public(codes.AlreadyExists, "user already in company")
			}
			if pqErr.Code == "23503" { // foreign_key_violation
				return Error.Public(codes.NotFound, "company not found")
			}
		}
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// GetCompanyEmployee Получение данных о сотруднике в компании (ошибка если сотрудника нет)
func (r *companyRepository) GetCompanyEmployee(ctx context.Context, companyUUID, userUUID string) (*entities.Employee, Error.CodeError) {
	query := `SELECT role, COALESCE(department_uuid::text, ''), joined_at FROM employees WHERE company_uuid = $1 AND user_uuid = $2;`

	employee := &entities.Employee{
		CompanyUUID: companyUUID,
		UserUUID:    userUUID,
	}

	err := r.db.QueryRowContext(ctx, query, companyUUID, userUUID).Scan(&employee.Role, &employee.DepartmentUUID, &employee.JoinedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, Error.Public(codes.NotFound, "user not in company")
		}
		return nil, Error.Internal(err)
	}
	return employee, Error.CodeError{}
}

// GetCompanyEmployees Возвращает сотрудников компании (сортировка по role, departmentUUID и ограничения через offset и count)
func (r *companyRepository) GetCompanyEmployees(ctx context.Context, companyUUID, departmentUUID, role string, offset, count int64) ([]*entities.Employee, Error.CodeError) {
	query := `SELECT 
		user_uuid, 
		role, 
		department_uuid,
		joined_at 
	FROM employees 
	WHERE company_uuid = $1 AND ($2 = '' OR role = $2) AND ($3 = '' OR department_uuid = $3)
	ORDER BY joined_at DESC 
	OFFSET $4 LIMIT $5;`

	rows, err := r.db.QueryContext(ctx, query, companyUUID, role, departmentUUID, offset, count)
	if err != nil {
		return nil, Error.Internal(err)
	}
	defer rows.Close()

	employees := make([]*entities.Employee, 0)
	for rows.Next() {
		employee := &entities.Employee{}
		err = rows.Scan(&employee.UserUUID, &employee.Role, &employee.DepartmentUUID, &employee.JoinedAt)
		if err != nil {
			return nil, Error.Internal(err)
		}

		employees = append(employees, employee)
	}

	if err = rows.Err(); err != nil {
		return nil, Error.Internal(err)
	}

	return employees, Error.CodeError{}
}

// GetCompanyEmployeesSummary Получение кол-ва сотрудников по ролям в компании
func (r *companyRepository) GetCompanyEmployeesSummary(ctx context.Context, companyUUID, departmentUUID string) (*entities.EmployeesSummary, Error.CodeError) {
	query := `SELECT
    	COUNT(CASE WHEN role = 'unemployed' THEN 1 END) as unemployed_count,
    	COUNT(CASE WHEN role = 'inspector' THEN 1 END) as inspector_count,
    	COUNT(CASE WHEN role = 'engineer' THEN 1 END) as engineer_count,
    	COUNT(CASE WHEN role = 'manager' THEN 1 END) as manager_count,
    	COUNT(CASE WHEN role = 'analytic' THEN 1 END) as analytic_count,
    	COUNT(CASE WHEN role = 'chief' THEN 1 END) as chief_count
	FROM employees
	WHERE company_uuid = $1 AND ($2 = '' OR department_uuid = $2);`

	employeeSummary := &entities.EmployeesSummary{
		CompanyUUID: companyUUID,
	}

	err := r.db.QueryRowContext(ctx, query, companyUUID, departmentUUID).
		Scan(
			&employeeSummary.UnemployedCount,
			&employeeSummary.InspectorCount,
			&employeeSummary.EngineerCount,
			&employeeSummary.ManagerCount,
			&employeeSummary.AnalyticCount,
			&employeeSummary.ChiefCount)

	if err != nil {
		return nil, Error.Internal(err)
	}
	return employeeSummary, Error.CodeError{}
}

// SetCompanyEmployeeRole Устанавливает новую роль сотруднику компании
func (r *companyRepository) SetCompanyEmployeeRole(ctx context.Context, companyUUID string, userUUID string, role string) Error.CodeError {
	query := `UPDATE employees SET role = $3 WHERE company_uuid = $1 AND user_uuid = $2;`

	res, err := r.db.ExecContext(ctx, query, companyUUID, userUUID, role)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Неверное значение enum
			if pqErr.Code == "22P02" {
				return Error.Public(codes.InvalidArgument, "invalid role value")
			}
		}
		return Error.Internal(err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return Error.Internal(err)
	}

	if rowsAffected == 0 {
		return Error.Public(codes.NotFound, "employee not found")
	}

	return Error.CodeError{}
}

// RemoveCompanyEmployee Удаление пользователя из списка сотрудников компании
func (r *companyRepository) RemoveCompanyEmployee(ctx context.Context, companyUUID, userUUID string) Error.CodeError {
	query := `DELETE FROM employees WHERE company_uuid = $1 AND user_uuid = $2;`

	res, err := r.db.ExecContext(ctx, query, companyUUID, userUUID)
	if err != nil {
		return Error.Internal(err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return Error.Internal(err)
	}

	if rowsAffected == 0 {
		return Error.Public(codes.NotFound, "user not in company")
	}
	return Error.CodeError{}
}

// CreateDepartment - Создание департамента
func (r *companyRepository) CreateDepartment(ctx context.Context, dto *entities.CreateDepartment) Error.CodeError {
	query := `INSERT INTO departments (uuid, company_uuid, title, created_by) VALUES ($1, $2, $3, $4);`

	_, err := r.db.ExecContext(ctx, query, dto.UUID, dto.CompanyUUID, dto.Title, dto.CreatedBy)
	if err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// AddEmployeeToDepartment - Добавление сотрудника в департамент
func (r *companyRepository) AddEmployeeToDepartment(ctx context.Context, departmentUUID, companyUUID, targetUUID string) Error.CodeError {
	query := `UPDATE employees SET department_uuid = $3 WHERE company_uuid = $1 AND user_uuid = $2;`

	res, err := r.db.ExecContext(ctx, query, companyUUID, targetUUID, departmentUUID)
	if err != nil {
		return Error.Internal(err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return Error.Internal(err)
	}

	if rowsAffected == 0 {
		return Error.Public(codes.NotFound, "employee not found")
	}

	return Error.CodeError{}
}

// GetDepartment - Получение полной информации о департаменте
func (r *companyRepository) GetDepartment(ctx context.Context, departmentUUID string) (*entities.Department, Error.CodeError) {
	query := `SELECT company_uuid, title, created_at, created_by FROM departments WHERE uuid = $1;`

	department := &entities.Department{
		UUID: departmentUUID,
	}

	err := r.db.QueryRowContext(ctx, query, departmentUUID).Scan(&department.CompanyUUID, &department.Title, &department.CreatedAt, &department.CreatedBy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, Error.Public(codes.NotFound, "department not found")
		}
		return nil, Error.Internal(err)
	}
	return department, Error.CodeError{}
}

// GetCompanyDepartments - Получение списка департаментов организации с фильтрацией (offset и count)
func (r *companyRepository) GetCompanyDepartments(ctx context.Context, companyUUID string, offset, count int64) ([]*entities.Department, Error.CodeError) {
	query := `SELECT uuid, title FROM departments WHERE company_uuid = $1 ORDER BY created_at DESC OFFSET $2 LIMIT $3;`

	rows, err := r.db.QueryContext(ctx, query, companyUUID, offset, count)
	if err != nil {
		return nil, Error.Internal(err)
	}
	defer rows.Close()

	departments := make([]*entities.Department, 0)
	for rows.Next() {
		department := &entities.Department{}
		err = rows.Scan(&department.UUID, &department.Title)
		if err != nil {
			return nil, Error.Internal(err)
		}

		departments = append(departments, department)
	}

	if err = rows.Err(); err != nil {
		return nil, Error.Internal(err)
	}

	return departments, Error.CodeError{}
}

// UpdateDepartmentTitle - Обновление названия департамента
func (r *companyRepository) UpdateDepartmentTitle(ctx context.Context, dto *entities.UpdateDepartment) Error.CodeError {
	query := `UPDATE departments SET title = $1 WHERE uuid = $2;`

	res, err := r.db.ExecContext(ctx, query, dto.Title, dto.UUID)
	if err != nil {
		return Error.Internal(err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return Error.Internal(err)
	}

	if rowsAffected == 0 {
		return Error.Public(codes.NotFound, "department not found")
	}

	return Error.CodeError{}
}

// DeleteDepartment - Удаление департамента
func (r *companyRepository) DeleteDepartment(ctx context.Context, departmentUUID string) Error.CodeError {
	query := `DELETE FROM departments WHERE uuid = $1;`

	res, err := r.db.ExecContext(ctx, query, departmentUUID)
	if err != nil {
		return Error.Internal(err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return Error.Internal(err)
	}

	if rowsAffected == 0 {
		return Error.Public(codes.NotFound, "department not found")
	}

	return Error.CodeError{}
}

// RemoveEmployeeFromDepartment - Удаление сотрудника из департамента
func (r *companyRepository) RemoveEmployeeFromDepartment(ctx context.Context, companyUUID, targetUUID string) Error.CodeError {
	query := `UPDATE employees SET department_uuid = NULL WHERE company_uuid = $1 AND user_uuid = $2;`

	res, err := r.db.ExecContext(ctx, query, companyUUID, targetUUID)
	if err != nil {
		return Error.Internal(err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return Error.Internal(err)
	}

	if rowsAffected == 0 {
		return Error.Public(codes.NotFound, "employee not found")
	}

	return Error.CodeError{}
}
