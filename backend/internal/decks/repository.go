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
		SELECT d.id, d.name, d.description, d.commander, d.colors, COUNT(c.id) AS card_count
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
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &commanderInt, &d.Colors, &d.CardCount); err != nil {
			return nil, err
		}
		d.Commander = commanderInt == 1
		result = append(result, d)
	}
	return result, nil
}

func (r *Repository) Create(input DeckInput) (int64, error) {
	commanderInt := 0
	if input.Commander {
		commanderInt = 1
	}
	res, err := r.db.Exec(
		`INSERT INTO decks (name, description, commander, colors) VALUES (?, ?, ?, ?)`,
		input.Name, input.Description, commanderInt, input.Colors,
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
		`UPDATE decks SET name=?, description=?, commander=?, colors=? WHERE id=?`,
		input.Name, input.Description, commanderInt, input.Colors, id,
	)
	return err
}

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM decks WHERE id=?`, id)
	return err
}
