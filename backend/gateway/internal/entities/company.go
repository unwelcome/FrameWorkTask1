package entities

import (
	"fmt"
	"strings"

	"github.com/unwelcome/FrameWorkTask1/backend/shared/helpers"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/validate"
)

// ─── Shared response types ────────────────────────────────────────────────────

type DepartmentListItem struct {
	DepartmentUUID string `json:"department_uuid"`
	Title          string `json:"title"`
}

// ─── CreateCompany ────────────────────────────────────────────────────────────

type CreateCompanyRequest struct {
	Title string `json:"title"`
}
type CreateCompanyResponse struct {
	CompanyUUID string `json:"company_uuid"`
}

func (e *CreateCompanyRequest) Validate() error {
	e.Title = strings.TrimSpace(e.Title)
	if err := validate.CompanyTitle(e.Title); err != nil {
		return err
	}
	return nil
}

// ─── GetCompany ───────────────────────────────────────────────────────────────

type GetCompanyRequest struct {
	CompanyUUID string `json:"-"`
}
type GetCompanyResponse struct {
	CompanyUUID string `json:"company_uuid"`
	Title       string `json:"title"`
	Status      string `json:"status"`
}

func (e *GetCompanyRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	return nil
}

// ─── GetCompanies ─────────────────────────────────────────────────────────────

type GetCompaniesRequest struct {
	Offset int64 `query:"offset"`
	Count  int64 `query:"count"`
}
type GetCompaniesResponse struct {
	Companies []*GetCompanyResponse `json:"companies"`
}

func (e *GetCompaniesRequest) Validate() error {
	if err := validate.Number(int(e.Offset), validate.IntPtr(0), nil, "offset"); err != nil {
		return err
	}
	if err := validate.Number(int(e.Count), validate.IntPtr(1), validate.IntPtr(100), "count"); err != nil {
		return err
	}
	return nil
}

// ─── GetUserCompanies ─────────────────────────────────────────────────────────

type GetUserCompaniesRequest struct{}
type GetUserCompaniesResponse struct {
	Companies []*GetCompanyResponse `json:"companies"`
}

// ─── UpdateCompanyTitle ───────────────────────────────────────────────────────

type UpdateCompanyTitleRequest struct {
	CompanyUUID string `json:"-"`
	Title       string `json:"title"`
}
type UpdateCompanyTitleResponse struct{}

func (e *UpdateCompanyTitleRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.Title = strings.TrimSpace(e.Title)
	if err := validate.CompanyTitle(e.Title); err != nil {
		return err
	}
	return nil
}

// ─── UpdateCompanyStatus ──────────────────────────────────────────────────────

type UpdateCompanyStatusRequest struct {
	CompanyUUID string `json:"-"`
	Status      string `json:"status"`
}
type UpdateCompanyStatusResponse struct{}

func (e *UpdateCompanyStatusRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.Status = strings.TrimSpace(e.Status)
	if !helpers.Contains([]string{"open", "close"}, e.Status) {
		return fmt.Errorf("incorrect company status")
	}
	return nil
}

// ─── DeleteCompany ────────────────────────────────────────────────────────────

type DeleteCompanyRequest struct {
	CompanyUUID string `json:"-"`
}
type DeleteCompanyResponse struct{}

func (e *DeleteCompanyRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	return nil
}

// ─── CreateCompanyJoinCode ────────────────────────────────────────────────────

type CreateCompanyJoinCodeRequest struct {
	CompanyUUID string `json:"-"`
	CodeTTL     int64  `json:"code_ttl"`
}
type CreateCompanyJoinCodeResponse struct {
	Code string `json:"code"`
}

func (e *CreateCompanyJoinCodeRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	if err := validate.Number(int(e.CodeTTL), validate.IntPtr(60), validate.IntPtr(60*60*24*7), "code_ttl"); err != nil {
		return err
	}
	return nil
}

// ─── GetCompanyJoinCodes ──────────────────────────────────────────────────────

