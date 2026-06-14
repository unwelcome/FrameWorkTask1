#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/.."
E2E_DIR="$BACKEND_DIR/e2e"
LOGS_DIR="$SCRIPT_DIR/../../logs/tests"

# Clear logs from the previous run
rm -f "$LOGS_DIR"/*.log 2>/dev/null || true

# Services that have e2e tests
E2E_SERVICES=("auth" "company" "application")

# Map service name → Go -run pattern
get_pattern() {
  case "$1" in
    auth)
      echo "^(TestRegister|TestLogin|TestRefreshToken|TestGetUser|TestUpdateUserBio|TestChangePassword|TestGetAllActiveSessions|TestRevokeSession|TestRevokeAllSessions|TestDeleteUser|TestRestoreAccount|TestAuthFullFlow|TestVerifyAccount|TestResendVerificationCode|TestForgotPassword|TestResetPassword|TestVerify2FA|TestUpdateUser2FA)"
      ;;
    company)
      echo "^(TestCreateCompany|TestGetCompany|TestGetCompaniesList|TestGetMyCompanies|TestUpdateCompanyTitle|TestUpdateCompanyStatus|TestDeleteCompany|TestCreateJoinCode|TestGetJoinCodes|TestJoinCompany|TestDeleteJoinCode|TestCompanyFullWorkflow|TestCreateDepartment|TestGetDepartment|TestGetCompanyDepartments|TestUpdateDepartmentTitle|TestDeleteDepartment|TestAddEmployeeToDepartment|TestRemoveEmployeeFromDepartment|TestDepartmentFullWorkflow|TestGetCompanyEmployee|TestGetCompanyEmployees|TestGetCompanyEmployeesSummary|TestUpdateEmployeeRole|TestRemoveCompanyEmployee|TestEmployeeFullWorkflow)"
      ;;
    application)
      echo "^(TestCreateApplication|TestGetApplication|TestGetApplications|TestUpdateApplicationStatus|TestAssignApplication|TestRedirectApplication|TestRecallApplication|TestTakeApplicationToVerification|TestReleaseApplicationVerification|TestAddApplicationFixLog|TestDeleteApplication|TestGetApplicationHistory)"
      ;;
    *)
      echo "^Test"
      ;;
  esac
}

run_e2e() {
  local service=$1

  # Validate service name
  local supported=false
  for s in "${E2E_SERVICES[@]}"; do
    [ "$s" = "$service" ] && supported=true && break
  done

  if [ "$supported" = false ]; then
    echo "ERROR: no e2e tests for service '$service'"
    echo "Available: ${E2E_SERVICES[*]}"
    exit 1
  fi

  echo ""
  echo "=== E2E tests: $service ==="
  cd "$E2E_DIR"
  go test -v -timeout 10m -count=1 -run "$(get_pattern "$service")" .
  echo "=== $service e2e: OK ==="
}

run_all() {
  cd "$E2E_DIR"
  echo ""
  echo "=== E2E tests: all services ==="
  go test -v -timeout 10m -count=1 .
  echo ""
  echo "=== All e2e tests passed ==="
}

# Usage info
usage() {
  echo "Usage: $0 [service]"
  echo ""
  echo "  $0              — run all e2e tests"
  echo "  $0 auth         — run auth service tests only"
  echo "  $0 company      — run company service tests only"
  echo "  $0 application  — run application service tests only"
  echo ""
  echo "Available services: ${E2E_SERVICES[*]}"
}

case "$1" in
  -h|--help) usage ;;
  "")        run_all ;;
  *)         run_e2e "$1" ;;
esac
