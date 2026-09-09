package httpserver

import (
	"errors"
	"net/http"

	"gin-start/src/internal/user"

	"github.com/labstack/echo/v5"
)

type errorResponse struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// handleError는 핸들러, 라우터, 미들웨어의 오류를 같은 JSON 형식으로 변환한다.
func handleError(c *echo.Context, err error) {
	if response, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil && response.Committed {
		return
	}

	status, detail := classifyError(err)
	var writeErr error
	if c.Request().Method == http.MethodHead {
		writeErr = c.NoContent(status)
	} else {
		writeErr = c.JSON(status, errorResponse{Error: detail})
	}
	if writeErr != nil {
		c.Logger().Error("failed to write error response", "error", writeErr)
	}
}

func classifyError(err error) (int, errorDetail) {
	switch {
	case errors.Is(err, user.ErrInvalidName):
		return http.StatusBadRequest, errorDetail{"INVALID_NAME", user.ErrInvalidName.Error()}
	case errors.Is(err, user.ErrInvalidEmail):
		return http.StatusBadRequest, errorDetail{"INVALID_EMAIL", user.ErrInvalidEmail.Error()}
	case errors.Is(err, user.ErrInvalidID):
		return http.StatusBadRequest, errorDetail{"INVALID_ID", user.ErrInvalidID.Error()}
	case errors.Is(err, user.ErrInvalidPagination):
		return http.StatusBadRequest, errorDetail{"INVALID_PAGINATION", user.ErrInvalidPagination.Error()}
	case errors.Is(err, user.ErrNotFound):
		return http.StatusNotFound, errorDetail{"USER_NOT_FOUND", user.ErrNotFound.Error()}
	case errors.Is(err, user.ErrEmailTaken):
		return http.StatusConflict, errorDetail{"EMAIL_TAKEN", user.ErrEmailTaken.Error()}
	case errors.Is(err, echo.ErrStatusRequestEntityTooLarge):
		// 길이 미상의 본문을 읽다가 발생한 413은 바인더의 400 오류 안에 감싸질 수 있다.
		return http.StatusRequestEntityTooLarge, errorDetail{"BODY_TOO_LARGE", "request body must not exceed 1 MiB"}
	}

	var httpErr echo.HTTPStatusCoder
	if errors.As(err, &httpErr) {
		status := httpErr.StatusCode()
		switch status {
		case http.StatusBadRequest:
			return status, errorDetail{"INVALID_BODY", "body must contain valid JSON with string name and email fields"}
		case http.StatusNotFound:
			return status, errorDetail{"ROUTE_NOT_FOUND", "route not found"}
		case http.StatusMethodNotAllowed:
			return status, errorDetail{"METHOD_NOT_ALLOWED", "method not allowed"}
		case http.StatusRequestEntityTooLarge:
			return status, errorDetail{"BODY_TOO_LARGE", "request body must not exceed 1 MiB"}
		case http.StatusUnsupportedMediaType:
			return status, errorDetail{"UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json"}
		}
		if status >= 400 && status < 500 {
			return status, errorDetail{"REQUEST_ERROR", http.StatusText(status)}
		}
	}
	// DB 오류, panic, Echo의 내부 오류 메시지를 응답에 그대로 노출하지 않는다.
	return http.StatusInternalServerError, errorDetail{"INTERNAL_ERROR", "an internal error occurred"}
}
