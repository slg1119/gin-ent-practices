package userhttp

import (
	"mime"
	"net/http"
	"strconv"

	"gin-start/src/internal/user"

	"github.com/labstack/echo/v5"
)

type UserHandler struct {
	userService *user.UserService
}

func NewUserHandler(userService *user.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Create(c *echo.Context) error {
	// 생성 API는 JSON 본문만 받는다. 쿼리/경로 값을 DTO에 섞어 바인딩하지 않는다.
	mediaType, _, err := mime.ParseMediaType(c.Request().Header.Get(echo.HeaderContentType))
	if err != nil || mediaType != echo.MIMEApplicationJSON {
		return echo.ErrUnsupportedMediaType
	}
	if c.Request().ContentLength == 0 {
		return echo.ErrBadRequest
	}
	// MIME 대소문자와 charset 매개변수를 정규화해 Echo 바인더에 전달한다.
	c.Request().Header.Set(echo.HeaderContentType, mediaType)
	var request createRequest
	if err := echo.BindBody(c, &request); err != nil {
		return err
	}

	u, err := h.userService.Create(c.Request().Context(), request.Name, request.Email)
	if err != nil {
		// 업무 오류를 HTTP 상태로 바꾸는 작업은 공통 오류 처리기에 맡긴다.
		return err
	}
	c.Response().Header().Set("Location", "/api/v1/users/"+strconv.Itoa(u.ID))
	return c.JSON(http.StatusCreated, toResponse(u))
}

func (h *UserHandler) Get(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return user.ErrInvalidID
	}
	u, err := h.userService.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(u))
}

func (h *UserHandler) List(c *echo.Context) error {
	limit, limitErr := strconv.Atoi(queryOrDefault(c, "limit", "20"))
	offset, offsetErr := strconv.Atoi(queryOrDefault(c, "offset", "0"))
	if limitErr != nil || offsetErr != nil {
		return user.ErrInvalidPagination
	}
	users, err := h.userService.List(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	response := listResponse{Users: make([]userResponse, 0, len(users)), Limit: limit, Offset: offset}
	for _, u := range users {
		response.Users = append(response.Users, toResponse(u))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *UserHandler) Deactivate(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return user.ErrInvalidID
	}
	u, err := h.userService.Deactivate(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(u))
}

func queryOrDefault(c *echo.Context, name, fallback string) string {
	// 파라미터 생략과 빈 값(?limit=)을 구분한다. 빈 값은 입력 오류다.
	if !c.QueryParams().Has(name) {
		return fallback
	}
	return c.QueryParam(name)
}
