package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── GetCompanyEmployee ───────────────────────────────────────────────────────

func TestGetCompanyEmployee(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := chief.get(fmt.Sprintf("/api/auth/company/%s/employee/%s/info", companyUUID, memberLogin.UserUUID))

		assert.Equal(t, http.StatusOK, status, "body: %s", body)
		var resp employeeInfoResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, memberLogin.UserUUID, resp.UserUUID)
		assert.NotEmpty(t, resp.Role)
		assert.NotEmpty(t, resp.JoinedAt)
	})

	t.Run("member_can_read", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberALogin := mustRegisterAndLogin(t, c)
		memberA := c.withToken(memberALogin.AccessToken)
		mustAddMember(t, chief, memberA, companyUUID)

		_, memberBLogin := mustRegisterAndLogin(t, c)
		memberB := c.withToken(memberBLogin.AccessToken)
		mustAddMember(t, chief, memberB, companyUUID)

		// memberA читает информацию о memberB
		status, body := memberA.get(fmt.Sprintf("/api/auth/company/%s/employee/%s/info", companyUUID, memberBLogin.UserUUID))
		assert.Equal(t, http.StatusOK, status, "member should be able to read peer info (body: %s)", body)
	})

	t.Run("not_found", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		status, body := chief.get(fmt.Sprintf("/api/auth/company/%s/employee/00000000-0000-0000-0000-000000000001/info", companyUUID))
		assert.Equal(t, http.StatusNotFound, status, "unknown employee should return 404 (body: %s)", body)
	})

	t.Run("outsider_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		_, outsiderLogin := mustRegisterAndLogin(t, c)
		outsider := c.withToken(outsiderLogin.AccessToken)

		status, body := outsider.get(fmt.Sprintf("/api/auth/company/%s/employee/%s/info", companyUUID, memberLogin.UserUUID))
		assert.Equal(t, http.StatusForbidden, status, "outsider should get 403 (body: %s)", body)
	})
}

// ─── GetCompanyEmployees ──────────────────────────────────────────────────────

func TestGetCompanyEmployees(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := chief.get("/api/auth/company/" + companyUUID + "/employees/list?offset=0&count=10")

		assert.Equal(t, http.StatusOK, status, "body: %s", body)
		var resp employeesListResp
		require.NoError(t, json.Unmarshal(body, &resp))
		// Компания: chief + 1 member = 2 сотрудника
		assert.GreaterOrEqual(t, len(resp.Employees), 2)

		uuids := make([]string, 0, len(resp.Employees))
		for _, e := range resp.Employees {
			uuids = append(uuids, e.UserUUID)
		}
		assert.Contains(t, uuids, chiefLogin.UserUUID)
		assert.Contains(t, uuids, memberLogin.UserUUID)
	})

	t.Run("filter_by_role", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)
		mustSetEmployeeRole(t, chief, companyUUID, memberLogin.UserUUID, "engineer")

		status, body := chief.get("/api/auth/company/" + companyUUID + "/employees/list?offset=0&count=10&role=engineer")

		assert.Equal(t, http.StatusOK, status, "body: %s", body)
		var resp employeesListResp
		require.NoError(t, json.Unmarshal(body, &resp))
		require.NotEmpty(t, resp.Employees, "should have at least one engineer")
		for _, e := range resp.Employees {
			assert.Equal(t, "engineer", e.Role)
		}
	})

	t.Run("filter_by_department", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		_, memberALogin := mustRegisterAndLogin(t, c)
		memberA := c.withToken(memberALogin.AccessToken)
		mustAddMember(t, chief, memberA, companyUUID)
		mustAddEmployeeToDepartment(t, chief, companyUUID, deptUUID, memberALogin.UserUUID)

		_, memberBLogin := mustRegisterAndLogin(t, c)
		memberB := c.withToken(memberBLogin.AccessToken)
		mustAddMember(t, chief, memberB, companyUUID)
		// memberB намеренно не добавляется в департамент

		status, body := chief.get("/api/auth/company/" + companyUUID + "/employees/list?offset=0&count=10&department_uuid=" + deptUUID)

		assert.Equal(t, http.StatusOK, status, "body: %s", body)
		var resp employeesListResp
		require.NoError(t, json.Unmarshal(body, &resp))

		uuids := make([]string, 0, len(resp.Employees))
		for _, e := range resp.Employees {
			uuids = append(uuids, e.UserUUID)
		}
		assert.Contains(t, uuids, memberALogin.UserUUID, "memberA should be in department filter result")
		assert.NotContains(t, uuids, memberBLogin.UserUUID, "memberB should not appear in department filter result")
	})

	t.Run("pagination", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := chief.get("/api/auth/company/" + companyUUID + "/employees/list?offset=0&count=1")

		assert.Equal(t, http.StatusOK, status, "body: %s", body)
		var resp employeesListResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.LessOrEqual(t, len(resp.Employees), 1)
	})

	t.Run("outsider_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, outsiderLogin := mustRegisterAndLogin(t, c)
		outsider := c.withToken(outsiderLogin.AccessToken)

		status, body := outsider.get("/api/auth/company/" + companyUUID + "/employees/list?offset=0&count=10")
		assert.Equal(t, http.StatusForbidden, status, "outsider should get 403 (body: %s)", body)
	})
}

