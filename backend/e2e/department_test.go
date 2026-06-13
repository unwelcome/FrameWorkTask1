package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── CreateDepartment ─────────────────────────────────────────────────────────

func TestCreateDepartment(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		chief := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		status, body := chief.post("/api/auth/company/"+companyUUID+"/department", map[string]string{
			"title": "Engineering",
		})

		assert.Equal(t, http.StatusCreated, status, "body: %s", body)
		var resp createDepartmentResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.NotEmpty(t, resp.DepartmentUUID)
	})

	t.Run("empty_title", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		chief := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		status, body := chief.post("/api/auth/company/"+companyUUID+"/department", map[string]string{
			"title": "",
		})
		assert.Equal(t, http.StatusBadRequest, status, "empty title should return 400 (body: %s)", body)
	})

	t.Run("title_too_long", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		chief := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		status, body := chief.post("/api/auth/company/"+companyUUID+"/department", map[string]string{
			"title": strings.Repeat("a", 256),
		})
		assert.Equal(t, http.StatusBadRequest, status, "title >255 chars should return 400 (body: %s)", body)
	})

	t.Run("non_chief_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := member.post("/api/auth/company/"+companyUUID+"/department", map[string]string{
			"title": "Engineering",
		})
		assert.Equal(t, http.StatusForbidden, status, "non-chief should get 403 (body: %s)", body)
	})
}

// ─── GetDepartment ────────────────────────────────────────────────────────────

func TestGetDepartment(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		chief := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptTitle := "Engineering"
		deptUUID := mustCreateDepartment(t, chief, companyUUID, deptTitle)

		status, body := chief.get(fmt.Sprintf("/api/auth/company/%s/department/%s", companyUUID, deptUUID))

		assert.Equal(t, http.StatusOK, status, "body: %s", body)
		var resp departmentResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, deptUUID, resp.DepartmentUUID)
		assert.Equal(t, companyUUID, resp.CompanyUUID)
		assert.Equal(t, deptTitle, resp.Title)
		assert.NotEmpty(t, resp.CreatedAt)
		assert.NotEmpty(t, resp.CreatedBy)
	})

	t.Run("member_can_read", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := member.get(fmt.Sprintf("/api/auth/company/%s/department/%s", companyUUID, deptUUID))
		assert.Equal(t, http.StatusOK, status, "member should be able to read department (body: %s)", body)
	})

	t.Run("not_found", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		chief := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		status, body := chief.get(fmt.Sprintf("/api/auth/company/%s/department/00000000-0000-0000-0000-000000000001", companyUUID))
		assert.Equal(t, http.StatusNotFound, status, "unknown department should return 404 (body: %s)", body)
	})

	t.Run("outsider_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		_, outsiderLogin := mustRegisterAndLogin(t, c)
		outsider := c.withToken(outsiderLogin.AccessToken)

		status, body := outsider.get(fmt.Sprintf("/api/auth/company/%s/department/%s", companyUUID, deptUUID))
		assert.Equal(t, http.StatusForbidden, status, "outsider should get 403 (body: %s)", body)
	})
}

// ─── GetCompanyDepartments ────────────────────────────────────────────────────

func TestGetCompanyDepartments(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		chief := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		d1 := mustCreateDepartment(t, chief, companyUUID, "Engineering")
		d2 := mustCreateDepartment(t, chief, companyUUID, "Marketing")

		status, body := chief.get("/api/auth/company/" + companyUUID + "/departments/list?offset=0&count=10")

		assert.Equal(t, http.StatusOK, status, "body: %s", body)
		var resp departmentsListResp
		require.NoError(t, json.Unmarshal(body, &resp))
		require.GreaterOrEqual(t, len(resp.Departments), 2, "should contain at least 2 departments")

		ids := make([]string, 0, len(resp.Departments))
		for _, d := range resp.Departments {
			ids = append(ids, d.DepartmentUUID)
		}
		assert.Contains(t, ids, d1)
		assert.Contains(t, ids, d2)
	})

	t.Run("empty_when_no_departments", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		chief := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		status, body := chief.get("/api/auth/company/" + companyUUID + "/departments/list?offset=0&count=10")

		assert.Equal(t, http.StatusOK, status, "body: %s", body)
		var resp departmentsListResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Empty(t, resp.Departments)
	})

	t.Run("pagination", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		chief := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		mustCreateDepartment(t, chief, companyUUID, "Dept 1")
		mustCreateDepartment(t, chief, companyUUID, "Dept 2")
		mustCreateDepartment(t, chief, companyUUID, "Dept 3")

		status, body := chief.get("/api/auth/company/" + companyUUID + "/departments/list?offset=0&count=2")

		assert.Equal(t, http.StatusOK, status, "body: %s", body)
		var resp departmentsListResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.LessOrEqual(t, len(resp.Departments), 2)
	})

	t.Run("outsider_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, outsiderLogin := mustRegisterAndLogin(t, c)
		outsider := c.withToken(outsiderLogin.AccessToken)

		status, body := outsider.get("/api/auth/company/" + companyUUID + "/departments/list?offset=0&count=10")
		assert.Equal(t, http.StatusForbidden, status, "outsider should get 403 (body: %s)", body)
	})
}

