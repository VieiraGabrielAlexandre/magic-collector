package reports

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Build() (CollectionReport, error) {
	report := CollectionReport{
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		SchemaVersion: "1.0",
	}

	var err error

	if report.Summary, err = r.querySummary(); err != nil {
		return report, err
	}
	if report.ByColor, err = r.queryByColor(); err != nil {
		return report, err
	}
	if report.TopExpensive, err = r.queryTopExpensive(10); err != nil {
		return report, err
	}
	if report.RecentDecks, err = r.queryRecentDecks(5); err != nil {
		return report, err
	}
	if report.RecentCards, err = r.queryRecentCards(10); err != nil {
		return report, err
	}
	if report.PriceAlerts, err = r.queryPriceAlerts(10); err != nil {
		return report, err
	}

	report.Message = FormatWhatsApp(report)

	return report, nil
}

func (r *Repository) querySummary() (Summary, error) {
	s := Summary{RarityBreakdown: map[string]int{}}

	// Aggregate: total copies, unique cards, total value, foil count
	err := r.db.QueryRow(`
		SELECT
			COALESCE(SUM(quantity), 0),
			COUNT(*),
			COALESCE(SUM(price_usd * quantity), 0),
			SUM(foil)
		FROM cards
	`).Scan(&s.TotalCards, &s.TotalUnique, &s.TotalValueUSD, &s.FoilCount)
	if err != nil {
		return s, err
	}

	// Without deck / in deck
	err = r.db.QueryRow(`SELECT COUNT(*) FROM cards WHERE deck_id = 0 OR deck_id IS NULL`).Scan(&s.CardsWithoutDeck)
	if err != nil {
		return s, err
	}
	err = r.db.QueryRow(`SELECT COUNT(*) FROM cards WHERE deck_id > 0`).Scan(&s.CardsInDeck)
	if err != nil {
		return s, err
	}

	// Total decks
	err = r.db.QueryRow(`SELECT COUNT(*) FROM decks`).Scan(&s.TotalDecks)
	if err != nil {
		return s, err
	}

	// Rarity breakdown
	rows, err := r.db.Query(`
		SELECT UPPER(COALESCE(NULLIF(rarity,''), '?')) AS r, COUNT(*)
		FROM cards
		GROUP BY r
	`)
	if err != nil {
		return s, err
	}
	defer rows.Close()
	for rows.Next() {
		var r string
		var cnt int
		if err := rows.Scan(&r, &cnt); err != nil {
			continue
		}
		s.RarityBreakdown[r] = cnt
	}

	return s, nil
}

var colorNames = map[string]string{
	"W":     "Branco",
	"U":     "Azul",
	"B":     "Preto",
	"R":     "Vermelho",
	"G":     "Verde",
	"C":     "Incolor",
	"Multi": "Multicolorida",
}

