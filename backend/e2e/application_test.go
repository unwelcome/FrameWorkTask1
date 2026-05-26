package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── TestCreateApplication ────────────────────────────────────────────────────

func TestCreateApplication(t *testing.T) {
	c := newClient()
	env := mustSetupAppEnv(t, c) // создаём окружение один раз на все подтесты

	t.Run("happy_path", func(t *testing.T) {
		code, body := env.Inspector.post("/api/auth/application/create", map[string]string{
			"company_uuid": env.CompanyUUID,
			"title":        "Fix the heating system",
			"description":  "The heating system on the 3rd floor stopped working.",
		})

		assert.Equal(t, http.StatusCreated, code, "body: %s", body)

		var resp createApplicationResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.NotEmpty(t, resp.ApplicationUUID, "should return a non-empty application_uuid")
	})

	t.Run("wrong_role_manager", func(t *testing.T) {
		code, body := env.Manager.post("/api/auth/application/create", map[string]string{
			"company_uuid": env.CompanyUUID,
			"title":        "Fix the heating system",
			"description":  "The heating system on the 3rd floor stopped working.",
		})

		assert.Equal(t, http.StatusForbidden, code, "manager must not be able to create applications (body: %s)", body)
	})
}

// ─── TestGetApplication ───────────────────────────────────────────────────────

