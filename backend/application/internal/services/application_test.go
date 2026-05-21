package services

import (
	"context"
	"strings"
	"testing"

	pb "github.com/unwelcome/FrameWorkTask1/backend/contracts/application/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/entities"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/company/generated"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	opID        = "op-test-1"
	initiatorID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	targetID    = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	appID       = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	companyID   = "dddddddd-dddd-dddd-dddd-dddddddddddd"
	deptID      = "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	otherDeptID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	otherUserID = "11111111-1111-1111-1111-111111111111"
)

// ─── CreateApplication ────────────────────────────────────────────────────────

func TestCreateApplication(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := emptyRepo()
		repo.createApplication = func(_ context.Context, _ entities.CreateApplicationDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleClient("inspector"))
		res, err := svc.CreateApplication(context.Background(), &pb.CreateApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			CompanyUuid:     companyID,
			ApplicationData: &pb.ApplicationData{Title: "Valid Title", Description: "Some description"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.GetApplicationUuid() == "" {
			t.Error("expected non-empty application uuid")
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("manager"))
		_, err := svc.CreateApplication(context.Background(), &pb.CreateApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			CompanyUuid:     companyID,
			ApplicationData: &pb.ApplicationData{Title: "Title", Description: "Some description"},
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("empty title", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("inspector"))
		_, err := svc.CreateApplication(context.Background(), &pb.CreateApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			CompanyUuid:     companyID,
			ApplicationData: &pb.ApplicationData{Title: "", Description: "Some description"},
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("title too long", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("inspector"))
		_, err := svc.CreateApplication(context.Background(), &pb.CreateApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			CompanyUuid:     companyID,
			ApplicationData: &pb.ApplicationData{Title: strings.Repeat("x", 256), Description: "Some description"},
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("company service error", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), errCompanyClient())
		_, err := svc.CreateApplication(context.Background(), &pb.CreateApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			CompanyUuid:     companyID,
			ApplicationData: &pb.ApplicationData{Title: "Title", Description: "Some description"},
		})
		if err == nil {
			t.Fatal("expected error from company service, got nil")
		}
	})

	t.Run("db error", func(t *testing.T) {
		repo := emptyRepo()
		repo.createApplication = func(_ context.Context, _ entities.CreateApplicationDTO) Error.CodeError { return internalErr() }

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.CreateApplication(context.Background(), &pb.CreateApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			CompanyUuid:     companyID,
			ApplicationData: &pb.ApplicationData{Title: "Title", Description: "Some description"},
		})
		assertCode(t, err, codes.Internal)
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("inspector"))
		_, err := svc.CreateApplication(context.Background(), &pb.CreateApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   "not-a-uuid",
			CompanyUuid:     companyID,
			ApplicationData: &pb.ApplicationData{Title: "Title", Description: "Desc"},
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_company_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("inspector"))
		_, err := svc.CreateApplication(context.Background(), &pb.CreateApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			CompanyUuid:     "not-a-uuid",
			ApplicationData: &pb.ApplicationData{Title: "Title", Description: "Desc"},
		})
		assertCode(t, err, codes.InvalidArgument)
	})
}

// ─── GetApplication ───────────────────────────────────────────────────────────

