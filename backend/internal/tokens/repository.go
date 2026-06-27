package tokens

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List() ([]Token, error) {
	rows, err := r.db.Query(`
		SELECT id, name, type_line, oracle_text, power, toughness, colors, set_code,
		       collection_number, mtg_id, image_url, double_faced,
		       back_name, back_type_line, back_oracle_text, back_image_url, back_power, back_toughness,
		       artist, quantity, foil,
		       COALESCE(DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%s'), '') AS created_at
		FROM tokens ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Token
	for rows.Next() {
		var t Token
		var doubleFacedInt, foilInt int
		if err := rows.Scan(
			&t.ID, &t.Name, &t.TypeLine, &t.OracleText, &t.Power, &t.Toughness, &t.Colors,
			&t.SetCode, &t.CollectionNumber, &t.MtgID, &t.ImageURL, &doubleFacedInt,
			&t.BackName, &t.BackTypeLine, &t.BackOracleText, &t.BackImageURL, &t.BackPower, &t.BackToughness,
			&t.Artist, &t.Quantity, &foilInt, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		t.DoubleFaced = doubleFacedInt == 1
		t.Foil = foilInt == 1
		result = append(result, t)
	}
	return result, nil
}

func (r *Repository) Create(t Token) (int64, error) {
	doubleFacedInt := 0
	if t.DoubleFaced {
		doubleFacedInt = 1
	}
	foilInt := 0
	if t.Foil {
		foilInt = 1
	}

	res, err := r.db.Exec(`
		INSERT INTO tokens
		  (name, type_line, oracle_text, power, toughness, colors, set_code,
		   collection_number, mtg_id, image_url, double_faced,
		   back_name, back_type_line, back_oracle_text, back_image_url, back_power, back_toughness,
		   artist, quantity, foil)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		t.Name, t.TypeLine, t.OracleText, t.Power, t.Toughness, t.Colors, t.SetCode,
		t.CollectionNumber, t.MtgID, t.ImageURL, doubleFacedInt,
		t.BackName, t.BackTypeLine, t.BackOracleText, t.BackImageURL, t.BackPower, t.BackToughness,
		t.Artist, t.Quantity, foilInt,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateQuantity(id string, quantity int) error {
	_, err := r.db.Exec(`UPDATE tokens SET quantity = ? WHERE id = ?`, quantity, id)
	return err
}

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM tokens WHERE id = ?`, id)
	return err
}
