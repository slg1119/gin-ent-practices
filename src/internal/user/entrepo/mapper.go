package entrepo

import (
	"gin-start/src/ent"
	"gin-start/src/internal/user"
)

func toDomain(row *ent.User) *user.User {
	return &user.User{
		ID:        row.ID,
		Name:      row.Name,
		Email:     row.Email,
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
