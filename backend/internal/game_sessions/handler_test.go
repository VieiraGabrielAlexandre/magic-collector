package game_sessions

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// ── mock service ─────────────────────────────────────────────────────────────

type mockGSSvc struct {
	listFn          func() ([]GameSession, error)
	getByIDFn       func(id int64) (*GameSession, error)
	createFn        func(input CreateSessionInput) (*GameSession, error)
	deleteFn        func(id int64) error
	addPlayerFn     func(sessionID int64, input PlayerInput) (*Player, error)
	updatePlayerFn  func(sessionID, playerID int64, input UpdatePlayerInput) (*Player, error)
	deletePlayerFn  func(sessionID, playerID int64) error
	resetFn         func(sessionID int64) (*GameSession, error)
	finishFn        func(sessionID int64) (*GameSession, error)
	restoreFn       func(sessionID int64) (*GameSession, error)
}

func (m *mockGSSvc) List() ([]GameSession, error)          { return m.listFn() }
func (m *mockGSSvc) GetByID(id int64) (*GameSession, error)    { return m.getByIDFn(id) }
func (m *mockGSSvc) Create(i CreateSessionInput) (*GameSession, error) { return m.createFn(i) }
func (m *mockGSSvc) Delete(id int64) error                     { return m.deleteFn(id) }
func (m *mockGSSvc) AddPlayer(sid int64, i PlayerInput) (*Player, error) { return m.addPlayerFn(sid, i) }
func (m *mockGSSvc) UpdatePlayer(sid, pid int64, i UpdatePlayerInput) (*Player, error) {
	return m.updatePlayerFn(sid, pid, i)
}
func (m *mockGSSvc) DeletePlayer(sid, pid int64) error { return m.deletePlayerFn(sid, pid) }
func (m *mockGSSvc) Reset(sid int64) (*GameSession, error)   { return m.resetFn(sid) }
func (m *mockGSSvc) Finish(sid int64) (*GameSession, error)  { return m.finishFn(sid) }
func (m *mockGSSvc) Restore(sid int64) (*GameSession, error) { return m.restoreFn(sid) }

func newGSRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.GET("/game-sessions", h.List)
	r.POST("/game-sessions", h.Create)
	r.GET("/game-sessions/:id", h.GetByID)
	r.DELETE("/game-sessions/:id", h.Delete)
	r.POST("/game-sessions/:id/players", h.AddPlayer)
	r.PATCH("/game-sessions/:id/players/:player_id", h.UpdatePlayer)
	r.DELETE("/game-sessions/:id/players/:player_id", h.DeletePlayer)
	r.POST("/game-sessions/:id/reset", h.Reset)
	r.POST("/game-sessions/:id/finish", h.Finish)
	r.POST("/game-sessions/:id/restore", h.Restore)
	return r
}

