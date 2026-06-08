package game_sessions

import (
	"errors"
	"testing"
)

// ── mock repository ──────────────────────────────────────────────────────────

type mockGSRepo struct {
	listFn          func() ([]GameSession, error)
	getByIDFn       func(id int64) (*GameSession, error)
	createFn        func(input CreateSessionInput) (*GameSession, error)
	deleteFn        func(id int64) error
	addPlayerFn     func(sessionID int64, input PlayerInput, startingLife int) (*Player, error)
	updatePlayerFn  func(sessionID, playerID int64, input UpdatePlayerInput) (*Player, error)
	deletePlayerFn  func(sessionID, playerID int64) error
	resetFn         func(sessionID int64) (*GameSession, error)
	finishFn        func(sessionID int64) (*GameSession, error)
	restoreFn       func(sessionID int64) (*GameSession, error)
}

func (m *mockGSRepo) List() ([]GameSession, error)              { return m.listFn() }
func (m *mockGSRepo) GetByID(id int64) (*GameSession, error)    { return m.getByIDFn(id) }
func (m *mockGSRepo) Create(i CreateSessionInput) (*GameSession, error) { return m.createFn(i) }
func (m *mockGSRepo) Delete(id int64) error                     { return m.deleteFn(id) }
func (m *mockGSRepo) AddPlayer(sid int64, inp PlayerInput, life int) (*Player, error) {
	return m.addPlayerFn(sid, inp, life)
}
func (m *mockGSRepo) UpdatePlayer(sid, pid int64, inp UpdatePlayerInput) (*Player, error) {
	return m.updatePlayerFn(sid, pid, inp)
}
func (m *mockGSRepo) DeletePlayer(sid, pid int64) error { return m.deletePlayerFn(sid, pid) }
func (m *mockGSRepo) Reset(sid int64) (*GameSession, error)   { return m.resetFn(sid) }
func (m *mockGSRepo) Finish(sid int64) (*GameSession, error)  { return m.finishFn(sid) }
func (m *mockGSRepo) Restore(sid int64) (*GameSession, error) { return m.restoreFn(sid) }

func twoPlayerSession() *GameSession {
	return &GameSession{
		ID:           1,
		Status:       "active",
		StartingLife: 40,
		Players: []Player{
			{ID: 10, Name: "Alice"},
			{ID: 11, Name: "Bob"},
		},
	}
}

// ── Service.Create validation ────────────────────────────────────────────────

func TestServiceCreate_TooFewPlayers(t *testing.T) {
	svc := &Service{repo: &mockGSRepo{}}
	_, err := svc.Create(CreateSessionInput{
		Name:    "Test",
		Players: []PlayerInput{{Name: "Alice", ShortCode: "A"}},
	})
	if err == nil || err.Error() != "mínimo de 2 jogadores" {
		t.Errorf("expected min players error, got %v", err)
	}
}

func TestServiceCreate_TooManyPlayers(t *testing.T) {
	svc := &Service{repo: &mockGSRepo{}}
	players := make([]PlayerInput, 9)
	for i := range players {
		players[i] = PlayerInput{Name: "P", ShortCode: "P"}
	}
	_, err := svc.Create(CreateSessionInput{Name: "Test", Players: players})
	if err == nil || err.Error() != "máximo de 8 jogadores" {
		t.Errorf("expected max players error, got %v", err)
	}
}

func TestServiceCreate_EmptyPlayerName(t *testing.T) {
	svc := &Service{repo: &mockGSRepo{}}
	_, err := svc.Create(CreateSessionInput{
		Name: "Test",
		Players: []PlayerInput{
			{Name: "", ShortCode: "A"},
			{Name: "Bob", ShortCode: "B"},
		},
	})
	if err == nil {
		t.Fatal("expected error for empty player name")
	}
}

func TestServiceCreate_ShortCodeTooLong(t *testing.T) {
	svc := &Service{repo: &mockGSRepo{}}
	_, err := svc.Create(CreateSessionInput{
		Name: "Test",
		Players: []PlayerInput{
			{Name: "Alice", ShortCode: "ABCD"}, // 4 chars > 3
			{Name: "Bob", ShortCode: "B"},
		},
	})
	if err == nil {
		t.Fatal("expected error for long short code")
	}
}

func TestServiceCreate_EmptyShortCode(t *testing.T) {
	svc := &Service{repo: &mockGSRepo{}}
	_, err := svc.Create(CreateSessionInput{
		Name: "Test",
		Players: []PlayerInput{
			{Name: "Alice", ShortCode: ""},
			{Name: "Bob", ShortCode: "B"},
		},
	})
	if err == nil {
		t.Fatal("expected error for empty short code")
	}
}

