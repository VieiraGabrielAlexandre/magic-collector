package battles

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRepositoryList_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "result", "opponents", "player_count", "game_style",
		"deck_id", "deck_name", "deck_is_mine", "notes", "played_at",
	}).
		AddRow(1, "win", `["Alice","Bob"]`, 3, "Commander", 0, "My Deck", 1, "good game", "2024-01-15T20:00:00").
		AddRow(2, "loss", `["Carol"]`, 2, "Standard", 5, "Deck B", 0, "", "2024-01-14T19:00:00")

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	repo := NewRepository(db)
	battles, err := repo.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(battles) != 2 {
		t.Errorf("expected 2 battles, got %d", len(battles))
	}
	if battles[0].Result != "win" {
		t.Errorf("Result = %q, want win", battles[0].Result)
	}
	if !battles[0].DeckIsMine {
		t.Error("expected DeckIsMine=true for first battle")
	}
	if len(battles[0].Opponents) != 2 {
		t.Errorf("expected 2 opponents, got %d", len(battles[0].Opponents))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRepositoryList_QueryError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnError(sqlmock.ErrCancelled)

	repo := NewRepository(db)
	_, err := repo.List()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRepositoryList_InvalidOpponentsJSON(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "result", "opponents", "player_count", "game_style",
		"deck_id", "deck_name", "deck_is_mine", "notes", "played_at",
	}).AddRow(1, "win", `invalid_json`, 2, "Casual", 0, "", 0, "", "2024-01-01T00:00:00")

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	repo := NewRepository(db)
	battles, err := repo.List()
	if err != nil {
		t.Fatal(err)
	}
	// Invalid JSON falls back to empty slice
	if len(battles[0].Opponents) != 0 {
		t.Errorf("expected 0 opponents for invalid JSON, got %d", len(battles[0].Opponents))
	}
}

func TestRepositoryCreate_OK(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectExec("INSERT INTO battles").
		WithArgs("win", `["Alice","Bob"]`, 3, "Commander", 0, "Deck A", 1, "").
		WillReturnResult(sqlmock.NewResult(7, 1))

	repo := NewRepository(db)
	id, err := repo.Create(BattleInput{
		Result:      "win",
		Opponents:   []string{"Alice", "Bob"},
		PlayerCount: 3,
		GameStyle:   "Commander",
		DeckIsMine:  true,
		DeckName:    "Deck A",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 7 {
		t.Errorf("id = %d, want 7", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRepositoryCreate_DefaultsPlayerCount(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	// PlayerCount=0 should default to 2
	mock.ExpectExec("INSERT INTO battles").
		WithArgs("loss", `[]`, 2, "", 0, "", 0, "").
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewRepository(db)
	_, err := repo.Create(BattleInput{Result: "loss", PlayerCount: 0})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryCreate_ExecError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectExec("INSERT INTO battles").WillReturnError(sqlmock.ErrCancelled)

	repo := NewRepository(db)
	_, err := repo.Create(BattleInput{Result: "win"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRepositoryDelete_OK(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectExec("DELETE FROM battles").
		WithArgs("5").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewRepository(db)
	if err := repo.Delete("5"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRepositoryDelete_Error(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectExec("DELETE FROM battles").WillReturnError(sqlmock.ErrCancelled)

	repo := NewRepository(db)
	if err := repo.Delete("99"); err == nil {
		t.Fatal("expected error")
	}
}
