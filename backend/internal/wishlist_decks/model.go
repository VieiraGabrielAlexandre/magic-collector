package wishlist_decks

type WishlistDeck struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Format        string `json:"format"`
	Colors        string `json:"colors"`
	Note          string `json:"note"`
	Acquired      bool   `json:"acquired"`
	LinkedDeckID  int64  `json:"linked_deck_id"`
	CreatedAt     string `json:"created_at"`
}

type AcquireCardResult struct {
	DeckID int64 `json:"deck_id"`
	CardID int64 `json:"card_id"`
}

type CreateWishlistDeckInput struct {
	Name   string `json:"name"`
	Format string `json:"format"`
	Colors string `json:"colors"`
	Note   string `json:"note"`
}

type DeckCard struct {
	ID               int64  `json:"id"`
	WishlistDeckID   int64  `json:"wishlist_deck_id"`
	SetCode          string `json:"set_code"`
	CollectionNumber string `json:"collection_number"`
	Name             string `json:"name"`
	PrintedName      string `json:"printed_name"`
	TypeLine         string `json:"type_line"`
	ImageURI         string `json:"image_uri"`
	Foil             bool   `json:"foil"`
	Note             string `json:"note"`
	Acquired         bool   `json:"acquired"`
	CreatedAt        string `json:"created_at"`
}

type AddDeckCardInput struct {
	SetCode          string `json:"set_code"`
	CollectionNumber string `json:"collection_number"`
	Foil             bool   `json:"foil"`
	Note             string `json:"note"`
}
