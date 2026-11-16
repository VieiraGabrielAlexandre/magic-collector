package trade

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// movementBase selects cards that have at least 2 price history entries,
// comparing oldest vs most-recent price.
const movementBase = `
SELECT
    c.id, c.name, c.set_code, c.rarity, c.foil, c.image_url, c.quantity,
    latest.price_usd  AS current_price,
    earliest.price_usd AS prev_price,
    ROUND(((latest.price_usd - earliest.price_usd) / earliest.price_usd) * 100, 2) AS change_pct
FROM cards c
JOIN (
    SELECT h.card_id, h.price_usd
    FROM card_price_history h
    INNER JOIN (SELECT card_id, MAX(recorded_at) AS d FROM card_price_history GROUP BY card_id) mx
        ON h.card_id = mx.card_id AND h.recorded_at = mx.d
) latest ON latest.card_id = c.id
JOIN (
    SELECT h.card_id, h.price_usd
    FROM card_price_history h
    INNER JOIN (SELECT card_id, MIN(recorded_at) AS d FROM card_price_history GROUP BY card_id) mn
        ON h.card_id = mn.card_id AND h.recorded_at = mn.d
) earliest ON earliest.card_id = c.id
JOIN (
    SELECT card_id FROM card_price_history GROUP BY card_id HAVING COUNT(*) >= 2
) has2 ON has2.card_id = c.id
WHERE latest.price_usd > 0.5
AND earliest.price_usd > 0
`

