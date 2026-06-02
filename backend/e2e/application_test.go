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

// ─── TestAssignApplication ────────────────────────────────────────────────────

func TestAssignApplication(t *testing.T) {
	c := newClient()
	env := mustSetupAppEnv(t, c)

	// happy_path — менеджер назначает инженера; статус становится assigned,
	// executed_by и managed_by выставляются корректно.
	t.Run("happy_path", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Assign happy path", "Application for assign test.")

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/assign", map[string]string{
			"target_uuid": env.EngineerUUID,
		})
		assert.Equal(t, http.StatusOK, code, "manager should assign application (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "assigned", app.Status)
		assert.Equal(t, env.EngineerUUID, app.ExecutedBy, "executed_by must be set to target engineer")
		assert.Equal(t, env.ManagerUUID, app.ManagedBy, "managed_by must be set to initiator manager")
	})

	// from_recalled — менеджер повторно назначает инженера после отзыва заявки.
	t.Run("from_recalled", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Assign from recalled", "Application recalled before reassign.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustRecallApplication(t, env.Manager, appUUID)

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/assign", map[string]string{
			"target_uuid": env.EngineerUUID,
		})
		assert.Equal(t, http.StatusOK, code, "manager should reassign after recall (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "assigned", app.Status)
		assert.Equal(t, env.EngineerUUID, app.ExecutedBy, "executed_by must be set after reassign")
	})

	// wrong_role — не-менеджер (инспектор) пытается назначить инженера → 403.
	t.Run("wrong_role", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Assign wrong role", "Inspector must not assign.")

		code, body := env.Inspector.patch("/api/auth/application/"+appUUID+"/assign", map[string]string{
			"target_uuid": env.EngineerUUID,
		})
		assert.Equal(t, http.StatusForbidden, code, "non-manager must not assign application (body: %s)", body)
	})

	// wrong_dept_manager — менеджер из другого департамента → 403.
	t.Run("wrong_dept_manager", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Assign wrong dept", "Dept2 manager must not assign dept1 application.")

		code, body := env.Manager2.patch("/api/auth/application/"+appUUID+"/assign", map[string]string{
			"target_uuid": env.EngineerUUID,
		})
		assert.Equal(t, http.StatusForbidden, code, "manager from another dept must not assign (body: %s)", body)
	})

	// engineer_from_wrong_dept — менеджер пытается назначить инженера из другого департамента → 400.
	t.Run("engineer_from_wrong_dept", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Assign engineer wrong dept", "Engineer from dept2 must not be assignable to dept1 application.")

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/assign", map[string]string{
			"target_uuid": env.Engineer2UUID,
		})
		assert.Equal(t, http.StatusBadRequest, code, "engineer from wrong dept must return 400 (body: %s)", body)
	})

	// target_not_engineer — менеджер пытается назначить не-инженера (инспектора) → 400.
	t.Run("target_not_engineer", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Assign target not engineer", "Inspector UUID as target must return 400.")

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/assign", map[string]string{
			"target_uuid": env.InspectorUUID,
		})
		assert.Equal(t, http.StatusBadRequest, code, "assigning non-engineer must return 400 (body: %s)", body)
	})

	// wrong_status — заявка в статусе in_progress не может быть назначена → 400.
	t.Run("wrong_status", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Assign wrong status", "In-progress application must not be assignable.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/assign", map[string]string{
			"target_uuid": env.EngineerUUID,
		})
		assert.Equal(t, http.StatusBadRequest, code, "in_progress application must not be assignable (body: %s)", body)
	})

	// not_found — заявка с несуществующим UUID → 404.
	t.Run("not_found", func(t *testing.T) {
		code, body := env.Manager.patch("/api/auth/application/00000000-0000-0000-0000-000000000000/assign", map[string]string{
			"target_uuid": env.EngineerUUID,
		})
		assert.Equal(t, http.StatusNotFound, code, "non-existent application must return 404 (body: %s)", body)
	})
}

// ─── TestRedirectApplication ──────────────────────────────────────────────────

