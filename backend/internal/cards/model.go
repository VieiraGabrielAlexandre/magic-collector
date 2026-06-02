package cards

type Card struct {
	ID               int    `json:"id"`
	MTGID            string `json:"mtg_id"`
	Name             string `json:"name"`
	Color            string `json:"color"`
	Type             string `json:"type"`
	Subtitle         string `json:"subtitle"`
	CollectionNumber string `json:"collection_number"`
	Rarity           string `json:"rarity"`
	SetCode          string `json:"set_code"`
	ManaCost         string `json:"mana_cost"`
	Colors           string `json:"colors"`
	Language         string `json:"language"`
	Year             int    `json:"year"`
	Artist           string `json:"artist"`
	Company          string `json:"company"`
	Foil             bool   `json:"foil"`
	PreRelease       bool    `json:"prerelease"`
	Commander        bool    `json:"commander"`
	PreconDeck       string  `json:"precon_deck"`
	DeckID           int     `json:"deck_id"`
	Quantity         int     `json:"quantity"`
	Condition        string  `json:"condition"`
	Notes            string  `json:"notes"`
	PriceUSD         float64 `json:"price_usd"`
	ImageURL         string  `json:"image_url"`
}

type PreviewCardInput struct {
	Name             string `json:"name"`
	SetCode          string `json:"set_code"`
	CollectionNumber string `json:"collection_number"`
	Language         string `json:"language"`
	Artist           string `json:"artist"`
	PreRelease       bool   `json:"prerelease"`
	Foil             bool   `json:"foil"`
}

type CreateCardInput struct {
	Name             string `json:"name" binding:"required"`
	Color            string `json:"color"`
	Colors           string `json:"colors"`
	Type             string `json:"type"`
	Subtitle         string `json:"subtitle"`
	CollectionNumber string `json:"collection_number"`
	Rarity           string `json:"rarity"`
	SetCode          string `json:"set_code"`
	Language         string `json:"language"`
	Year             int    `json:"year"`
	Artist           string `json:"artist"`
	Company          string `json:"company"`
	Foil             bool   `json:"foil"`
	PreRelease       bool   `json:"prerelease"`
	Commander        bool   `json:"commander"`
	PreconDeck       string `json:"precon_deck"`
	DeckID           int    `json:"deck_id"`
	Quantity         int    `json:"quantity"`
	Condition        string `json:"condition"`
	Notes            string `json:"notes"`
}

type UpdateCardInput struct {
	Name             string `json:"name" binding:"required"`
	Color            string `json:"color"`
	Colors           string `json:"colors"`
	Type             string `json:"type"`
	Subtitle         string `json:"subtitle"`
	CollectionNumber string `json:"collection_number"`
	Rarity           string `json:"rarity"`
	SetCode          string `json:"set_code"`
	Language         string `json:"language"`
	Year             int    `json:"year"`
	Artist           string `json:"artist"`
	Company          string `json:"company"`
	Foil             bool   `json:"foil"`
	PreRelease       bool   `json:"prerelease"`
	Commander        bool   `json:"commander"`
	PreconDeck       string `json:"precon_deck"`
	DeckID           int    `json:"deck_id"`
	Quantity         int    `json:"quantity"`
	Condition        string `json:"condition"`
	Notes            string `json:"notes"`
	Propagate        bool   `json:"propagate"`
}
