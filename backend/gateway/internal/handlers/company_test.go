package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/company/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/entities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ─── CreateCompany ────────────────────────────────────────────────────────────

func TestCreateCompany(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodPost, "/company/create",
			newCompanyHandler(&mockCompanyClient{
				createCompany: func(_ context.Context, _ *company_proto.CreateCompanyRequest, _ ...grpc.CallOption) (*company_proto.CreateCompanyResponse, error) {
					return &company_proto.CreateCompanyResponse{CompanyUuid: companyID}, nil
				},
			}).CreateCompany,
		)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/company/create", `{"title":"My Company"}`))
		assertStatus(t, resp, fiber.StatusCreated)

		res := decodeBody[entities.CreateCompanyResponse](t, resp)
		if res.CompanyUUID != companyID {
			t.Errorf("expected company_uuid=%q, got %q", companyID, res.CompanyUUID)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		app := newApp(http.MethodPost, "/company/create",
			newCompanyHandler(&mockCompanyClient{}).CreateCompany,
		)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/company/create", `{"title":""}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - internal", func(t *testing.T) {
		app := newApp(http.MethodPost, "/company/create",
			newCompanyHandler(&mockCompanyClient{
				createCompany: func(_ context.Context, _ *company_proto.CreateCompanyRequest, _ ...grpc.CallOption) (*company_proto.CreateCompanyResponse, error) {
					return nil, grpcErr(codes.Internal)
				},
			}).CreateCompany,
		)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/company/create", `{"title":"My Company"}`))
		assertStatus(t, resp, fiber.StatusInternalServerError)
	})
}

// ─── GetCompany ───────────────────────────────────────────────────────────────

func TestGetCompany(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid",
			newCompanyHandler(&mockCompanyClient{
				getCompany: func(_ context.Context, _ *company_proto.GetCompanyRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyResponse, error) {
					return &company_proto.GetCompanyResponse{
						CompanyUuid: companyID,
						Title:       "My Company",
						Status:      "open",
					}, nil
				},
			}).GetCompany,
		)
		resp, _ := app.Test(getReq("/company/" + companyID))
		assertStatus(t, resp, fiber.StatusOK)

		res := decodeBody[entities.GetCompanyResponse](t, resp)
		if res.CompanyUUID != companyID {
			t.Errorf("expected company_uuid=%q, got %q", companyID, res.CompanyUUID)
		}
		if res.Title != "My Company" {
			t.Errorf("expected title=%q, got %q", "My Company", res.Title)
		}
	})

	t.Run("invalid uuid in path", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid",
			newCompanyHandler(&mockCompanyClient{}).GetCompany,
		)
		resp, _ := app.Test(getReq("/company/not-a-uuid"))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid",
			newCompanyHandler(&mockCompanyClient{
				getCompany: func(_ context.Context, _ *company_proto.GetCompanyRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyResponse, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).GetCompany,
		)
		resp, _ := app.Test(getReq("/company/" + companyID))
		assertStatus(t, resp, fiber.StatusNotFound)
	})
}

// ─── GetCompanies ─────────────────────────────────────────────────────────────

func TestGetCompanies(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/list",
			newCompanyHandler(&mockCompanyClient{
				getCompanies: func(_ context.Context, _ *company_proto.GetCompaniesRequest, _ ...grpc.CallOption) (*company_proto.GetCompaniesResponse, error) {
					return &company_proto.GetCompaniesResponse{
						Companies: []*company_proto.Company{
							{CompanyUuid: companyID, Title: "My Company", Status: "open"},
						},
					}, nil
				},
			}).GetCompanies,
		)
		resp, _ := app.Test(getReq("/company/list?count=10&offset=0"))
		assertStatus(t, resp, fiber.StatusOK)

		res := decodeBody[entities.GetCompaniesResponse](t, resp)
		if len(res.Companies) != 1 {
			t.Errorf("expected 1 company, got %d", len(res.Companies))
		}
	})

	t.Run("count=0 - validation error", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/list",
			newCompanyHandler(&mockCompanyClient{}).GetCompanies,
		)
		resp, _ := app.Test(getReq("/company/list?count=0&offset=0"))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("count>100 - validation error", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/list",
			newCompanyHandler(&mockCompanyClient{}).GetCompanies,
		)
		resp, _ := app.Test(getReq("/company/list?count=200&offset=0"))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/list",
			newCompanyHandler(&mockCompanyClient{
				getCompanies: func(_ context.Context, _ *company_proto.GetCompaniesRequest, _ ...grpc.CallOption) (*company_proto.GetCompaniesResponse, error) {
					return nil, grpcErr(codes.Internal)
				},
			}).GetCompanies,
		)
		resp, _ := app.Test(getReq("/company/list?count=10&offset=0"))
		assertStatus(t, resp, fiber.StatusInternalServerError)
	})
}

