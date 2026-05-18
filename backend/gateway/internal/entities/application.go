package entities

import (
	"fmt"

	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/utils"
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

func (e *CreateApplicationRequest) Validate() error {
	if err := utils.ValidateApplicationTitle(e.Title); err != nil {
		return err
	}
	return utils.ValidateApplicationDescription(e.Description)
}

type CreateApplicationResponse struct {
	ApplicationUUID string `json:"application_uuid"`
}

// ─── GetApplication ───────────────────────────────────────────────────────────

type GetApplicationRequest struct {
	ApplicationUUID string `json:"-"`
}

func (e *GetApplicationRequest) Validate() error {
	return utils.ValidateUUID(e.ApplicationUUID)
}

type GetApplicationResponse struct {
	Application *ApplicationResponse `json:"application"`
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

func (e *GetApplicationsRequest) Validate() error {
	return utils.ValidateUUID(e.CompanyUUID)
}

type GetApplicationsResponse struct {
	Applications []*ApplicationListItem `json:"applications"`
}

// ─── UpdateApplicationStatus ──────────────────────────────────────────────────

type UpdateApplicationStatusRequest struct {
	ApplicationUUID string `json:"-"`
	Status          string `json:"status"`
}

func (e *UpdateApplicationStatusRequest) Validate() error {
	if err := utils.ValidateUUID(e.ApplicationUUID); err != nil {
		return err
	}
	if e.Status == "" {
		return fmt.Errorf("status missed")
	}
	return nil
}

type UpdateApplicationStatusResponse struct{}

// ─── AssignApplication ────────────────────────────────────────────────────────

type AssignApplicationRequest struct {
	ApplicationUUID string `json:"-"`
	TargetUUID      string `json:"target_uuid"`
}

func (e *AssignApplicationRequest) Validate() error {
	if err := utils.ValidateUUID(e.ApplicationUUID); err != nil {
		return err
	}
	return utils.ValidateUUID(e.TargetUUID)
}

type AssignApplicationResponse struct{}

// ─── RedirectApplication ──────────────────────────────────────────────────────

type RedirectApplicationRequest struct {
	ApplicationUUID      string `json:"-"`
	TargetDepartmentUUID string `json:"target_department_uuid"`
	Message              string `json:"message"`
}

func (e *RedirectApplicationRequest) Validate() error {
	if err := utils.ValidateUUID(e.ApplicationUUID); err != nil {
		return err
	}
	if err := utils.ValidateUUID(e.TargetDepartmentUUID); err != nil {
		return fmt.Errorf("target_department_uuid: %w", err)
	}
	if e.Message == "" {
		return fmt.Errorf("message missed")
	}
	return nil
}

type RedirectApplicationResponse struct{}

// ─── RecallApplication ────────────────────────────────────────────────────────

type RecallApplicationRequest struct {
	ApplicationUUID string `json:"-"`
	Message         string `json:"message"`
}

func (e *RecallApplicationRequest) Validate() error {
	if err := utils.ValidateUUID(e.ApplicationUUID); err != nil {
		return err
	}
	if e.Message == "" {
		return fmt.Errorf("message missed")
	}
	return nil
}

type RecallApplicationResponse struct{}

// ─── TakeApplicationToVerification ───────────────────────────────────────────

type TakeApplicationToVerificationRequest struct {
	ApplicationUUID string `json:"-"`
}

func (e *TakeApplicationToVerificationRequest) Validate() error {
	return utils.ValidateUUID(e.ApplicationUUID)
}

type TakeApplicationToVerificationResponse struct{}

// ─── ReleaseApplicationVerification ──────────────────────────────────────────

type ReleaseApplicationVerificationRequest struct {
	ApplicationUUID string `json:"-"`
	Message         string `json:"message"`
}

func (e *ReleaseApplicationVerificationRequest) Validate() error {
	if err := utils.ValidateUUID(e.ApplicationUUID); err != nil {
		return err
	}
	if e.Message == "" {
		return fmt.Errorf("message missed")
	}
	return nil
}

type ReleaseApplicationVerificationResponse struct{}

// ─── AddApplicationFixLog ─────────────────────────────────────────────────────

type AddApplicationFixLogRequest struct {
	ApplicationUUID string `json:"-"`
	Message         string `json:"message"`
}

func (e *AddApplicationFixLogRequest) Validate() error {
	if err := utils.ValidateUUID(e.ApplicationUUID); err != nil {
		return err
	}
	if e.Message == "" {
		return fmt.Errorf("message missed")
	}
	return nil
}

type AddApplicationFixLogResponse struct{}

// ─── DeleteApplication ────────────────────────────────────────────────────────

type DeleteApplicationRequest struct {
	ApplicationUUID string `json:"-"`
	Message         string `json:"message"`
}

func (e *DeleteApplicationRequest) Validate() error {
	return utils.ValidateUUID(e.ApplicationUUID)
}

type DeleteApplicationResponse struct{}
