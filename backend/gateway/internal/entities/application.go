package entities

import (
	"fmt"
	"strings"

	"github.com/unwelcome/FrameWorkTask1/backend/shared/validate"
)

// ─── Shared response types ────────────────────────────────────────────────────

type FixLogResponse struct {
	UUID      string `json:"uuid"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
	CreatedBy string `json:"created_by"`
}

type ApplicationResponse struct {
	ApplicationUUID string            `json:"application_uuid"`
	CompanyUUID     string            `json:"company_uuid"`
	DepartmentUUID  string            `json:"department_uuid"`
	Version         int64             `json:"version"`
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	Status          string            `json:"status"`
	RevisionCount   int64             `json:"revision_count"`
	CreatedAt       string            `json:"created_at"`
	CreatedBy       string            `json:"created_by"`
	UpdatedAt       string            `json:"updated_at"`
	UpdatedBy       string            `json:"updated_by"`
	ManagedBy       string            `json:"managed_by"`
	ExecutedBy      string            `json:"executed_by"`
	InspectedBy     string            `json:"inspected_by"`
	ClosedAt        string            `json:"closed_at"`
	DeletedAt       string            `json:"deleted_at"`
	DeletedBy       string            `json:"deleted_by"`
	FixLogs         []*FixLogResponse `json:"fix_logs,omitempty"`
}

type ApplicationListItem struct {
	ApplicationUUID string `json:"application_uuid"`
	Title           string `json:"title"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// ─── CreateApplication ────────────────────────────────────────────────────────

type CreateApplicationRequest struct {
	CompanyUUID string `json:"company_uuid"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
type CreateApplicationResponse struct {
	ApplicationUUID string `json:"application_uuid"`
}

func (e *CreateApplicationRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.Title = strings.TrimSpace(e.Title)
	if err := validate.ApplicationTitle(e.Title); err != nil {
		return err
	}
	e.Description = strings.TrimSpace(e.Description)
	if err := validate.ApplicationDescription(e.Description); err != nil {
		return err
	}
	return nil
}

// ─── GetApplication ───────────────────────────────────────────────────────────

type GetApplicationRequest struct {
	ApplicationUUID string `json:"-"`
}
type GetApplicationResponse struct {
	Application *ApplicationResponse `json:"application"`
}

func (e *GetApplicationRequest) Validate() error {
	e.ApplicationUUID = strings.TrimSpace(e.ApplicationUUID)
	if err := validate.UUID(e.ApplicationUUID); err != nil {
		return err
	}
	return nil
}

// ─── GetApplications ──────────────────────────────────────────────────────────

type GetApplicationsRequest struct {
	CompanyUUID    string   `json:"-"`
	DepartmentUUID string   `query:"department_uuid"`
	Statuses       []string `query:"statuses"`
	Count          int64    `query:"count"`
	Offset         int64    `query:"offset"`
	IsDeleted      bool     `query:"is_deleted"`
	FromPool       bool     `query:"from_pool"`
}
type GetApplicationsResponse struct {
	Applications []*ApplicationListItem `json:"applications"`
}

func (e *GetApplicationsRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.DepartmentUUID = strings.TrimSpace(e.DepartmentUUID)
	if err := validate.UUID(e.DepartmentUUID); err != nil && e.DepartmentUUID != "" {
		return err
	}
	if err := validate.Number(int(e.Count), validate.IntPtr(1), validate.IntPtr(100), "count"); err != nil {
		return err
	}
	if err := validate.Number(int(e.Offset), validate.IntPtr(0), nil, "offset"); err != nil {
		return err
	}
	return nil
}

// ─── UpdateApplicationStatus ──────────────────────────────────────────────────

type UpdateApplicationStatusRequest struct {
	ApplicationUUID string `json:"-"`
	Status          string `json:"status"`
}
type UpdateApplicationStatusResponse struct{}

func (e *UpdateApplicationStatusRequest) Validate() error {
	e.ApplicationUUID = strings.TrimSpace(e.ApplicationUUID)
	if err := validate.UUID(e.ApplicationUUID); err != nil {
		return err
	}
	e.Status = strings.TrimSpace(e.Status)
	if e.Status == "" {
		return fmt.Errorf("status missed")
	}
	return nil
}

// ─── AssignApplication ────────────────────────────────────────────────────────

type AssignApplicationRequest struct {
	ApplicationUUID string `json:"-"`
	TargetUUID      string `json:"target_uuid"`
}
type AssignApplicationResponse struct{}

func (e *AssignApplicationRequest) Validate() error {
	e.ApplicationUUID = strings.TrimSpace(e.ApplicationUUID)
	if err := validate.UUID(e.ApplicationUUID); err != nil {
		return err
	}
	e.TargetUUID = strings.TrimSpace(e.TargetUUID)
	if err := validate.UUID(e.TargetUUID); err != nil {
		return err
	}
	return nil
}

// ─── RedirectApplication ──────────────────────────────────────────────────────

type RedirectApplicationRequest struct {
	ApplicationUUID      string `json:"-"`
	TargetDepartmentUUID string `json:"target_department_uuid"`
	Message              string `json:"message"`
}
type RedirectApplicationResponse struct{}

func (e *RedirectApplicationRequest) Validate() error {
	e.ApplicationUUID = strings.TrimSpace(e.ApplicationUUID)
	if err := validate.UUID(e.ApplicationUUID); err != nil {
		return err
	}
	e.TargetDepartmentUUID = strings.TrimSpace(e.TargetDepartmentUUID)
	if err := validate.UUID(e.TargetDepartmentUUID); err != nil {
		return err
	}
	e.Message = strings.TrimSpace(e.Message)
	if e.Message == "" {
		return fmt.Errorf("message missed")
	}
	return nil
}

// ─── RecallApplication ────────────────────────────────────────────────────────

type RecallApplicationRequest struct {
	ApplicationUUID string `json:"-"`
	Message         string `json:"message"`
}
type RecallApplicationResponse struct{}

func (e *RecallApplicationRequest) Validate() error {
	e.ApplicationUUID = strings.TrimSpace(e.ApplicationUUID)
	if err := validate.UUID(e.ApplicationUUID); err != nil {
		return err
	}
	e.Message = strings.TrimSpace(e.Message)
	if e.Message == "" {
		return fmt.Errorf("message missed")
	}
	return nil
}

// ─── TakeApplicationToVerification ───────────────────────────────────────────

type TakeApplicationToVerificationRequest struct {
	ApplicationUUID string `json:"-"`
}
type TakeApplicationToVerificationResponse struct{}

func (e *TakeApplicationToVerificationRequest) Validate() error {
	e.ApplicationUUID = strings.TrimSpace(e.ApplicationUUID)
	if err := validate.UUID(e.ApplicationUUID); err != nil {
		return err
	}
	return nil
}

// ─── ReleaseApplicationVerification ──────────────────────────────────────────

type ReleaseApplicationVerificationRequest struct {
	ApplicationUUID string `json:"-"`
	Message         string `json:"message"`
}
type ReleaseApplicationVerificationResponse struct{}

func (e *ReleaseApplicationVerificationRequest) Validate() error {
	e.ApplicationUUID = strings.TrimSpace(e.ApplicationUUID)
	if err := validate.UUID(e.ApplicationUUID); err != nil {
		return err
	}
	e.Message = strings.TrimSpace(e.Message)
	if e.Message == "" {
		return fmt.Errorf("message missed")
	}
	return nil
}

// ─── AddApplicationFixLog ─────────────────────────────────────────────────────

type AddApplicationFixLogRequest struct {
	ApplicationUUID string `json:"-"`
	Message         string `json:"message"`
}
type AddApplicationFixLogResponse struct{}

func (e *AddApplicationFixLogRequest) Validate() error {
	e.ApplicationUUID = strings.TrimSpace(e.ApplicationUUID)
	if err := validate.UUID(e.ApplicationUUID); err != nil {
		return err
	}
	e.Message = strings.TrimSpace(e.Message)
	if e.Message == "" {
		return fmt.Errorf("message missed")
	}
	return nil
}

// ─── DeleteApplication ────────────────────────────────────────────────────────

type DeleteApplicationRequest struct {
	ApplicationUUID string `json:"-"`
	Message         string `json:"message"`
}
type DeleteApplicationResponse struct{}

func (e *DeleteApplicationRequest) Validate() error {
	e.ApplicationUUID = strings.TrimSpace(e.ApplicationUUID)
	if err := validate.UUID(e.ApplicationUUID); err != nil {
		return err
	}
	e.Message = strings.TrimSpace(e.Message)
	if e.Message == "" {
		return fmt.Errorf("message missed")
	}
	return nil
}

// ─── GetApplicationHistory ────────────────────────────────────────────────────

type GetApplicationHistoryRequest struct {
	ApplicationUUID string `json:"-"`
	Count           int64  `query:"count"`
	Offset          int64  `query:"offset"`
}
type GetApplicationHistoryResponse struct {
	History []*ApplicationResponse `json:"history"`
}

func (e *GetApplicationHistoryRequest) Validate() error {
	e.ApplicationUUID = strings.TrimSpace(e.ApplicationUUID)
	if err := validate.UUID(e.ApplicationUUID); err != nil {
		return err
	}
	if err := validate.Number(int(e.Count), validate.IntPtr(1), validate.IntPtr(100), "count"); err != nil {
		return err
	}
	if err := validate.Number(int(e.Offset), validate.IntPtr(0), nil, "offset"); err != nil {
		return err
	}
	return nil
}