func TestGetApplication(t *testing.T) {
	c := newClient()
	env := mustSetupAppEnv(t, c) // создаём окружение один раз на все подтесты

	// creator_inspector_access — инспектор читает заявку, которую сам создал (CreatedBy).
	t.Run("creator_inspector_access", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Leaking pipe", "Cold water pipe is leaking on floor 2.")

		code, body := env.Inspector.get("/api/auth/application/" + appUUID)

		assert.Equal(t, http.StatusOK, code, "creator inspector should access own application (body: %s)", body)

		var resp applicationDetailResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, appUUID, resp.Application.ApplicationUUID)
		assert.Equal(t, "created", resp.Application.Status)
	})

	// manager_access — менеджер читает заявку, которую сам назначил (ManagedBy).
	t.Run("manager_access", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Broken elevator", "Elevator in building A is out of service.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)

		code, body := env.Manager.get("/api/auth/application/" + appUUID)

		assert.Equal(t, http.StatusOK, code, "responsible manager (ManagedBy) should access application (body: %s)", body)

		var resp applicationDetailResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, env.ManagerUUID, resp.Application.ManagedBy)
	})

	// engineer_access — инженер читает заявку, назначенную ему (ExecutedBy).
	t.Run("engineer_access", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Power outage", "No electricity in rooms 201-205.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)

		code, body := env.Engineer.get("/api/auth/application/" + appUUID)

		assert.Equal(t, http.StatusOK, code, "responsible engineer (ExecutedBy) should access application (body: %s)", body)

		var resp applicationDetailResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, env.EngineerUUID, resp.Application.ExecutedBy)
	})

	// inspected_by_access — инспектор читает заявку, которую сам взял на проверку (InspectedBy).
	// Inspector2 берёт заявку на проверку → должен иметь доступ по полю InspectedBy.
	t.Run("inspected_by_access", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Faulty wiring", "Wiring in room 105 causes short circuits.")

		// Доводим заявку до статуса on_verification через Inspector2.
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")
		mustSetAppStatus(t, env.Engineer, appUUID, "pending_verification")
		mustTakeToVerification(t, env.Inspector2, appUUID)

		code, body := env.Inspector2.get("/api/auth/application/" + appUUID)

		assert.Equal(t, http.StatusOK, code, "inspector holding verification (InspectedBy) should access application (body: %s)", body)

		var resp applicationDetailResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, env.Inspector2UUID, resp.Application.InspectedBy)
		assert.Equal(t, "on_verification", resp.Application.Status)
	})

	// stranger_inspector_denied — инспектор, не связанный с заявкой, получает 403.
	t.Run("stranger_inspector_denied", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Broken window", "Window in hall B is cracked.")

		_, strangerLogin := mustRegisterAndLogin(t, c)
		stranger := c.withToken(strangerLogin.AccessToken)
		mustAddMember(t, env.Chief, stranger, env.CompanyUUID)
		mustSetEmployeeRole(t, env.Chief, env.CompanyUUID, strangerLogin.UserUUID, "inspector")

		code, body := stranger.get("/api/auth/application/" + appUUID)

		assert.Equal(t, http.StatusForbidden, code, "unrelated inspector must not access application (body: %s)", body)
	})

	// stranger_engineer_denied — инженер, которому заявка не назначена, получает 403.
	t.Run("stranger_engineer_denied", func(t *testing.T) {
		// Заявка остаётся в статусе created: executed_by не задан.
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Noisy HVAC", "HVAC unit on roof makes loud grinding noise.")

		_, strangerLogin := mustRegisterAndLogin(t, c)
		stranger := c.withToken(strangerLogin.AccessToken)
		mustAddMember(t, env.Chief, stranger, env.CompanyUUID)
		mustSetEmployeeRole(t, env.Chief, env.CompanyUUID, strangerLogin.UserUUID, "engineer")

		code, body := stranger.get("/api/auth/application/" + appUUID)

		assert.Equal(t, http.StatusForbidden, code, "unrelated engineer must not access application (body: %s)", body)
	})

	// stranger_manager_denied — менеджер, не назначавший эту заявку, получает 403.
	t.Run("stranger_manager_denied", func(t *testing.T) {
		// Заявка остаётся в статусе created: managed_by не задан.
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Roof leak", "Rain water drips through roof on 5th floor.")

		_, strangerLogin := mustRegisterAndLogin(t, c)
		stranger := c.withToken(strangerLogin.AccessToken)
		mustAddMember(t, env.Chief, stranger, env.CompanyUUID)
		mustSetEmployeeRole(t, env.Chief, env.CompanyUUID, strangerLogin.UserUUID, "manager")

		code, body := stranger.get("/api/auth/application/" + appUUID)

		assert.Equal(t, http.StatusForbidden, code, "unrelated manager must not access application (body: %s)", body)
	})

	// other_dept_inspector_denied — инспектор из другого департамента не видит заявку.
	t.Run("other_dept_inspector_denied", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"AC breakdown", "Air conditioning stopped in wing B.")

		code, body := env.Inspector3.get("/api/auth/application/" + appUUID)

		assert.Equal(t, http.StatusForbidden, code, "inspector from another dept must not access application (body: %s)", body)
	})

	// other_dept_manager_denied — менеджер из другого департамента не видит заявку.
	t.Run("other_dept_manager_denied", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Elevator noise", "Elevator in block C makes noise.")

		code, body := env.Manager2.get("/api/auth/application/" + appUUID)

		assert.Equal(t, http.StatusForbidden, code, "manager from another dept must not access application (body: %s)", body)
	})

	// other_dept_engineer_denied — инженер из другого департамента не видит заявку.
	t.Run("other_dept_engineer_denied", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Broken lock", "Entry door lock is broken.")

		code, body := env.Engineer2.get("/api/auth/application/" + appUUID)

		assert.Equal(t, http.StatusForbidden, code, "engineer from another dept must not access application (body: %s)", body)
	})

	// chief_access — chief может получить любую заявку компании.
	t.Run("chief_access", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Water pump failure", "Pump in basement stopped working.")

		code, body := env.Chief.get("/api/auth/application/" + appUUID)

		assert.Equal(t, http.StatusOK, code, "chief should access any application (body: %s)", body)

		var resp applicationDetailResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, appUUID, resp.Application.ApplicationUUID)
	})

	// analytic_access — аналитик может получить любую заявку компании.
	t.Run("analytic_access", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Gas smell", "Faint gas smell reported near boiler room.")

		code, body := env.Analytic.get("/api/auth/application/" + appUUID)

		assert.Equal(t, http.StatusOK, code, "analytic should access any application (body: %s)", body)

		var resp applicationDetailResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, appUUID, resp.Application.ApplicationUUID)
	})
}

