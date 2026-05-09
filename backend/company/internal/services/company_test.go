package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb "github.com/unwelcome/FrameWorkTask1/backend/company/api/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/company/internal/entities"
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
	testCode    = "123456"
	deptID      = "dept-uuid-1"
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
	// Health вызывает db.Ping, которому нужно реальное соединение — пропускаем в unit-тестах
	t.Skip("requires real DB/Redis connection")
}

// ─── CreateCompany ────────────────────────────────────────────────────────────

func TestCreateCompany(t *testing.T) {
	ctx := context.Background()
	req := &pb.CreateCompanyRequest{OperationId: opID, Title: "ACME", InitiatorUuid: initiatorID}

	t.Run("success", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.createCompany = func(_ context.Context, _ *entities.CreateCompany) Error.CodeError { return ok() }
		pg.joinCompany = func(_ context.Context, _, _ string) Error.CodeError { return ok() }
		pg.setCompanyEmployeeRole = func(_ context.Context, _, _, _ string) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.CreateCompany(ctx, req)
		assertNoError(t, err)
		if res.GetCompanyUuid() == "" {
			t.Error("expected non-empty company uuid in response")
		}
	})

	t.Run("create company db error", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.createCompany = func(_ context.Context, _ *entities.CreateCompany) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.CreateCompany(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("join company fails — company is rolled back", func(t *testing.T) {
		deleted := false
		pg := emptyPGRepo()
		pg.createCompany = func(_ context.Context, _ *entities.CreateCompany) Error.CodeError { return ok() }
		pg.joinCompany = func(_ context.Context, _, _ string) Error.CodeError { return internalErr() }
		pg.deleteCompany = func(_ context.Context, _ string) Error.CodeError {
			deleted = true
			return ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.CreateCompany(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
		if !deleted {
			t.Error("expected DeleteCompany rollback to be called")
		}
	})

	t.Run("set role fails — company is rolled back", func(t *testing.T) {
		deleted := false
		pg := emptyPGRepo()
		pg.createCompany = func(_ context.Context, _ *entities.CreateCompany) Error.CodeError { return ok() }
		pg.joinCompany = func(_ context.Context, _, _ string) Error.CodeError { return ok() }
		pg.setCompanyEmployeeRole = func(_ context.Context, _, _, _ string) Error.CodeError { return internalErr() }
		pg.deleteCompany = func(_ context.Context, _ string) Error.CodeError {
			deleted = true
			return ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.CreateCompany(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
		if !deleted {
			t.Error("expected DeleteCompany rollback to be called")
		}
	})
}

// ─── GetCompany ───────────────────────────────────────────────────────────────

func TestGetCompany(t *testing.T) {
	ctx := context.Background()
	req := &pb.GetCompanyRequest{OperationId: opID, CompanyUuid: companyID}

	t.Run("success", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return &entities.Company{CompanyUUID: companyID, Title: "ACME", Status: "open"}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetCompany(ctx, req)
		assertNoError(t, err)
		if res.GetTitle() != "ACME" {
			t.Errorf("expected title %q, got %q", "ACME", res.GetTitle())
		}
	})

	t.Run("company not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompany(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})
}

// ─── GetCompanies ─────────────────────────────────────────────────────────────

func TestGetCompanies(t *testing.T) {
	ctx := context.Background()

	t.Run("success — returns mapped companies", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompanies = func(_ context.Context, _, _ int64) ([]*entities.GetCompanies, Error.CodeError) {
			return []*entities.GetCompanies{
				{CompanyUUID: "c1", Title: "Company 1", Status: "open"},
				{CompanyUUID: "c2", Title: "Company 2", Status: "closed"},
			}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetCompanies(ctx, &pb.GetCompaniesRequest{OperationId: opID, Offset: 0, Count: 10})
		assertNoError(t, err)
		if len(res.GetCompanies()) != 2 {
			t.Errorf("expected 2 companies, got %d", len(res.GetCompanies()))
		}
	})

	t.Run("negative offset", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetCompanies(ctx, &pb.GetCompaniesRequest{OperationId: opID, Offset: -1, Count: 10})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("count zero", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetCompanies(ctx, &pb.GetCompaniesRequest{OperationId: opID, Offset: 0, Count: 0})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("count too large", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetCompanies(ctx, &pb.GetCompaniesRequest{OperationId: opID, Offset: 0, Count: 101})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("db error", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompanies = func(_ context.Context, _, _ int64) ([]*entities.GetCompanies, Error.CodeError) {
			return nil, internalErr()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanies(ctx, &pb.GetCompaniesRequest{OperationId: opID, Offset: 0, Count: 10})
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── UpdateCompanyTitle ───────────────────────────────────────────────────────

func TestUpdateCompanyTitle(t *testing.T) {
	ctx := context.Background()
	req := &pb.UpdateCompanyTitleRequest{
		OperationId: opID, CompanyUuid: companyID, InitiatorUuid: initiatorID, Title: "New Title",
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.updateCompanyTitle = func(_ context.Context, _, _ string) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyTitle(ctx, req)
		assertNoError(t, err)
	})

	t.Run("company not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyTitle(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("user not in company", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyTitle(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("wrong role — engineer cannot update title", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
			return &entities.Employee{Role: "engineer"}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyTitle(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("db update error", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.updateCompanyTitle = func(_ context.Context, _, _ string) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyTitle(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── UpdateCompanyStatus ──────────────────────────────────────────────────────

func TestUpdateCompanyStatus(t *testing.T) {
	ctx := context.Background()
	req := &pb.UpdateCompanyStatusRequest{
		OperationId: opID, CompanyUuid: companyID, InitiatorUuid: initiatorID, Status: "closed",
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.updateCompanyStatus = func(_ context.Context, _, _ string) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyStatus(ctx, req)
		assertNoError(t, err)
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyStatus(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("db update error", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.updateCompanyStatus = func(_ context.Context, _, _ string) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyStatus(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── DeleteCompany ────────────────────────────────────────────────────────────

func TestDeleteCompany(t *testing.T) {
	ctx := context.Background()
	req := &pb.DeleteCompanyRequest{OperationId: opID, CompanyUuid: companyID, InitiatorUuid: initiatorID}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.deleteCompany = func(_ context.Context, _ string) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteCompany(ctx, req)
		assertNoError(t, err)
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteCompany(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("db delete error", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.deleteCompany = func(_ context.Context, _ string) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteCompany(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── CreateCompanyJoinCode ────────────────────────────────────────────────────

func TestCreateCompanyJoinCode(t *testing.T) {
	ctx := context.Background()
	validReq := &pb.CreateCompanyJoinCodeRequest{
		OperationId: opID, CompanyUuid: companyID, InitiatorUuid: initiatorID, CodeTtl: 3600,
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ string) Error.CodeError { return notFound() }
		rdb.createCompanyJoinCode = func(_ context.Context, _, _ string, _ time.Duration) Error.CodeError {
			return ok()
		}

		svc := newTestService(pg, rdb)
		res, err := svc.CreateCompanyJoinCode(ctx, validReq)
		assertNoError(t, err)
		if len(res.GetJoinCode()) != JoinCodeLength {
			t.Errorf("expected join code length %d, got %d", JoinCodeLength, len(res.GetJoinCode()))
		}
	})

	t.Run("ttl too short", func(t *testing.T) {
		svc := newTestService(pgRepoWithChief(), emptyRedisRepo())
		_, err := svc.CreateCompanyJoinCode(ctx, &pb.CreateCompanyJoinCodeRequest{
			OperationId: opID, CompanyUuid: companyID, InitiatorUuid: initiatorID, CodeTtl: 59,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("ttl too long", func(t *testing.T) {
		svc := newTestService(pgRepoWithChief(), emptyRedisRepo())
		_, err := svc.CreateCompanyJoinCode(ctx, &pb.CreateCompanyJoinCodeRequest{
			OperationId: opID, CompanyUuid: companyID, InitiatorUuid: initiatorID, CodeTtl: 60*60*24*7 + 1,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.CreateCompanyJoinCode(ctx, validReq)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("all 10 attempts produce non-unique codes — internal error", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ string) Error.CodeError { return ok() }

		svc := newTestService(pg, rdb)
		_, err := svc.CreateCompanyJoinCode(ctx, validReq)
		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("save join code error", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ string) Error.CodeError { return notFound() }
		rdb.createCompanyJoinCode = func(_ context.Context, _, _ string, _ time.Duration) Error.CodeError {
			return internalErr()
		}

		svc := newTestService(pg, rdb)
		_, err := svc.CreateCompanyJoinCode(ctx, validReq)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── GetCompanyJoinCodes ──────────────────────────────────────────────────────

func TestGetCompanyJoinCodes(t *testing.T) {
	ctx := context.Background()
	req := &pb.GetCompanyJoinCodesRequest{OperationId: opID, CompanyUuid: companyID, InitiatorUuid: initiatorID}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.getCompanyJoinCodes = func(_ context.Context, _ string) ([]string, Error.CodeError) {
			return []string{"111111", "222222"}, ok()
		}

		svc := newTestService(pg, rdb)
		res, err := svc.GetCompanyJoinCodes(ctx, req)
		assertNoError(t, err)
		if len(res.GetCodes()) != 2 {
			t.Errorf("expected 2 codes, got %d", len(res.GetCodes()))
		}
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanyJoinCodes(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("cache error", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.getCompanyJoinCodes = func(_ context.Context, _ string) ([]string, Error.CodeError) {
			return nil, internalErr()
		}

		svc := newTestService(pg, rdb)
		_, err := svc.GetCompanyJoinCodes(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── DeleteCompanyJoinCode ────────────────────────────────────────────────────

func TestDeleteCompanyJoinCode(t *testing.T) {
	ctx := context.Background()
	req := &pb.DeleteCompanyJoinCodeRequest{
		OperationId: opID, CompanyUuid: companyID, InitiatorUuid: initiatorID, Code: testCode,
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ string) Error.CodeError { return ok() }
		rdb.checkJoinCodeBelongToCompany = func(_ context.Context, _, _ string) Error.CodeError { return ok() }
		rdb.deleteCompanyJoinCode = func(_ context.Context, _, _ string) Error.CodeError { return ok() }

		svc := newTestService(pg, rdb)
		_, err := svc.DeleteCompanyJoinCode(ctx, req)
		assertNoError(t, err)
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteCompanyJoinCode(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("join code not found", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ string) Error.CodeError { return notFound() }

		svc := newTestService(pg, rdb)
		_, err := svc.DeleteCompanyJoinCode(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("join code does not belong to company", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ string) Error.CodeError { return ok() }
		rdb.checkJoinCodeBelongToCompany = func(_ context.Context, _, _ string) Error.CodeError {
			return Error.CodeError{Code: int(codes.PermissionDenied), Err: fmt.Errorf("not belong")}
		}

		svc := newTestService(pg, rdb)
		_, err := svc.DeleteCompanyJoinCode(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("delete cache error", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ string) Error.CodeError { return ok() }
		rdb.checkJoinCodeBelongToCompany = func(_ context.Context, _, _ string) Error.CodeError { return ok() }
		rdb.deleteCompanyJoinCode = func(_ context.Context, _, _ string) Error.CodeError { return internalErr() }

		svc := newTestService(pg, rdb)
		_, err := svc.DeleteCompanyJoinCode(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── JoinCompany ──────────────────────────────────────────────────────────────

func TestJoinCompany(t *testing.T) {
	ctx := context.Background()
	req := &pb.JoinCompanyRequest{OperationId: opID, JoinCode: testCode, InitiatorUuid: targetID}

	joinRdb := func() *mockRedisCompanyRepo {
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ string) Error.CodeError { return ok() }
		rdb.getCompanyByJoinCode = func(_ context.Context, _ string) (string, Error.CodeError) {
			return companyID, ok()
		}
		return rdb
	}

	t.Run("success", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
			return nil, notFound()
		}
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return &entities.Company{Status: "open"}, ok()
		}
		pg.joinCompany = func(_ context.Context, _, _ string) Error.CodeError { return ok() }

		svc := newTestService(pg, joinRdb())
		res, err := svc.JoinCompany(ctx, req)
		assertNoError(t, err)
		if res.GetCompanyUuid() != companyID {
			t.Errorf("expected company uuid %q, got %q", companyID, res.GetCompanyUuid())
		}
		if res.GetRole() != "unemployed" {
			t.Errorf("expected role %q, got %q", "unemployed", res.GetRole())
		}
	})

	t.Run("join code not found", func(t *testing.T) {
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ string) Error.CodeError { return notFound() }

		svc := newTestService(emptyPGRepo(), rdb)
		_, err := svc.JoinCompany(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("get company by code error", func(t *testing.T) {
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ string) Error.CodeError { return ok() }
		rdb.getCompanyByJoinCode = func(_ context.Context, _ string) (string, Error.CodeError) {
			return "", internalErr()
		}

		svc := newTestService(emptyPGRepo(), rdb)
		_, err := svc.JoinCompany(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("user already in company", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
			return chiefEmployee(), ok()
		}

		svc := newTestService(pg, joinRdb())
		_, err := svc.JoinCompany(ctx, req)
		assertGRPCCode(t, err, codes.AlreadyExists)
	})

	t.Run("get employee internal error", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
			return nil, internalErr()
		}

		svc := newTestService(pg, joinRdb())
		_, err := svc.JoinCompany(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("company is closed", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
			return nil, notFound()
		}
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return &entities.Company{Status: "closed"}, ok()
		}

		svc := newTestService(pg, joinRdb())
		_, err := svc.JoinCompany(ctx, req)
		assertGRPCCode(t, err, codes.Canceled)
	})

	t.Run("join company db error", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
			return nil, notFound()
		}
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return &entities.Company{Status: "open"}, ok()
		}
		pg.joinCompany = func(_ context.Context, _, _ string) Error.CodeError { return internalErr() }

		svc := newTestService(pg, joinRdb())
		_, err := svc.JoinCompany(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── GetCompanyEmployee ───────────────────────────────────────────────────────

func TestGetCompanyEmployee(t *testing.T) {
	ctx := context.Background()
	req := &pb.GetCompanyEmployeeRequest{
		OperationId: opID, CompanyUuid: companyID, InitiatorUuid: initiatorID, TargetUuid: targetID,
	}

	t.Run("success", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, userUUID string) (*entities.Employee, Error.CodeError) {
			if userUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return &entities.Employee{Role: "engineer", JoinedAt: "2024-01-01"}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetCompanyEmployee(ctx, req)
		assertNoError(t, err)
		if res.GetRole() != "engineer" {
			t.Errorf("expected role %q, got %q", "engineer", res.GetRole())
		}
	})

	t.Run("check role fails — initiator not in company", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanyEmployee(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("target employee not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, userUUID string) (*entities.Employee, Error.CodeError) {
			if userUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanyEmployee(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})
}

// ─── GetCompanyEmployees ──────────────────────────────────────────────────────

func TestGetCompanyEmployees(t *testing.T) {
	ctx := context.Background()

	makeReq := func(role, departmentUUID string, count, offset int64) *pb.GetCompanyEmployeesRequest {
		return &pb.GetCompanyEmployeesRequest{
			OperationId:    opID,
			CompanyUuid:    companyID,
			InitiatorUuid:  initiatorID,
			Role:           role,
			DepartmentUuid: departmentUUID,
			Count:          count,
			Offset:         offset,
		}
	}

	t.Run("success without filters", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.getCompanyEmployees = func(_ context.Context, _, _, _ string, _, _ int64) ([]*entities.Employee, Error.CodeError) {
			return []*entities.Employee{
				{UserUUID: "u1", Role: "engineer"},
				{UserUUID: "u2", Role: "manager"},
			}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetCompanyEmployees(ctx, makeReq("", "", 10, 0))
		assertNoError(t, err)
		if len(res.GetEmployees()) != 2 {
			t.Errorf("expected 2 employees, got %d", len(res.GetEmployees()))
		}
	})

	t.Run("success with role filter", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.getCompanyEmployees = func(_ context.Context, _, _, role string, _, _ int64) ([]*entities.Employee, Error.CodeError) {
			if role != "engineer" {
				return nil, internalErr()
			}
			return []*entities.Employee{
				{UserUUID: "u1", Role: "engineer"},
			}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetCompanyEmployees(ctx, makeReq("engineer", "", 10, 0))
		assertNoError(t, err)
		if len(res.GetEmployees()) != 1 {
			t.Errorf("expected 1 employee, got %d", len(res.GetEmployees()))
		}
	})

	t.Run("success with department filter", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.getCompanyEmployees = func(_ context.Context, _, deptUUID, _ string, _, _ int64) ([]*entities.Employee, Error.CodeError) {
			if deptUUID != deptID {
				return nil, internalErr()
			}
			return []*entities.Employee{
				{UserUUID: "u1", Role: "engineer", DepartmentUUID: deptID},
			}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetCompanyEmployees(ctx, makeReq("", deptID, 10, 0))
		assertNoError(t, err)
		if len(res.GetEmployees()) != 1 {
			t.Errorf("expected 1 employee, got %d", len(res.GetEmployees()))
		}
	})

	t.Run("invalid role", func(t *testing.T) {
		svc := newTestService(pgRepoWithChief(), emptyRedisRepo())
		_, err := svc.GetCompanyEmployees(ctx, makeReq("hacker", "", 10, 0))
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("count zero", func(t *testing.T) {
		svc := newTestService(pgRepoWithChief(), emptyRedisRepo())
		_, err := svc.GetCompanyEmployees(ctx, makeReq("", "", 0, 0))
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("count too large", func(t *testing.T) {
		svc := newTestService(pgRepoWithChief(), emptyRedisRepo())
		_, err := svc.GetCompanyEmployees(ctx, makeReq("", "", 101, 0))
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("negative offset", func(t *testing.T) {
		svc := newTestService(pgRepoWithChief(), emptyRedisRepo())
		_, err := svc.GetCompanyEmployees(ctx, makeReq("", "", 10, -1))
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanyEmployees(ctx, makeReq("", "", 10, 0))
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("db error", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.getCompanyEmployees = func(_ context.Context, _, _, _ string, _, _ int64) ([]*entities.Employee, Error.CodeError) {
			return nil, internalErr()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanyEmployees(ctx, makeReq("", "", 10, 0))
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── GetCompanyEmployeesSummary ───────────────────────────────────────────────

func TestGetCompanyEmployeesSummary(t *testing.T) {
	ctx := context.Background()
	req := &pb.GetCompanyEmployeesSummaryRequest{
		OperationId: opID, CompanyUuid: companyID, InitiatorUuid: initiatorID,
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.getCompanyEmployeesSummary = func(_ context.Context, _, _ string) (*entities.EmployeesSummary, Error.CodeError) {
			return &entities.EmployeesSummary{
				ChiefCount: 1, EngineerCount: 3, ManagerCount: 2, AnalyticCount: 1, UnemployedCount: 5, InspectorCount: 6,
			}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetCompanyEmployeesSummary(ctx, req)
		assertNoError(t, err)
		if res.GetChiefCount() != 1 || res.GetEngineerCount() != 3 {
			t.Errorf("unexpected summary counts: chief=%d engineer=%d", res.GetChiefCount(), res.GetEngineerCount())
		}
	})

	t.Run("success with department filter", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.getCompanyEmployeesSummary = func(_ context.Context, _, deptUUID string) (*entities.EmployeesSummary, Error.CodeError) {
			if deptUUID != deptID {
				return nil, internalErr()
			}
			return &entities.EmployeesSummary{EngineerCount: 2}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetCompanyEmployeesSummary(ctx, &pb.GetCompanyEmployeesSummaryRequest{
			OperationId: opID, CompanyUuid: companyID, InitiatorUuid: initiatorID, DepartmentUuid: deptID,
		})
		assertNoError(t, err)
		if res.GetEngineerCount() != 2 {
			t.Errorf("expected 2 engineers, got %d", res.GetEngineerCount())
		}
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanyEmployeesSummary(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("db error", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.getCompanyEmployeesSummary = func(_ context.Context, _, _ string) (*entities.EmployeesSummary, Error.CodeError) {
			return nil, internalErr()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanyEmployeesSummary(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── UpdateEmployeeRole ───────────────────────────────────────────────────────

func TestUpdateEmployeeRole(t *testing.T) {
	ctx := context.Background()
	req := &pb.UpdateEmployeeRoleRequest{
		OperationId:   opID,
		CompanyUuid:   companyID,
		InitiatorUuid: initiatorID,
		TargetUuid:    targetID,
		Role:          "chief",
	}

	t.Run("success", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, userUUID string) (*entities.Employee, Error.CodeError) {
			if userUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return &entities.Employee{Role: "engineer"}, ok()
		}
		pg.setCompanyEmployeeRole = func(_ context.Context, _, _, _ string) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateEmployeeRole(ctx, req)
		assertNoError(t, err)
	})

	t.Run("update self role", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return nil, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateEmployeeRole(ctx, &pb.UpdateEmployeeRoleRequest{
			OperationId:   opID,
			CompanyUuid:   companyID,
			InitiatorUuid: initiatorID,
			TargetUuid:    initiatorID,
			Role:          "chief",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateEmployeeRole(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("target employee not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, userUUID string) (*entities.Employee, Error.CodeError) {
			if userUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateEmployeeRole(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("set role db error", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, userUUID string) (*entities.Employee, Error.CodeError) {
			if userUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return &entities.Employee{Role: "engineer"}, ok()
		}
		pg.setCompanyEmployeeRole = func(_ context.Context, _, _, _ string) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateEmployeeRole(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── RemoveCompanyEmployee ────────────────────────────────────────────────────

func TestRemoveCompanyEmployee(t *testing.T) {
	ctx := context.Background()

	makeReq := func(initiator, target string) *pb.RemoveCompanyEmployeeRequest {
		return &pb.RemoveCompanyEmployeeRequest{
			OperationId:   opID,
			CompanyUuid:   companyID,
			InitiatorUuid: initiator,
			TargetUuid:    target,
		}
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.removeCompanyEmployee = func(_ context.Context, _, _ string) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveCompanyEmployee(ctx, makeReq(initiatorID, targetID))
		assertNoError(t, err)
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveCompanyEmployee(ctx, makeReq(initiatorID, targetID))
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("cannot remove yourself", func(t *testing.T) {
		svc := newTestService(pgRepoWithChief(), emptyRedisRepo())
		_, err := svc.RemoveCompanyEmployee(ctx, makeReq(initiatorID, initiatorID))
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("target employee not found", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.removeCompanyEmployee = func(_ context.Context, _, _ string) Error.CodeError { return notFound() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveCompanyEmployee(ctx, makeReq(initiatorID, targetID))
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("db delete error", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.removeCompanyEmployee = func(_ context.Context, _, _ string) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveCompanyEmployee(ctx, makeReq(initiatorID, targetID))
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── CreateDepartment ─────────────────────────────────────────────────────────

func TestCreateDepartment(t *testing.T) {
	ctx := context.Background()
	req := &pb.CreateDepartmentRequest{
		OperationId: opID, CompanyUuid: companyID, InitiatorUuid: initiatorID, Title: "Engineering",
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.createDepartment = func(_ context.Context, _ *entities.CreateDepartment) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.CreateDepartment(ctx, req)
		assertNoError(t, err)
		if res.GetDepartmentUuid() == "" {
			t.Error("expected non-empty department uuid in response")
		}
	})

	t.Run("invalid title — empty", func(t *testing.T) {
		svc := newTestService(pgRepoWithChief(), emptyRedisRepo())
		_, err := svc.CreateDepartment(ctx, &pb.CreateDepartmentRequest{
			OperationId: opID, CompanyUuid: companyID, InitiatorUuid: initiatorID, Title: "   ",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("check role fails — company not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.CreateDepartment(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("check role fails — not chief", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
			return &entities.Employee{Role: "engineer"}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.CreateDepartment(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("db error", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.createDepartment = func(_ context.Context, _ *entities.CreateDepartment) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.CreateDepartment(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── AddEmployeeToDepartment ──────────────────────────────────────────────────

func TestAddEmployeeToDepartment(t *testing.T) {
	ctx := context.Background()
	req := &pb.AddEmployeeToDepartmentRequest{
		OperationId: opID, InitiatorUuid: initiatorID, DepartmentUuid: deptID, TargetUuid: targetID,
	}

	t.Run("success", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ string) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, userUUID string) (*entities.Employee, Error.CodeError) {
			if userUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return &entities.Employee{Role: "engineer"}, ok()
		}
		pg.addEmployeeToDepartment = func(_ context.Context, _, _, _ string) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.AddEmployeeToDepartment(ctx, req)
		assertNoError(t, err)
	})

	t.Run("department not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ string) (*entities.Department, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.AddEmployeeToDepartment(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("initiator not chief", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ string) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
			return &entities.Employee{Role: "manager"}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.AddEmployeeToDepartment(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("target not in company", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ string) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, userUUID string) (*entities.Employee, Error.CodeError) {
			if userUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.AddEmployeeToDepartment(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("db error", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ string) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, userUUID string) (*entities.Employee, Error.CodeError) {
			if userUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return &entities.Employee{Role: "engineer"}, ok()
		}
		pg.addEmployeeToDepartment = func(_ context.Context, _, _, _ string) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.AddEmployeeToDepartment(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── GetDepartment ────────────────────────────────────────────────────────────

func TestGetDepartment(t *testing.T) {
	ctx := context.Background()
	req := &pb.GetDepartmentRequest{
		OperationId: opID, InitiatorUuid: initiatorID, DepartmentUuid: deptID,
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChiefAndDept()

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetDepartment(ctx, req)
		assertNoError(t, err)
		if res.GetTitle() != "Test Dept" {
			t.Errorf("expected title %q, got %q", "Test Dept", res.GetTitle())
		}
		if res.GetCompanyUuid() != companyID {
			t.Errorf("expected company uuid %q, got %q", companyID, res.GetCompanyUuid())
		}
	})

	t.Run("department not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ string) (*entities.Department, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetDepartment(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("initiator not in company", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ string) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetDepartment(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})
}

// ─── GetCompanyDepartments ────────────────────────────────────────────────────

func TestGetCompanyDepartments(t *testing.T) {
	ctx := context.Background()

	makeReq := func(offset, count int64) *pb.GetCompanyDepartmentsRequest {
		return &pb.GetCompanyDepartmentsRequest{
			OperationId: opID, CompanyUuid: companyID, InitiatorUuid: initiatorID,
			Offset: offset, Count: count,
		}
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.getCompanyDepartments = func(_ context.Context, _ string, _, _ int64) ([]*entities.Department, Error.CodeError) {
			return []*entities.Department{
				{UUID: "d1", Title: "Engineering"},
				{UUID: "d2", Title: "Marketing"},
			}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetCompanyDepartments(ctx, makeReq(0, 10))
		assertNoError(t, err)
		if len(res.GetDepartments()) != 2 {
			t.Errorf("expected 2 departments, got %d", len(res.GetDepartments()))
		}
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanyDepartments(ctx, makeReq(0, 10))
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("negative offset", func(t *testing.T) {
		svc := newTestService(pgRepoWithChief(), emptyRedisRepo())
		_, err := svc.GetCompanyDepartments(ctx, makeReq(-1, 10))
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("count zero", func(t *testing.T) {
		svc := newTestService(pgRepoWithChief(), emptyRedisRepo())
		_, err := svc.GetCompanyDepartments(ctx, makeReq(0, 0))
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("count too large", func(t *testing.T) {
		svc := newTestService(pgRepoWithChief(), emptyRedisRepo())
		_, err := svc.GetCompanyDepartments(ctx, makeReq(0, 101))
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("db error", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.getCompanyDepartments = func(_ context.Context, _ string, _, _ int64) ([]*entities.Department, Error.CodeError) {
			return nil, internalErr()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanyDepartments(ctx, makeReq(0, 10))
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── UpdateDepartmentTitle ────────────────────────────────────────────────────

func TestUpdateDepartmentTitle(t *testing.T) {
	ctx := context.Background()
	req := &pb.UpdateDepartmentTitleRequest{
		OperationId: opID, InitiatorUuid: initiatorID, DepartmentUuid: deptID, Title: "New Name",
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChiefAndDept()
		pg.updateDepartmentTitle = func(_ context.Context, _ *entities.UpdateDepartment) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateDepartmentTitle(ctx, req)
		assertNoError(t, err)
	})

	t.Run("invalid title — empty", func(t *testing.T) {
		svc := newTestService(pgRepoWithChiefAndDept(), emptyRedisRepo())
		_, err := svc.UpdateDepartmentTitle(ctx, &pb.UpdateDepartmentTitleRequest{
			OperationId: opID, InitiatorUuid: initiatorID, DepartmentUuid: deptID, Title: "  ",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("department not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ string) (*entities.Department, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateDepartmentTitle(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("initiator not chief", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ string) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
			return &entities.Employee{Role: "manager"}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateDepartmentTitle(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("db update error", func(t *testing.T) {
		pg := pgRepoWithChiefAndDept()
		pg.updateDepartmentTitle = func(_ context.Context, _ *entities.UpdateDepartment) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateDepartmentTitle(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── DeleteDepartment ─────────────────────────────────────────────────────────

func TestDeleteDepartment(t *testing.T) {
	ctx := context.Background()
	req := &pb.DeleteDepartmentRequest{
		OperationId: opID, InitiatorUuid: initiatorID, DepartmentUuid: deptID,
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChiefAndDept()
		pg.deleteDepartment = func(_ context.Context, _ string) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteDepartment(ctx, req)
		assertNoError(t, err)
	})

	t.Run("department not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ string) (*entities.Department, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteDepartment(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("initiator not chief", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ string) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
			return &entities.Employee{Role: "manager"}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteDepartment(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("db delete error", func(t *testing.T) {
		pg := pgRepoWithChiefAndDept()
		pg.deleteDepartment = func(_ context.Context, _ string) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteDepartment(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── RemoveEmployeeFromDepartment ─────────────────────────────────────────────

func TestRemoveEmployeeFromDepartment(t *testing.T) {
	ctx := context.Background()
	req := &pb.RemoveEmployeeFromDepartmentRequest{
		OperationId: opID, InitiatorUuid: initiatorID, DepartmentUuid: deptID, TargetUuid: targetID,
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChiefAndDept()
		// target is in this department
		pg.getCompanyEmployee = func(_ context.Context, _, userUUID string) (*entities.Employee, Error.CodeError) {
			if userUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return &entities.Employee{Role: "engineer", DepartmentUUID: deptID}, ok()
		}
		pg.removeEmployeeFromDepartment = func(_ context.Context, _, _ string) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveEmployeeFromDepartment(ctx, req)
		assertNoError(t, err)
	})

	t.Run("department not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ string) (*entities.Department, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveEmployeeFromDepartment(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("initiator not chief", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ string) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ string) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _, _ string) (*entities.Employee, Error.CodeError) {
			return &entities.Employee{Role: "manager"}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveEmployeeFromDepartment(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("target employee not found", func(t *testing.T) {
		pg := pgRepoWithChiefAndDept()
		pg.getCompanyEmployee = func(_ context.Context, _, userUUID string) (*entities.Employee, Error.CodeError) {
			if userUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveEmployeeFromDepartment(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("target not in this department", func(t *testing.T) {
		pg := pgRepoWithChiefAndDept()
		pg.getCompanyEmployee = func(_ context.Context, _, userUUID string) (*entities.Employee, Error.CodeError) {
			if userUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return &entities.Employee{Role: "engineer", DepartmentUUID: "other-dept-uuid"}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveEmployeeFromDepartment(ctx, req)
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("db remove error", func(t *testing.T) {
		pg := pgRepoWithChiefAndDept()
		pg.getCompanyEmployee = func(_ context.Context, _, userUUID string) (*entities.Employee, Error.CodeError) {
			if userUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return &entities.Employee{Role: "engineer", DepartmentUUID: deptID}, ok()
		}
		pg.removeEmployeeFromDepartment = func(_ context.Context, _, _ string) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveEmployeeFromDepartment(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}