type GetCompanyJoinCodesRequest struct {
	CompanyUUID string `json:"-"`
}
type GetCompanyJoinCodesResponse struct {
	Codes []string `json:"codes"`
}

func (e *GetCompanyJoinCodesRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	return nil
}

// ─── DeleteCompanyJoinCode ────────────────────────────────────────────────────

type DeleteCompanyJoinCodeRequest struct {
	CompanyUUID string `json:"-"`
	Code        string `json:"code"`
}
type DeleteCompanyJoinCodeResponse struct{}

func (e *DeleteCompanyJoinCodeRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.Code = strings.TrimSpace(e.Code)
	if err := validate.CompanyJoinCode(e.Code); err != nil {
		return err
	}
	return nil
}

// ─── JoinCompany ──────────────────────────────────────────────────────────────

type JoinCompanyRequest struct {
	Code string `json:"code"`
}
type JoinCompanyResponse struct {
	CompanyUUID string `json:"company_uuid"`
	Role        string `json:"role"`
}

func (e *JoinCompanyRequest) Validate() error {
	e.Code = strings.TrimSpace(e.Code)
	if err := validate.CompanyJoinCode(e.Code); err != nil {
		return err
	}
	return nil
}

// ─── GetCompanyEmployee ───────────────────────────────────────────────────────

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
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.TargetUUID = strings.TrimSpace(e.TargetUUID)
	if err := validate.UUID(e.TargetUUID); err != nil {
		return err
	}
	return nil
}

// ─── GetCompanyEmployees ──────────────────────────────────────────────────────

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
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.DepartmentUUID = strings.TrimSpace(e.DepartmentUUID)
	if e.DepartmentUUID != "" {
		if err := validate.UUID(e.DepartmentUUID); err != nil {
			return err
		}
	}
	e.Role = strings.TrimSpace(e.Role)
	if !helpers.Contains([]string{"", "unemployed", "inspector", "engineer", "manager", "analytic", "chief"}, e.Role) {
		return fmt.Errorf("incorrect employee role")
	}
	if err := validate.Number(int(e.Offset), validate.IntPtr(0), nil, "offset"); err != nil {
		return err
	}
	if err := validate.Number(int(e.Count), validate.IntPtr(1), validate.IntPtr(100), "count"); err != nil {
		return err
	}
	return nil
}

// ─── GetCompanyEmployeesSummary ───────────────────────────────────────────────

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
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.DepartmentUUID = strings.TrimSpace(e.DepartmentUUID)
	if e.DepartmentUUID != "" {
		if err := validate.UUID(e.DepartmentUUID); err != nil {
			return err
		}
	}
	return nil
}

// ─── UpdateEmployeeRole ───────────────────────────────────────────────────────

type UpdateEmployeeRoleRequest struct {
	CompanyUUID string `json:"-"`
	TargetUUID  string `json:"-"`
	Role        string `json:"role"`
}
type UpdateEmployeeRoleResponse struct{}

func (e *UpdateEmployeeRoleRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.TargetUUID = strings.TrimSpace(e.TargetUUID)
	if err := validate.UUID(e.TargetUUID); err != nil {
		return err
	}
	e.Role = strings.TrimSpace(e.Role)
	if !helpers.Contains([]string{"unemployed", "engineer", "manager", "analytic", "inspector", "chief"}, e.Role) {
		return fmt.Errorf("incorrect employee role")
	}
	return nil
}

// ─── RemoveCompanyEmployee ────────────────────────────────────────────────────

type RemoveCompanyEmployeeRequest struct {
	TargetUUID  string `json:"-"`
	CompanyUUID string `json:"-"`
}
type RemoveCompanyEmployeeResponse struct{}

func (e *RemoveCompanyEmployeeRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.TargetUUID = strings.TrimSpace(e.TargetUUID)
	if err := validate.UUID(e.TargetUUID); err != nil {
		return err
	}
	return nil
}

// ─── CreateDepartment ─────────────────────────────────────────────────────────

type CreateDepartmentRequest struct {
	CompanyUUID string `json:"-"`
	Title       string `json:"title"`
}
type CreateDepartmentResponse struct {
	DepartmentUUID string `json:"department_uuid"`
}

