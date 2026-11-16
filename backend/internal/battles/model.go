package battles

type DeckStat struct {
	DeckName string  `json:"deck_name"`
	Total    int     `json:"total"`
	Wins     int     `json:"wins"`
	WinRate  float64 `json:"win_rate"`
}

type MonthStat struct {
	Month   string `json:"month"` // "YYYY-MM"
	Total   int    `json:"total"`
	Wins    int    `json:"wins"`
}

type OpponentStat struct {
	Name   string `json:"name"`
	Losses int    `json:"losses"` // vezes que esse adversário perdeu para mim
}

type BattleStats struct {
	Total     int            `json:"total"`
	Wins      int            `json:"wins"`
	Losses    int            `json:"losses"`
	Draws     int            `json:"draws"`
	WinRate   float64        `json:"win_rate"`
	ByDeck    []DeckStat     `json:"by_deck"`
	ByMonth   []MonthStat    `json:"by_month"`
	TopOpponents []OpponentStat `json:"top_opponents"`
}

type Battle struct {
	ID          int      `json:"id"`
	Result      string   `json:"result"` // "win" | "loss"
	Opponents   []string `json:"opponents"`
	PlayerCount int      `json:"player_count"`
	GameStyle   string   `json:"game_style"`
	DeckID      int      `json:"deck_id"`
	DeckName    string   `json:"deck_name"`
	DeckIsMine  bool     `json:"deck_is_mine"`
	Notes       string   `json:"notes"`
	PlayedAt    string   `json:"played_at"`
}

type BattleInput struct {
	Result      string   `json:"result" binding:"required"`
	Opponents   []string `json:"opponents"`
	PlayerCount int      `json:"player_count"`
	GameStyle   string   `json:"game_style"`
	DeckID      int      `json:"deck_id"`
	DeckName    string   `json:"deck_name"`
	DeckIsMine  bool     `json:"deck_is_mine"`
	Notes       string   `json:"notes"`
}