// ─── GetCompanyEmployeesSummary ───────────────────────────────────────────────

func TestGetCompanyEmployeesSummary(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := chief.get("/api/auth/company/" + companyUUID + "/employees/summary")

		assert.Equal(t, http.StatusOK, status, "body: %s", body)
		var resp employeesSummaryResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, int64(1), resp.ChiefCount, "should have exactly 1 chief")
		assert.Equal(t, int64(1), resp.UnemployedCount, "new member should be unemployed")
	})

	t.Run("filter_by_department", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)
		mustSetEmployeeRole(t, chief, companyUUID, memberLogin.UserUUID, "engineer")
		mustAddEmployeeToDepartment(t, chief, companyUUID, deptUUID, memberLogin.UserUUID)

		status, body := chief.get("/api/auth/company/" + companyUUID + "/employees/summary?department_uuid=" + deptUUID)

		assert.Equal(t, http.StatusOK, status, "body: %s", body)
		var resp employeesSummaryResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, int64(1), resp.EngineerCount, "department summary should show 1 engineer")
		assert.Equal(t, int64(0), resp.UnemployedCount, "department should have no unemployed")
	})

	t.Run("empty_company", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		status, body := chief.get("/api/auth/company/" + companyUUID + "/employees/summary")

		assert.Equal(t, http.StatusOK, status, "body: %s", body)
		var resp employeesSummaryResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, int64(1), resp.ChiefCount)
		assert.Equal(t, int64(0), resp.UnemployedCount)
		assert.Equal(t, int64(0), resp.EngineerCount)
		assert.Equal(t, int64(0), resp.ManagerCount)
		assert.Equal(t, int64(0), resp.InspectorCount)
		assert.Equal(t, int64(0), resp.AnalyticsCount)
	})

	t.Run("outsider_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, outsiderLogin := mustRegisterAndLogin(t, c)
		outsider := c.withToken(outsiderLogin.AccessToken)

		status, body := outsider.get("/api/auth/company/" + companyUUID + "/employees/summary")
		assert.Equal(t, http.StatusForbidden, status, "outsider should get 403 (body: %s)", body)
	})
}

// ─── UpdateEmployeeRole ───────────────────────────────────────────────────────

