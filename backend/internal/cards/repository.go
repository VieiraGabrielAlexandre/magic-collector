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
	DeckIDFilter    *int   // nil = all; 0 = without deck; >0 = specific deck
	FoilOnly        bool   // true = somente foil
	FullArtOnly     bool   // true = somente full art
	RarityFilter    string // "" = todas; "L","C","U","R","M","T"
	ColorsFilter    string // "" = todas; "W" | "U,G" | "W,U,B" etc.
	TypeFilter      string // "" = todos; busca parcial no type_line
	SubtypeFilter   string // "" = todas; busca na estirpe (após o —)
	ConditionFilter string // "" = todas; "mint" | "near_mint" | "played" | "damaged"
	LangFilter      string // "" = todas; "EN","PT","JP","ES","FR","DE"
	PrereleaseFilter string // "" = todas; "1" = somente prerelease; "0" = excluir
	ProxyFilter     string // "" = todas; "1" = somente proxy; "0" = excluir proxy
	CommanderFilter string // "" = todas; "1" = somente commander
	PricedFilter    string // "" = todas; "priced" = com preço; "unpriced" = sem preço
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
	"created_at":        "created_at",
}

// selectCols lista as colunas na mesma ordem que os Scan abaixo.
// `condition` e `type` são palavras reservadas no MySQL e precisam de backticks.
const selectCols = `id, mtg_id, name, color, ` + "`type`" + `, subtitle, collection_number,
	       rarity, set_code, mana_cost, colors, language, year,
	       artist, foil, quantity, ` + "`condition`" + `, notes, prerelease, commander, precon_deck, deck_id, price_usd, COALESCE(DATE_FORMAT(price_updated_at,'%Y-%m-%dT%H:%i:%s'),'') AS price_updated_at, image_url, full_art, double_faced, image_url_back, proxy, storage_id, COALESCE(precon_original,1) AS precon_original`