// ─── UpdateCompanyTitle ───────────────────────────────────────────────────────

func TestUpdateCompanyTitle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/company/:company_uuid/title",
			newCompanyHandler(&mockCompanyClient{
				updateCompanyTitle: func(_ context.Context, _ *company_proto.UpdateCompanyTitleRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).UpdateCompanyTitle,
		)
		body := `{"title":"New Company Name"}`
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/company/"+companyID+"/title", body))
		assertStatus(t, resp, fiber.StatusOK)
	})

	t.Run("invalid company uuid", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/company/:company_uuid/title",
			newCompanyHandler(&mockCompanyClient{}).UpdateCompanyTitle,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/company/bad-uuid/title", `{"title":"Name"}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("empty title", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/company/:company_uuid/title",
			newCompanyHandler(&mockCompanyClient{}).UpdateCompanyTitle,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/company/"+companyID+"/title", `{"title":""}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - forbidden", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/company/:company_uuid/title",
			newCompanyHandler(&mockCompanyClient{
				updateCompanyTitle: func(_ context.Context, _ *company_proto.UpdateCompanyTitleRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.PermissionDenied)
				},
			}).UpdateCompanyTitle,
		)
		body := `{"title":"New Name"}`
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/company/"+companyID+"/title", body))
		assertStatus(t, resp, fiber.StatusForbidden)
	})
}

// ─── UpdateCompanyStatus ──────────────────────────────────────────────────────

func TestUpdateCompanyStatus(t *testing.T) {
	t.Run("success - open", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/company/:company_uuid/status",
			newCompanyHandler(&mockCompanyClient{
				updateCompanyStatus: func(_ context.Context, _ *company_proto.UpdateCompanyStatusRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).UpdateCompanyStatus,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/company/"+companyID+"/status", `{"status":"open"}`))
		assertStatus(t, resp, fiber.StatusOK)
	})

	t.Run("invalid status", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/company/:company_uuid/status",
			newCompanyHandler(&mockCompanyClient{}).UpdateCompanyStatus,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/company/"+companyID+"/status", `{"status":"suspended"}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/company/:company_uuid/status",
			newCompanyHandler(&mockCompanyClient{
				updateCompanyStatus: func(_ context.Context, _ *company_proto.UpdateCompanyStatusRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).UpdateCompanyStatus,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/company/"+companyID+"/status", `{"status":"close"}`))
		assertStatus(t, resp, fiber.StatusNotFound)
	})
}

// ─── DeleteCompany ────────────────────────────────────────────────────────────

func TestDeleteCompany(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/company/:company_uuid",
			newCompanyHandler(&mockCompanyClient{
				deleteCompany: func(_ context.Context, _ *company_proto.DeleteCompanyRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).DeleteCompany,
		)
		resp, _ := app.Test(deleteReq("/company/" + companyID))
		assertStatus(t, resp, fiber.StatusOK)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/company/:company_uuid",
			newCompanyHandler(&mockCompanyClient{}).DeleteCompany,
		)
		resp, _ := app.Test(deleteReq("/company/bad"))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - forbidden", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/company/:company_uuid",
			newCompanyHandler(&mockCompanyClient{
				deleteCompany: func(_ context.Context, _ *company_proto.DeleteCompanyRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.PermissionDenied)
				},
			}).DeleteCompany,
		)
		resp, _ := app.Test(deleteReq("/company/" + companyID))
		assertStatus(t, resp, fiber.StatusForbidden)
	})
}

