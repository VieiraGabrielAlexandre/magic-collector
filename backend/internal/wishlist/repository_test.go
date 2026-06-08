package wishlist

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var wishlistCols = []string{
	"id", "mtg_id", "set_code", "collection_number", "name", "printed_name",
	"image_uri", "artist", "rarity", "colors", "color",
	"price_usd", "price_usd_foil",
	"foil", "reason", "acquired", "created_at",
}

func newWishMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func wishRow(id int, name string) []driver.Value {
	return []driver.Value{
		int64(id), "mtg-1", "GRN", "100", name, name,
		"https://img", "Artist", "R", `["R"]`, "Vermelho",
		1.50, 0.0,
		int64(0), "want it", int64(0), "2024-01-01T00:00:00",
	}
}

// ── List ─────────────────────────────────────────────────────────────────────

func TestWishlistRepoList_OK(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id`).
		WillReturnRows(sqlmock.NewRows(wishlistCols).
			AddRow(wishRow(1, "Lightning Bolt")...).
			AddRow(wishRow(2, "Counterspell")...))

	items, err := r.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("len = %d, want 2", len(items))
	}
}

func TestWishlistRepoList_Error(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)
	mock.ExpectQuery(`SELECT id`).WillReturnError(sql.ErrConnDone)
	_, err := r.List()
	if err == nil {
		t.Error("expected error")
	}
}

func TestWishlistRepoList_FoilAndAcquired(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)

	row := []driver.Value{
		int64(1), "mtg-1", "GRN", "100", "Bolt", "Bolt",
		"", "Art", "R", `["R"]`, "Vermelho",
		2.0, 5.0,
		int64(1), "", int64(1), "2024-01-01T00:00:00",
	}
	mock.ExpectQuery(`SELECT id`).
		WillReturnRows(sqlmock.NewRows(wishlistCols).AddRow(row...))

	items, err := r.List()
	if err != nil {
		t.Fatalf("List foil+acquired: %v", err)
	}
	if !items[0].Foil || !items[0].Acquired {
		t.Errorf("foil=%v acquired=%v, want both true", items[0].Foil, items[0].Acquired)
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestWishlistRepoGetByID_OK(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id`).
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows(wishlistCols).AddRow(wishRow(1, "Bolt")...))

	w, err := r.GetByID("1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if w.Name != "Bolt" {
		t.Errorf("name = %q, want Bolt", w.Name)
	}
}

func TestWishlistRepoGetByID_NotFound(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)
	mock.ExpectQuery(`SELECT id`).WithArgs("99").WillReturnError(sql.ErrNoRows)
	_, err := r.GetByID("99")
	if err == nil {
		t.Error("expected error")
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestWishlistRepoCreate_OK(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`INSERT INTO wishlist_cards`).
		WillReturnResult(sqlmock.NewResult(7, 1))

	id, err := r.Create(WishlistCard{Name: "Bolt", PriceUSD: 1.50, Foil: false})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 7 {
		t.Errorf("id = %d, want 7", id)
	}
}

func TestWishlistRepoCreate_FoilWithPrices(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`INSERT INTO wishlist_cards`).
		WillReturnResult(sqlmock.NewResult(8, 1))

	id, err := r.Create(WishlistCard{Name: "Foil Bolt", PriceUSD: 1.50, PriceUSDFoil: 3.00, Foil: true})
	if err != nil {
		t.Fatalf("Create foil: %v", err)
	}
	if id != 8 {
		t.Errorf("id = %d, want 8", id)
	}
}

func TestWishlistRepoCreate_Error(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)
	mock.ExpectExec(`INSERT INTO wishlist_cards`).WillReturnError(sql.ErrConnDone)
	_, err := r.Create(WishlistCard{Name: "x"})
	if err == nil {
		t.Error("expected error")
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestWishlistRepoUpdate_OK(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE wishlist_cards SET foil`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := r.Update("1", WishlistUpdateInput{Foil: false, Reason: "updated"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestWishlistRepoUpdate_Foil(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE wishlist_cards SET foil`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := r.Update("2", WishlistUpdateInput{Foil: true})
	if err != nil {
		t.Fatalf("Update foil: %v", err)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestWishlistRepoDelete_OK(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)
	mock.ExpectExec(`DELETE FROM wishlist_cards`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.Delete("1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestWishlistRepoDelete_Error(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)
	mock.ExpectExec(`DELETE FROM wishlist_cards`).WillReturnError(sql.ErrConnDone)
	if err := r.Delete("1"); err == nil {
		t.Error("expected error")
	}
}

// ── Acquire ───────────────────────────────────────────────────────────────────

func TestWishlistRepoAcquire_OK(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)

	// GetByID
	mock.ExpectQuery(`SELECT id`).
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows(wishlistCols).AddRow(wishRow(1, "Bolt")...))

	// INSERT INTO cards
	mock.ExpectExec(`INSERT INTO cards`).
		WillReturnResult(sqlmock.NewResult(99, 1))

	// UPDATE wishlist_cards SET acquired
	mock.ExpectExec(`UPDATE wishlist_cards SET acquired`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	cardID, err := r.Acquire("1", AcquireInput{Condition: "near_mint"})
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	if cardID != 99 {
		t.Errorf("cardID = %d, want 99", cardID)
	}
}

func TestWishlistRepoAcquire_GetByIDError(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id`).WithArgs("99").WillReturnError(sql.ErrNoRows)

	_, err := r.Acquire("99", AcquireInput{})
	if err == nil {
		t.Error("expected error")
	}
}

func TestWishlistRepoAcquire_WithFoilPriceAndNoName(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)

	// A card with no name (uses setCode+number), foil=true, foil price > 0
	row := []driver.Value{
		int64(1), "mtg-1", "GRN", "200", "", "",
		"", "Art", "C", `[]`, "",
		1.0, 4.0,
		int64(1), "", int64(0), "2024-01-01T00:00:00",
	}
	mock.ExpectQuery(`SELECT id`).
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows(wishlistCols).AddRow(row...))

	mock.ExpectExec(`INSERT INTO cards`).
		WillReturnResult(sqlmock.NewResult(55, 1))

	mock.ExpectExec(`UPDATE wishlist_cards SET acquired`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	cardID, err := r.Acquire("1", AcquireInput{Commander: true, PreRelease: true})
	if err != nil {
		t.Fatalf("Acquire no-name foil: %v", err)
	}
	if cardID != 55 {
		t.Errorf("cardID = %d, want 55", cardID)
	}
}

func TestWishlistRepoAcquire_InsertError(t *testing.T) {
	db, mock := newWishMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id`).
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows(wishlistCols).AddRow(wishRow(1, "Bolt")...))

	mock.ExpectExec(`INSERT INTO cards`).WillReturnError(sql.ErrConnDone)

	_, err := r.Acquire("1", AcquireInput{})
	if err == nil {
		t.Error("expected error")
	}
}
