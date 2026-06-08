package wishlist

import (
	"errors"
	"testing"

	"magic-collection-api/internal/mtgapi"
)

// ── parsePrice ────────────────────────────────────────────────────────────────

func TestParsePrice(t *testing.T) {
	tests := []struct {
		name   string
		prices map[string]string
		key    string
		want   float64
	}{
		{"nil map", nil, "usd", 0},
		{"key not found", map[string]string{"eur": "1.50"}, "usd", 0},
		{"empty value", map[string]string{"usd": ""}, "usd", 0},
		{"valid usd", map[string]string{"usd": "2.50"}, "usd", 2.50},
		{"valid usd_foil", map[string]string{"usd_foil": "5.00"}, "usd_foil", 5.00},
		{"invalid value", map[string]string{"usd": "abc"}, "usd", 0},
		{"zero price", map[string]string{"usd": "0.00"}, "usd", 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parsePrice(tc.prices, tc.key)
			if got != tc.want {
				t.Errorf("parsePrice(%v, %q) = %v, want %v", tc.prices, tc.key, got, tc.want)
			}
		})
	}
}

// ── colorsJSONToDisplay ───────────────────────────────────────────────────────

func TestWishlistColorsJSONToDisplay(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"null", ""},
		{"[]", ""},
		{`["W"]`, "Branco"},
		{`["U","B"]`, "Azul/Preto"},
		{"invalid", ""},
	}
	for _, tc := range tests {
		got := colorsJSONToDisplay(tc.input)
		if got != tc.want {
			t.Errorf("colorsJSONToDisplay(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ── mock repository ───────────────────────────────────────────────────────────

type mockWishlistRepo struct {
	listFn     func() ([]WishlistCard, error)
	getByIDFn  func(id string) (*WishlistCard, error)
	createFn   func(w WishlistCard) (int64, error)
	updateFn   func(id string, input WishlistUpdateInput) error
	deleteFn   func(id string) error
	acquireFn  func(id string, input AcquireInput) (int64, error)
}

func (m *mockWishlistRepo) List() ([]WishlistCard, error) { return m.listFn() }
func (m *mockWishlistRepo) GetByID(id string) (*WishlistCard, error) { return m.getByIDFn(id) }
func (m *mockWishlistRepo) Create(w WishlistCard) (int64, error)     { return m.createFn(w) }
func (m *mockWishlistRepo) Update(id string, i WishlistUpdateInput) error { return m.updateFn(id, i) }
func (m *mockWishlistRepo) Delete(id string) error                        { return m.deleteFn(id) }
func (m *mockWishlistRepo) Acquire(id string, i AcquireInput) (int64, error) {
	return m.acquireFn(id, i)
}

type mockWishlistMtg struct {
	searchFn func(setCode, number, lang, artist string) (*mtgapi.ExternalCard, error)
}

func (m *mockWishlistMtg) Search(setCode, number, lang, artist string) (*mtgapi.ExternalCard, error) {
	return m.searchFn(setCode, number, lang, artist)
}

// ── Service.List ──────────────────────────────────────────────────────────────

func TestWishlistServiceList(t *testing.T) {
	repo := &mockWishlistRepo{
		listFn: func() ([]WishlistCard, error) { return []WishlistCard{{ID: 1}}, nil },
	}
	svc := &Service{repo: repo}
	got, err := svc.List()
	if err != nil || len(got) == 0 {
		t.Fatal("expected items")
	}
}

// ── Service.GetByID ───────────────────────────────────────────────────────────

func TestWishlistServiceGetByID_OK(t *testing.T) {
	repo := &mockWishlistRepo{
		getByIDFn: func(id string) (*WishlistCard, error) { return &WishlistCard{ID: 1}, nil },
	}
	svc := &Service{repo: repo}
	item, err := svc.GetByID("1")
	if err != nil || item == nil {
		t.Fatal("expected item")
	}
}

func TestWishlistServiceGetByID_Error(t *testing.T) {
	repo := &mockWishlistRepo{
		getByIDFn: func(id string) (*WishlistCard, error) { return nil, errors.New("not found") },
	}
	svc := &Service{repo: repo}
	_, err := svc.GetByID("99")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ── Service.Create ────────────────────────────────────────────────────────────

func TestWishlistServiceCreate_NoScryfall(t *testing.T) {
	repo := &mockWishlistRepo{
		createFn: func(w WishlistCard) (int64, error) {
			if w.SetCode != "M21" {
				t.Errorf("SetCode = %q, want M21", w.SetCode)
			}
			return 5, nil
		},
	}
	mtg := &mockWishlistMtg{
		searchFn: func(_, _, _, _ string) (*mtgapi.ExternalCard, error) { return nil, nil },
	}
	svc := &Service{repo: repo, mtgClient: mtg}
	id, err := svc.Create(WishlistCardInput{SetCode: "M21", CollectionNumber: "278"})
	if err != nil || id != 5 {
		t.Fatalf("expected id=5, got %d err=%v", id, err)
	}
}

func TestWishlistServiceCreate_WithScryfall(t *testing.T) {
	ext := &mtgapi.ExternalCard{
		ID:       "uuid-x",
		Name:     "Island",
		Rarity:   "L",
		Colors:   []string{},
		ImageURL: "https://img.example.com/island.jpg",
		Prices:   map[string]string{"usd": "0.10", "usd_foil": "0.50"},
	}
	var savedCard WishlistCard
	repo := &mockWishlistRepo{
		createFn: func(w WishlistCard) (int64, error) { savedCard = w; return 1, nil },
	}
	mtg := &mockWishlistMtg{
		searchFn: func(_, _, _, _ string) (*mtgapi.ExternalCard, error) { return ext, nil },
	}
	svc := &Service{repo: repo, mtgClient: mtg}
	_, err := svc.Create(WishlistCardInput{SetCode: "M21", CollectionNumber: "278"})
	if err != nil {
		t.Fatal(err)
	}
	if savedCard.MTGID != "uuid-x" {
		t.Errorf("MTGID = %q, want uuid-x", savedCard.MTGID)
	}
	if savedCard.PriceUSD != 0.10 {
		t.Errorf("PriceUSD = %v, want 0.10", savedCard.PriceUSD)
	}
	if savedCard.PriceUSDFoil != 0.50 {
		t.Errorf("PriceUSDFoil = %v, want 0.50", savedCard.PriceUSDFoil)
	}
}

func TestWishlistServiceCreate_NormalizesSetCode(t *testing.T) {
	var searchedCode string
	repo := &mockWishlistRepo{
		createFn: func(w WishlistCard) (int64, error) { return 1, nil },
	}
	mtg := &mockWishlistMtg{
		searchFn: func(setCode, _, _, _ string) (*mtgapi.ExternalCard, error) {
			searchedCode = setCode
			return nil, nil
		},
	}
	svc := &Service{repo: repo, mtgClient: mtg}
	_, _ = svc.Create(WishlistCardInput{SetCode: "m21", CollectionNumber: "1"})
	if searchedCode != "M21" {
		t.Errorf("expected uppercased set code, got %q", searchedCode)
	}
}

// ── Service.Update / Delete ───────────────────────────────────────────────────

func TestWishlistServiceUpdate(t *testing.T) {
	repo := &mockWishlistRepo{
		updateFn: func(id string, i WishlistUpdateInput) error { return nil },
	}
	svc := &Service{repo: repo}
	if err := svc.Update("1", WishlistUpdateInput{Reason: "want it"}); err != nil {
		t.Fatal(err)
	}
}

func TestWishlistServiceDelete(t *testing.T) {
	repo := &mockWishlistRepo{
		deleteFn: func(id string) error { return nil },
	}
	svc := &Service{repo: repo}
	if err := svc.Delete("1"); err != nil {
		t.Fatal(err)
	}
}

// ── Service.Acquire ───────────────────────────────────────────────────────────

func TestWishlistServiceAcquire(t *testing.T) {
	repo := &mockWishlistRepo{
		acquireFn: func(id string, i AcquireInput) (int64, error) { return 42, nil },
	}
	svc := &Service{repo: repo}
	cardID, err := svc.Acquire("1", AcquireInput{Condition: "near_mint"})
	if err != nil || cardID != 42 {
		t.Fatalf("expected cardID=42, got %d err=%v", cardID, err)
	}
}