// ─── UpdateDepartmentTitle ────────────────────────────────────────────────────

func TestUpdateDepartmentTitle(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		chief := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		status, body := chief.patch(
			fmt.Sprintf("/api/auth/company/%s/department/%s/title", companyUUID, deptUUID),
			map[string]string{"title": "Backend Engineering"},
		)
		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		// Проверяем, что название действительно изменилось
		status, body = chief.get(fmt.Sprintf("/api/auth/company/%s/department/%s", companyUUID, deptUUID))
		require.Equal(t, http.StatusOK, status)
		var resp departmentResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, "Backend Engineering", resp.Title)
	})

	t.Run("empty_title", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		chief := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		status, body := chief.patch(
			fmt.Sprintf("/api/auth/company/%s/department/%s/title", companyUUID, deptUUID),
			map[string]string{"title": ""},
		)
		assert.Equal(t, http.StatusBadRequest, status, "empty title should return 400 (body: %s)", body)
	})

	t.Run("non_chief_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := member.patch(
			fmt.Sprintf("/api/auth/company/%s/department/%s/title", companyUUID, deptUUID),
			map[string]string{"title": "Hacked Name"},
		)
		assert.Equal(t, http.StatusForbidden, status, "non-chief should get 403 (body: %s)", body)
	})

	t.Run("not_found", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		chief := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		status, body := chief.patch(
			fmt.Sprintf("/api/auth/company/%s/department/00000000-0000-0000-0000-000000000001/title", companyUUID),
			map[string]string{"title": "New Name"},
		)
		assert.Equal(t, http.StatusNotFound, status, "unknown department should return 404 (body: %s)", body)
	})
}

// ─── DeleteDepartment ─────────────────────────────────────────────────────────

func TestDeleteDepartment(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		chief := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		status, body := chief.delete(
			fmt.Sprintf("/api/auth/company/%s/department/%s", companyUUID, deptUUID),
			nil,
		)
		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		// После удаления департамент должен быть недоступен
		status, body = chief.get(fmt.Sprintf("/api/auth/company/%s/department/%s", companyUUID, deptUUID))
		assert.Equal(t, http.StatusNotFound, status, "deleted department should return 404 (body: %s)", body)
	})

	t.Run("non_chief_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := member.delete(
			fmt.Sprintf("/api/auth/company/%s/department/%s", companyUUID, deptUUID),
			nil,
		)
		assert.Equal(t, http.StatusForbidden, status, "non-chief should get 403 (body: %s)", body)
	})

	t.Run("not_found", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		chief := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		status, body := chief.delete(
			fmt.Sprintf("/api/auth/company/%s/department/00000000-0000-0000-0000-000000000001", companyUUID),
			nil,
		)
		assert.Equal(t, http.StatusNotFound, status, "unknown department should return 404 (body: %s)", body)
	})
}

// ─── AddEmployeeToDepartment ──────────────────────────────────────────────────

