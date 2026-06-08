package decks

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var deckListCols = []string{
	"id", "name", "description", "commander", "colors", "set_code",
	"icon_uri", "theme_color", "card_count", "evaluation", "evaluated_at",
	"battle_wins", "battle_losses", "battle_draws", "battle_total",
}

func newDeckMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// ── List ─────────────────────────────────────────────────────────────────────

func TestDecksRepoList_OK(t *testing.T) {
	db, mock := newDeckMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT d.id`).
		WillReturnRows(sqlmock.NewRows(deckListCols).
			AddRow(1, "My Deck", "Desc", 1, "W,U", "GRN", "", "sapphire", 30, "", "", 2, 1, 0, 3))

	decks, err := r.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(decks) != 1 || !decks[0].Commander {
		t.Errorf("unexpected: %+v", decks)
	}
}

func TestDecksRepoList_Empty(t *testing.T) {
	db, mock := newDeckMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT d.id`).
		WillReturnRows(sqlmock.NewRows(deckListCols))

	decks, err := r.List()
	if err != nil {
		t.Fatalf("List empty: %v", err)
	}
	if len(decks) != 0 {
		t.Errorf("expected 0 decks, got %d", len(decks))
	}
}

func TestDecksRepoList_QueryError(t *testing.T) {
	db, mock := newDeckMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT d.id`).WillReturnError(sql.ErrConnDone)

	_, err := r.List()
	if err == nil {
		t.Error("expected error")
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestDecksRepoGetByID_OK(t *testing.T) {
	db, mock := newDeckMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id, name`).
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "commander", "colors", "set_code", "icon_uri", "theme_color"}).
			AddRow(1, "My Deck", "Desc", 0, "R", "M21", "", "ember"))

	d, err := r.GetByID("1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if d.Name != "My Deck" {
		t.Errorf("name = %q, want My Deck", d.Name)
	}
}

func TestDecksRepoGetByID_NotFound(t *testing.T) {
	db, mock := newDeckMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id, name`).
		WithArgs("99").
		WillReturnError(sql.ErrNoRows)

	_, err := r.GetByID("99")
	if err == nil {
		t.Error("expected error")
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestDecksRepoCreate_OK(t *testing.T) {
	db, mock := newDeckMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`INSERT INTO decks`).
		WillReturnResult(sqlmock.NewResult(10, 1))

	id, err := r.Create(DeckInput{Name: "Test", Commander: true})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 10 {
		t.Errorf("id = %d, want 10", id)
	}
}

func TestDecksRepoCreate_Error(t *testing.T) {
	db, mock := newDeckMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`INSERT INTO decks`).WillReturnError(sql.ErrConnDone)

	_, err := r.Create(DeckInput{Name: "x"})
	if err == nil {
		t.Error("expected error")
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestDecksRepoUpdate_OK(t *testing.T) {
	db, mock := newDeckMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE decks SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := r.Update("1", DeckInput{Name: "Updated"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestDecksRepoUpdate_Commander(t *testing.T) {
	db, mock := newDeckMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE decks SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := r.Update("2", DeckInput{Name: "Commander Deck", Commander: true})
	if err != nil {
		t.Fatalf("Update commander: %v", err)
	}
}

// ── UpdateIcon ────────────────────────────────────────────────────────────────

func TestDecksRepoUpdateIcon_OK(t *testing.T) {
	db, mock := newDeckMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE decks SET icon_uri`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := r.UpdateIcon("1", "https://icon.svg")
	if err != nil {
		t.Fatalf("UpdateIcon: %v", err)
	}
}

// ── UpdateEvaluation ──────────────────────────────────────────────────────────

func TestDecksRepoUpdateEvaluation_OK(t *testing.T) {
	db, mock := newDeckMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE decks SET evaluation`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := r.UpdateEvaluation("1", "Strong deck")
	if err != nil {
		t.Fatalf("UpdateEvaluation: %v", err)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDecksRepoDelete_OK(t *testing.T) {
	db, mock := newDeckMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`DELETE FROM decks`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := r.Delete("1")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestDecksRepoDelete_Error(t *testing.T) {
	db, mock := newDeckMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`DELETE FROM decks`).WillReturnError(sql.ErrConnDone)

	err := r.Delete("1")
	if err == nil {
		t.Error("expected error")
	}
}
