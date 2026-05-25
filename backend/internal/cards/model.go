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
	Quantity         int    `json:"quantity"`
	Condition        string `json:"condition"`
	Notes            string `json:"notes"`
}

type CreateCardInput struct {
	Name             string `json:"name" binding:"required"`
	Color            string `json:"color"`
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
	Quantity         int    `json:"quantity"`
	Condition        string `json:"condition"`
	Notes            string `json:"notes"`
}
