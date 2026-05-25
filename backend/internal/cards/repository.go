package cards

import (
	"database/sql"
	"fmt"
	"strings"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type ListParams struct {
	Search   string
	Page     int
	PageSize int
	Sort     string
	Order    string
}

type ListResult struct {
	Cards      []Card
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

var allowedSortFields = map[string]string{
	"name":              "name",
	"set_code":          "set_code",
	"rarity":            "rarity",
	"color":             "color",
	"year":              "year",
	"collection_number": "collection_number",
}

func (r *Repository) List(params ListParams) (ListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	sortCol := "name"
	if col, ok := allowedSortFields[params.Sort]; ok {
		sortCol = col
	}
	order := "ASC"
	if strings.ToUpper(params.Order) == "DESC" {
		order = "DESC"
	}

	where := ""
	args := []any{}
	if params.Search != "" {
		where = `WHERE name LIKE ? OR set_code LIKE ? OR color LIKE ? OR type LIKE ? OR artist LIKE ?`
		like := "%" + params.Search + "%"
		args = append(args, like, like, like, like, like)
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM cards %s`, where)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return ListResult{}, err
	}

	offset := (params.Page - 1) * params.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, mtg_id, name, color, type, subtitle, collection_number,
		       rarity, set_code, mana_cost, colors, language, year,
		       artist, company, foil, quantity, condition, notes
		FROM cards %s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, where, sortCol, order)

	rows, err := r.db.Query(dataQuery, append(args, params.PageSize, offset)...)
	if err != nil {
		return ListResult{}, err
	}
	defer rows.Close()

	var result []Card
	for rows.Next() {
		var c Card
		var foilInt int
		err := rows.Scan(
			&c.ID, &c.MTGID, &c.Name, &c.Color, &c.Type, &c.Subtitle,
			&c.CollectionNumber, &c.Rarity, &c.SetCode, &c.ManaCost,
			&c.Colors, &c.Language, &c.Year, &c.Artist, &c.Company,
			&foilInt, &c.Quantity, &c.Condition, &c.Notes,
		)
		if err != nil {
			return ListResult{}, err
		}
		c.Foil = foilInt == 1
		result = append(result, c)
	}

	totalPages := (total + params.PageSize - 1) / params.PageSize
	if totalPages < 1 {
		totalPages = 1
	}

	return ListResult{
		Cards:      result,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *Repository) Create(card Card) (int64, error) {
	stmt, err := r.db.Prepare(`
		INSERT INTO cards (
			mtg_id, name, color, type, subtitle, collection_number,
			rarity, set_code, mana_cost, colors, language, year,
			artist, company, foil, quantity, condition, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	foilInt := 0
	if card.Foil {
		foilInt = 1
	}

	result, err := stmt.Exec(
		card.MTGID, card.Name, card.Color, card.Type, card.Subtitle,
		card.CollectionNumber, card.Rarity, card.SetCode, card.ManaCost,
		card.Colors, card.Language, card.Year, card.Artist, card.Company,
		foilInt, card.Quantity, card.Condition, card.Notes,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) GetByID(id string) (*Card, error) {
	row := r.db.QueryRow(`
		SELECT id, mtg_id, name, color, type, subtitle, collection_number,
		       rarity, set_code, mana_cost, colors, language, year,
		       artist, company, foil, quantity, condition, notes
		FROM cards WHERE id = ?
	`, id)

	var c Card
	var foilInt int
	err := row.Scan(
		&c.ID, &c.MTGID, &c.Name, &c.Color, &c.Type, &c.Subtitle,
		&c.CollectionNumber, &c.Rarity, &c.SetCode, &c.ManaCost,
		&c.Colors, &c.Language, &c.Year, &c.Artist, &c.Company,
		&foilInt, &c.Quantity, &c.Condition, &c.Notes,
	)
	if err != nil {
		return nil, err
	}
	c.Foil = foilInt == 1
	return &c, nil
}

func (r *Repository) UpdateMTGID(id, mtgID string) error {
	_, err := r.db.Exec(`UPDATE cards SET mtg_id = ? WHERE id = ?`, mtgID, id)
	return err
}

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM cards WHERE id = ?`, id)
	return err
}