func (r *Repository) queryByColor() ([]ColorBucket, error) {
	rows, err := r.db.Query(`
		SELECT COALESCE(colors, '[]'), COUNT(*), COALESCE(SUM(quantity), 0)
		FROM cards
		WHERE deck_id = 0 OR deck_id IS NULL
		GROUP BY colors
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buckets := map[string]*ColorBucket{}

	for rows.Next() {
		var colorsJSON string
		var cnt, qty int
		if err := rows.Scan(&colorsJSON, &cnt, &qty); err != nil {
			continue
		}

		var codes []string
		if err := json.Unmarshal([]byte(colorsJSON), &codes); err != nil || len(codes) == 0 {
			codes = []string{"C"}
		}

		key := "C"
		if len(codes) > 1 {
			key = "Multi"
		} else if len(codes) == 1 {
			key = codes[0]
		}

		if b, ok := buckets[key]; ok {
			b.Count += cnt
			b.Quantity += qty
		} else {
			buckets[key] = &ColorBucket{
				Color:   key,
				ColorPT: colorNames[key],
				Count:   cnt,
				Quantity: qty,
			}
		}
	}

	order := []string{"W", "U", "B", "R", "G", "Multi", "C"}
	result := make([]ColorBucket, 0, len(buckets))
	for _, k := range order {
		if b, ok := buckets[k]; ok {
			result = append(result, *b)
		}
	}
	return result, nil
}

func (r *Repository) queryTopExpensive(limit int) ([]ExpensiveCard, error) {
	rows, err := r.db.Query(`
		SELECT c.id, c.name, COALESCE(c.set_code,''), COALESCE(c.rarity,''),
		       c.price_usd, c.foil, COALESCE(c.image_url,''),
		       COALESCE(c.deck_id,0), COALESCE(d.name,'')
		FROM cards c
		LEFT JOIN decks d ON d.id = c.deck_id AND c.deck_id > 0
		WHERE c.price_usd > 0
		ORDER BY c.price_usd DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ExpensiveCard
	for rows.Next() {
		var e ExpensiveCard
		var foilInt int
		if err := rows.Scan(&e.ID, &e.Name, &e.SetCode, &e.Rarity,
			&e.PriceUSD, &foilInt, &e.ImageURL, &e.DeckID, &e.DeckName); err != nil {
			continue
		}
		e.Foil = foilInt == 1
		result = append(result, e)
	}
	return result, nil
}

func (r *Repository) queryRecentDecks(limit int) ([]RecentDeck, error) {
	rows, err := r.db.Query(`
		SELECT d.id, d.name, COALESCE(d.colors,''),
		       COUNT(c.id),
		       COALESCE(SUM(c.price_usd * c.quantity), 0),
		       DATE_FORMAT(d.created_at, '%Y-%m-%dT%H:%i:%s')
		FROM decks d
		LEFT JOIN cards c ON c.deck_id = d.id
		GROUP BY d.id, d.name, d.colors, d.created_at
		ORDER BY d.created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []RecentDeck
	for rows.Next() {
		var d RecentDeck
		if err := rows.Scan(&d.ID, &d.Name, &d.Colors,
			&d.CardCount, &d.EstimatedValue, &d.CreatedAt); err != nil {
			continue
		}
		result = append(result, d)
	}
	return result, nil
}

func (r *Repository) queryPriceAlerts(limit int) ([]PriceAlert, error) {
	// Compara as duas entradas mais recentes do histórico de preço de cada carta.
	// Retorna somente cartas com variação absoluta >= 15%.
	rows, err := r.db.Query(`
		SELECT c.id, c.name, COALESCE(c.set_code,''), COALESCE(c.rarity,''), c.foil,
		       h1.price_usd AS cur_price, h2.price_usd AS prev_price,
		       ROUND((h1.price_usd - h2.price_usd) / h2.price_usd * 100, 1) AS change_pct
		FROM cards c
		INNER JOIN card_price_history h1 ON h1.card_id = c.id
		  AND h1.recorded_at = (SELECT MAX(hh.recorded_at) FROM card_price_history hh WHERE hh.card_id = c.id)
		INNER JOIN card_price_history h2 ON h2.card_id = c.id
		  AND h2.recorded_at = (
		      SELECT MAX(hh.recorded_at) FROM card_price_history hh
		      WHERE hh.card_id = c.id AND hh.recorded_at < h1.recorded_at
		  )
		WHERE h1.price_usd > 0.5 AND h2.price_usd > 0
		  AND ABS((h1.price_usd - h2.price_usd) / h2.price_usd * 100) >= 15
		ORDER BY ABS((h1.price_usd - h2.price_usd) / h2.price_usd * 100) DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PriceAlert
	for rows.Next() {
		var a PriceAlert
		var foilInt int
		if err := rows.Scan(&a.ID, &a.Name, &a.SetCode, &a.Rarity, &foilInt,
			&a.CurrentPrice, &a.PrevPrice, &a.ChangePct); err != nil {
			continue
		}
		a.Foil = foilInt == 1
		if a.ChangePct >= 0 {
			a.Direction = "up"
		} else {
			a.Direction = "down"
		}
		result = append(result, a)
	}
	return result, nil
}

func (r *Repository) queryRecentCards(limit int) ([]RecentCard, error) {
	rows, err := r.db.Query(`
		SELECT id, name, COALESCE(set_code,''), COALESCE(rarity,''),
		       foil, COALESCE(image_url,''),
		       DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%s')
		FROM cards
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []RecentCard
	for rows.Next() {
		var c RecentCard
		var foilInt int
		if err := rows.Scan(&c.ID, &c.Name, &c.SetCode, &c.Rarity,
			&foilInt, &c.ImageURL, &c.AddedAt); err != nil {
			continue
		}
		c.Foil = foilInt == 1
		result = append(result, c)
	}
	return result, nil
}
