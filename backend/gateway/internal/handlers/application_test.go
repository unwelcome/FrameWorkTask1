package handlers

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	application_proto "github.com/unwelcome/FrameWorkTask1/backend/application/api/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/entities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ─── CreateApplication ────────────────────────────────────────────────────────

func TestCreateApplication(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodPost, "/application/create",
			newApplicationHandler(&mockApplicationClient{
				createApplication: func(_ context.Context, _ *application_proto.CreateApplicationRequest, _ ...grpc.CallOption) (*application_proto.CreateApplicationResponse, error) {
					return &application_proto.CreateApplicationResponse{ApplicationUuid: appID}, nil
				},
			}).CreateApplication,
		)
		body := fmt.Sprintf(`{"company_uuid":%q,"title":"Pipeline broken","description":"The main pipeline is down"}`, companyID)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/application/create", body))
		assertStatus(t, resp, fiber.StatusCreated)

		res := decodeBody[entities.CreateApplicationResponse](t, resp)
		if res.ApplicationUUID != appID {
			t.Errorf("expected application_uuid=%q, got %q", appID, res.ApplicationUUID)
		}
	})

	t.Run("missing title", func(t *testing.T) {
		app := newApp(http.MethodPost, "/application/create",
			newApplicationHandler(&mockApplicationClient{}).CreateApplication,
		)
		body := fmt.Sprintf(`{"company_uuid":%q,"description":"desc"}`, companyID)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/application/create", body))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("missing description", func(t *testing.T) {
		app := newApp(http.MethodPost, "/application/create",
			newApplicationHandler(&mockApplicationClient{}).CreateApplication,
		)
		body := fmt.Sprintf(`{"company_uuid":%q,"title":"Fix pipeline"}`, companyID)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/application/create", body))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - permission denied", func(t *testing.T) {
		app := newApp(http.MethodPost, "/application/create",
			newApplicationHandler(&mockApplicationClient{
				createApplication: func(_ context.Context, _ *application_proto.CreateApplicationRequest, _ ...grpc.CallOption) (*application_proto.CreateApplicationResponse, error) {
					return nil, grpcErr(codes.PermissionDenied)
				},
			}).CreateApplication,
		)
		body := fmt.Sprintf(`{"company_uuid":%q,"title":"Fix pipeline","description":"desc"}`, companyID)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/application/create", body))
		assertStatus(t, resp, fiber.StatusForbidden)
	})

	t.Run("service error - internal", func(t *testing.T) {
		app := newApp(http.MethodPost, "/application/create",
			newApplicationHandler(&mockApplicationClient{
				createApplication: func(_ context.Context, _ *application_proto.CreateApplicationRequest, _ ...grpc.CallOption) (*application_proto.CreateApplicationResponse, error) {
					return nil, grpcErr(codes.Internal)
				},
			}).CreateApplication,
		)
		body := fmt.Sprintf(`{"company_uuid":%q,"title":"Fix pipeline","description":"desc"}`, companyID)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/application/create", body))
		assertStatus(t, resp, fiber.StatusInternalServerError)
	})
}

// ─── GetApplication ───────────────────────────────────────────────────────────