func TestRedirectApplication(t *testing.T) {
	c := newClient()
	env := mustSetupAppEnv(t, c)

	// happy_path — менеджер dept1 перенаправляет заявку в dept2.
	// Проверяем: статус redirected, department_uuid = dept2, fix_log записан.
	t.Run("happy_path", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Redirect happy path", "Application to be redirected to dept2.")

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/redirect", map[string]string{
			"target_department_uuid": env.Dept2UUID,
			"message":                "Redirected: not our area of responsibility.",
		})
		assert.Equal(t, http.StatusOK, code, "manager should redirect application (body: %s)", body)

		// Inspector созданной заявки по-прежнему имеет доступ (поле created_by).
		app := mustGetApplicationDetail(t, env.Inspector, appUUID)
		assert.Equal(t, "redirected", app.Status)
		assert.Equal(t, env.Dept2UUID, app.DepartmentUUID, "department_uuid must switch to target dept after redirect")
		require.Len(t, app.FixLogs, 1, "redirect must write exactly one fix log entry")
		assert.Equal(t, "Redirected: not our area of responsibility.", app.FixLogs[0].Text)
	})

	// from_recalled — менеджер перенаправляет заявку, которую предварительно отозвал.
	t.Run("from_recalled", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Redirect from recalled", "Recalled application can be redirected.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustRecallApplication(t, env.Manager, appUUID)

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/redirect", map[string]string{
			"target_department_uuid": env.Dept2UUID,
			"message":                "Redirecting after recall.",
		})
		assert.Equal(t, http.StatusOK, code, "manager should redirect recalled application (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Inspector, appUUID)
		assert.Equal(t, "redirected", app.Status)
		assert.Equal(t, env.Dept2UUID, app.DepartmentUUID, "department_uuid must switch to target dept")
	})

	// wrong_role — не-менеджер (инспектор) пытается перенаправить → 403.
	t.Run("wrong_role", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Redirect wrong role", "Inspector must not redirect.")

		code, body := env.Inspector.patch("/api/auth/application/"+appUUID+"/redirect", map[string]string{
			"target_department_uuid": env.Dept2UUID,
			"message":                "Trying to redirect.",
		})
		assert.Equal(t, http.StatusForbidden, code, "non-manager must not redirect application (body: %s)", body)
	})

	// wrong_dept_manager — менеджер из другого департамента пытается перенаправить → 403.
	t.Run("wrong_dept_manager", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Redirect wrong dept", "Dept2 manager must not redirect dept1 application.")

		code, body := env.Manager2.patch("/api/auth/application/"+appUUID+"/redirect", map[string]string{
			"target_department_uuid": env.Dept2UUID,
			"message":                "Attempting cross-dept redirect.",
		})
		assert.Equal(t, http.StatusForbidden, code, "manager from wrong dept must not redirect (body: %s)", body)
	})

	// wrong_status — заявка в статусе assigned не может быть перенаправлена → 403.
	t.Run("wrong_status", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Redirect wrong status", "Assigned application must not be redirectable.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/redirect", map[string]string{
			"target_department_uuid": env.Dept2UUID,
			"message":                "Redirect from assigned.",
		})
		assert.Equal(t, http.StatusForbidden, code, "assigned application must not be redirectable (body: %s)", body)
	})

	// empty_message — пустое поле message → 400.
	t.Run("empty_message", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Redirect empty message", "Message must not be empty.")

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/redirect", map[string]string{
			"target_department_uuid": env.Dept2UUID,
			"message":                "",
		})
		assert.Equal(t, http.StatusBadRequest, code, "empty message must return 400 (body: %s)", body)
	})

	// not_found — заявка с несуществующим UUID → 404.
	t.Run("not_found", func(t *testing.T) {
		code, body := env.Manager.patch("/api/auth/application/00000000-0000-0000-0000-000000000000/redirect", map[string]string{
			"target_department_uuid": env.Dept2UUID,
			"message":                "Redirect non-existent.",
		})
		assert.Equal(t, http.StatusNotFound, code, "non-existent application must return 404 (body: %s)", body)
	})
}

// ─── TestRecallApplication ────────────────────────────────────────────────────

