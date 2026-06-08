package cards

import (
	"database/sql"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// cardCols lists the columns returned by selectCols (27 columns).
var cardCols = []string{
	"id", "mtg_id", "name", "color", "type", "subtitle", "collection_number",
	"rarity", "set_code", "mana_cost", "colors", "language", "year",
	"artist", "company", "foil", "quantity", "condition", "notes",
	"prerelease", "commander", "precon_deck", "deck_id", "price_usd", "image_url", "full_art",
}

// cardRow returns a standard slice of 27 values matching cardCols.
func cardRow(id int, name, setCode string) []driver.Value {
	return []driver.Value{
		int64(id), "mtgid-1", name, "Azul", "Instant", "", "100",
		"U", setCode, "{U}", `["U"]`, "EN", int64(2023),
		"Artist", "WOTC", int64(0), int64(1), "near_mint", "", int64(0), int64(0), "", int64(0), 1.50, "", int64(0),
	}
}

func newMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// ── List ─────────────────────────────────────────────────────────────────────

func TestCardsRepoList_Basic(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*), COALESCE(SUM(quantity), 0) FROM cards`)).
		WillReturnRows(sqlmock.NewRows([]string{"count", "qty"}).AddRow(2, 3))

	rows := sqlmock.NewRows(cardCols).
		AddRow(cardRow(1, "Counterspell", "GRN")...).
		AddRow(cardRow(2, "Lightning Bolt", "M21")...)
	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	result, err := r.List(ListParams{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Cards) != 2 {
		t.Errorf("cards = %d, want 2", len(result.Cards))
	}
	if result.Total != 2 {
		t.Errorf("total = %d, want 2", result.Total)
	}
}

func TestCardsRepoList_WithSearch(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count", "qty"}).AddRow(1, 1))
	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows(cardCols).AddRow(cardRow(1, "Counterspell", "GRN")...))

	_, err := r.List(ListParams{Search: "Counter"})
	if err != nil {
		t.Fatalf("List with search: %v", err)
	}
}

func TestCardsRepoList_WithNumberSearch(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count", "qty"}).AddRow(1, 1))
	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows(cardCols).AddRow(cardRow(1, "Counterspell", "GRN")...))

	_, err := r.List(ListParams{Search: "#100"})
	if err != nil {
		t.Fatalf("List with #search: %v", err)
	}
}

func TestCardsRepoList_WithDeckFilter(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	deckID := 5
	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count", "qty"}).AddRow(0, 0))
	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows(cardCols))

	_, err := r.List(ListParams{DeckIDFilter: &deckID})
	if err != nil {
		t.Fatalf("List with deckFilter: %v", err)
	}
}

func TestCardsRepoList_FoilAndRarityFilter(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count", "qty"}).AddRow(0, 0))
	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows(cardCols))

	_, err := r.List(ListParams{FoilOnly: true, RarityFilter: "R"})
	if err != nil {
		t.Fatalf("List foil+rarity: %v", err)
	}
}

func TestCardsRepoList_ColorsFilter(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count", "qty"}).AddRow(0, 0))
	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows(cardCols))

	_, err := r.List(ListParams{ColorsFilter: "U,G"})
	if err != nil {
		t.Fatalf("List with colorsFilter: %v", err)
	}
}

func TestCardsRepoList_ColorsFilterNone(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count", "qty"}).AddRow(0, 0))
	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows(cardCols))

	_, err := r.List(ListParams{ColorsFilter: "none"})
	if err != nil {
		t.Fatalf("List with colorsFilter=none: %v", err)
	}
}

func TestCardsRepoList_CountError(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT COUNT`).WillReturnError(sql.ErrConnDone)

	_, err := r.List(ListParams{})
	if err == nil {
		t.Error("expected error on count query failure")
	}
}

func TestCardsRepoList_QueryError(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count", "qty"}).AddRow(1, 1))
	mock.ExpectQuery(`SELECT`).WillReturnError(sql.ErrConnDone)

	_, err := r.List(ListParams{})
	if err == nil {
		t.Error("expected error on list query failure")
	}
}

