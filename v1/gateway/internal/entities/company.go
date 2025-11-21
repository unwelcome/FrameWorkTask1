package entities

import (
	"fmt"

	"github.com/unwelcome/FrameWorkTask1/v1/gateway/pkg/utils"
)

type CreateCompanyRequest struct {
	Title string `json:"title"`
}
type CreateCompanyResponse struct {
	CompanyUUID string `json:"company_uuid"`
}

func (e *CreateCompanyRequest) Validate() error {
	// Валидация title
	if err := utils.ValidateCompanyTitle(e.Title); err != nil {
		return err
	}

	return nil
}

type GetCompanyRequest struct {
	CompanyUUID string `json:"company_uuid"`
}
type GetCompanyResponse struct {
	CompanyUUID string `json:"company_uuid"`
	Title       string `json:"title"`
	Status      string `json:"status"`
}

func (e *GetCompanyRequest) Validate() error {
	// Валидация company_uuid
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}

	return nil
}

type GetCompaniesRequest struct {
	Offset int64 `query:"offset"`
	Count  int64 `query:"count"`
}
type GetCompaniesResponse struct {
	Companies []*GetCompanyResponse `json:"companies"`
}

func (e *GetCompaniesRequest) Validate() error {
	// Валидация offset
	if err := utils.ValidateNumber(int(e.Offset), 0, 0, "offset"); err != nil {
		return err
	}

	// Валидация count
	if err := utils.ValidateNumber(int(e.Count), 1, 100, "count"); err != nil {
		return err
	}

	return nil
}

type UpdateCompanyTitleRequest struct {
	Title string `json:"title"`
}
type UpdateCompanyTitleRequestFull struct {
	CompanyUUID string `json:"company_uuid"`
	Title       string `json:"title"`
}
type UpdateCompanyTitleResponse struct {
}

func (e *UpdateCompanyTitleRequestFull) Validate() error {
	// Валидация company_uuid
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}

	// Валидация title
	if err := utils.ValidateCompanyTitle(e.Title); err != nil {
		return err
	}

	return nil
}

type UpdateCompanyStatusRequest struct {
	Status string `json:"status"`
}
type UpdateCompanyStatusRequestFull struct {
	CompanyUUID string `json:"company_uuid"`
	Status      string `json:"status"`
}
type UpdateCompanyStatusResponse struct {
}

func (e *UpdateCompanyStatusRequestFull) Validate() error {
	// Валидация company_uuid
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}

	// Валидация status
	if !utils.ValidateIsArrayContain(e.Status, []string{"unemployed", "engineer", "manager", "analytic", "chief"}) {
		return fmt.Errorf("incorrect company status")
	}

	return nil
}

type DeleteCompanyRequest struct {
	CompanyUUID string `json:"company_uuid"`
}
type DeleteCompanyResponse struct {
}

func (e *DeleteCompanyRequest) Validate() error {
	// Валидация company_uuid
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}

	return nil
}

type CreateCompanyJoinCodeRequest struct {
	CompanyUUID string `json:"company_uuid"`
}
type CreateCompanyJoinCodeResponse struct {
	Code string `json:"code"`
}

func (e *CreateCompanyJoinCodeRequest) Validate() error {
	// Валидация company_uuid
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}

	return nil
}

type GetCompanyJoinCodesRequest struct {
	CompanyUUID string `json:"company_uuid"`
}
type GetCompanyJoinCodesResponse struct {
	Codes []string `json:"codes"`
}

func (e *GetCompanyJoinCodesRequest) Validate() error {
	// Валидация company_uuid
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}

	return nil
}

type DeleteCompanyJoinCodeRequest struct {
	Code string `json:"code"`
}
type DeleteCompanyJoinCodeRequestFull struct {
	CompanyUUID string `json:"company_uuid"`
	Code        string `json:"code"`
}
type DeleteCompanyJoinCodeResponse struct {
}

func (e *DeleteCompanyJoinCodeRequestFull) Validate() error {
	// Валидация company_uuid
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}

	// Валидация join code
	if err := utils.ValidateCompanyJoinCode(e.Code); err != nil {
		return err
	}

	return nil
}

type JoinCompanyRequest struct {
	Code string `json:"code"`
}
type JoinCompanyResponse struct {
	CompanyUUID string `json:"company_uuid"`
	Role        string `json:"role"`
}

func (e *JoinCompanyRequest) Validate() error {
	// Валидация join code
	if err := utils.ValidateCompanyJoinCode(e.Code); err != nil {
		return err
	}

	return nil
}

type GetCompanyEmployeeRequest struct {
	TargetUUID  string `json:"target_uuid"`
	CompanyUUID string `json:"company_uuid"`
}
type GetCompanyEmployeeResponse struct {
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}

func (e *GetCompanyEmployeeRequest) Validate() error {
	// Валидация target_uuid
	if err := utils.ValidateUUID(e.TargetUUID); err != nil {
		return err
	}

	// Валидация company_uuid
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}

	return nil
}

type GetCompanyEmployeesSummaryRequest struct {
	CompanyUUID string `json:"company_uuid"`
}
type GetCompanyEmployeesSummaryResponse struct {
	ChiefCount     int64 `json:"chief_count"`
	AnalyticsCount int64 `json:"analytics_count"`
	ManagerCount   int64 `json:"manager_count"`
	EngineerCount  int64 `json:"engineer_count"`
	UnemployedCoun int64 `json:"unemployed_count"`
}

func (e *GetCompanyEmployeesSummaryRequest) Validate() error {
	// Валидация company_uuid
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}

	return nil
}

type RemoveCompanyEmployeeRequest struct {
	TargetUUID  string `json:"target_uuid"`
	CompanyUUID string `json:"company_uuid"`
}
type RemoveCompanyEmployeeResponse struct {
}

func (e *RemoveCompanyEmployeeRequest) Validate() error {
	// Валидация target_uuid
	if err := utils.ValidateUUID(e.TargetUUID); err != nil {
		return err
	}

	// Валидация company_uuid
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}

	return nil
}