func TestRecallApplication(t *testing.T) {
	c := newClient()
	env := mustSetupAppEnv(t, c)

	// from_assigned — менеджер отзывает назначенную заявку.
	// Проверяем: статус recalled, executed_by очищен, fix_log записан.
	t.Run("from_assigned", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Recall from assigned", "Application to be recalled from assigned.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/recall", map[string]string{
			"message": "Recalled: requirements changed.",
		})
		assert.Equal(t, http.StatusOK, code, "manager should recall assigned application (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "recalled", app.Status)
		assert.Empty(t, app.ExecutedBy, "executed_by must be cleared after recall")
		require.Len(t, app.FixLogs, 1, "recall must write exactly one fix log entry")
		assert.Equal(t, "Recalled: requirements changed.", app.FixLogs[0].Text)
	})

	// from_in_progress — менеджер отзывает заявку, которую инженер уже взял в работу.
	t.Run("from_in_progress", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Recall from in_progress", "Application to be recalled from in_progress.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/recall", map[string]string{
			"message": "Recalled while in progress.",
		})
		assert.Equal(t, http.StatusOK, code, "manager should recall in_progress application (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "recalled", app.Status)
		assert.Empty(t, app.ExecutedBy, "executed_by must be cleared after recall")
	})

	// from_on_hold — менеджер отзывает приостановленную заявку.
	t.Run("from_on_hold", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Recall from on_hold", "Application to be recalled from on_hold.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")
		mustSetAppStatus(t, env.Engineer, appUUID, "on_hold")

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/recall", map[string]string{
			"message": "Recalled while on hold.",
		})
		assert.Equal(t, http.StatusOK, code, "manager should recall on_hold application (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "recalled", app.Status)
		assert.Empty(t, app.ExecutedBy, "executed_by must be cleared after recall")
	})

	// wrong_manager_not_responsible — другой менеджер (не ManagedBy) пытается отозвать → 403.
	t.Run("wrong_manager_not_responsible", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Recall wrong manager", "Only responsible manager can recall.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)

		code, body := env.Manager2.patch("/api/auth/application/"+appUUID+"/recall", map[string]string{
			"message": "Trying to recall by wrong manager.",
		})
		assert.Equal(t, http.StatusForbidden, code, "non-responsible manager must not recall (body: %s)", body)
	})

	// wrong_role — не-менеджер (инспектор) пытается отозвать заявку → 403.
	t.Run("wrong_role", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Recall wrong role", "Inspector must not recall.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)

		code, body := env.Inspector.patch("/api/auth/application/"+appUUID+"/recall", map[string]string{
			"message": "Inspector trying to recall.",
		})
		assert.Equal(t, http.StatusForbidden, code, "non-manager must not recall (body: %s)", body)
	})

	// wrong_status — заявка в статусе created не может быть отозвана → 403.
	t.Run("wrong_status", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Recall wrong status", "Created application must not be recallable.")

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/recall", map[string]string{
			"message": "Trying to recall created application.",
		})
		assert.Equal(t, http.StatusForbidden, code, "created application must not be recallable (body: %s)", body)
	})

	// empty_message — пустое поле message → 400.
	t.Run("empty_message", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Recall empty message", "Message must not be empty.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/recall", map[string]string{
			"message": "",
		})
		assert.Equal(t, http.StatusBadRequest, code, "empty message must return 400 (body: %s)", body)
	})

	// not_found — заявка с несуществующим UUID → 404.
	t.Run("not_found", func(t *testing.T) {
		code, body := env.Manager.patch("/api/auth/application/00000000-0000-0000-0000-000000000000/recall", map[string]string{
			"message": "Recall non-existent.",
		})
		assert.Equal(t, http.StatusNotFound, code, "non-existent application must return 404 (body: %s)", body)
	})
}

// ─── TestTakeApplicationToVerification ───────────────────────────────────────

