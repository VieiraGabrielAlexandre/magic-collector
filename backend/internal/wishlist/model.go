package wishlist

import (
	"encoding/json"
	"strings"
)

var colorCodeToPT = map[string]string{
	"W": "Branco", "U": "Azul", "B": "Preto",
	"R": "Vermelho", "G": "Verde", "C": "Incolor",
}

func colorsJSONToDisplay(colorsJSON string) string {
	if colorsJSON == "" || colorsJSON == "null" || colorsJSON == "[]" {
		return ""
	}
	var codes []string
	if err := json.Unmarshal([]byte(colorsJSON), &codes); err != nil {
		return ""
	}
	parts := make([]string, 0, len(codes))
	for _, c := range codes {
		if pt, ok := colorCodeToPT[c]; ok {
			parts = append(parts, pt)
		}
	}
	return strings.Join(parts, "/")
}

type WishlistCard struct {
	ID               int64   `json:"id"`
	MTGID            string  `json:"mtg_id"`
	SetCode          string  `json:"set_code"`
	CollectionNumber string  `json:"collection_number"`
	Name             string  `json:"name"`
	PrintedName      string  `json:"printed_name"`
	ImageURI         string  `json:"image_uri"`
	Artist           string  `json:"artist"`
	Rarity           string  `json:"rarity"`
	Colors           string  `json:"colors"`
	Color            string  `json:"color"`
	PriceUSD         float64 `json:"price_usd"`
	PriceUSDFoil     float64 `json:"price_usd_foil"`
	Foil             bool    `json:"foil"`
	Reason           string  `json:"reason"`
	Acquired         bool    `json:"acquired"`
	CreatedAt        string  `json:"created_at"`
}

type WishlistCardInput struct {
	SetCode          string `json:"set_code" binding:"required"`
	CollectionNumber string `json:"collection_number" binding:"required"`
	Foil             bool   `json:"foil"`
	Reason           string `json:"reason"`
}

type WishlistUpdateInput struct {
	Foil   bool   `json:"foil"`
	Reason string `json:"reason"`
}

type AcquireInput struct {
	DeckID     int    `json:"deck_id"`
	Condition  string `json:"condition"`
	Commander  bool   `json:"commander"`
	PreRelease bool   `json:"prerelease"`
}
