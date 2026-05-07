package entities

import (
	"fmt"

	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/utils"
)

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
	if err := utils.ValidateApplicationTitle(e.Title); err != nil {
		return err
	}
	if err := utils.ValidateApplicationDescription(e.Description); err != nil {
		return err
	}
	return nil
}

// ─── Shared response types ────────────────────────────────────────────────────

// FixLogResponse — запись fix log в ответе
type FixLogResponse struct {
	UUID      string `json:"uuid"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
	CreatedBy string `json:"created_by"`
}

// ApplicationResponse — объект заявки в ответе
type ApplicationResponse struct {
	ApplicationUUID     string            `json:"application_uuid"`
	CompanyUUID         string            `json:"company_uuid"`
	Title               string            `json:"title"`
	Description         string            `json:"description"`
	Status              string            `json:"status"`
	ResponsibleManager  string            `json:"responsible_manager"`
	ResponsibleEngineer string            `json:"responsible_engineer"`
	CreatedAt           string            `json:"created_at"`
	CreatedBy           string            `json:"created_by"`
	ClosedAt            string            `json:"closed_at"`
	FixLogs             []*FixLogResponse `json:"fix_logs,omitempty"`
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
	CompanyUUID string
	Count       int64 `query:"count"`
	Offset      int64 `query:"offset"`
}

func (e *GetApplicationsRequest) Validate() error {
	return utils.ValidateUUID(e.CompanyUUID)
}

type GetApplicationsResponse struct {
	Applications []*ApplicationResponse `json:"applications"`
}

// ─── GetCompanyApplicationStatistic ──────────────────────────────────────────

type GetCompanyApplicationStatisticRequest struct {
	CompanyUUID string
}

func (e *GetCompanyApplicationStatisticRequest) Validate() error {
	return utils.ValidateUUID(e.CompanyUUID)
}

type GetCompanyApplicationStatisticResponse struct {
	Created          int64 `json:"created"`
	Assigned         int64 `json:"assigned"`
	InProgress       int64 `json:"in_progress"`
	OnHold           int64 `json:"on_hold"`
	AwaitingApproval int64 `json:"awaiting_approval"`
	Completed        int64 `json:"completed"`
	Cancelled        int64 `json:"cancelled"`
	Failed           int64 `json:"failed"`
	Archived         int64 `json:"archived"`
}

// ─── GetEmployeeApplicationStatistic ─────────────────────────────────────────

type GetEmployeeApplicationStatisticRequest struct {
	CompanyUUID string
	TargetUUID  string
}

func (e *GetEmployeeApplicationStatisticRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	return utils.ValidateUUID(e.TargetUUID)
}

type GetEmployeeApplicationStatisticResponse struct {
	Created          int64 `json:"created"`
	Assigned         int64 `json:"assigned"`
	InProgress       int64 `json:"in_progress"`
	OnHold           int64 `json:"on_hold"`
	AwaitingApproval int64 `json:"awaiting_approval"`
	Completed        int64 `json:"completed"`
	Cancelled        int64 `json:"cancelled"`
	Failed           int64 `json:"failed"`
	Archived         int64 `json:"archived"`
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

// ─── AssignApplicationToEmployee ─────────────────────────────────────────────

type AssignApplicationToEmployeeRequest struct {
	TargetUUID string `json:"target_uuid"`
}

type AssignApplicationToEmployeeRequestFull struct {
	ApplicationUUID string
	TargetUUID      string
}

func (e *AssignApplicationToEmployeeRequestFull) Validate() error {
	if err := utils.ValidateUUID(e.ApplicationUUID); err != nil {
		return err
	}
	return utils.ValidateUUID(e.TargetUUID)
}

type AssignApplicationToEmployeeResponse struct{}

// ─── AddApplicationFixLog ─────────────────────────────────────────────────────

type AddApplicationFixLogRequest struct {
	LogText string `json:"log_text"`
}

type AddApplicationFixLogRequestFull struct {
	ApplicationUUID string
	LogText         string
}

func (e *AddApplicationFixLogRequestFull) Validate() error {
	if err := utils.ValidateUUID(e.ApplicationUUID); err != nil {
		return err
	}
	if e.LogText == "" {
		return fmt.Errorf("log_text missed")
	}
	return nil
}

type AddApplicationFixLogResponse struct{}

// ─── DeleteApplication ────────────────────────────────────────────────────────

type DeleteApplicationRequest struct {
	ApplicationUUID string
}

func (e *DeleteApplicationRequest) Validate() error {
	return utils.ValidateUUID(e.ApplicationUUID)
}

type DeleteApplicationResponse struct{}
