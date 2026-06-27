package tokens

type Token struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	TypeLine         string `json:"type_line"`
	OracleText       string `json:"oracle_text"`
	Power            string `json:"power"`
	Toughness        string `json:"toughness"`
	Colors           string `json:"colors"`
	SetCode          string `json:"set_code"`
	CollectionNumber string `json:"collection_number"`
	MtgID            string `json:"mtg_id"`
	ImageURL         string `json:"image_url"`
	DoubleFaced      bool   `json:"double_faced"`
	BackName         string `json:"back_name"`
	BackTypeLine     string `json:"back_type_line"`
	BackOracleText   string `json:"back_oracle_text"`
	BackImageURL     string `json:"back_image_url"`
	BackPower        string `json:"back_power"`
	BackToughness    string `json:"back_toughness"`
	Artist           string `json:"artist"`
	Quantity         int    `json:"quantity"`
	Foil             bool   `json:"foil"`
	CreatedAt        string `json:"created_at"`
}

type CreateTokenInput struct {
	SetCode              string `json:"set_code" binding:"required"`
	CollectionNumber     string `json:"collection_number" binding:"required"`
	BackSetCode          string `json:"back_set_code"`
	BackCollectionNumber string `json:"back_collection_number"`
	Quantity             int    `json:"quantity"`
	Foil                 bool   `json:"foil"`
}