func TestUpdateEmployeeRole(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := chief.patch(
			fmt.Sprintf("/api/auth/company/%s/employee/%s/role", companyUUID, memberLogin.UserUUID),
			map[string]string{"role": "engineer"},
		)
		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		// Проверяем, что роль действительно изменилась
		getStatus, getBody := chief.get(fmt.Sprintf("/api/auth/company/%s/employee/%s/info", companyUUID, memberLogin.UserUUID))
		require.Equal(t, http.StatusOK, getStatus)
		var emp employeeInfoResp
		require.NoError(t, json.Unmarshal(getBody, &emp))
		assert.Equal(t, "engineer", emp.Role)
	})

	t.Run("invalid_role", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := chief.patch(
			fmt.Sprintf("/api/auth/company/%s/employee/%s/role", companyUUID, memberLogin.UserUUID),
			map[string]string{"role": "superuser"},
		)
		assert.Equal(t, http.StatusBadRequest, status, "invalid role should return 400 (body: %s)", body)
	})

	t.Run("self_change", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		status, body := chief.patch(
			fmt.Sprintf("/api/auth/company/%s/employee/%s/role", companyUUID, chiefLogin.UserUUID),
			map[string]string{"role": "engineer"},
		)
		assert.Equal(t, http.StatusBadRequest, status, "self role change should return 400 (body: %s)", body)
	})

	t.Run("non_chief_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberALogin := mustRegisterAndLogin(t, c)
		memberA := c.withToken(memberALogin.AccessToken)
		mustAddMember(t, chief, memberA, companyUUID)

		_, memberBLogin := mustRegisterAndLogin(t, c)
		memberB := c.withToken(memberBLogin.AccessToken)
		mustAddMember(t, chief, memberB, companyUUID)

		// memberA (не chief) пытается изменить роль memberB
		status, body := memberA.patch(
			fmt.Sprintf("/api/auth/company/%s/employee/%s/role", companyUUID, memberBLogin.UserUUID),
			map[string]string{"role": "engineer"},
		)
		assert.Equal(t, http.StatusForbidden, status, "non-chief should get 403 (body: %s)", body)
	})

	t.Run("target_not_in_company", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, outsiderLogin := mustRegisterAndLogin(t, c)

		status, body := chief.patch(
			fmt.Sprintf("/api/auth/company/%s/employee/%s/role", companyUUID, outsiderLogin.UserUUID),
			map[string]string{"role": "engineer"},
		)
		assert.Equal(t, http.StatusNotFound, status, "target not in company should return 404 (body: %s)", body)
	})
}

// ─── RemoveCompanyEmployee ────────────────────────────────────────────────────

func TestRemoveCompanyEmployee(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := chief.delete(
			fmt.Sprintf("/api/auth/company/%s/employee/%s", companyUUID, memberLogin.UserUUID),
			nil,
		)
		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		// После удаления сотрудник не должен быть найден в компании
		getStatus, getBody := chief.get(fmt.Sprintf("/api/auth/company/%s/employee/%s/info", companyUUID, memberLogin.UserUUID))
		assert.Equal(t, http.StatusNotFound, getStatus, "removed employee should return 404 (body: %s)", getBody)
	})

	t.Run("self_removal", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		status, body := chief.delete(
			fmt.Sprintf("/api/auth/company/%s/employee/%s", companyUUID, chiefLogin.UserUUID),
			nil,
		)
		assert.Equal(t, http.StatusBadRequest, status, "self removal should return 400 (body: %s)", body)
	})

	t.Run("non_chief_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberALogin := mustRegisterAndLogin(t, c)
		memberA := c.withToken(memberALogin.AccessToken)
		mustAddMember(t, chief, memberA, companyUUID)

		_, memberBLogin := mustRegisterAndLogin(t, c)
		memberB := c.withToken(memberBLogin.AccessToken)
		mustAddMember(t, chief, memberB, companyUUID)

		// memberA (не chief) пытается удалить memberB
		status, body := memberA.delete(
			fmt.Sprintf("/api/auth/company/%s/employee/%s", companyUUID, memberBLogin.UserUUID),
			nil,
		)
		assert.Equal(t, http.StatusForbidden, status, "non-chief should get 403 (body: %s)", body)
	})

	t.Run("not_found", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		status, body := chief.delete(
			fmt.Sprintf("/api/auth/company/%s/employee/00000000-0000-0000-0000-000000000001", companyUUID),
			nil,
		)
		assert.Equal(t, http.StatusNotFound, status, "unknown employee should return 404 (body: %s)", body)
	})
}

// ─── Full employee workflow ───────────────────────────────────────────────────

