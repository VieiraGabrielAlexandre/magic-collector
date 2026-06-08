package cards

import (
	"errors"
	"testing"

	"magic-collection-api/internal/mtgapi"
)

// ── parsePriceUSD ────────────────────────────────────────────────────────────

func TestParsePriceUSD(t *testing.T) {
	tests := []struct {
		name   string
		prices map[string]string
		foil   bool
		want   float64
	}{
		{"nil prices", nil, false, 0},
		{"nil prices foil", nil, true, 0},
		{"normal price", map[string]string{"usd": "1.50"}, false, 1.50},
		{"foil prefers usd_foil", map[string]string{"usd": "1.50", "usd_foil": "3.00"}, true, 3.00},
		{"foil fallback to usd when no usd_foil", map[string]string{"usd": "2.00"}, true, 2.00},
		{"foil with empty usd_foil falls back to usd", map[string]string{"usd": "2.00", "usd_foil": ""}, true, 2.00},
		{"missing usd key", map[string]string{"eur": "1.00"}, false, 0},
		{"empty usd value", map[string]string{"usd": ""}, false, 0},
		{"invalid usd value", map[string]string{"usd": "abc"}, false, 0},
		{"zero price", map[string]string{"usd": "0.00"}, false, 0},
		{"large price", map[string]string{"usd": "9999.99"}, false, 9999.99},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parsePriceUSD(tc.prices, tc.foil)
			if got != tc.want {
				t.Errorf("parsePriceUSD(%v, %v) = %v, want %v", tc.prices, tc.foil, got, tc.want)
			}
		})
	}
}

// ── mock implementations ─────────────────────────────────────────────────────

type mockCardRepository struct {
	createFn              func(card Card) (int64, error)
	listFn                func(params ListParams) (ListResult, error)
	getByIDFn             func(id string) (*Card, error)
	updateFn              func(id string, card Card) error
	updateSharedFn        func(oldName, oldSetCode, oldCollNum, oldLang string, oldFoil bool, card Card) error
	updateMTGIDFn         func(id, mtgID string) error
	deleteFn              func(id string) error
	setQuantityFn         func(id string, quantity int) error
	setDeckFn             func(id string, deckID int) error
	listAllFn             func() ([]Card, error)
	listAllPriceFn        func() ([]CardForPriceRefresh, error)
	listEmptyPriceFn      func() ([]CardForPriceRefresh, error)
	updatePriceFn         func(id int, price float64) error
	updatePriceMTGFn      func(id int, mtgID string, price float64) error
	updateImageFn         func(id int, imageURL string) error
	updateImageMTGFn      func(id int, mtgID, imageURL string) error
	listNoColorsFn        func() ([]CardForPriceRefresh, error)
	normalizeRaritiesFn   func() (NormalizeRarityResult, error)
	updateColorsFn        func(id int, colors, color string) error
	updateColorsMTGFn     func(id int, mtgID, colors, color string) error
	listColorCombosFn     func() ([]ColorCombo, error)
	getStatsFn            func() (CollectionStats, error)
	listForDeckBuilderFn  func() ([]DeckBuilderCard, error)
}

