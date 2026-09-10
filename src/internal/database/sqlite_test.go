package database_test

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"gin-start/src/ent"
	"gin-start/src/internal/database"
)

func TestOpenPreservesUsersWhenAddingUpdatedAt(t *testing.T) {
	dsn := "file:" + filepath.Join(t.TempDir(), "legacy.db") + "?_fk=1"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	// mixin을 적용하기 전 스키마와 기존 데이터를 재현한다.
	for _, statement := range []string{
		`CREATE TABLE users (
			id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
			name text NOT NULL,
			email text NOT NULL,
			is_active bool NOT NULL DEFAULT true,
			created_at datetime NOT NULL
		)`,
		`CREATE UNIQUE INDEX users_email_key ON users (email)`,
		`INSERT INTO users (id, name, email, is_active, created_at)
		 VALUES (7, 'Existing', 'existing@example.com', false, '2020-01-01 00:00:00')`,
	} {
		if _, err := db.ExecContext(t.Context(), statement); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	migrationStarted := time.Now().UTC().Truncate(time.Second)
	client, err := database.Open(t.Context(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	u, err := client.User.Get(t.Context(), 7)
	if err != nil {
		t.Fatal(err)
	}
	createdAt := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	if u.Name != "Existing" || u.Email != "existing@example.com" || u.IsActive || !u.CreatedAt.Equal(createdAt) || u.UpdatedAt.IsZero() {
		t.Fatalf("legacy user was not preserved with timestamps: %+v", u)
	}
	if u.UpdatedAt.Before(migrationStarted) || u.UpdatedAt.After(time.Now()) {
		t.Fatalf("legacy updated_at = %v, want migration time", u.UpdatedAt)
	}
	if _, err := client.User.Create().SetName("Duplicate").SetEmail(u.Email).Save(t.Context()); !ent.IsConstraintError(err) {
		t.Fatalf("migration did not preserve unique email constraint: %v", err)
	}
	created, err := client.User.Create().SetName("New").SetEmail("new@example.com").Save(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if created.ID <= u.ID || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("new user after migration = %+v", created)
	}
}
