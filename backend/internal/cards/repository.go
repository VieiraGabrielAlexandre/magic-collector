package cards

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
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
	DeckIDFilter *int   // nil = all; 0 = without deck; >0 = specific deck
	FoilOnly     bool   // true = somente foil
	RarityFilter string // "" = todas; "L","C","U","R","M","T"
	ColorsFilter string // "" = todas; "W" | "U,G" | "W,U,B" etc.
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
	"price_usd":         "price_usd",
	"quantity":          "quantity",
}

// selectCols lista as colunas na mesma ordem que os Scan abaixo.
// `condition` e `type` são palavras reservadas no MySQL e precisam de backticks.
const selectCols = `id, mtg_id, name, color, ` + "`type`" + `, subtitle, collection_number,
	       rarity, set_code, mana_cost, colors, language, year,
	       artist, company, foil, quantity, ` + "`condition`" + `, notes, prerelease, commander, precon_deck, deck_id, price_usd, image_url`

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

	var clauses []string
	args := []any{}
	if params.Search != "" {
		if strings.HasPrefix(params.Search, "#") {
			// Busca exata por número da carta quando o termo começa com '#'
			num := params.Search[1:]
			clauses = append(clauses, "collection_number = ?")
			args = append(args, num)
		} else {
			clauses = append(clauses, "(name LIKE ? OR set_code LIKE ? OR color LIKE ? OR `type` LIKE ? OR artist LIKE ? OR collection_number LIKE ?)")
			like := "%" + params.Search + "%"
			args = append(args, like, like, like, like, like, like)
		}
	}
	if params.DeckIDFilter != nil {
		clauses = append(clauses, "deck_id = ?")
		args = append(args, *params.DeckIDFilter)
	}
	if params.FoilOnly {
		clauses = append(clauses, "foil = 1")
	}
	if params.RarityFilter != "" {
		clauses = append(clauses, "rarity = ?")
		args = append(args, params.RarityFilter)
	}
	if params.ColorsFilter != "" {
		if params.ColorsFilter == "none" {
			clauses = append(clauses, `(colors IS NULL OR colors = '' OR colors = '[]' OR colors = 'null')`)
		} else {
			codes := strings.Split(params.ColorsFilter, ",")
			clean := make([]string, 0, len(codes))
			for _, c := range codes {
				if c = strings.TrimSpace(strings.ToUpper(c)); c != "" {
					clean = append(clean, c)
				}
			}
			if len(clean) > 0 {
				clauses = append(clauses, `(colors IS NOT NULL AND colors != '' AND colors != '[]' AND colors != 'null')`)
				for _, code := range clean {
					clauses = append(clauses, `JSON_CONTAINS(colors, ?, '$') = 1`)
					args = append(args, `"`+code+`"`)
				}
				clauses = append(clauses, `JSON_LENGTH(colors) = ?`)
				args = append(args, len(clean))
			}
		}
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
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
			&foilInt, &c.Quantity, &c.Condition, &c.Notes, &prereleaseInt, &commanderInt, &c.PreconDeck, &c.DeckID, &c.PriceUSD, &c.ImageURL,
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
		" artist, company, foil, quantity, `condition`, notes, prerelease, commander, precon_deck, deck_id, price_usd, image_url)" +
		" VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
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
		foilInt, card.Quantity, card.Condition, card.Notes, prereleaseInt, commanderInt, card.PreconDeck, card.DeckID, card.PriceUSD, card.ImageURL,
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
		&foilInt, &c.Quantity, &c.Condition, &c.Notes, &prereleaseInt, &commanderInt, &c.PreconDeck, &c.DeckID, &c.PriceUSD, &c.ImageURL,
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
		"UPDATE cards SET name=?, color=?, colors=?, `type`=?, subtitle=?, collection_number=?,"+
			" rarity=?, set_code=?, language=?, year=?, artist=?, company=?,"+
			" foil=?, prerelease=?, commander=?, precon_deck=?, deck_id=?, quantity=?, `condition`=?, notes=?, price_usd=?, image_url=? WHERE id=?",
		card.Name, card.Color, card.Colors, card.Type, card.Subtitle, card.CollectionNumber,
		card.Rarity, card.SetCode, card.Language, card.Year, card.Artist,
		card.Company, foilInt, prereleaseInt, commanderInt, card.PreconDeck, card.DeckID,
		card.Quantity, card.Condition, card.Notes, card.PriceUSD, card.ImageURL, id,
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
		"UPDATE cards SET name=?, color=?, colors=?, `type`=?, subtitle=?, collection_number=?,"+
			" rarity=?, set_code=?, language=?, year=?, artist=?, company=?, foil=?"+
			" WHERE name=? AND set_code=? AND collection_number=? AND language=? AND foil=?",
		card.Name, card.Color, card.Colors, card.Type, card.Subtitle, card.CollectionNumber,
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

func (r *Repository) SetQuantity(id string, quantity int) error {
	_, err := r.db.Exec(`UPDATE cards SET quantity = ? WHERE id = ?`, quantity, id)
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
			&foilInt, &c.Quantity, &c.Condition, &c.Notes, &prereleaseInt, &commanderInt, &c.PreconDeck, &c.DeckID, &c.PriceUSD, &c.ImageURL,
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

// ── Atualização de preços ────────────────────────────────────────────────

// CardForPriceRefresh contém os campos necessários para buscar e atualizar o preço via Scryfall.
type CardForPriceRefresh struct {
	ID               int
	MTGID            string
	SetCode          string
	CollectionNumber string
	Language         string
	Artist           string
	Foil             bool
	Name             string
}

func (r *Repository) ListAllForPriceRefresh() ([]CardForPriceRefresh, error) {
	return r.listForPriceRefresh(false)
}

// ListEmptyPricesForRefresh retorna apenas cartas com price_usd = 0.
func (r *Repository) ListEmptyPricesForRefresh() ([]CardForPriceRefresh, error) {
	return r.listForPriceRefresh(true)
}

func (r *Repository) listForPriceRefresh(emptyOnly bool) ([]CardForPriceRefresh, error) {
	q := `SELECT id, COALESCE(mtg_id,''), COALESCE(set_code,''),
		       COALESCE(collection_number,''), COALESCE(language,'EN'),
		       COALESCE(artist,''), foil, name
		FROM cards`
	if emptyOnly {
		q += ` WHERE price_usd = 0`
	}
	q += ` ORDER BY id`

	rows, err := r.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []CardForPriceRefresh
	for rows.Next() {
		var c CardForPriceRefresh
		var foilInt int
		if err := rows.Scan(&c.ID, &c.MTGID, &c.SetCode, &c.CollectionNumber, &c.Language, &c.Artist, &foilInt, &c.Name); err != nil {
			return nil, err
		}
		c.Foil = foilInt == 1
		result = append(result, c)
	}
	return result, nil
}

func (r *Repository) UpdatePrice(id int, price float64) error {
	_, err := r.db.Exec(`UPDATE cards SET price_usd = ? WHERE id = ?`, price, id)
	return err
}

func (r *Repository) UpdatePriceAndMTGID(id int, mtgID string, price float64) error {
	_, err := r.db.Exec(`UPDATE cards SET price_usd = ?, mtg_id = ? WHERE id = ?`, price, mtgID, id)
	return err
}

func (r *Repository) UpdateImageURL(id int, imageURL string) error {
	_, err := r.db.Exec(`UPDATE cards SET image_url = ? WHERE id = ?`, imageURL, id)
	return err
}

func (r *Repository) UpdateImageURLAndMTGID(id int, mtgID, imageURL string) error {
	_, err := r.db.Exec(`UPDATE cards SET image_url = ?, mtg_id = ? WHERE id = ?`, imageURL, mtgID, id)
	return err
}

// ListCardsWithoutColors retorna cartas que ainda não têm cores definidas
// (colors NULL/''/null, mas NÃO '[]' que significa colorless já verificado)
// e que têm set_code ou mtg_id para busca na Scryfall.
func (r *Repository) ListCardsWithoutColors() ([]CardForPriceRefresh, error) {
	q := `SELECT id, COALESCE(mtg_id,''), COALESCE(set_code,''),
		       COALESCE(collection_number,''), COALESCE(language,'EN'),
		       COALESCE(artist,''), foil, name
		FROM cards
		WHERE (colors IS NULL OR colors = '' OR colors = 'null')
		  AND (mtg_id != '' OR (set_code != '' AND collection_number != ''))
		ORDER BY id`

	rows, err := r.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []CardForPriceRefresh
	for rows.Next() {
		var c CardForPriceRefresh
		var foilInt int
		if err := rows.Scan(&c.ID, &c.MTGID, &c.SetCode, &c.CollectionNumber, &c.Language, &c.Artist, &foilInt, &c.Name); err != nil {
			return nil, err
		}
		c.Foil = foilInt == 1
		result = append(result, c)
	}
	return result, nil
}

// NormalizeRarityResult descreve o resultado da normalização em lote.
type NormalizeRarityResult struct {
	Updated int            `json:"updated"`
	Skipped int            `json:"skipped"` // já estavam corretas
	ByValue map[string]int `json:"by_value"` // quantas de cada valor antigo foram corrigidas
}

// NormalizeRarities converte todos os valores de raridade não-padrão para o código de 1 letra.
func (r *Repository) NormalizeRarities() (NormalizeRarityResult, error) {
	// mapeamento: valor antigo → novo
	conversions := map[string]string{
		"common":     "C",
		"Common":     "C",
		"COMMON":     "C",
		"uncommon":   "U",
		"Uncommon":   "U",
		"UNCOMMON":   "U",
		"rare":       "R",
		"Rare":       "R",
		"RARE":       "R",
		"special":    "R",
		"Special":    "R",
		"bonus":      "R",
		"timeshifted": "R",
		"mythic":     "M",
		"Mythic":     "M",
		"MYTHIC":     "M",
		"mythic rare": "M",
		"Mythic Rare": "M",
		"land":       "L",
		"Land":       "L",
		"basic land": "L",
		"token":      "T",
		"Token":      "T",
	}

	result := NormalizeRarityResult{ByValue: map[string]int{}}

	for old, newVal := range conversions {
		res, err := r.db.Exec(`UPDATE cards SET rarity = ? WHERE rarity = ?`, newVal, old)
		if err != nil {
			return result, err
		}
		n, _ := res.RowsAffected()
		if n > 0 {
			result.Updated += int(n)
			result.ByValue[old+" → "+newVal] = int(n)
		}
	}

	return result, nil
}

func (r *Repository) UpdateColors(id int, colors, color string) error {
	_, err := r.db.Exec(`UPDATE cards SET colors = ?, color = ? WHERE id = ?`, colors, color, id)
	return err
}

func (r *Repository) UpdateColorsAndMTGID(id int, mtgID, colors, color string) error {
	_, err := r.db.Exec(`UPDATE cards SET colors = ?, color = ?, mtg_id = ? WHERE id = ?`, colors, color, mtgID, id)
	return err
}

// EvalCardInfo contém os campos mínimos necessários para gerar a avaliação IA de um deck.
type EvalCardInfo struct {
	Name     string
	Type     string
	ManaCost string
	Rarity   string
}

// ColorCombo representa uma combinação de cores disponível na coleção.
type ColorCombo struct {
	Codes string `json:"codes"` // ex: "W,U" — ordem WUBRG
	Count int    `json:"count"`
}

var wubrgOrder = map[string]int{"W": 0, "U": 1, "B": 2, "R": 3, "G": 4, "C": 5}

// ListColorCombos retorna as combinações de cores distintas presentes nas cartas.
func (r *Repository) ListColorCombos() ([]ColorCombo, error) {
	rows, err := r.db.Query(`
		SELECT colors, COUNT(*) AS cnt
		FROM cards
		WHERE colors IS NOT NULL AND colors != '' AND colors != '[]' AND colors != 'null'
		  AND JSON_LENGTH(colors) > 0
		GROUP BY colors
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totals := map[string]int{}
	var ordered []string

	for rows.Next() {
		var colorsJSON string
		var cnt int
		if err := rows.Scan(&colorsJSON, &cnt); err != nil {
			continue
		}
		var codes []string
		if err := json.Unmarshal([]byte(colorsJSON), &codes); err != nil {
			continue
		}
		sort.Slice(codes, func(i, j int) bool {
			return wubrgOrder[codes[i]] < wubrgOrder[codes[j]]
		})
		key := strings.Join(codes, ",")
		if _, exists := totals[key]; !exists {
			ordered = append(ordered, key)
		}
		totals[key] += cnt
	}

	// Ordena: mono → duo → trio… e dentro de cada grupo por contagem desc
	sort.Slice(ordered, func(i, j int) bool {
		ni := strings.Count(ordered[i], ",")
		nj := strings.Count(ordered[j], ",")
		if ni != nj {
			return ni < nj
		}
		return totals[ordered[i]] > totals[ordered[j]]
	})

	result := make([]ColorCombo, 0, len(ordered))
	for _, key := range ordered {
		result = append(result, ColorCombo{Codes: key, Count: totals[key]})
	}

	// Adiciona "sem cor" ao final — cartas sem cor identificável
	var noColorCount int
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM cards WHERE colors IS NULL OR colors = '' OR colors = '[]' OR colors = 'null'`).Scan(&noColorCount)
	if noColorCount > 0 {
		result = append(result, ColorCombo{Codes: "none", Count: noColorCount})
	}

	return result, nil
}

// DeckBuilderCard representa uma carta (agrupada por nome) para análise de deck-building.
type DeckBuilderCard struct {
	Name     string
	Type     string
	ManaCost string
	Rarity   string
	Colors   string
	Quantity int
}

// ── Estatísticas da coleção ─────────────────────────────────────────────

type CollectionStats struct {
	TotalQuantity  int            `json:"total_quantity"`
	UniqueCards    int            `json:"unique_cards"`
	FoilCount      int            `json:"foil_count"`
	FoilQuantity   int            `json:"foil_quantity"`
	EstimatedValue float64        `json:"estimated_value_usd"`
	PricedCards    int            `json:"priced_cards"`
	ByRarity       []RarityCount  `json:"by_rarity"`
	ByColor        []ColorCount   `json:"by_color"`
	TopSets        []SetCount     `json:"top_sets"`
}

type RarityCount struct {
	Rarity   string `json:"rarity"`
	Count    int    `json:"count"`
	Quantity int    `json:"quantity"`
}

type ColorCount struct {
	Color    string `json:"color"`
	Count    int    `json:"count"`
}

type SetCount struct {
	SetCode  string `json:"set_code"`
	Count    int    `json:"count"`
	Quantity int    `json:"quantity"`
}

func (r *Repository) GetStats() (CollectionStats, error) {
	var s CollectionStats

	// Totais gerais
	err := r.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(quantity),0),
		       COALESCE(SUM(CASE WHEN foil=1 THEN 1 ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN foil=1 THEN quantity ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN price_usd>0 THEN price_usd*quantity ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN price_usd>0 THEN 1 ELSE 0 END),0)
		FROM cards`).Scan(
		&s.UniqueCards, &s.TotalQuantity,
		&s.FoilCount, &s.FoilQuantity,
		&s.EstimatedValue, &s.PricedCards)
	if err != nil {
		return s, err
	}

	// Por raridade
	rows, err := r.db.Query(`
		SELECT COALESCE(NULLIF(rarity,''),'?') AS r, COUNT(*), COALESCE(SUM(quantity),0)
		FROM cards GROUP BY r
		ORDER BY FIELD(r,'M','R','U','C','L','T','?')`)
	if err != nil {
		return s, err
	}
	defer rows.Close()
	for rows.Next() {
		var rc RarityCount
		if err := rows.Scan(&rc.Rarity, &rc.Count, &rc.Quantity); err != nil {
			return s, err
		}
		s.ByRarity = append(s.ByRarity, rc)
	}
	rows.Close()

	// Top 15 sets
	setRows, err := r.db.Query(`
		SELECT UPPER(COALESCE(NULLIF(set_code,''),'?')) AS sc, COUNT(*), COALESCE(SUM(quantity),0)
		FROM cards GROUP BY sc ORDER BY COUNT(*) DESC LIMIT 15`)
	if err != nil {
		return s, err
	}
	defer setRows.Close()
	for setRows.Next() {
		var sc SetCount
		if err := setRows.Scan(&sc.SetCode, &sc.Count, &sc.Quantity); err != nil {
			return s, err
		}
		s.TopSets = append(s.TopSets, sc)
	}
	setRows.Close()

	// Distribuição de cores: busca o campo JSON e processa em Go
	colorRows, err := r.db.Query(`SELECT COALESCE(colors,'[]') FROM cards`)
	if err != nil {
		return s, err
	}
	defer colorRows.Close()
	colorMap := map[string]int{}
	for colorRows.Next() {
		var colorsJSON string
		colorRows.Scan(&colorsJSON)
		var colors []string
		json.Unmarshal([]byte(colorsJSON), &colors)
		if len(colors) == 0 {
			colorMap["C"]++
		} else {
			for _, c := range colors {
				colorMap[c]++
			}
		}
	}
	for _, code := range []string{"W", "U", "B", "R", "G", "C"} {
		if n, ok := colorMap[code]; ok {
			s.ByColor = append(s.ByColor, ColorCount{Color: code, Count: n})
		}
	}

	return s, nil
}

// ListForDeckBuilder retorna todas as cartas sem deck agrupadas por nome,
// com a quantidade total disponível de cada uma.
func (r *Repository) ListForDeckBuilder() ([]DeckBuilderCard, error) {
	rows, err := r.db.Query(`
		SELECT name,
		       COALESCE(MAX(` + "`type`" + `), '') AS type,
		       COALESCE(MAX(mana_cost), '') AS mana_cost,
		       COALESCE(MAX(rarity), '') AS rarity,
		       COALESCE(MAX(colors), '') AS colors,
		       SUM(quantity) AS total_qty
		FROM cards
		WHERE deck_id = 0
		GROUP BY name
		ORDER BY ` + "`type`" + `, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []DeckBuilderCard
	for rows.Next() {
		var c DeckBuilderCard
		if err := rows.Scan(&c.Name, &c.Type, &c.ManaCost, &c.Rarity, &c.Colors, &c.Quantity); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, nil
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
