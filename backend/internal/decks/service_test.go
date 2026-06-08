package decks

import (
	"errors"
	"strings"
	"testing"

	"magic-collection-api/internal/cards"
	"magic-collection-api/internal/mtgapi"
)

// ── mock repository ───────────────────────────────────────────────────────────

type mockDeckRepo struct {
	listFn             func() ([]Deck, error)
	getByIDFn          func(id string) (*Deck, error)
	createFn           func(input DeckInput) (int64, error)
	updateFn           func(id string, input DeckInput) error
	updateIconFn       func(id, iconURI string) error
	updateEvaluationFn func(id, evaluation string) error
	deleteFn           func(id string) error
}

func (m *mockDeckRepo) List() ([]Deck, error)                             { return m.listFn() }
func (m *mockDeckRepo) GetByID(id string) (*Deck, error)                  { return m.getByIDFn(id) }
func (m *mockDeckRepo) Create(i DeckInput) (int64, error)                 { return m.createFn(i) }
func (m *mockDeckRepo) Update(id string, i DeckInput) error               { return m.updateFn(id, i) }
func (m *mockDeckRepo) UpdateIcon(id, uri string) error                   { return m.updateIconFn(id, uri) }
func (m *mockDeckRepo) UpdateEvaluation(id, eval string) error             { return m.updateEvaluationFn(id, eval) }
func (m *mockDeckRepo) Delete(id string) error                            { return m.deleteFn(id) }

type mockDeckMtg struct {
	getSetFn func(code string) (*mtgapi.SetInfo, error)
}

func (m *mockDeckMtg) GetSetByCode(code string) (*mtgapi.SetInfo, error) { return m.getSetFn(code) }

type mockDeckCardRepo struct {
	listForEvalFn func(deckID int) ([]cards.EvalCardInfo, error)
}

func (m *mockDeckCardRepo) ListForEval(deckID int) ([]cards.EvalCardInfo, error) {
	return m.listForEvalFn(deckID)
}

type mockDeckAI struct {
	completeFn func(prompt string) (string, error)
}

func (m *mockDeckAI) Complete(prompt string) (string, error) { return m.completeFn(prompt) }

// ── buildEvalPrompt ───────────────────────────────────────────────────────────

func TestBuildEvalPrompt_ContainsDeckName(t *testing.T) {
	deck := &Deck{Name: "Grixis Control", Colors: "U,B,R", Commander: true}
	evalCards := []cards.EvalCardInfo{
		{Name: "Counterspell", Type: "Instant", ManaCost: "{U}{U}"},
		{Name: "Swamp", Type: "Basic Land — Swamp", ManaCost: ""},
	}
	prompt := buildEvalPrompt(deck, evalCards)
	if !strings.Contains(prompt, "Grixis Control") {
		t.Error("prompt should contain deck name")
	}
	if !strings.Contains(prompt, "Commander") {
		t.Error("prompt should mention Commander format")
	}
	if !strings.Contains(prompt, "Counterspell") {
		t.Error("prompt should list card names")
	}
}

func TestBuildEvalPrompt_NoColors(t *testing.T) {
	deck := &Deck{Name: "Colorless", Colors: ""}
	prompt := buildEvalPrompt(deck, []cards.EvalCardInfo{{Name: "Eldrazi", Type: "Creature"}})
	if !strings.Contains(prompt, "Incolor") {
		t.Error("prompt should say Incolor when no colors")
	}
}