func (r *Repository) List(params ListParams) (ListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 500 {
		params.PageSize = 20
	}

	sortCol := "created_at"
	if col, ok := allowedSortFields[params.Sort]; ok {
		sortCol = col
	}
	order := "DESC"
	if strings.ToUpper(params.Order) == "ASC" {
		order = "ASC"
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
			like := "%" + params.Search + "%"
			// FULLTEXT para buscas de palavra única com 3+ chars; LIKE para o restante.
			if len(params.Search) >= 3 && !strings.ContainsAny(params.Search, " \t") {
				clauses = append(clauses, "(MATCH(name) AGAINST(? IN BOOLEAN MODE) OR set_code LIKE ? OR color LIKE ? OR `type` LIKE ? OR artist LIKE ? OR collection_number LIKE ?)")
				args = append(args, params.Search+"*", like, like, like, like, like)
			} else {
				clauses = append(clauses, "(name LIKE ? OR set_code LIKE ? OR color LIKE ? OR `type` LIKE ? OR artist LIKE ? OR collection_number LIKE ?)")
				args = append(args, like, like, like, like, like, like)
			}
		}
	}
	if params.DeckIDFilter != nil {
		clauses = append(clauses, "deck_id = ?")
		args = append(args, *params.DeckIDFilter)
	}
	if params.FoilOnly {
		clauses = append(clauses, "foil = 1")
	}
	if params.FullArtOnly {
		clauses = append(clauses, "full_art = 1")
	}
	if params.RarityFilter != "" {
		clauses = append(clauses, "rarity = ?")
		args = append(args, params.RarityFilter)
	}
	if params.TypeFilter == "Nonbasic Land" {
		clauses = append(clauses, "`type` LIKE '%Land%' AND `type` NOT LIKE '%Basic%'")
	} else if params.TypeFilter != "" {
		clauses = append(clauses, "`type` LIKE ?")
		args = append(args, "%"+params.TypeFilter+"%")
	}
	if params.SubtypeFilter != "" {
		// busca após o traço em type (ex: "Creature — Elf Warrior")
		clauses = append(clauses, "`type` LIKE ?")
		args = append(args, "%— %"+params.SubtypeFilter+"%")
	}
	if params.ConditionFilter != "" {
		clauses = append(clauses, "`condition` = ?")
		args = append(args, params.ConditionFilter)
	}
	if params.LangFilter != "" {
		clauses = append(clauses, "UPPER(COALESCE(language,'EN')) = ?")
		args = append(args, strings.ToUpper(params.LangFilter))
	}
	if params.PrereleaseFilter == "1" {
		clauses = append(clauses, "prerelease = 1")
	} else if params.PrereleaseFilter == "0" {
		clauses = append(clauses, "prerelease = 0")
	}
	if params.ProxyFilter == "1" {
		clauses = append(clauses, "proxy = 1")
	} else if params.ProxyFilter == "0" {
		clauses = append(clauses, "proxy = 0")
	}
	if params.CommanderFilter == "1" {
		clauses = append(clauses, "commander = 1")
	}
	if params.PricedFilter == "priced" {
		clauses = append(clauses, "price_usd > 0")
	} else if params.PricedFilter == "unpriced" {
		clauses = append(clauses, "price_usd = 0")
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
		var foilInt, prereleaseInt, commanderInt, fullArtInt, proxyInt, preconOriginalInt int
		err := rows.Scan(
			&c.ID, &c.MTGID, &c.Name, &c.Color, &c.Type, &c.Subtitle,
			&c.CollectionNumber, &c.Rarity, &c.SetCode, &c.ManaCost,
			&c.Colors, &c.Language, &c.Year, &c.Artist,
			&foilInt, &c.Quantity, &c.Condition, &c.Notes, &prereleaseInt, &commanderInt, &c.PreconDeck, &c.DeckID, &c.PriceUSD, &c.PriceUpdatedAt, &c.ImageURL, &fullArtInt, &c.DoubleFaced, &c.ImageURLBack, &proxyInt, &c.StorageID, &preconOriginalInt,
		)
		if err != nil {
			return ListResult{}, err
		}
		c.Foil = foilInt == 1
		c.PreRelease = prereleaseInt == 1
		c.Commander = commanderInt == 1
		c.FullArt = fullArtInt == 1
		c.PreconOriginal = preconOriginalInt == 1
		c.Proxy = proxyInt == 1
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
		" artist, foil, quantity, `condition`, notes, prerelease, commander, precon_deck, deck_id, price_usd, image_url, full_art, double_faced, image_url_back, proxy, storage_id, precon_original)" +
		" VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
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
	fullArtInt := 0
	if card.FullArt {
		fullArtInt = 1
	}
	doubleFacedInt := 0
	if card.DoubleFaced {
		doubleFacedInt = 1
	}
	proxyInt := 0
	if card.Proxy {
		proxyInt = 1
	}
	preconOriginalInt := 1
	if !card.PreconOriginal {
		preconOriginalInt = 0
	}

	result, err := stmt.Exec(
		card.MTGID, card.Name, card.Color, card.Type, card.Subtitle,
		card.CollectionNumber, card.Rarity, card.SetCode, card.ManaCost,
		card.Colors, card.Language, card.Year, card.Artist,
		foilInt, card.Quantity, card.Condition, card.Notes, prereleaseInt, commanderInt, card.PreconDeck, card.DeckID, card.PriceUSD, card.ImageURL, fullArtInt, doubleFacedInt, card.ImageURLBack, proxyInt, card.StorageID, preconOriginalInt,
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
	var foilInt, prereleaseInt, commanderInt, fullArtInt, proxyInt, preconOriginalInt int
	err := row.Scan(
		&c.ID, &c.MTGID, &c.Name, &c.Color, &c.Type, &c.Subtitle,
		&c.CollectionNumber, &c.Rarity, &c.SetCode, &c.ManaCost,
		&c.Colors, &c.Language, &c.Year, &c.Artist,
		&foilInt, &c.Quantity, &c.Condition, &c.Notes, &prereleaseInt, &commanderInt, &c.PreconDeck, &c.DeckID, &c.PriceUSD, &c.PriceUpdatedAt, &c.ImageURL, &fullArtInt, &c.DoubleFaced, &c.ImageURLBack, &proxyInt, &c.StorageID, &preconOriginalInt,
	)
	if err != nil {
		return nil, err
	}
	c.Foil = foilInt == 1
	c.PreRelease = prereleaseInt == 1
	c.Commander = commanderInt == 1
	c.FullArt = fullArtInt == 1
	c.Proxy = proxyInt == 1
	c.PreconOriginal = preconOriginalInt == 1
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
	fullArtIntU := 0
	if card.FullArt {
		fullArtIntU = 1
	}
	doubleFacedIntU := 0
	if card.DoubleFaced {
		doubleFacedIntU = 1
	}
	proxyIntU := 0
	if card.Proxy {
		proxyIntU = 1
	}
	preconOriginalIntU := 1
	if !card.PreconOriginal {
		preconOriginalIntU = 0
	}
	_, err := r.db.Exec(
		"UPDATE cards SET mtg_id=?, name=?, color=?, colors=?, `type`=?, subtitle=?, collection_number=?,"+
			" rarity=?, set_code=?, mana_cost=?, language=?, year=?, artist=?,"+
			" foil=?, prerelease=?, commander=?, precon_deck=?, deck_id=?, quantity=?, `condition`=?, notes=?, price_usd=?, image_url=?, full_art=?, double_faced=?, image_url_back=?, proxy=?, storage_id=?, precon_original=? WHERE id=?",
		card.MTGID, card.Name, card.Color, card.Colors, card.Type, card.Subtitle, card.CollectionNumber,
		card.Rarity, card.SetCode, card.ManaCost, card.Language, card.Year, card.Artist,
		foilInt, prereleaseInt, commanderInt, card.PreconDeck, card.DeckID,
		card.Quantity, card.Condition, card.Notes, card.PriceUSD, card.ImageURL, fullArtIntU, doubleFacedIntU, card.ImageURLBack, proxyIntU, card.StorageID, preconOriginalIntU, id,
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
			" rarity=?, set_code=?, language=?, year=?, artist=?, foil=?"+
			" WHERE name=? AND set_code=? AND collection_number=? AND language=? AND foil=?",
		card.Name, card.Color, card.Colors, card.Type, card.Subtitle, card.CollectionNumber,
		card.Rarity, card.SetCode, card.Language, card.Year, card.Artist, newFoilInt,
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

func (r *Repository) SetCommander(id string, value bool) error {
	v := 0
	if value {
		v = 1
	}
	_, err := r.db.Exec(`UPDATE cards SET commander = ? WHERE id = ?`, v, id)
	return err
}

func (r *Repository) SetPreconOriginal(id string, value bool) error {
	v := 0
	if value {
		v = 1
	}
	_, err := r.db.Exec(`UPDATE cards SET precon_original = ? WHERE id = ?`, v, id)
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
		var foilInt, prereleaseInt, commanderInt, fullArtInt, proxyInt, preconOriginalInt int
		err := rows.Scan(
			&c.ID, &c.MTGID, &c.Name, &c.Color, &c.Type, &c.Subtitle,
			&c.CollectionNumber, &c.Rarity, &c.SetCode, &c.ManaCost,
			&c.Colors, &c.Language, &c.Year, &c.Artist,
			&foilInt, &c.Quantity, &c.Condition, &c.Notes, &prereleaseInt, &commanderInt, &c.PreconDeck, &c.DeckID, &c.PriceUSD, &c.PriceUpdatedAt, &c.ImageURL, &fullArtInt, &c.DoubleFaced, &c.ImageURLBack, &proxyInt, &c.StorageID, &preconOriginalInt,
		)
		if err != nil {
			return nil, err
		}
		c.Foil = foilInt == 1
		c.PreRelease = prereleaseInt == 1
		c.Commander = commanderInt == 1
		c.FullArt = fullArtInt == 1
		c.Proxy = proxyInt == 1
		c.PreconOriginal = preconOriginalInt == 1
		result = append(result, c)
	}
	return result, nil
}

func (r *Repository) ListProxies() ([]Card, error) {
	rows, err := r.db.Query(`SELECT ` + selectCols + ` FROM cards WHERE proxy = 1 ORDER BY name ASC, set_code ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Card
	for rows.Next() {
		var c Card
		var foilInt, prereleaseInt, commanderInt, fullArtInt, proxyInt, preconOriginalInt int
		err := rows.Scan(
			&c.ID, &c.MTGID, &c.Name, &c.Color, &c.Type, &c.Subtitle,
			&c.CollectionNumber, &c.Rarity, &c.SetCode, &c.ManaCost,
			&c.Colors, &c.Language, &c.Year, &c.Artist,
			&foilInt, &c.Quantity, &c.Condition, &c.Notes, &prereleaseInt, &commanderInt, &c.PreconDeck, &c.DeckID, &c.PriceUSD, &c.PriceUpdatedAt, &c.ImageURL, &fullArtInt, &c.DoubleFaced, &c.ImageURLBack, &proxyInt, &c.StorageID, &preconOriginalInt,
		)
		if err != nil {
			return nil, err
		}
		c.Foil = foilInt == 1
		c.PreRelease = prereleaseInt == 1
		c.Commander = commanderInt == 1
		c.FullArt = fullArtInt == 1
		c.Proxy = proxyInt == 1
		c.PreconOriginal = preconOriginalInt == 1
		result = append(result, c)
	}
	return result, nil
}

func (r *Repository) SetStorage(id string, storageID int) error {
	_, err := r.db.Exec(`UPDATE cards SET storage_id = ? WHERE id = ?`, storageID, id)
	return err
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
	_, err := r.db.Exec(`UPDATE cards SET price_usd = ?, price_updated_at = NOW() WHERE id = ?`, price, id)
	return err
}

func (r *Repository) UpdatePriceAndMTGID(id int, mtgID string, price float64) error {
	_, err := r.db.Exec(`UPDATE cards SET price_usd = ?, price_updated_at = NOW(), mtg_id = ? WHERE id = ?`, price, mtgID, id)
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

// EvalCardInfo contém os campos necessários para gerar a avaliação IA de um deck.
type EvalCardInfo struct {
	Name             string
	Type             string
	ManaCost         string
	Rarity           string
	SetCode          string
	CollectionNumber string
	Quantity         int
	IsCommander      bool
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

// AnalysisCard inclui set_code, collection_number e image_url para a tela de análise completa.
type AnalysisCard struct {
	Name             string
	Type             string
	ManaCost         string
	Rarity           string
	Colors           string
	Quantity         int
	SetCode          string
	CollectionNumber string
	ImageURL         string
}

// ── Estatísticas da coleção ─────────────────────────────────────────────

type TopCard struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	PriceUSD float64 `json:"price_usd"`
	ImageURL string  `json:"image_url"`
	SetCode  string  `json:"set_code"`
	Rarity   string  `json:"rarity"`
	Foil     bool    `json:"foil"`
}

type CollectionStats struct {
	TotalQuantity  int              `json:"total_quantity"`
	UniqueCards    int              `json:"unique_cards"`
	FoilCount      int              `json:"foil_count"`
	FoilQuantity   int              `json:"foil_quantity"`
	EstimatedValue float64          `json:"estimated_value_usd"`
	DeckValue      float64          `json:"deck_value_usd"`
	NonDeckValue   float64          `json:"non_deck_value_usd"`
	PricedCards    int              `json:"priced_cards"`
	CardsThisWeek  int              `json:"cards_this_week"`
	ByRarity       []RarityCount    `json:"by_rarity"`
	ByColor        []ColorCount     `json:"by_color"`
	TopSets        []SetCount       `json:"top_sets"`
	TopCards       []TopCard        `json:"top_cards"`
	TopByRarity    []TopCard        `json:"top_by_rarity"`
	ByType         []TypeCount      `json:"by_type"`
	TopQuantity    []TopQuantCard   `json:"top_quantity"`
}

type TypeCount struct {
	TypeGroup string `json:"type_group"`
	Count     int    `json:"count"`
	Quantity  int    `json:"quantity"`
}

type TopQuantCard struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	SetCode  string `json:"set_code"`
	Rarity   string `json:"rarity"`
	ImageURL string `json:"image_url"`
	Foil     bool   `json:"foil"`
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

func (r *Repository) GetSubtypes() ([]string, error) {
	rows, err := r.db.Query("SELECT DISTINCT `type` FROM cards WHERE `type` LIKE '%—%' AND `type` != ''")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seen := make(map[string]struct{})
	for rows.Next() {
		var typeLine string
		if err := rows.Scan(&typeLine); err != nil {
			continue
		}
		// pega o que vem após o "—"
		idx := strings.Index(typeLine, "—")
		if idx < 0 {
			continue
		}
		after := strings.TrimSpace(typeLine[idx+len("—"):])
		for _, part := range strings.Fields(after) {
			part = strings.Trim(part, ".,;'\"")
			if len(part) >= 2 {
				seen[part] = struct{}{}
			}
		}
	}

	result := make([]string, 0, len(seen))
	for s := range seen {
		result = append(result, s)
	}
	sort.Strings(result)
	return result, nil
}

func (r *Repository) GetDuplicates() ([]int, error) {
	rows, err := r.db.Query(`
		SELECT DISTINCT c1.id
		FROM cards c1
		WHERE c1.mtg_id != ''
		AND EXISTS (
			SELECT 1 FROM cards c2
			WHERE c2.mtg_id = c1.mtg_id AND c2.id != c1.id
		)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repository) GetStats() (CollectionStats, error) {
	var s CollectionStats

	// Totais gerais
	err := r.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(quantity),0),
		       COALESCE(SUM(CASE WHEN foil=1 THEN 1 ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN foil=1 THEN quantity ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN price_usd>0 THEN price_usd*quantity ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN price_usd>0 THEN 1 ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN price_usd>0 AND deck_id>0 THEN price_usd*quantity ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN price_usd>0 AND (deck_id=0 OR deck_id IS NULL) THEN price_usd*quantity ELSE 0 END),0)
		FROM cards`).Scan(
		&s.UniqueCards, &s.TotalQuantity,
		&s.FoilCount, &s.FoilQuantity,
		&s.EstimatedValue, &s.PricedCards,
		&s.DeckValue, &s.NonDeckValue)
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

	// Cartas adicionadas nos últimos 7 dias
	r.db.QueryRow(`SELECT COUNT(*) FROM cards WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)`).Scan(&s.CardsThisWeek)

	// Distribuição por tipo (agrupada)
	typeRows, err := r.db.Query(`
		SELECT
		  CASE
		    WHEN type LIKE '%Creature%'     THEN 'Criatura'
		    WHEN type LIKE '%Instant%'      THEN 'Instantâneo'
		    WHEN type LIKE '%Sorcery%'      THEN 'Feitiço'
		    WHEN type LIKE '%Enchantment%'  THEN 'Encantamento'
		    WHEN type LIKE '%Artifact%'     THEN 'Artefato'
		    WHEN type LIKE '%Land%'         THEN 'Terreno'
		    WHEN type LIKE '%Planeswalker%' THEN 'Planeswalker'
		    WHEN type LIKE '%Battle%'       THEN 'Battle'
		    ELSE 'Outro'
		  END AS tg,
		  COUNT(*) as cnt,
		  COALESCE(SUM(quantity),0) as qty
		FROM cards
		WHERE type IS NOT NULL AND type != ''
		GROUP BY tg
		ORDER BY cnt DESC`)
	if err == nil {
		defer typeRows.Close()
		for typeRows.Next() {
			var tc TypeCount
			typeRows.Scan(&tc.TypeGroup, &tc.Count, &tc.Quantity)
			s.ByType = append(s.ByType, tc)
		}
	}

	// Top 6 cartas com mais cópias
	qtyRows, err := r.db.Query(`
		SELECT id, name, quantity, COALESCE(image_url,''), COALESCE(set_code,''), COALESCE(rarity,''), foil
		FROM cards WHERE quantity > 1 ORDER BY quantity DESC LIMIT 6`)
	if err == nil {
		defer qtyRows.Close()
		for qtyRows.Next() {
			var tq TopQuantCard
			var foilInt int
			qtyRows.Scan(&tq.ID, &tq.Name, &tq.Quantity, &tq.ImageURL, &tq.SetCode, &tq.Rarity, &foilInt)
			tq.Foil = foilInt == 1
			s.TopQuantity = append(s.TopQuantity, tq)
		}
	}

	// Top 10 cartas mais valiosas
	topRows, err := r.db.Query(`
		SELECT id, name, price_usd, COALESCE(image_url,''), COALESCE(set_code,''), COALESCE(rarity,''), foil
		FROM cards WHERE price_usd > 0 ORDER BY price_usd DESC LIMIT 10`)
	if err == nil {
		defer topRows.Close()
		for topRows.Next() {
			var tc TopCard
			var foilInt int
			if err := topRows.Scan(&tc.ID, &tc.Name, &tc.PriceUSD, &tc.ImageURL, &tc.SetCode, &tc.Rarity, &foilInt); err == nil {
				tc.Foil = foilInt == 1
				s.TopCards = append(s.TopCards, tc)
			}
		}
	}

	// Top 1 carta mais valiosa por raridade
	rarityRows, err := r.db.Query(`
		SELECT id, name, price_usd, COALESCE(image_url,''), COALESCE(set_code,''), COALESCE(rarity,''), foil
		FROM (
			SELECT id, name, price_usd, image_url, set_code, rarity, foil,
			       ROW_NUMBER() OVER (PARTITION BY rarity ORDER BY price_usd DESC) AS rn
			FROM cards WHERE price_usd > 0 AND rarity IN ('M','R','U','C','L','T')
		) t WHERE rn = 1
		ORDER BY FIELD(rarity,'M','R','U','C','L','T')`)
	if err == nil {
		defer rarityRows.Close()
		for rarityRows.Next() {
			var tc TopCard
			var foilInt int
			if err := rarityRows.Scan(&tc.ID, &tc.Name, &tc.PriceUSD, &tc.ImageURL, &tc.SetCode, &tc.Rarity, &foilInt); err == nil {
				tc.Foil = foilInt == 1
				s.TopByRarity = append(s.TopByRarity, tc)
			}
		}
	}

	return s, nil
}

// ListAllForAnalysis retorna TODAS as cartas da coleção (independente de deck)
// agrupadas por nome, para a tela de análise completa.
func (r *Repository) ListAllForAnalysis() ([]AnalysisCard, error) {
	rows, err := r.db.Query(`
		SELECT name,
		       COALESCE(MAX(` + "`type`" + `), '') AS type,
		       COALESCE(MAX(mana_cost), '') AS mana_cost,
		       COALESCE(MAX(rarity), '') AS rarity,
		       COALESCE(MAX(colors), '') AS colors,
		       SUM(quantity) AS total_qty,
		       COALESCE(MAX(set_code), '') AS set_code,
		       COALESCE(MAX(collection_number), '') AS collection_number,
		       COALESCE(MAX(image_url), '') AS image_url
		FROM cards
		GROUP BY name
		ORDER BY ` + "`type`" + `, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []AnalysisCard
	for rows.Next() {
		var c AnalysisCard
		if err := rows.Scan(&c.Name, &c.Type, &c.ManaCost, &c.Rarity, &c.Colors, &c.Quantity, &c.SetCode, &c.CollectionNumber, &c.ImageURL); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, nil
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

// ListForEval retorna todos os dados necessários para gerar a avaliação IA de um deck.
func (r *Repository) ListForEval(deckID int) ([]EvalCardInfo, error) {
	rows, err := r.db.Query(
		`SELECT name, COALESCE(`+"`type`"+`, ''), COALESCE(mana_cost, ''), COALESCE(rarity, ''),
		        COALESCE(set_code, ''), COALESCE(collection_number, ''), quantity, commander
		 FROM cards WHERE deck_id = ? ORDER BY commander DESC, `+"`type`"+`, name`, deckID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []EvalCardInfo
	for rows.Next() {
		var c EvalCardInfo
		var cmdInt int
		if err := rows.Scan(&c.Name, &c.Type, &c.ManaCost, &c.Rarity, &c.SetCode, &c.CollectionNumber, &c.Quantity, &cmdInt); err != nil {
			return nil, err
		}
		c.IsCommander = cmdInt == 1
		result = append(result, c)
	}
	return result, nil
}

// ── Histórico de preço ────────────────────────────────────────────────────────

type PricePoint struct {
	Date     string  `json:"date"`
	PriceUSD float64 `json:"price_usd"`
}

func (r *Repository) InsertPriceHistory(cardID int, priceUSD float64, date string) error {
	_, err := r.db.Exec(
		`INSERT INTO card_price_history (card_id, price_usd, recorded_at)
		 VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE price_usd = VALUES(price_usd)`,
		cardID, priceUSD, date,
	)
	return err
}

func (r *Repository) GetPriceHistory(cardID string) ([]PricePoint, error) {
	rows, err := r.db.Query(
		`SELECT recorded_at, price_usd FROM card_price_history
		 WHERE card_id = ? ORDER BY recorded_at ASC LIMIT 90`,
		cardID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []PricePoint
	for rows.Next() {
		var p PricePoint
		if err := rows.Scan(&p.Date, &p.PriceUSD); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}
