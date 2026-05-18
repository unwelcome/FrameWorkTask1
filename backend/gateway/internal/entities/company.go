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
	return utils.ValidateCompanyTitle(e.Title)
}

type GetCompanyRequest struct {
	CompanyUUID string `json:"-"`
}
type GetCompanyResponse struct {
	CompanyUUID string `json:"company_uuid"`
	Title       string `json:"title"`
	Status      string `json:"status"`
}

func (e *GetCompanyRequest) Validate() error {
	return utils.ValidateUUID(e.CompanyUUID)
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
	CompanyUUID string `json:"-"`
	Title       string `json:"title"`
}
type UpdateCompanyTitleResponse struct{}

func (e *UpdateCompanyTitleRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	return utils.ValidateCompanyTitle(e.Title)
}

type UpdateCompanyStatusRequest struct {
	CompanyUUID string `json:"-"`
	Status      string `json:"status"`
}
type UpdateCompanyStatusResponse struct{}

func (e *UpdateCompanyStatusRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if !utils.ValidateIsArrayContain(e.Status, []string{"open", "close"}) {
		return fmt.Errorf("incorrect company status")
	}
	return nil
}

type DeleteCompanyRequest struct {
	CompanyUUID string `json:"-"`
}
type DeleteCompanyResponse struct{}

func (e *DeleteCompanyRequest) Validate() error {
	return utils.ValidateUUID(e.CompanyUUID)
}

type CreateCompanyJoinCodeRequest struct {
	CompanyUUID string `json:"-"`
	CodeTTL     int64  `json:"code_ttl"`
}
type CreateCompanyJoinCodeResponse struct {
	Code string `json:"code"`
}

func (e *CreateCompanyJoinCodeRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	return utils.ValidateNumber(int(e.CodeTTL), 60, 60*60*24*7, "code_ttl")
}

type GetCompanyJoinCodesRequest struct {
	CompanyUUID string `json:"-"`
}
type GetCompanyJoinCodesResponse struct {
	Codes []string `json:"codes"`
}

func (e *GetCompanyJoinCodesRequest) Validate() error {
	return utils.ValidateUUID(e.CompanyUUID)
}

type DeleteCompanyJoinCodeRequest struct {
	CompanyUUID string `json:"-"`
	Code        string `json:"code"`
}
type DeleteCompanyJoinCodeResponse struct{}

func (e *DeleteCompanyJoinCodeRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	return utils.ValidateCompanyJoinCode(e.Code)
}

type JoinCompanyRequest struct {
	Code string `json:"code"`
}
type JoinCompanyResponse struct {
	CompanyUUID string `json:"company_uuid"`
	Role        string `json:"role"`
}

func (e *JoinCompanyRequest) Validate() error {
	return utils.ValidateCompanyJoinCode(e.Code)
}

type GetCompanyEmployeeRequest struct {
	TargetUUID  string `json:"-"`
	CompanyUUID string `json:"-"`
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
	return utils.ValidateUUID(e.CompanyUUID)
}

type GetCompanyEmployeesRequest struct {
	CompanyUUID    string `json:"-"`
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
	return utils.ValidateNumber(int(e.Count), 1, 100, "count")
}

type GetCompanyEmployeesSummaryRequest struct {
	CompanyUUID    string `json:"-"`
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
	CompanyUUID string `json:"-"`
	TargetUUID  string `json:"-"`
	Role        string `json:"role"`
}
type UpdateEmployeeRoleResponse struct{}

func (e *UpdateEmployeeRoleRequest) Validate() error {
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
	TargetUUID  string `json:"-"`
	CompanyUUID string `json:"-"`
}
type RemoveCompanyEmployeeResponse struct{}

func (e *RemoveCompanyEmployeeRequest) Validate() error {
	if err := utils.ValidateUUID(e.TargetUUID); err != nil {
		return err
	}
	return utils.ValidateUUID(e.CompanyUUID)
}

// Department

type CreateDepartmentRequest struct {
	CompanyUUID string `json:"-"`
	Title       string `json:"title"`
}
type CreateDepartmentResponse struct {
	DepartmentUUID string `json:"department_uuid"`
}

func (e *CreateDepartmentRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	return utils.ValidateCompanyTitle(e.Title)
}

type GetDepartmentRequest struct {
	CompanyUUID    string `json:"-"`
	DepartmentUUID string `json:"-"`
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
	return utils.ValidateUUID(e.DepartmentUUID)
}

type GetCompanyDepartmentsRequest struct {
	CompanyUUID string `json:"-"`
	Offset      int64  `query:"offset"`
	Count       int64  `query:"count"`
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
	return utils.ValidateNumber(int(e.Count), 1, 100, "count")
}

type UpdateDepartmentTitleRequest struct {
	CompanyUUID    string `json:"-"`
	DepartmentUUID string `json:"-"`
	Title          string `json:"title"`
}
type UpdateDepartmentTitleResponse struct{}

func (e *UpdateDepartmentTitleRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if err := utils.ValidateUUID(e.DepartmentUUID); err != nil {
		return err
	}
	return utils.ValidateCompanyTitle(e.Title)
}

type DeleteDepartmentRequest struct {
	CompanyUUID    string `json:"-"`
	DepartmentUUID string `json:"-"`
}
type DeleteDepartmentResponse struct{}

func (e *DeleteDepartmentRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	return utils.ValidateUUID(e.DepartmentUUID)
}

type AddEmployeeToDepartmentRequest struct {
	CompanyUUID    string `json:"-"`
	DepartmentUUID string `json:"-"`
	TargetUUID     string `json:"-"`
}
type AddEmployeeToDepartmentResponse struct{}

func (e *AddEmployeeToDepartmentRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if err := utils.ValidateUUID(e.DepartmentUUID); err != nil {
		return err
	}
	return utils.ValidateUUID(e.TargetUUID)
}

type RemoveEmployeeFromDepartmentRequest struct {
	CompanyUUID    string `json:"-"`
	DepartmentUUID string `json:"-"`
	TargetUUID     string `json:"-"`
}
type RemoveEmployeeFromDepartmentResponse struct{}

func (e *RemoveEmployeeFromDepartmentRequest) Validate() error {
	if err := utils.ValidateUUID(e.CompanyUUID); err != nil {
		return err
	}
	if err := utils.ValidateUUID(e.DepartmentUUID); err != nil {
		return err
	}
	return utils.ValidateUUID(e.TargetUUID)
}
