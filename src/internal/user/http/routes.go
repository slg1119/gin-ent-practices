package userhttp

import "github.com/labstack/echo/v5"

func (h *UserHandler) RegisterRoutes(api *echo.Group) {
	users := api.Group("/users")
	users.POST("", h.Create)
	users.GET("", h.List)
	users.GET("/:id", h.Get)
	users.PATCH("/:id/deactivate", h.Deactivate)
}