// ─── CreateCompanyJoinCode ────────────────────────────────────────────────────

func TestCreateCompanyJoinCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodPost, "/company/:company_uuid/code",
			newCompanyHandler(&mockCompanyClient{
				createCompanyJoinCode: func(_ context.Context, _ *company_proto.CreateCompanyJoinCodeRequest, _ ...grpc.CallOption) (*company_proto.CreateCompanyJoinCodeResponse, error) {
					return &company_proto.CreateCompanyJoinCodeResponse{JoinCode: "123456"}, nil
				},
			}).CreateCompanyJoinCode,
		)
		// code_ttl: 3600 = 1 час (между 60 и 604800)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/company/"+companyID+"/code", `{"code_ttl":3600}`))
		assertStatus(t, resp, fiber.StatusCreated)

		res := decodeBody[entities.CreateCompanyJoinCodeResponse](t, resp)
		if res.Code != "123456" {
			t.Errorf("expected code=%q, got %q", "123456", res.Code)
		}
	})

	t.Run("code_ttl too small (< 60)", func(t *testing.T) {
		app := newApp(http.MethodPost, "/company/:company_uuid/code",
			newCompanyHandler(&mockCompanyClient{}).CreateCompanyJoinCode,
		)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/company/"+companyID+"/code", `{"code_ttl":10}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("code_ttl too large (> 1 week)", func(t *testing.T) {
		app := newApp(http.MethodPost, "/company/:company_uuid/code",
			newCompanyHandler(&mockCompanyClient{}).CreateCompanyJoinCode,
		)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/company/"+companyID+"/code", `{"code_ttl":999999}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error", func(t *testing.T) {
		app := newApp(http.MethodPost, "/company/:company_uuid/code",
			newCompanyHandler(&mockCompanyClient{
				createCompanyJoinCode: func(_ context.Context, _ *company_proto.CreateCompanyJoinCodeRequest, _ ...grpc.CallOption) (*company_proto.CreateCompanyJoinCodeResponse, error) {
					return nil, grpcErr(codes.Internal)
				},
			}).CreateCompanyJoinCode,
		)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/company/"+companyID+"/code", `{"code_ttl":3600}`))
		assertStatus(t, resp, fiber.StatusInternalServerError)
	})
}

// ─── GetCompanyJoinCodes ──────────────────────────────────────────────────────

func TestGetCompanyJoinCodes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/codes",
			newCompanyHandler(&mockCompanyClient{
				getCompanyJoinCodes: func(_ context.Context, _ *company_proto.GetCompanyJoinCodesRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyJoinCodesResponse, error) {
					return &company_proto.GetCompanyJoinCodesResponse{Codes: []string{"123456", "654321"}}, nil
				},
			}).GetCompanyJoinCodes,
		)
		resp, _ := app.Test(getReq("/company/" + companyID + "/codes"))
		assertStatus(t, resp, fiber.StatusOK)

		res := decodeBody[entities.GetCompanyJoinCodesResponse](t, resp)
		if len(res.Codes) != 2 {
			t.Errorf("expected 2 codes, got %d", len(res.Codes))
		}
	})

	t.Run("invalid uuid", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/codes",
			newCompanyHandler(&mockCompanyClient{}).GetCompanyJoinCodes,
		)
		resp, _ := app.Test(getReq("/company/bad-uuid/codes"))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - forbidden", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/codes",
			newCompanyHandler(&mockCompanyClient{
				getCompanyJoinCodes: func(_ context.Context, _ *company_proto.GetCompanyJoinCodesRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyJoinCodesResponse, error) {
					return nil, grpcErr(codes.PermissionDenied)
				},
			}).GetCompanyJoinCodes,
		)
		resp, _ := app.Test(getReq("/company/" + companyID + "/codes"))
		assertStatus(t, resp, fiber.StatusForbidden)
	})
}