func (e *CreateDepartmentRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.Title = strings.TrimSpace(e.Title)
	if err := validate.DepartmentTitle(e.Title); err != nil {
		return err
	}
	return nil
}

// ─── GetDepartment ────────────────────────────────────────────────────────────

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
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.DepartmentUUID = strings.TrimSpace(e.DepartmentUUID)
	if err := validate.UUID(e.DepartmentUUID); err != nil {
		return err
	}
	return nil
}

// ─── GetCompanyDepartments ────────────────────────────────────────────────────

type GetCompanyDepartmentsRequest struct {
	CompanyUUID string `json:"-"`
	Offset      int64  `query:"offset"`
	Count       int64  `query:"count"`
}
type GetCompanyDepartmentsResponse struct {
	Departments []*DepartmentListItem `json:"departments"`
}

func (e *GetCompanyDepartmentsRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	if err := validate.Number(int(e.Offset), validate.IntPtr(0), nil, "offset"); err != nil {
		return err
	}
	return validate.Number(int(e.Count), validate.IntPtr(1), validate.IntPtr(100), "count")
}

// ─── UpdateDepartmentTitle ────────────────────────────────────────────────────

type UpdateDepartmentTitleRequest struct {
	CompanyUUID    string `json:"-"`
	DepartmentUUID string `json:"-"`
	Title          string `json:"title"`
}
type UpdateDepartmentTitleResponse struct{}

func (e *UpdateDepartmentTitleRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.DepartmentUUID = strings.TrimSpace(e.DepartmentUUID)
	if err := validate.UUID(e.DepartmentUUID); err != nil {
		return err
	}
	e.Title = strings.TrimSpace(e.Title)
	if err := validate.DepartmentTitle(e.Title); err != nil {
		return err
	}
	return nil
}

// ─── DeleteDepartment ─────────────────────────────────────────────────────────

type DeleteDepartmentRequest struct {
	CompanyUUID    string `json:"-"`
	DepartmentUUID string `json:"-"`
}
type DeleteDepartmentResponse struct{}

func (e *DeleteDepartmentRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.DepartmentUUID = strings.TrimSpace(e.DepartmentUUID)
	if err := validate.UUID(e.DepartmentUUID); err != nil {
		return err
	}
	return nil
}

// ─── AddEmployeeToDepartment ──────────────────────────────────────────────────

type AddEmployeeToDepartmentRequest struct {
	CompanyUUID    string `json:"-"`
	DepartmentUUID string `json:"-"`
	TargetUUID     string `json:"-"`
}
type AddEmployeeToDepartmentResponse struct{}

func (e *AddEmployeeToDepartmentRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.DepartmentUUID = strings.TrimSpace(e.DepartmentUUID)
	if err := validate.UUID(e.DepartmentUUID); err != nil {
		return err
	}
	e.TargetUUID = strings.TrimSpace(e.TargetUUID)
	if err := validate.UUID(e.TargetUUID); err != nil {
		return err
	}
	return nil
}

// ─── RemoveEmployeeFromDepartment ─────────────────────────────────────────────

type RemoveEmployeeFromDepartmentRequest struct {
	CompanyUUID    string `json:"-"`
	DepartmentUUID string `json:"-"`
	TargetUUID     string `json:"-"`
}
type RemoveEmployeeFromDepartmentResponse struct{}

func (e *RemoveEmployeeFromDepartmentRequest) Validate() error {
	e.CompanyUUID = strings.TrimSpace(e.CompanyUUID)
	if err := validate.UUID(e.CompanyUUID); err != nil {
		return err
	}
	e.DepartmentUUID = strings.TrimSpace(e.DepartmentUUID)
	if err := validate.UUID(e.DepartmentUUID); err != nil {
		return err
	}
	e.TargetUUID = strings.TrimSpace(e.TargetUUID)
	if err := validate.UUID(e.TargetUUID); err != nil {
		return err
	}
	return nil
}
