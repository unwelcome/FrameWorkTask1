package postgresDB

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
)

// 'created', 'assigned', 'in_progress', 'on_hold', 'completed', 'failed', 'redirected', 'rejected', 'recalled', 'pending_verification', 'on_verification', 'on_revision'

type ApplicationRepository interface {
	CreateApplication(ctx context.Context, dto entities.CreateApplicationDTO) Error.CodeError
	AddApplicationFixLog(ctx context.Context, dto entities.AddFixLogDTO) Error.CodeError
	GetApplication(ctx context.Context, dto entities.GetApplicationDTO) (*entities.Application, Error.CodeError)
	GetApplicationFixLogs(ctx context.Context, dto entities.GetApplicationFixLogsDTO) ([]*entities.FixLog, Error.CodeError)
	GetApplications(ctx context.Context, dto entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError)
	UpdateApplicationStatus(ctx context.Context, dto entities.UpdateApplicationStatusDTO) Error.CodeError
	AssignApplicationToEmployee(ctx context.Context, dto entities.AssignApplicationDTO) Error.CodeError
	RedirectApplication(ctx context.Context, dto entities.RedirectApplicationDTO) Error.CodeError
	RecallApplication(ctx context.Context, dto entities.RecallApplicationDTO) Error.CodeError
	TakeApplicationToVerification(ctx context.Context, dto entities.TakeApplicationToVerificationDTO) Error.CodeError
	ReleaseApplicationVerification(ctx context.Context, dto entities.ReleaseApplicationVerificationDTO) Error.CodeError
	DeleteApplication(ctx context.Context, dto entities.DeleteApplicationDTO) Error.CodeError
}

type applicationRepository struct {
	db *sql.DB
}

func NewApplicationRepository(db *sql.DB) ApplicationRepository {
	return &applicationRepository{db: db}
}