func TestTakeApplicationToVerification(t *testing.T) {
	c := newClient()
	env := mustSetupAppEnv(t, c)

	// happy_path — инспектор берёт заявку на проверку.
	// pending_verification → on_verification, inspected_by = inspector2UUID.
	t.Run("happy_path", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Take verification happy path", "Application to take to verification.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")
		mustSetAppStatus(t, env.Engineer, appUUID, "pending_verification")

		code, body := env.Inspector2.patch("/api/auth/application/"+appUUID+"/take-verification", nil)
		assert.Equal(t, http.StatusOK, code, "inspector should take application to verification (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "on_verification", app.Status)
		assert.Equal(t, env.Inspector2UUID, app.InspectedBy, "inspected_by must be set to initiator inspector")
	})

	// wrong_dept_inspector — инспектор из другого департамента → 403.
	t.Run("wrong_dept_inspector", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Take verification wrong dept", "Dept2 inspector must not take dept1 application.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")
		mustSetAppStatus(t, env.Engineer, appUUID, "pending_verification")

		code, body := env.Inspector3.patch("/api/auth/application/"+appUUID+"/take-verification", nil)
		assert.Equal(t, http.StatusForbidden, code, "inspector from wrong dept must not take (body: %s)", body)
	})

	// wrong_role — не-инспектор (менеджер) пытается взять заявку на проверку → 403.
	t.Run("wrong_role", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Take verification wrong role", "Manager must not take application to verification.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")
		mustSetAppStatus(t, env.Engineer, appUUID, "pending_verification")

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/take-verification", nil)
		assert.Equal(t, http.StatusForbidden, code, "non-inspector must not take application to verification (body: %s)", body)
	})

	// wrong_status — заявка не в статусе pending_verification → 403.
	t.Run("wrong_status", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Take verification wrong status", "Assigned application must not be takeable.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		// Status is "assigned", not "pending_verification".

		code, body := env.Inspector2.patch("/api/auth/application/"+appUUID+"/take-verification", nil)
		assert.Equal(t, http.StatusForbidden, code, "non-pending_verification application must not be takeable (body: %s)", body)
	})

	// not_found — заявка с несуществующим UUID → 404.
	t.Run("not_found", func(t *testing.T) {
		code, body := env.Inspector2.patch("/api/auth/application/00000000-0000-0000-0000-000000000000/take-verification", nil)
		assert.Equal(t, http.StatusNotFound, code, "non-existent application must return 404 (body: %s)", body)
	})
}

// ─── TestReleaseApplicationVerification ──────────────────────────────────────

func TestReleaseApplicationVerification(t *testing.T) {
	c := newClient()
	env := mustSetupAppEnv(t, c)

	// happy_path — инспектор отпускает заявку (on_verification → pending_verification).
	// Проверяем: статус pending_verification, inspected_by очищен, fix_log записан.
	t.Run("happy_path", func(t *testing.T) {
		appUUID := mustAdvanceToOnVerification(t, env, "Release verification happy path")

		code, body := env.Inspector2.patch("/api/auth/application/"+appUUID+"/release-verification", map[string]string{
			"message": "Releasing for another inspector to take.",
		})
		assert.Equal(t, http.StatusOK, code, "inspector should release verification (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		assert.Equal(t, "pending_verification", app.Status)
		assert.Empty(t, app.InspectedBy, "inspected_by must be cleared after release")
		require.Len(t, app.FixLogs, 1, "release must write exactly one fix log entry")
		assert.Equal(t, "Releasing for another inspector to take.", app.FixLogs[0].Text)
	})

	// wrong_inspector_not_responsible — инспектор, не взявший заявку (не InspectedBy), пытается отпустить → 403.
	t.Run("wrong_inspector_not_responsible", func(t *testing.T) {
		appUUID := mustAdvanceToOnVerification(t, env, "Release verification wrong inspector")
		// InspectedBy = Inspector2; Inspector is a different inspector in the same dept.

		code, body := env.Inspector.patch("/api/auth/application/"+appUUID+"/release-verification", map[string]string{
			"message": "Trying to release by wrong inspector.",
		})
		assert.Equal(t, http.StatusForbidden, code, "non-responsible inspector must not release (body: %s)", body)
	})

	// wrong_role — не-инспектор (менеджер) пытается отпустить заявку → 403.
	t.Run("wrong_role", func(t *testing.T) {
		appUUID := mustAdvanceToOnVerification(t, env, "Release verification wrong role")

		code, body := env.Manager.patch("/api/auth/application/"+appUUID+"/release-verification", map[string]string{
			"message": "Manager trying to release.",
		})
		assert.Equal(t, http.StatusForbidden, code, "non-inspector must not release verification (body: %s)", body)
	})

	// wrong_status — заявка не в статусе on_verification → 403.
	t.Run("wrong_status", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Release wrong status", "Pending verification application must not be releaseable.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")
		mustSetAppStatus(t, env.Engineer, appUUID, "pending_verification")
		// Status is "pending_verification", not "on_verification".

		code, body := env.Inspector2.patch("/api/auth/application/"+appUUID+"/release-verification", map[string]string{
			"message": "Trying to release pending_verification.",
		})
		assert.Equal(t, http.StatusForbidden, code, "non-on_verification application must not be releaseable (body: %s)", body)
	})

	// empty_message — пустое поле message → 400.
	t.Run("empty_message", func(t *testing.T) {
		appUUID := mustAdvanceToOnVerification(t, env, "Release verification empty message")

		code, body := env.Inspector2.patch("/api/auth/application/"+appUUID+"/release-verification", map[string]string{
			"message": "",
		})
		assert.Equal(t, http.StatusBadRequest, code, "empty message must return 400 (body: %s)", body)
	})

	// not_found — заявка с несуществующим UUID → 404.
	t.Run("not_found", func(t *testing.T) {
		code, body := env.Inspector2.patch("/api/auth/application/00000000-0000-0000-0000-000000000000/release-verification", map[string]string{
			"message": "Release non-existent.",
		})
		assert.Equal(t, http.StatusNotFound, code, "non-existent application must return 404 (body: %s)", body)
	})
}

