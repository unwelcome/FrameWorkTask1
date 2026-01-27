package postgresDB

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/unwelcome/FrameWorkTask1/v1/application/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/v1/application/pkg/errors"
	"google.golang.org/grpc/codes"
)

type ApplicationRepository interface {
	CreateApplication(ctx context.Context, dto *entities.CreateApplication) Error.CodeError
	AddApplicationFixLog(ctx context.Context, dto *entities.CreateFixLog) Error.CodeError
	GetApplication(ctx context.Context, applicationUUID string) (*entities.Application, Error.CodeError)
	GetApplicationFixLogs(ctx context.Context, applicationUUID string) ([]*entities.FixLog, Error.CodeError)
	GetApplications(ctx context.Context, status string, count int, offset int) ([]*entities.Application, Error.CodeError)
}

type applicationRepository struct {
	db *sql.DB
}

func NewApplicationRepository(db *sql.DB) ApplicationRepository {
	return &applicationRepository{db: db}
}

func (r *applicationRepository) CreateApplication(ctx context.Context, dto *entities.CreateApplication) Error.CodeError {
	query := `INSERT INTO applications 
	(uuid, version, title, description, status, created_by) VALUES 
	($1, 1, $2, $3, 'created', $4);`

	_, err := r.db.ExecContext(ctx, query, dto.ApplicationUUID, dto.Title, dto.Description, dto.CreatedBy)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	return Error.CodeError{Code: -1, Err: nil}
}

func (r *applicationRepository) AddApplicationFixLog(ctx context.Context, dto *entities.CreateFixLog) Error.CodeError {
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

func (r *applicationRepository) GetApplication(ctx context.Context, applicationUUID string) (*entities.Application, Error.CodeError) {
	query := `SELECT 
			title, 
			description, 
			status, 
			created_at, 
			created_by, 
			closed_at, 
			managed_by, 
			executed_by 
		FROM applications 
		WHERE uuid = $1;`

	application := &entities.Application{ApplicationUUID: applicationUUID}
	err := r.db.QueryRowContext(ctx, query, applicationUUID).Scan(
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

func (r *applicationRepository) GetApplicationFixLogs(ctx context.Context, applicationUUID string) ([]*entities.FixLog, Error.CodeError) {
	query := `SELECT 
			id, 
			text, 
			created_at, 
			created_by 
		FROM application_fix_logs 
		WHERE application_uuid = $1;`

	res, err := r.db.QueryContext(ctx, query, applicationUUID)
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

func (r *applicationRepository) GetApplications(ctx context.Context, status string, count int, offset int) ([]*entities.Application, Error.CodeError) {
	query := `
		SELECT 
			uuid, 
			title, 
			description, 
			status, 
			created_at, 
			created_by, 
			closed_at, 
			managed_by, 
			executed_by 
		FROM applications 
		WHERE status = $1 
		ORDER BY created_at DESC, uuid
		OFFSET $2 LIMIT $3;`

	rows, err := r.db.QueryContext(ctx, query, status, offset, count)
	if err != nil {
		return nil, Error.CodeError{Code: 0, Err: err}
	}
	defer rows.Close()

	applications := make([]*entities.Application, 0)
	for rows.Next() {
		item := &entities.Application{}
		err = rows.Scan(
			&item.ApplicationUUID,
			&item.Title,
			&item.Description,
			&item.Status,
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