// ─── DeleteCompanyJoinCode ────────────────────────────────────────────────────

func TestDeleteCompanyJoinCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/company/:company_uuid/code",
			newCompanyHandler(&mockCompanyClient{
				deleteCompanyJoinCode: func(_ context.Context, _ *company_proto.DeleteCompanyJoinCodeRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).DeleteCompanyJoinCode,
		)
		resp, _ := app.Test(jsonReq(http.MethodDelete, "/company/"+companyID+"/code", `{"code":"123456"}`))
		assertStatus(t, resp, fiber.StatusOK)
	})

	t.Run("invalid join code format", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/company/:company_uuid/code",
			newCompanyHandler(&mockCompanyClient{}).DeleteCompanyJoinCode,
		)
		resp, _ := app.Test(jsonReq(http.MethodDelete, "/company/"+companyID+"/code", `{"code":"abc"}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/company/:company_uuid/code",
			newCompanyHandler(&mockCompanyClient{
				deleteCompanyJoinCode: func(_ context.Context, _ *company_proto.DeleteCompanyJoinCodeRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).DeleteCompanyJoinCode,
		)
		resp, _ := app.Test(jsonReq(http.MethodDelete, "/company/"+companyID+"/code", `{"code":"123456"}`))
		assertStatus(t, resp, fiber.StatusNotFound)
	})
}

// ─── JoinCompany ──────────────────────────────────────────────────────────────

func TestJoinCompany(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodPost, "/company/join",
			newCompanyHandler(&mockCompanyClient{
				joinCompany: func(_ context.Context, _ *company_proto.JoinCompanyRequest, _ ...grpc.CallOption) (*company_proto.JoinCompanyResponse, error) {
					return &company_proto.JoinCompanyResponse{CompanyUuid: companyID, Role: "unemployed"}, nil
				},
			}).JoinCompany,
		)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/company/join", `{"code":"123456"}`))
		assertStatus(t, resp, fiber.StatusOK)

		res := decodeBody[entities.JoinCompanyResponse](t, resp)
		if res.CompanyUUID != companyID {
			t.Errorf("expected company_uuid=%q, got %q", companyID, res.CompanyUUID)
		}
	})

	t.Run("invalid code format (not 6 digits)", func(t *testing.T) {
		app := newApp(http.MethodPost, "/company/join",
			newCompanyHandler(&mockCompanyClient{}).JoinCompany,
		)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/company/join", `{"code":"abc"}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - code not found", func(t *testing.T) {
		app := newApp(http.MethodPost, "/company/join",
			newCompanyHandler(&mockCompanyClient{
				joinCompany: func(_ context.Context, _ *company_proto.JoinCompanyRequest, _ ...grpc.CallOption) (*company_proto.JoinCompanyResponse, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).JoinCompany,
		)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/company/join", `{"code":"999999"}`))
		assertStatus(t, resp, fiber.StatusNotFound)
	})
}

// ─── GetCompanyEmployee ───────────────────────────────────────────────────────

func TestGetCompanyEmployee(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/employee/:employee_uuid/info",
			newCompanyHandler(&mockCompanyClient{
				getCompanyEmployee: func(_ context.Context, _ *company_proto.GetCompanyEmployeeRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error) {
					return &company_proto.GetCompanyEmployeeResponse{Role: "engineer", JoinedAt: "2024-01-01"}, nil
				},
			}).GetCompanyEmployee,
		)
		url := fmt.Sprintf("/company/%s/employee/%s/info", companyID, targetID)
		resp, _ := app.Test(getReq(url))
		assertStatus(t, resp, fiber.StatusOK)

		res := decodeBody[entities.GetCompanyEmployeeResponse](t, resp)
		if res.Role != "engineer" {
			t.Errorf("expected role=%q, got %q", "engineer", res.Role)
		}
	})

	t.Run("invalid employee uuid", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/employee/:employee_uuid/info",
			newCompanyHandler(&mockCompanyClient{}).GetCompanyEmployee,
		)
		resp, _ := app.Test(getReq("/company/" + companyID + "/employee/bad/info"))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/employee/:employee_uuid/info",
			newCompanyHandler(&mockCompanyClient{
				getCompanyEmployee: func(_ context.Context, _ *company_proto.GetCompanyEmployeeRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).GetCompanyEmployee,
		)
		url := fmt.Sprintf("/company/%s/employee/%s/info", companyID, targetID)
		resp, _ := app.Test(getReq(url))
		assertStatus(t, resp, fiber.StatusNotFound)
	})
}

