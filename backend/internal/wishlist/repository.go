package wishlist

import (
	"database/sql"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List() ([]WishlistCard, error) {
	rows, err := r.db.Query(`
		SELECT id, mtg_id, set_code, collection_number, name, printed_name,
		       image_uri, artist, rarity, COALESCE(colors,'[]'), COALESCE(color,''),
		       COALESCE(price_usd,0), COALESCE(price_usd_foil,0),
		       foil, COALESCE(reason,''), acquired,
		       DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%s')
		FROM wishlist_cards
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]WishlistCard, 0)
	for rows.Next() {
		var w WishlistCard
		var foil, acquired int
		if err := rows.Scan(
			&w.ID, &w.MTGID, &w.SetCode, &w.CollectionNumber, &w.Name, &w.PrintedName,
			&w.ImageURI, &w.Artist, &w.Rarity, &w.Colors, &w.Color,
			&w.PriceUSD, &w.PriceUSDFoil,
			&foil, &w.Reason, &acquired, &w.CreatedAt,
		); err != nil {
			return nil, err
		}
		w.Foil = foil == 1
		w.Acquired = acquired == 1
		items = append(items, w)
	}
	return items, nil
}

func (r *Repository) GetByID(id string) (*WishlistCard, error) {
	var w WishlistCard
	var foil, acquired int
	err := r.db.QueryRow(`
		SELECT id, mtg_id, set_code, collection_number, name, printed_name,
		       image_uri, artist, rarity, COALESCE(colors,'[]'), COALESCE(color,''),
		       COALESCE(price_usd,0), COALESCE(price_usd_foil,0),
		       foil, COALESCE(reason,''), acquired,
		       DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%s')
		FROM wishlist_cards WHERE id = ?
	`, id).Scan(
		&w.ID, &w.MTGID, &w.SetCode, &w.CollectionNumber, &w.Name, &w.PrintedName,
		&w.ImageURI, &w.Artist, &w.Rarity, &w.Colors, &w.Color,
		&w.PriceUSD, &w.PriceUSDFoil,
		&foil, &w.Reason, &acquired, &w.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	w.Foil = foil == 1
	w.Acquired = acquired == 1
	return &w, nil
}

func (r *Repository) Create(w WishlistCard) (int64, error) {
	now := time.Now().Format("2006-01-02 15:04:05")

	var priceUSD, priceUSDFoil interface{}
	if w.PriceUSD > 0 {
		priceUSD = w.PriceUSD
	}
	if w.PriceUSDFoil > 0 {
		priceUSDFoil = w.PriceUSDFoil
	}

	var foil int
	if w.Foil {
		foil = 1
	}

	result, err := r.db.Exec(`
		INSERT INTO wishlist_cards
		  (mtg_id, set_code, collection_number, name, printed_name,
		   image_uri, artist, rarity, colors, color,
		   price_usd, price_usd_foil, foil, reason, acquired, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)
	`,
		w.MTGID, w.SetCode, w.CollectionNumber, w.Name, w.PrintedName,
		w.ImageURI, w.Artist, w.Rarity, w.Colors, w.Color,
		priceUSD, priceUSDFoil, foil, w.Reason, now, now,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(id string, input WishlistUpdateInput) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	var foil int
	if input.Foil {
		foil = 1
	}
	_, err := r.db.Exec(
		`UPDATE wishlist_cards SET foil = ?, reason = ?, updated_at = ? WHERE id = ?`,
		foil, input.Reason, now, id,
	)
	return err
}

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM wishlist_cards WHERE id = ?`, id)
	return err
}

func (r *Repository) Acquire(id string, input AcquireInput) (int64, error) {
	w, err := r.GetByID(id)
	if err != nil {
		return 0, err
	}

	condition := input.Condition
	if condition == "" {
		condition = "near_mint"
	}

	colors := w.Colors
	if colors == "" {
		colors = "[]"
	}
	colorDisplay := colorsJSONToDisplay(colors)

	name := w.Name
	if name == "" {
		name = fmt.Sprintf("%s #%s", w.SetCode, w.CollectionNumber)
	}

	var foil, commander, prerelease int
	if w.Foil {
		foil = 1
	}
	if input.Commander {
		commander = 1
	}
	if input.PreRelease {
		prerelease = 1
	}

	priceUSD := w.PriceUSD
	if w.Foil && w.PriceUSDFoil > 0 {
		priceUSD = w.PriceUSDFoil
	}

	insertSQL := "INSERT INTO cards " +
		"(mtg_id, name, set_code, collection_number, artist, rarity, " +
		"colors, color, foil, commander, prerelease, deck_id, " +
		"`condition`, notes, price_usd, image_url, quantity, language, " +
		"`type`, mana_cost, subtitle, company, year, precon_deck) " +
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', ?, ?, 1, 'EN', '', '', '', 'Wizards of the Coast', 0, '')"

	result, err := r.db.Exec(insertSQL,
		w.MTGID, name, w.SetCode, w.CollectionNumber, w.Artist, w.Rarity,
		colors, colorDisplay, foil, commander, prerelease, input.DeckID,
		condition, priceUSD, w.ImageURI,
	)
	if err != nil {
		return 0, err
	}

	newCardID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	_, _ = r.db.Exec(`UPDATE wishlist_cards SET acquired = 1, updated_at = ? WHERE id = ?`, now, id)

	return newCardID, nil
}
