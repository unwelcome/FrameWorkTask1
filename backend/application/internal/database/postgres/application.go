package postgresDB

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
)

type ApplicationRepository interface {
	CreateApplication(ctx context.Context, dto entities.CreateApplicationDTO) Error.CodeError
	AddApplicationFixLog(ctx context.Context, dto entities.CreateFixLogDTO) Error.CodeError
	GetApplication(ctx context.Context, dto entities.GetApplicationDTO) (*entities.Application, Error.CodeError)
	GetApplicationFixLogs(ctx context.Context, dto entities.GetApplicationFixLogsDTO) ([]*entities.FixLog, Error.CodeError)
	GetApplications(ctx context.Context, dto entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError)
	GetCompanyApplicationStatistic(ctx context.Context, dto entities.GetCompanyApplicationStatisticDTO) (*entities.ApplicationStatistic, Error.CodeError)
	GetEmployeeApplicationStatistic(ctx context.Context, dto entities.GetEmployeeApplicationStatisticDTO) (*entities.ApplicationStatistic, Error.CodeError)
	UpdateApplicationStatus(ctx context.Context, dto entities.UpdateApplicationStatusDTO) Error.CodeError
	AssignApplicationToEmployee(ctx context.Context, dto entities.AssignApplicationToEmployeeDTO) Error.CodeError
	DeleteApplicationRequest(ctx context.Context, dto entities.DeleteApplicationDTO) Error.CodeError
}

type applicationRepository struct {
	db *sql.DB
}

func NewApplicationRepository(db *sql.DB) ApplicationRepository {
	return &applicationRepository{db: db}
}

// CreateApplication Создание заявки
func (r *applicationRepository) CreateApplication(ctx context.Context, dto entities.CreateApplicationDTO) Error.CodeError {
	query := `INSERT INTO applications 
	(uuid, company_uuid, version, title, description, status, created_by) VALUES 
	($1, $2, 1, $3, $4, 'created', $5);`

	_, err := r.db.ExecContext(ctx, query, dto.ApplicationUUID, dto.CompanyUUID, dto.Title, dto.Description, dto.CreatedBy)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	return Error.CodeError{Code: -1, Err: nil}
}

// AddApplicationFixLog Добавление записи в fix log-и заявки
func (r *applicationRepository) AddApplicationFixLog(ctx context.Context, dto entities.CreateFixLogDTO) Error.CodeError {
	query := `INSERT INTO application_fix_logs
	(application_uuid, text, created_by) VALUES 
	($1, $2, $3);`

	_, err := r.db.ExecContext(ctx, query, dto.ApplicationUUID, dto.Text, dto.CreatedBy)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" { // foreign_key_violation
				return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("application not found")}
			}
		}

		return Error.CodeError{Code: 0, Err: err}
	}
	return Error.CodeError{Code: -1, Err: nil}
}

// GetApplication Получение информации по заявке
func (r *applicationRepository) GetApplication(ctx context.Context, dto entities.GetApplicationDTO) (*entities.Application, Error.CodeError) {
	query := `SELECT
			company_uuid,
			title,
			description,
			status,
			created_at,
			created_by,
			COALESCE(closed_at::text, ''),
			COALESCE(managed_by, ''),
			COALESCE(executed_by, '')
		FROM applications
		WHERE uuid = $1;`

	application := &entities.Application{ApplicationUUID: dto.ApplicationUUID}
	err := r.db.QueryRowContext(ctx, query, dto.ApplicationUUID).Scan(
		&application.CompanyUUID,
		&application.Title,
		&application.Description,
		&application.Status,
		&application.CreatedAt,
		&application.CreatedBy,
		&application.ClosedAt,
		&application.ResponsibleManager,
		&application.ResponsibleEngineer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("application not found")}
		}
		return nil, Error.CodeError{Code: 0, Err: err}
	}
	return application, Error.CodeError{Code: -1, Err: nil}
}