func TestAddEmployeeToDepartment(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := chief.post(
			fmt.Sprintf("/api/auth/company/%s/department/%s/employee/%s", companyUUID, deptUUID, memberLogin.UserUUID),
			nil,
		)
		assert.Equal(t, http.StatusOK, status, "body: %s", body)
	})

	t.Run("already_in_department", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)
		mustAddEmployeeToDepartment(t, chief, companyUUID, deptUUID, memberLogin.UserUUID)

		// Пытаемся добавить снова
		status, body := chief.post(
			fmt.Sprintf("/api/auth/company/%s/department/%s/employee/%s", companyUUID, deptUUID, memberLogin.UserUUID),
			nil,
		)
		assert.Equal(t, http.StatusConflict, status, "already in department should return 409 (body: %s)", body)
	})

	t.Run("target_not_in_company", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		// Регистрируем пользователя, который НЕ вступил в компанию
		_, outsiderLogin := mustRegisterAndLogin(t, c)

		status, body := chief.post(
			fmt.Sprintf("/api/auth/company/%s/department/%s/employee/%s", companyUUID, deptUUID, outsiderLogin.UserUUID),
			nil,
		)
		assert.Equal(t, http.StatusForbidden, status, "target not in company should return 403 (body: %s)", body)
	})

	t.Run("non_chief_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		_, memberALogin := mustRegisterAndLogin(t, c)
		memberA := c.withToken(memberALogin.AccessToken)
		mustAddMember(t, chief, memberA, companyUUID)

		_, memberBLogin := mustRegisterAndLogin(t, c)
		memberB := c.withToken(memberBLogin.AccessToken)
		mustAddMember(t, chief, memberB, companyUUID)

		// memberA (не chief) пытается добавить memberB в департамент
		status, body := memberA.post(
			fmt.Sprintf("/api/auth/company/%s/department/%s/employee/%s", companyUUID, deptUUID, memberBLogin.UserUUID),
			nil,
		)
		assert.Equal(t, http.StatusForbidden, status, "non-chief should get 403 (body: %s)", body)
	})

	t.Run("department_not_found", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := chief.post(
			fmt.Sprintf("/api/auth/company/%s/department/00000000-0000-0000-0000-000000000001/employee/%s", companyUUID, memberLogin.UserUUID),
			nil,
		)
		assert.Equal(t, http.StatusNotFound, status, "unknown department should return 404 (body: %s)", body)
	})
}

// ─── RemoveEmployeeFromDepartment ─────────────────────────────────────────────

func TestRemoveEmployeeFromDepartment(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)
		mustAddEmployeeToDepartment(t, chief, companyUUID, deptUUID, memberLogin.UserUUID)

		status, body := chief.delete(
			fmt.Sprintf("/api/auth/company/%s/department/%s/employee/%s", companyUUID, deptUUID, memberLogin.UserUUID),
			nil,
		)
		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		// Проверяем, что у сотрудника теперь нет департамента
		status, body = chief.get(fmt.Sprintf("/api/auth/company/%s/employee/%s/info", companyUUID, memberLogin.UserUUID))
		require.Equal(t, http.StatusOK, status)
		var empResp employeeInfoResp
		require.NoError(t, json.Unmarshal(body, &empResp))
		assert.Empty(t, empResp.DepartmentUUID, "employee should have no department after removal")
	})

	t.Run("not_in_department", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)
		// Намеренно не добавляем member в департамент

		status, body := chief.delete(
			fmt.Sprintf("/api/auth/company/%s/department/%s/employee/%s", companyUUID, deptUUID, memberLogin.UserUUID),
			nil,
		)
		assert.Equal(t, http.StatusBadRequest, status, "employee not in department should return 400 (body: %s)", body)
	})

	t.Run("non_chief_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

		_, memberALogin := mustRegisterAndLogin(t, c)
		memberA := c.withToken(memberALogin.AccessToken)
		mustAddMember(t, chief, memberA, companyUUID)

		_, memberBLogin := mustRegisterAndLogin(t, c)
		memberB := c.withToken(memberBLogin.AccessToken)
		mustAddMember(t, chief, memberB, companyUUID)
		mustAddEmployeeToDepartment(t, chief, companyUUID, deptUUID, memberBLogin.UserUUID)

		// memberA (не chief) пытается убрать memberB из департамента
		status, body := memberA.delete(
			fmt.Sprintf("/api/auth/company/%s/department/%s/employee/%s", companyUUID, deptUUID, memberBLogin.UserUUID),
			nil,
		)
		assert.Equal(t, http.StatusForbidden, status, "non-chief should get 403 (body: %s)", body)
	})

	t.Run("department_not_found", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := chief.delete(
			fmt.Sprintf("/api/auth/company/%s/department/00000000-0000-0000-0000-000000000001/employee/%s", companyUUID, memberLogin.UserUUID),
			nil,
		)
		assert.Equal(t, http.StatusNotFound, status, "unknown department should return 404 (body: %s)", body)
	})
}