func TestBuildEvalPrompt_CategorizesByType(t *testing.T) {
	deck := &Deck{Name: "Test", Commander: true}
	evalCards := []cards.EvalCardInfo{
		{Name: "Birds of Paradise", Type: "Creature — Bird"},
		{Name: "Sol Ring", Type: "Artifact"},
		{Name: "Counterspell", Type: "Instant"},
		{Name: "Wrath of God", Type: "Sorcery"},
		{Name: "Ghostly Prison", Type: "Enchantment"},
		{Name: "Jace TMS", Type: "Planeswalker"},
		{Name: "Forest", Type: "Basic Land — Forest"},
		{Name: "Unknown Thing", Type: "Scheme"},
	}
	prompt := buildEvalPrompt(deck, evalCards)
	for _, section := range []string{"Criatura", "Artefato", "Mágica Imediata", "Feitiço", "Encantamento", "Planeswalker", "Terreno", "Outros"} {
		if !strings.Contains(prompt, section) {
			t.Errorf("prompt should contain section %q", section)
		}
	}
}

// ── Service.List ──────────────────────────────────────────────────────────────

func TestDeckServiceList(t *testing.T) {
	repo := &mockDeckRepo{
		listFn: func() ([]Deck, error) { return []Deck{{ID: 1, Name: "Aggro"}}, nil },
	}
	svc := &Service{repo: repo}
	got, err := svc.List()
	if err != nil || len(got) == 0 {
		t.Fatal("expected decks")
	}
}

// ── Service.Create ────────────────────────────────────────────────────────────

func TestDeckServiceCreate_NoSetCode(t *testing.T) {
	repo := &mockDeckRepo{
		createFn: func(i DeckInput) (int64, error) { return 5, nil },
	}
	svc := &Service{repo: repo, mtgClient: &mockDeckMtg{}}
	id, err := svc.Create(DeckInput{Name: "Test"})
	if err != nil || id != 5 {
		t.Fatalf("expected id=5, got %d err=%v", id, err)
	}
}

func TestDeckServiceCreate_WithSetCode_FetchesIcon(t *testing.T) {
	iconCalled := false
	repo := &mockDeckRepo{
		createFn:     func(i DeckInput) (int64, error) { return 1, nil },
		updateIconFn: func(id, uri string) error { iconCalled = true; return nil },
	}
	mtg := &mockDeckMtg{
		getSetFn: func(code string) (*mtgapi.SetInfo, error) {
			return &mtgapi.SetInfo{Code: code, IconSVGURI: "https://svg.example.com/icon.svg"}, nil
		},
	}
	svc := &Service{repo: repo, mtgClient: mtg}
	_, err := svc.Create(DeckInput{Name: "Test", SetCode: "M21"})
	if err != nil {
		t.Fatal(err)
	}
	if !iconCalled {
		t.Error("expected UpdateIcon to be called")
	}
}