// ─── TestAddApplicationFixLog ─────────────────────────────────────────────────

func TestAddApplicationFixLog(t *testing.T) {
	c := newClient()
	env := mustSetupAppEnv(t, c)

	// happy_path — ответственный инженер добавляет fix log к заявке в статусе in_progress.
	// Проверяем: 201, в FixLogs одна запись с правильным текстом и created_by = engineerUUID.
	t.Run("happy_path", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Fix log happy path", "Application for fix log test.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")

		code, body := env.Engineer.post("/api/auth/application/"+appUUID+"/fix-log", map[string]string{
			"message": "Replaced the broken valve.",
		})
		assert.Equal(t, http.StatusCreated, code, "engineer should add fix log (body: %s)", body)

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		require.Len(t, app.FixLogs, 1, "should have exactly one fix log entry")
		assert.Equal(t, "Replaced the broken valve.", app.FixLogs[0].Text)
		assert.Equal(t, env.EngineerUUID, app.FixLogs[0].CreatedBy, "fix log must be attributed to the engineer")
	})

	// multiple_entries — инженер добавляет две записи подряд.
	// Проверяем: FixLogs содержит обе записи независимо от их порядка.
	t.Run("multiple_entries", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Fix log multiple entries", "Application for multiple fix log entries.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")

		mustAddFixLog(t, env.Engineer, appUUID, "Step 1: drained the pipe.")
		mustAddFixLog(t, env.Engineer, appUUID, "Step 2: replaced the gasket.")

		app := mustGetApplicationDetail(t, env.Chief, appUUID)
		require.Len(t, app.FixLogs, 2, "should have exactly two fix log entries")

		texts := []string{app.FixLogs[0].Text, app.FixLogs[1].Text}
		assert.Contains(t, texts, "Step 1: drained the pipe.")
		assert.Contains(t, texts, "Step 2: replaced the gasket.")
	})

	// not_responsible_engineer — Engineer2 (не ExecutedBy) пытается добавить запись → 403.
	t.Run("not_responsible_engineer", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Fix log not responsible", "Only ExecutedBy engineer may add fix logs.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")

		code, body := env.Engineer2.post("/api/auth/application/"+appUUID+"/fix-log", map[string]string{
			"message": "Engineer2 trying to add fix log.",
		})
		assert.Equal(t, http.StatusForbidden, code, "non-responsible engineer must not add fix log (body: %s)", body)
	})

	// wrong_role — не-инженер (менеджер) пытается добавить запись → 403.
	t.Run("wrong_role", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Fix log wrong role", "Manager must not add fix logs.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")

		code, body := env.Manager.post("/api/auth/application/"+appUUID+"/fix-log", map[string]string{
			"message": "Manager trying to add fix log.",
		})
		assert.Equal(t, http.StatusForbidden, code, "non-engineer must not add fix log (body: %s)", body)
	})

	// empty_message — пустое поле message → 400.
	t.Run("empty_message", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Fix log empty message", "Message must not be empty.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")

		code, body := env.Engineer.post("/api/auth/application/"+appUUID+"/fix-log", map[string]string{
			"message": "",
		})
		assert.Equal(t, http.StatusBadRequest, code, "empty message must return 400 (body: %s)", body)
	})

	// not_found — заявка с несуществующим UUID → 404.
	t.Run("not_found", func(t *testing.T) {
		code, body := env.Engineer.post("/api/auth/application/00000000-0000-0000-0000-000000000000/fix-log", map[string]string{
			"message": "Fix log for non-existent application.",
		})
		assert.Equal(t, http.StatusNotFound, code, "non-existent application must return 404 (body: %s)", body)
	})
}

