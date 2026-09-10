package userhttp

import (
	"time"

	"gin-start/src/internal/user"
)

type createRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// 응답 필드를 명시해 DB 컬럼 추가가 API에 자동으로 노출되지 않게 한다.
type userResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type listResponse struct {
	Users  []userResponse `json:"users"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

func toResponse(u *user.User) userResponse {
	return userResponse{
		ID: u.ID, Name: u.Name, Email: u.Email,
		IsActive: u.IsActive, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}
}
