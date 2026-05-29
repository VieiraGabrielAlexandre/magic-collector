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
	Search       string
	Page         int
	PageSize     int
	Sort         string
	Order        string
	DeckIDFilter *int // nil = all; 0 = without deck; >0 = specific deck
}

type ListResult struct {
	Cards         []Card
	Total         int
	TotalQuantity int
	Page          int
	PageSize      int
	TotalPages    int
}

var allowedSortFields = map[string]string{
	"name":              "name",
	"set_code":          "set_code",
	"rarity":            "rarity",
	"color":             "color",
	"year":              "year",
	"collection_number": "collection_number",
}

// selectCols lista as colunas na mesma ordem que os Scan abaixo.
// `condition` e `type` são palavras reservadas no MySQL e precisam de backticks.
const selectCols = `id, mtg_id, name, color, ` + "`type`" + `, subtitle, collection_number,
	       rarity, set_code, mana_cost, colors, language, year,
	       artist, company, foil, quantity, ` + "`condition`" + `, notes, prerelease, commander, precon_deck, deck_id`

func (r *Repository) List(params ListParams) (ListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 500 {
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
		where = "WHERE name LIKE ? OR set_code LIKE ? OR color LIKE ? OR `type` LIKE ? OR artist LIKE ?"
		like := "%" + params.Search + "%"
		args = append(args, like, like, like, like, like)
	}
	if params.DeckIDFilter != nil {
		clause := "deck_id = ?"
		if where == "" {
			where = "WHERE " + clause
		} else {
			where += " AND " + clause
		}
		args = append(args, *params.DeckIDFilter)
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*), COALESCE(SUM(quantity), 0) FROM cards %s`, where)
	var total, totalQuantity int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total, &totalQuantity); err != nil {
		return ListResult{}, err
	}

	offset := (params.Page - 1) * params.PageSize
	dataQuery := fmt.Sprintf(`SELECT %s FROM cards %s ORDER BY %s %s LIMIT ? OFFSET ?`,
		selectCols, where, sortCol, order)

	rows, err := r.db.Query(dataQuery, append(args, params.PageSize, offset)...)
	if err != nil {
		return ListResult{}, err
	}
	defer rows.Close()

	var result []Card
	for rows.Next() {
		var c Card
		var foilInt, prereleaseInt, commanderInt int
		err := rows.Scan(
			&c.ID, &c.MTGID, &c.Name, &c.Color, &c.Type, &c.Subtitle,
			&c.CollectionNumber, &c.Rarity, &c.SetCode, &c.ManaCost,
			&c.Colors, &c.Language, &c.Year, &c.Artist, &c.Company,
			&foilInt, &c.Quantity, &c.Condition, &c.Notes, &prereleaseInt, &commanderInt, &c.PreconDeck, &c.DeckID,
		)
		if err != nil {
			return ListResult{}, err
		}
		c.Foil = foilInt == 1
		c.PreRelease = prereleaseInt == 1
		c.Commander = commanderInt == 1
		result = append(result, c)
	}

	totalPages := (total + params.PageSize - 1) / params.PageSize
	if totalPages < 1 {
		totalPages = 1
	}

	return ListResult{
		Cards:         result,
		Total:         total,
		TotalQuantity: totalQuantity,
		Page:          params.Page,
		PageSize:      params.PageSize,
		TotalPages:    totalPages,
	}, nil
}

func (r *Repository) Create(card Card) (int64, error) {
	stmt, err := r.db.Prepare("INSERT INTO cards " +
		"(mtg_id, name, color, `type`, subtitle, collection_number," +
		" rarity, set_code, mana_cost, colors, language, year," +
		" artist, company, foil, quantity, `condition`, notes, prerelease, commander, precon_deck, deck_id)" +
		" VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	foilInt := 0
	if card.Foil {
		foilInt = 1
	}
	prereleaseInt := 0
	if card.PreRelease {
		prereleaseInt = 1
	}
	commanderInt := 0
	if card.Commander {
		commanderInt = 1
	}

	result, err := stmt.Exec(
		card.MTGID, card.Name, card.Color, card.Type, card.Subtitle,
		card.CollectionNumber, card.Rarity, card.SetCode, card.ManaCost,
		card.Colors, card.Language, card.Year, card.Artist, card.Company,
		foilInt, card.Quantity, card.Condition, card.Notes, prereleaseInt, commanderInt, card.PreconDeck, card.DeckID,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) GetByID(id string) (*Card, error) {
	row := r.db.QueryRow(
		"SELECT "+selectCols+" FROM cards WHERE id = ?", id)

	var c Card
	var foilInt, prereleaseInt, commanderInt int
	err := row.Scan(
		&c.ID, &c.MTGID, &c.Name, &c.Color, &c.Type, &c.Subtitle,
		&c.CollectionNumber, &c.Rarity, &c.SetCode, &c.ManaCost,
		&c.Colors, &c.Language, &c.Year, &c.Artist, &c.Company,
		&foilInt, &c.Quantity, &c.Condition, &c.Notes, &prereleaseInt, &commanderInt, &c.PreconDeck, &c.DeckID,
	)
	if err != nil {
		return nil, err
	}
	c.Foil = foilInt == 1
	c.PreRelease = prereleaseInt == 1
	c.Commander = commanderInt == 1
	return &c, nil
}

func (r *Repository) Update(id string, card Card) error {
	foilInt := 0
	if card.Foil {
		foilInt = 1
	}
	prereleaseInt := 0
	if card.PreRelease {
		prereleaseInt = 1
	}
	commanderInt := 0
	if card.Commander {
		commanderInt = 1
	}
	_, err := r.db.Exec(
		"UPDATE cards SET name=?, color=?, `type`=?, subtitle=?, collection_number=?,"+
			" rarity=?, set_code=?, language=?, year=?, artist=?, company=?,"+
			" foil=?, prerelease=?, commander=?, precon_deck=?, deck_id=?, quantity=?, `condition`=?, notes=? WHERE id=?",
		card.Name, card.Color, card.Type, card.Subtitle, card.CollectionNumber,
		card.Rarity, card.SetCode, card.Language, card.Year, card.Artist,
		card.Company, foilInt, prereleaseInt, commanderInt, card.PreconDeck, card.DeckID,
		card.Quantity, card.Condition, card.Notes, id,
	)
	return err
}

// UpdateSharedByIdentity atualiza os campos compartilhados (tudo exceto quantity, condition, notes)
// em todas as cartas com a mesma identidade (name+set+number+language+foil).
func (r *Repository) UpdateSharedByIdentity(oldName, oldSetCode, oldCollNum, oldLang string, oldFoil bool, card Card) error {
	oldFoilInt := 0
	if oldFoil {
		oldFoilInt = 1
	}
	newFoilInt := 0
	if card.Foil {
		newFoilInt = 1
	}
	_, err := r.db.Exec(
		"UPDATE cards SET name=?, color=?, `type`=?, subtitle=?, collection_number=?,"+
			" rarity=?, set_code=?, language=?, year=?, artist=?, company=?, foil=?"+
			" WHERE name=? AND set_code=? AND collection_number=? AND language=? AND foil=?",
		card.Name, card.Color, card.Type, card.Subtitle, card.CollectionNumber,
		card.Rarity, card.SetCode, card.Language, card.Year, card.Artist, card.Company, newFoilInt,
		oldName, oldSetCode, oldCollNum, oldLang, oldFoilInt,
	)
	return err
}

func (r *Repository) UpdateMTGID(id, mtgID string) error {
	_, err := r.db.Exec(`UPDATE cards SET mtg_id = ? WHERE id = ?`, mtgID, id)
	return err
}

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM cards WHERE id = ?`, id)
	return err
}

