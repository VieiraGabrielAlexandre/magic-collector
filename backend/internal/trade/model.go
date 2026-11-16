package trade

type CardMovement struct {
	CardID       int     `json:"card_id"`
	Name         string  `json:"name"`
	SetCode      string  `json:"set_code"`
	Rarity       string  `json:"rarity"`
	Foil         bool    `json:"foil"`
	CurrentPrice float64 `json:"current_price"`
	PrevPrice    float64 `json:"prev_price"`
	ChangePct    float64 `json:"change_pct"`
	Direction    string  `json:"direction"`
	ImageURL     string  `json:"image_url"`
	Quantity     int     `json:"quantity"`
}

type TradeSuggestion struct {
	CardID         int     `json:"card_id"`
	Name           string  `json:"name"`
	SetCode        string  `json:"set_code"`
	Rarity         string  `json:"rarity"`
	Foil           bool    `json:"foil"`
	CurrentPrice   float64 `json:"current_price"`
	SuggestedValue float64 `json:"suggested_value"`
	ImageURL       string  `json:"image_url"`
	Quantity       int     `json:"quantity"`
	ChangePct      float64 `json:"change_pct"`
	AtPeak         bool    `json:"at_peak"`
}

type PortfolioPoint struct {
	Date          string  `json:"date"`
	TotalValueUSD float64 `json:"total_value_usd"`
}

type PriceSpike struct {
	CardID       int     `json:"card_id"`
	Name         string  `json:"name"`
	SetCode      string  `json:"set_code"`
	Rarity       string  `json:"rarity"`
	Foil         bool    `json:"foil"`
	ImageURL     string  `json:"image_url"`
	Quantity     int     `json:"quantity"`
	CurrentPrice float64 `json:"current_price"`
	PrevPrice    float64 `json:"prev_price"`
	ChangePct    float64 `json:"change_pct"`
	UpdatedAt    string  `json:"updated_at"`
}
