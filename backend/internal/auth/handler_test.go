package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

type mockAuthSvc struct {
	loginFn      func(username, password string) (*Session, *User, error)
	getSessionFn func(token string) (*Session, *User, error)
	logoutFn     func(token string) error
}

func (m *mockAuthSvc) Login(u, p string) (*Session, *User, error) { return m.loginFn(u, p) }
func (m *mockAuthSvc) GetSession(token string) (*Session, *User, error) {
	return m.getSessionFn(token)
}
func (m *mockAuthSvc) Logout(token string) error { return m.logoutFn(token) }

func newAuthRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.POST("/auth/login", h.Login)
	r.POST("/auth/logout", h.Logout)
	r.GET("/auth/me", h.Me)
	r.Use(h.Middleware())
	return r
}

func authReq(r *gin.Engine, method, path string, body any, token string) *httptest.ResponseRecorder {
	var buf *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func validSession() (*Session, *User) {
	return &Session{Token: "tok123", CreatedAt: time.Now()}, &User{ID: 1, Username: "alice"}
}

// ── Login ─────────────────────────────────────────────────────────────────────

func TestAuthHandlerLogin_OK(t *testing.T) {
	svc := &mockAuthSvc{loginFn: func(_, _ string) (*Session, *User, error) {
		s, usr := validSession()
		return s, usr, nil
	}}
	r := newAuthRouter(&Handler{service: svc})
	w := authReq(r, http.MethodPost, "/auth/login", map[string]any{
		"username": "alice", "password": "secret",
	}, "")
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] != "tok123" {
		t.Errorf("token = %v, want tok123", resp["token"])
	}
}

func TestAuthHandlerLogin_BadJSON(t *testing.T) {
	svc := &mockAuthSvc{}
	r := newAuthRouter(&Handler{service: svc})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestAuthHandlerLogin_InvalidCredentials(t *testing.T) {
	svc := &mockAuthSvc{loginFn: func(u, p string) (*Session, *User, error) {
		return nil, nil, errors.New("credenciais inválidas")
	}}
	r := newAuthRouter(&Handler{service: svc})
	w := authReq(r, http.MethodPost, "/auth/login", map[string]any{
		"username": "alice", "password": "wrong",
	}, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// ── Logout ────────────────────────────────────────────────────────────────────

func TestAuthHandlerLogout_WithToken(t *testing.T) {
	var loggedOut string
	svc := &mockAuthSvc{logoutFn: func(token string) error { loggedOut = token; return nil }}
	r := newAuthRouter(&Handler{service: svc})
	w := authReq(r, http.MethodPost, "/auth/logout", nil, "tok123")
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if loggedOut != "tok123" {
		t.Errorf("logged out token = %q, want tok123", loggedOut)
	}
}

func TestAuthHandlerLogout_NoToken(t *testing.T) {
	svc := &mockAuthSvc{}
	r := newAuthRouter(&Handler{service: svc})
	w := authReq(r, http.MethodPost, "/auth/logout", nil, "")
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// ── Me ────────────────────────────────────────────────────────────────────────

func TestAuthHandlerMe_OK(t *testing.T) {
	sess, user := validSession()
	svc := &mockAuthSvc{getSessionFn: func(token string) (*Session, *User, error) {
		return sess, user, nil
	}}
	r := newAuthRouter(&Handler{service: svc})
	w := authReq(r, http.MethodGet, "/auth/me", nil, "tok123")
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestAuthHandlerMe_NoToken(t *testing.T) {
	svc := &mockAuthSvc{}
	r := newAuthRouter(&Handler{service: svc})
	w := authReq(r, http.MethodGet, "/auth/me", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuthHandlerMe_InvalidSession(t *testing.T) {
	svc := &mockAuthSvc{getSessionFn: func(token string) (*Session, *User, error) {
		return nil, nil, nil
	}}
	r := newAuthRouter(&Handler{service: svc})
	w := authReq(r, http.MethodGet, "/auth/me", nil, "badtoken")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuthHandlerMe_SessionError(t *testing.T) {
	svc := &mockAuthSvc{getSessionFn: func(token string) (*Session, *User, error) {
		return nil, nil, errors.New("db error")
	}}
	r := newAuthRouter(&Handler{service: svc})
	w := authReq(r, http.MethodGet, "/auth/me", nil, "tok")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// ── Middleware ────────────────────────────────────────────────────────────────

func TestAuthMiddleware_ValidToken(t *testing.T) {
	sess, user := validSession()
	svc := &mockAuthSvc{getSessionFn: func(token string) (*Session, *User, error) {
		return sess, user, nil
	}}
	h := &Handler{service: svc}

	r := gin.New()
	r.Use(h.Middleware())
	r.GET("/protected", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer validtoken")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	h := &Handler{service: &mockAuthSvc{}}

	r := gin.New()
	r.Use(h.Middleware())
	r.GET("/protected", func(c *gin.Context) { c.JSON(http.StatusOK, nil) })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_InvalidSession(t *testing.T) {
	svc := &mockAuthSvc{getSessionFn: func(token string) (*Session, *User, error) {
		return nil, nil, nil
	}}
	h := &Handler{service: svc}

	r := gin.New()
	r.Use(h.Middleware())
	r.GET("/protected", func(c *gin.Context) { c.JSON(http.StatusOK, nil) })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer expired")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// ── extractToken ──────────────────────────────────────────────────────────────

func TestExtractToken_Bearer(t *testing.T) {
	r := gin.New()
	var captured string
	r.GET("/test", func(c *gin.Context) {
		captured = extractToken(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer mytoken123")
	httptest.NewRecorder()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if captured != "mytoken123" {
		t.Errorf("extractToken = %q, want mytoken123", captured)
	}
}

func TestExtractToken_NoHeader(t *testing.T) {
	r := gin.New()
	var captured string
	r.GET("/test", func(c *gin.Context) {
		captured = extractToken(c)
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if captured != "" {
		t.Errorf("expected empty token, got %q", captured)
	}
}
