package importer

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

// ── parsePriceUSD (pure function in package) ──────────────────────────────────

func TestImporterParsePriceUSD(t *testing.T) {
	tests := []struct {
		name   string
		prices map[string]string
		foil   bool
		want   float64
	}{
		{"nil", nil, false, 0},
		{"normal", map[string]string{"usd": "2.50"}, false, 2.50},
		{"foil prefers usd_foil", map[string]string{"usd": "1.00", "usd_foil": "4.00"}, true, 4.00},
		{"foil fallback", map[string]string{"usd": "1.50"}, true, 1.50},
		{"empty value", map[string]string{"usd": ""}, false, 0},
		{"invalid", map[string]string{"usd": "abc"}, false, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parsePriceUSD(tc.prices, tc.foil)
			if got != tc.want {
				t.Errorf("parsePriceUSD = %v, want %v", got, tc.want)
			}
		})
	}
}

// ── mock service ──────────────────────────────────────────────────────────────

type mockImporterSvc struct {
	importPreconFn      func(input ImportPreconInput) (ImportResult, error)
	importDeckListFn    func(input ImportDeckListInput) (ImportResult, error)
	importIntoDeckFn    func(deckID int64, input ImportCardsToDeckInput) (ImportResult, error)
}

func (m *mockImporterSvc) ImportPrecon(i ImportPreconInput) (ImportResult, error) {
	return m.importPreconFn(i)
}
func (m *mockImporterSvc) ImportDeckList(i ImportDeckListInput) (ImportResult, error) {
	return m.importDeckListFn(i)
}
func (m *mockImporterSvc) ImportCardsIntoDeck(deckID int64, i ImportCardsToDeckInput) (ImportResult, error) {
	return m.importIntoDeckFn(deckID, i)
}

func newImporterRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.POST("/decks/import-precon", h.ImportPrecon)
	r.POST("/decks/import-list", h.ImportDeckList)
	r.POST("/decks/:id/import-cards", h.ImportCardsIntoDeck)
	return r
}

func impReq(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
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

// ── ImportPrecon ──────────────────────────────────────────────────────────────

func TestImporterHandlerPrecon_OK(t *testing.T) {
	svc := &mockImporterSvc{importPreconFn: func(i ImportPreconInput) (ImportResult, error) {
		return ImportResult{DeckID: 1, Imported: 100}, nil
	}}
	r := newImporterRouter(&Handler{svc: svc})
	w := impReq(r, http.MethodPost, "/decks/import-precon", map[string]any{
		"set_code": "NCC", "deck_name": "New Capenna Commander",
	})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp ImportResult
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Imported != 100 {
		t.Errorf("Imported = %d, want 100", resp.Imported)
	}
}

func TestImporterHandlerPrecon_BadJSON(t *testing.T) {
	svc := &mockImporterSvc{}
	r := newImporterRouter(&Handler{svc: svc})
	req := httptest.NewRequest(http.MethodPost, "/decks/import-precon", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestImporterHandlerPrecon_ServiceError(t *testing.T) {
	svc := &mockImporterSvc{importPreconFn: func(i ImportPreconInput) (ImportResult, error) {
		return ImportResult{}, errors.New("scryfall error")
	}}
	r := newImporterRouter(&Handler{svc: svc})
	w := impReq(r, http.MethodPost, "/decks/import-precon", map[string]any{
		"set_code": "BAD", "deck_name": "Bad Deck",
	})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── ImportDeckList ────────────────────────────────────────────────────────────

func TestImporterHandlerDeckList_OK(t *testing.T) {
	svc := &mockImporterSvc{importDeckListFn: func(i ImportDeckListInput) (ImportResult, error) {
		return ImportResult{DeckID: 5, Imported: 60}, nil
	}}
	r := newImporterRouter(&Handler{svc: svc})
	w := impReq(r, http.MethodPost, "/decks/import-list", map[string]any{
		"deck_name": "My Aggro", "deck_list": "4 Lightning Bolt\n20 Mountain",
	})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestImporterHandlerDeckList_BadJSON(t *testing.T) {
	svc := &mockImporterSvc{}
	r := newImporterRouter(&Handler{svc: svc})
	req := httptest.NewRequest(http.MethodPost, "/decks/import-list", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestImporterHandlerDeckList_ServiceError(t *testing.T) {
	svc := &mockImporterSvc{importDeckListFn: func(i ImportDeckListInput) (ImportResult, error) {
		return ImportResult{}, errors.New("parse error")
	}}
	r := newImporterRouter(&Handler{svc: svc})
	w := impReq(r, http.MethodPost, "/decks/import-list", map[string]any{
		"deck_name": "Test", "deck_list": "invalid",
	})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── ImportCardsIntoDeck ───────────────────────────────────────────────────────

func TestImporterHandlerCardsIntoDeck_OK(t *testing.T) {
	svc := &mockImporterSvc{importIntoDeckFn: func(deckID int64, i ImportCardsToDeckInput) (ImportResult, error) {
		return ImportResult{DeckID: deckID, Imported: 30}, nil
	}}
	r := newImporterRouter(&Handler{svc: svc})
	w := impReq(r, http.MethodPost, "/decks/3/import-cards", map[string]any{
		"deck_list": "4 Counterspell",
	})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestImporterHandlerCardsIntoDeck_InvalidDeckID(t *testing.T) {
	svc := &mockImporterSvc{}
	r := newImporterRouter(&Handler{svc: svc})
	w := impReq(r, http.MethodPost, "/decks/abc/import-cards", map[string]any{
		"deck_list": "4 Counterspell",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestImporterHandlerCardsIntoDeck_ServiceError(t *testing.T) {
	svc := &mockImporterSvc{importIntoDeckFn: func(deckID int64, i ImportCardsToDeckInput) (ImportResult, error) {
		return ImportResult{}, errors.New("deck not found")
	}}
	r := newImporterRouter(&Handler{svc: svc})
	w := impReq(r, http.MethodPost, "/decks/99/import-cards", map[string]any{
		"deck_list": "4 Lightning Bolt",
	})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}
