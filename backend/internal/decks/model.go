package decks

type Deck struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Commander   bool   `json:"commander"`
	Colors      string `json:"colors"`
	SetCode     string `json:"set_code"`
	IconURI     string `json:"icon_uri"`
	CardCount   int    `json:"card_count"`
}

type DeckInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Commander   bool   `json:"commander"`
	Colors      string `json:"colors"`
	SetCode     string `json:"set_code"`
}