func (m *mockCardRepository) Create(card Card) (int64, error)  { return m.createFn(card) }
func (m *mockCardRepository) List(p ListParams) (ListResult, error) { return m.listFn(p) }
func (m *mockCardRepository) GetByID(id string) (*Card, error) { return m.getByIDFn(id) }
func (m *mockCardRepository) Update(id string, card Card) error { return m.updateFn(id, card) }
func (m *mockCardRepository) UpdateSharedByIdentity(n, sc, cn, l string, f bool, c Card) error {
	return m.updateSharedFn(n, sc, cn, l, f, c)
}
func (m *mockCardRepository) UpdateMTGID(id, mtgID string) error { return m.updateMTGIDFn(id, mtgID) }
func (m *mockCardRepository) Delete(id string) error             { return m.deleteFn(id) }
func (m *mockCardRepository) SetQuantity(id string, q int) error { return m.setQuantityFn(id, q) }
func (m *mockCardRepository) SetDeck(id string, deckID int) error { return m.setDeckFn(id, deckID) }
func (m *mockCardRepository) ListAll() ([]Card, error)           { return m.listAllFn() }
func (m *mockCardRepository) ListAllForPriceRefresh() ([]CardForPriceRefresh, error) {
	return m.listAllPriceFn()
}
func (m *mockCardRepository) ListEmptyPricesForRefresh() ([]CardForPriceRefresh, error) {
	return m.listEmptyPriceFn()
}
func (m *mockCardRepository) UpdatePrice(id int, price float64) error {
	return m.updatePriceFn(id, price)
}
func (m *mockCardRepository) UpdatePriceAndMTGID(id int, mtgID string, price float64) error {
	return m.updatePriceMTGFn(id, mtgID, price)
}
func (m *mockCardRepository) UpdateImageURL(id int, imageURL string) error {
	return m.updateImageFn(id, imageURL)
}
func (m *mockCardRepository) UpdateImageURLAndMTGID(id int, mtgID, imageURL string) error {
	return m.updateImageMTGFn(id, mtgID, imageURL)
}
func (m *mockCardRepository) ListCardsWithoutColors() ([]CardForPriceRefresh, error) {
	return m.listNoColorsFn()
}
func (m *mockCardRepository) NormalizeRarities() (NormalizeRarityResult, error) {
	return m.normalizeRaritiesFn()
}
func (m *mockCardRepository) UpdateColors(id int, colors, color string) error {
	return m.updateColorsFn(id, colors, color)
}
func (m *mockCardRepository) UpdateColorsAndMTGID(id int, mtgID, colors, color string) error {
	return m.updateColorsMTGFn(id, mtgID, colors, color)
}
func (m *mockCardRepository) ListColorCombos() ([]ColorCombo, error) { return m.listColorCombosFn() }
func (m *mockCardRepository) GetStats() (CollectionStats, error)      { return m.getStatsFn() }
func (m *mockCardRepository) ListForDeckBuilder() ([]DeckBuilderCard, error) {
	return m.listForDeckBuilderFn()
}

type mockMTGClient struct {
	searchFn       func(setCode, number, lang, artist string) (*mtgapi.ExternalCard, error)
	searchPreFn    func(name, lang, artist string) (*mtgapi.ExternalCard, error)
	getByMTGIDFn   func(id string) (*mtgapi.ExternalCard, error)
}

func (m *mockMTGClient) Search(setCode, number, lang, artist string) (*mtgapi.ExternalCard, error) {
	return m.searchFn(setCode, number, lang, artist)
}
func (m *mockMTGClient) SearchPreRelease(name, lang, artist string) (*mtgapi.ExternalCard, error) {
	return m.searchPreFn(name, lang, artist)
}
func (m *mockMTGClient) GetByMTGID(id string) (*mtgapi.ExternalCard, error) {
	return m.getByMTGIDFn(id)
}

func newNilMTG() *mockMTGClient {
	return &mockMTGClient{
		searchFn:     func(_, _, _, _ string) (*mtgapi.ExternalCard, error) { return nil, nil },
		searchPreFn:  func(_, _, _ string) (*mtgapi.ExternalCard, error) { return nil, nil },
		getByMTGIDFn: func(_ string) (*mtgapi.ExternalCard, error) { return nil, nil },
	}
}

// ── Service.Create ───────────────────────────────────────────────────────────