// ─── TestDeleteApplication ────────────────────────────────────────────────────

func TestDeleteApplication(t *testing.T) {
	c := newClient()
	env := mustSetupAppEnv(t, c)

	// happy_path — инспектор (CreatedBy) удаляет свою заявку в статусе created.
	// Проверяем: 200; после удаления GET возвращает заявку с заполненными deleted_at и deleted_by,
	// а в FixLogs присутствует запись с текстом удаления.
	t.Run("happy_path", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Delete happy path", "Application to be soft-deleted.")

		code, body := env.Inspector.delete("/api/auth/application/"+appUUID, map[string]string{
			"message": "Application created by mistake.",
		})
		assert.Equal(t, http.StatusOK, code, "inspector should delete own application (body: %s)", body)

		// Soft delete: заявка доступна через GetApplication, поле deleted_at заполнено.
		app := mustGetApplicationDetail(t, env.Inspector, appUUID)
		assert.NotEmpty(t, app.DeletedAt, "deleted_at must be set after deletion")
		assert.Equal(t, env.InspectorUUID, app.DeletedBy, "deleted_by must be set to the initiator inspector")
		require.Len(t, app.FixLogs, 1, "delete must write exactly one fix log entry")
		assert.Equal(t, "Application created by mistake.", app.FixLogs[0].Text)
	})

	// wrong_creator — другой инспектор (не CreatedBy) пытается удалить → 403.
	t.Run("wrong_creator", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Delete wrong creator", "Only the creator may delete.")

		code, body := env.Inspector2.delete("/api/auth/application/"+appUUID, map[string]string{
			"message": "Inspector2 trying to delete.",
		})
		assert.Equal(t, http.StatusForbidden, code, "non-creator must not delete application (body: %s)", body)
	})

	// wrong_status — заявка уже назначена (статус assigned); создатель пытается удалить → 403.
	t.Run("wrong_status", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Delete wrong status", "Assigned application must not be deletable.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		// Status is now "assigned" — only "created" is deletable.

		code, body := env.Inspector.delete("/api/auth/application/"+appUUID, map[string]string{
			"message": "Trying to delete assigned application.",
		})
		assert.Equal(t, http.StatusForbidden, code, "non-created application must not be deletable (body: %s)", body)
	})

	// empty_message — пустое поле message → 400.
	t.Run("empty_message", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"Delete empty message", "Message must not be empty.")

		code, body := env.Inspector.delete("/api/auth/application/"+appUUID, map[string]string{
			"message": "",
		})
		assert.Equal(t, http.StatusBadRequest, code, "empty message must return 400 (body: %s)", body)
	})

	// not_found — заявка с несуществующим UUID → 404.
	t.Run("not_found", func(t *testing.T) {
		code, body := env.Inspector.delete("/api/auth/application/00000000-0000-0000-0000-000000000000", map[string]string{
			"message": "Delete non-existent.",
		})
		assert.Equal(t, http.StatusNotFound, code, "non-existent application must return 404 (body: %s)", body)
	})
}

// ─── TestGetApplicationHistory ────────────────────────────────────────────────