// saveVersion сохраняет снапшот текущего состояния заявки в application_versions внутри транзакции.
// Использует SELECT FOR UPDATE, чтобы заблокировать строку на время транзакции.
func (r *applicationRepository) saveVersion(ctx context.Context, tx *sql.Tx, applicationUUID string) error {
	var app entities.Application
	err := tx.QueryRowContext(ctx, `
		SELECT
			uuid, 
			company_uuid, 
			department_uuid, 
			version, 
			title, 
			description, 
			status,
			revision_count,
			created_at::text, 
			created_by,
		    COALESCE(updated_at::text, ''),
			COALESCE(updated_by, ''),
			COALESCE(managed_by::text, ''),
			COALESCE(executed_by::text, ''),
			COALESCE(inspected_by::text, ''),
			COALESCE(closed_at::text, ''),
			COALESCE(deleted_at::text, ''),
			COALESCE(deleted_by::text, '')
		FROM applications
		WHERE uuid = $1
		FOR UPDATE`,
		applicationUUID,
	).Scan(
		&app.ApplicationUUID,
		&app.CompanyUUID,
		&app.DepartmentUUID,
		&app.Version,
		&app.Title,
		&app.Description,
		&app.Status,
		&app.RevisionCount,
		&app.CreatedAt,
		&app.CreatedBy,
		&app.UpdatedAt,
		&app.UpdatedBy,
		&app.ManagedBy,
		&app.ExecutedBy,
		&app.InspectedBy,
		&app.ClosedAt,
		&app.DeletedAt,
		&app.DeletedBy,
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(app)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO application_versions (application_uuid, version, body) VALUES ($1, $2, $3)`,
		app.ApplicationUUID, app.Version, body,
	)
	return err
}

// CreateApplication Создание заявки
// 'created'				- inspector
func (r *applicationRepository) CreateApplication(ctx context.Context, dto entities.CreateApplicationDTO) Error.CodeError {
	query := `INSERT INTO applications
	(uuid, 
	 company_uuid, 
	 department_uuid, 
	 version, 
	 title, 
	 description, 
	 status, 
	 revision_count,
	 created_by
	 ) VALUES
	($1, $2, $3, 1, $4, $5, 'created', 0, $6);`

	_, err := r.db.ExecContext(ctx, query, dto.ApplicationUUID, dto.CompanyUUID, dto.DepartmentUUID, dto.Title, dto.Description, dto.CreatedBy)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	return Error.CodeError{Code: -1, Err: nil}
}

// AddApplicationFixLog Добавление записи в fix log-и заявки
func (r *applicationRepository) AddApplicationFixLog(ctx context.Context, dto entities.AddFixLogDTO) Error.CodeError {
	query := `INSERT INTO application_fix_logs
	(uuid, application_uuid, text, created_by) VALUES
	($1, $2, $3, $4);`

	_, err := r.db.ExecContext(ctx, query, uuid.New().String(), dto.ApplicationUUID, dto.Text, dto.CreatedBy)
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
			department_uuid,
			version,
			title,
			description,
			status,
			revision_count,
			created_at::text,
			created_by,
			COALESCE(updated_at::text, ''),
			COALESCE(updated_by, ''),
			COALESCE(managed_by, ''),
			COALESCE(executed_by, ''),
			COALESCE(inspected_by, ''),
			COALESCE(closed_at::text, ''),
			COALESCE(deleted_at::text, ''),
			COALESCE(deleted_by, '')
		FROM applications
		WHERE uuid = $1;`

	application := &entities.Application{ApplicationUUID: dto.ApplicationUUID}
	err := r.db.QueryRowContext(ctx, query, dto.ApplicationUUID).Scan(
		&application.CompanyUUID,
		&application.DepartmentUUID,
		&application.Version,
		&application.Title,
		&application.Description,
		&application.Status,
		&application.RevisionCount,
		&application.CreatedAt,
		&application.CreatedBy,
		&application.UpdatedAt,
		&application.UpdatedBy,
		&application.ManagedBy,
		&application.ExecutedBy,
		&application.InspectedBy,
		&application.ClosedAt,
		&application.DeletedAt,
		&application.DeletedBy,
	)
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
			uuid,
			text,
			created_at::text,
			created_by
		FROM application_fix_logs
		WHERE application_uuid = $1;`

	rows, err := r.db.QueryContext(ctx, query, dto.ApplicationUUID)
	if err != nil {
		return nil, Error.CodeError{Code: 0, Err: fmt.Errorf("failed get application fix logs")}
	}
	defer rows.Close()

	fixLogs := make([]*entities.FixLog, 0)
	for rows.Next() {
		item := &entities.FixLog{}
		err = rows.Scan(&item.UUID, &item.Text, &item.CreatedAt, &item.CreatedBy)
		if err != nil {
			return nil, Error.CodeError{Code: 0, Err: err}
		}
		fixLogs = append(fixLogs, item)
	}

	if err = rows.Err(); err != nil {
		return nil, Error.CodeError{Code: 0, Err: err}
	}

	return fixLogs, Error.CodeError{Code: -1, Err: nil}
}

// GetApplications Получение списка заявок по uuid компании с сортировкой по статусу и департаменту с offset и count (отдельно удаленные заявки)
func (r *applicationRepository) GetApplications(ctx context.Context, dto entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError) {
	query := `
		SELECT
			uuid,
			department_uuid,
			version,
			title,
			description,
			status,
			revision_count,
			created_at::text,
			created_by,
			COALESCE(updated_at::text, ''),
			COALESCE(updated_by, ''),
			COALESCE(managed_by, ''),
			COALESCE(executed_by, ''),
			COALESCE(inspected_by, ''),
			COALESCE(closed_at::text, ''),
			COALESCE(deleted_at::text, ''),
			COALESCE(deleted_by, '')
		FROM applications
		WHERE company_uuid = $1 
		  	AND (ARRAY_LENGTH($2::text[], 1) IS NULL OR status = ANY($2::text[])) 
		  	AND ($3 = '' OR created_by = $3)
			AND ($4 = '' OR managed_by = $4)
			AND ($5 = '' OR executed_by = $5)
			AND ($6 = '' OR inspected_by = $6)
			AND (NOT $7 OR executed_by IS NULL)
		  	AND ($8 = '' OR department_uuid = $8) 
		  	AND ((NOT $9 AND deleted_at IS NULL) OR ($9 AND deleted_at IS NOT NULL))
		ORDER BY created_at DESC, uuid
		OFFSET $10 LIMIT $11;`

	rows, err := r.db.QueryContext(ctx, query,
		dto.CompanyUUID,        // 1
		pq.Array(dto.Statuses), // 2
		dto.CreatedBy,          // 3
		dto.ManagedBy,          // 4
		dto.ExecutedBy,         // 5
		dto.InspectedBy,        // 6
		dto.ExecutedByIsNull,   // 7
		dto.DepartmentUUID,     // 8
		dto.IsDeleted,          // 9
		dto.Offset,             // 10
		dto.Count,              // 11
	)
	if err != nil {
		return nil, Error.CodeError{Code: 0, Err: err}
	}
	defer rows.Close()

	applications := make([]*entities.Application, 0)
	for rows.Next() {
		application := &entities.Application{CompanyUUID: dto.CompanyUUID}
		err = rows.Scan(
			&application.ApplicationUUID,
			&application.DepartmentUUID,
			&application.Version,
			&application.Title,
			&application.Description,
			&application.Status,
			&application.RevisionCount,
			&application.CreatedAt,
			&application.CreatedBy,
			&application.UpdatedAt,
			&application.UpdatedBy,
			&application.ManagedBy,
			&application.ExecutedBy,
			&application.InspectedBy,
			&application.ClosedAt,
			&application.DeletedAt,
			&application.DeletedBy,
		)
		if err != nil {
			return nil, Error.CodeError{Code: 0, Err: err}
		}

		applications = append(applications, application)
	}

	if err = rows.Err(); err != nil {
		return nil, Error.CodeError{Code: 0, Err: err}
	}

	return applications, Error.CodeError{Code: -1, Err: nil}
}

// UpdateApplicationStatus Обновление статуса заявки
// 'rejected' 				- manager
// 'in_progress'  			- engineer
// 'on_hold' 				- engineer
// 'pending_verification' 	- engineer
// 'completed'	 			- inspector
// 'failed'					- inspector
// 'on_revision'			- inspector
func (r *applicationRepository) UpdateApplicationStatus(ctx context.Context, dto entities.UpdateApplicationStatusDTO) Error.CodeError {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	defer tx.Rollback()

	if err := r.saveVersion(ctx, tx, dto.ApplicationUUID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("application not found")}
		}
		return Error.CodeError{Code: 0, Err: err}
	}

	query := `UPDATE applications
	SET
		version = version + 1,
		status = $2,
		
		revision_count = CASE
		    WHEN $2 = 'on_revision' THEN revision_count + 1
	        ELSE revision_count 
		END,
		
		updated_at = CURRENT_TIMESTAMP,
		updated_by = $3,

		managed_by = CASE
		    WHEN $4 THEN NULL
			WHEN $2 = 'rejected' THEN $3
			ELSE managed_by
		END,
	    
	    executed_by = CASE
	        WHEN $5 THEN NULL
	        ELSE executed_by
	    END,
	    
	    inspected_by = NULL,

		closed_at = CASE
			WHEN $2 IN ('completed', 'failed') THEN CURRENT_TIMESTAMP
			ELSE NULL
		END
	WHERE uuid = $1 AND deleted_at IS NULL;`

	res, err := tx.ExecContext(ctx, query, dto.ApplicationUUID, dto.Status, dto.InitiatorUUID, dto.DropManagedBy, dto.DropExecutedBy)
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

	if err = tx.Commit(); err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	return Error.CodeError{Code: -1, Err: nil}
}

// AssignApplicationToEmployee Назначение заявки инженеру
// 'assigned'				- manager
func (r *applicationRepository) AssignApplicationToEmployee(ctx context.Context, dto entities.AssignApplicationDTO) Error.CodeError {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	defer tx.Rollback()

	if err := r.saveVersion(ctx, tx, dto.ApplicationUUID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("application not found")}
		}
		return Error.CodeError{Code: 0, Err: err}
	}

	query := `UPDATE applications
	SET
		version = version + 1,
		status = 'assigned',
		updated_at = CURRENT_TIMESTAMP,
		updated_by = $2,
		managed_by = $2,
		executed_by = $3
	WHERE uuid = $1 AND deleted_at IS NULL;`

	res, err := tx.ExecContext(ctx, query, dto.ApplicationUUID, dto.InitiatorUUID, dto.TargetUUID)
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

	if err = tx.Commit(); err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	return Error.CodeError{Code: -1, Err: nil}
}

// RedirectApplication Передача заявки в другой департамент
// 'redirected'				- manager
func (r *applicationRepository) RedirectApplication(ctx context.Context, dto entities.RedirectApplicationDTO) Error.CodeError {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	defer tx.Rollback()

	if err := r.saveVersion(ctx, tx, dto.ApplicationUUID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("application not found")}
		}
		return Error.CodeError{Code: 0, Err: err}
	}

	query := `UPDATE applications
	SET 
	    department_uuid = $2,
	    version = version + 1,
	    status = 'redirected',
	    updated_at = CURRENT_TIMESTAMP,
	    updated_by = $3,
	    managed_by = $3,
	    executed_by = NULL,
	    inspected_by = NULL
	WHERE uuid = $1 AND deleted_at IS NULL;`

	res, err := tx.ExecContext(ctx, query, dto.ApplicationUUID, dto.TargetDepartmentUUID, dto.InitiatorUUID)
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

	if err = tx.Commit(); err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	return Error.CodeError{Code: -1, Err: nil}
}

// RecallApplication Отзыв заявки у инженера
// 'recalled'				- manager
func (r *applicationRepository) RecallApplication(ctx context.Context, dto entities.RecallApplicationDTO) Error.CodeError {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	defer tx.Rollback()

	if err := r.saveVersion(ctx, tx, dto.ApplicationUUID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("application not found")}
		}
		return Error.CodeError{Code: 0, Err: err}
	}

	query := `UPDATE applications
	SET
	    version = version + 1,
	    status = 'recalled',
	    updated_at = CURRENT_TIMESTAMP,
	    updated_by = $2,
	    managed_by = $2,
	    executed_by = NULL
	WHERE uuid = $1 AND deleted_at IS NULL;`

	res, err := tx.ExecContext(ctx, query, dto.ApplicationUUID, dto.InitiatorUUID)
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

	if err = tx.Commit(); err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	return Error.CodeError{Code: -1, Err: nil}
}

// TakeApplicationToVerification Взятие заявки на проверку
// 'on_verification'		- inspector
func (r *applicationRepository) TakeApplicationToVerification(ctx context.Context, dto entities.TakeApplicationToVerificationDTO) Error.CodeError {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	defer tx.Rollback()

	if err := r.saveVersion(ctx, tx, dto.ApplicationUUID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("application not found")}
		}
		return Error.CodeError{Code: 0, Err: err}
	}

	query := `UPDATE applications
	SET 
		version = version + 1,
		status = 'on_verification',
		updated_at = CURRENT_TIMESTAMP,
		updated_by = $2,
		inspected_by = $2
	WHERE uuid = $1 AND deleted_at IS NULL;`

	res, err := tx.ExecContext(ctx, query, dto.ApplicationUUID, dto.InitiatorUUID)
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

	if err = tx.Commit(); err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	return Error.CodeError{Code: -1, Err: nil}
}

// ReleaseApplicationVerification Отмена взятия заявки на проверку
// 'pending_verification'	- inspector
func (r *applicationRepository) ReleaseApplicationVerification(ctx context.Context, dto entities.ReleaseApplicationVerificationDTO) Error.CodeError {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	defer tx.Rollback()

	if err := r.saveVersion(ctx, tx, dto.ApplicationUUID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("application not found")}
		}
		return Error.CodeError{Code: 0, Err: err}
	}

	query := `UPDATE applications
	SET
		version = version + 1,
		status = 'pending_verification',
		updated_at = CURRENT_TIMESTAMP,
		updated_by = $2,
		inspected_by = NULL
	WHERE uuid = $1 AND deleted_at IS NULL;`

	res, err := tx.ExecContext(ctx, query, dto.ApplicationUUID, dto.InitiatorUUID)
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

	if err = tx.Commit(); err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	return Error.CodeError{Code: -1, Err: nil}
}

// DeleteApplication Мягкое удаление заявки
func (r *applicationRepository) DeleteApplication(ctx context.Context, dto entities.DeleteApplicationDTO) Error.CodeError {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	defer tx.Rollback()

	if err := r.saveVersion(ctx, tx, dto.ApplicationUUID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("application not found")}
		}
		return Error.CodeError{Code: 0, Err: err}
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE applications
		SET 
		    version = version + 1,
		    updated_at = CURRENT_TIMESTAMP,
		    updated_by = $2,
		    deleted_at = CURRENT_TIMESTAMP, 
		    deleted_by = $2
		WHERE uuid = $1 AND deleted_at IS NULL;`,
		dto.ApplicationUUID, dto.DeletedBy,
	)
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

	if err = tx.Commit(); err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	return Error.CodeError{Code: -1, Err: nil}
}