func TestGetApplication(t *testing.T) {
	t.Run("success with fix logs", func(t *testing.T) {
		app := newApp(http.MethodGet, "/application/:application_uuid",
			newApplicationHandler(&mockApplicationClient{
				getApplication: func(_ context.Context, _ *application_proto.GetApplicationRequest, _ ...grpc.CallOption) (*application_proto.GetApplicationResponse, error) {
					return &application_proto.GetApplicationResponse{
						Application: &application_proto.Application{
							ApplicationUuid: appID,
							CompanyUuid:     companyID,
							Title:           "Fix pipeline",
							Status:          "assigned",
							FixLogs: []*application_proto.FixLog{
								{Text: "Replaced valve A", CreatedAt: "2024-01-02", CreatedBy: targetID},
							},
						},
					}, nil
				},
			}).GetApplication,
		)
		resp, _ := app.Test(getReq("/application/" + appID))
		assertStatus(t, resp, fiber.StatusOK)

		res := decodeBody[entities.GetApplicationResponse](t, resp)
		if res.Application == nil {
			t.Fatal("expected non-nil application")
		}
		if res.Application.ApplicationUUID != appID {
			t.Errorf("expected uuid=%q, got %q", appID, res.Application.ApplicationUUID)
		}
		if len(res.Application.FixLogs) != 1 {
			t.Errorf("expected 1 fix log, got %d", len(res.Application.FixLogs))
		}
	})

	t.Run("invalid uuid in path", func(t *testing.T) {
		app := newApp(http.MethodGet, "/application/:application_uuid",
			newApplicationHandler(&mockApplicationClient{}).GetApplication,
		)
		resp, _ := app.Test(getReq("/application/not-a-uuid"))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found", func(t *testing.T) {
		app := newApp(http.MethodGet, "/application/:application_uuid",
			newApplicationHandler(&mockApplicationClient{
				getApplication: func(_ context.Context, _ *application_proto.GetApplicationRequest, _ ...grpc.CallOption) (*application_proto.GetApplicationResponse, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).GetApplication,
		)
		resp, _ := app.Test(getReq("/application/" + appID))
		assertStatus(t, resp, fiber.StatusNotFound)
	})

	t.Run("service error - permission denied", func(t *testing.T) {
		app := newApp(http.MethodGet, "/application/:application_uuid",
			newApplicationHandler(&mockApplicationClient{
				getApplication: func(_ context.Context, _ *application_proto.GetApplicationRequest, _ ...grpc.CallOption) (*application_proto.GetApplicationResponse, error) {
					return nil, grpcErr(codes.PermissionDenied)
				},
			}).GetApplication,
		)
		resp, _ := app.Test(getReq("/application/" + appID))
		assertStatus(t, resp, fiber.StatusForbidden)
	})
}

// ─── GetApplications ──────────────────────────────────────────────────────────

func TestGetApplications(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/applications/list",
			newApplicationHandler(&mockApplicationClient{
				getApplications: func(_ context.Context, _ *application_proto.GetApplicationsRequest, _ ...grpc.CallOption) (*application_proto.GetApplicationsResponse, error) {
					return &application_proto.GetApplicationsResponse{
						Applications: []*application_proto.Application{
							{ApplicationUuid: appID, Title: "Fix pipeline", Status: "created"},
							{ApplicationUuid: targetID, Title: "Replace valve", Status: "assigned"},
						},
					}, nil
				},
			}).GetApplications,
		)
		url := "/company/" + companyID + "/applications/list?count=10&offset=0"
		resp, _ := app.Test(getReq(url))
		assertStatus(t, resp, fiber.StatusOK)

		res := decodeBody[entities.GetApplicationsResponse](t, resp)
		if len(res.Applications) != 2 {
			t.Errorf("expected 2 applications, got %d", len(res.Applications))
		}
	})

	t.Run("invalid company uuid", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/applications/list",
			newApplicationHandler(&mockApplicationClient{}).GetApplications,
		)
		resp, _ := app.Test(getReq("/company/bad-uuid/applications/list?count=10&offset=0"))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - permission denied", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/applications/list",
			newApplicationHandler(&mockApplicationClient{
				getApplications: func(_ context.Context, _ *application_proto.GetApplicationsRequest, _ ...grpc.CallOption) (*application_proto.GetApplicationsResponse, error) {
					return nil, grpcErr(codes.PermissionDenied)
				},
			}).GetApplications,
		)
		resp, _ := app.Test(getReq("/company/" + companyID + "/applications/list?count=10&offset=0"))
		assertStatus(t, resp, fiber.StatusForbidden)
	})

	t.Run("service error - internal", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/applications/list",
			newApplicationHandler(&mockApplicationClient{
				getApplications: func(_ context.Context, _ *application_proto.GetApplicationsRequest, _ ...grpc.CallOption) (*application_proto.GetApplicationsResponse, error) {
					return nil, grpcErr(codes.Internal)
				},
			}).GetApplications,
		)
		resp, _ := app.Test(getReq("/company/" + companyID + "/applications/list?count=10&offset=0"))
		assertStatus(t, resp, fiber.StatusInternalServerError)
	})
}

