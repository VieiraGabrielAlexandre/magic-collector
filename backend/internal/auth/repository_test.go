package auth

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAuthRepoGetUserByUsername_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "username", "display_name", "password_hash", "created_at"}).
		AddRow(1, "alice", "Alice", "$2a$10$hash", now)
	mock.ExpectQuery("SELECT").WithArgs("alice").WillReturnRows(rows)

	repo := NewRepository(db)
	u, err := repo.GetUserByUsername("alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u == nil {
		t.Fatal("expected user, got nil")
	}
	if u.Username != "alice" {
		t.Errorf("Username = %q, want alice", u.Username)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAuthRepoGetUserByUsername_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "username", "display_name", "password_hash", "created_at"})
	mock.ExpectQuery("SELECT").WithArgs("unknown").WillReturnRows(rows)

	repo := NewRepository(db)
	u, err := repo.GetUserByUsername("unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u != nil {
		t.Error("expected nil for unknown user")
	}
}

func TestAuthRepoCreateSession_OK(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectExec("INSERT INTO sessions").
		WithArgs("tok123", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewRepository(db)
	if err := repo.CreateSession("tok123", 1); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAuthRepoCreateSession_Error(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectExec("INSERT INTO sessions").WillReturnError(sqlmock.ErrCancelled)

	repo := NewRepository(db)
	if err := repo.CreateSession("bad", 1); err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthRepoGetSessionWithUser_OK(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"token", "user_id", "created_at", "expires_at",
		"id", "username", "display_name", "created_at",
	}).AddRow("tok123", 1, now, now.Add(720*time.Hour), 1, "alice", "Alice", now)

	mock.ExpectQuery("SELECT").WithArgs("tok123").WillReturnRows(rows)

	repo := NewRepository(db)
	sess, user, err := repo.GetSessionWithUser("tok123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess == nil || user == nil {
		t.Fatal("expected session and user")
	}
	if sess.Token != "tok123" {
		t.Errorf("Token = %q, want tok123", sess.Token)
	}
	if user.Username != "alice" {
		t.Errorf("Username = %q, want alice", user.Username)
	}
}

func TestAuthRepoGetSessionWithUser_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"token", "user_id", "created_at", "expires_at",
		"id", "username", "display_name", "created_at",
	})
	mock.ExpectQuery("SELECT").WithArgs("expired").WillReturnRows(rows)

	repo := NewRepository(db)
	sess, user, err := repo.GetSessionWithUser("expired")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess != nil || user != nil {
		t.Error("expected nil for expired/missing session")
	}
}

func TestAuthRepoGetSessionWithUser_DBError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT").WithArgs("bad").WillReturnError(sql.ErrConnDone)

	repo := NewRepository(db)
	_, _, err := repo.GetSessionWithUser("bad")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthRepoDeleteSession_OK(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectExec("DELETE FROM sessions").
		WithArgs("tok123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewRepository(db)
	if err := repo.DeleteSession("tok123"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
