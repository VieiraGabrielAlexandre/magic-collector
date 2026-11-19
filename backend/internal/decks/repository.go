package decks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List() ([]Deck, error) {
	rows, err := r.db.Query(`
		SELECT d.id, d.name, d.description, d.commander, d.colors, d.set_code, d.icon_uri, d.theme_color,
		       d.deckbox_color, COALESCE(d.deck_origin,'custom') AS deck_origin,
		       COALESCE(SUM(c.quantity), 0) AS card_count,
		       COALESCE(SUM(c.price_usd * c.quantity), 0) AS estimated_value_usd,
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
		         d.icon_uri, d.theme_color, d.deckbox_color, d.deck_origin, d.evaluation, d.evaluated_at,
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
			&d.IconURI, &d.ThemeColor, &d.DeckboxColor, &d.DeckOrigin, &d.CardCount, &d.EstimatedValueUSD,
			&d.Evaluation, &d.EvaluatedAt,
			&d.BattleWins, &d.BattleLosses, &d.BattleDraws, &d.BattleTotal,
		); err != nil {
			return nil, err
		}
		d.Commander = commanderInt == 1
		result = append(result, d)
	}

	// Popula comandantes de cada deck em uma única query
	cmdRows, err := r.db.Query(
		`SELECT deck_id, id, name, image_url FROM cards WHERE commander = 1 AND deck_id > 0`)
	if err == nil {
		defer cmdRows.Close()
		cmdMap := map[int][]CommanderCard{}
		for cmdRows.Next() {
			var deckID int
			var cc CommanderCard
			if cmdRows.Scan(&deckID, &cc.ID, &cc.Name, &cc.ImageURL) == nil {
				cmdMap[deckID] = append(cmdMap[deckID], cc)
			}
		}
		for i := range result {
			if cmds, ok := cmdMap[result[i].ID]; ok {
				result[i].Commanders = cmds
			}
		}
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
	origin := input.DeckOrigin
	if origin == "" {
		origin = "custom"
	}
	res, err := r.db.Exec(
		`INSERT INTO decks (name, description, commander, colors, set_code, theme_color, deckbox_color, deck_origin) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		input.Name, input.Description, commanderInt, input.Colors, input.SetCode, input.ThemeColor, input.DeckboxColor, origin,
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
	updateOrigin := input.DeckOrigin
	if updateOrigin == "" {
		updateOrigin = "custom"
	}
	_, err := r.db.Exec(
		`UPDATE decks SET name=?, description=?, commander=?, colors=?, set_code=?, theme_color=?, deckbox_color=?, deck_origin=? WHERE id=?`,
		input.Name, input.Description, commanderInt, input.Colors, input.SetCode, input.ThemeColor, input.DeckboxColor, updateOrigin, id,
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

// Validate executa as regras de Commander:
//  1. Deck marcado como Commander
//  2. Exatamente 100 cartas (soma de quantity)
//  3. Pelo menos 1 carta marcada como commander = 1
//  4. Singleton: quantity > 1 apenas para Basic Lands
//  5. Identidade de cor: cores das cartas devem ser subconjunto das cores do deck
func (r *Repository) Validate(deckID string) (DeckValidation, error) {
	idInt, _ := strconv.Atoi(deckID)
	v := DeckValidation{DeckID: idInt, Issues: []ValidationIssue{}}

	var deckColors string
	var isCommander int
	if err := r.db.QueryRow(
		`SELECT COALESCE(colors,''), commander FROM decks WHERE id = ?`, deckID,
	).Scan(&deckColors, &isCommander); err != nil {
		return v, err
	}

	if isCommander == 0 {
		v.Issues = append(v.Issues, ValidationIssue{
			Code:    "not_commander_deck",
			Message: "Deck não está marcado como Commander",
		})
	}

	// Contagem total de cartas
	if err := r.db.QueryRow(
		`SELECT COALESCE(SUM(quantity), 0) FROM cards WHERE deck_id = ?`, deckID,
	).Scan(&v.CardCount); err != nil {
		return v, err
	}
	if v.CardCount != 100 {
		v.Issues = append(v.Issues, ValidationIssue{
			Code:    "wrong_count",
			Message: fmt.Sprintf("Deck tem %d cartas (Commander exige exatamente 100)", v.CardCount),
		})
	}

	// Carta comandante definida
	var cmdCount int
	if err := r.db.QueryRow(
		`SELECT COUNT(*) FROM cards WHERE deck_id = ? AND commander = 1`, deckID,
	).Scan(&cmdCount); err != nil {
		return v, err
	}
	if cmdCount == 0 {
		v.Issues = append(v.Issues, ValidationIssue{
			Code:    "no_commander",
			Message: "Nenhuma carta marcada como comandante",
		})
	}

	// Regra singleton: quantity > 1 para não-Basic Land
	sRows, err := r.db.Query(`
		SELECT id, name, quantity FROM cards
		WHERE deck_id = ? AND quantity > 1 AND (` + "`type`" + ` NOT LIKE '%Basic Land%' OR ` + "`type`" + ` IS NULL)
	`, deckID)
	if err != nil {
		return v, err
	}
	defer sRows.Close()
	for sRows.Next() {
		var cid int
		var cname string
		var qty int
		if sRows.Scan(&cid, &cname, &qty) == nil {
			v.Issues = append(v.Issues, ValidationIssue{
				Code:     "not_singleton",
				Message:  fmt.Sprintf("%s: %d cópias (permitido: 1)", cname, qty),
				CardID:   cid,
				CardName: cname,
			})
		}
	}

	// Identidade de cor
	if isCommander == 1 && deckColors != "" {
		allowed := map[string]bool{"C": true}
		for _, c := range strings.Split(deckColors, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				allowed[c] = true
			}
		}

		cRows, err := r.db.Query(`
			SELECT id, name, COALESCE(colors,'[]') FROM cards
			WHERE deck_id = ?
		`, deckID)
		if err == nil {
			defer cRows.Close()
			for cRows.Next() {
				var cid int
				var cname, colorsJSON string
				if cRows.Scan(&cid, &cname, &colorsJSON) != nil {
					continue
				}
				var codes []string
				if json.Unmarshal([]byte(colorsJSON), &codes) != nil {
					continue
				}
				for _, code := range codes {
					if !allowed[code] {
						v.Issues = append(v.Issues, ValidationIssue{
							Code:     "color_violation",
							Message:  fmt.Sprintf("%s: cor '%s' está fora da identidade do deck", cname, code),
							CardID:   cid,
							CardName: cname,
						})
						break
					}
				}
			}
		}
	}

	v.Valid = len(v.Issues) == 0
	return v, nil
}
