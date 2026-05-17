package entities

import (
	"fmt"

	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/utils"
)

type CreateCompanyRequest struct {
	Title string `json:"title"`
}
type CreateCompanyResponse struct {
	CompanyUUID string `json:"company_uuid"`
}

func (e *CreateCompanyRequest) Validate() error {
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
	if err := utils.ValidateNumber(int(e.Offset), 0, 0, "offset"); err != nil {
		return err
	}
	if err := utils.ValidateNumber(int(e.Count), 1, 100, "count"); err != nil {
		return err
	}
	return nil
}

type GetUserCompaniesResponse struct {
	Companies []*GetCompanyResponse `json:"companies"`
}

type UpdateCompanyTitleRequest struct {
	Title string `json:"title"`
}
type UpdateCompanyTitleRequestFull struct {
	CompanyUUID string
	Title       string
}
type UpdateCompanyTitleResponse struct{}

func (e *UpdateCompanyTitleRequestFull) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if err := utils.ValidateCompanyTitle(e.Title); err != nil {
		return err
	}
	return nil
}

type UpdateCompanyStatusRequest struct {
	Status string `json:"status"`
}
type UpdateCompanyStatusRequestFull struct {
	CompanyUUID string
	Status      string
}
type UpdateCompanyStatusResponse struct{}

func (e *UpdateCompanyStatusRequestFull) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if !utils.ValidateIsArrayContain(e.Status, []string{"open", "close"}) {
		return fmt.Errorf("incorrect company status")
	}
	return nil
}

type DeleteCompanyRequest struct {
	CompanyUUID string
}
type DeleteCompanyResponse struct{}

func (e *DeleteCompanyRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	return nil
}

type CreateCompanyJoinCodeRequest struct {
	CodeTTL int64 `json:"code_ttl"`
}
type CreateCompanyJoinCodeRequestFull struct {
	CompanyUUID string
	CodeTTL     int64
}
type CreateCompanyJoinCodeResponse struct {
	Code string `json:"code"`
}

func (e *CreateCompanyJoinCodeRequestFull) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if err := utils.ValidateNumber(int(e.CodeTTL), 60, 60*60*24*7, "code_ttl"); err != nil {
		return err
	}
	return nil
}

type GetCompanyJoinCodesRequest struct {
	CompanyUUID string
}
type GetCompanyJoinCodesResponse struct {
	Codes []string `json:"codes"`
}

func (e *GetCompanyJoinCodesRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	return nil
}

type DeleteCompanyJoinCodeRequest struct {
	Code string `json:"code"`
}
type DeleteCompanyJoinCodeRequestFull struct {
	CompanyUUID string
	Code        string
}
type DeleteCompanyJoinCodeResponse struct{}

func (e *DeleteCompanyJoinCodeRequestFull) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
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
	if err := utils.ValidateCompanyJoinCode(e.Code); err != nil {
		return err
	}
	return nil
}

type GetCompanyEmployeeRequest struct {
	TargetUUID  string
	CompanyUUID string
}
type GetCompanyEmployeeResponse struct {
	UserUUID       string `json:"user_uuid"`
	Role           string `json:"role"`
	DepartmentUUID string `json:"department_uuid"`
	JoinedAt       string `json:"joined_at"`
}

func (e *GetCompanyEmployeeRequest) Validate() error {
	if err := utils.ValidateUUID(e.TargetUUID); err != nil {
		return err
	}
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	return nil
}

type GetCompanyEmployeesRequest struct {
	CompanyUUID    string
	DepartmentUUID string `query:"department_uuid"`
	Role           string `query:"role"`
	Offset         int64  `query:"offset"`
	Count          int64  `query:"count"`
}
type GetCompanyEmployeesResponse struct {
	Employees []*GetCompanyEmployeeResponse `json:"employees"`
}

func (e *GetCompanyEmployeesRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if e.DepartmentUUID != "" {
		if err := utils.ValidateUUID(e.DepartmentUUID); err != nil {
			return err
		}
	}
	if !utils.ValidateIsArrayContain(e.Role, []string{"", "unemployed", "inspector", "engineer", "manager", "analytic", "chief"}) {
		return fmt.Errorf("incorrect employee role")
	}
	if err := utils.ValidateNumber(int(e.Offset), 0, 0, "offset"); err != nil {
		return err
	}
	if err := utils.ValidateNumber(int(e.Count), 1, 100, "count"); err != nil {
		return err
	}
	return nil
}

type GetCompanyEmployeesSummaryRequest struct {
	CompanyUUID    string
	DepartmentUUID string `query:"department_uuid"`
}
type GetCompanyEmployeesSummaryResponse struct {
	ChiefCount      int64 `json:"chief_count"`
	AnalyticsCount  int64 `json:"analytics_count"`
	ManagerCount    int64 `json:"manager_count"`
	EngineerCount   int64 `json:"engineer_count"`
	InspectorCount  int64 `json:"inspector_count"`
	UnemployedCount int64 `json:"unemployed_count"`
}

func (e *GetCompanyEmployeesSummaryRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if e.DepartmentUUID != "" {
		if err := utils.ValidateUUID(e.DepartmentUUID); err != nil {
			return err
		}
	}
	return nil
}

type UpdateEmployeeRoleRequest struct {
	Role string `json:"role"`
}
type UpdateEmployeeRoleRequestFull struct {
	CompanyUUID string
	TargetUUID  string
	Role        string
}
type UpdateEmployeeRoleResponse struct{}

func (e *UpdateEmployeeRoleRequestFull) Validate() error {
	if err := utils.ValidateUUID(e.TargetUUID); err != nil {
		return err
	}
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if !utils.ValidateIsArrayContain(e.Role, []string{"unemployed", "engineer", "manager", "analytic", "inspector", "chief"}) {
		return fmt.Errorf("incorrect employee role")
	}
	return nil
}

type RemoveCompanyEmployeeRequest struct {
	TargetUUID  string
	CompanyUUID string
}
type RemoveCompanyEmployeeResponse struct{}

func (e *RemoveCompanyEmployeeRequest) Validate() error {
	if err := utils.ValidateUUID(e.TargetUUID); err != nil {
		return err
	}
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	return nil
}

// Department

type CreateDepartmentRequest struct {
	Title string `json:"title"`
}
type CreateDepartmentRequestFull struct {
	CompanyUUID string
	Title       string
}
type CreateDepartmentResponse struct {
	DepartmentUUID string `json:"department_uuid"`
}

func (e *CreateDepartmentRequestFull) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if err := utils.ValidateCompanyTitle(e.Title); err != nil {
		return err
	}
	return nil
}

type GetDepartmentRequest struct {
	CompanyUUID    string
	DepartmentUUID string
}
type GetDepartmentResponse struct {
	DepartmentUUID string `json:"department_uuid"`
	CompanyUUID    string `json:"company_uuid"`
	Title          string `json:"title"`
	CreatedAt      string `json:"created_at"`
	CreatedBy      string `json:"created_by"`
}

func (e *GetDepartmentRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if err := utils.ValidateUUID(e.DepartmentUUID); err != nil {
		return err
	}
	return nil
}

type GetCompanyDepartmentsRequest struct {
	CompanyUUID string
	Offset      int64 `query:"offset"`
	Count       int64 `query:"count"`
}
type DepartmentListItem struct {
	DepartmentUUID string `json:"department_uuid"`
	Title          string `json:"title"`
}
type GetCompanyDepartmentsResponse struct {
	Departments []*DepartmentListItem `json:"departments"`
}

func (e *GetCompanyDepartmentsRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if err := utils.ValidateNumber(int(e.Offset), 0, 0, "offset"); err != nil {
		return err
	}
	if err := utils.ValidateNumber(int(e.Count), 1, 100, "count"); err != nil {
		return err
	}
	return nil
}

type UpdateDepartmentTitleRequest struct {
	Title string `json:"title"`
}
type UpdateDepartmentTitleRequestFull struct {
	CompanyUUID    string
	DepartmentUUID string
	Title          string
}
type UpdateDepartmentTitleResponse struct{}

func (e *UpdateDepartmentTitleRequestFull) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if err := utils.ValidateUUID(e.DepartmentUUID); err != nil {
		return err
	}
	if err := utils.ValidateCompanyTitle(e.Title); err != nil {
		return err
	}
	return nil
}

type DeleteDepartmentRequest struct {
	CompanyUUID    string
	DepartmentUUID string
}
type DeleteDepartmentResponse struct{}

func (e *DeleteDepartmentRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if err := utils.ValidateUUID(e.DepartmentUUID); err != nil {
		return err
	}
	return nil
}

type AddEmployeeToDepartmentRequest struct {
	CompanyUUID    string
	DepartmentUUID string
	TargetUUID     string
}
type AddEmployeeToDepartmentResponse struct{}

func (e *AddEmployeeToDepartmentRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if err := utils.ValidateUUID(e.DepartmentUUID); err != nil {
		return err
	}
	if err := utils.ValidateUUID(e.TargetUUID); err != nil {
		return err
	}
	return nil
}

type RemoveEmployeeFromDepartmentRequest struct {
	CompanyUUID    string
	DepartmentUUID string
	TargetUUID     string
}
type RemoveEmployeeFromDepartmentResponse struct{}

func (e *RemoveEmployeeFromDepartmentRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if err := utils.ValidateUUID(e.DepartmentUUID); err != nil {
		return err
	}
	if err := utils.ValidateUUID(e.TargetUUID); err != nil {
		return err
	}
	return nil
}
