package gamechangers

type GameChanger struct {
	ID         int    `json:"id"`
	ScryfallID string `json:"scryfall_id"`
	Name       string `json:"name"`
	SetCode    string `json:"set_code"`
	Rarity     string `json:"rarity"`
	Colors     string `json:"colors"`
	ImageURL   string `json:"image_url"`
	ManaCost   string `json:"mana_cost"`
	TypeLine   string `json:"type_line"`
	SyncedAt   string `json:"synced_at"`
	OwnedCount int    `json:"owned_count"`
}

type SyncResult struct {
	Inserted int `json:"inserted"`
	Updated  int `json:"updated"`
	Total    int `json:"total"`
}
