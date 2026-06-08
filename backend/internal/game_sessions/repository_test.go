package game_sessions

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var sessionCols = []string{
	"id", "name", "format", "status", "starting_life",
	"created_at", "updated_at", "ended_at",
}

var playerCols = []string{
	"id", "session_id", "name", "short_code", "life", "poison", "commander_damage_received",
	"is_eliminated", "eliminated_reason", "created_at", "updated_at",
}

func newGSMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func sessionRow(id int64, name, status string) []driver.Value {
	return []driver.Value{id, name, "Commander", status, int64(40), "2024-01-01T00:00:00", "2024-01-01T00:00:00", nil}
}

func playerRow(id, sessionID int64, name string) []driver.Value {
	return []driver.Value{id, sessionID, name, "P1", int64(40), int64(0), int64(0), false, "", "2024-01-01T00:00:00", "2024-01-01T00:00:00"}
}

// ── List ─────────────────────────────────────────────────────────────────────

func TestGSRepoList_Empty(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id, name`).
		WillReturnRows(sqlmock.NewRows(sessionCols))

	sessions, err := r.List()
	if err != nil {
		t.Fatalf("List empty: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0, got %d", len(sessions))
	}
}

func TestGSRepoList_WithPlayers(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id, name`).
		WillReturnRows(sqlmock.NewRows(sessionCols).
			AddRow(sessionRow(1, "Game 1", "active")...).
			AddRow(sessionRow(2, "Game 2", "finished")...))

	mock.ExpectQuery(`SELECT id, session_id`).
		WillReturnRows(sqlmock.NewRows(playerCols).
			AddRow(playerRow(1, 1, "Alice")...).
			AddRow(playerRow(2, 1, "Bob")...).
			AddRow(playerRow(3, 2, "Carol")...))

	sessions, err := r.List()
	if err != nil {
		t.Fatalf("List with players: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("sessions = %d, want 2", len(sessions))
	}
	if len(sessions[0].Players) != 2 {
		t.Errorf("players[0] = %d, want 2", len(sessions[0].Players))
	}
}

func TestGSRepoList_SessionQueryError(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)
	mock.ExpectQuery(`SELECT id, name`).WillReturnError(sql.ErrConnDone)
	_, err := r.List()
	if err == nil {
		t.Error("expected error")
	}
}

func TestGSRepoList_PlayerQueryError(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id, name`).
		WillReturnRows(sqlmock.NewRows(sessionCols).
			AddRow(sessionRow(1, "Game 1", "active")...))

	mock.ExpectQuery(`SELECT id, session_id`).WillReturnError(sql.ErrConnDone)

	_, err := r.List()
	if err == nil {
		t.Error("expected error from player query")
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestGSRepoGetByID_OK(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id, name`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(sessionCols).AddRow(sessionRow(1, "My Game", "active")...))

	mock.ExpectQuery(`SELECT id, session_id`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(playerCols).
			AddRow(playerRow(1, 1, "Alice")...))

	s, err := r.GetByID(1)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if s == nil || s.Name != "My Game" {
		t.Errorf("unexpected: %+v", s)
	}
}