func gsReq(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var buf *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func okSession() *GameSession { return &GameSession{ID: 1, Status: "active"} }

// ── List ──────────────────────────────────────────────────────────────────────

func TestGSHandlerList_OK(t *testing.T) {
	svc := &mockGSSvc{listFn: func() ([]GameSession, error) { return []GameSession{{ID: 1}}, nil }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodGet, "/game-sessions", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestGSHandlerList_Error(t *testing.T) {
	svc := &mockGSSvc{listFn: func() ([]GameSession, error) { return nil, errors.New("db error") }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodGet, "/game-sessions", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── GetByID ──────────────────────────────────────────────────────────────────

func TestGSHandlerGetByID_OK(t *testing.T) {
	svc := &mockGSSvc{getByIDFn: func(id int64) (*GameSession, error) { return okSession(), nil }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodGet, "/game-sessions/1", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestGSHandlerGetByID_Nil(t *testing.T) {
	svc := &mockGSSvc{getByIDFn: func(id int64) (*GameSession, error) { return nil, nil }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodGet, "/game-sessions/999", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestGSHandlerGetByID_Error(t *testing.T) {
	svc := &mockGSSvc{getByIDFn: func(id int64) (*GameSession, error) { return nil, errors.New("fail") }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodGet, "/game-sessions/1", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestGSHandlerCreate_OK(t *testing.T) {
	svc := &mockGSSvc{createFn: func(i CreateSessionInput) (*GameSession, error) { return okSession(), nil }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodPost, "/game-sessions", map[string]any{
		"name":    "Test",
		"players": []map[string]any{{"name": "Alice", "short_code": "A"}, {"name": "Bob", "short_code": "B"}},
	})
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
}

func TestGSHandlerCreate_ValidationError(t *testing.T) {
	svc := &mockGSSvc{createFn: func(i CreateSessionInput) (*GameSession, error) {
		return nil, errors.New("mínimo de 2 jogadores")
	}}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodPost, "/game-sessions", map[string]any{
		"name":    "Test",
		"players": []map[string]any{{"name": "Solo", "short_code": "S"}},
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestGSHandlerDelete_OK(t *testing.T) {
	svc := &mockGSSvc{deleteFn: func(id int64) error { return nil }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodDelete, "/game-sessions/1", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestGSHandlerDelete_Error(t *testing.T) {
	svc := &mockGSSvc{deleteFn: func(id int64) error { return errors.New("fail") }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodDelete, "/game-sessions/1", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── AddPlayer ────────────────────────────────────────────────────────────────

func TestGSHandlerAddPlayer_OK(t *testing.T) {
	svc := &mockGSSvc{addPlayerFn: func(sid int64, i PlayerInput) (*Player, error) {
		return &Player{ID: 5, Name: i.Name}, nil
	}}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodPost, "/game-sessions/1/players", map[string]any{
		"name": "Carol", "short_code": "C",
	})
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
}

func TestGSHandlerAddPlayer_Error(t *testing.T) {
	svc := &mockGSSvc{addPlayerFn: func(sid int64, i PlayerInput) (*Player, error) {
		return nil, errors.New("max players")
	}}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodPost, "/game-sessions/1/players", map[string]any{
		"name": "Carol", "short_code": "C",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ── UpdatePlayer ──────────────────────────────────────────────────────────────

func TestGSHandlerUpdatePlayer_OK(t *testing.T) {
	svc := &mockGSSvc{updatePlayerFn: func(sid, pid int64, i UpdatePlayerInput) (*Player, error) {
		return &Player{ID: pid, Life: i.Life}, nil
	}}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodPatch, "/game-sessions/1/players/10", map[string]any{"life": 35})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// ── DeletePlayer ──────────────────────────────────────────────────────────────

func TestGSHandlerDeletePlayer_OK(t *testing.T) {
	svc := &mockGSSvc{deletePlayerFn: func(sid, pid int64) error { return nil }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodDelete, "/game-sessions/1/players/10", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestGSHandlerDeletePlayer_Error(t *testing.T) {
	svc := &mockGSSvc{deletePlayerFn: func(sid, pid int64) error { return errors.New("min players") }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodDelete, "/game-sessions/1/players/10", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ── Reset / Finish / Restore ──────────────────────────────────────────────────

func TestGSHandlerReset_OK(t *testing.T) {
	svc := &mockGSSvc{resetFn: func(sid int64) (*GameSession, error) { return okSession(), nil }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodPost, "/game-sessions/1/reset", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestGSHandlerReset_Error(t *testing.T) {
	svc := &mockGSSvc{resetFn: func(sid int64) (*GameSession, error) { return nil, errors.New("finished") }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodPost, "/game-sessions/1/reset", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestGSHandlerFinish_OK(t *testing.T) {
	svc := &mockGSSvc{finishFn: func(sid int64) (*GameSession, error) { return okSession(), nil }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodPost, "/game-sessions/1/finish", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestGSHandlerFinish_Error(t *testing.T) {
	svc := &mockGSSvc{finishFn: func(sid int64) (*GameSession, error) { return nil, errors.New("already finished") }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodPost, "/game-sessions/1/finish", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestGSHandlerRestore_OK(t *testing.T) {
	svc := &mockGSSvc{restoreFn: func(sid int64) (*GameSession, error) { return okSession(), nil }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodPost, "/game-sessions/1/restore", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestGSHandlerRestore_Error(t *testing.T) {
	svc := &mockGSSvc{restoreFn: func(sid int64) (*GameSession, error) { return nil, errors.New("fail") }}
	r := newGSRouter(&Handler{svc: svc})
	w := gsReq(r, http.MethodPost, "/game-sessions/1/restore", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
