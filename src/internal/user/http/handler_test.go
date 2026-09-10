package userhttp_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"gin-start/src/ent"
	"gin-start/src/internal/bootstrap"
	"gin-start/src/internal/database"

	"github.com/labstack/echo/v5"
)

func newAPI(t *testing.T) (*ent.Client, *echo.Echo) {
	t.Helper()
	client, err := database.Open(t.Context(), "file:"+filepath.Join(t.TempDir(), "api.db")+"?_fk=1")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	application := bootstrap.NewApplication(client)
	return client, application.Router
}

func request(t *testing.T, router http.Handler, method, path, body string, wantStatus int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != wantStatus {
		t.Fatalf("%s %s: status=%d want=%d body=%s", method, path, recorder.Code, wantStatus, recorder.Body.String())
	}
	return recorder
}

func TestAPILifecycle(t *testing.T) {
	_, router := newAPI(t)
	const base = "/api/v1/users"
	empty := request(t, router, "GET", base, "", 200)
	if !strings.Contains(empty.Body.String(), `"users":[]`) {
		t.Fatalf("empty list must be [], got %s", empty.Body.String())
	}
	created := request(t, router, "POST", base, `{"name":"  테스트  ","email":" DEMO@EXAMPLE.COM "}`, 201)
	type responseUser struct {
		ID        int       `json:"id"`
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		IsActive  bool      `json:"isActive"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}
	var u responseUser
	if err := json.Unmarshal(created.Body.Bytes(), &u); err != nil {
		t.Fatal(err)
	}
	path := base + "/" + strconv.Itoa(u.ID)
	if u.ID <= 0 || u.Name != "테스트" || u.Email != "demo@example.com" || !u.IsActive || u.CreatedAt.IsZero() || u.UpdatedAt.IsZero() || created.Header().Get("Location") != path {
		t.Fatalf("unexpected user or Location: %s", created.Body.String())
	}
	var fields map[string]any
	if err := json.Unmarshal(created.Body.Bytes(), &fields); err != nil || len(fields) != 6 {
		t.Fatalf("response fields = %v, error = %v", fields, err)
	}
	request(t, router, "GET", path, "", 200)
	request(t, router, "POST", base, `{"name":"Other","email":"demo@example.com"}`, 409)
	lastUpdatedAt := u.UpdatedAt
	for range 2 {
		inactive := request(t, router, "PATCH", path+"/deactivate", "", 200)
		var updated responseUser
		if err := json.Unmarshal(inactive.Body.Bytes(), &updated); err != nil {
			t.Fatal(err)
		}
		if updated.IsActive || !updated.CreatedAt.Equal(u.CreatedAt) || updated.UpdatedAt.Before(lastUpdatedAt) {
			t.Fatalf("unexpected deactivation response: %s", inactive.Body.String())
		}
		lastUpdatedAt = updated.UpdatedAt
	}
	got := request(t, router, "GET", path, "", 200)
	var fetched responseUser
	if err := json.Unmarshal(got.Body.Bytes(), &fetched); err != nil {
		t.Fatal(err)
	}
	if fetched.IsActive || !fetched.CreatedAt.Equal(u.CreatedAt) || !fetched.UpdatedAt.Equal(lastUpdatedAt) {
		t.Fatalf("inactive state or timestamp not persisted: %s", got.Body.String())
	}
}

func TestAPIErrorContracts(t *testing.T) {
	_, router := newAPI(t)
	const base = "/api/v1/users"
	tests := []struct {
		method, path, body string
		status             int
		code               string
	}{
		{"POST", base, "{", 400, "INVALID_BODY"},
		{"POST", base, "", 400, "INVALID_BODY"},
		{"POST", base, `{"name":123,"email":"demo@example.com"}`, 400, "INVALID_BODY"},
		{"POST", base, `{"name":"  ","email":"demo@example.com"}`, 400, "INVALID_NAME"},
		{"POST", base, `{"name":"Demo","email":"bad"}`, 400, "INVALID_EMAIL"},
		{"GET", base + "/abc", "", 400, "INVALID_ID"},
		{"GET", base + "/0", "", 400, "INVALID_ID"},
		{"GET", base + "/999999", "", 404, "USER_NOT_FOUND"},
		{"PATCH", base + "/abc/deactivate", "", 400, "INVALID_ID"},
		{"PATCH", base + "/999999/deactivate", "", 404, "USER_NOT_FOUND"},
		{"GET", base + "?limit=0", "", 400, "INVALID_PAGINATION"},
		{"GET", base + "?limit=101", "", 400, "INVALID_PAGINATION"},
		{"GET", base + "?offset=-1", "", 400, "INVALID_PAGINATION"},
		{"GET", base + "?offset=abc", "", 400, "INVALID_PAGINATION"},
		{"GET", base + "?limit=", "", 400, "INVALID_PAGINATION"},
		{"GET", "/missing", "", 404, "ROUTE_NOT_FOUND"},
		{"DELETE", base, "", 405, "METHOD_NOT_ALLOWED"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path+" "+tt.code, func(t *testing.T) {
			response := request(t, router, tt.method, tt.path, tt.body, tt.status)
			var result struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil || result.Error.Code != tt.code {
				t.Fatalf("error contract mismatch: %s", response.Body.String())
			}
		})
	}
}

func TestAPIBodyLimit(t *testing.T) {
	_, router := newAPI(t)
	body := `{"name":"` + strings.Repeat("x", 1<<20) + `","email":"demo@example.com"}`
	for _, streamed := range []bool{false, true} {
		t.Run(strconv.FormatBool(streamed), func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if streamed {
				req.ContentLength = -1
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			if recorder.Code != http.StatusRequestEntityTooLarge || !strings.Contains(recorder.Body.String(), `"code":"BODY_TOO_LARGE"`) {
				t.Fatalf("body limit response = %d %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestAPIHidesDatabaseFailure(t *testing.T) {
	client, router := newAPI(t)
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	response := request(t, router, "GET", "/api/v1/users/1", "", 500)
	if strings.TrimSpace(response.Body.String()) != `{"error":{"code":"INTERNAL_ERROR","message":"an internal error occurred"}}` {
		t.Fatalf("internal details exposed: %s", response.Body.String())
	}
}

func TestAPIRequiresJSONContentType(t *testing.T) {
	_, router := newAPI(t)
	body := `{"name":"Demo","email":"demo@example.com"}`
	for _, contentType := range []string{"", "text/plain", "application/xml"} {
		t.Run(contentType, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(body))
			req.Header.Set("Content-Type", contentType)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			if recorder.Code != 415 || !strings.Contains(recorder.Body.String(), `"code":"UNSUPPORTED_MEDIA_TYPE"`) {
				t.Fatalf("media type response = %d %s", recorder.Code, recorder.Body.String())
			}
		})
	}
	req := httptest.NewRequest("POST", "/api/v1/users?name=Query&email=query@example.com", strings.NewReader(body))
	req.Header.Set("Content-Type", "Application/JSON; charset=utf-8")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != 201 || !strings.Contains(recorder.Body.String(), `"email":"demo@example.com"`) {
		t.Fatalf("JSON body binding response = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestAPIPanicReturnsGenericError(t *testing.T) {
	_, router := newAPI(t)
	router.GET("/panic", func(c *echo.Context) error { panic("private panic detail") })
	response := request(t, router, "GET", "/panic", "", 500)
	if strings.TrimSpace(response.Body.String()) != `{"error":{"code":"INTERNAL_ERROR","message":"an internal error occurred"}}` {
		t.Fatalf("panic details exposed: %s", response.Body.String())
	}
}