func TestServiceCreate_NoScryfall(t *testing.T) {
	repo := &mockCardRepository{
		createFn: func(card Card) (int64, error) {
			if card.Quantity != 1 {
				t.Errorf("expected quantity=1, got %d", card.Quantity)
			}
			return 42, nil
		},
	}
	svc := &Service{repository: repo, mtgClient: newNilMTG()}
	id, err := svc.Create(CreateCardInput{Name: "Lightning Bolt", SetCode: "M11", CollectionNumber: "149"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("expected id=42, got %d", id)
	}
}

func TestServiceCreate_DefaultsQuantityToOne(t *testing.T) {
	var savedCard Card
	repo := &mockCardRepository{
		createFn: func(card Card) (int64, error) { savedCard = card; return 1, nil },
	}
	svc := &Service{repository: repo, mtgClient: newNilMTG()}
	_, _ = svc.Create(CreateCardInput{Name: "Island", Quantity: 0})
	if savedCard.Quantity != 1 {
		t.Errorf("expected quantity=1 when input is 0, got %d", savedCard.Quantity)
	}
}

func TestServiceCreate_ScryfallEnrichesCard(t *testing.T) {
	extCard := &mtgapi.ExternalCard{
		ID:       "uuid-123",
		Name:     "Lightning Bolt",
		Set:      "m11",
		Rarity:   "C",
		Type:     "Instant",
		ManaCost: "{R}",
		Colors:   []string{"R"},
		ImageURL: "https://img.example.com/bolt.jpg",
		FullArt:  false,
		Prices:   map[string]string{"usd": "0.50"},
	}
	mtg := &mockMTGClient{
		searchFn:     func(_, _, _, _ string) (*mtgapi.ExternalCard, error) { return extCard, nil },
		searchPreFn:  func(_, _, _ string) (*mtgapi.ExternalCard, error) { return nil, nil },
		getByMTGIDFn: func(_ string) (*mtgapi.ExternalCard, error) { return nil, nil },
	}

	var savedCard Card
	repo := &mockCardRepository{
		createFn: func(card Card) (int64, error) { savedCard = card; return 1, nil },
	}
	svc := &Service{repository: repo, mtgClient: mtg}
	_, err := svc.Create(CreateCardInput{Name: "Lightning Bolt", SetCode: "M11", CollectionNumber: "149"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if savedCard.MTGID != "uuid-123" {
		t.Errorf("MTGID not set, got %q", savedCard.MTGID)
	}
	if savedCard.PriceUSD != 0.50 {
		t.Errorf("PriceUSD not set, got %v", savedCard.PriceUSD)
	}
	if savedCard.ImageURL != extCard.ImageURL {
		t.Errorf("ImageURL not set")
	}
}

func TestServiceCreate_PreRelease(t *testing.T) {
	mtg := &mockMTGClient{
		searchFn:    func(_, _, _, _ string) (*mtgapi.ExternalCard, error) { return nil, errors.New("should not call Search") },
		searchPreFn: func(name, lang, artist string) (*mtgapi.ExternalCard, error) {
			if name != "Sol Ring" {
				t.Errorf("expected name=Sol Ring, got %q", name)
			}
			return nil, nil
		},
		getByMTGIDFn: func(_ string) (*mtgapi.ExternalCard, error) { return nil, nil },
	}
	repo := &mockCardRepository{
		createFn: func(card Card) (int64, error) { return 1, nil },
	}
	svc := &Service{repository: repo, mtgClient: mtg}
	_, err := svc.Create(CreateCardInput{Name: "Sol Ring", PreRelease: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceCreate_RepositoryError(t *testing.T) {
	repo := &mockCardRepository{
		createFn: func(card Card) (int64, error) { return 0, errors.New("db error") },
	}
	svc := &Service{repository: repo, mtgClient: newNilMTG()}
	_, err := svc.Create(CreateCardInput{Name: "Island"})
	if err == nil {
		t.Fatal("expected error from repository")
	}
}

// ── Service.List ─────────────────────────────────────────────────────────────

func TestServiceList(t *testing.T) {
	want := ListResult{Cards: []Card{{ID: 1, Name: "Island"}}, Total: 1, Page: 1, PageSize: 20, TotalPages: 1}
	repo := &mockCardRepository{
		listFn: func(params ListParams) (ListResult, error) { return want, nil },
	}
	svc := &Service{repository: repo}
	got, err := svc.List(ListParams{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	if got.Total != want.Total {
		t.Errorf("Total = %d, want %d", got.Total, want.Total)
	}
}

// ── Service.Delete ───────────────────────────────────────────────────────────

func TestServiceDelete(t *testing.T) {
	repo := &mockCardRepository{
		deleteFn: func(id string) error {
			if id != "7" {
				t.Errorf("expected id=7, got %q", id)
			}
			return nil
		},
	}
	svc := &Service{repository: repo}
	if err := svc.Delete("7"); err != nil {
		t.Fatal(err)
	}
}

func TestServiceDelete_Error(t *testing.T) {
	repo := &mockCardRepository{
		deleteFn: func(id string) error { return errors.New("not found") },
	}
	svc := &Service{repository: repo}
	if err := svc.Delete("99"); err == nil {
		t.Fatal("expected error")
	}
}

// ── Service.Update ───────────────────────────────────────────────────────────

func TestServiceUpdate_NoPropagate(t *testing.T) {
	current := &Card{ID: 1, Name: "Island", SetCode: "m11", CollectionNumber: "231", Language: "EN", Foil: false}
	repo := &mockCardRepository{
		getByIDFn: func(id string) (*Card, error) { return current, nil },
		updateFn:  func(id string, card Card) error { return nil },
	}
	svc := &Service{repository: repo}
	err := svc.Update("1", UpdateCardInput{Name: "Island", Propagate: false})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceUpdate_Propagate(t *testing.T) {
	current := &Card{ID: 1, Name: "Island", SetCode: "m11", CollectionNumber: "231", Language: "EN", Foil: false}
	propagateCalled := false
	repo := &mockCardRepository{
		getByIDFn: func(id string) (*Card, error) { return current, nil },
		updateSharedFn: func(_, _, _, _ string, _ bool, _ Card) error {
			propagateCalled = true
			return nil
		},
		updateFn: func(id string, card Card) error { return nil },
	}
	svc := &Service{repository: repo}
	err := svc.Update("1", UpdateCardInput{Name: "Island", Propagate: true})
	if err != nil {
		t.Fatal(err)
	}
	if !propagateCalled {
		t.Error("expected UpdateSharedByIdentity to be called")
	}
}

// ── Service.GetByID ──────────────────────────────────────────────────────────

func TestServiceGetByID_WithMTGID(t *testing.T) {
	card := &Card{ID: 1, Name: "Lightning Bolt", MTGID: "uuid-abc"}
	extCard := &mtgapi.ExternalCard{ID: "uuid-abc", Name: "Lightning Bolt"}
	repo := &mockCardRepository{
		getByIDFn:     func(id string) (*Card, error) { return card, nil },
		updateMTGIDFn: func(id, mtgID string) error { return nil },
	}
	mtg := &mockMTGClient{
		searchFn:     func(_, _, _, _ string) (*mtgapi.ExternalCard, error) { return nil, nil },
		searchPreFn:  func(_, _, _ string) (*mtgapi.ExternalCard, error) { return nil, nil },
		getByMTGIDFn: func(id string) (*mtgapi.ExternalCard, error) { return extCard, nil },
	}
	svc := &Service{repository: repo, mtgClient: mtg}
	result, err := svc.GetByID("1")
	if err != nil {
		t.Fatal(err)
	}
	if result["external"] == nil {
		t.Error("expected external card in result")
	}
}

func TestServiceGetByID_NotFound(t *testing.T) {
	repo := &mockCardRepository{
		getByIDFn: func(id string) (*Card, error) { return nil, errors.New("not found") },
	}
	svc := &Service{repository: repo}
	_, err := svc.GetByID("999")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ── Service.SetQuantity / SetDeck ────────────────────────────────────────────

func TestServiceSetQuantity(t *testing.T) {
	repo := &mockCardRepository{
		setQuantityFn: func(id string, q int) error {
			if id != "5" || q != 3 {
				t.Errorf("unexpected call: id=%s q=%d", id, q)
			}
			return nil
		},
	}
	svc := &Service{repository: repo}
	if err := svc.SetQuantity("5", 3); err != nil {
		t.Fatal(err)
	}
}

func TestServiceSetDeck(t *testing.T) {
	repo := &mockCardRepository{
		setDeckFn: func(id string, deckID int) error { return nil },
	}
	svc := &Service{repository: repo}
	if err := svc.SetDeck("3", 10); err != nil {
		t.Fatal(err)
	}
}

// ── Service.ExportAll ────────────────────────────────────────────────────────

func TestServiceExportAll(t *testing.T) {
	cards := []Card{{ID: 1}, {ID: 2}}
	repo := &mockCardRepository{
		listAllFn: func() ([]Card, error) { return cards, nil },
	}
	svc := &Service{repository: repo}
	got, err := svc.ExportAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 cards, got %d", len(got))
	}
}

// ── Service.GetStats / ListColorCombos / NormalizeRarities ──────────────────

func TestServiceGetStats(t *testing.T) {
	want := CollectionStats{UniqueCards: 100}
	repo := &mockCardRepository{
		getStatsFn: func() (CollectionStats, error) { return want, nil },
	}
	svc := &Service{repository: repo}
	got, err := svc.GetStats()
	if err != nil {
		t.Fatal(err)
	}
	if got.UniqueCards != 100 {
		t.Errorf("UniqueCards = %d, want 100", got.UniqueCards)
	}
}

func TestServiceListColorCombos(t *testing.T) {
	repo := &mockCardRepository{
		listColorCombosFn: func() ([]ColorCombo, error) { return []ColorCombo{{Codes: "W,U"}}, nil },
	}
	svc := &Service{repository: repo}
	got, err := svc.ListColorCombos()
	if err != nil || len(got) == 0 {
		t.Fatal("expected combos")
	}
}

func TestServiceNormalizeRarities(t *testing.T) {
	repo := &mockCardRepository{
		normalizeRaritiesFn: func() (NormalizeRarityResult, error) {
			return NormalizeRarityResult{Updated: 5}, nil
		},
	}
	svc := &Service{repository: repo}
	got, err := svc.NormalizeRarities()
	if err != nil || got.Updated != 5 {
		t.Fatal("unexpected result")
	}
}

// ── Service.Preview ──────────────────────────────────────────────────────────

func TestServicePreview_Normal(t *testing.T) {
	ext := &mtgapi.ExternalCard{Name: "Forest"}
	mtg := &mockMTGClient{
		searchFn:    func(_, _, _, _ string) (*mtgapi.ExternalCard, error) { return ext, nil },
		searchPreFn: func(_, _, _ string) (*mtgapi.ExternalCard, error) { return nil, nil },
		getByMTGIDFn: func(_ string) (*mtgapi.ExternalCard, error) { return nil, nil },
	}
	svc := &Service{mtgClient: mtg}
	got, err := svc.Preview(PreviewCardInput{SetCode: "M21", CollectionNumber: "278"})
	if err != nil || got == nil {
		t.Fatal("expected external card")
	}
}

func TestServicePreview_PreRelease(t *testing.T) {
	ext := &mtgapi.ExternalCard{Name: "Sol Ring"}
	mtg := &mockMTGClient{
		searchFn:    func(_, _, _, _ string) (*mtgapi.ExternalCard, error) { return nil, errors.New("should not call") },
		searchPreFn: func(_, _, _ string) (*mtgapi.ExternalCard, error) { return ext, nil },
		getByMTGIDFn: func(_ string) (*mtgapi.ExternalCard, error) { return nil, nil },
	}
	svc := &Service{mtgClient: mtg}
	got, err := svc.Preview(PreviewCardInput{Name: "Sol Ring", PreRelease: true})
	if err != nil || got == nil {
		t.Fatal("expected external card for prerelease")
	}
}