func TestCardsRepoList_SortAndOrderDesc(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count", "qty"}).AddRow(0, 0))
	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows(cardCols))

	_, err := r.List(ListParams{Sort: "price_usd", Order: "desc"})
	if err != nil {
		t.Fatalf("List sort/order: %v", err)
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestCardsRepoCreate_OK(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectPrepare("INSERT INTO cards").
		ExpectExec().
		WillReturnResult(sqlmock.NewResult(42, 1))

	id, err := r.Create(Card{Name: "Counterspell", Foil: true, PreRelease: true, Commander: true, FullArt: true})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 42 {
		t.Errorf("id = %d, want 42", id)
	}
}

func TestCardsRepoCreate_PrepareError(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectPrepare("INSERT INTO cards").WillReturnError(sql.ErrConnDone)

	_, err := r.Create(Card{Name: "x"})
	if err == nil {
		t.Error("expected error on prepare failure")
	}
}

func TestCardsRepoCreate_ExecError(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectPrepare("INSERT INTO cards").
		ExpectExec().WillReturnError(sql.ErrConnDone)

	_, err := r.Create(Card{Name: "x"})
	if err == nil {
		t.Error("expected error on exec failure")
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestCardsRepoGetByID_OK(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT`).
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows(cardCols).AddRow(cardRow(1, "Counterspell", "GRN")...))

	c, err := r.GetByID("1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if c.Name != "Counterspell" {
		t.Errorf("name = %q, want Counterspell", c.Name)
	}
}

func TestCardsRepoGetByID_NotFound(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT`).
		WithArgs("999").
		WillReturnError(sql.ErrNoRows)

	_, err := r.GetByID("999")
	if err == nil {
		t.Error("expected error for not found")
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestCardsRepoUpdate_OK(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE cards SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := r.Update("1", Card{Name: "x", Foil: false, PreRelease: false, Commander: false, FullArt: false})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestCardsRepoUpdate_WithFlags(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE cards SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := r.Update("2", Card{Name: "Foil Card", Foil: true, PreRelease: true, Commander: true, FullArt: true})
	if err != nil {
		t.Fatalf("Update with flags: %v", err)
	}
}

// ── UpdateSharedByIdentity ────────────────────────────────────────────────────

func TestCardsRepoUpdateSharedByIdentity_OK(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE cards SET`).
		WillReturnResult(sqlmock.NewResult(0, 3))

	err := r.UpdateSharedByIdentity("Bolt", "M21", "100", "EN", false, Card{Name: "Lightning Bolt", Foil: true})
	if err != nil {
		t.Fatalf("UpdateSharedByIdentity: %v", err)
	}
}

// ── Simple updates ────────────────────────────────────────────────────────────

func TestCardsRepoUpdateMTGID(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectExec(`UPDATE cards SET mtg_id`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.UpdateMTGID("1", "abc-123"); err != nil {
		t.Fatalf("UpdateMTGID: %v", err)
	}
}

func TestCardsRepoDelete_OK(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectExec(`DELETE FROM cards`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.Delete("1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestCardsRepoSetQuantity(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectExec(`UPDATE cards SET quantity`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.SetQuantity("1", 3); err != nil {
		t.Fatalf("SetQuantity: %v", err)
	}
}

func TestCardsRepoSetDeck(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectExec(`UPDATE cards SET deck_id`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.SetDeck("1", 5); err != nil {
		t.Fatalf("SetDeck: %v", err)
	}
}

func TestCardsRepoUpdatePrice(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectExec(`UPDATE cards SET price_usd`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.UpdatePrice(1, 2.99); err != nil {
		t.Fatalf("UpdatePrice: %v", err)
	}
}

func TestCardsRepoUpdatePriceAndMTGID(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectExec(`UPDATE cards SET price_usd`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.UpdatePriceAndMTGID(1, "mtg-1", 5.00); err != nil {
		t.Fatalf("UpdatePriceAndMTGID: %v", err)
	}
}

func TestCardsRepoUpdateImageURL(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectExec(`UPDATE cards SET image_url`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.UpdateImageURL(1, "http://img"); err != nil {
		t.Fatalf("UpdateImageURL: %v", err)
	}
}

func TestCardsRepoUpdateImageURLAndMTGID(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectExec(`UPDATE cards SET image_url`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.UpdateImageURLAndMTGID(1, "mtg-1", "http://img"); err != nil {
		t.Fatalf("UpdateImageURLAndMTGID: %v", err)
	}
}

func TestCardsRepoUpdateColors(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectExec(`UPDATE cards SET colors`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.UpdateColors(1, `["U"]`, "Azul"); err != nil {
		t.Fatalf("UpdateColors: %v", err)
	}
}

func TestCardsRepoUpdateColorsAndMTGID(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectExec(`UPDATE cards SET colors`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.UpdateColorsAndMTGID(1, "mtg-1", `["U"]`, "Azul"); err != nil {
		t.Fatalf("UpdateColorsAndMTGID: %v", err)
	}
}

// ── ListAll ───────────────────────────────────────────────────────────────────

func TestCardsRepoListAll_OK(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows(cardCols).AddRow(cardRow(1, "Counterspell", "GRN")...))

	cards, err := r.ListAll()
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(cards) != 1 {
		t.Errorf("len = %d, want 1", len(cards))
	}
}

func TestCardsRepoListAll_QueryError(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectQuery(`SELECT`).WillReturnError(sql.ErrConnDone)
	_, err := r.ListAll()
	if err == nil {
		t.Error("expected error")
	}
}

// ── listForPriceRefresh ───────────────────────────────────────────────────────

var priceRefreshCols = []string{"id", "mtg_id", "set_code", "collection_number", "language", "artist", "foil", "name"}

func TestCardsRepoListAllForPriceRefresh_OK(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id`).
		WillReturnRows(sqlmock.NewRows(priceRefreshCols).
			AddRow(1, "mtg-1", "GRN", "100", "EN", "Art", 0, "Bolt"))

	cards, err := r.ListAllForPriceRefresh()
	if err != nil {
		t.Fatalf("ListAllForPriceRefresh: %v", err)
	}
	if len(cards) != 1 || cards[0].Name != "Bolt" {
		t.Errorf("unexpected result: %+v", cards)
	}
}

func TestCardsRepoListEmptyPricesForRefresh_OK(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id`).
		WillReturnRows(sqlmock.NewRows(priceRefreshCols).
			AddRow(2, "", "M21", "99", "EN", "", 1, "Foil"))

	cards, err := r.ListEmptyPricesForRefresh()
	if err != nil {
		t.Fatalf("ListEmptyPricesForRefresh: %v", err)
	}
	if len(cards) != 1 || !cards[0].Foil {
		t.Errorf("unexpected result: %+v", cards)
	}
}

func TestCardsRepoListAllForPriceRefresh_Error(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectQuery(`SELECT id`).WillReturnError(sql.ErrConnDone)
	_, err := r.ListAllForPriceRefresh()
	if err == nil {
		t.Error("expected error")
	}
}

// ── ListCardsWithoutColors ────────────────────────────────────────────────────

func TestCardsRepoListCardsWithoutColors_OK(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id`).
		WillReturnRows(sqlmock.NewRows(priceRefreshCols).
			AddRow(3, "", "GRN", "50", "EN", "", 0, "Mystery"))

	cards, err := r.ListCardsWithoutColors()
	if err != nil {
		t.Fatalf("ListCardsWithoutColors: %v", err)
	}
	if len(cards) != 1 {
		t.Errorf("len = %d, want 1", len(cards))
	}
}

func TestCardsRepoListCardsWithoutColors_Error(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectQuery(`SELECT id`).WillReturnError(sql.ErrConnDone)
	_, err := r.ListCardsWithoutColors()
	if err == nil {
		t.Error("expected error")
	}
}

// ── NormalizeRarities ─────────────────────────────────────────────────────────

func TestCardsRepoNormalizeRarities_OK(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	// NormalizeRarities iterates over conversions map (non-deterministic order).
	// Map has 23 entries; set up 30 expectations to be safe.
	mock.MatchExpectationsInOrder(false)
	for i := 0; i < 30; i++ {
		mock.ExpectExec(`UPDATE cards SET rarity`).
			WillReturnResult(sqlmock.NewResult(0, 0))
	}

	result, err := r.NormalizeRarities()
	if err != nil {
		t.Fatalf("NormalizeRarities: %v", err)
	}
	_ = result
}

func TestCardsRepoNormalizeRarities_ExecError(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.MatchExpectationsInOrder(false)
	mock.ExpectExec(`UPDATE cards SET rarity`).WillReturnError(sql.ErrConnDone)
	// other calls won't execute due to early return
	for i := 0; i < 20; i++ {
		mock.ExpectExec(`UPDATE cards SET rarity`).WillReturnResult(sqlmock.NewResult(0, 0))
	}

	_, err := r.NormalizeRarities()
	if err == nil {
		t.Error("expected error")
	}
}

// ── ListColorCombos ───────────────────────────────────────────────────────────

func TestCardsRepoListColorCombos_OK(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT colors`).
		WillReturnRows(sqlmock.NewRows([]string{"colors", "cnt"}).
			AddRow(`["U"]`, 5).
			AddRow(`["W","U"]`, 2))

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	combos, err := r.ListColorCombos()
	if err != nil {
		t.Fatalf("ListColorCombos: %v", err)
	}
	if len(combos) < 2 {
		t.Errorf("combos = %d, want >=2", len(combos))
	}
}

func TestCardsRepoListColorCombos_Error(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectQuery(`SELECT colors`).WillReturnError(sql.ErrConnDone)
	_, err := r.ListColorCombos()
	if err == nil {
		t.Error("expected error")
	}
}

// ── GetStats ──────────────────────────────────────────────────────────────────

func TestCardsRepoGetStats_OK(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	// Main totals query
	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"unique", "total_qty", "foil_count", "foil_qty", "est_val", "priced"}).
			AddRow(100, 200, 10, 15, 50.0, 80))

	// By rarity
	mock.ExpectQuery(`SELECT COALESCE`).
		WillReturnRows(sqlmock.NewRows([]string{"r", "count", "qty"}).
			AddRow("R", 20, 30).AddRow("C", 80, 170))

	// Top sets
	mock.ExpectQuery(`SELECT UPPER`).
		WillReturnRows(sqlmock.NewRows([]string{"sc", "count", "qty"}).
			AddRow("GRN", 50, 100))

	// Color distribution
	mock.ExpectQuery(`SELECT COALESCE\(colors`).
		WillReturnRows(sqlmock.NewRows([]string{"colors"}).
			AddRow(`["U"]`).AddRow(`[]`))

	stats, err := r.GetStats()
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.UniqueCards != 100 {
		t.Errorf("UniqueCards = %d, want 100", stats.UniqueCards)
	}
}

func TestCardsRepoGetStats_MainQueryError(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectQuery(`SELECT COUNT`).WillReturnError(sql.ErrConnDone)
	_, err := r.GetStats()
	if err == nil {
		t.Error("expected error")
	}
}

// ── ListForDeckBuilder ────────────────────────────────────────────────────────

func TestCardsRepoListForDeckBuilder_OK(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT name`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "type", "mana_cost", "rarity", "colors", "total_qty"}).
			AddRow("Bolt", "Instant", "{R}", "C", `["R"]`, 4))

	cards, err := r.ListForDeckBuilder()
	if err != nil {
		t.Fatalf("ListForDeckBuilder: %v", err)
	}
	if len(cards) != 1 || cards[0].Name != "Bolt" {
		t.Errorf("unexpected: %+v", cards)
	}
}

func TestCardsRepoListForDeckBuilder_Error(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectQuery(`SELECT name`).WillReturnError(sql.ErrConnDone)
	_, err := r.ListForDeckBuilder()
	if err == nil {
		t.Error("expected error")
	}
}

// ── ListForEval ───────────────────────────────────────────────────────────────

func TestCardsRepoListForEval_OK(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT name`).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"name", "type", "mana_cost", "rarity"}).
			AddRow("Bolt", "Instant", "{R}", "C"))

	cards, err := r.ListForEval(7)
	if err != nil {
		t.Fatalf("ListForEval: %v", err)
	}
	if len(cards) != 1 {
		t.Errorf("len = %d, want 1", len(cards))
	}
}

func TestCardsRepoListForEval_Error(t *testing.T) {
	db, mock := newMock(t)
	r := NewRepository(db)
	mock.ExpectQuery(`SELECT name`).WithArgs(0).WillReturnError(sql.ErrConnDone)
	_, err := r.ListForEval(0)
	if err == nil {
		t.Error("expected error")
	}
}