// ─── Full department workflow ─────────────────────────────────────────────────

func TestDepartmentFullWorkflow(t *testing.T) {
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

	// 3. Chief создаёт департамент
	deptUUID := mustCreateDepartment(t, chief, companyUUID, "Engineering")

	// Департамент виден в списке
	listStatus, listBody := chief.get("/api/auth/company/" + companyUUID + "/departments/list?offset=0&count=10")
	require.Equal(t, http.StatusOK, listStatus)
	var listResp departmentsListResp
	require.NoError(t, json.Unmarshal(listBody, &listResp))
	deptIDs := make([]string, 0, len(listResp.Departments))
	for _, d := range listResp.Departments {
		deptIDs = append(deptIDs, d.DepartmentUUID)
	}
	assert.Contains(t, deptIDs, deptUUID, "department should appear in list")

	// 4. Chief добавляет A в департамент
	mustAddEmployeeToDepartment(t, chief, companyUUID, deptUUID, memberALogin.UserUUID)

	// Проверяем, что у A проставлен department_uuid
	infoStatus, infoBody := chief.get(fmt.Sprintf("/api/auth/company/%s/employee/%s/info", companyUUID, memberALogin.UserUUID))
	require.Equal(t, http.StatusOK, infoStatus)
	var empA employeeInfoResp
	require.NoError(t, json.Unmarshal(infoBody, &empA))
	assert.Equal(t, deptUUID, empA.DepartmentUUID, "employee A should be in the department")

	// 5. Chief переименовывает департамент
	renameStatus, renameBody := chief.patch(
		fmt.Sprintf("/api/auth/company/%s/department/%s/title", companyUUID, deptUUID),
		map[string]string{"title": "Backend Engineering"},
	)
	assert.Equal(t, http.StatusOK, renameStatus, "rename failed (body: %s)", renameBody)

	// Проверяем новое название
	getStatus, getBody := chief.get(fmt.Sprintf("/api/auth/company/%s/department/%s", companyUUID, deptUUID))
	require.Equal(t, http.StatusOK, getStatus)
	var deptResp departmentResp
	require.NoError(t, json.Unmarshal(getBody, &deptResp))
	assert.Equal(t, "Backend Engineering", deptResp.Title)

	// 6. Chief убирает A из департамента
	removeStatus, removeBody := chief.delete(
		fmt.Sprintf("/api/auth/company/%s/department/%s/employee/%s", companyUUID, deptUUID, memberALogin.UserUUID),
		nil,
	)
	assert.Equal(t, http.StatusOK, removeStatus, "remove from dept failed (body: %s)", removeBody)

	// У A теперь нет департамента
	infoStatus, infoBody = chief.get(fmt.Sprintf("/api/auth/company/%s/employee/%s/info", companyUUID, memberALogin.UserUUID))
	require.Equal(t, http.StatusOK, infoStatus)
	require.NoError(t, json.Unmarshal(infoBody, &empA))
	assert.Empty(t, empA.DepartmentUUID, "employee A should have no department after removal")

	// 7. Chief добавляет B в департамент
	mustAddEmployeeToDepartment(t, chief, companyUUID, deptUUID, memberBLogin.UserUUID)

	// 8. Chief удаляет департамент
	deleteStatus, deleteBody := chief.delete(
		fmt.Sprintf("/api/auth/company/%s/department/%s", companyUUID, deptUUID),
		nil,
	)
	assert.Equal(t, http.StatusOK, deleteStatus, "delete department failed (body: %s)", deleteBody)

	// После удаления департамент недоступен
	getStatus, _ = chief.get(fmt.Sprintf("/api/auth/company/%s/department/%s", companyUUID, deptUUID))
	assert.Equal(t, http.StatusNotFound, getStatus, "deleted department should return 404")
}