func TestEmployeeFullWorkflow(t *testing.T) {
	c := newClient()

	// 1. Chief создаёт компанию и открывает её
	_, chiefLogin := mustRegisterAndLogin(t, c)
	chief := c.withToken(chiefLogin.AccessToken)
	companyUUID := mustCreateCompany(t, chief, randomTitle())
	mustOpenCompany(t, chief, companyUUID)

	// 2. Два сотрудника вступают в компанию
	_, memberALogin := mustRegisterAndLogin(t, c)
	memberA := c.withToken(memberALogin.AccessToken)
	mustAddMember(t, chief, memberA, companyUUID)

	_, memberBLogin := mustRegisterAndLogin(t, c)
	memberB := c.withToken(memberBLogin.AccessToken)
	mustAddMember(t, chief, memberB, companyUUID)

	// 3. Проверяем сводку: 1 chief + 2 unemployed
	summaryStatus, summaryBody := chief.get("/api/auth/company/" + companyUUID + "/employees/summary")
	require.Equal(t, http.StatusOK, summaryStatus)
	var summary employeesSummaryResp
	require.NoError(t, json.Unmarshal(summaryBody, &summary))
	assert.Equal(t, int64(1), summary.ChiefCount, "should have 1 chief")
	assert.Equal(t, int64(2), summary.UnemployedCount, "should have 2 unemployed")

	// 4. Chief назначает memberA инженером
	mustSetEmployeeRole(t, chief, companyUUID, memberALogin.UserUUID, "engineer")

	// Проверяем роль через GET
	infoStatus, infoBody := chief.get(fmt.Sprintf("/api/auth/company/%s/employee/%s/info", companyUUID, memberALogin.UserUUID))
	require.Equal(t, http.StatusOK, infoStatus)
	var empA employeeInfoResp
	require.NoError(t, json.Unmarshal(infoBody, &empA))
	assert.Equal(t, "engineer", empA.Role, "memberA should be engineer")

	// 5. Фильтрация по роли: только инженеры
	listStatus, listBody := chief.get("/api/auth/company/" + companyUUID + "/employees/list?offset=0&count=10&role=engineer")
	require.Equal(t, http.StatusOK, listStatus)
	var engineers employeesListResp
	require.NoError(t, json.Unmarshal(listBody, &engineers))
	require.NotEmpty(t, engineers.Employees)
	engineerUUIDs := make([]string, 0, len(engineers.Employees))
	for _, e := range engineers.Employees {
		engineerUUIDs = append(engineerUUIDs, e.UserUUID)
	}
	assert.Contains(t, engineerUUIDs, memberALogin.UserUUID, "memberA should be in engineers list")
	assert.NotContains(t, engineerUUIDs, memberBLogin.UserUUID, "memberB should not be in engineers list")

	// 6. Chief добавляет memberA в департамент и проверяет сводку по нему
	deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")
	mustAddEmployeeToDepartment(t, chief, companyUUID, deptUUID, memberALogin.UserUUID)

	deptSummaryStatus, deptSummaryBody := chief.get("/api/auth/company/" + companyUUID + "/employees/summary?department_uuid=" + deptUUID)
	require.Equal(t, http.StatusOK, deptSummaryStatus)
	var deptSummary employeesSummaryResp
	require.NoError(t, json.Unmarshal(deptSummaryBody, &deptSummary))
	assert.Equal(t, int64(1), deptSummary.EngineerCount, "department should have 1 engineer")

	// 7. Chief удаляет memberB из компании
	removeStatus, removeBody := chief.delete(
		fmt.Sprintf("/api/auth/company/%s/employee/%s", companyUUID, memberBLogin.UserUUID),
		nil,
	)
	assert.Equal(t, http.StatusOK, removeStatus, "remove failed (body: %s)", removeBody)

	// memberB больше не найден в компании
	getStatus, _ := chief.get(fmt.Sprintf("/api/auth/company/%s/employee/%s/info", companyUUID, memberBLogin.UserUUID))
	assert.Equal(t, http.StatusNotFound, getStatus, "removed employee should return 404")

	// 8. Итоговая сводка: 1 chief + 1 engineer (memberA), memberB убран
	finalSummaryStatus, finalSummaryBody := chief.get("/api/auth/company/" + companyUUID + "/employees/summary")
	require.Equal(t, http.StatusOK, finalSummaryStatus)
	var finalSummary employeesSummaryResp
	require.NoError(t, json.Unmarshal(finalSummaryBody, &finalSummary))
	assert.Equal(t, int64(1), finalSummary.ChiefCount, "should still have 1 chief")
	assert.Equal(t, int64(1), finalSummary.EngineerCount, "should have 1 engineer (memberA)")
	assert.Equal(t, int64(0), finalSummary.UnemployedCount, "no unemployed after memberB removed")
}