// ─── TestGetApplications ──────────────────────────────────────────────────────

func TestGetApplications(t *testing.T) {
	c := newClient()
	env := mustSetupAppEnv(t, c) // создаём окружение один раз на все подтесты

	// chief_success — chief видит заявки с любым фильтром по статусу.
	t.Run("chief_success", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Chief list test", "Application to verify chief list visibility.")

		code, body := env.Chief.get(fmt.Sprintf(
			"/api/auth/company/%s/applications/list?count=50&offset=0&statuses=created",
			env.CompanyUUID,
		))

		assert.Equal(t, http.StatusOK, code, "body: %s", body)

		var resp applicationListResp
		require.NoError(t, json.Unmarshal(body, &resp))

		found := false
		for _, app := range resp.Applications {
			if app.ApplicationUUID == appUUID {
				found = true
				break
			}
		}
		assert.True(t, found, "created application should appear in chief's list")
	})

	// inspector_from_pool — инспектор видит заявки в статусе pending_verification своего департамента.
	t.Run("inspector_from_pool", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Pool app", "Application advanced to pending_verification.")

		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")
		mustSetAppStatus(t, env.Engineer, appUUID, "pending_verification")

		// Inspector2 смотрит пул (Inspector создал заявку — его личный список, не пул).
		code, body := env.Inspector2.get(fmt.Sprintf(
			"/api/auth/company/%s/applications/list?count=50&offset=0&from_pool=true",
			env.CompanyUUID,
		))

		assert.Equal(t, http.StatusOK, code, "body: %s", body)

		var resp applicationListResp
		require.NoError(t, json.Unmarshal(body, &resp))

		found := false
		for _, app := range resp.Applications {
			if app.ApplicationUUID == appUUID {
				found = true
				break
			}
		}
		assert.True(t, found, "pending_verification application should appear in inspector pool")
	})

	// inspector_personal_no_status — без фильтра по статусу инспектор видит созданные им заявки (по CreatedBy).
	t.Run("inspector_personal_no_status", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Inspector personal", "Application created by this inspector.")

		code, body := env.Inspector.get(fmt.Sprintf(
			"/api/auth/company/%s/applications/list?count=50&offset=0",
			env.CompanyUUID,
		))

		assert.Equal(t, http.StatusOK, code, "body: %s", body)

		var resp applicationListResp
		require.NoError(t, json.Unmarshal(body, &resp))

		found := false
		for _, app := range resp.Applications {
			if app.ApplicationUUID == appUUID {
				found = true
				break
			}
		}
		assert.True(t, found, "inspector should see applications they created when no status filter is applied")
	})

	// inspector_personal_on_verification — с фильтром on_verification инспектор видит заявки,
	// которые он взял на проверку (по InspectedBy).
	t.Run("inspector_personal_on_verification", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"On verification app", "Application being verified by Inspector2.")

		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")
		mustSetAppStatus(t, env.Engineer, appUUID, "pending_verification")
		mustTakeToVerification(t, env.Inspector2, appUUID)

		code, body := env.Inspector2.get(fmt.Sprintf(
			"/api/auth/company/%s/applications/list?count=50&offset=0&statuses=on_verification",
			env.CompanyUUID,
		))

		assert.Equal(t, http.StatusOK, code, "body: %s", body)

		var resp applicationListResp
		require.NoError(t, json.Unmarshal(body, &resp))

		found := false
		for _, app := range resp.Applications {
			if app.ApplicationUUID == appUUID {
				found = true
				break
			}
		}
		assert.True(t, found, "inspector should see application they are currently verifying (InspectedBy)")
	})

	// manager_from_pool — менеджер видит заявки с executed_by=null в своём департаменте.
	t.Run("manager_from_pool", func(t *testing.T) {
		// Заявка только что создана: executed_by=null, status=created — попадает в пул менеджеров.
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Manager pool app", "Unassigned application for manager pool.")

		code, body := env.Manager.get(fmt.Sprintf(
			"/api/auth/company/%s/applications/list?count=50&offset=0&from_pool=true",
			env.CompanyUUID,
		))

		assert.Equal(t, http.StatusOK, code, "body: %s", body)

		var resp applicationListResp
		require.NoError(t, json.Unmarshal(body, &resp))

		found := false
		for _, app := range resp.Applications {
			if app.ApplicationUUID == appUUID {
				found = true
				break
			}
		}
		assert.True(t, found, "unassigned application (executed_by=null) should appear in manager pool")
	})

	// manager_personal — менеджер видит заявки, которые он сам назначил (по ManagedBy).
	t.Run("manager_personal", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Manager personal app", "Application assigned by this manager.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)

		code, body := env.Manager.get(fmt.Sprintf(
			"/api/auth/company/%s/applications/list?count=50&offset=0",
			env.CompanyUUID,
		))

		assert.Equal(t, http.StatusOK, code, "body: %s", body)

		var resp applicationListResp
		require.NoError(t, json.Unmarshal(body, &resp))

		found := false
		for _, app := range resp.Applications {
			if app.ApplicationUUID == appUUID {
				found = true
				break
			}
		}
		assert.True(t, found, "application assigned by this manager should appear in their personal list")
	})

	// engineer_personal — инженер видит заявки, назначенные ему (по ExecutedBy).
	t.Run("engineer_personal", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Engineer personal app", "Application assigned to this engineer.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)

		// Для инженера фильтр statuses обязателен: допустимые значения — assigned, in_progress, on_hold, on_revision.
		code, body := env.Engineer.get(fmt.Sprintf(
			"/api/auth/company/%s/applications/list?count=50&offset=0&statuses=assigned",
			env.CompanyUUID,
		))

		assert.Equal(t, http.StatusOK, code, "body: %s", body)

		var resp applicationListResp
		require.NoError(t, json.Unmarshal(body, &resp))

		found := false
		for _, app := range resp.Applications {
			if app.ApplicationUUID == appUUID {
				found = true
				break
			}
		}
		assert.True(t, found, "application assigned to this engineer should appear in their personal list")
	})

	// inspector_pool_dept_isolation — инспектор из другого департамента НЕ должен видеть
	// заявки в pending_verification из чужого отдела.
	t.Run("inspector_pool_dept_isolation", func(t *testing.T) {
		// Создаём и продвигаем заявку в dept1 до pending_verification.
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Dept1 pool isolation", "Application from dept1 that dept2 must not see.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")
		mustSetAppStatus(t, env.Engineer, appUUID, "pending_verification")

		// Inspector3 из dept2 смотрит свой пул — заявки из dept1 не должны туда попадать.
		code, body := env.Inspector3.get(fmt.Sprintf(
			"/api/auth/company/%s/applications/list?count=50&offset=0&from_pool=true",
			env.CompanyUUID,
		))

		assert.Equal(t, http.StatusOK, code, "body: %s", body)

		var resp applicationListResp
		require.NoError(t, json.Unmarshal(body, &resp))

		for _, app := range resp.Applications {
			assert.NotEqual(t, appUUID, app.ApplicationUUID,
				"dept2 inspector pool must not contain applications from dept1")
		}
	})

	// manager_pool_dept_isolation — менеджер из другого департамента НЕ должен видеть
	// неназначенные заявки из чужого отдела.
	t.Run("manager_pool_dept_isolation", func(t *testing.T) {
		// Заявка создана в dept1 и не назначена (executed_by=null) — попадает в пул менеджеров dept1.
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Dept1 manager pool isolation", "Unassigned application from dept1.")

		// Manager2 из dept2 смотрит свой пул — заявка из dept1 не должна туда попадать.
		code, body := env.Manager2.get(fmt.Sprintf(
			"/api/auth/company/%s/applications/list?count=50&offset=0&from_pool=true",
			env.CompanyUUID,
		))

		assert.Equal(t, http.StatusOK, code, "body: %s", body)

		var resp applicationListResp
		require.NoError(t, json.Unmarshal(body, &resp))

		for _, app := range resp.Applications {
			assert.NotEqual(t, appUUID, app.ApplicationUUID,
				"dept2 manager pool must not contain applications from dept1")
		}
	})

	// invalid_count_zero — count=0 не допускается; валидация происходит до проверки членства в компании.
	t.Run("invalid_count_zero", func(t *testing.T) {
		code, body := env.Chief.get(fmt.Sprintf(
			"/api/auth/company/%s/applications/list?count=0&offset=0",
			env.CompanyUUID,
		))
		assert.Equal(t, http.StatusBadRequest, code, "count=0 should return 400 (body: %s)", body)
	})

	// invalid_count_above_max — count>100 не допускается.
	t.Run("invalid_count_above_max", func(t *testing.T) {
		code, body := env.Chief.get(fmt.Sprintf(
			"/api/auth/company/%s/applications/list?count=101&offset=0",
			env.CompanyUUID,
		))
		assert.Equal(t, http.StatusBadRequest, code, "count=101 should return 400 (body: %s)", body)
	})

	// negative_offset — отрицательный offset не допускается.
	t.Run("negative_offset", func(t *testing.T) {
		code, body := env.Chief.get(fmt.Sprintf(
			"/api/auth/company/%s/applications/list?count=10&offset=-1",
			env.CompanyUUID,
		))
		assert.Equal(t, http.StatusBadRequest, code, "offset=-1 should return 400 (body: %s)", body)
	})
}