func TestGSRepoGetByID_NotFound(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id, name`).
		WithArgs(int64(99)).
		WillReturnError(sql.ErrNoRows)

	s, err := r.GetByID(99)
	if err != nil || s != nil {
		t.Errorf("expected nil, nil; got %v, %v", s, err)
	}
}

func TestGSRepoGetByID_DBError(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT id, name`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	_, err := r.GetByID(1)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGSRepoGetByID_WithEndedAt(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	endedAt := "2024-02-01T10:00:00"
	row := []driver.Value{int64(1), "Finished Game", "Commander", "finished", int64(40),
		"2024-01-01T00:00:00", "2024-02-01T10:00:00", endedAt}

	mock.ExpectQuery(`SELECT id, name`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(sessionCols).AddRow(row...))

	mock.ExpectQuery(`SELECT id, session_id`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(playerCols))

	s, err := r.GetByID(1)
	if err != nil {
		t.Fatalf("GetByID finished: %v", err)
	}
	if s.EndedAt == nil || *s.EndedAt != endedAt {
		t.Errorf("ended_at = %v, want %s", s.EndedAt, endedAt)
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestGSRepoCreate_OK(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	// INSERT session
	mock.ExpectExec(`INSERT INTO game_sessions`).
		WillReturnResult(sqlmock.NewResult(5, 1))

	// INSERT player 1
	mock.ExpectExec(`INSERT INTO game_session_players`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// INSERT player 2
	mock.ExpectExec(`INSERT INTO game_session_players`).
		WillReturnResult(sqlmock.NewResult(2, 1))

	// GetByID → SELECT session
	mock.ExpectQuery(`SELECT id, name`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows(sessionCols).AddRow(sessionRow(5, "New Game", "active")...))

	// GetByID → SELECT players
	mock.ExpectQuery(`SELECT id, session_id`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows(playerCols).
			AddRow(playerRow(1, 5, "Alice")...).
			AddRow(playerRow(2, 5, "Bob")...))

	input := CreateSessionInput{
		Name:         "New Game",
		Format:       "Commander",
		StartingLife: 40,
		Players:      []PlayerInput{{Name: "Alice", ShortCode: "A"}, {Name: "Bob", ShortCode: "B"}},
	}

	s, err := r.Create(input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if s.Name != "New Game" {
		t.Errorf("name = %q, want New Game", s.Name)
	}
}

func TestGSRepoCreate_InsertError(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`INSERT INTO game_sessions`).WillReturnError(sql.ErrConnDone)

	_, err := r.Create(CreateSessionInput{Name: "x", Players: []PlayerInput{{Name: "A"}}})
	if err == nil {
		t.Error("expected error")
	}
}

func TestGSRepoCreate_PlayerInsertError(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`INSERT INTO game_sessions`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`INSERT INTO game_session_players`).WillReturnError(sql.ErrConnDone)

	_, err := r.Create(CreateSessionInput{
		Name: "x", StartingLife: 40,
		Players: []PlayerInput{{Name: "A"}},
	})
	if err == nil {
		t.Error("expected error on player insert")
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestGSRepoDelete_OK(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`DELETE FROM game_session_players`).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(`DELETE FROM game_sessions`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := r.Delete(1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestGSRepoDelete_PlayerDeleteError(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`DELETE FROM game_session_players`).WillReturnError(sql.ErrConnDone)

	if err := r.Delete(1); err == nil {
		t.Error("expected error")
	}
}

// ── AddPlayer ─────────────────────────────────────────────────────────────────

func TestGSRepoAddPlayer_OK(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`INSERT INTO game_session_players`).
		WillReturnResult(sqlmock.NewResult(10, 1))

	// getPlayerByID
	mock.ExpectQuery(`SELECT id, session_id`).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows(playerCols).AddRow(playerRow(10, 1, "Dave")...))

	p, err := r.AddPlayer(1, PlayerInput{Name: "Dave", ShortCode: "D"}, 40)
	if err != nil {
		t.Fatalf("AddPlayer: %v", err)
	}
	if p.Name != "Dave" {
		t.Errorf("name = %q, want Dave", p.Name)
	}
}

func TestGSRepoAddPlayer_InsertError(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`INSERT INTO game_session_players`).WillReturnError(sql.ErrConnDone)

	_, err := r.AddPlayer(1, PlayerInput{Name: "x"}, 40)
	if err == nil {
		t.Error("expected error")
	}
}

// ── UpdatePlayer ──────────────────────────────────────────────────────────────

func TestGSRepoUpdatePlayer_OK(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE game_session_players`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(`SELECT id, session_id`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(playerCols).AddRow(playerRow(1, 1, "Alice")...))

	p, err := r.UpdatePlayer(1, 1, UpdatePlayerInput{Life: 35, Poison: 0, CommanderDamageReceived: 0})
	if err != nil {
		t.Fatalf("UpdatePlayer: %v", err)
	}
	_ = p
}

func TestGSRepoUpdatePlayer_Eliminated_Life(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE game_session_players`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id, session_id`).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows(playerCols).AddRow(playerRow(2, 1, "Bob")...))

	_, err := r.UpdatePlayer(1, 2, UpdatePlayerInput{Life: 0, Poison: 0, CommanderDamageReceived: 0})
	if err != nil {
		t.Fatalf("UpdatePlayer life=0: %v", err)
	}
}

func TestGSRepoUpdatePlayer_Eliminated_Poison(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE game_session_players`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id, session_id`).
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows(playerCols).AddRow(playerRow(3, 1, "Carol")...))

	_, err := r.UpdatePlayer(1, 3, UpdatePlayerInput{Life: 10, Poison: 10, CommanderDamageReceived: 0})
	if err != nil {
		t.Fatalf("UpdatePlayer poison=10: %v", err)
	}
}

func TestGSRepoUpdatePlayer_Eliminated_CommanderDamage(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE game_session_players`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id, session_id`).
		WithArgs(int64(4)).
		WillReturnRows(sqlmock.NewRows(playerCols).AddRow(playerRow(4, 1, "Dave")...))

	_, err := r.UpdatePlayer(1, 4, UpdatePlayerInput{Life: 20, Poison: 0, CommanderDamageReceived: 21})
	if err != nil {
		t.Fatalf("UpdatePlayer cmd_damage=21: %v", err)
	}
}

func TestGSRepoUpdatePlayer_ExecError(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE game_session_players`).WillReturnError(sql.ErrConnDone)

	_, err := r.UpdatePlayer(1, 1, UpdatePlayerInput{Life: 30})
	if err == nil {
		t.Error("expected error")
	}
}

// ── DeletePlayer ──────────────────────────────────────────────────────────────

func TestGSRepoDeletePlayer_OK(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)
	mock.ExpectExec(`DELETE FROM game_session_players`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.DeletePlayer(1, 2); err != nil {
		t.Fatalf("DeletePlayer: %v", err)
	}
}

// ── Reset ─────────────────────────────────────────────────────────────────────

func TestGSRepoReset_OK(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	// get starting_life
	mock.ExpectQuery(`SELECT starting_life`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"starting_life"}).AddRow(40))

	// update players
	mock.ExpectExec(`UPDATE game_session_players`).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// GetByID
	mock.ExpectQuery(`SELECT id, name`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(sessionCols).AddRow(sessionRow(1, "Game", "active")...))
	mock.ExpectQuery(`SELECT id, session_id`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(playerCols))

	s, err := r.Reset(1)
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if s.Name != "Game" {
		t.Errorf("name = %q", s.Name)
	}
}

