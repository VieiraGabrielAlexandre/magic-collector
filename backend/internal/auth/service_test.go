package auth

import (
	"errors"
	"testing"
	"time"
)

// ── mock repository ──────────────────────────────────────────────────────────

type mockAuthRepo struct {
	getUserFn         func(username string) (*userWithHash, error)
	createSessionFn   func(token string, userID int) error
	getSessionFn      func(token string) (*Session, *User, error)
	deleteSessionFn   func(token string) error
}

func (m *mockAuthRepo) GetUserByUsername(u string) (*userWithHash, error) { return m.getUserFn(u) }
func (m *mockAuthRepo) CreateSession(token string, userID int) error      { return m.createSessionFn(token, userID) }
func (m *mockAuthRepo) GetSessionWithUser(token string) (*Session, *User, error) {
	return m.getSessionFn(token)
}
func (m *mockAuthRepo) DeleteSession(token string) error { return m.deleteSessionFn(token) }

// ── HashPassword ──────────────────────────────────────────────────────────────

func TestHashPassword_Roundtrip(t *testing.T) {
	hash, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == "mypassword" {
		t.Fatal("hash should differ from plaintext")
	}
}

func TestHashPassword_DifferentForSameInput(t *testing.T) {
	h1, _ := HashPassword("password")
	h2, _ := HashPassword("password")
	if h1 == h2 {
		t.Error("bcrypt should produce different hashes for same input (salt)")
	}
}

// ── generateToken ─────────────────────────────────────────────────────────────

func TestGenerateToken_Length(t *testing.T) {
	tok, err := generateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 32 bytes → 64 hex chars
	if len(tok) != 64 {
		t.Errorf("token length = %d, want 64", len(tok))
	}
}

func TestGenerateToken_Unique(t *testing.T) {
	tok1, _ := generateToken()
	tok2, _ := generateToken()
	if tok1 == tok2 {
		t.Error("tokens should be unique")
	}
}

// ── Service.Login ─────────────────────────────────────────────────────────────

func TestServiceLogin_Success(t *testing.T) {
	hash, _ := HashPassword("secret")
	now := time.Now()
	repo := &mockAuthRepo{
		getUserFn: func(u string) (*userWithHash, error) {
			return &userWithHash{
				User:         User{ID: 1, Username: u},
				PasswordHash: hash,
			}, nil
		},
		createSessionFn: func(token string, userID int) error { return nil },
		getSessionFn: func(token string) (*Session, *User, error) {
			return &Session{Token: token, CreatedAt: now}, &User{ID: 1, Username: "alice"}, nil
		},
	}
	svc := &Service{repository: repo}
	sess, user, err := svc.Login("alice", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess == nil || user == nil {
		t.Fatal("expected session and user")
	}
	if user.Username != "alice" {
		t.Errorf("Username = %q, want alice", user.Username)
	}
}

func TestServiceLogin_UserNotFound(t *testing.T) {
	repo := &mockAuthRepo{
		getUserFn: func(u string) (*userWithHash, error) { return nil, nil },
	}
	svc := &Service{repository: repo}
	_, _, err := svc.Login("unknown", "pass")
	if err == nil {
		t.Fatal("expected error for unknown user")
	}
}

func TestServiceLogin_WrongPassword(t *testing.T) {
	hash, _ := HashPassword("correct")
	repo := &mockAuthRepo{
		getUserFn: func(u string) (*userWithHash, error) {
			return &userWithHash{User: User{ID: 1}, PasswordHash: hash}, nil
		},
	}
	svc := &Service{repository: repo}
	_, _, err := svc.Login("alice", "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestServiceLogin_RepositoryError(t *testing.T) {
	repo := &mockAuthRepo{
		getUserFn: func(u string) (*userWithHash, error) { return nil, errors.New("db error") },
	}
	svc := &Service{repository: repo}
	_, _, err := svc.Login("alice", "pass")
	if err == nil {
		t.Fatal("expected error when repo fails")
	}
}

func TestServiceLogin_CreateSessionError(t *testing.T) {
	hash, _ := HashPassword("pass")
	repo := &mockAuthRepo{
		getUserFn: func(u string) (*userWithHash, error) {
			return &userWithHash{User: User{ID: 1}, PasswordHash: hash}, nil
		},
		createSessionFn: func(token string, userID int) error { return errors.New("session error") },
	}
	svc := &Service{repository: repo}
	_, _, err := svc.Login("alice", "pass")
	if err == nil {
		t.Fatal("expected error when CreateSession fails")
	}
}

// ── Service.GetSession ────────────────────────────────────────────────────────

func TestServiceGetSession(t *testing.T) {
	now := time.Now()
	repo := &mockAuthRepo{
		getSessionFn: func(token string) (*Session, *User, error) {
			return &Session{Token: token, CreatedAt: now}, &User{ID: 1}, nil
		},
	}
	svc := &Service{repository: repo}
	sess, user, err := svc.GetSession("mytoken")
	if err != nil || sess == nil || user == nil {
		t.Fatal("expected valid session")
	}
}

// ── Service.Logout ────────────────────────────────────────────────────────────

func TestServiceLogout_OK(t *testing.T) {
	var deletedToken string
	repo := &mockAuthRepo{
		deleteSessionFn: func(token string) error {
			deletedToken = token
			return nil
		},
	}
	svc := &Service{repository: repo}
	err := svc.Logout("mytoken")
	if err != nil {
		t.Fatal(err)
	}
	if deletedToken != "mytoken" {
		t.Errorf("deleted token = %q, want mytoken", deletedToken)
	}
}
