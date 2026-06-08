package cards

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"magic-collection-api/internal/mtgapi"
)

func init() { gin.SetMode(gin.TestMode) }

// ── parseDeckBuilderOutput ───────────────────────────────────────────────────

func TestParseDeckBuilderOutput_ValidDeck(t *testing.T) {
	raw := `<<<DECK>>>{"nome":"Aggro Red","cores":"R","commander":false,"lista":"4 Lightning Bolt","descricao":"fast deck"}<<<FIM_DECK>>>
some analysis text`

	s, analysis, errIA, roles := parseDeckBuilderOutput(raw)
	if s == nil {
		t.Fatal("expected a deck suggestion")
	}
	if s.Nome != "Aggro Red" {
		t.Errorf("Nome = %q, want %q", s.Nome, "Aggro Red")
	}
	if analysis == "" {
		t.Error("expected analysis text")
	}
	if errIA != "" {
		t.Errorf("unexpected error: %q", errIA)
	}
	if roles != nil {
		t.Error("expected no card roles without CARTAS block")
	}
}

func TestParseDeckBuilderOutput_WithCardRoles(t *testing.T) {
	raw := `<<<DECK>>>{"nome":"Control","cores":"U","commander":false,"lista":"4 Counterspell","descricao":"control"}<<<FIM_DECK>>>
analysis
<<<CARTAS>>>{"nao_terrenos":[{"nome":"Counterspell","papel":"Counter spell"}],"terrenos":{"total":24,"motivo":"Standard mana base"}}<<<FIM_CARTAS>>>
more text`

	s, _, _, roles := parseDeckBuilderOutput(raw)
	if s == nil {
		t.Fatal("expected deck suggestion")
	}
	if roles == nil {
		t.Fatal("expected card roles")
	}
	if len(roles.NaoTerrenos) != 1 {
		t.Errorf("expected 1 non-land role, got %d", len(roles.NaoTerrenos))
	}
	if roles.Terrenos.Total != 24 {
		t.Errorf("Terrenos.Total = %d, want 24", roles.Terrenos.Total)
	}
}

func TestParseDeckBuilderOutput_ErrorBlock(t *testing.T) {
	raw := `<<<ERRO>>>{"motivo":"Cartas insuficientes para montar um deck"}<<<FIM_ERRO>>>`
	s, _, errIA, _ := parseDeckBuilderOutput(raw)
	if s != nil {
		t.Error("expected nil suggestion on error")
	}
	if errIA == "" {
		t.Error("expected error message")
	}
}

func TestParseDeckBuilderOutput_NoDeckBlock(t *testing.T) {
	raw := `just some random text without deck tags`
	s, analysis, errIA, _ := parseDeckBuilderOutput(raw)
	if s != nil {
		t.Error("expected nil suggestion")
	}
	if analysis != raw {
		t.Errorf("expected raw text as analysis, got %q", analysis)
	}
	if errIA != "" {
		t.Errorf("unexpected error: %q", errIA)
	}
}

func TestParseDeckBuilderOutput_InvalidJSON(t *testing.T) {
	raw := `<<<DECK>>>not valid json<<<FIM_DECK>>>analysis`
	s, _, _, _ := parseDeckBuilderOutput(raw)
	if s != nil {
		t.Error("expected nil suggestion with invalid JSON")
	}
}

func TestParseDeckBuilderOutput_MissingEndTag(t *testing.T) {
	raw := `<<<DECK>>>{"nome":"test"}` // no end tag
	s, _, _, _ := parseDeckBuilderOutput(raw)
	if s != nil {
		t.Error("expected nil suggestion when end tag missing")
	}
}

// ── mock cardService ─────────────────────────────────────────────────────────

type mockCardService struct {
	createFn              func(input CreateCardInput) (int64, error)
	listFn                func(params ListParams) (ListResult, error)
	getByIDFn             func(id string) (map[string]any, error)
	updateFn              func(id string, input UpdateCardInput) error
	deleteFn              func(id string) error
	setDeckFn             func(id string, deckID int) error
	normalizeRaritiesFn   func() (NormalizeRarityResult, error)
	refreshColorsFn       func() (RefreshColorsResult, error)
	listColorCombosFn     func() ([]ColorCombo, error)
	setQuantityFn         func(id string, quantity int) error
	previewFn             func(input PreviewCardInput) (*mtgapi.ExternalCard, error)
	getStatsFn            func() (CollectionStats, error)
	refreshImagesFn       func() (ImageRefreshResult, error)
	refreshPricesFn       func(emptyOnly bool) (PriceRefreshResult, error)
	getCardsForDeckFn     func() ([]DeckBuilderCard, error)
	exportAllFn           func() ([]Card, error)
}