// ─── GetCompanyApplicationStatistic ──────────────────────────────────────────

func TestGetCompanyApplicationStatistic(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/applications/statistic",
			newApplicationHandler(&mockApplicationClient{
				getCompanyApplicationStatistic: func(_ context.Context, _ *application_proto.GetCompanyApplicationStatisticRequest, _ ...grpc.CallOption) (*application_proto.GetCompanyApplicationStatisticResponse, error) {
					return &application_proto.GetCompanyApplicationStatisticResponse{
						Created:    3,
						Assigned:   2,
						InProgress: 1,
						Completed:  5,
					}, nil
				},
			}).GetCompanyApplicationStatistic,
		)
		resp, _ := app.Test(getReq("/company/" + companyID + "/applications/statistic"))
		assertStatus(t, resp, fiber.StatusOK)

		res := decodeBody[entities.GetCompanyApplicationStatisticResponse](t, resp)
		if res.Created != 3 {
			t.Errorf("expected created=3, got %d", res.Created)
		}
		if res.Completed != 5 {
			t.Errorf("expected completed=5, got %d", res.Completed)
		}
	})

	t.Run("invalid company uuid", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/applications/statistic",
			newApplicationHandler(&mockApplicationClient{}).GetCompanyApplicationStatistic,
		)
		resp, _ := app.Test(getReq("/company/bad/applications/statistic"))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - permission denied", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/applications/statistic",
			newApplicationHandler(&mockApplicationClient{
				getCompanyApplicationStatistic: func(_ context.Context, _ *application_proto.GetCompanyApplicationStatisticRequest, _ ...grpc.CallOption) (*application_proto.GetCompanyApplicationStatisticResponse, error) {
					return nil, grpcErr(codes.PermissionDenied)
				},
			}).GetCompanyApplicationStatistic,
		)
		resp, _ := app.Test(getReq("/company/" + companyID + "/applications/statistic"))
		assertStatus(t, resp, fiber.StatusForbidden)
	})
}

// ─── GetEmployeeApplicationStatistic ─────────────────────────────────────────

func TestGetEmployeeApplicationStatistic(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/employee/:employee_uuid/applications/statistic",
			newApplicationHandler(&mockApplicationClient{
				getEmployeeApplicationStatistic: func(_ context.Context, _ *application_proto.GetEmployeeApplicationStatisticRequest, _ ...grpc.CallOption) (*application_proto.GetEmployeeApplicationStatisticResponse, error) {
					return &application_proto.GetEmployeeApplicationStatisticResponse{
						Assigned:   2,
						InProgress: 3,
					}, nil
				},
			}).GetEmployeeApplicationStatistic,
		)
		url := fmt.Sprintf("/company/%s/employee/%s/applications/statistic", companyID, targetID)
		resp, _ := app.Test(getReq(url))
		assertStatus(t, resp, fiber.StatusOK)

		res := decodeBody[entities.GetEmployeeApplicationStatisticResponse](t, resp)
		if res.InProgress != 3 {
			t.Errorf("expected in_progress=3, got %d", res.InProgress)
		}
	})

	t.Run("invalid employee uuid", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/employee/:employee_uuid/applications/statistic",
			newApplicationHandler(&mockApplicationClient{}).GetEmployeeApplicationStatistic,
		)
		resp, _ := app.Test(getReq("/company/" + companyID + "/employee/bad/applications/statistic"))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found (employee not in company)", func(t *testing.T) {
		app := newApp(http.MethodGet, "/company/:company_uuid/employee/:employee_uuid/applications/statistic",
			newApplicationHandler(&mockApplicationClient{
				getEmployeeApplicationStatistic: func(_ context.Context, _ *application_proto.GetEmployeeApplicationStatisticRequest, _ ...grpc.CallOption) (*application_proto.GetEmployeeApplicationStatisticResponse, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).GetEmployeeApplicationStatistic,
		)
		url := fmt.Sprintf("/company/%s/employee/%s/applications/statistic", companyID, targetID)
		resp, _ := app.Test(getReq(url))
		assertStatus(t, resp, fiber.StatusNotFound)
	})
}

