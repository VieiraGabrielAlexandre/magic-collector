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
	BattleTotal       int             `json:"battle_total"`
	EstimatedValueUSD float64         `json:"estimated_value_usd"`
	DeckboxColor      string          `json:"deckbox_color"`
	DeckOrigin        string          `json:"deck_origin"` // custom | precon | precon_modified
	Commanders        []CommanderCard `json:"commanders,omitempty"`
}

type DeckInput struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	Commander    bool   `json:"commander"`
	Colors       string `json:"colors"`
	SetCode      string `json:"set_code"`
	ThemeColor   string `json:"theme_color"`
	DeckboxColor string `json:"deckbox_color"`
	DeckOrigin   string `json:"deck_origin"`
}

// ValidationIssue descreve um problema encontrado durante a validação do deck.
type ValidationIssue struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	CardID   int    `json:"card_id,omitempty"`
	CardName string `json:"card_name,omitempty"`
}

// DeckValidation é o resultado de GET /decks/:id/validate.
type DeckValidation struct {
	DeckID    int               `json:"deck_id"`
	Valid     bool              `json:"valid"`
	CardCount int               `json:"card_count"`
	Issues    []ValidationIssue `json:"issues"`
}
