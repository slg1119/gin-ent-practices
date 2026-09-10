// Package entrepo는 회원 저장소를 Ent와 SQLite로 구현한다.
package entrepo

import (
	"context"
	"errors"
	"fmt"

	"gin-start/src/ent"
	entuser "gin-start/src/ent/user"
	"gin-start/src/internal/user"

	"github.com/mattn/go-sqlite3"
)

type UserRepository struct {
	client *ent.Client
}

var _ user.UserRepository = (*UserRepository)(nil)

func NewUserRepository(client *ent.Client) *UserRepository {
	return &UserRepository{client: client}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) (*user.User, error) {
	row, err := r.client.User.Create().
		SetName(u.Name).
		SetEmail(u.Email).
		SetIsActive(u.IsActive).
		Save(ctx)
	if err != nil {
		// 이 스키마에서 사용자가 입력하는 unique 필드는 email 하나다.
		// 다른 제약 위반이나 연결 오류까지 이메일 중복으로 처리하지 않는다.
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return nil, user.ErrEmailTaken
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return toDomain(row), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int) (*user.User, error) {
	row, err := r.client.User.Get(ctx, id)
	if ent.IsNotFound(err) {
		return nil, user.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return toDomain(row), nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	rows, err := r.client.User.Query().
		Order(ent.Asc(entuser.FieldID)).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	users := make([]*user.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, toDomain(row))
	}
	return users, nil
}

func (r *UserRepository) Deactivate(ctx context.Context, id int) (*user.User, error) {
	// 상태와 mixin의 수정 시각을 갱신하고 이름/이메일을 덮어쓰지 않는다.
	row, err := r.client.User.UpdateOneID(id).SetIsActive(false).Save(ctx)
	if ent.IsNotFound(err) {
		return nil, user.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("deactivate user: %w", err)
	}
	return toDomain(row), nil
}