func TestGetApplication(t *testing.T) {
	t.Run("success as creator", func(t *testing.T) {
		repo := repoWithApp(testApp())
		repo.getApplicationFixLogs = func(_ context.Context, _ entities.GetApplicationFixLogsDTO) ([]*entities.FixLog, Error.CodeError) {
			return []*entities.FixLog{}, ok()
		}

		svc := newAppTestService(repo, roleClient("inspector"))
		res, err := svc.GetApplication(context.Background(), &pb.GetApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID, // matches CreatedBy
			ApplicationUuid: appID,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.GetApplication().GetApplicationUuid() != appID {
			t.Errorf("expected uuid %q, got %q", appID, res.GetApplication().GetApplicationUuid())
		}
	})

	t.Run("success as chief (unrestricted access)", func(t *testing.T) {
		repo := repoWithApp(testApp())
		repo.getApplicationFixLogs = func(_ context.Context, _ entities.GetApplicationFixLogsDTO) ([]*entities.FixLog, Error.CodeError) {
			return nil, ok()
		}

		svc := newAppTestService(repo, roleClient("chief"))
		_, err := svc.GetApplication(context.Background(), &pb.GetApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   otherUserID,
			ApplicationUuid: appID,
		})
		if err != nil {
			t.Fatalf("chief should access any application, got: %v", err)
		}
	})

	t.Run("permission denied (stranger)", func(t *testing.T) {
		repo := repoWithApp(testApp())

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.GetApplication(context.Background(), &pb.GetApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   otherUserID,
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("not found", func(t *testing.T) {
		repo := emptyRepo()
		repo.getApplication = func(_ context.Context, _ entities.GetApplicationDTO) (*entities.Application, Error.CodeError) {
			return nil, notFound()
		}

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.GetApplication(context.Background(), &pb.GetApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.NotFound)
	})

	t.Run("company service error", func(t *testing.T) {
		repo := repoWithApp(testApp())

		svc := newAppTestService(repo, errCompanyClient())
		_, err := svc.GetApplication(context.Background(), &pb.GetApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("fix logs db error", func(t *testing.T) {
		repo := repoWithApp(testApp())
		repo.getApplicationFixLogs = func(_ context.Context, _ entities.GetApplicationFixLogsDTO) ([]*entities.FixLog, Error.CodeError) {
			return nil, internalErr()
		}

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.GetApplication(context.Background(), &pb.GetApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.Internal)
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("inspector"))
		_, err := svc.GetApplication(context.Background(), &pb.GetApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   "not-a-uuid",
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_application_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("inspector"))
		_, err := svc.GetApplication(context.Background(), &pb.GetApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: "not-a-uuid",
		})
		assertCode(t, err, codes.InvalidArgument)
	})
}

// ─── GetApplications ──────────────────────────────────────────────────────────

func TestGetApplications(t *testing.T) {
	t.Run("invalid count", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("chief"))
		_, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Count:         0,
			Offset:        0,
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid offset", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("chief"))
		_, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Count:         10,
			Offset:        -1,
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("chief success", func(t *testing.T) {
		repo := emptyRepo()
		repo.getApplications = func(_ context.Context, _ entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError) {
			return []*entities.Application{testApp()}, ok()
		}

		svc := newAppTestService(repo, roleClient("chief"))
		res, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Statuses:      []string{"created"},
			Count:         10,
			Offset:        0,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.GetApplications()) != 1 {
			t.Errorf("expected 1 application, got %d", len(res.GetApplications()))
		}
	})

	t.Run("chief invalid status", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("chief"))
		_, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Statuses:      []string{"nonexistent_status"},
			Count:         10,
			Offset:        0,
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("inspector from pool (pending_verification only)", func(t *testing.T) {
		var capturedDTO entities.GetApplicationsDTO
		repo := emptyRepo()
		repo.getApplications = func(_ context.Context, dto entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError) {
			capturedDTO = dto
			return []*entities.Application{}, ok()
		}

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Count:         10,
			Offset:        0,
			FromPool:      true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(capturedDTO.Statuses) != 1 || capturedDTO.Statuses[0] != "pending_verification" {
			t.Errorf("pool inspector should query only pending_verification, got %v", capturedDTO.Statuses)
		}
	})

	t.Run("inspector personal: no statuses → created_by filter", func(t *testing.T) {
		var capturedDTO entities.GetApplicationsDTO
		repo := emptyRepo()
		repo.getApplications = func(_ context.Context, dto entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError) {
			capturedDTO = dto
			return []*entities.Application{}, ok()
		}

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Count:         10,
			Offset:        0,
			Statuses:      []string{},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if capturedDTO.CreatedBy != initiatorID {
			t.Errorf("expected CreatedBy=%q, got %q", initiatorID, capturedDTO.CreatedBy)
		}
	})

	t.Run("inspector personal: on_verification → inspected_by filter", func(t *testing.T) {
		var capturedDTO entities.GetApplicationsDTO
		repo := emptyRepo()
		repo.getApplications = func(_ context.Context, dto entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError) {
			capturedDTO = dto
			return []*entities.Application{}, ok()
		}

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Count:         10,
			Offset:        0,
			Statuses:      []string{"on_verification"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if capturedDTO.InspectedBy != initiatorID {
			t.Errorf("expected InspectedBy=%q, got %q", initiatorID, capturedDTO.InspectedBy)
		}
	})

	t.Run("inspector personal: invalid status", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("inspector"))
		_, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Count:         10,
			Offset:        0,
			Statuses:      []string{"created"}, // only on_verification is valid here
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("manager from pool (executed_by is null)", func(t *testing.T) {
		var capturedDTO entities.GetApplicationsDTO
		repo := emptyRepo()
		repo.getApplications = func(_ context.Context, dto entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError) {
			capturedDTO = dto
			return []*entities.Application{}, ok()
		}

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Count:         10,
			Offset:        0,
			FromPool:      true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !capturedDTO.ExecutedByIsNull {
			t.Error("pool manager query should have ExecutedByIsNull=true")
		}
	})

	t.Run("manager personal (managed_by filter)", func(t *testing.T) {
		var capturedDTO entities.GetApplicationsDTO
		repo := emptyRepo()
		repo.getApplications = func(_ context.Context, dto entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError) {
			capturedDTO = dto
			return []*entities.Application{}, ok()
		}

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Count:         10,
			Offset:        0,
			FromPool:      false,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if capturedDTO.ManagedBy != initiatorID {
			t.Errorf("expected ManagedBy=%q, got %q", initiatorID, capturedDTO.ManagedBy)
		}
	})

	t.Run("engineer success", func(t *testing.T) {
		repo := emptyRepo()
		repo.getApplications = func(_ context.Context, _ entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError) {
			return []*entities.Application{}, ok()
		}

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Count:         10,
			Offset:        0,
			Statuses:      []string{"assigned", "in_progress"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("engineer invalid status", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("engineer"))
		_, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Count:         10,
			Offset:        0,
			Statuses:      []string{"completed"},
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("unemployed denied (company service error)", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), errCompanyClient())
		_, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Count:         10,
			Offset:        0,
		})
		if err == nil {
			t.Fatal("expected error for unknown employee, got nil")
		}
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("chief"))
		_, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: "not-a-uuid",
			CompanyUuid:   companyID,
			Count:         10,
			Offset:        0,
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_company_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("chief"))
		_, err := svc.GetApplications(context.Background(), &pb.GetApplicationsRequest{
			OperationId:   opID,
			InitiatorUuid: initiatorID,
			CompanyUuid:   "not-a-uuid",
			Count:         10,
			Offset:        0,
		})
		assertCode(t, err, codes.InvalidArgument)
	})
}

