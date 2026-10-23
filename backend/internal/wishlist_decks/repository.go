package wishlist_decks

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List() ([]WishlistDeck, error) {
	rows, err := r.db.Query(`
		SELECT id, name, format, colors, COALESCE(note,''), acquired,
			DATE_FORMAT(created_at,'%Y-%m-%dT%H:%i:%s')
		FROM wishlist_decks ORDER BY acquired ASC, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]WishlistDeck, 0)
	for rows.Next() {
		var d WishlistDeck
		var acquired int
		if err := rows.Scan(&d.ID, &d.Name, &d.Format, &d.Colors, &d.Note, &acquired, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.Acquired = acquired == 1
		items = append(items, d)
	}
	return items, nil
}

func (r *Repository) Create(input CreateWishlistDeckInput) (*WishlistDeck, error) {
	res, err := r.db.Exec(`
		INSERT INTO wishlist_decks (name, format, colors, note)
		VALUES (?, ?, ?, ?)
	`, input.Name, input.Format, input.Colors, input.Note)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return r.getByID(id)
}

func (r *Repository) Acquire(id int64) (*WishlistDeck, error) {
	_, err := r.db.Exec(`UPDATE wishlist_decks SET acquired = 1 WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	return r.getByID(id)
}

func (r *Repository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM wishlist_decks WHERE id = ?`, id)
	return err
}

func (r *Repository) ListCards(deckID int64) ([]DeckCard, error) {
	rows, err := r.db.Query(`
		SELECT id, wishlist_deck_id, set_code, collection_number,
			COALESCE(name,''), COALESCE(printed_name,''), COALESCE(type_line,''),
			COALESCE(image_uri,''), foil, COALESCE(note,''), COALESCE(acquired,0),
			DATE_FORMAT(created_at,'%Y-%m-%dT%H:%i:%s')
		FROM wishlist_deck_cards WHERE wishlist_deck_id = ? ORDER BY id ASC
	`, deckID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cards := make([]DeckCard, 0)
	for rows.Next() {
		var c DeckCard
		var foil, acquired int
		if err := rows.Scan(&c.ID, &c.WishlistDeckID, &c.SetCode, &c.CollectionNumber,
			&c.Name, &c.PrintedName, &c.TypeLine, &c.ImageURI, &foil, &c.Note, &acquired, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.Foil = foil == 1
		c.Acquired = acquired == 1
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *Repository) ToggleCardAcquired(deckID, cardID int64) (*DeckCard, error) {
	_, err := r.db.Exec(`
		UPDATE wishlist_deck_cards SET acquired = NOT acquired
		WHERE id = ? AND wishlist_deck_id = ?
	`, cardID, deckID)
	if err != nil {
		return nil, err
	}
	var c DeckCard
	var foil, acquired int
	err = r.db.QueryRow(`
		SELECT id, wishlist_deck_id, set_code, collection_number,
			COALESCE(name,''), COALESCE(printed_name,''), COALESCE(type_line,''),
			COALESCE(image_uri,''), foil, COALESCE(note,''), COALESCE(acquired,0),
			DATE_FORMAT(created_at,'%Y-%m-%dT%H:%i:%s')
		FROM wishlist_deck_cards WHERE id = ?
	`, cardID).Scan(&c.ID, &c.WishlistDeckID, &c.SetCode, &c.CollectionNumber,
		&c.Name, &c.PrintedName, &c.TypeLine, &c.ImageURI, &foil, &c.Note, &acquired, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	c.Foil = foil == 1
	c.Acquired = acquired == 1
	return &c, nil
}

func (r *Repository) AddCard(deckID int64, input AddDeckCardInput, name, printedName, typeLine, imageURI string) (*DeckCard, error) {
	foil := 0
	if input.Foil {
		foil = 1
	}
	res, err := r.db.Exec(`
		INSERT INTO wishlist_deck_cards
			(wishlist_deck_id, set_code, collection_number, name, printed_name, type_line, image_uri, foil, note)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, deckID, input.SetCode, input.CollectionNumber, name, printedName, typeLine, imageURI, foil, input.Note)
	if err != nil {
		// fallback: type_line ainda não existe no banco
		res, err = r.db.Exec(`
			INSERT INTO wishlist_deck_cards
				(wishlist_deck_id, set_code, collection_number, name, printed_name, image_uri, foil, note)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, deckID, input.SetCode, input.CollectionNumber, name, printedName, imageURI, foil, input.Note)
		if err != nil {
			return nil, err
		}
	}
	id, _ := res.LastInsertId()
	var c DeckCard
	var acquiredInt int
	err = r.db.QueryRow(`
		SELECT id, wishlist_deck_id, set_code, collection_number,
			COALESCE(name,''), COALESCE(printed_name,''), COALESCE(type_line,''),
			COALESCE(image_uri,''), foil, COALESCE(note,''), COALESCE(acquired,0),
			DATE_FORMAT(created_at,'%Y-%m-%dT%H:%i:%s')
		FROM wishlist_deck_cards WHERE id = ?
	`, id).Scan(&c.ID, &c.WishlistDeckID, &c.SetCode, &c.CollectionNumber,
		&c.Name, &c.PrintedName, &c.TypeLine, &c.ImageURI, &foil, &c.Note, &acquiredInt, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	c.Foil = foil == 1
	c.Acquired = acquiredInt == 1
	return &c, nil
}

func (r *Repository) DeleteCard(deckID, cardID int64) error {
	_, err := r.db.Exec(`DELETE FROM wishlist_deck_cards WHERE id = ? AND wishlist_deck_id = ?`, cardID, deckID)
	return err
}

func (r *Repository) getByID(id int64) (*WishlistDeck, error) {
	var d WishlistDeck
	var acquired int
	err := r.db.QueryRow(`
		SELECT id, name, format, colors, COALESCE(note,''), acquired,
			DATE_FORMAT(created_at,'%Y-%m-%dT%H:%i:%s')
		FROM wishlist_decks WHERE id = ?
	`, id).Scan(&d.ID, &d.Name, &d.Format, &d.Colors, &d.Note, &acquired, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	d.Acquired = acquired == 1
	return &d, nil
}
