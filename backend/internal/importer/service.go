package importer

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"magic-collection-api/internal/cards"
	"magic-collection-api/internal/decks"
	"magic-collection-api/internal/mtgapi"
)

type ImportPreconInput struct {
	SetCode     string `json:"set_code" binding:"required"`
	Language    string `json:"language"`
	DeckName    string `json:"deck_name" binding:"required"`
	Description string `json:"description"`
	Commander   bool   `json:"commander"`
	Colors      string `json:"colors"`
	ThemeColor  string `json:"theme_color"`
}

type ImportResult struct {
	DeckID       int64  `json:"deck_id"`
	DeckName     string `json:"deck_name"`
	Imported     int    `json:"imported"`
	Failed       int    `json:"failed"`
	TotalFromAPI int    `json:"total_from_api"`
}

type Service struct {
	deckRepo  *decks.Repository
	cardRepo  *cards.Repository
	mtgClient *mtgapi.Client
}

func NewService(deckRepo *decks.Repository, cardRepo *cards.Repository, mtgClient *mtgapi.Client) *Service {
	return &Service{deckRepo: deckRepo, cardRepo: cardRepo, mtgClient: mtgClient}
}

func (s *Service) ImportPrecon(input ImportPreconInput) (ImportResult, error) {
	lang := strings.ToUpper(strings.TrimSpace(input.Language))
	if lang == "" {
		lang = "EN"
	}

	deckID, err := s.deckRepo.Create(decks.DeckInput{
		Name:        input.DeckName,
		Description: input.Description,
		Commander:   input.Commander,
		Colors:      input.Colors,
		SetCode:     strings.ToLower(strings.TrimSpace(input.SetCode)),
		ThemeColor:  input.ThemeColor,
	})
	if err != nil {
		return ImportResult{}, fmt.Errorf("criando deck: %w", err)
	}

	apiCards, err := s.mtgClient.FetchSetCards(input.SetCode, lang)
	if err != nil {
		return ImportResult{}, fmt.Errorf("buscando cartas do set: %w", err)
	}

	result := ImportResult{
		DeckID:       deckID,
		DeckName:     input.DeckName,
		TotalFromAPI: len(apiCards),
	}

	for _, ext := range apiCards {
		cardName := ext.PrintedName
		if cardName == "" {
			cardName = ext.Name
		}

		colorsJSON, _ := json.Marshal(ext.Colors)

		card := cards.Card{
			MTGID:            ext.ID,
			Name:             cardName,
			Type:             ext.Type,
			CollectionNumber: ext.Number,
			Rarity:           ext.Rarity,
			SetCode:          ext.Set,
			ManaCost:         ext.ManaCost,
			Colors:           string(colorsJSON),
			Language:         lang,
			Artist:           ext.Artist,
			Company:          "Wizards of the Coast",
			Quantity:         1,
			Condition:        "Mint",
			DeckID:           int(deckID),
		}

		if _, err := s.cardRepo.Create(card); err != nil {
			result.Failed++
		} else {
			result.Imported++
		}
	}

	return result, nil
}

type ImportDeckListInput struct {
	DeckName    string `json:"deck_name" binding:"required"`
	DeckList    string `json:"deck_list" binding:"required"`
	SetCode     string `json:"set_code"`
	Language    string `json:"language"`
	Description string `json:"description"`
	Commander   bool   `json:"commander"`
	Colors      string `json:"colors"`
	ThemeColor  string `json:"theme_color"`
}

func (s *Service) ImportDeckList(input ImportDeckListInput) (ImportResult, error) {
	lang := strings.ToUpper(strings.TrimSpace(input.Language))
	if lang == "" {
		lang = "EN"
	}

	entries := ParseDeckList(input.DeckList)
	if len(entries) == 0 {
		return ImportResult{}, fmt.Errorf("nenhuma carta encontrada na lista")
	}

	deckID, err := s.deckRepo.Create(decks.DeckInput{
		Name:        input.DeckName,
		Description: input.Description,
		Commander:   input.Commander,
		Colors:      input.Colors,
		SetCode:     strings.ToLower(strings.TrimSpace(input.SetCode)),
		ThemeColor:  input.ThemeColor,
	})
	if err != nil {
		return ImportResult{}, fmt.Errorf("criando deck: %w", err)
	}

	result := ImportResult{
		DeckID:       deckID,
		DeckName:     input.DeckName,
		TotalFromAPI: len(entries),
	}

	for _, entry := range entries {
		time.Sleep(75 * time.Millisecond)
		ext, err := s.mtgClient.SearchByName(entry.Name, input.SetCode, lang)
		if err != nil || ext == nil {
			result.Failed++
			continue
		}

		cardName := ext.PrintedName
		if cardName == "" {
			cardName = ext.Name
		}

		colorsJSON, _ := json.Marshal(ext.Colors)

		card := cards.Card{
			MTGID:            ext.ID,
			Name:             cardName,
			Type:             ext.Type,
			CollectionNumber: ext.Number,
			Rarity:           ext.Rarity,
			SetCode:          ext.Set,
			ManaCost:         ext.ManaCost,
			Colors:           string(colorsJSON),
			Language:         lang,
			Artist:           ext.Artist,
			Company:          "Wizards of the Coast",
			Quantity:         entry.Quantity,
			Condition:        "Mint",
			DeckID:           int(deckID),
		}

		if _, err := s.cardRepo.Create(card); err != nil {
			result.Failed++
		} else {
			result.Imported++
		}
	}

	return result, nil
}