func TestGSRepoReset_NotFound(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectQuery(`SELECT starting_life`).
		WithArgs(int64(99)).
		WillReturnError(sql.ErrNoRows)

	_, err := r.Reset(99)
	if err == nil {
		t.Error("expected error")
	}
}

// ── Finish ────────────────────────────────────────────────────────────────────

func TestGSRepoFinish_OK(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE game_sessions SET status = 'finished'`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// GetByID
	mock.ExpectQuery(`SELECT id, name`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(sessionCols).AddRow(sessionRow(1, "Game", "finished")...))
	mock.ExpectQuery(`SELECT id, session_id`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(playerCols))

	s, err := r.Finish(1)
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if s.Status != "finished" {
		t.Errorf("status = %q", s.Status)
	}
}

func TestGSRepoFinish_Error(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE game_sessions SET status = 'finished'`).WillReturnError(sql.ErrConnDone)

	_, err := r.Finish(1)
	if err == nil {
		t.Error("expected error")
	}
}

// ── Restore ───────────────────────────────────────────────────────────────────

func TestGSRepoRestore_OK(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE game_sessions SET status = 'active'`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(`SELECT id, name`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(sessionCols).AddRow(sessionRow(1, "Game", "active")...))
	mock.ExpectQuery(`SELECT id, session_id`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(playerCols))

	s, err := r.Restore(1)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if s.Status != "active" {
		t.Errorf("status = %q", s.Status)
	}
}

func TestGSRepoRestore_Error(t *testing.T) {
	db, mock := newGSMock(t)
	r := NewRepository(db)

	mock.ExpectExec(`UPDATE game_sessions SET status = 'active'`).WillReturnError(sql.ErrConnDone)

	_, err := r.Restore(1)
	if err == nil {
		t.Error("expected error")
	}
}