func TestServiceCreate_DefaultsFormat(t *testing.T) {
	var capturedInput CreateSessionInput
	repo := &mockGSRepo{
		createFn: func(i CreateSessionInput) (*GameSession, error) {
			capturedInput = i
			return &GameSession{ID: 1}, nil
		},
	}
	svc := &Service{repo: repo}
	_, err := svc.Create(CreateSessionInput{
		Name: "Test",
		Players: []PlayerInput{
			{Name: "Alice", ShortCode: "A"},
			{Name: "Bob", ShortCode: "B"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if capturedInput.Format != "Commander" {
		t.Errorf("Format = %q, want Commander", capturedInput.Format)
	}
	if capturedInput.StartingLife != 40 {
		t.Errorf("StartingLife = %d, want 40", capturedInput.StartingLife)
	}
}

func TestServiceCreate_CasualFormat_DefaultsTo20Life(t *testing.T) {
	var capturedInput CreateSessionInput
	repo := &mockGSRepo{
		createFn: func(i CreateSessionInput) (*GameSession, error) {
			capturedInput = i
			return &GameSession{ID: 1}, nil
		},
	}
	svc := &Service{repo: repo}
	_, err := svc.Create(CreateSessionInput{
		Name:   "Casual",
		Format: "casual",
		Players: []PlayerInput{
			{Name: "Alice", ShortCode: "A"},
			{Name: "Bob", ShortCode: "B"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if capturedInput.StartingLife != 20 {
		t.Errorf("StartingLife = %d, want 20 for casual", capturedInput.StartingLife)
	}
}

func TestServiceCreate_OK(t *testing.T) {
	repo := &mockGSRepo{
		createFn: func(i CreateSessionInput) (*GameSession, error) {
			return &GameSession{ID: 42, Name: i.Name}, nil
		},
	}
	svc := &Service{repo: repo}
	sess, err := svc.Create(CreateSessionInput{
		Name: "Friday Night",
		Players: []PlayerInput{
			{Name: "Alice", ShortCode: "A"},
			{Name: "Bob", ShortCode: "B"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if sess.ID != 42 {
		t.Errorf("ID = %d, want 42", sess.ID)
	}
}

// ── Service.AddPlayer ────────────────────────────────────────────────────────

func TestServiceAddPlayer_SessionNotFound(t *testing.T) {
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return nil, errors.New("not found") },
	}
	svc := &Service{repo: repo}
	_, err := svc.AddPlayer(1, PlayerInput{Name: "Eve", ShortCode: "E"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServiceAddPlayer_FinishedSession(t *testing.T) {
	session := &GameSession{ID: 1, Status: "finished", Players: []Player{{}, {}}}
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return session, nil },
	}
	svc := &Service{repo: repo}
	_, err := svc.AddPlayer(1, PlayerInput{Name: "Eve", ShortCode: "E"})
	if err == nil || err.Error() != "sessão encerrada não pode ser alterada" {
		t.Errorf("expected finished error, got %v", err)
	}
}

func TestServiceAddPlayer_MaxPlayers(t *testing.T) {
	players := make([]Player, 8)
	session := &GameSession{ID: 1, Status: "active", Players: players, StartingLife: 40}
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return session, nil },
	}
	svc := &Service{repo: repo}
	_, err := svc.AddPlayer(1, PlayerInput{Name: "Nine", ShortCode: "9"})
	if err == nil || err.Error() != "máximo de 8 jogadores" {
		t.Errorf("expected max players error, got %v", err)
	}
}

func TestServiceAddPlayer_LongShortCode(t *testing.T) {
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return twoPlayerSession(), nil },
	}
	svc := &Service{repo: repo}
	_, err := svc.AddPlayer(1, PlayerInput{Name: "Eve", ShortCode: "LONG"})
	if err == nil {
		t.Fatal("expected short code error")
	}
}

func TestServiceAddPlayer_OK(t *testing.T) {
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return twoPlayerSession(), nil },
		addPlayerFn: func(sid int64, inp PlayerInput, life int) (*Player, error) {
			return &Player{ID: 99, Name: inp.Name}, nil
		},
	}
	svc := &Service{repo: repo}
	p, err := svc.AddPlayer(1, PlayerInput{Name: "Eve", ShortCode: "E"})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Eve" {
		t.Errorf("Name = %q, want Eve", p.Name)
	}
}

// ── Service.DeletePlayer ─────────────────────────────────────────────────────

func TestServiceDeletePlayer_MinPlayers(t *testing.T) {
	session := twoPlayerSession() // has exactly 2 players (IDs 10,11)
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return session, nil },
	}
	svc := &Service{repo: repo}
	// Deleting player 10 leaves 1 player → should fail
	err := svc.DeletePlayer(1, 10)
	if err == nil || err.Error() != "mínimo de 2 jogadores por sessão" {
		t.Errorf("expected min players error, got %v", err)
	}
}

func TestServiceDeletePlayer_OK(t *testing.T) {
	session := &GameSession{
		ID:     1,
		Status: "active",
		Players: []Player{
			{ID: 10, Name: "Alice"},
			{ID: 11, Name: "Bob"},
			{ID: 12, Name: "Carol"},
		},
	}
	repo := &mockGSRepo{
		getByIDFn:      func(id int64) (*GameSession, error) { return session, nil },
		deletePlayerFn: func(sid, pid int64) error { return nil },
	}
	svc := &Service{repo: repo}
	err := svc.DeletePlayer(1, 10)
	if err != nil {
		t.Fatal(err)
	}
}

// ── Service.UpdatePlayer ─────────────────────────────────────────────────────

func TestServiceUpdatePlayer_FinishedSession(t *testing.T) {
	session := &GameSession{ID: 1, Status: "finished"}
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return session, nil },
	}
	svc := &Service{repo: repo}
	_, err := svc.UpdatePlayer(1, 10, UpdatePlayerInput{Life: 35})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServiceUpdatePlayer_OK(t *testing.T) {
	session := twoPlayerSession()
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return session, nil },
		updatePlayerFn: func(sid, pid int64, inp UpdatePlayerInput) (*Player, error) {
			return &Player{ID: pid, Life: inp.Life}, nil
		},
	}
	svc := &Service{repo: repo}
	p, err := svc.UpdatePlayer(1, 10, UpdatePlayerInput{Life: 35})
	if err != nil {
		t.Fatal(err)
	}
	if p.Life != 35 {
		t.Errorf("Life = %d, want 35", p.Life)
	}
}