// GetApplicationFixLogs Получение всех fix log-ов заявки
func (r *applicationRepository) GetApplicationFixLogs(ctx context.Context, dto entities.GetApplicationFixLogsDTO) ([]*entities.FixLog, Error.CodeError) {
	query := `SELECT 
			id, 
			text, 
			created_at, 
			created_by 
		FROM application_fix_logs 
		WHERE application_uuid = $1;`

	res, err := r.db.QueryContext(ctx, query, dto.ApplicationUUID)
	if err != nil {
		return nil, Error.CodeError{Code: 0, Err: fmt.Errorf("failed get application fix logs")}
	}
	defer res.Close()

	fixLogs := make([]*entities.FixLog, 0)
	for res.Next() {
		item := &entities.FixLog{}
		err = res.Scan(&item.ID, &item.Text, &item.CreatedAt, &item.CreatedBy)
		if err != nil {
			return nil, Error.CodeError{Code: 0, Err: err}
		}
		fixLogs = append(fixLogs, item)
	}

	return fixLogs, Error.CodeError{Code: -1, Err: nil}
}

// GetApplications Получение списка заявок по uuid компании и статусу
func (r *applicationRepository) GetApplications(ctx context.Context, dto entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError) {
	query := `
		SELECT
			uuid,
			version,
			title,
			description,
			created_at,
			created_by,
			COALESCE(closed_at::text, ''),
			COALESCE(managed_by, ''),
			COALESCE(executed_by, '')
		FROM applications
		WHERE company_uuid = $1 AND ($2 = '' OR status = $2)
		ORDER BY created_at DESC, uuid
		OFFSET $3 LIMIT $4;`

	rows, err := r.db.QueryContext(ctx, query, dto.CompanyUUID, dto.Status, dto.Offset, dto.Count)
	if err != nil {
		return nil, Error.CodeError{Code: 0, Err: err}
	}
	defer rows.Close()

	applications := make([]*entities.Application, 0)
	for rows.Next() {
		item := &entities.Application{CompanyUUID: dto.CompanyUUID, Status: dto.Status}
		err = rows.Scan(
			&item.ApplicationUUID,
			&item.Version,
			&item.Title,
			&item.Description,
			&item.CreatedAt,
			&item.CreatedBy,
			&item.ClosedAt,
			&item.ResponsibleManager,
			&item.ResponsibleEngineer)
		if err != nil {
			return nil, Error.CodeError{Code: 0, Err: err}
		}

		applications = append(applications, item)
	}

	return applications, Error.CodeError{Code: -1, Err: nil}
}

// GetCompanyApplicationStatistic Получение статистики компании по заявкам
func (r *applicationRepository) GetCompanyApplicationStatistic(ctx context.Context, dto entities.GetCompanyApplicationStatisticDTO) (*entities.ApplicationStatistic, Error.CodeError) {
	query := `SELECT 
		COUNT(*) FILTER (WHERE status = 'created'),
		COUNT(*) FILTER (WHERE status = 'assigned'),
		COUNT(*) FILTER (WHERE status = 'in_progress'),
		COUNT(*) FILTER (WHERE status = 'on_hold'),
		COUNT(*) FILTER (WHERE status = 'awaiting_approval'),
		COUNT(*) FILTER (WHERE status = 'completed'),
		COUNT(*) FILTER (WHERE status = 'cancelled'),
		COUNT(*) FILTER (WHERE status = 'failed'),
		COUNT(*) FILTER (WHERE status = 'archived')
	FROM applications
	WHERE company_uuid = $1 AND deleted_at IS NULL;`

	statistic := &entities.ApplicationStatistic{}

	err := r.db.QueryRow(query, dto.CompanyUUID).Scan(
		&statistic.Created,
		&statistic.Assigned,
		&statistic.InProgress,
		&statistic.OnHold,
		&statistic.AwaitingApproval,
		&statistic.Completed,
		&statistic.Cancelled,
		&statistic.Failed,
		&statistic.Archived,
	)

	if err != nil {
		return nil, Error.CodeError{Code: 0, Err: err}
	}

	return statistic, Error.CodeError{Code: -1, Err: nil}
}