func TestGetApplicationHistory(t *testing.T) {
	c := newClient()
	env := mustSetupAppEnv(t, c)

	// happy_path — создаём заявку и прогоняем два перехода состояния.
	// saveVersion вызывается перед каждой мутацией, поэтому:
	//   assign       → snapshot(v=1, status=created)
	//   in_progress  → snapshot(v=2, status=assigned)
	// История отсортирована по version DESC.
	t.Run("happy_path", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"History happy path", "Check history order.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
		mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")

		code, resp := getApplicationHistory(t, env.Inspector, appUUID, 10, 0)
		assert.Equal(t, http.StatusOK, code)
		require.Len(t, resp.History, 2)

		// Первая запись — самый свежий снапшот (перед in_progress)
		assert.Equal(t, int64(2), resp.History[0].Version)
		assert.Equal(t, "assigned", resp.History[0].Status)
		assert.Equal(t, appUUID, resp.History[0].ApplicationUUID)

		// Вторая запись — снапшот перед assign
		assert.Equal(t, int64(1), resp.History[1].Version)
		assert.Equal(t, "created", resp.History[1].Status)
		assert.Equal(t, appUUID, resp.History[1].ApplicationUUID)
	})

	// pagination — 4 перехода = 4 записи (версии 1..4).
	// mustAdvanceToOnVerification: create→assign→in_progress→pending_verification→on_verification
	// Снапшоты (version, status): (1,created), (2,assigned), (3,in_progress), (4,pending_verification)
	t.Run("pagination", func(t *testing.T) {
		appUUID := mustAdvanceToOnVerification(t, env, "History pagination test")

		// Первая страница — версии 4 и 3
		code, resp := getApplicationHistory(t, env.Inspector, appUUID, 2, 0)
		assert.Equal(t, http.StatusOK, code)
		require.Len(t, resp.History, 2)
		assert.Equal(t, int64(4), resp.History[0].Version)
		assert.Equal(t, int64(3), resp.History[1].Version)

		// Вторая страница — версии 2 и 1
		code, resp = getApplicationHistory(t, env.Inspector, appUUID, 2, 2)
		assert.Equal(t, http.StatusOK, code)
		require.Len(t, resp.History, 2)
		assert.Equal(t, int64(2), resp.History[0].Version)
		assert.Equal(t, int64(1), resp.History[1].Version)

		// offset=3 — только последняя (самая старая) запись
		code, resp = getApplicationHistory(t, env.Inspector, appUUID, 2, 3)
		assert.Equal(t, http.StatusOK, code)
		require.Len(t, resp.History, 1)
		assert.Equal(t, int64(1), resp.History[0].Version)
	})

	// as_managed_by — manager является участником заявки (managed_by).
	t.Run("as_managed_by", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"History managed_by", "Desc.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID) // manager → ManagedBy

		code, resp := getApplicationHistory(t, env.Manager, appUUID, 10, 0)
		assert.Equal(t, http.StatusOK, code, "manager (ManagedBy) should access history")
		assert.Len(t, resp.History, 1)
	})

	// as_executed_by — engineer является участником заявки (executed_by).
	t.Run("as_executed_by", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"History executed_by", "Desc.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID) // engineer → ExecutedBy

		code, resp := getApplicationHistory(t, env.Engineer, appUUID, 10, 0)
		assert.Equal(t, http.StatusOK, code, "engineer (ExecutedBy) should access history")
		assert.Len(t, resp.History, 1)
	})

	// as_chief — chief не является участником, но доступ разрешён по роли.
	t.Run("as_chief", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"History chief access", "Desc.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)

		code, _ := getApplicationHistory(t, env.Chief, appUUID, 10, 0)
		assert.Equal(t, http.StatusOK, code, "chief should access any application history")
	})

	// as_analytic — analytic не является участником, но доступ разрешён по роли.
	t.Run("as_analytic", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"History analytic access", "Desc.")
		mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)

		code, _ := getApplicationHistory(t, env.Analytic, appUUID, 10, 0)
		assert.Equal(t, http.StatusOK, code, "analytic should access any application history")
	})

	// permission_denied — engineer2 из другого отдела, не участник заявки.
	t.Run("permission_denied", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"History denied", "Desc.")

		code, _ := getApplicationHistory(t, env.Engineer2, appUUID, 10, 0)
		assert.Equal(t, http.StatusForbidden, code)
	})

	// not_found — несуществующий UUID заявки.
	t.Run("not_found", func(t *testing.T) {
		code, _ := getApplicationHistory(t, env.Inspector, "00000000-0000-0000-0000-000000000000", 10, 0)
		assert.Equal(t, http.StatusNotFound, code)
	})

	// invalid_count_zero — count=0 должен вернуть 400.
	t.Run("invalid_count_zero", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"History invalid count zero", "Desc.")
		code, _ := getApplicationHistory(t, env.Inspector, appUUID, 0, 0)
		assert.Equal(t, http.StatusBadRequest, code)
	})

	// invalid_count_overflow — count=101 должен вернуть 400.
	t.Run("invalid_count_overflow", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"History invalid count overflow", "Desc.")
		code, _ := getApplicationHistory(t, env.Inspector, appUUID, 101, 0)
		assert.Equal(t, http.StatusBadRequest, code)
	})

	// invalid_offset — offset=-1 должен вернуть 400.
	t.Run("invalid_offset", func(t *testing.T) {
		appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID,
			"History invalid offset", "Desc.")
		code, _ := getApplicationHistory(t, env.Inspector, appUUID, 10, -1)
		assert.Equal(t, http.StatusBadRequest, code)
	})
}