// ─── GetCompanyEmployees ──────────────────────────────────────────────────────

func TestGetCompanyEmployees(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/employees/list",
			newCompanyHandler(&mockCompanyClient{
				getCompanyEmployees: func(_ context.Context, _ *company_proto.GetCompanyEmployeesRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeesResponse, error) {
					return &company_proto.GetCompanyEmployeesResponse{
						Employees: []*company_proto.Employee{
							{UserUuid: targetID, Role: "engineer", JoinedAt: "2024-01-01"},
						},
					}, nil
				},
			}).GetCompanyEmployees,
		)
		url := "/company/" + companyID + "/employees/list?count=10&offset=0"
		resp, _ := app.Test(getReq(url))
		assertStatus(t, resp, fiber.StatusOK)

		res := decodeBody[entities.GetCompanyEmployeesResponse](t, resp)
		if len(res.Employees) != 1 {
			t.Errorf("expected 1 employee, got %d", len(res.Employees))
		}
	})

	t.Run("invalid role filter", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/employees/list",
			newCompanyHandler(&mockCompanyClient{}).GetCompanyEmployees,
		)
		url := "/company/" + companyID + "/employees/list?count=10&offset=0&role=superadmin"
		resp, _ := app.Test(getReq(url))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("count out of range", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/employees/list",
			newCompanyHandler(&mockCompanyClient{}).GetCompanyEmployees,
		)
		url := "/company/" + companyID + "/employees/list?count=0&offset=0"
		resp, _ := app.Test(getReq(url))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/employees/list",
			newCompanyHandler(&mockCompanyClient{
				getCompanyEmployees: func(_ context.Context, _ *company_proto.GetCompanyEmployeesRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeesResponse, error) {
					return nil, grpcErr(codes.Internal)
				},
			}).GetCompanyEmployees,
		)
		resp, _ := app.Test(getReq("/company/" + companyID + "/employees/list?count=10&offset=0"))
		assertStatus(t, resp, fiber.StatusInternalServerError)
	})
}

// ─── GetCompanyEmployeesSummary ───────────────────────────────────────────────

func TestGetCompanyEmployeesSummary(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/employees/summary",
			newCompanyHandler(&mockCompanyClient{
				getCompanyEmployeesSummary: func(_ context.Context, _ *company_proto.GetCompanyEmployeesSummaryRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeesSummaryResponse, error) {
					return &company_proto.GetCompanyEmployeesSummaryResponse{
						ChiefCount:     1,
						ManagerCount:   3,
						EngineerCount:  5,
						UnemployedCount: 2,
					}, nil
				},
			}).GetCompanyEmployeesSummary,
		)
		resp, _ := app.Test(getReq("/company/" + companyID + "/employees/summary"))
		assertStatus(t, resp, fiber.StatusOK)

		var res entities.GetCompanyEmployeesSummaryResponse
		json.NewDecoder(resp.Body).Decode(&res)
		if res.ChiefCount != 1 {
			t.Errorf("expected chief_count=1, got %d", res.ChiefCount)
		}
		if res.EngineerCount != 5 {
			t.Errorf("expected engineer_count=5, got %d", res.EngineerCount)
		}
	})

	t.Run("invalid uuid", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/employees/summary",
			newCompanyHandler(&mockCompanyClient{}).GetCompanyEmployeesSummary,
		)
		resp, _ := app.Test(getReq("/company/not-uuid/employees/summary"))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/employees/summary",
			newCompanyHandler(&mockCompanyClient{
				getCompanyEmployeesSummary: func(_ context.Context, _ *company_proto.GetCompanyEmployeesSummaryRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeesSummaryResponse, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).GetCompanyEmployeesSummary,
		)
		resp, _ := app.Test(getReq("/company/" + companyID + "/employees/summary"))
		assertStatus(t, resp, fiber.StatusNotFound)
	})
}