// ─── UpdateApplicationStatus ──────────────────────────────────────────────────

func TestUpdateApplicationStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/application/:application_uuid/status",
			newApplicationHandler(&mockApplicationClient{
				updateApplicationStatus: func(_ context.Context, _ *application_proto.UpdateApplicationStatusRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).UpdateApplicationStatus,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/application/"+appID+"/status", `{"status":"in_progress"}`))
		assertStatus(t, resp, fiber.StatusOK)
	})

	t.Run("missing status in body", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/application/:application_uuid/status",
			newApplicationHandler(&mockApplicationClient{}).UpdateApplicationStatus,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/application/"+appID+"/status", `{}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("invalid application uuid", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/application/:application_uuid/status",
			newApplicationHandler(&mockApplicationClient{}).UpdateApplicationStatus,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/application/bad/status", `{"status":"in_progress"}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - permission denied", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/application/:application_uuid/status",
			newApplicationHandler(&mockApplicationClient{
				updateApplicationStatus: func(_ context.Context, _ *application_proto.UpdateApplicationStatusRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.PermissionDenied)
				},
			}).UpdateApplicationStatus,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/application/"+appID+"/status", `{"status":"completed"}`))
		assertStatus(t, resp, fiber.StatusForbidden)
	})

	t.Run("service error - failed precondition (wrong current status)", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/application/:application_uuid/status",
			newApplicationHandler(&mockApplicationClient{
				updateApplicationStatus: func(_ context.Context, _ *application_proto.UpdateApplicationStatusRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.FailedPrecondition)
				},
			}).UpdateApplicationStatus,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/application/"+appID+"/status", `{"status":"in_progress"}`))
		assertStatus(t, resp, fiber.StatusPreconditionFailed)
	})
}

// ─── AssignApplicationToEmployee ─────────────────────────────────────────────

func TestAssignApplicationToEmployee(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/application/:application_uuid/assign",
			newApplicationHandler(&mockApplicationClient{
				assignApplicationToEmployee: func(_ context.Context, _ *application_proto.AssignApplicationToEmployeeRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).AssignApplicationToEmployee,
		)
		body := fmt.Sprintf(`{"target_uuid":%q}`, targetID)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/application/"+appID+"/assign", body))
		assertStatus(t, resp, fiber.StatusOK)
	})

	t.Run("invalid target uuid in body", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/application/:application_uuid/assign",
			newApplicationHandler(&mockApplicationClient{}).AssignApplicationToEmployee,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/application/"+appID+"/assign", `{"target_uuid":"bad"}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("invalid application uuid", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/application/:application_uuid/assign",
			newApplicationHandler(&mockApplicationClient{}).AssignApplicationToEmployee,
		)
		body := fmt.Sprintf(`{"target_uuid":%q}`, targetID)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/application/bad/assign", body))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/application/:application_uuid/assign",
			newApplicationHandler(&mockApplicationClient{
				assignApplicationToEmployee: func(_ context.Context, _ *application_proto.AssignApplicationToEmployeeRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).AssignApplicationToEmployee,
		)
		body := fmt.Sprintf(`{"target_uuid":%q}`, targetID)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/application/"+appID+"/assign", body))
		assertStatus(t, resp, fiber.StatusNotFound)
	})

	t.Run("service error - invalid argument (target not engineer)", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/application/:application_uuid/assign",
			newApplicationHandler(&mockApplicationClient{
				assignApplicationToEmployee: func(_ context.Context, _ *application_proto.AssignApplicationToEmployeeRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.InvalidArgument)
				},
			}).AssignApplicationToEmployee,
		)
		body := fmt.Sprintf(`{"target_uuid":%q}`, targetID)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/application/"+appID+"/assign", body))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})
}

