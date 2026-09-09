package entrepo_test

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"gin-start/src/ent"
	"gin-start/src/internal/database"
	"gin-start/src/internal/user"
	"gin-start/src/internal/user/entrepo"
)

func newRepository(t *testing.T) (*ent.Client, *entrepo.UserRepository) {
	t.Helper()
	client, err := database.Open(t.Context(), "file:"+filepath.Join(t.TempDir(), "test.db")+"?_fk=1")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	return client, entrepo.NewUserRepository(client)
}

func TestRepositoryLifecycle(t *testing.T) {
	_, repo := newRepository(t)
	ctx := t.Context()
	empty, err := repo.List(ctx, 20, 0)
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("empty list = %v, error = %v", empty, err)
	}
	var created []*user.User
	for i := range 3 {
		u, err := repo.Create(ctx, &user.User{Name: "Demo", Email: fmt.Sprintf("demo%d@example.com", i), IsActive: true})
		if err != nil {
			t.Fatal(err)
		}
		if u.ID <= 0 || u.CreatedAt.IsZero() || !u.IsActive {
			t.Fatalf("generated values missing: %+v", u)
		}
		created = append(created, u)
	}
	page, err := repo.List(ctx, 1, 1)
	if err != nil || len(page) != 1 || page[0].ID != created[1].ID {
		t.Fatalf("ordered page = %+v, error = %v", page, err)
	}
	for range 2 {
		u, err := repo.Deactivate(ctx, created[0].ID)
		if err != nil || u.IsActive {
			t.Fatalf("deactivation failed: %+v, %v", u, err)
		}
	}
	got, err := repo.FindByID(ctx, created[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.IsActive || got.Name != created[0].Name || got.Email != created[0].Email || !got.CreatedAt.Equal(created[0].CreatedAt) {
		t.Fatalf("deactivation changed unrelated fields or was not persisted: %+v", got)
	}
	for _, operation := range []func(context.Context, int) (*user.User, error){repo.FindByID, repo.Deactivate} {
		if _, err := operation(ctx, 999999); !errors.Is(err, user.ErrNotFound) {
			t.Fatalf("missing ID error = %v", err)
		}
	}
}

func TestConcurrentDuplicateEmailUsesDatabaseConstraint(t *testing.T) {
	_, repo := newRepository(t)
	svc := user.NewUserService(repo)
	ctx := t.Context()
	const attempts = 8
	start := make(chan struct{})
	results := make(chan error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			email := "same@example.com"
			if i%2 == 0 {
				email = " SAME@EXAMPLE.COM "
			}
			_, err := svc.Create(ctx, "Demo", email)
			results <- err
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)
	success, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			success++
		case errors.Is(err, user.ErrEmailTaken):
			conflicts++
		default:
			t.Fatalf("unexpected insert error: %v", err)
		}
	}
	rows, err := repo.List(ctx, 20, 0)
	if err != nil || success != 1 || conflicts != attempts-1 || len(rows) != 1 {
		t.Fatalf("success=%d conflicts=%d rows=%d error=%v", success, conflicts, len(rows), err)
	}
}

func TestRepositoryPreservesCancellationAndDoesNotMislabelDatabaseFailure(t *testing.T) {
	client, repo := newRepository(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := repo.FindByID(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled lookup error = %v", err)
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := repo.Create(t.Context(), &user.User{Name: "Demo", Email: "demo@example.com", IsActive: true})
	if err == nil || errors.Is(err, user.ErrEmailTaken) {
		t.Fatalf("closed DB error incorrectly translated: %v", err)
	}
}

func TestDataSurvivesReopen(t *testing.T) {
	dsn := "file:" + filepath.Join(t.TempDir(), "persistent.db") + "?_fk=1"
	client, err := database.Open(t.Context(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	u, err := user.NewUserService(entrepo.NewUserRepository(client)).Create(t.Context(), "Demo", "demo@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := database.Open(t.Context(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { reopened.Close() })
	got, err := entrepo.NewUserRepository(reopened).FindByID(t.Context(), u.ID)
	if err != nil || got.Email != "demo@example.com" {
		t.Fatalf("reopened user = %+v, error = %v", got, err)
	}
}
