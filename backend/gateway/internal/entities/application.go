package entities

import (
	"fmt"

	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/utils"
)

// ─── Shared response types ────────────────────────────────────────────────────

// FixLogResponse — запись fix log в ответе
type FixLogResponse struct {
	UUID      string `json:"uuid"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
	CreatedBy string `json:"created_by"`
}

// ApplicationResponse — полный объект заявки (используется в GetApplication)
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

// ApplicationListItem — краткий объект заявки в списке (используется в GetApplications)
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
	if err := utils.ValidateApplicationDescription(e.Description); err != nil {
		return err
	}
	return nil
}

type CreateApplicationResponse struct {
	ApplicationUUID string `json:"application_uuid"`
}

// ─── GetApplication ───────────────────────────────────────────────────────────

type GetApplicationRequest struct {
	ApplicationUUID string
}

func (e *GetApplicationRequest) Validate() error {
	return utils.ValidateUUID(e.ApplicationUUID)
}

type GetApplicationResponse struct {
	Application *ApplicationResponse `json:"application"`
}

// ─── GetApplications ──────────────────────────────────────────────────────────

type GetApplicationsRequest struct {
	CompanyUUID    string
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
	Status string `json:"status"`
}

type UpdateApplicationStatusRequestFull struct {
	ApplicationUUID string
	Status          string
}

func (e *UpdateApplicationStatusRequestFull) Validate() error {
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
	TargetUUID string `json:"target_uuid"`
}

type AssignApplicationRequestFull struct {
	ApplicationUUID string
	TargetUUID      string
}

func (e *AssignApplicationRequestFull) Validate() error {
	if err := utils.ValidateUUID(e.ApplicationUUID); err != nil {
		return err
	}
	return utils.ValidateUUID(e.TargetUUID)
}

type AssignApplicationResponse struct{}

// ─── RedirectApplication ──────────────────────────────────────────────────────

type RedirectApplicationRequest struct {
	TargetDepartmentUUID string `json:"target_department_uuid"`
	Message              string `json:"message"`
}

type RedirectApplicationRequestFull struct {
	ApplicationUUID      string
	TargetDepartmentUUID string
	Message              string
}

func (e *RedirectApplicationRequestFull) Validate() error {
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
	Message string `json:"message"`
}

type RecallApplicationRequestFull struct {
	ApplicationUUID string
	Message         string
}

func (e *RecallApplicationRequestFull) Validate() error {
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
	ApplicationUUID string
}

func (e *TakeApplicationToVerificationRequest) Validate() error {
	return utils.ValidateUUID(e.ApplicationUUID)
}

type TakeApplicationToVerificationResponse struct{}

// ─── ReleaseApplicationVerification ──────────────────────────────────────────

type ReleaseApplicationVerificationRequest struct {
	Message string `json:"message"`
}

type ReleaseApplicationVerificationRequestFull struct {
	ApplicationUUID string
	Message         string
}

func (e *ReleaseApplicationVerificationRequestFull) Validate() error {
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
	Message string `json:"message"`
}

type AddApplicationFixLogRequestFull struct {
	ApplicationUUID string
	Message         string
}

func (e *AddApplicationFixLogRequestFull) Validate() error {
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
	Message string `json:"message"`
}

type DeleteApplicationRequestFull struct {
	ApplicationUUID string
	Message         string
}

func (e *DeleteApplicationRequestFull) Validate() error {
	return utils.ValidateUUID(e.ApplicationUUID)
}

type DeleteApplicationResponse struct{}
