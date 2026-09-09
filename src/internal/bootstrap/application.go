// Package bootstrap은 기능별 모듈과 HTTP 라우터를 조립한다.
package bootstrap

import (
	"gin-start/src/ent"
	"gin-start/src/internal/httpserver"

	"github.com/labstack/echo/v5"
)

type Application struct {
	Router *echo.Echo
}

// NewApplication은 모듈을 만들고 외부에 제공할 라우트를 연결한다.
// DB 연결의 생성·종료와 HTTP 서버의 시작·종료는 main에서 관리한다.
func NewApplication(entClient *ent.Client) *Application {
	userModule := NewUserModule(entClient)

	router := httpserver.New()
	api := router.Group("/api/v1")
	userModule.UserHandler.RegisterRoutes(api)

	return &Application{Router: router}
}