func (m *mockCardService) Create(i CreateCardInput) (int64, error) { return m.createFn(i) }
func (m *mockCardService) List(p ListParams) (ListResult, error)   { return m.listFn(p) }
func (m *mockCardService) GetByID(id string) (map[string]any, error) { return m.getByIDFn(id) }
func (m *mockCardService) Update(id string, i UpdateCardInput) error  { return m.updateFn(id, i) }
func (m *mockCardService) Delete(id string) error                     { return m.deleteFn(id) }
func (m *mockCardService) SetDeck(id string, deckID int) error        { return m.setDeckFn(id, deckID) }
func (m *mockCardService) NormalizeRarities() (NormalizeRarityResult, error) {
	return m.normalizeRaritiesFn()
}
func (m *mockCardService) RefreshColors() (RefreshColorsResult, error) { return m.refreshColorsFn() }
func (m *mockCardService) ListColorCombos() ([]ColorCombo, error)      { return m.listColorCombosFn() }
func (m *mockCardService) SetQuantity(id string, q int) error          { return m.setQuantityFn(id, q) }
func (m *mockCardService) Preview(i PreviewCardInput) (*mtgapi.ExternalCard, error) {
	return m.previewFn(i)
}
func (m *mockCardService) GetStats() (CollectionStats, error) { return m.getStatsFn() }
func (m *mockCardService) RefreshImages() (ImageRefreshResult, error) { return m.refreshImagesFn() }
func (m *mockCardService) RefreshPrices(emptyOnly bool) (PriceRefreshResult, error) {
	return m.refreshPricesFn(emptyOnly)
}
func (m *mockCardService) GetCardsForDeckBuilder() ([]DeckBuilderCard, error) {
	return m.getCardsForDeckFn()
}
func (m *mockCardService) ExportAll() ([]Card, error) { return m.exportAllFn() }

type mockAI struct {
	completeFn func(prompt string) (string, error)
}

func (m *mockAI) Complete(prompt string) (string, error) { return m.completeFn(prompt) }

func newRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.GET("/cards", h.List)
	r.POST("/cards", h.Create)
	r.GET("/cards/export", h.Export)
	r.GET("/cards/colors", h.ListColors)
	r.GET("/cards/stats", h.Stats)
	r.POST("/cards/preview", h.Preview)
	r.POST("/cards/refresh-images", h.RefreshImages)
	r.POST("/cards/refresh-prices", h.RefreshPrices)
	r.POST("/cards/suggest-decks", h.SuggestDecks)
	r.GET("/cards/:id", h.GetByID)
	r.PUT("/cards/:id", h.Update)
	r.DELETE("/cards/:id", h.Delete)
	r.PATCH("/cards/:id/deck", h.SetDeck)
	r.PATCH("/cards/:id/quantity", h.UpdateQuantity)
	return r
}