// ── Service.Reset / Finish / Restore ────────────────────────────────────────

func TestServiceReset_FinishedSession(t *testing.T) {
	session := &GameSession{ID: 1, Status: "finished"}
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return session, nil },
	}
	svc := &Service{repo: repo}
	_, err := svc.Reset(1)
	if err == nil {
		t.Fatal("expected error for finished session reset")
	}
}

func TestServiceReset_OK(t *testing.T) {
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return twoPlayerSession(), nil },
		resetFn:   func(sid int64) (*GameSession, error) { return twoPlayerSession(), nil },
	}
	svc := &Service{repo: repo}
	sess, err := svc.Reset(1)
	if err != nil || sess == nil {
		t.Fatal("expected success")
	}
}

func TestServiceFinish_AlreadyFinished(t *testing.T) {
	session := &GameSession{ID: 1, Status: "finished"}
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return session, nil },
	}
	svc := &Service{repo: repo}
	_, err := svc.Finish(1)
	if err == nil || err.Error() != "sessão já encerrada" {
		t.Errorf("expected already finished error, got %v", err)
	}
}

func TestServiceFinish_OK(t *testing.T) {
	finished := &GameSession{ID: 1, Status: "finished"}
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return twoPlayerSession(), nil },
		finishFn:  func(sid int64) (*GameSession, error) { return finished, nil },
	}
	svc := &Service{repo: repo}
	sess, err := svc.Finish(1)
	if err != nil || sess.Status != "finished" {
		t.Fatalf("expected finished session, err=%v", err)
	}
}

func TestServiceRestore_SessionNotFound(t *testing.T) {
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return nil, nil },
	}
	svc := &Service{repo: repo}
	_, err := svc.Restore(99)
	if err == nil {
		t.Fatal("expected error for nil session")
	}
}

func TestServiceRestore_OK(t *testing.T) {
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return twoPlayerSession(), nil },
		restoreFn: func(sid int64) (*GameSession, error) { return twoPlayerSession(), nil },
	}
	svc := &Service{repo: repo}
	sess, err := svc.Restore(1)
	if err != nil || sess == nil {
		t.Fatal("expected success")
	}
}

// ── Service.List / GetByID / Delete ─────────────────────────────────────────

func TestServiceList(t *testing.T) {
	sessions := []GameSession{{ID: 1}, {ID: 2}}
	repo := &mockGSRepo{
		listFn: func() ([]GameSession, error) { return sessions, nil },
	}
	svc := &Service{repo: repo}
	got, err := svc.List()
	if err != nil || len(got) != 2 {
		t.Fatal("unexpected result")
	}
}

func TestServiceGetByID_OK(t *testing.T) {
	repo := &mockGSRepo{
		getByIDFn: func(id int64) (*GameSession, error) { return twoPlayerSession(), nil },
	}
	svc := &Service{repo: repo}
	sess, err := svc.GetByID(1)
	if err != nil || sess == nil {
		t.Fatal("expected session")
	}
}

func TestServiceDelete_OK(t *testing.T) {
	repo := &mockGSRepo{
		deleteFn: func(id int64) error { return nil },
	}
	svc := &Service{repo: repo}
	if err := svc.Delete(1); err != nil {
		t.Fatal(err)
	}
}
