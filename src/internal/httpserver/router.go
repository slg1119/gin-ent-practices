// Package httpserver는 Echo 라우터, 미들웨어, 공통 HTTP 오류 처리를 구성한다.
package httpserver

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// New는 공통 미들웨어와 상태 확인 라우트를 구성한다.
// 기능별 라우트는 bootstrap에서 등록한다.
func New() *echo.Echo {
	e := echo.New()
	e.IPExtractor = echo.ExtractIPDirect()
	e.HTTPErrorHandler = handleError
	e.Use(middleware.RequestLogger(), middleware.Recover(), middleware.BodyLimit(1<<20))
	e.GET("/healthz", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	return e
}