func doRequest(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
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

// ── Handler.List ─────────────────────────────────────────────────────────────

func TestHandlerList_OK(t *testing.T) {
	svc := &mockCardService{
		listFn: func(p ListParams) (ListResult, error) {
			return ListResult{Cards: []Card{{ID: 1}}, Total: 1, Page: 1, PageSize: 20, TotalPages: 1}, nil
		},
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodGet, "/cards", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestHandlerList_Error(t *testing.T) {
	svc := &mockCardService{
		listFn: func(p ListParams) (ListResult, error) { return ListResult{}, errors.New("db down") },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodGet, "/cards", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestHandlerList_WithFilters(t *testing.T) {
	var capturedParams ListParams
	svc := &mockCardService{
		listFn: func(p ListParams) (ListResult, error) {
			capturedParams = p
			return ListResult{}, nil
		},
	}
	r := newRouter(&Handler{service: svc})
	doRequest(r, http.MethodGet, "/cards?foil=1&full_art=1&rarity=R&q=bolt", nil)
	if !capturedParams.FoilOnly {
		t.Error("FoilOnly should be true")
	}
	if !capturedParams.FullArtOnly {
		t.Error("FullArtOnly should be true")
	}
	if capturedParams.RarityFilter != "R" {
		t.Errorf("RarityFilter = %q, want R", capturedParams.RarityFilter)
	}
}

// ── Handler.Create ───────────────────────────────────────────────────────────

func TestHandlerCreate_OK(t *testing.T) {
	svc := &mockCardService{
		createFn: func(i CreateCardInput) (int64, error) { return 99, nil },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodPost, "/cards", map[string]any{"name": "Island"})
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
}

func TestHandlerCreate_BadJSON(t *testing.T) {
	svc := &mockCardService{}
	r := newRouter(&Handler{service: svc})
	req := httptest.NewRequest(http.MethodPost, "/cards", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandlerCreate_ServiceError(t *testing.T) {
	svc := &mockCardService{
		createFn: func(i CreateCardInput) (int64, error) { return 0, errors.New("db error") },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodPost, "/cards", map[string]any{"name": "Island"})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── Handler.GetByID ──────────────────────────────────────────────────────────

func TestHandlerGetByID_OK(t *testing.T) {
	svc := &mockCardService{
		getByIDFn: func(id string) (map[string]any, error) {
			return map[string]any{"local": Card{ID: 1}}, nil
		},
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodGet, "/cards/1", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestHandlerGetByID_NotFound(t *testing.T) {
	svc := &mockCardService{
		getByIDFn: func(id string) (map[string]any, error) { return nil, errors.New("not found") },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodGet, "/cards/999", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// ── Handler.Update ───────────────────────────────────────────────────────────

func TestHandlerUpdate_OK(t *testing.T) {
	svc := &mockCardService{
		updateFn: func(id string, i UpdateCardInput) error { return nil },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodPut, "/cards/1", map[string]any{"name": "Island"})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestHandlerUpdate_Error(t *testing.T) {
	svc := &mockCardService{
		updateFn: func(id string, i UpdateCardInput) error { return errors.New("fail") },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodPut, "/cards/1", map[string]any{"name": "Island"})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── Handler.Delete ───────────────────────────────────────────────────────────

func TestHandlerDelete_OK(t *testing.T) {
	svc := &mockCardService{
		deleteFn: func(id string) error { return nil },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodDelete, "/cards/1", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestHandlerDelete_Error(t *testing.T) {
	svc := &mockCardService{
		deleteFn: func(id string) error { return errors.New("fail") },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodDelete, "/cards/1", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── Handler.SetDeck ──────────────────────────────────────────────────────────

func TestHandlerSetDeck_OK(t *testing.T) {
	svc := &mockCardService{
		setDeckFn: func(id string, deckID int) error { return nil },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodPatch, "/cards/1/deck", map[string]any{"deck_id": 5})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// ── Handler.UpdateQuantity ───────────────────────────────────────────────────

func TestHandlerUpdateQuantity_OK(t *testing.T) {
	svc := &mockCardService{
		setQuantityFn: func(id string, q int) error { return nil },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodPatch, "/cards/1/quantity", map[string]any{"quantity": 3})
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

func TestHandlerUpdateQuantity_InvalidQuantity(t *testing.T) {
	svc := &mockCardService{}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodPatch, "/cards/1/quantity", map[string]any{"quantity": 0})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ── Handler.Export ───────────────────────────────────────────────────────────

func TestHandlerExport_OK(t *testing.T) {
	svc := &mockCardService{
		exportAllFn: func() ([]Card, error) { return []Card{{ID: 1}}, nil },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodGet, "/cards/export", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// ── Handler.ListColors / Stats ───────────────────────────────────────────────

func TestHandlerListColors_OK(t *testing.T) {
	svc := &mockCardService{
		listColorCombosFn: func() ([]ColorCombo, error) { return []ColorCombo{}, nil },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodGet, "/cards/colors", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestHandlerStats_OK(t *testing.T) {
	svc := &mockCardService{
		getStatsFn: func() (CollectionStats, error) { return CollectionStats{}, nil },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodGet, "/cards/stats", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// ── Handler.Preview ──────────────────────────────────────────────────────────

func TestHandlerPreview_Found(t *testing.T) {
	svc := &mockCardService{
		previewFn: func(i PreviewCardInput) (*mtgapi.ExternalCard, error) {
			return &mtgapi.ExternalCard{Name: "Island"}, nil
		},
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodPost, "/cards/preview", map[string]any{"set_code": "M21", "collection_number": "278"})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if found, _ := resp["found"].(bool); !found {
		t.Error("expected found=true")
	}
}

func TestHandlerPreview_NotFound(t *testing.T) {
	svc := &mockCardService{
		previewFn: func(i PreviewCardInput) (*mtgapi.ExternalCard, error) { return nil, nil },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodPost, "/cards/preview", map[string]any{"set_code": "XXX", "collection_number": "1"})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if found, _ := resp["found"].(bool); found {
		t.Error("expected found=false")
	}
}

// ── Handler.RefreshImages ────────────────────────────────────────────────────

func TestHandlerRefreshImages_OK(t *testing.T) {
	svc := &mockCardService{
		refreshImagesFn: func() (ImageRefreshResult, error) { return ImageRefreshResult{Updated: 5}, nil },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodPost, "/cards/refresh-images", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestHandlerRefreshImages_Error(t *testing.T) {
	svc := &mockCardService{
		refreshImagesFn: func() (ImageRefreshResult, error) { return ImageRefreshResult{}, errors.New("fail") },
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodPost, "/cards/refresh-images", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── Handler.RefreshPrices ─────────────────────────────────────────────────────

func TestHandlerRefreshPrices_OK(t *testing.T) {
	var capturedEmpty bool
	svc := &mockCardService{
		refreshPricesFn: func(emptyOnly bool) (PriceRefreshResult, error) {
			capturedEmpty = emptyOnly
			return PriceRefreshResult{Updated: 10}, nil
		},
	}
	r := newRouter(&Handler{service: svc})
	w := doRequest(r, http.MethodPost, "/cards/refresh-prices?empty_only=1", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !capturedEmpty {
		t.Error("expected emptyOnly=true")
	}
}

// ── Handler.SuggestDecks ─────────────────────────────────────────────────────

func TestHandlerSuggestDecks_NoCards(t *testing.T) {
	svc := &mockCardService{
		getCardsForDeckFn: func() ([]DeckBuilderCard, error) { return []DeckBuilderCard{}, nil },
	}
	r := newRouter(&Handler{service: svc, aiClient: &mockAI{}})
	w := doRequest(r, http.MethodPost, "/cards/suggest-decks", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandlerSuggestDecks_AIError(t *testing.T) {
	svc := &mockCardService{
		getCardsForDeckFn: func() ([]DeckBuilderCard, error) {
			return []DeckBuilderCard{{Name: "Island", Type: "Land"}}, nil
		},
	}
	ai := &mockAI{completeFn: func(prompt string) (string, error) { return "", errors.New("api down") }}
	r := newRouter(&Handler{service: svc, aiClient: ai})
	w := doRequest(r, http.MethodPost, "/cards/suggest-decks", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestHandlerSuggestDecks_WithDeckOutput(t *testing.T) {
	svc := &mockCardService{
		getCardsForDeckFn: func() ([]DeckBuilderCard, error) {
			return []DeckBuilderCard{{Name: "Lightning Bolt", Type: "Instant", Quantity: 4}}, nil
		},
	}
	deckJSON := `{"nome":"Aggro","cores":"R","commander":false,"lista":"4 Lightning Bolt","descricao":"fast"}`
	aiResp := `<<<DECK>>>` + deckJSON + `<<<FIM_DECK>>>
Analysis text here`
	ai := &mockAI{completeFn: func(prompt string) (string, error) { return aiResp, nil }}
	r := newRouter(&Handler{service: svc, aiClient: ai})
	w := doRequest(r, http.MethodPost, "/cards/suggest-decks", map[string]any{"format": "Standard", "goal": "aggro"})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["deck_name"] != "Aggro" {
		t.Errorf("deck_name = %v, want Aggro", resp["deck_name"])
	}
}
