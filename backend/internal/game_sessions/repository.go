package game_sessions

import (
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List() ([]GameSession, error) {
	rows, err := r.db.Query(`
		SELECT id, name, format, status, starting_life,
			DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%s'),
			DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%s'),
			DATE_FORMAT(ended_at, '%Y-%m-%dT%H:%i:%s')
		FROM game_sessions ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]GameSession, 0)
	sessionMap := make(map[int64]int)
	for rows.Next() {
		var s GameSession
		var endedAt sql.NullString
		if err := rows.Scan(&s.ID, &s.Name, &s.Format, &s.Status, &s.StartingLife,
			&s.CreatedAt, &s.UpdatedAt, &endedAt); err != nil {
			return nil, err
		}
		if endedAt.Valid {
			s.EndedAt = &endedAt.String
		}
		s.Players = make([]Player, 0)
		sessionMap[s.ID] = len(sessions)
		sessions = append(sessions, s)
	}

	if len(sessions) == 0 {
		return sessions, nil
	}

	playerRows, err := r.db.Query(`
		SELECT id, session_id, name, short_code, life, poison, commander_damage_received,
			is_eliminated, COALESCE(eliminated_reason, ''),
			DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%s'),
			DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%s')
		FROM game_session_players ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer playerRows.Close()

	for playerRows.Next() {
		var p Player
		if err := playerRows.Scan(&p.ID, &p.SessionID, &p.Name, &p.ShortCode,
			&p.Life, &p.Poison, &p.CommanderDamageReceived, &p.IsEliminated,
			&p.EliminatedReason, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if idx, ok := sessionMap[p.SessionID]; ok {
			sessions[idx].Players = append(sessions[idx].Players, p)
		}
	}

	return sessions, nil
}

func (r *Repository) GetByID(id int64) (*GameSession, error) {
	var s GameSession
	var endedAt sql.NullString
	err := r.db.QueryRow(`
		SELECT id, name, format, status, starting_life,
			DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%s'),
			DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%s'),
			DATE_FORMAT(ended_at, '%Y-%m-%dT%H:%i:%s')
		FROM game_sessions WHERE id = ?
	`, id).Scan(&s.ID, &s.Name, &s.Format, &s.Status, &s.StartingLife,
		&s.CreatedAt, &s.UpdatedAt, &endedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if endedAt.Valid {
		s.EndedAt = &endedAt.String
	}

	rows, err := r.db.Query(`
		SELECT id, session_id, name, short_code, life, poison, commander_damage_received,
			is_eliminated, COALESCE(eliminated_reason, ''),
			DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%s'),
			DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%s')
		FROM game_session_players WHERE session_id = ? ORDER BY id ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	s.Players = make([]Player, 0)
	for rows.Next() {
		var p Player
		if err := rows.Scan(&p.ID, &p.SessionID, &p.Name, &p.ShortCode,
			&p.Life, &p.Poison, &p.CommanderDamageReceived, &p.IsEliminated,
			&p.EliminatedReason, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		s.Players = append(s.Players, p)
	}

	return &s, nil
}

func (r *Repository) Create(input CreateSessionInput) (*GameSession, error) {
	res, err := r.db.Exec(`
		INSERT INTO game_sessions (name, format, status, starting_life)
		VALUES (?, ?, 'active', ?)
	`, input.Name, input.Format, input.StartingLife)
	if err != nil {
		return nil, err
	}
	sessionID, _ := res.LastInsertId()

	for _, pi := range input.Players {
		_, err := r.db.Exec(`
			INSERT INTO game_session_players
				(session_id, name, short_code, life, poison, commander_damage_received, is_eliminated)
			VALUES (?, ?, ?, ?, 0, 0, 0)
		`, sessionID, pi.Name, pi.ShortCode, input.StartingLife)
		if err != nil {
			return nil, err
		}
	}

	return r.GetByID(sessionID)
}

func (r *Repository) Delete(id int64) error {
	if _, err := r.db.Exec("DELETE FROM game_session_players WHERE session_id = ?", id); err != nil {
		return err
	}
	_, err := r.db.Exec("DELETE FROM game_sessions WHERE id = ?", id)
	return err
}

func (r *Repository) getPlayerByID(id int64) (*Player, error) {
	var p Player
	err := r.db.QueryRow(`
		SELECT id, session_id, name, short_code, life, poison, commander_damage_received,
			is_eliminated, COALESCE(eliminated_reason, ''),
			DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%s'),
			DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%s')
		FROM game_session_players WHERE id = ?
	`, id).Scan(&p.ID, &p.SessionID, &p.Name, &p.ShortCode,
		&p.Life, &p.Poison, &p.CommanderDamageReceived, &p.IsEliminated,
		&p.EliminatedReason, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) AddPlayer(sessionID int64, input PlayerInput, startingLife int) (*Player, error) {
	res, err := r.db.Exec(`
		INSERT INTO game_session_players
			(session_id, name, short_code, life, poison, commander_damage_received, is_eliminated)
		VALUES (?, ?, ?, ?, 0, 0, 0)
	`, sessionID, input.Name, input.ShortCode, startingLife)
	if err != nil {
		return nil, err
	}
	playerID, _ := res.LastInsertId()
	return r.getPlayerByID(playerID)
}

func (r *Repository) UpdatePlayer(sessionID, playerID int64, input UpdatePlayerInput) (*Player, error) {
	isEliminated := false
	eliminatedReason := ""

	if input.CommanderDamageReceived >= 21 {
		isEliminated = true
		eliminatedReason = "commander_damage"
	} else if input.Poison >= 10 {
		isEliminated = true
		eliminatedReason = "poison"
	} else if input.Life <= 0 {
		isEliminated = true
		eliminatedReason = "life"
	}

	var eliminatedReasonVal interface{}
	if isEliminated {
		eliminatedReasonVal = eliminatedReason
	}

	_, err := r.db.Exec(`
		UPDATE game_session_players
		SET life = ?, poison = ?, commander_damage_received = ?,
			is_eliminated = ?, eliminated_reason = ?
		WHERE id = ? AND session_id = ?
	`, input.Life, input.Poison, input.CommanderDamageReceived,
		isEliminated, eliminatedReasonVal, playerID, sessionID)
	if err != nil {
		return nil, err
	}

	return r.getPlayerByID(playerID)
}

func (r *Repository) DeletePlayer(sessionID, playerID int64) error {
	_, err := r.db.Exec("DELETE FROM game_session_players WHERE id = ? AND session_id = ?", playerID, sessionID)
	return err
}

func (r *Repository) Reset(sessionID int64) (*GameSession, error) {
	var startingLife int
	err := r.db.QueryRow("SELECT starting_life FROM game_sessions WHERE id = ?", sessionID).Scan(&startingLife)
	if err != nil {
		return nil, fmt.Errorf("sessão não encontrada")
	}

	_, err = r.db.Exec(`
		UPDATE game_session_players
		SET life = ?, poison = 0, commander_damage_received = 0, is_eliminated = 0, eliminated_reason = NULL
		WHERE session_id = ?
	`, startingLife, sessionID)
	if err != nil {
		return nil, err
	}

	return r.GetByID(sessionID)
}

func (r *Repository) Finish(sessionID int64) (*GameSession, error) {
	_, err := r.db.Exec(`UPDATE game_sessions SET status = 'finished', ended_at = NOW() WHERE id = ?`, sessionID)
	if err != nil {
		return nil, err
	}
	return r.GetByID(sessionID)
}

func (r *Repository) Restore(sessionID int64) (*GameSession, error) {
	_, err := r.db.Exec(`UPDATE game_sessions SET status = 'active', ended_at = NULL WHERE id = ?`, sessionID)
	if err != nil {
		return nil, err
	}
	return r.GetByID(sessionID)
}