func (r *Repository) SetDeck(id string, deckID int) error {
	_, err := r.db.Exec(`UPDATE cards SET deck_id = ? WHERE id = ?`, deckID, id)
	return err
}

func (r *Repository) ListAll() ([]Card, error) {
	rows, err := r.db.Query(`SELECT ` + selectCols + ` FROM cards ORDER BY name ASC, set_code ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Card
	for rows.Next() {
		var c Card
		var foilInt, prereleaseInt, commanderInt int
		err := rows.Scan(
			&c.ID, &c.MTGID, &c.Name, &c.Color, &c.Type, &c.Subtitle,
			&c.CollectionNumber, &c.Rarity, &c.SetCode, &c.ManaCost,
			&c.Colors, &c.Language, &c.Year, &c.Artist, &c.Company,
			&foilInt, &c.Quantity, &c.Condition, &c.Notes, &prereleaseInt, &commanderInt, &c.PreconDeck, &c.DeckID,
		)
		if err != nil {
			return nil, err
		}
		c.Foil = foilInt == 1
		c.PreRelease = prereleaseInt == 1
		c.Commander = commanderInt == 1
		result = append(result, c)
	}
	return result, nil
}

// EvalCardInfo contém os campos mínimos necessários para gerar a avaliação IA de um deck.
type EvalCardInfo struct {
	Name     string
	Type     string
	ManaCost string
	Rarity   string
}

// ListForEval retorna nome, tipo, custo de mana e raridade de todas as cartas de um deck.
func (r *Repository) ListForEval(deckID int) ([]EvalCardInfo, error) {
	rows, err := r.db.Query(
		`SELECT name, COALESCE(` + "`type`" + `, ''), COALESCE(mana_cost, ''), COALESCE(rarity, '')
		 FROM cards WHERE deck_id = ? ORDER BY ` + "`type`" + `, name`, deckID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []EvalCardInfo
	for rows.Next() {
		var c EvalCardInfo
		if err := rows.Scan(&c.Name, &c.Type, &c.ManaCost, &c.Rarity); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, nil
}