func (r *Repository) scanMovements(rows *sql.Rows) ([]CardMovement, error) {
	var out []CardMovement
	for rows.Next() {
		var m CardMovement
		var foilInt int
		if err := rows.Scan(
			&m.CardID, &m.Name, &m.SetCode, &m.Rarity, &foilInt,
			&m.ImageURL, &m.Quantity, &m.CurrentPrice, &m.PrevPrice, &m.ChangePct,
		); err != nil {
			return nil, err
		}
		m.Foil = foilInt == 1
		if m.ChangePct >= 0 {
			m.Direction = "up"
		} else {
			m.Direction = "down"
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) Gainers(minPct float64, limit int) ([]CardMovement, error) {
	rows, err := r.db.Query(movementBase+
		` HAVING change_pct >= ? ORDER BY change_pct DESC LIMIT ?`, minPct, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanMovements(rows)
}

func (r *Repository) Losers(maxPct float64, limit int) ([]CardMovement, error) {
	rows, err := r.db.Query(movementBase+
		` HAVING change_pct <= ? ORDER BY change_pct ASC LIMIT ?`, maxPct, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanMovements(rows)
}

func (r *Repository) Stable(limit int) ([]CardMovement, error) {
	rows, err := r.db.Query(movementBase+
		` HAVING ABS(change_pct) < 5 ORDER BY latest.price_usd DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanMovements(rows)
}

func (r *Repository) TradeSuggestions(minPrice float64, limit int) ([]TradeSuggestion, error) {
	q := `
	SELECT
	    c.id, c.name, c.set_code, c.rarity, c.foil, c.image_url, c.quantity,
	    latest.price_usd AS current_price,
	    ROUND(latest.price_usd * 0.88, 2) AS suggested_value,
	    COALESCE(ROUND(((latest.price_usd - earliest.price_usd) / NULLIF(earliest.price_usd, 0)) * 100, 2), 0) AS change_pct,
	    CASE WHEN latest.price_usd >= maxp.max_price THEN 1 ELSE 0 END AS at_peak
	FROM cards c
	JOIN (
	    SELECT h.card_id, h.price_usd
	    FROM card_price_history h
	    INNER JOIN (SELECT card_id, MAX(recorded_at) AS d FROM card_price_history GROUP BY card_id) mx
	        ON h.card_id = mx.card_id AND h.recorded_at = mx.d
	) latest ON latest.card_id = c.id
	LEFT JOIN (
	    SELECT h.card_id, h.price_usd
	    FROM card_price_history h
	    INNER JOIN (SELECT card_id, MIN(recorded_at) AS d FROM card_price_history GROUP BY card_id) mn
	        ON h.card_id = mn.card_id AND h.recorded_at = mn.d
	) earliest ON earliest.card_id = c.id
	JOIN (
	    SELECT card_id, MAX(price_usd) AS max_price FROM card_price_history GROUP BY card_id
	) maxp ON maxp.card_id = c.id
	WHERE latest.price_usd >= ?
	ORDER BY latest.price_usd DESC
	LIMIT ?
	`
	rows, err := r.db.Query(q, minPrice, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TradeSuggestion
	for rows.Next() {
		var s TradeSuggestion
		var foilInt, atPeakInt int
		if err := rows.Scan(
			&s.CardID, &s.Name, &s.SetCode, &s.Rarity, &foilInt,
			&s.ImageURL, &s.Quantity, &s.CurrentPrice, &s.SuggestedValue,
			&s.ChangePct, &atPeakInt,
		); err != nil {
			return nil, err
		}
		s.Foil = foilInt == 1
		s.AtPeak = atPeakInt == 1
		out = append(out, s)
	}
	return out, rows.Err()
}

// PriceSpikes retorna cartas cuja última entrada de preço subiu >= threshold%
// em relação à entrada imediatamente anterior (ex: ontem vs hoje após refresh).
func (r *Repository) PriceSpikes(threshold float64) ([]PriceSpike, error) {
	q := `
	SELECT
	    c.id, c.name, c.set_code, c.rarity, c.foil, COALESCE(c.image_url,''),
	    c.quantity,
	    latest.price_usd  AS current_price,
	    prev.price_usd    AS prev_price,
	    ROUND(((latest.price_usd - prev.price_usd) / prev.price_usd) * 100, 2) AS change_pct,
	    DATE_FORMAT(latest.recorded_at, '%Y-%m-%dT%H:%i:%s') AS updated_at
	FROM cards c
	JOIN (
	    SELECT h.card_id, h.price_usd, h.recorded_at
	    FROM card_price_history h
	    INNER JOIN (SELECT card_id, MAX(recorded_at) AS d FROM card_price_history GROUP BY card_id) mx
	        ON h.card_id = mx.card_id AND h.recorded_at = mx.d
	) latest ON latest.card_id = c.id
	JOIN (
	    SELECT h.card_id, h.price_usd
	    FROM card_price_history h
	    INNER JOIN (
	        SELECT ph2.card_id, MAX(ph2.recorded_at) AS d
	        FROM card_price_history ph2
	        JOIN (SELECT card_id, MAX(recorded_at) AS max_d FROM card_price_history GROUP BY card_id) lx
	            ON ph2.card_id = lx.card_id AND ph2.recorded_at < lx.max_d
	        GROUP BY ph2.card_id
	    ) prev_max ON h.card_id = prev_max.card_id AND h.recorded_at = prev_max.d
	) prev ON prev.card_id = c.id
	WHERE latest.price_usd > 0.5
	  AND prev.price_usd  > 0
	HAVING change_pct >= ?
	ORDER BY change_pct DESC
	LIMIT 50
	`
	rows, err := r.db.Query(q, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PriceSpike
	for rows.Next() {
		var s PriceSpike
		var foilInt int
		if err := rows.Scan(
			&s.CardID, &s.Name, &s.SetCode, &s.Rarity, &foilInt,
			&s.ImageURL, &s.Quantity, &s.CurrentPrice, &s.PrevPrice, &s.ChangePct, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		s.Foil = foilInt == 1
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) PortfolioHistory() ([]PortfolioPoint, error) {
	rows, err := r.db.Query(`
	    SELECT DATE_FORMAT(h.recorded_at, '%Y-%m-%d') AS date,
	           ROUND(SUM(h.price_usd * c.quantity), 2) AS total_value_usd
	    FROM card_price_history h
	    JOIN cards c ON c.id = h.card_id
	    GROUP BY h.recorded_at
	    ORDER BY h.recorded_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PortfolioPoint
	for rows.Next() {
		var p PortfolioPoint
		if err := rows.Scan(&p.Date, &p.TotalValueUSD); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
