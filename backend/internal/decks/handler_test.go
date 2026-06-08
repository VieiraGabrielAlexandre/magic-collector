package decks

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

// ── mock deck service ──────────────────────────────────────────────────────────

type mockDeckSvc struct {
	listFn        func() ([]Deck, error)
	createFn      func(input DeckInput) (int64, error)
	updateFn      func(id string, input DeckInput) error
	deleteFn      func(id string) error
	fetchIconFn   func(id string) (string, error)
	evaluateFn    func(id string) (*Deck, error)
}

func (m *mockDeckSvc) List() ([]Deck, error)                  { return m.listFn() }
func (m *mockDeckSvc) Create(i DeckInput) (int64, error)      { return m.createFn(i) }
func (m *mockDeckSvc) Update(id string, i DeckInput) error    { return m.updateFn(id, i) }
func (m *mockDeckSvc) Delete(id string) error                 { return m.deleteFn(id) }
func (m *mockDeckSvc) FetchIcon(id string) (string, error)    { return m.fetchIconFn(id) }
func (m *mockDeckSvc) EvaluateDeck(id string) (*Deck, error)  { return m.evaluateFn(id) }

func newDeckRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.GET("/decks", h.List)
	r.POST("/decks", h.Create)
	r.PUT("/decks/:id", h.Update)
	r.DELETE("/decks/:id", h.Delete)
	r.PATCH("/decks/:id/icon", h.FetchIcon)
	r.POST("/decks/:id/evaluate", h.Evaluate)
	return r
}

func dkReq(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
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

// ── List ──────────────────────────────────────────────────────────────────────

func TestDeckHandlerList_OK(t *testing.T) {
	svc := &mockDeckSvc{listFn: func() ([]Deck, error) { return []Deck{{ID: 1, Name: "Aggro"}}, nil }}
	r := newDeckRouter(&Handler{svc: svc})
	w := dkReq(r, http.MethodGet, "/decks", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestDeckHandlerList_Empty(t *testing.T) {
	svc := &mockDeckSvc{listFn: func() ([]Deck, error) { return nil, nil }}
	r := newDeckRouter(&Handler{svc: svc})
	w := dkReq(r, http.MethodGet, "/decks", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestDeckHandlerList_Error(t *testing.T) {
	svc := &mockDeckSvc{listFn: func() ([]Deck, error) { return nil, errors.New("db error") }}
	r := newDeckRouter(&Handler{svc: svc})
	w := dkReq(r, http.MethodGet, "/decks", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestDeckHandlerCreate_OK(t *testing.T) {
	svc := &mockDeckSvc{createFn: func(i DeckInput) (int64, error) { return 3, nil }}
	r := newDeckRouter(&Handler{svc: svc})
	w := dkReq(r, http.MethodPost, "/decks", map[string]any{"name": "Grixis"})
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
}

func TestDeckHandlerCreate_BadJSON(t *testing.T) {
	svc := &mockDeckSvc{}
	r := newDeckRouter(&Handler{svc: svc})
	req := httptest.NewRequest(http.MethodPost, "/decks", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestDeckHandlerCreate_ServiceError(t *testing.T) {
	svc := &mockDeckSvc{createFn: func(i DeckInput) (int64, error) { return 0, errors.New("fail") }}
	r := newDeckRouter(&Handler{svc: svc})
	w := dkReq(r, http.MethodPost, "/decks", map[string]any{"name": "Test"})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestDeckHandlerUpdate_OK(t *testing.T) {
	svc := &mockDeckSvc{updateFn: func(id string, i DeckInput) error { return nil }}
	r := newDeckRouter(&Handler{svc: svc})
	w := dkReq(r, http.MethodPut, "/decks/1", map[string]any{"name": "Updated"})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestDeckHandlerUpdate_Error(t *testing.T) {
	svc := &mockDeckSvc{updateFn: func(id string, i DeckInput) error { return errors.New("fail") }}
	r := newDeckRouter(&Handler{svc: svc})
	w := dkReq(r, http.MethodPut, "/decks/1", map[string]any{"name": "Test"})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDeckHandlerDelete_OK(t *testing.T) {
	svc := &mockDeckSvc{deleteFn: func(id string) error { return nil }}
	r := newDeckRouter(&Handler{svc: svc})
	w := dkReq(r, http.MethodDelete, "/decks/1", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestDeckHandlerDelete_Error(t *testing.T) {
	svc := &mockDeckSvc{deleteFn: func(id string) error { return errors.New("fail") }}
	r := newDeckRouter(&Handler{svc: svc})
	w := dkReq(r, http.MethodDelete, "/decks/1", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── FetchIcon ─────────────────────────────────────────────────────────────────

func TestDeckHandlerFetchIcon_OK(t *testing.T) {
	svc := &mockDeckSvc{fetchIconFn: func(id string) (string, error) { return "https://icon.svg", nil }}
	r := newDeckRouter(&Handler{svc: svc})
	w := dkReq(r, http.MethodPatch, "/decks/1/icon", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestDeckHandlerFetchIcon_Error(t *testing.T) {
	svc := &mockDeckSvc{fetchIconFn: func(id string) (string, error) { return "", errors.New("fail") }}
	r := newDeckRouter(&Handler{svc: svc})
	w := dkReq(r, http.MethodPatch, "/decks/1/icon", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── Evaluate ─────────────────────────────────────────────────────────────────

func TestDeckHandlerEvaluate_OK(t *testing.T) {
	svc := &mockDeckSvc{evaluateFn: func(id string) (*Deck, error) {
		return &Deck{ID: 1, Evaluation: "Great deck!"}, nil
	}}
	r := newDeckRouter(&Handler{svc: svc})
	w := dkReq(r, http.MethodPost, "/decks/1/evaluate", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestDeckHandlerEvaluate_Error(t *testing.T) {
	svc := &mockDeckSvc{evaluateFn: func(id string) (*Deck, error) { return nil, errors.New("ai down") }}
	r := newDeckRouter(&Handler{svc: svc})
	w := dkReq(r, http.MethodPost, "/decks/1/evaluate", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}
