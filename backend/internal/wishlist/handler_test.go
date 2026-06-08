package wishlist

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

// ── mock service ──────────────────────────────────────────────────────────────

type mockWishlistSvc struct {
	listFn    func() ([]WishlistCard, error)
	getByIDFn func(id string) (*WishlistCard, error)
	createFn  func(input WishlistCardInput) (int64, error)
	updateFn  func(id string, input WishlistUpdateInput) error
	deleteFn  func(id string) error
	acquireFn func(id string, input AcquireInput) (int64, error)
}

func (m *mockWishlistSvc) List() ([]WishlistCard, error) { return m.listFn() }
func (m *mockWishlistSvc) GetByID(id string) (*WishlistCard, error) { return m.getByIDFn(id) }
func (m *mockWishlistSvc) Create(i WishlistCardInput) (int64, error) { return m.createFn(i) }
func (m *mockWishlistSvc) Update(id string, i WishlistUpdateInput) error { return m.updateFn(id, i) }
func (m *mockWishlistSvc) Delete(id string) error                        { return m.deleteFn(id) }
func (m *mockWishlistSvc) Acquire(id string, i AcquireInput) (int64, error) {
	return m.acquireFn(id, i)
}

func newWishlistRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.GET("/wishlist", h.List)
	r.GET("/wishlist/:id", h.GetByID)
	r.POST("/wishlist", h.Create)
	r.PUT("/wishlist/:id", h.Update)
	r.DELETE("/wishlist/:id", h.Delete)
	r.POST("/wishlist/:id/acquire", h.Acquire)
	return r
}

func wlReq(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
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

func TestWishlistHandlerList_OK(t *testing.T) {
	svc := &mockWishlistSvc{listFn: func() ([]WishlistCard, error) { return []WishlistCard{{ID: 1}}, nil }}
	r := newWishlistRouter(&Handler{svc: svc})
	w := wlReq(r, http.MethodGet, "/wishlist", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestWishlistHandlerList_Error(t *testing.T) {
	svc := &mockWishlistSvc{listFn: func() ([]WishlistCard, error) { return nil, errors.New("fail") }}
	r := newWishlistRouter(&Handler{svc: svc})
	w := wlReq(r, http.MethodGet, "/wishlist", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestWishlistHandlerGetByID_OK(t *testing.T) {
	svc := &mockWishlistSvc{getByIDFn: func(id string) (*WishlistCard, error) { return &WishlistCard{ID: 1}, nil }}
	r := newWishlistRouter(&Handler{svc: svc})
	w := wlReq(r, http.MethodGet, "/wishlist/1", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestWishlistHandlerGetByID_NotFound(t *testing.T) {
	svc := &mockWishlistSvc{getByIDFn: func(id string) (*WishlistCard, error) { return nil, errors.New("not found") }}
	r := newWishlistRouter(&Handler{svc: svc})
	w := wlReq(r, http.MethodGet, "/wishlist/99", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestWishlistHandlerCreate_OK(t *testing.T) {
	svc := &mockWishlistSvc{createFn: func(i WishlistCardInput) (int64, error) { return 7, nil }}
	r := newWishlistRouter(&Handler{svc: svc})
	w := wlReq(r, http.MethodPost, "/wishlist", map[string]any{
		"set_code": "M21", "collection_number": "278",
	})
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
}

func TestWishlistHandlerCreate_ServiceError(t *testing.T) {
	svc := &mockWishlistSvc{createFn: func(i WishlistCardInput) (int64, error) { return 0, errors.New("fail") }}
	r := newWishlistRouter(&Handler{svc: svc})
	w := wlReq(r, http.MethodPost, "/wishlist", map[string]any{
		"set_code": "M21", "collection_number": "278",
	})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestWishlistHandlerUpdate_OK(t *testing.T) {
	svc := &mockWishlistSvc{updateFn: func(id string, i WishlistUpdateInput) error { return nil }}
	r := newWishlistRouter(&Handler{svc: svc})
	w := wlReq(r, http.MethodPut, "/wishlist/1", map[string]any{"reason": "want it"})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestWishlistHandlerUpdate_Error(t *testing.T) {
	svc := &mockWishlistSvc{updateFn: func(id string, i WishlistUpdateInput) error { return errors.New("fail") }}
	r := newWishlistRouter(&Handler{svc: svc})
	w := wlReq(r, http.MethodPut, "/wishlist/1", map[string]any{"reason": "x"})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestWishlistHandlerDelete_OK(t *testing.T) {
	svc := &mockWishlistSvc{deleteFn: func(id string) error { return nil }}
	r := newWishlistRouter(&Handler{svc: svc})
	w := wlReq(r, http.MethodDelete, "/wishlist/1", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestWishlistHandlerDelete_Error(t *testing.T) {
	svc := &mockWishlistSvc{deleteFn: func(id string) error { return errors.New("fail") }}
	r := newWishlistRouter(&Handler{svc: svc})
	w := wlReq(r, http.MethodDelete, "/wishlist/1", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── Acquire ───────────────────────────────────────────────────────────────────

func TestWishlistHandlerAcquire_OK(t *testing.T) {
	svc := &mockWishlistSvc{acquireFn: func(id string, i AcquireInput) (int64, error) { return 42, nil }}
	r := newWishlistRouter(&Handler{svc: svc})
	w := wlReq(r, http.MethodPost, "/wishlist/1/acquire", map[string]any{"condition": "near_mint"})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["card_id"].(float64) != 42 {
		t.Errorf("card_id = %v, want 42", resp["card_id"])
	}
}

func TestWishlistHandlerAcquire_Error(t *testing.T) {
	svc := &mockWishlistSvc{acquireFn: func(id string, i AcquireInput) (int64, error) {
		return 0, errors.New("fail")
	}}
	r := newWishlistRouter(&Handler{svc: svc})
	w := wlReq(r, http.MethodPost, "/wishlist/1/acquire", map[string]any{"condition": "near_mint"})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}