// ─── UpdateApplicationStatus ──────────────────────────────────────────────────

func TestUpdateApplicationStatus(t *testing.T) {
	t.Run("invalid status (not allowed in this method)", func(t *testing.T) {
		repo := repoWithApp(testApp())

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Status:          "created",
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("inspector: completed from on_verification", func(t *testing.T) {
		repo := repoWithApp(onVerificationApp())
		repo.updateApplicationStatus = func(_ context.Context, _ entities.UpdateApplicationStatusDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Status:          "completed",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("inspector: failed from on_verification", func(t *testing.T) {
		repo := repoWithApp(onVerificationApp())
		repo.updateApplicationStatus = func(_ context.Context, _ entities.UpdateApplicationStatusDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Status:          "failed",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("inspector: on_revision without escalation (revision_count=0)", func(t *testing.T) {
		repo := repoWithApp(onVerificationApp()) // RevisionCount=0, (0+1)%5 != 0
		var capturedDTO entities.UpdateApplicationStatusDTO
		repo.updateApplicationStatus = func(_ context.Context, dto entities.UpdateApplicationStatusDTO) Error.CodeError {
			capturedDTO = dto
			return ok()
		}

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Status:          "on_revision",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if capturedDTO.DropManagedBy || capturedDTO.DropExecutedBy {
			t.Error("should not escalate when (RevisionCount+1)%5 != 0")
		}
	})

	t.Run("inspector: on_revision with escalation (revision_count=4)", func(t *testing.T) {
		app := onVerificationApp()
		app.RevisionCount = 4 // (4+1)%5 == 0 → escalation
		repo := repoWithApp(app)
		var capturedDTO entities.UpdateApplicationStatusDTO
		repo.updateApplicationStatus = func(_ context.Context, dto entities.UpdateApplicationStatusDTO) Error.CodeError {
			capturedDTO = dto
			return ok()
		}

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Status:          "on_revision",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !capturedDTO.DropManagedBy || !capturedDTO.DropExecutedBy {
			t.Error("should escalate when (RevisionCount+1)%5 == 0")
		}
	})

	t.Run("inspector: not responsible (InspectedBy mismatch)", func(t *testing.T) {
		repo := repoWithApp(onVerificationApp())

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   otherUserID,
			ApplicationUuid: appID,
			Status:          "completed",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("inspector: wrong current status (not on_verification)", func(t *testing.T) {
		repo := repoWithApp(assignedApp())

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Status:          "completed",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("inspector: target status not allowed (rejected)", func(t *testing.T) {
		repo := repoWithApp(onVerificationApp())

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Status:          "rejected",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("manager: rejected from created", func(t *testing.T) {
		repo := repoWithApp(testApp()) // status=created, DepartmentUUID=deptID
		repo.updateApplicationStatus = func(_ context.Context, _ entities.UpdateApplicationStatusDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Status:          "rejected",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("manager: wrong department", func(t *testing.T) {
		app := testApp()
		app.DepartmentUUID = otherDeptID
		repo := repoWithApp(app)

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Status:          "rejected",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("manager: target status not allowed (completed)", func(t *testing.T) {
		repo := repoWithApp(testApp())

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Status:          "completed",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("manager: wrong current status (on_verification)", func(t *testing.T) {
		repo := repoWithApp(onVerificationApp())

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Status:          "rejected",
		})
		assertCode(t, err, codes.FailedPrecondition)
	})

	t.Run("engineer: in_progress from assigned", func(t *testing.T) {
		repo := repoWithApp(assignedApp()) // ExecutedBy=targetID
		repo.updateApplicationStatus = func(_ context.Context, _ entities.UpdateApplicationStatusDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   targetID,
			ApplicationUuid: appID,
			Status:          "in_progress",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("engineer: not responsible (ExecutedBy mismatch)", func(t *testing.T) {
		repo := repoWithApp(assignedApp())

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   otherUserID,
			ApplicationUuid: appID,
			Status:          "in_progress",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("engineer: wrong current status (created)", func(t *testing.T) {
		repo := repoWithApp(testApp())

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Status:          "in_progress",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("chief: denied", func(t *testing.T) {
		repo := repoWithApp(testApp())

		svc := newAppTestService(repo, roleClient("chief"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Status:          "rejected",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("manager"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   "not-a-uuid",
			ApplicationUuid: appID,
			Status:          "rejected",
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_application_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("manager"))
		_, err := svc.UpdateApplicationStatus(context.Background(), &pb.UpdateApplicationStatusRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: "not-a-uuid",
			Status:          "rejected",
		})
		assertCode(t, err, codes.InvalidArgument)
	})
}

// ─── AssignApplication ────────────────────────────────────────────────────────

func TestAssignApplication(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := repoWithApp(testApp())
		repo.assignApplicationToEmployee = func(_ context.Context, _ entities.AssignApplicationDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleByTargetClient(map[string]string{
			initiatorID: "manager",
			targetID:    "engineer",
		}))
		_, err := svc.AssignApplication(context.Background(), &pb.AssignApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			TargetUuid:      targetID,
			ApplicationUuid: appID,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("wrong status (on_verification)", func(t *testing.T) {
		repo := repoWithApp(onVerificationApp())

		svc := newAppTestService(repo, roleByTargetClient(map[string]string{
			initiatorID: "manager",
			targetID:    "engineer",
		}))
		_, err := svc.AssignApplication(context.Background(), &pb.AssignApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			TargetUuid:      targetID,
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("on_revision not escalated (revision_count=3)", func(t *testing.T) {
		repo := repoWithApp(onRevisionApp()) // RevisionCount=3, not divisible by 5

		svc := newAppTestService(repo, roleByTargetClient(map[string]string{
			initiatorID: "manager",
			targetID:    "engineer",
		}))
		_, err := svc.AssignApplication(context.Background(), &pb.AssignApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			TargetUuid:      targetID,
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("on_revision escalated (revision_count=5)", func(t *testing.T) {
		repo := repoWithApp(escalatedOnRevisionApp()) // RevisionCount=5
		repo.assignApplicationToEmployee = func(_ context.Context, _ entities.AssignApplicationDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleByTargetClient(map[string]string{
			initiatorID: "manager",
			targetID:    "engineer",
		}))
		_, err := svc.AssignApplication(context.Background(), &pb.AssignApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			TargetUuid:      targetID,
			ApplicationUuid: appID,
		})
		if err != nil {
			t.Fatalf("escalated on_revision should be assignable, got: %v", err)
		}
	})

	t.Run("initiator is not manager", func(t *testing.T) {
		repo := repoWithApp(testApp())

		svc := newAppTestService(repo, roleByTargetClient(map[string]string{
			initiatorID: "engineer",
			targetID:    "engineer",
		}))
		_, err := svc.AssignApplication(context.Background(), &pb.AssignApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			TargetUuid:      targetID,
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("manager department mismatch", func(t *testing.T) {
		app := testApp()
		app.DepartmentUUID = otherDeptID // app in otherDeptID, manager in deptID
		repo := repoWithApp(app)

		svc := newAppTestService(repo, roleByTargetClient(map[string]string{
			initiatorID: "manager",
			targetID:    "engineer",
		}))
		_, err := svc.AssignApplication(context.Background(), &pb.AssignApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			TargetUuid:      targetID,
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("target is not engineer", func(t *testing.T) {
		repo := repoWithApp(testApp())

		svc := newAppTestService(repo, roleByTargetClient(map[string]string{
			initiatorID: "manager",
			targetID:    "manager",
		}))
		_, err := svc.AssignApplication(context.Background(), &pb.AssignApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			TargetUuid:      targetID,
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("db error", func(t *testing.T) {
		repo := repoWithApp(testApp())
		repo.assignApplicationToEmployee = func(_ context.Context, _ entities.AssignApplicationDTO) Error.CodeError { return internalErr() }

		svc := newAppTestService(repo, roleByTargetClient(map[string]string{
			initiatorID: "manager",
			targetID:    "engineer",
		}))
		_, err := svc.AssignApplication(context.Background(), &pb.AssignApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			TargetUuid:      targetID,
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.Internal)
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("manager"))
		_, err := svc.AssignApplication(context.Background(), &pb.AssignApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   "not-a-uuid",
			TargetUuid:      targetID,
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_application_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("manager"))
		_, err := svc.AssignApplication(context.Background(), &pb.AssignApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			TargetUuid:      targetID,
			ApplicationUuid: "not-a-uuid",
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_target_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("manager"))
		_, err := svc.AssignApplication(context.Background(), &pb.AssignApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			TargetUuid:      "not-a-uuid",
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.InvalidArgument)
	})
}

// ─── RedirectApplication ──────────────────────────────────────────────────────

func TestRedirectApplication(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := repoWithApp(testApp())
		repo.addApplicationFixLog = func(_ context.Context, _ entities.AddFixLogDTO) Error.CodeError { return ok() }
		repo.redirectApplication = func(_ context.Context, _ entities.RedirectApplicationDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, redirectClient("manager", companyID))
		_, err := svc.RedirectApplication(context.Background(), &pb.RedirectApplicationRequest{
			OperationId:          opID,
			InitiatorUuid:        initiatorID,
			ApplicationUuid:      appID,
			TargetDepartmentUuid: otherDeptID,
			Message:              "Redirecting to another department",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty message", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), redirectClient("manager", companyID))
		_, err := svc.RedirectApplication(context.Background(), &pb.RedirectApplicationRequest{
			OperationId:          opID,
			InitiatorUuid:        initiatorID,
			ApplicationUuid:      appID,
			TargetDepartmentUuid: otherDeptID,
			Message:              "   ",
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid status (on_verification)", func(t *testing.T) {
		repo := repoWithApp(onVerificationApp())

		svc := newAppTestService(repo, redirectClient("manager", companyID))
		_, err := svc.RedirectApplication(context.Background(), &pb.RedirectApplicationRequest{
			OperationId:          opID,
			InitiatorUuid:        initiatorID,
			ApplicationUuid:      appID,
			TargetDepartmentUuid: otherDeptID,
			Message:              "message",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("initiator is not manager", func(t *testing.T) {
		repo := repoWithApp(testApp())

		svc := newAppTestService(repo, redirectClient("engineer", companyID))
		_, err := svc.RedirectApplication(context.Background(), &pb.RedirectApplicationRequest{
			OperationId:          opID,
			InitiatorUuid:        initiatorID,
			ApplicationUuid:      appID,
			TargetDepartmentUuid: otherDeptID,
			Message:              "message",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("manager department mismatch", func(t *testing.T) {
		app := testApp()
		app.DepartmentUUID = otherDeptID // manager is in deptID, app in otherDeptID
		repo := repoWithApp(app)

		svc := newAppTestService(repo, redirectClient("manager", companyID))
		_, err := svc.RedirectApplication(context.Background(), &pb.RedirectApplicationRequest{
			OperationId:          opID,
			InitiatorUuid:        initiatorID,
			ApplicationUuid:      appID,
			TargetDepartmentUuid: deptID,
			Message:              "message",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("target department belongs to another company", func(t *testing.T) {
		repo := repoWithApp(testApp())

		svc := newAppTestService(repo, redirectClient("manager", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"))
		_, err := svc.RedirectApplication(context.Background(), &pb.RedirectApplicationRequest{
			OperationId:          opID,
			InitiatorUuid:        initiatorID,
			ApplicationUuid:      appID,
			TargetDepartmentUuid: otherDeptID,
			Message:              "message",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("db error on fix log", func(t *testing.T) {
		repo := repoWithApp(testApp())
		repo.addApplicationFixLog = func(_ context.Context, _ entities.AddFixLogDTO) Error.CodeError { return internalErr() }

		svc := newAppTestService(repo, redirectClient("manager", companyID))
		_, err := svc.RedirectApplication(context.Background(), &pb.RedirectApplicationRequest{
			OperationId:          opID,
			InitiatorUuid:        initiatorID,
			ApplicationUuid:      appID,
			TargetDepartmentUuid: otherDeptID,
			Message:              "message",
		})
		assertCode(t, err, codes.Internal)
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), redirectClient("manager", companyID))
		_, err := svc.RedirectApplication(context.Background(), &pb.RedirectApplicationRequest{
			OperationId:          opID,
			InitiatorUuid:        "not-a-uuid",
			ApplicationUuid:      appID,
			TargetDepartmentUuid: otherDeptID,
			Message:              "message",
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_application_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), redirectClient("manager", companyID))
		_, err := svc.RedirectApplication(context.Background(), &pb.RedirectApplicationRequest{
			OperationId:          opID,
			InitiatorUuid:        initiatorID,
			ApplicationUuid:      "not-a-uuid",
			TargetDepartmentUuid: otherDeptID,
			Message:              "message",
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_target_department_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), redirectClient("manager", companyID))
		_, err := svc.RedirectApplication(context.Background(), &pb.RedirectApplicationRequest{
			OperationId:          opID,
			InitiatorUuid:        initiatorID,
			ApplicationUuid:      appID,
			TargetDepartmentUuid: "not-a-uuid",
			Message:              "message",
		})
		assertCode(t, err, codes.InvalidArgument)
	})
}

// ─── RecallApplication ────────────────────────────────────────────────────────

func TestRecallApplication(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := repoWithApp(assignedApp()) // ManagedBy=initiatorID
		repo.addApplicationFixLog = func(_ context.Context, _ entities.AddFixLogDTO) Error.CodeError { return ok() }
		repo.recallApplication = func(_ context.Context, _ entities.RecallApplicationDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.RecallApplication(context.Background(), &pb.RecallApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "Recalling for rework",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty message", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("manager"))
		_, err := svc.RecallApplication(context.Background(), &pb.RecallApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "",
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid status (created)", func(t *testing.T) {
		repo := repoWithApp(testApp())

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.RecallApplication(context.Background(), &pb.RecallApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "message",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("not responsible manager (ManagedBy mismatch)", func(t *testing.T) {
		repo := repoWithApp(assignedApp()) // ManagedBy=initiatorID

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.RecallApplication(context.Background(), &pb.RecallApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   otherUserID,
			ApplicationUuid: appID,
			Message:         "message",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("manager role changed", func(t *testing.T) {
		repo := repoWithApp(assignedApp()) // ManagedBy=initiatorID

		svc := newAppTestService(repo, roleClient("engineer")) // role changed
		_, err := svc.RecallApplication(context.Background(), &pb.RecallApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "message",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("db error", func(t *testing.T) {
		repo := repoWithApp(assignedApp())
		repo.addApplicationFixLog = func(_ context.Context, _ entities.AddFixLogDTO) Error.CodeError { return ok() }
		repo.recallApplication = func(_ context.Context, _ entities.RecallApplicationDTO) Error.CodeError { return internalErr() }

		svc := newAppTestService(repo, roleClient("manager"))
		_, err := svc.RecallApplication(context.Background(), &pb.RecallApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "message",
		})
		assertCode(t, err, codes.Internal)
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("manager"))
		_, err := svc.RecallApplication(context.Background(), &pb.RecallApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   "not-a-uuid",
			ApplicationUuid: appID,
			Message:         "message",
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_application_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("manager"))
		_, err := svc.RecallApplication(context.Background(), &pb.RecallApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: "not-a-uuid",
			Message:         "message",
		})
		assertCode(t, err, codes.InvalidArgument)
	})
}

// ─── TakeApplicationToVerification ───────────────────────────────────────────

func TestTakeApplicationToVerification(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := repoWithApp(pendingVerificationApp()) // DepartmentUUID=deptID
		repo.takeApplicationToVerification = func(_ context.Context, _ entities.TakeApplicationToVerificationDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.TakeApplicationToVerification(context.Background(), &pb.TakeApplicationToVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("wrong status (not pending_verification)", func(t *testing.T) {
		repo := repoWithApp(testApp())

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.TakeApplicationToVerification(context.Background(), &pb.TakeApplicationToVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("initiator is not inspector", func(t *testing.T) {
		repo := repoWithApp(pendingVerificationApp())

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.TakeApplicationToVerification(context.Background(), &pb.TakeApplicationToVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("department mismatch", func(t *testing.T) {
		app := pendingVerificationApp()
		app.DepartmentUUID = otherDeptID // inspector in deptID, app in otherDeptID
		repo := repoWithApp(app)

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.TakeApplicationToVerification(context.Background(), &pb.TakeApplicationToVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("db error", func(t *testing.T) {
		repo := repoWithApp(pendingVerificationApp())
		repo.takeApplicationToVerification = func(_ context.Context, _ entities.TakeApplicationToVerificationDTO) Error.CodeError {
			return internalErr()
		}

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.TakeApplicationToVerification(context.Background(), &pb.TakeApplicationToVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.Internal)
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("inspector"))
		_, err := svc.TakeApplicationToVerification(context.Background(), &pb.TakeApplicationToVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   "not-a-uuid",
			ApplicationUuid: appID,
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_application_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("inspector"))
		_, err := svc.TakeApplicationToVerification(context.Background(), &pb.TakeApplicationToVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: "not-a-uuid",
		})
		assertCode(t, err, codes.InvalidArgument)
	})
}

// ─── ReleaseApplicationVerification ──────────────────────────────────────────

func TestReleaseApplicationVerification(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := repoWithApp(onVerificationApp()) // InspectedBy=initiatorID
		repo.addApplicationFixLog = func(_ context.Context, _ entities.AddFixLogDTO) Error.CodeError { return ok() }
		repo.releaseApplicationVerification = func(_ context.Context, _ entities.ReleaseApplicationVerificationDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.ReleaseApplicationVerification(context.Background(), &pb.ReleaseApplicationVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "Releasing verification",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty message", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("inspector"))
		_, err := svc.ReleaseApplicationVerification(context.Background(), &pb.ReleaseApplicationVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "  ",
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("wrong status (not on_verification)", func(t *testing.T) {
		repo := repoWithApp(pendingVerificationApp())

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.ReleaseApplicationVerification(context.Background(), &pb.ReleaseApplicationVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "message",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("not responsible inspector (InspectedBy mismatch)", func(t *testing.T) {
		repo := repoWithApp(onVerificationApp()) // InspectedBy=initiatorID

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.ReleaseApplicationVerification(context.Background(), &pb.ReleaseApplicationVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   otherUserID,
			ApplicationUuid: appID,
			Message:         "message",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("inspector role changed", func(t *testing.T) {
		repo := repoWithApp(onVerificationApp())

		svc := newAppTestService(repo, roleClient("manager")) // role changed
		_, err := svc.ReleaseApplicationVerification(context.Background(), &pb.ReleaseApplicationVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "message",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("db error", func(t *testing.T) {
		repo := repoWithApp(onVerificationApp())
		repo.addApplicationFixLog = func(_ context.Context, _ entities.AddFixLogDTO) Error.CodeError { return ok() }
		repo.releaseApplicationVerification = func(_ context.Context, _ entities.ReleaseApplicationVerificationDTO) Error.CodeError {
			return internalErr()
		}

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.ReleaseApplicationVerification(context.Background(), &pb.ReleaseApplicationVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "message",
		})
		assertCode(t, err, codes.Internal)
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("inspector"))
		_, err := svc.ReleaseApplicationVerification(context.Background(), &pb.ReleaseApplicationVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   "not-a-uuid",
			ApplicationUuid: appID,
			Message:         "message",
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_application_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("inspector"))
		_, err := svc.ReleaseApplicationVerification(context.Background(), &pb.ReleaseApplicationVerificationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: "not-a-uuid",
			Message:         "message",
		})
		assertCode(t, err, codes.InvalidArgument)
	})
}

// ─── AddApplicationFixLog ─────────────────────────────────────────────────────

func TestAddApplicationFixLog(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := repoWithApp(inProgressApp()) // ExecutedBy=initiatorID
		repo.addApplicationFixLog = func(_ context.Context, _ entities.AddFixLogDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.AddApplicationFixLog(context.Background(), &pb.AddApplicationFixLogRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "Fixed the issue",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty message", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("engineer"))
		_, err := svc.AddApplicationFixLog(context.Background(), &pb.AddApplicationFixLogRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "   ",
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("application not found", func(t *testing.T) {
		repo := emptyRepo()
		repo.getApplication = func(_ context.Context, _ entities.GetApplicationDTO) (*entities.Application, Error.CodeError) {
			return nil, notFound()
		}

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.AddApplicationFixLog(context.Background(), &pb.AddApplicationFixLogRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "message",
		})
		assertCode(t, err, codes.NotFound)
	})

	t.Run("not responsible engineer (ExecutedBy mismatch)", func(t *testing.T) {
		repo := repoWithApp(inProgressApp()) // ExecutedBy=initiatorID

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.AddApplicationFixLog(context.Background(), &pb.AddApplicationFixLogRequest{
			OperationId:     opID,
			InitiatorUuid:   otherUserID,
			ApplicationUuid: appID,
			Message:         "message",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("db error", func(t *testing.T) {
		repo := repoWithApp(inProgressApp())
		repo.addApplicationFixLog = func(_ context.Context, _ entities.AddFixLogDTO) Error.CodeError { return internalErr() }

		svc := newAppTestService(repo, roleClient("engineer"))
		_, err := svc.AddApplicationFixLog(context.Background(), &pb.AddApplicationFixLogRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "message",
		})
		assertCode(t, err, codes.Internal)
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("engineer"))
		_, err := svc.AddApplicationFixLog(context.Background(), &pb.AddApplicationFixLogRequest{
			OperationId:     opID,
			InitiatorUuid:   "not-a-uuid",
			ApplicationUuid: appID,
			Message:         "message",
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_application_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("engineer"))
		_, err := svc.AddApplicationFixLog(context.Background(), &pb.AddApplicationFixLogRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: "not-a-uuid",
			Message:         "message",
		})
		assertCode(t, err, codes.InvalidArgument)
	})
}

// ─── DeleteApplication ────────────────────────────────────────────────────────

func TestDeleteApplication(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := repoWithApp(testApp()) // CreatedBy=initiatorID, status=created
		repo.addApplicationFixLog = func(_ context.Context, _ entities.AddFixLogDTO) Error.CodeError { return ok() }
		repo.deleteApplication = func(_ context.Context, _ entities.DeleteApplicationDTO) Error.CodeError { return ok() }

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.DeleteApplication(context.Background(), &pb.DeleteApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "Deleting application",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("not creator", func(t *testing.T) {
		repo := repoWithApp(testApp())

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.DeleteApplication(context.Background(), &pb.DeleteApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   otherUserID,
			ApplicationUuid: appID,
			Message:         "reason",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("status is not created (assigned)", func(t *testing.T) {
		repo := repoWithApp(assignedApp())

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.DeleteApplication(context.Background(), &pb.DeleteApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "reason",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("creator is no longer inspector", func(t *testing.T) {
		repo := repoWithApp(testApp())

		svc := newAppTestService(repo, roleClient("manager")) // role changed
		_, err := svc.DeleteApplication(context.Background(), &pb.DeleteApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "reason",
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("db error", func(t *testing.T) {
		repo := repoWithApp(testApp())
		repo.addApplicationFixLog = func(_ context.Context, _ entities.AddFixLogDTO) Error.CodeError { return ok() }
		repo.deleteApplication = func(_ context.Context, _ entities.DeleteApplicationDTO) Error.CodeError { return internalErr() }

		svc := newAppTestService(repo, roleClient("inspector"))
		_, err := svc.DeleteApplication(context.Background(), &pb.DeleteApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: appID,
			Message:         "reason",
		})
		assertCode(t, err, codes.Internal)
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("inspector"))
		_, err := svc.DeleteApplication(context.Background(), &pb.DeleteApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   "not-a-uuid",
			ApplicationUuid: appID,
			Message:         "reason",
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_application_uuid", func(t *testing.T) {
		svc := newAppTestService(emptyRepo(), roleClient("inspector"))
		_, err := svc.DeleteApplication(context.Background(), &pb.DeleteApplicationRequest{
			OperationId:     opID,
			InitiatorUuid:   initiatorID,
			ApplicationUuid: "not-a-uuid",
			Message:         "reason",
		})
		assertCode(t, err, codes.InvalidArgument)
	})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func assertCode(t *testing.T, err error, expected codes.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected gRPC error with code %v, got nil", expected)
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != expected {
		t.Errorf("expected code %v, got %v: %s", expected, st.Code(), st.Message())
	}
}

// Compile-time check: mockCompanyClient satisfies the interface
var _ company_proto.CompanyServiceClient = (*mockCompanyClient)(nil)

// Suppress unused import: grpc is used in mock function signatures in mocks_test.go
var _ grpc.CallOption
