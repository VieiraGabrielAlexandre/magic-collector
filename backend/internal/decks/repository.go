package decks

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List() ([]Deck, error) {
	rows, err := r.db.Query(`
		SELECT d.id, d.name, d.description, d.commander, d.colors, d.set_code, d.icon_uri, d.theme_color,
		       COUNT(DISTINCT c.id) AS card_count,
		       COALESCE(d.evaluation, '') AS evaluation,
		       COALESCE(DATE_FORMAT(d.evaluated_at, '%Y-%m-%dT%H:%i:%s'), '') AS evaluated_at,
		       COALESCE(bs.wins,   0) AS battle_wins,
		       COALESCE(bs.losses, 0) AS battle_losses,
		       COALESCE(bs.draws,  0) AS battle_draws,
		       COALESCE(bs.total,  0) AS battle_total
		FROM decks d
		LEFT JOIN cards c ON c.deck_id = d.id
		LEFT JOIN (
		    SELECT deck_id,
		           SUM(CASE WHEN result='win'  THEN 1 ELSE 0 END) AS wins,
		           SUM(CASE WHEN result='loss' THEN 1 ELSE 0 END) AS losses,
		           SUM(CASE WHEN result='draw' THEN 1 ELSE 0 END) AS draws,
		           COUNT(*) AS total
		    FROM battles WHERE deck_is_mine = 1
		    GROUP BY deck_id
		) bs ON bs.deck_id = d.id
		GROUP BY d.id, d.name, d.description, d.commander, d.colors, d.set_code,
		         d.icon_uri, d.theme_color, d.evaluation, d.evaluated_at,
		         bs.wins, bs.losses, bs.draws, bs.total
		ORDER BY d.name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Deck
	for rows.Next() {
		var d Deck
		var commanderInt int
		if err := rows.Scan(
			&d.ID, &d.Name, &d.Description, &commanderInt, &d.Colors, &d.SetCode,
			&d.IconURI, &d.ThemeColor, &d.CardCount, &d.Evaluation, &d.EvaluatedAt,
			&d.BattleWins, &d.BattleLosses, &d.BattleDraws, &d.BattleTotal,
		); err != nil {
			return nil, err
		}
		d.Commander = commanderInt == 1
		result = append(result, d)
	}
	return result, nil
}

func (r *Repository) GetByID(id string) (*Deck, error) {
	row := r.db.QueryRow(
		`SELECT id, name, description, commander, colors, set_code, icon_uri, theme_color FROM decks WHERE id = ?`, id)
	var d Deck
	var commanderInt int
	if err := row.Scan(&d.ID, &d.Name, &d.Description, &commanderInt, &d.Colors, &d.SetCode, &d.IconURI, &d.ThemeColor); err != nil {
		return nil, err
	}
	d.Commander = commanderInt == 1
	return &d, nil
}

func (r *Repository) Create(input DeckInput) (int64, error) {
	commanderInt := 0
	if input.Commander {
		commanderInt = 1
	}
	res, err := r.db.Exec(
		`INSERT INTO decks (name, description, commander, colors, set_code, theme_color) VALUES (?, ?, ?, ?, ?, ?)`,
		input.Name, input.Description, commanderInt, input.Colors, input.SetCode, input.ThemeColor,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) Update(id string, input DeckInput) error {
	commanderInt := 0
	if input.Commander {
		commanderInt = 1
	}
	_, err := r.db.Exec(
		`UPDATE decks SET name=?, description=?, commander=?, colors=?, set_code=?, theme_color=? WHERE id=?`,
		input.Name, input.Description, commanderInt, input.Colors, input.SetCode, input.ThemeColor, id,
	)
	return err
}

func (r *Repository) UpdateIcon(id, iconURI string) error {
	_, err := r.db.Exec(`UPDATE decks SET icon_uri=? WHERE id=?`, iconURI, id)
	return err
}

func (r *Repository) UpdateEvaluation(id, evaluation string) error {
	_, err := r.db.Exec(`UPDATE decks SET evaluation=?, evaluated_at=NOW() WHERE id=?`, evaluation, id)
	return err
}

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM decks WHERE id=?`, id)
	return err
}