// ─── AddApplicationFixLog ─────────────────────────────────────────────────────

func TestAddApplicationFixLog(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodPost, "/application/:application_uuid/fix-log",
			newApplicationHandler(&mockApplicationClient{
				addApplicationFixLog: func(_ context.Context, _ *application_proto.AddApplicationFixLogRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).AddApplicationFixLog,
		)
		body := `{"log_text":"Replaced valve B7 successfully"}`
		resp, _ := app.Test(jsonReq(http.MethodPost, "/application/"+appID+"/fix-log", body))
		assertStatus(t, resp, fiber.StatusCreated)
	})

	t.Run("missing log_text", func(t *testing.T) {
		app := newApp(http.MethodPost, "/application/:application_uuid/fix-log",
			newApplicationHandler(&mockApplicationClient{}).AddApplicationFixLog,
		)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/application/"+appID+"/fix-log", `{}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("invalid application uuid", func(t *testing.T) {
		app := newApp(http.MethodPost, "/application/:application_uuid/fix-log",
			newApplicationHandler(&mockApplicationClient{}).AddApplicationFixLog,
		)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/application/bad/fix-log", `{"log_text":"Fixed"}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found", func(t *testing.T) {
		app := newApp(http.MethodPost, "/application/:application_uuid/fix-log",
			newApplicationHandler(&mockApplicationClient{
				addApplicationFixLog: func(_ context.Context, _ *application_proto.AddApplicationFixLogRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).AddApplicationFixLog,
		)
		body := `{"log_text":"Replaced valve B7"}`
		resp, _ := app.Test(jsonReq(http.MethodPost, "/application/"+appID+"/fix-log", body))
		assertStatus(t, resp, fiber.StatusNotFound)
	})

	t.Run("service error - permission denied (not responsible engineer)", func(t *testing.T) {
		app := newApp(http.MethodPost, "/application/:application_uuid/fix-log",
			newApplicationHandler(&mockApplicationClient{
				addApplicationFixLog: func(_ context.Context, _ *application_proto.AddApplicationFixLogRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.PermissionDenied)
				},
			}).AddApplicationFixLog,
		)
		body := `{"log_text":"Replaced valve B7"}`
		resp, _ := app.Test(jsonReq(http.MethodPost, "/application/"+appID+"/fix-log", body))
		assertStatus(t, resp, fiber.StatusForbidden)
	})
}

// ─── DeleteApplication ────────────────────────────────────────────────────────

func TestDeleteApplication(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/application/:application_uuid",
			newApplicationHandler(&mockApplicationClient{
				deleteApplication: func(_ context.Context, _ *application_proto.DeleteApplicationRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).DeleteApplication,
		)
		resp, _ := app.Test(deleteReq("/application/" + appID))
		assertStatus(t, resp, fiber.StatusOK)
	})

	t.Run("invalid application uuid", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/application/:application_uuid",
			newApplicationHandler(&mockApplicationClient{}).DeleteApplication,
		)
		resp, _ := app.Test(deleteReq("/application/not-a-uuid"))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/application/:application_uuid",
			newApplicationHandler(&mockApplicationClient{
				deleteApplication: func(_ context.Context, _ *application_proto.DeleteApplicationRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).DeleteApplication,
		)
		resp, _ := app.Test(deleteReq("/application/" + appID))
		assertStatus(t, resp, fiber.StatusNotFound)
	})

	t.Run("service error - permission denied (not inspector or already assigned)", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/application/:application_uuid",
			newApplicationHandler(&mockApplicationClient{
				deleteApplication: func(_ context.Context, _ *application_proto.DeleteApplicationRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.PermissionDenied)
				},
			}).DeleteApplication,
		)
		resp, _ := app.Test(deleteReq("/application/" + appID))
		assertStatus(t, resp, fiber.StatusForbidden)
	})
}
