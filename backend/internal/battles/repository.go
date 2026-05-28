package battles

import (
	"database/sql"
	"encoding/json"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List() ([]Battle, error) {
	rows, err := r.db.Query(`
		SELECT id, result, COALESCE(opponents, '[]'), player_count, game_style,
		       deck_id, deck_name, deck_is_mine, notes,
		       DATE_FORMAT(played_at, '%Y-%m-%dT%H:%i:%s')
		FROM battles
		ORDER BY played_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Battle
	for rows.Next() {
		var b Battle
		var mineInt int
		var opponentsJSON string
		err := rows.Scan(
			&b.ID, &b.Result, &opponentsJSON, &b.PlayerCount, &b.GameStyle,
			&b.DeckID, &b.DeckName, &mineInt, &b.Notes, &b.PlayedAt,
		)
		if err != nil {
			return nil, err
		}
		b.DeckIsMine = mineInt == 1
		if err := json.Unmarshal([]byte(opponentsJSON), &b.Opponents); err != nil {
			b.Opponents = []string{}
		}
		result = append(result, b)
	}
	return result, nil
}

func (r *Repository) Create(b BattleInput) (int64, error) {
	mineInt := 0
	if b.DeckIsMine {
		mineInt = 1
	}
	if b.PlayerCount <= 0 {
		b.PlayerCount = 2
	}
	if b.Opponents == nil {
		b.Opponents = []string{}
	}
	opponentsJSON, _ := json.Marshal(b.Opponents)

	res, err := r.db.Exec(
		`INSERT INTO battles (result, opponents, player_count, game_style, deck_id, deck_name, deck_is_mine, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		b.Result, string(opponentsJSON), b.PlayerCount, b.GameStyle,
		b.DeckID, b.DeckName, mineInt, b.Notes,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM battles WHERE id = ?`, id)
	return err
}
