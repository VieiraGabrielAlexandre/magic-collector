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
		SELECT d.id, d.name, d.description, d.commander, d.colors, d.set_code, d.icon_uri, d.theme_color, COUNT(c.id) AS card_count
		FROM decks d
		LEFT JOIN cards c ON c.deck_id = d.id
		GROUP BY d.id
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
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &commanderInt, &d.Colors, &d.SetCode, &d.IconURI, &d.ThemeColor, &d.CardCount); err != nil {
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

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM decks WHERE id=?`, id)
	return err
}