// ─── UpdateEmployeeRole ───────────────────────────────────────────────────────

func TestUpdateEmployeeRole(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/company/:company_uuid/employee/:employee_uuid/role",
			newCompanyHandler(&mockCompanyClient{
				updateEmployeeRole: func(_ context.Context, _ *company_proto.UpdateEmployeeRoleRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).UpdateEmployeeRole,
		)
		url := fmt.Sprintf("/company/%s/employee/%s/role", companyID, targetID)
		resp, _ := app.Test(jsonReq(http.MethodPatch, url, `{"role":"engineer"}`))
		assertStatus(t, resp, fiber.StatusOK)
	})

	t.Run("invalid role", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/company/:company_uuid/employee/:employee_uuid/role",
			newCompanyHandler(&mockCompanyClient{}).UpdateEmployeeRole,
		)
		url := fmt.Sprintf("/company/%s/employee/%s/role", companyID, targetID)
		resp, _ := app.Test(jsonReq(http.MethodPatch, url, `{"role":"superadmin"}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("invalid target uuid", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/company/:company_uuid/employee/:employee_uuid/role",
			newCompanyHandler(&mockCompanyClient{}).UpdateEmployeeRole,
		)
		url := fmt.Sprintf("/company/%s/employee/bad-uuid/role", companyID)
		resp, _ := app.Test(jsonReq(http.MethodPatch, url, `{"role":"engineer"}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - forbidden", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/company/:company_uuid/employee/:employee_uuid/role",
			newCompanyHandler(&mockCompanyClient{
				updateEmployeeRole: func(_ context.Context, _ *company_proto.UpdateEmployeeRoleRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.PermissionDenied)
				},
			}).UpdateEmployeeRole,
		)
		url := fmt.Sprintf("/company/%s/employee/%s/role", companyID, targetID)
		resp, _ := app.Test(jsonReq(http.MethodPatch, url, `{"role":"manager"}`))
		assertStatus(t, resp, fiber.StatusForbidden)
	})
}

// ─── RemoveCompanyEmployee ────────────────────────────────────────────────────

func TestRemoveCompanyEmployee(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/company/:company_uuid/employee/:employee_uuid",
			newCompanyHandler(&mockCompanyClient{
				removeCompanyEmployee: func(_ context.Context, _ *company_proto.RemoveCompanyEmployeeRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).RemoveCompanyEmployee,
		)
		url := fmt.Sprintf("/company/%s/employee/%s", companyID, targetID)
		resp, _ := app.Test(deleteReq(url))
		assertStatus(t, resp, fiber.StatusOK)
	})

	t.Run("invalid employee uuid", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/company/:company_uuid/employee/:employee_uuid",
			newCompanyHandler(&mockCompanyClient{}).RemoveCompanyEmployee,
		)
		url := fmt.Sprintf("/company/%s/employee/bad", companyID)
		resp, _ := app.Test(deleteReq(url))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/company/:company_uuid/employee/:employee_uuid",
			newCompanyHandler(&mockCompanyClient{
				removeCompanyEmployee: func(_ context.Context, _ *company_proto.RemoveCompanyEmployeeRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).RemoveCompanyEmployee,
		)
		url := fmt.Sprintf("/company/%s/employee/%s", companyID, targetID)
		resp, _ := app.Test(deleteReq(url))
		assertStatus(t, resp, fiber.StatusNotFound)
	})
}
