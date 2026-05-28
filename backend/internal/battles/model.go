package battles

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
