package battles

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

type mockBattleRepo struct {
	listFn   func() ([]Battle, error)
	createFn func(b BattleInput) (int64, error)
	deleteFn func(id string) error
}

func (m *mockBattleRepo) List() ([]Battle, error)          { return m.listFn() }
func (m *mockBattleRepo) Create(b BattleInput) (int64, error) { return m.createFn(b) }
func (m *mockBattleRepo) Delete(id string) error              { return m.deleteFn(id) }

func newBattleRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.GET("/battles", h.List)
	r.POST("/battles", h.Create)
	r.DELETE("/battles/:id", h.Delete)
	return r
}

func doReq(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
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

// ── List ─────────────────────────────────────────────────────────────────────

func TestBattleHandlerList_OK(t *testing.T) {
	repo := &mockBattleRepo{
		listFn: func() ([]Battle, error) {
			return []Battle{{ID: 1, Result: "win"}, {ID: 2, Result: "loss"}}, nil
		},
	}
	r := newBattleRouter(&Handler{repo: repo})
	w := doReq(r, http.MethodGet, "/battles", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp []Battle
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 2 {
		t.Errorf("expected 2 battles, got %d", len(resp))
	}
}

func TestBattleHandlerList_Empty(t *testing.T) {
	repo := &mockBattleRepo{
		listFn: func() ([]Battle, error) { return nil, nil },
	}
	r := newBattleRouter(&Handler{repo: repo})
	w := doReq(r, http.MethodGet, "/battles", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	// Should return [] not null
	body := w.Body.String()
	if body == "null\n" || body == "null" {
		t.Error("expected empty array [], got null")
	}
}

func TestBattleHandlerList_Error(t *testing.T) {
	repo := &mockBattleRepo{
		listFn: func() ([]Battle, error) { return nil, errors.New("db error") },
	}
	r := newBattleRouter(&Handler{repo: repo})
	w := doReq(r, http.MethodGet, "/battles", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestBattleHandlerCreate_OK(t *testing.T) {
	repo := &mockBattleRepo{
		createFn: func(b BattleInput) (int64, error) {
			if b.Result != "win" {
				t.Errorf("Result = %q, want win", b.Result)
			}
			return 7, nil
		},
	}
	r := newBattleRouter(&Handler{repo: repo})
	w := doReq(r, http.MethodPost, "/battles", map[string]any{
		"result":     "win",
		"game_style": "Commander",
		"player_count": 4,
	})
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["id"].(float64) != 7 {
		t.Errorf("id = %v, want 7", resp["id"])
	}
}

func TestBattleHandlerCreate_BadJSON(t *testing.T) {
	repo := &mockBattleRepo{}
	r := newBattleRouter(&Handler{repo: repo})
	req := httptest.NewRequest(http.MethodPost, "/battles", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBattleHandlerCreate_Error(t *testing.T) {
	repo := &mockBattleRepo{
		createFn: func(b BattleInput) (int64, error) { return 0, errors.New("db error") },
	}
	r := newBattleRouter(&Handler{repo: repo})
	w := doReq(r, http.MethodPost, "/battles", map[string]any{"result": "win"})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestBattleHandlerDelete_OK(t *testing.T) {
	repo := &mockBattleRepo{
		deleteFn: func(id string) error {
			if id != "3" {
				t.Errorf("id = %q, want 3", id)
			}
			return nil
		},
	}
	r := newBattleRouter(&Handler{repo: repo})
	w := doReq(r, http.MethodDelete, "/battles/3", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestBattleHandlerDelete_Error(t *testing.T) {
	repo := &mockBattleRepo{
		deleteFn: func(id string) error { return errors.New("not found") },
	}
	r := newBattleRouter(&Handler{repo: repo})
	w := doReq(r, http.MethodDelete, "/battles/99", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}
