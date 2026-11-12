package storage

type StorageLocation struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // pasta | caixa | sleeve | deckbox | outro
	Color       string `json:"color"`
	Description string `json:"description"`
	CardCount   int    `json:"card_count"`
}

type StorageInput struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type"`
	Color       string `json:"color"`
	Description string `json:"description"`
}
