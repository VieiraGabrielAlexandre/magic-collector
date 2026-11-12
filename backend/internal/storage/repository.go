package storage

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List() ([]StorageLocation, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.name, s.type, s.color, s.description,
		       COALESCE(COUNT(c.id), 0) AS card_count
		FROM storage_locations s
		LEFT JOIN cards c ON c.storage_id = s.id
		GROUP BY s.id, s.name, s.type, s.color, s.description
		ORDER BY s.name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []StorageLocation
	for rows.Next() {
		var s StorageLocation
		if err := rows.Scan(&s.ID, &s.Name, &s.Type, &s.Color, &s.Description, &s.CardCount); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

func (r *Repository) Create(input StorageInput) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO storage_locations (name, type, color, description) VALUES (?, ?, ?, ?)`,
		input.Name, input.Type, input.Color, input.Description,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) Update(id string, input StorageInput) error {
	_, err := r.db.Exec(
		`UPDATE storage_locations SET name=?, type=?, color=?, description=? WHERE id=?`,
		input.Name, input.Type, input.Color, input.Description, id,
	)
	return err
}

func (r *Repository) Delete(id string) error {
	r.db.Exec(`UPDATE cards SET storage_id = 0 WHERE storage_id = ?`, id)
	_, err := r.db.Exec(`DELETE FROM storage_locations WHERE id=?`, id)
	return err
}
