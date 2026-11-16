package reports

// CollectionReport is the full response payload of GET /reports/collection.
type CollectionReport struct {
	GeneratedAt   string         `json:"generated_at"`    // RFC3339
	SchemaVersion string         `json:"schema_version"`  // "1.0"
	Summary       Summary        `json:"summary"`
	ByColor       []ColorBucket  `json:"cards_without_deck_by_color"`
	TopExpensive  []ExpensiveCard `json:"top_expensive_cards"`
	RecentDecks   []RecentDeck   `json:"recent_decks"`
	RecentCards   []RecentCard   `json:"recent_cards"`
	PriceAlerts   []PriceAlert   `json:"price_alerts"`
	Message       string         `json:"message"` // ready-to-send WhatsApp text (consumed by ArcanePulse)
}

// PriceAlert sinaliza uma carta com variação de preço >= 15% entre as duas
// últimas entradas do histórico.
type PriceAlert struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	SetCode      string  `json:"set_code"`
	Rarity       string  `json:"rarity"`
	Foil         bool    `json:"foil"`
	CurrentPrice float64 `json:"current_price_usd"`
	PrevPrice    float64 `json:"prev_price_usd"`
	ChangePct    float64 `json:"change_pct"`    // positivo = alta, negativo = queda
	Direction    string  `json:"direction"`     // "up" | "down"
}

// Summary holds aggregate numbers for the entire collection.
type Summary struct {
	TotalCards       int              `json:"total_cards"`        // SUM(quantity)
	TotalUnique      int              `json:"total_unique_cards"` // COUNT(*)
	TotalDecks       int              `json:"total_decks"`
	TotalValueUSD    float64          `json:"total_value_usd"`
	CardsWithoutDeck int              `json:"cards_without_deck"` // unique cards with deck_id = 0
	CardsInDeck      int              `json:"cards_in_deck"`      // unique cards with deck_id > 0
	FoilCount        int              `json:"foil_count"`
	RarityBreakdown  map[string]int   `json:"rarity_breakdown"`   // M/R/U/C/L/T → count
}

// ColorBucket groups cards without a deck by their primary color.
type ColorBucket struct {
	Color    string `json:"color"`     // WUBRG / Multi / C
	ColorPT  string `json:"color_pt"`  // human-readable in PT-BR
	Count    int    `json:"count"`     // unique cards
	Quantity int    `json:"quantity"`  // total copies
}

// ExpensiveCard is a card entry in the top-expensive list.
type ExpensiveCard struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	SetCode  string  `json:"set_code"`
	Rarity   string  `json:"rarity"`
	PriceUSD float64 `json:"price_usd"`
	Foil     bool    `json:"foil"`
	ImageURL string  `json:"image_url"`
	DeckID   int     `json:"deck_id"`
	DeckName string  `json:"deck_name,omitempty"`
}

// RecentDeck is a deck entry in the recent-decks list.
type RecentDeck struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Colors         string  `json:"colors"`
	CardCount      int     `json:"card_count"`
	EstimatedValue float64 `json:"estimated_value_usd"`
	CreatedAt      string  `json:"created_at"`
}

// RecentCard is a card entry in the recently-added list.
type RecentCard struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	SetCode  string `json:"set_code"`
	Rarity   string `json:"rarity"`
	Foil     bool   `json:"foil"`
	ImageURL string `json:"image_url"`
	AddedAt  string `json:"added_at"`
}
