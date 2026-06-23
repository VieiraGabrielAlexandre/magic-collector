package decks

type CommanderCard struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

type Deck struct {
	ID           int             `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Commander    bool            `json:"commander"`
	Colors       string          `json:"colors"`
	SetCode      string          `json:"set_code"`
	IconURI      string          `json:"icon_uri"`
	ThemeColor   string          `json:"theme_color"`
	CardCount    int             `json:"card_count"`
	Evaluation   string          `json:"evaluation,omitempty"`
	EvaluatedAt  string          `json:"evaluated_at,omitempty"`
	BattleWins   int             `json:"battle_wins"`
	BattleLosses int             `json:"battle_losses"`
	BattleDraws  int             `json:"battle_draws"`
	BattleTotal  int             `json:"battle_total"`
	Commanders   []CommanderCard `json:"commanders,omitempty"`
}

type DeckInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Commander   bool   `json:"commander"`
	Colors      string `json:"colors"`
	SetCode     string `json:"set_code"`
	ThemeColor  string `json:"theme_color"`
}