// GetEmployeeApplicationStatistic Получение статистики работника по заявкам
func (r *applicationRepository) GetEmployeeApplicationStatistic(ctx context.Context, dto entities.GetEmployeeApplicationStatisticDTO) (*entities.ApplicationStatistic, Error.CodeError) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'created' AND created_by = $2) as created,
			COUNT(*) FILTER (WHERE status = 'assigned') as assigned,
			COUNT(*) FILTER (WHERE status = 'in_progress') as in_progress,
			COUNT(*) FILTER (WHERE status = 'on_hold') as on_hold,
			COUNT(*) FILTER (WHERE status = 'awaiting_approval') as awaiting_approval,
			COUNT(*) FILTER (WHERE status = 'completed') as completed,
			COUNT(*) FILTER (WHERE status = 'cancelled') as cancelled,
			COUNT(*) FILTER (WHERE status = 'failed') as failed,
			COUNT(*) FILTER (WHERE status = 'archived') as archived
		FROM applications
		WHERE deleted_at IS NULL 
			AND company_uuid = $1
			AND (created_by = $2 OR managed_by = $2 OR executed_by = $2);`

	statistic := &entities.ApplicationStatistic{}

	err := r.db.QueryRowContext(ctx, query, dto.CompanyUUID, dto.TargetUUID).Scan(
		&statistic.Created,
		&statistic.Assigned,
		&statistic.InProgress,
		&statistic.OnHold,
		&statistic.AwaitingApproval,
		&statistic.Completed,
		&statistic.Cancelled,
		&statistic.Failed,
		&statistic.Archived,
	)
	if err != nil {
		return nil, Error.CodeError{Code: 0, Err: err}
	}

	return statistic, Error.CodeError{Code: -1, Err: nil}
}

// UpdateApplicationStatus Обновление статуса заявки
func (r *applicationRepository) UpdateApplicationStatus(ctx context.Context, dto entities.UpdateApplicationStatusDTO) Error.CodeError {
	query := `UPDATE applications 
	SET 
		version = version + 1,
		status = $2,
		
		-- Логика для managed_by
		managed_by = CASE 
			WHEN $2 IN ('assigned', 'completed', 'cancelled', 'failed') THEN $3 
			ELSE managed_by 
		END,
		
		-- Логика для executed_by
		executed_by = CASE 
			WHEN $2 IN ('in_progress', 'on_hold', 'awaiting_approval') THEN $3 
			ELSE executed_by 
		END,
		
		-- Логика для closed_at
		closed_at = CASE 
			WHEN $2 IN ('completed', 'cancelled', 'failed') THEN CURRENT_TIMESTAMP 
			ELSE NULL
		END
	WHERE uuid = $1 AND deleted_at IS NULL;`

	res, err := r.db.ExecContext(ctx, query, dto.ApplicationUUID, dto.Status, dto.InitiatorUUID)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	if affected == 0 {
		return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("application not found")}
	}

	return Error.CodeError{Code: -1, Err: nil}
}

// AssignApplicationToEmployee Назначение заявки инженеру
func (r *applicationRepository) AssignApplicationToEmployee(ctx context.Context, dto entities.AssignApplicationToEmployeeDTO) Error.CodeError {
	query := `UPDATE applications 
	SET 
		version = version + 1,
		status = 'assigned',
		managed_by = $2,
		executed_by = $3
	WHERE uuid = $1 AND deleted_at IS NOT NULL;`

	res, err := r.db.ExecContext(ctx, query, dto.ApplicationUUID, dto.InitiatorUUID, dto.TargetUUID)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	if affected == 0 {
		return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("application not found")}
	}

	return Error.CodeError{Code: -1, Err: nil}
}

// DeleteApplicationRequest Удаление заявки (заявка помечается удаленной, но не стирается из бд)
func (r *applicationRepository) DeleteApplicationRequest(ctx context.Context, dto entities.DeleteApplicationDTO) Error.CodeError {
	query := `UPDATE applications 
	SET
		deleted_at = CURRENT_TIMESTAMP
	WHERE uuid = $1 AND deleted_at IS NULL;`

	res, err := r.db.ExecContext(ctx, query, dto.ApplicationUUID)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	if affected == 0 {
		return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("application not found")}
	}

	return Error.CodeError{Code: -1, Err: nil}
}
