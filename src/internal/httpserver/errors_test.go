package httpserver

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gin-start/src/internal/user"

	"github.com/labstack/echo/v5"
)

func TestErrorHandlerPreservesStatusAndHidesInternalDetails(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"wrapped domain error", fmt.Errorf("lookup failed: %w", user.ErrNotFound), 404, "USER_NOT_FOUND"},
		{"wrapped conflict", fmt.Errorf("insert failed: %w", user.ErrEmailTaken), 409, "EMAIL_TAKEN"},
		{"streaming body limit", echo.ErrBadRequest.Wrap(echo.ErrStatusRequestEntityTooLarge), 413, "BODY_TOO_LARGE"},
		{"explicit body limit", echo.NewHTTPError(413, "private error detail"), 413, "BODY_TOO_LARGE"},
		{"unexpected error", errors.New("private error detail"), 500, "INTERNAL_ERROR"},
		{"framework internal error", echo.NewHTTPError(500, "private error detail"), 500, "INTERNAL_ERROR"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.HTTPErrorHandler = handleError
			e.GET("/error", func(c *echo.Context) error { return tt.err })
			recorder := httptest.NewRecorder()
			e.ServeHTTP(recorder, httptest.NewRequest("GET", "/error", nil))
			if recorder.Code != tt.status || !strings.Contains(recorder.Body.String(), `"code":"`+tt.code+`"`) {
				t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
			}
			if strings.Contains(recorder.Body.String(), "private error detail") {
				t.Fatalf("internal details exposed: %s", recorder.Body.String())
			}
		})
	}
}

func TestErrorHandlerDoesNotAppendToCommittedResponse(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = handleError
	e.GET("/committed", func(c *echo.Context) error {
		if err := c.String(http.StatusOK, "already sent"); err != nil {
			return err
		}
		return errors.New("late error")
	})
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest("GET", "/committed", nil))
	if recorder.Code != 200 || recorder.Body.String() != "already sent" {
		t.Fatalf("response was overwritten: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestErrorHandlerReturnsNoBodyForHEAD(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = handleError
	e.HEAD("/error", func(c *echo.Context) error { return user.ErrNotFound })
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest("HEAD", "/error", nil))
	if recorder.Code != 404 || recorder.Body.Len() != 0 {
		t.Fatalf("HEAD error response = %d %s", recorder.Code, recorder.Body.String())
	}
}
