package services

import (
	"context"
	"testing"

	pb "github.com/unwelcome/FrameWorkTask1/backend/application/api/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─── Тестовые константы ───────────────────────────────────────────────────────

const (
	opID        = "op-test-1"
	companyID   = "company-uuid-1"
	initiatorID = "user-uuid-1"
	targetID    = "user-uuid-2"
	appID       = "app-uuid-1"
)

// ─── Утилиты утверждений ──────────────────────────────────────────────────────

func assertGRPCCode(t *testing.T, err error, expected codes.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %v, got nil", expected)
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status error: %v", err)
	}
	if st.Code() != expected {
		t.Errorf("expected gRPC code %v, got %v (message: %q)", expected, st.Code(), st.Message())
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// ─── Health ───────────────────────────────────────────────────────────────────

func TestHealth(t *testing.T) {
	svc := newAppTestService(emptyRepo(), roleClient("inspector"))
	res, err := svc.Health(context.Background(), &pb.HealthRequest{OperationId: opID})
	assertNoError(t, err)
	if res.GetHealth() != "healthy" {
		t.Errorf("expected %q, got %q", "healthy", res.GetHealth())
	}
}

// ─── CreateApplication ────────────────────────────────────────────────────────

func TestCreateApplication(t *testing.T) {
	ctx := context.Background()
	req := &pb.CreateApplicationRequest{
		OperationId:   opID,
		InitiatorUuid: initiatorID,
		CompanyUuid:   companyID,
		Title:         "Fix pipeline",
		Description:   "Pipeline is broken",
	}

	t.Run("success", func(t *testing.T) {
		repo := emptyRepo()
		repo.createApplication = func(_ context.Context, _ entities.CreateApplicationDTO) Error.CodeError {
			return ok()
		}

		svc := newAppTestService(repo, roleClient("inspector"))
		res, err := svc.CreateApplication(ctx, req)
		assertNoError(t, err)
		if res.GetApplicationUuid() == "" {
			t.Error("expected non-empty application uuid in response")
		}
	})

	t.Run("company service error", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), errCompanyClient())
		_, err := svc.CreateApplication(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("wrong role - only inspector can create", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("engineer"))
		_, err := svc.CreateApplication(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("db create error", func(t *testing.T) {
		repo := emptyRepo()
		repo.createApplication = func(_ context.Context, _ entities.CreateApplicationDTO) Error.CodeError {
			return internalErr()
		}

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.CreateApplication(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── GetApplication ───────────────────────────────────────────────────────────

func TestGetApplication(t *testing.T) {
	ctx := context.Background()
	req := &pb.GetApplicationRequest{
		OperationId:     opID,
		InitiatorUuid:   initiatorID,
		ApplicationUuid: appID,
	}

	t.Run("success", func(t *testing.T) {
		repo := repoWithApp(testApp())
		repo.getApplicationFixLogs = func(_ context.Context, _ entities.GetApplicationFixLogsDTO) ([]*entities.FixLog, Error.CodeError) {
			return []*entities.FixLog{
				{Text: "Fixed pipe A", CreatedAt: "2024-01-02", CreatedBy: targetID},
			}, ok()
		}

		svc := newAppTestService(repo, roleClient("engineer"))
		res, err := svc.GetApplication(ctx, req)
		assertNoError(t, err)
		if res.GetApplication().GetTitle() != "Test Application" {
			t.Errorf("unexpected title: %q", res.GetApplication().GetTitle())
		}
		if len(res.GetApplication().GetFixLogs()) != 1 {
			t.Errorf("expected 1 fix log, got %d", len(res.GetApplication().GetFixLogs()))
		}
	})

	t.Run("application not found", func(t *testing.T) {
		repo := emptyRepo()
		repo.getApplication = func(_ context.Context, _ entities.GetApplicationDTO) (*entities.Application, Error.CodeError) {
			return nil, notFound()
		}

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.GetApplication(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("company service error", func(t *testing.T) {
		svc := newAppTestService(repoWithApp(testApp()), errCompanyClient())
		_, err := svc.GetApplication(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("unemployed not allowed", func(t *testing.T) {
		svc := newAppTestService(repoWithApp(testApp()), roleClient("unemployed"))
		_, err := svc.GetApplication(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("get fix logs db error", func(t *testing.T) {
		repo := repoWithApp(testApp())
		repo.getApplicationFixLogs = func(_ context.Context, _ entities.GetApplicationFixLogsDTO) ([]*entities.FixLog, Error.CodeError) {
			return nil, internalErr()
		}

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.GetApplication(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── GetApplications ──────────────────────────────────────────────────────────

func TestGetApplications(t *testing.T) {
	ctx := context.Background()

	makeReq := func(count, offset int64) *pb.GetApplicationsRequest {
		return &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Count:         count,
			Offset:        offset,
		}
	}

	t.Run("success", func(t *testing.T) {
		repo := emptyRepo()
		repo.getApplications = func(_ context.Context, _ entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError) {
			return []*entities.Application{testApp(), testApp()}, ok()
		}

		svc := newAppTestService(repo, roleClient("manager"))
		res, err := svc.GetApplications(ctx, makeReq(10, 0))
		assertNoError(t, err)
		if len(res.GetApplications()) != 2 {
			t.Errorf("expected 2 applications, got %d", len(res.GetApplications()))
		}
	})

	t.Run("company service error", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), errCompanyClient())
		_, err := svc.GetApplications(ctx, makeReq(10, 0))
		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("unemployed not allowed", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("unemployed"))
		_, err := svc.GetApplications(ctx, makeReq(10, 0))
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("count zero", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("engineer"))
		_, err := svc.GetApplications(ctx, makeReq(0, 0))
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("count too large", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("engineer"))
		_, err := svc.GetApplications(ctx, makeReq(101, 0))
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("negative offset", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("engineer"))
		_, err := svc.GetApplications(ctx, makeReq(10, -1))
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("db error", func(t *testing.T) {
		repo := emptyRepo()
		repo.getApplications = func(_ context.Context, _ entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError) {
			return nil, internalErr()
		}

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.GetApplications(ctx, makeReq(10, 0))
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── GetCompanyApplicationStatistic ──────────────────────────────────────────

func TestGetCompanyApplicationStatistic(t *testing.T) {
	ctx := context.Background()
	req := &pb.GetCompanyApplicationStatisticRequest{
		OperationId:   opID,
		InitiatorUuid: initiatorID,
		CompanyUuid:   companyID,
	}

	t.Run("success as analytic", func(t *testing.T) {
		repo := emptyRepo()
		repo.getCompanyApplicationStatistic = func(_ context.Context, _ entities.GetCompanyApplicationStatisticDTO) (*entities.ApplicationStatistic, Error.CodeError) {
			return testStatistic(), ok()
		}

		svc := newAppTestService(repo, roleClient("analytic"))
		res, err := svc.GetCompanyApplicationStatistic(ctx, req)
		assertNoError(t, err)
		if res.GetCreated() != 2 {
			t.Errorf("expected created=2, got %d", res.GetCreated())
		}
	})

	t.Run("success as chief", func(t *testing.T) {
		repo := emptyRepo()
		repo.getCompanyApplicationStatistic = func(_ context.Context, _ entities.GetCompanyApplicationStatisticDTO) (*entities.ApplicationStatistic, Error.CodeError) {
			return testStatistic(), ok()
		}

		svc := newAppTestService(repo, roleClient("chief"))
		_, err := svc.GetCompanyApplicationStatistic(ctx, req)
		assertNoError(t, err)
	})

	t.Run("company service error", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), errCompanyClient())
		_, err := svc.GetCompanyApplicationStatistic(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("engineer not allowed", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("engineer"))
		_, err := svc.GetCompanyApplicationStatistic(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("manager not allowed", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("manager"))
		_, err := svc.GetCompanyApplicationStatistic(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("db error", func(t *testing.T) {
		repo := emptyRepo()
		repo.getCompanyApplicationStatistic = func(_ context.Context, _ entities.GetCompanyApplicationStatisticDTO) (*entities.ApplicationStatistic, Error.CodeError) {
			return nil, internalErr()
		}

		svc := newAppTestService(repo, roleClient("analytic"))
		_, err := svc.GetCompanyApplicationStatistic(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── GetEmployeeApplicationStatistic ─────────────────────────────────────────

func TestGetEmployeeApplicationStatistic(t *testing.T) {
	ctx := context.Background()
	req := &pb.GetEmployeeApplicationStatisticRequest{
		OperationId:   opID,
		InitiatorUuid: initiatorID,
		CompanyUuid:   companyID,
		TargetUuid:    targetID,
	}

	t.Run("success", func(t *testing.T) {
		repo := emptyRepo()
		repo.getEmployeeApplicationStatistic = func(_ context.Context, _ entities.GetEmployeeApplicationStatisticDTO) (*entities.ApplicationStatistic, Error.CodeError) {
			return testStatistic(), ok()
		}

		// Первый вызов (self-check initiator): manager; второй вызов (target): engineer
		svc := newAppTestService(repo, roleByTargetClient(map[string]string{
			initiatorID: "manager",
			targetID:    "engineer",
		}))
		res, err := svc.GetEmployeeApplicationStatistic(ctx, req)
		assertNoError(t, err)
		if res.GetInProgress() != 3 {
			t.Errorf("expected in_progress=3, got %d", res.GetInProgress())
		}
	})

	t.Run("company service error on initiator check", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), errCompanyClient())
		_, err := svc.GetEmployeeApplicationStatistic(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("unemployed initiator not allowed", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleByTargetClient(map[string]string{
			initiatorID: "unemployed",
		}))
		_, err := svc.GetEmployeeApplicationStatistic(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("target not in company", func(t *testing.T) {
		// targetID отсутствует в карте → roleByTargetClient вернёт NotFound
		svc := newAppTestService(emptyRepo(), roleByTargetClient(map[string]string{
			initiatorID: "manager",
		}))
		_, err := svc.GetEmployeeApplicationStatistic(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("db error", func(t *testing.T) {
		repo := emptyRepo()
		repo.getEmployeeApplicationStatistic = func(_ context.Context, _ entities.GetEmployeeApplicationStatisticDTO) (*entities.ApplicationStatistic, Error.CodeError) {
			return nil, internalErr()
		}

		svc := newAppTestService(repo, roleByTargetClient(map[string]string{
			initiatorID: "manager",
			targetID:    "engineer",
		}))
		_, err := svc.GetEmployeeApplicationStatistic(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── UpdateApplicationStatus ──────────────────────────────────────────────────

func TestUpdateApplicationStatus(t *testing.T) {
	ctx := context.Background()

	makeReq := func(newStatus string) *pb.UpdateApplicationStatusRequest {
		return &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Status:          newStatus,
		}
	}

	t.Run("success engineer sets in_progress", func(t *testing.T) {
		app := assignedApp()
		app.ResponsibleEngineer = initiatorID

		repo := repoWithApp(app)
		repo.updateApplicationStatus = func(_ context.Context, _ entities.UpdateApplicationStatusDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.UpdateApplicationStatus(ctx, makeReq("in_progress"))
		assertNoError(t, err)
	})

	t.Run("success manager sets completed", func(t *testing.T) {
		app := awaitingApprovalApp()
		app.ResponsibleManager = initiatorID

		repo := repoWithApp(app)
		repo.updateApplicationStatus = func(_ context.Context, _ entities.UpdateApplicationStatusDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.UpdateApplicationStatus(ctx, makeReq("completed"))
		assertNoError(t, err)
	})

	t.Run("invalid status", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("engineer"))
		_, err := svc.UpdateApplicationStatus(ctx, makeReq("invalid_status"))
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("application not found", func(t *testing.T) {
		repo := emptyRepo()
		repo.getApplication = func(_ context.Context, _ entities.GetApplicationDTO) (*entities.Application, Error.CodeError) {
			return nil, notFound()
		}

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.UpdateApplicationStatus(ctx, makeReq("in_progress"))
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("forbidden role - inspector cannot change status", func(t *testing.T) {
		svc := newAppTestService(repoWithApp(assignedApp()), roleClient("inspector"))
		_, err := svc.UpdateApplicationStatus(ctx, makeReq("in_progress"))
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("forbidden role - chief cannot change status", func(t *testing.T) {
		svc := newAppTestService(repoWithApp(assignedApp()), roleClient("chief"))
		_, err := svc.UpdateApplicationStatus(ctx, makeReq("in_progress"))
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("engineer is not the responsible engineer", func(t *testing.T) {
		app := assignedApp()
		app.ResponsibleEngineer = targetID // инициатор != ответственный

		svc := newAppTestService(repoWithApp(app), roleClient("engineer"))
		_, err := svc.UpdateApplicationStatus(ctx, makeReq("in_progress"))
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("engineer cannot transition from created status", func(t *testing.T) {
		app := testApp() // status=created — не в validEngineerCurrentStatuses
		app.ResponsibleEngineer = initiatorID

		svc := newAppTestService(repoWithApp(app), roleClient("engineer"))
		_, err := svc.UpdateApplicationStatus(ctx, makeReq("in_progress"))
		assertGRPCCode(t, err, codes.FailedPrecondition)
	})

	t.Run("engineer cannot set manager statuses", func(t *testing.T) {
		app := awaitingApprovalApp()
		app.ResponsibleEngineer = initiatorID

		svc := newAppTestService(repoWithApp(app), roleClient("engineer"))
		_, err := svc.UpdateApplicationStatus(ctx, makeReq("completed"))
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("manager is not the responsible manager", func(t *testing.T) {
		app := awaitingApprovalApp()
		app.ResponsibleManager = targetID // инициатор != ответственный

		svc := newAppTestService(repoWithApp(app), roleClient("manager"))
		_, err := svc.UpdateApplicationStatus(ctx, makeReq("completed"))
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("manager cannot transition from assigned status", func(t *testing.T) {
		app := assignedApp() // status=assigned — не в validManagerCurrentStatuses
		app.ResponsibleManager = initiatorID

		svc := newAppTestService(repoWithApp(app), roleClient("manager"))
		_, err := svc.UpdateApplicationStatus(ctx, makeReq("completed"))
		assertGRPCCode(t, err, codes.FailedPrecondition)
	})

	t.Run("manager cannot set engineer statuses", func(t *testing.T) {
		app := awaitingApprovalApp()
		app.ResponsibleManager = initiatorID

		svc := newAppTestService(repoWithApp(app), roleClient("manager"))
		_, err := svc.UpdateApplicationStatus(ctx, makeReq("in_progress"))
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("db update error", func(t *testing.T) {
		app := assignedApp()
		app.ResponsibleEngineer = initiatorID

		repo := repoWithApp(app)
		repo.updateApplicationStatus = func(_ context.Context, _ entities.UpdateApplicationStatusDTO) Error.CodeError { return internalErr() }

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.UpdateApplicationStatus(ctx, makeReq("in_progress"))
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── AssignApplicationToEmployee ─────────────────────────────────────────────

func TestAssignApplicationToEmployee(t *testing.T) {
	ctx := context.Background()
	req := &pb.AssignApplicationToEmployeeRequest{
		OperationId:     opID,
		InitiatorUuid:   initiatorID,
		TargetUuid:      targetID,
		ApplicationUuid: appID,
	}

	t.Run("success", func(t *testing.T) {
		repo := repoWithApp(testApp())
		repo.assignApplicationToEmployee = func(_ context.Context, _ entities.AssignApplicationToEmployeeDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleByTargetClient(map[string]string{
			initiatorID: "manager",
			targetID:    "engineer",
		}))
		_, err := svc.AssignApplicationToEmployee(ctx, req)
		assertNoError(t, err)
	})

	t.Run("application not found", func(t *testing.T) {
		repo := emptyRepo()
		repo.getApplication = func(_ context.Context, _ entities.GetApplicationDTO) (*entities.Application, Error.CodeError) {
			return nil, notFound()
		}

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.AssignApplicationToEmployee(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("company service error on initiator check", func(t *testing.T) {
		svc := newAppTestService(repoWithApp(testApp()), errCompanyClient())
		_, err := svc.AssignApplicationToEmployee(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("initiator is not manager", func(t *testing.T) {
		svc := newAppTestService(repoWithApp(testApp()), roleByTargetClient(map[string]string{
			initiatorID: "engineer",
		}))
		_, err := svc.AssignApplicationToEmployee(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("target is not engineer", func(t *testing.T) {
		svc := newAppTestService(repoWithApp(testApp()), roleByTargetClient(map[string]string{
			initiatorID: "manager",
			targetID:    "manager",
		}))
		_, err := svc.AssignApplicationToEmployee(ctx, req)
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("target not in company", func(t *testing.T) {
		// targetID отсутствует в карте → NotFound
		svc := newAppTestService(repoWithApp(testApp()), roleByTargetClient(map[string]string{
			initiatorID: "manager",
		}))
		_, err := svc.AssignApplicationToEmployee(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("db assign error", func(t *testing.T) {
		repo := repoWithApp(testApp())
		repo.assignApplicationToEmployee = func(_ context.Context, _ entities.AssignApplicationToEmployeeDTO) Error.CodeError { return internalErr() }

		svc := newAppTestService(repo, roleByTargetClient(map[string]string{
			initiatorID: "manager",
			targetID:    "engineer",
		}))
		_, err := svc.AssignApplicationToEmployee(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── AddApplicationFixLog ─────────────────────────────────────────────────────

func TestAddApplicationFixLog(t *testing.T) {
	ctx := context.Background()
	req := &pb.AddApplicationFixLogRequest{
		OperationId:     opID,
		InitiatorUuid:   initiatorID,
		ApplicationUuid: appID,
		LogText:         "Replaced filter unit B7",
	}

	t.Run("success", func(t *testing.T) {
		app := assignedApp()
		app.ResponsibleEngineer = initiatorID

		repo := repoWithApp(app)
		repo.addApplicationFixLog = func(_ context.Context, _ entities.CreateFixLogDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.AddApplicationFixLog(ctx, req)
		assertNoError(t, err)
	})

	t.Run("application not found", func(t *testing.T) {
		repo := emptyRepo()
		repo.getApplication = func(_ context.Context, _ entities.GetApplicationDTO) (*entities.Application, Error.CodeError) {
			return nil, notFound()
		}

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.AddApplicationFixLog(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("initiator is not the responsible engineer", func(t *testing.T) {
		app := assignedApp()
		app.ResponsibleEngineer = targetID // initiatorID != targetID

		svc := newAppTestService(repoWithApp(app), roleClient("engineer"))
		_, err := svc.AddApplicationFixLog(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("db add error", func(t *testing.T) {
		app := assignedApp()
		app.ResponsibleEngineer = initiatorID

		repo := repoWithApp(app)
		repo.addApplicationFixLog = func(_ context.Context, _ entities.CreateFixLogDTO) Error.CodeError { return internalErr() }

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.AddApplicationFixLog(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── DeleteApplication ────────────────────────────────────────────────────────

func TestDeleteApplication(t *testing.T) {
	ctx := context.Background()
	req := &pb.DeleteApplicationRequest{
		OperationId:     opID,
		InitiatorUuid:   initiatorID,
		ApplicationUuid: appID,
	}

	t.Run("success", func(t *testing.T) {
		repo := repoWithApp(testApp()) // status=created
		repo.deleteApplicationRequest = func(_ context.Context, _ entities.DeleteApplicationDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.DeleteApplication(ctx, req)
		assertNoError(t, err)
	})

	t.Run("application not found", func(t *testing.T) {
		repo := emptyRepo()
		repo.getApplication = func(_ context.Context, _ entities.GetApplicationDTO) (*entities.Application, Error.CodeError) {
			return nil, notFound()
		}

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.DeleteApplication(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("application already assigned - cannot delete", func(t *testing.T) {
		svc := newAppTestService(repoWithApp(assignedApp()), roleClient("inspector"))
		_, err := svc.DeleteApplication(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("application in progress - cannot delete", func(t *testing.T) {
		app := testApp()
		app.Status = "in_progress"

		svc := newAppTestService(repoWithApp(app), roleClient("inspector"))
		_, err := svc.DeleteApplication(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("company service error", func(t *testing.T) {
		svc := newAppTestService(repoWithApp(testApp()), errCompanyClient())
		_, err := svc.DeleteApplication(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("initiator is not inspector", func(t *testing.T) {
		svc := newAppTestService(repoWithApp(testApp()), roleClient("manager"))
		_, err := svc.DeleteApplication(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("db delete error", func(t *testing.T) {
		repo := repoWithApp(testApp())
		repo.deleteApplicationRequest = func(_ context.Context, _ entities.DeleteApplicationDTO) Error.CodeError { return internalErr() }

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.DeleteApplication(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}
