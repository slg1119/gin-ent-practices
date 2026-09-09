package user_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"gin-start/src/internal/user"
)

func TestNewUserNormalizesAndValidates(t *testing.T) {
	tests := []struct {
		name, inputName, email string
		wantErr                error
	}{
		{"normalization", "  테스트  ", "  DEMO@EXAMPLE.COM  ", nil},
		{"unicode boundary", strings.Repeat("가", 100), "demo@example.com", nil},
		{"blank name", " \t\n ", "demo@example.com", user.ErrInvalidName},
		{"long name", strings.Repeat("가", 101), "demo@example.com", user.ErrInvalidName},
		{"invalid utf8", string([]byte{0xff}), "demo@example.com", user.ErrInvalidName},
		{"missing email", "Demo", "", user.ErrInvalidEmail},
		{"invalid email", "Demo", "invalid", user.ErrInvalidEmail},
		{"display name", "Demo", "Demo <demo@example.com>", user.ErrInvalidEmail},
		{"multiple addresses", "Demo", "a@example.com,b@example.com", user.ErrInvalidEmail},
		{"long email", "Demo", strings.Repeat("a", 243) + "@example.com", user.ErrInvalidEmail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := user.NewUser(tt.inputName, tt.email)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if err == nil {
				if u.Name != strings.TrimSpace(tt.inputName) || u.Email != "demo@example.com" || !u.IsActive {
					t.Fatalf("unexpected normalized user: %+v", u)
				}
			}
		})
	}
}

// 유효하지 않은 입력은 저장소에 도달하면 안 된다. 호출되면 nil 함수로 테스트가 실패한다.
type stubRepository struct {
	create     func(context.Context, *user.User) (*user.User, error)
	find       func(context.Context, int) (*user.User, error)
	list       func(context.Context, int, int) ([]*user.User, error)
	deactivate func(context.Context, int) (*user.User, error)
}

func (s stubRepository) Create(ctx context.Context, u *user.User) (*user.User, error) {
	return s.create(ctx, u)
}
func (s stubRepository) FindByID(ctx context.Context, id int) (*user.User, error) {
	return s.find(ctx, id)
}
func (s stubRepository) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	return s.list(ctx, limit, offset)
}
func (s stubRepository) Deactivate(ctx context.Context, id int) (*user.User, error) {
	return s.deactivate(ctx, id)
}

func TestServiceRejectsInvalidInputBeforePersistence(t *testing.T) {
	svc := user.NewUserService(stubRepository{})
	ctx := t.Context()
	if _, err := svc.Create(ctx, " ", "demo@example.com"); !errors.Is(err, user.ErrInvalidName) {
		t.Fatalf("Create error = %v", err)
	}
	if _, err := svc.Create(ctx, "Demo", "invalid"); !errors.Is(err, user.ErrInvalidEmail) {
		t.Fatalf("Create error = %v", err)
	}
	for _, id := range []int{0, -1} {
		if _, err := svc.Get(ctx, id); !errors.Is(err, user.ErrInvalidID) {
			t.Fatalf("Get error = %v", err)
		}
		if _, err := svc.Deactivate(ctx, id); !errors.Is(err, user.ErrInvalidID) {
			t.Fatalf("Deactivate error = %v", err)
		}
	}
	for _, page := range [][2]int{{0, 0}, {101, 0}, {20, -1}} {
		if _, err := svc.List(ctx, page[0], page[1]); !errors.Is(err, user.ErrInvalidPagination) {
			t.Fatalf("List error = %v", err)
		}
	}
}

func TestServicePreservesContextAndRepositoryErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	called := false
	svc := user.NewUserService(stubRepository{create: func(got context.Context, u *user.User) (*user.User, error) {
		called = true
		if got != ctx || u.Name != "Demo" || u.Email != "demo@example.com" || !u.IsActive {
			t.Fatal("context or normalized user was not forwarded")
		}
		return nil, got.Err()
	}})
	if _, err := svc.Create(ctx, " Demo ", "DEMO@EXAMPLE.COM"); !errors.Is(err, context.Canceled) || !called {
		t.Fatalf("error = %v, repository called = %v", err, called)
	}
}