func TestDeckServiceCreate_RepoError(t *testing.T) {
	repo := &mockDeckRepo{
		createFn: func(i DeckInput) (int64, error) { return 0, errors.New("db error") },
	}
	svc := &Service{repo: repo}
	_, err := svc.Create(DeckInput{Name: "Test"})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ── Service.Update ────────────────────────────────────────────────────────────

func TestDeckServiceUpdate_OK(t *testing.T) {
	repo := &mockDeckRepo{
		updateFn:     func(id string, i DeckInput) error { return nil },
		updateIconFn: func(id, uri string) error { return nil },
	}
	mtg := &mockDeckMtg{
		getSetFn: func(code string) (*mtgapi.SetInfo, error) {
			return &mtgapi.SetInfo{IconSVGURI: "uri"}, nil
		},
	}
	svc := &Service{repo: repo, mtgClient: mtg}
	err := svc.Update("1", DeckInput{Name: "Updated", SetCode: "GRN"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeckServiceUpdate_RepoError(t *testing.T) {
	repo := &mockDeckRepo{
		updateFn: func(id string, i DeckInput) error { return errors.New("fail") },
	}
	svc := &Service{repo: repo}
	err := svc.Update("1", DeckInput{Name: "Test"})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ── Service.Delete ────────────────────────────────────────────────────────────

func TestDeckServiceDelete(t *testing.T) {
	repo := &mockDeckRepo{
		deleteFn: func(id string) error { return nil },
	}
	svc := &Service{repo: repo}
	if err := svc.Delete("1"); err != nil {
		t.Fatal(err)
	}
}

// ── Service.FetchIcon ─────────────────────────────────────────────────────────

func TestDeckServiceFetchIcon_NoSetCode(t *testing.T) {
	repo := &mockDeckRepo{
		getByIDFn: func(id string) (*Deck, error) { return &Deck{ID: 1}, nil },
	}
	svc := &Service{repo: repo}
	uri, err := svc.FetchIcon("1")
	if err != nil || uri != "" {
		t.Errorf("expected empty uri and no error for deck without set_code, got uri=%q err=%v", uri, err)
	}
}

func TestDeckServiceFetchIcon_OK(t *testing.T) {
	repo := &mockDeckRepo{
		getByIDFn:    func(id string) (*Deck, error) { return &Deck{ID: 1, SetCode: "M21"}, nil },
		updateIconFn: func(id, uri string) error { return nil },
	}
	mtg := &mockDeckMtg{
		getSetFn: func(code string) (*mtgapi.SetInfo, error) {
			return &mtgapi.SetInfo{IconSVGURI: "https://icon.svg"}, nil
		},
	}
	svc := &Service{repo: repo, mtgClient: mtg}
	uri, err := svc.FetchIcon("1")
	if err != nil || uri == "" {
		t.Errorf("expected non-empty URI, got %q err=%v", uri, err)
	}
}

// ── Service.EvaluateDeck ──────────────────────────────────────────────────────

func TestDeckServiceEvaluateDeck_NoCards(t *testing.T) {
	repo := &mockDeckRepo{
		getByIDFn: func(id string) (*Deck, error) { return &Deck{ID: 1, Name: "Empty"}, nil },
	}
	cardRepo := &mockDeckCardRepo{
		listForEvalFn: func(deckID int) ([]cards.EvalCardInfo, error) { return nil, nil },
	}
	svc := &Service{repo: repo, cardRepo: cardRepo}
	_, err := svc.EvaluateDeck("1")
	if err == nil {
		t.Fatal("expected error for deck with no cards")
	}
}

func TestDeckServiceEvaluateDeck_OK(t *testing.T) {
	evalCards := []cards.EvalCardInfo{{Name: "Island", Type: "Basic Land"}}
	repo := &mockDeckRepo{
		getByIDFn:          func(id string) (*Deck, error) { return &Deck{ID: 1, Name: "Test"}, nil },
		updateEvaluationFn: func(id, eval string) error { return nil },
	}
	cardRepo := &mockDeckCardRepo{
		listForEvalFn: func(deckID int) ([]cards.EvalCardInfo, error) { return evalCards, nil },
	}
	aiClient := &mockDeckAI{
		completeFn: func(prompt string) (string, error) { return "## Analysis\nGreat deck!", nil },
	}
	svc := &Service{repo: repo, cardRepo: cardRepo, aiClient: aiClient}
	deck, err := svc.EvaluateDeck("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(deck.Evaluation, "Great deck!") {
		t.Error("expected evaluation text in deck")
	}
}

func TestDeckServiceEvaluateDeck_AIError(t *testing.T) {
	evalCards := []cards.EvalCardInfo{{Name: "Forest", Type: "Land"}}
	repo := &mockDeckRepo{
		getByIDFn: func(id string) (*Deck, error) { return &Deck{ID: 1, Name: "Test"}, nil },
	}
	cardRepo := &mockDeckCardRepo{
		listForEvalFn: func(deckID int) ([]cards.EvalCardInfo, error) { return evalCards, nil },
	}
	aiClient := &mockDeckAI{
		completeFn: func(prompt string) (string, error) { return "", errors.New("api error") },
	}
	svc := &Service{repo: repo, cardRepo: cardRepo, aiClient: aiClient}
	_, err := svc.EvaluateDeck("1")
	if err == nil {
		t.Fatal("expected error from AI")
	}
}