// ─── TestUpdateApplicationStatus ─────────────────────────────────────────────

func TestUpdateApplicationStatus(t *testing.T) {
	c := newClient()
	env := mustSetupAppEnv(t, c)

	// ── Inspector ────────────────────────────────────────────────────────────────

	// inspector_complete — инспектор завершает заявку (on_verification → completed).
	// Проверяем: closed_at заполнен, inspected_by сброшен.
	t.Run("inspector_complete", func(t *testing.T) {
		appUUID := mustAdvanceToOnVerification(t, env, "Inspector complete test")

		code, body := env.Inspector2.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "completed",
		})
		assert.Equal(t, http.StatusOK, code, "inspector should complete application (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "completed", app.Status)
		assert.NotEmpty(t, app.ClosedAt, "closed_at must be set on completion")
		assert.Equal(t, env.Inspector2UUID, app.InspectedBy, "inspected_by must be unchanged")
		assert.Equal(t, env.ManagerUUID, app.ManagedBy, "managed_by must be unchanged")
		assert.Equal(t, env.EngineerUUID, app.ExecutedBy, "executed_by must be unchanged")
	})

	// inspector_fail — инспектор проваливает заявку (on_verification → failed).
	// Проверяем: closed_at заполнен, inspected_by сброшен.
	t.Run("inspector_fail", func(t *testing.T) {
		appUUID := mustAdvanceToOnVerification(t, env, "Inspector fail test")

		code, body := env.Inspector2.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "failed",
		})
		assert.Equal(t, http.StatusOK, code, "inspector should fail application (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "failed", app.Status)
		assert.NotEmpty(t, app.ClosedAt, "closed_at must be set on failure")
		assert.Equal(t, env.Inspector2UUID, app.InspectedBy, "inspected_by must be unchanged")
	})

	// inspector_on_revision — инспектор отправляет заявку на доработку (on_verification → on_revision).
	// Проверяем: revision_count увеличился на 1, managed_by и executed_by не сброшены.
	t.Run("inspector_on_revision", func(t *testing.T) {
		appUUID := mustAdvanceToOnVerification(t, env, "Inspector on_revision test")

		code, body := env.Inspector2.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "on_revision",
		})
		assert.Equal(t, http.StatusOK, code, "inspector should set on_revision (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "on_revision", app.Status)
		assert.Equal(t, int64(1), app.RevisionCount, "revision_count must be incremented")
		assert.Equal(t, env.ManagerUUID, app.ManagedBy, "managed_by must not be dropped on revision")
		assert.Equal(t, env.EngineerUUID, app.ExecutedBy, "executed_by must not be dropped on revision")
		assert.Empty(t, app.InspectedBy, "inspected_by must be cleared")
	})

	// inspector_not_responsible — инспектор, не взявший заявку на проверку, получает 403.
	t.Run("inspector_not_responsible", func(t *testing.T) {
		appUUID := mustAdvanceToOnVerification(t, env, "Inspector not responsible test")

		// Inspector (not Inspector2) tries to close the application.
		code, body := env.Inspector.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "completed",
		})
		assert.Equal(t, http.StatusForbidden, code, "non-responsible inspector must be denied (body: %s)", body)
	})

	// inspector_target_status_denied — инспектор пытается выставить недопустимый для него статус (rejected).
	t.Run("inspector_target_status_denied", func(t *testing.T) {
		appUUID := mustAdvanceToOnVerification(t, env, "Inspector target status denied test")

		code, body := env.Inspector2.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "rejected",
		})
		assert.Equal(t, http.StatusForbidden, code, "inspector must not set rejected (body: %s)", body)
	})

	// ── Manager ──────────────────────────────────────────────────────────────────

	// manager_reject_created — менеджер отклоняет заявку из статуса created.
	// Проверяем: managed_by устанавливается как инициатор.
	t.Run("manager_reject_created", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Manager reject created test", "Reject from created.")

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "rejected",
		})
		assert.Equal(t, http.StatusOK, code, "manager should reject from created (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "rejected", app.Status)
		assert.Equal(t, env.ManagerUUID, app.ManagedBy, "managed_by must be set to initiator on rejection")
	})

	// manager_reject_recalled — менеджер отклоняет заявку из статуса recalled.
	t.Run("manager_reject_recalled", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Manager reject recalled test", "Reject from recalled.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustRecallApplication(t, env.Manager, appUUID)

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "rejected",
		})
		assert.Equal(t, http.StatusOK, code, "manager should reject from recalled (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "rejected", app.Status)
	})

	// manager_reject_on_revision — менеджер отклоняет заявку из статуса on_revision.
	t.Run("manager_reject_on_revision", func(t *testing.T) {
		appUUID := mustAdvanceToOnRevision(t, env, "Manager reject on_revision test")

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "rejected",
		})
		assert.Equal(t, http.StatusOK, code, "manager should reject from on_revision (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "rejected", app.Status)
	})

	// manager_reject_after_redirect — заявка перенаправлена в dept2; менеджер из dept2 её отклоняет.
	t.Run("manager_reject_after_redirect", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Manager reject after redirect test", "Redirect then reject.")
		mustRedirectApplication(t, env.Manager, appUUID, env.Dept2UUID)

		code, body := env.Manager2.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "rejected",
		})
		assert.Equal(t, http.StatusOK, code, "dept2 manager should reject redirected application (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "rejected", app.Status)
	})

	// manager_wrong_dept — менеджер из другого департамента не может изменить статус.
	t.Run("manager_wrong_dept", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Manager wrong dept test", "Dept2 manager should be denied.")

		// Manager2 is from dept2, application is in dept1.
		code, body := env.Manager2.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "rejected",
		})
		assert.Equal(t, http.StatusForbidden, code, "manager from another dept must be denied (body: %s)", body)
	})

	// manager_wrong_current_status — менеджер пытается отклонить заявку в статусе assigned (недопустимый переход).
	t.Run("manager_wrong_current_status", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Manager wrong current status test", "Should return 412.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		// Status is now "assigned" — not in {created, redirected, recalled, on_revision}.

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "rejected",
		})
		assert.Equal(t, http.StatusPreconditionFailed, code, "manager reject from assigned must return 412 (body: %s)", body)
	})

	// ── Engineer ─────────────────────────────────────────────────────────────────

	// engineer_start_progress — инженер начинает работу (assigned → in_progress).
	t.Run("engineer_start_progress", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Engineer start progress test", "Assign then start.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)

		code, body := env.Engineer.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "in_progress",
		})
		assert.Equal(t, http.StatusOK, code, "engineer should start progress (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "in_progress", app.Status)
	})

	// engineer_on_hold — инженер приостанавливает работу (in_progress → on_hold).
	t.Run("engineer_on_hold", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Engineer on_hold test", "in_progress then on_hold.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")

		code, body := env.Engineer.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "on_hold",
		})
		assert.Equal(t, http.StatusOK, code, "engineer should pause to on_hold (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "on_hold", app.Status)
	})

	// engineer_pending_verification — инженер передаёт заявку на проверку (in_progress → pending_verification).
	t.Run("engineer_pending_verification", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Engineer pending verification test", "in_progress then pending_verification.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")

		code, body := env.Engineer.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "pending_verification",
		})
		assert.Equal(t, http.StatusOK, code, "engineer should submit for verification (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "pending_verification", app.Status)
	})

	// engineer_resume_from_revision — инженер возобновляет работу после доработки (on_revision → in_progress).
	// Проверяем, что executed_by не был сброшен и инженер может продолжить работу.
	t.Run("engineer_resume_from_revision", func(t *testing.T) {
		appUUID := mustAdvanceToOnRevision(t, env, "Engineer resume from revision test")

		code, body := env.Engineer.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "in_progress",
		})
		assert.Equal(t, http.StatusOK, code, "engineer should resume from on_revision (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "in_progress", app.Status)
		assert.Equal(t, env.EngineerUUID, app.ExecutedBy, "executed_by must remain the same engineer after revision")
	})

	// engineer_not_responsible — инженер, которому заявка не назначена, получает 403.
	t.Run("engineer_not_responsible", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Engineer not responsible test", "Engineer2 should be denied.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		// Engineer2 is not ExecutedBy for this application.

		code, body := env.Engineer2.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "in_progress",
		})
		assert.Equal(t, http.StatusForbidden, code, "non-responsible engineer must be denied (body: %s)", body)
	})

	// engineer_wrong_current_status — инженер пытается изменить статус из pending_verification (недопустимый переход).
	t.Run("engineer_wrong_current_status", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Engineer wrong current status test", "pending_verification should return 412.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")
		mustSetAppStatus(t, env.Engineer, appUUID, "pending_verification")
		// Status is now "pending_verification" — not in engineer's allowed source statuses.

		code, body := env.Engineer.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "in_progress",
		})
		assert.Equal(t, http.StatusPreconditionFailed, code, "engineer from pending_verification must get 412 (body: %s)", body)
	})

	// ── Role denied ───────────────────────────────────────────────────────────────

	// chief_denied — chief не может изменять статус заявок.
	t.Run("chief_denied", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Chief denied test", "Chief must not update status.")

		code, body := env.Chief.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "rejected",
		})
		assert.Equal(t, http.StatusForbidden, code, "chief must not update application status (body: %s)", body)
	})

	// analytic_denied — аналитик не может изменять статус заявок.
	t.Run("analytic_denied", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Analytic denied test", "Analytic must not update status.")

		code, body := env.Analytic.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "rejected",
		})
		assert.Equal(t, http.StatusForbidden, code, "analytic must not update application status (body: %s)", body)
	})

	// ── Validation ────────────────────────────────────────────────────────────────

	// invalid_status_value — статус "created" не входит в допустимое множество для этого метода.
	t.Run("invalid_status_value", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Invalid status value test", "Status created is not settable via this endpoint.")

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
			"status": "created",
		})
		assert.Equal(t, http.StatusBadRequest, code, "status=created must return 400 (body: %s)", body)
	})

	// not_found — заявка с несуществующим UUID должна вернуть 404.
	t.Run("not_found", func(t *testing.T) {
		code, body := env.Manager.patch("/api/auth/application/00000000-0000-0000-0000-000000000000/status", map[string]string{
			"status": "rejected",
		})
		assert.Equal(t, http.StatusNotFound, code, "non-existent application must return 404 (body: %s)", body)
	})
}
