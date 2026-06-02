package services

import (
	"context"
	"testing"

	"github.com/unwelcome/FrameWorkTask1/backend/company/internal/entities"
	pb "github.com/unwelcome/FrameWorkTask1/backend/contracts/company/generated"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─── Тестовые константы ───────────────────────────────────────────────────────

const (
	opID        = "op-test-1"
	companyID   = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	initiatorID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	targetID    = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	testCode    = "123456"
	deptID      = "dddddddd-dddd-dddd-dddd-dddddddddddd"
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
	req := &pb.CreateCompanyRequest{Title: "ACME", InitiatorUuid: initiatorID}

	t.Run("success", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.createCompany = func(_ context.Context, _ *entities.CreateCompany) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.CreateCompany(ctx, req)
		assertNoError(t, err)
		if res.GetCompanyUuid() == "" {
			t.Error("expected non-empty company uuid in response")
		}
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.CreateCompany(ctx, &pb.CreateCompanyRequest{
			InitiatorUuid: "not-a-uuid",
			Title:         "ACME",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_title", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.CreateCompany(ctx, &pb.CreateCompanyRequest{
			InitiatorUuid: initiatorID,
			Title:         "",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("db error", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.createCompany = func(_ context.Context, _ *entities.CreateCompany) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.CreateCompany(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── GetCompany ───────────────────────────────────────────────────────────────

func TestGetCompany(t *testing.T) {
	ctx := context.Background()
	// InitiatorUuid обязателен — GetCompany его валидирует
	req := &pb.GetCompanyRequest{InitiatorUuid: initiatorID, CompanyUuid: companyID}

	t.Run("success", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return &entities.Company{CompanyUUID: companyID, Title: "ACME", Status: "open"}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetCompany(ctx, req)
		assertNoError(t, err)
		if res.GetTitle() != "ACME" {
			t.Errorf("expected title %q, got %q", "ACME", res.GetTitle())
		}
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetCompany(ctx, &pb.GetCompanyRequest{
			InitiatorUuid: "not-a-uuid",
			CompanyUuid:   companyID,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_company_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetCompany(ctx, &pb.GetCompanyRequest{
			InitiatorUuid: initiatorID,
			CompanyUuid:   "not-a-uuid",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("company not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
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
		pg.getCompanies = func(_ context.Context, _ entities.GetCompaniesDTO) ([]*entities.GetCompanies, Error.CodeError) {
			return []*entities.GetCompanies{
				{CompanyUUID: "c1", Title: "Company 1", Status: "open"},
				{CompanyUUID: "c2", Title: "Company 2", Status: "closed"},
			}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetCompanies(ctx, &pb.GetCompaniesRequest{Offset: 0, Count: 10})
		assertNoError(t, err)
		if len(res.GetCompanies()) != 2 {
			t.Errorf("expected 2 companies, got %d", len(res.GetCompanies()))
		}
	})

	t.Run("negative offset", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetCompanies(ctx, &pb.GetCompaniesRequest{Offset: -1, Count: 10})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("count zero", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetCompanies(ctx, &pb.GetCompaniesRequest{Offset: 0, Count: 0})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("count too large", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetCompanies(ctx, &pb.GetCompaniesRequest{Offset: 0, Count: 101})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("db error", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompanies = func(_ context.Context, _ entities.GetCompaniesDTO) ([]*entities.GetCompanies, Error.CodeError) {
			return nil, internalErr()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanies(ctx, &pb.GetCompaniesRequest{Offset: 0, Count: 10})
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── UpdateCompanyTitle ───────────────────────────────────────────────────────

func TestUpdateCompanyTitle(t *testing.T) {
	ctx := context.Background()
	req := &pb.UpdateCompanyTitleRequest{
		CompanyUuid: companyID, InitiatorUuid: initiatorID, Title: "New Title",
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.updateCompanyTitle = func(_ context.Context, _ entities.UpdateCompanyTitleDTO) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyTitle(ctx, req)
		assertNoError(t, err)
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.UpdateCompanyTitle(ctx, &pb.UpdateCompanyTitleRequest{
			InitiatorUuid: "not-a-uuid",
			CompanyUuid:   companyID,
			Title:         "New Title",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_company_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.UpdateCompanyTitle(ctx, &pb.UpdateCompanyTitleRequest{
			InitiatorUuid: initiatorID,
			CompanyUuid:   "not-a-uuid",
			Title:         "New Title",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_title", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.UpdateCompanyTitle(ctx, &pb.UpdateCompanyTitleRequest{
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Title:         "",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("company not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyTitle(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("user not in company", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyTitle(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("wrong role — engineer cannot update title", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			return &entities.Employee{Role: "engineer"}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyTitle(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("db update error", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.updateCompanyTitle = func(_ context.Context, _ entities.UpdateCompanyTitleDTO) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyTitle(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── UpdateCompanyStatus ──────────────────────────────────────────────────────

func TestUpdateCompanyStatus(t *testing.T) {
	ctx := context.Background()
	// Допустимые статусы: "open" и "close" (не "closed")
	req := &pb.UpdateCompanyStatusRequest{
		CompanyUuid: companyID, InitiatorUuid: initiatorID, Status: "close",
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.updateCompanyStatus = func(_ context.Context, _ entities.UpdateCompanyStatusDTO) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyStatus(ctx, req)
		assertNoError(t, err)
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.UpdateCompanyStatus(ctx, &pb.UpdateCompanyStatusRequest{
			InitiatorUuid: "not-a-uuid",
			CompanyUuid:   companyID,
			Status:        "close",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_status", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.UpdateCompanyStatus(ctx, &pb.UpdateCompanyStatusRequest{
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Status:        "closed", // "closed" не входит в AllStatuses
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyStatus(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("db update error", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.updateCompanyStatus = func(_ context.Context, _ entities.UpdateCompanyStatusDTO) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateCompanyStatus(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── DeleteCompany ────────────────────────────────────────────────────────────

func TestDeleteCompany(t *testing.T) {
	ctx := context.Background()
	req := &pb.DeleteCompanyRequest{CompanyUuid: companyID, InitiatorUuid: initiatorID}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.deleteCompany = func(_ context.Context, _ entities.DeleteCompanyDTO) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteCompany(ctx, req)
		assertNoError(t, err)
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.DeleteCompany(ctx, &pb.DeleteCompanyRequest{
			InitiatorUuid: "not-a-uuid",
			CompanyUuid:   companyID,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_company_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.DeleteCompany(ctx, &pb.DeleteCompanyRequest{
			InitiatorUuid: initiatorID,
			CompanyUuid:   "not-a-uuid",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteCompany(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("db delete error", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.deleteCompany = func(_ context.Context, _ entities.DeleteCompanyDTO) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteCompany(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── CreateCompanyJoinCode ────────────────────────────────────────────────────

func TestCreateCompanyJoinCode(t *testing.T) {
	ctx := context.Background()
	validReq := &pb.CreateCompanyJoinCodeRequest{
		CompanyUuid: companyID, InitiatorUuid: initiatorID, CodeTtl: 3600,
	}

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.CreateCompanyJoinCode(ctx, &pb.CreateCompanyJoinCodeRequest{
			InitiatorUuid: "not-a-uuid",
			CompanyUuid:   companyID,
			CodeTtl:       3600,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_company_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.CreateCompanyJoinCode(ctx, &pb.CreateCompanyJoinCodeRequest{
			InitiatorUuid: initiatorID,
			CompanyUuid:   "not-a-uuid",
			CodeTtl:       3600,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ entities.CheckJoinCodeExistsDTO) Error.CodeError { return notFound() }
		rdb.createCompanyJoinCode = func(_ context.Context, _ entities.CreateCompanyJoinCodeDTO) Error.CodeError {
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
			CompanyUuid: companyID, InitiatorUuid: initiatorID, CodeTtl: 59,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("ttl too long", func(t *testing.T) {
		svc := newTestService(pgRepoWithChief(), emptyRedisRepo())
		_, err := svc.CreateCompanyJoinCode(ctx, &pb.CreateCompanyJoinCodeRequest{
			CompanyUuid: companyID, InitiatorUuid: initiatorID, CodeTtl: 60*60*24*7 + 1,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.CreateCompanyJoinCode(ctx, validReq)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("all 10 attempts produce non-unique codes — internal error", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ entities.CheckJoinCodeExistsDTO) Error.CodeError { return ok() }

		svc := newTestService(pg, rdb)
		_, err := svc.CreateCompanyJoinCode(ctx, validReq)
		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("save join code error", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ entities.CheckJoinCodeExistsDTO) Error.CodeError { return notFound() }
		rdb.createCompanyJoinCode = func(_ context.Context, _ entities.CreateCompanyJoinCodeDTO) Error.CodeError {
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
	req := &pb.GetCompanyJoinCodesRequest{CompanyUuid: companyID, InitiatorUuid: initiatorID}

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetCompanyJoinCodes(ctx, &pb.GetCompanyJoinCodesRequest{
			InitiatorUuid: "not-a-uuid",
			CompanyUuid:   companyID,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.getCompanyJoinCodes = func(_ context.Context, _ entities.GetCompanyJoinCodesDTO) ([]string, Error.CodeError) {
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
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanyJoinCodes(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("cache error", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.getCompanyJoinCodes = func(_ context.Context, _ entities.GetCompanyJoinCodesDTO) ([]string, Error.CodeError) {
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
		CompanyUuid: companyID, InitiatorUuid: initiatorID, Code: testCode,
	}

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.DeleteCompanyJoinCode(ctx, &pb.DeleteCompanyJoinCodeRequest{
			InitiatorUuid: "not-a-uuid",
			CompanyUuid:   companyID,
			Code:          testCode,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_code_format", func(t *testing.T) {
		// Код должен быть ровно 6 цифр
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.DeleteCompanyJoinCode(ctx, &pb.DeleteCompanyJoinCodeRequest{
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			Code:          "1234", // 4 цифры вместо 6
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ entities.CheckJoinCodeExistsDTO) Error.CodeError { return ok() }
		rdb.checkJoinCodeBelongToCompany = func(_ context.Context, _ entities.CheckJoinCodeBelongToCompanyDTO) Error.CodeError { return ok() }
		rdb.deleteCompanyJoinCode = func(_ context.Context, _ entities.DeleteCompanyJoinCodeDTO) Error.CodeError { return ok() }

		svc := newTestService(pg, rdb)
		_, err := svc.DeleteCompanyJoinCode(ctx, req)
		assertNoError(t, err)
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteCompanyJoinCode(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("join code not found", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ entities.CheckJoinCodeExistsDTO) Error.CodeError { return notFound() }

		svc := newTestService(pg, rdb)
		_, err := svc.DeleteCompanyJoinCode(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("join code does not belong to company", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ entities.CheckJoinCodeExistsDTO) Error.CodeError { return ok() }
		rdb.checkJoinCodeBelongToCompany = func(_ context.Context, _ entities.CheckJoinCodeBelongToCompanyDTO) Error.CodeError {
			return Error.Public(codes.PermissionDenied, "not belong")
		}

		svc := newTestService(pg, rdb)
		_, err := svc.DeleteCompanyJoinCode(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("delete cache error", func(t *testing.T) {
		pg := pgRepoWithChief()
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ entities.CheckJoinCodeExistsDTO) Error.CodeError { return ok() }
		rdb.checkJoinCodeBelongToCompany = func(_ context.Context, _ entities.CheckJoinCodeBelongToCompanyDTO) Error.CodeError { return ok() }
		rdb.deleteCompanyJoinCode = func(_ context.Context, _ entities.DeleteCompanyJoinCodeDTO) Error.CodeError { return internalErr() }

		svc := newTestService(pg, rdb)
		_, err := svc.DeleteCompanyJoinCode(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── JoinCompany ──────────────────────────────────────────────────────────────

func TestJoinCompany(t *testing.T) {
	ctx := context.Background()
	req := &pb.JoinCompanyRequest{JoinCode: testCode, InitiatorUuid: targetID}

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.JoinCompany(ctx, &pb.JoinCompanyRequest{
			InitiatorUuid: "not-a-uuid",
			JoinCode:      testCode,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_join_code", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.JoinCompany(ctx, &pb.JoinCompanyRequest{
			InitiatorUuid: targetID,
			JoinCode:      "abc", // не 6 цифр
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	joinRdb := func() *mockRedisCompanyRepo {
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ entities.CheckJoinCodeExistsDTO) Error.CodeError { return ok() }
		rdb.getCompanyByJoinCode = func(_ context.Context, _ entities.GetCompanyByJoinCodeDTO) (string, Error.CodeError) {
			return companyID, ok()
		}
		return rdb
	}

	t.Run("success", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			return nil, notFound()
		}
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return &entities.Company{Status: "open"}, ok()
		}
		pg.joinCompany = func(_ context.Context, _ entities.JoinCompanyDTO) Error.CodeError { return ok() }

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
		rdb.checkJoinCodeExists = func(_ context.Context, _ entities.CheckJoinCodeExistsDTO) Error.CodeError { return notFound() }

		svc := newTestService(emptyPGRepo(), rdb)
		_, err := svc.JoinCompany(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("get company by code error", func(t *testing.T) {
		rdb := emptyRedisRepo()
		rdb.checkJoinCodeExists = func(_ context.Context, _ entities.CheckJoinCodeExistsDTO) Error.CodeError { return ok() }
		rdb.getCompanyByJoinCode = func(_ context.Context, _ entities.GetCompanyByJoinCodeDTO) (string, Error.CodeError) {
			return "", internalErr()
		}

		svc := newTestService(emptyPGRepo(), rdb)
		_, err := svc.JoinCompany(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("user already in company", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			return chiefEmployee(), ok()
		}

		svc := newTestService(pg, joinRdb())
		_, err := svc.JoinCompany(ctx, req)
		assertGRPCCode(t, err, codes.AlreadyExists)
	})

	t.Run("get employee internal error", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			return nil, internalErr()
		}

		svc := newTestService(pg, joinRdb())
		_, err := svc.JoinCompany(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("company is closed", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			return nil, notFound()
		}
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return &entities.Company{Status: "closed"}, ok()
		}

		svc := newTestService(pg, joinRdb())
		_, err := svc.JoinCompany(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("join company db error", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			return nil, notFound()
		}
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return &entities.Company{Status: "open"}, ok()
		}
		pg.joinCompany = func(_ context.Context, _ entities.JoinCompanyDTO) Error.CodeError { return internalErr() }

		svc := newTestService(pg, joinRdb())
		_, err := svc.JoinCompany(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── GetCompanyEmployee ───────────────────────────────────────────────────────

func TestGetCompanyEmployee(t *testing.T) {
	ctx := context.Background()
	req := &pb.GetCompanyEmployeeRequest{
		CompanyUuid: companyID, InitiatorUuid: initiatorID, TargetUuid: targetID,
	}

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetCompanyEmployee(ctx, &pb.GetCompanyEmployeeRequest{
			InitiatorUuid: "not-a-uuid",
			CompanyUuid:   companyID,
			TargetUuid:    targetID,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_target_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetCompanyEmployee(ctx, &pb.GetCompanyEmployeeRequest{
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			TargetUuid:    "not-a-uuid",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("success", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, dto entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			if dto.UserUUID == initiatorID {
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
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanyEmployee(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("target employee not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, dto entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			if dto.UserUUID == initiatorID {
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
		pg.getCompanyEmployees = func(_ context.Context, _ entities.GetCompanyEmployeesDTO) ([]*entities.Employee, Error.CodeError) {
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
		pg.getCompanyEmployees = func(_ context.Context, dto entities.GetCompanyEmployeesDTO) ([]*entities.Employee, Error.CodeError) {
			if dto.Role != "engineer" {
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
		pg.getCompanyEmployees = func(_ context.Context, dto entities.GetCompanyEmployeesDTO) ([]*entities.Employee, Error.CodeError) {
			if dto.DepartmentUUID != deptID {
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

	t.Run("invalid_department_uuid", func(t *testing.T) {
		// Непустой невалидный UUID в поле DepartmentUuid → InvalidArgument
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetCompanyEmployees(ctx, makeReq("", "not-a-uuid", 10, 0))
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
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanyEmployees(ctx, makeReq("", "", 10, 0))
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("db error", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.getCompanyEmployees = func(_ context.Context, _ entities.GetCompanyEmployeesDTO) ([]*entities.Employee, Error.CodeError) {
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
		CompanyUuid: companyID, InitiatorUuid: initiatorID,
	}

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.getCompanyEmployeesSummary = func(_ context.Context, _ entities.GetCompanyEmployeesSummaryDTO) (*entities.EmployeesSummary, Error.CodeError) {
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
		pg.getCompanyEmployeesSummary = func(_ context.Context, dto entities.GetCompanyEmployeesSummaryDTO) (*entities.EmployeesSummary, Error.CodeError) {
			if dto.DepartmentUUID != deptID {
				return nil, internalErr()
			}
			return &entities.EmployeesSummary{EngineerCount: 2}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetCompanyEmployeesSummary(ctx, &pb.GetCompanyEmployeesSummaryRequest{
			CompanyUuid: companyID, InitiatorUuid: initiatorID, DepartmentUuid: deptID,
		})
		assertNoError(t, err)
		if res.GetEngineerCount() != 2 {
			t.Errorf("expected 2 engineers, got %d", res.GetEngineerCount())
		}
	})

	t.Run("invalid_department_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetCompanyEmployeesSummary(ctx, &pb.GetCompanyEmployeesSummaryRequest{
			InitiatorUuid:  initiatorID,
			CompanyUuid:    companyID,
			DepartmentUuid: "not-a-uuid",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetCompanyEmployeesSummary(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("db error", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.getCompanyEmployeesSummary = func(_ context.Context, _ entities.GetCompanyEmployeesSummaryDTO) (*entities.EmployeesSummary, Error.CodeError) {
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
		CompanyUuid:   companyID,
		InitiatorUuid: initiatorID,
		TargetUuid:    targetID,
		Role:          "engineer",
	}

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.UpdateEmployeeRole(ctx, &pb.UpdateEmployeeRoleRequest{
			InitiatorUuid: "not-a-uuid",
			CompanyUuid:   companyID,
			TargetUuid:    targetID,
			Role:          "engineer",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_target_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.UpdateEmployeeRole(ctx, &pb.UpdateEmployeeRoleRequest{
			InitiatorUuid: initiatorID,
			CompanyUuid:   companyID,
			TargetUuid:    "not-a-uuid",
			Role:          "engineer",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("success", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, dto entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			if dto.UserUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return &entities.Employee{Role: "engineer"}, ok()
		}
		pg.setCompanyEmployeeRole = func(_ context.Context, _ entities.SetCompanyEmployeeRoleDTO) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateEmployeeRole(ctx, req)
		assertNoError(t, err)
	})

	t.Run("update self role", func(t *testing.T) {
		// Проверка initiator == target происходит до checkEmployeeRole — мок БД не нужен
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.UpdateEmployeeRole(ctx, &pb.UpdateEmployeeRoleRequest{
			CompanyUuid:   companyID,
			InitiatorUuid: initiatorID,
			TargetUuid:    initiatorID,
			Role:          "chief",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateEmployeeRole(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("target employee not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, dto entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			if dto.UserUUID == initiatorID {
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
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, dto entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			if dto.UserUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return &entities.Employee{Role: "engineer"}, ok()
		}
		pg.setCompanyEmployeeRole = func(_ context.Context, _ entities.SetCompanyEmployeeRoleDTO) Error.CodeError { return internalErr() }

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
			CompanyUuid:   companyID,
			InitiatorUuid: initiator,
			TargetUuid:    target,
		}
	}

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.RemoveCompanyEmployee(ctx, makeReq("not-a-uuid", targetID))
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.removeCompanyEmployee = func(_ context.Context, _ entities.RemoveCompanyEmployeeDTO) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveCompanyEmployee(ctx, makeReq(initiatorID, targetID))
		assertNoError(t, err)
	})

	t.Run("check role fails", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
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
		pg.removeCompanyEmployee = func(_ context.Context, _ entities.RemoveCompanyEmployeeDTO) Error.CodeError { return notFound() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveCompanyEmployee(ctx, makeReq(initiatorID, targetID))
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("db delete error", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.removeCompanyEmployee = func(_ context.Context, _ entities.RemoveCompanyEmployeeDTO) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveCompanyEmployee(ctx, makeReq(initiatorID, targetID))
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── CreateDepartment ─────────────────────────────────────────────────────────

func TestCreateDepartment(t *testing.T) {
	ctx := context.Background()
	req := &pb.CreateDepartmentRequest{
		CompanyUuid: companyID, InitiatorUuid: initiatorID, Title: "Engineering",
	}

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.CreateDepartment(ctx, &pb.CreateDepartmentRequest{
			InitiatorUuid: "not-a-uuid",
			CompanyUuid:   companyID,
			Title:         "Engineering",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_company_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.CreateDepartment(ctx, &pb.CreateDepartmentRequest{
			InitiatorUuid: initiatorID,
			CompanyUuid:   "not-a-uuid",
			Title:         "Engineering",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

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
			CompanyUuid: companyID, InitiatorUuid: initiatorID, Title: "",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("check role fails — company not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.CreateDepartment(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("check role fails — not chief", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
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
		InitiatorUuid: initiatorID, DepartmentUuid: deptID, TargetUuid: targetID,
	}

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.AddEmployeeToDepartment(ctx, &pb.AddEmployeeToDepartmentRequest{
			InitiatorUuid:  "not-a-uuid",
			DepartmentUuid: deptID,
			TargetUuid:     targetID,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_department_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.AddEmployeeToDepartment(ctx, &pb.AddEmployeeToDepartmentRequest{
			InitiatorUuid:  initiatorID,
			DepartmentUuid: "not-a-uuid",
			TargetUuid:     targetID,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_target_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.AddEmployeeToDepartment(ctx, &pb.AddEmployeeToDepartmentRequest{
			InitiatorUuid:  initiatorID,
			DepartmentUuid: deptID,
			TargetUuid:     "not-a-uuid",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("success", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, dto entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			if dto.UserUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return &entities.Employee{Role: "engineer"}, ok()
		}
		pg.addEmployeeToDepartment = func(_ context.Context, _ entities.AddEmployeeToDepartmentDTO) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.AddEmployeeToDepartment(ctx, req)
		assertNoError(t, err)
	})

	t.Run("department not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.AddEmployeeToDepartment(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("initiator not chief", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			return &entities.Employee{Role: "manager"}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.AddEmployeeToDepartment(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("target not in company", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, dto entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			if dto.UserUUID == initiatorID {
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
		pg.getDepartment = func(_ context.Context, _ entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, dto entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			if dto.UserUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return &entities.Employee{Role: "engineer"}, ok()
		}
		pg.addEmployeeToDepartment = func(_ context.Context, _ entities.AddEmployeeToDepartmentDTO) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.AddEmployeeToDepartment(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── GetDepartment ────────────────────────────────────────────────────────────

func TestGetDepartment(t *testing.T) {
	ctx := context.Background()
	req := &pb.GetDepartmentRequest{
		InitiatorUuid: initiatorID, DepartmentUuid: deptID,
	}

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetDepartment(ctx, &pb.GetDepartmentRequest{
			InitiatorUuid:  "not-a-uuid",
			DepartmentUuid: deptID,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_department_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetDepartment(ctx, &pb.GetDepartmentRequest{
			InitiatorUuid:  initiatorID,
			DepartmentUuid: "not-a-uuid",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

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
		pg.getDepartment = func(_ context.Context, _ entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetDepartment(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("initiator not in company", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
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
			CompanyUuid: companyID, InitiatorUuid: initiatorID,
			Offset: offset, Count: count,
		}
	}

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetCompanyDepartments(ctx, &pb.GetCompanyDepartmentsRequest{
			InitiatorUuid: "not-a-uuid",
			CompanyUuid:   companyID,
			Offset:        0,
			Count:         10,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChief()
		pg.getCompanyDepartments = func(_ context.Context, _ entities.GetCompanyDepartmentsDTO) ([]*entities.Department, Error.CodeError) {
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
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
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
		pg.getCompanyDepartments = func(_ context.Context, _ entities.GetCompanyDepartmentsDTO) ([]*entities.Department, Error.CodeError) {
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
		InitiatorUuid: initiatorID, DepartmentUuid: deptID, Title: "New Name",
	}

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.UpdateDepartmentTitle(ctx, &pb.UpdateDepartmentTitleRequest{
			InitiatorUuid:  "not-a-uuid",
			DepartmentUuid: deptID,
			Title:          "New Name",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_department_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.UpdateDepartmentTitle(ctx, &pb.UpdateDepartmentTitleRequest{
			InitiatorUuid:  initiatorID,
			DepartmentUuid: "not-a-uuid",
			Title:          "New Name",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

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
			InitiatorUuid: initiatorID, DepartmentUuid: deptID, Title: "",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("department not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.UpdateDepartmentTitle(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("initiator not chief", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
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
		InitiatorUuid: initiatorID, DepartmentUuid: deptID,
	}

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.DeleteDepartment(ctx, &pb.DeleteDepartmentRequest{
			InitiatorUuid:  "not-a-uuid",
			DepartmentUuid: deptID,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_department_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.DeleteDepartment(ctx, &pb.DeleteDepartmentRequest{
			InitiatorUuid:  initiatorID,
			DepartmentUuid: "not-a-uuid",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChiefAndDept()
		pg.deleteDepartment = func(_ context.Context, _ entities.DeleteDepartmentDTO) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteDepartment(ctx, req)
		assertNoError(t, err)
	})

	t.Run("department not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteDepartment(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("initiator not chief", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			return &entities.Employee{Role: "manager"}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteDepartment(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("db delete error", func(t *testing.T) {
		pg := pgRepoWithChiefAndDept()
		pg.deleteDepartment = func(_ context.Context, _ entities.DeleteDepartmentDTO) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.DeleteDepartment(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── GetUserCompanies ─────────────────────────────────────────────────────────

func TestGetUserCompanies(t *testing.T) {
	ctx := context.Background()
	req := &pb.GetUserCompaniesRequest{InitiatorUuid: initiatorID}

	t.Run("success — returns user companies", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getUserCompanies = func(_ context.Context, _ entities.GetUserCompaniesDTO) ([]*entities.GetCompanies, Error.CodeError) {
			return []*entities.GetCompanies{
				{CompanyUUID: companyID, Title: "My Company", Status: "open"},
			}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetUserCompanies(ctx, req)
		assertNoError(t, err)
		if len(res.GetCompanies()) != 1 {
			t.Errorf("expected 1 company, got %d", len(res.GetCompanies()))
		}
		if res.GetCompanies()[0].GetCompanyUuid() != companyID {
			t.Errorf("expected company uuid %q, got %q", companyID, res.GetCompanies()[0].GetCompanyUuid())
		}
	})

	t.Run("success — empty list", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getUserCompanies = func(_ context.Context, _ entities.GetUserCompaniesDTO) ([]*entities.GetCompanies, Error.CodeError) {
			return []*entities.GetCompanies{}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		res, err := svc.GetUserCompanies(ctx, req)
		assertNoError(t, err)
		if len(res.GetCompanies()) != 0 {
			t.Errorf("expected 0 companies, got %d", len(res.GetCompanies()))
		}
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.GetUserCompanies(ctx, &pb.GetUserCompaniesRequest{
			InitiatorUuid: "not-a-uuid",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("db error", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getUserCompanies = func(_ context.Context, _ entities.GetUserCompaniesDTO) ([]*entities.GetCompanies, Error.CodeError) {
			return nil, internalErr()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.GetUserCompanies(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}

// ─── RemoveEmployeeFromDepartment ─────────────────────────────────────────────

func TestRemoveEmployeeFromDepartment(t *testing.T) {
	ctx := context.Background()
	req := &pb.RemoveEmployeeFromDepartmentRequest{
		InitiatorUuid: initiatorID, DepartmentUuid: deptID, TargetUuid: targetID,
	}

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.RemoveEmployeeFromDepartment(ctx, &pb.RemoveEmployeeFromDepartmentRequest{
			InitiatorUuid:  "not-a-uuid",
			DepartmentUuid: deptID,
			TargetUuid:     targetID,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_department_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.RemoveEmployeeFromDepartment(ctx, &pb.RemoveEmployeeFromDepartmentRequest{
			InitiatorUuid:  initiatorID,
			DepartmentUuid: "not-a-uuid",
			TargetUuid:     targetID,
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_target_uuid", func(t *testing.T) {
		svc := newTestService(emptyPGRepo(), emptyRedisRepo())
		_, err := svc.RemoveEmployeeFromDepartment(ctx, &pb.RemoveEmployeeFromDepartmentRequest{
			InitiatorUuid:  initiatorID,
			DepartmentUuid: deptID,
			TargetUuid:     "not-a-uuid",
		})
		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("success", func(t *testing.T) {
		pg := pgRepoWithChiefAndDept()
		// target is in this department
		pg.getCompanyEmployee = func(_ context.Context, dto entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			if dto.UserUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return &entities.Employee{Role: "engineer", DepartmentUUID: deptID}, ok()
		}
		pg.removeEmployeeFromDepartment = func(_ context.Context, _ entities.RemoveEmployeeFromDepartmentDTO) Error.CodeError { return ok() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveEmployeeFromDepartment(ctx, req)
		assertNoError(t, err)
	})

	t.Run("department not found", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
			return nil, notFound()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveEmployeeFromDepartment(ctx, req)
		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("initiator not chief", func(t *testing.T) {
		pg := emptyPGRepo()
		pg.getDepartment = func(_ context.Context, _ entities.GetDepartmentDTO) (*entities.Department, Error.CodeError) {
			return departmentEntity(), ok()
		}
		pg.getCompany = func(_ context.Context, _ entities.GetCompanyDTO) (*entities.Company, Error.CodeError) {
			return companyEntity(), ok()
		}
		pg.getCompanyEmployee = func(_ context.Context, _ entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			return &entities.Employee{Role: "manager"}, ok()
		}

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveEmployeeFromDepartment(ctx, req)
		assertGRPCCode(t, err, codes.PermissionDenied)
	})

	t.Run("target employee not found", func(t *testing.T) {
		pg := pgRepoWithChiefAndDept()
		pg.getCompanyEmployee = func(_ context.Context, dto entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			if dto.UserUUID == initiatorID {
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
		pg.getCompanyEmployee = func(_ context.Context, dto entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			if dto.UserUUID == initiatorID {
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
		pg.getCompanyEmployee = func(_ context.Context, dto entities.GetCompanyEmployeeDTO) (*entities.Employee, Error.CodeError) {
			if dto.UserUUID == initiatorID {
				return chiefEmployee(), ok()
			}
			return &entities.Employee{Role: "engineer", DepartmentUUID: deptID}, ok()
		}
		pg.removeEmployeeFromDepartment = func(_ context.Context, _ entities.RemoveEmployeeFromDepartmentDTO) Error.CodeError { return internalErr() }

		svc := newTestService(pg, emptyRedisRepo())
		_, err := svc.RemoveEmployeeFromDepartment(ctx, req)
		assertGRPCCode(t, err, codes.Internal)
	})
}
