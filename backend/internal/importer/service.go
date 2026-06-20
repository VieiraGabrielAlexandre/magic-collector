package importer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"magic-collection-api/internal/cards"
	"magic-collection-api/internal/decks"
	"magic-collection-api/internal/mtgapi"
)

func parsePriceUSD(prices map[string]string, foil bool) float64 {
	if prices == nil {
		return 0
	}
	if foil {
		if v, ok := prices["usd_foil"]; ok && v != "" {
			f, _ := strconv.ParseFloat(v, 64)
			return f
		}
	}
	if v, ok := prices["usd"]; ok && v != "" {
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}
	return 0
}

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
	DeckID       int64    `json:"deck_id"`
	DeckName     string   `json:"deck_name"`
	Imported     int      `json:"imported"`
	Failed       int      `json:"failed"`
	TotalFromAPI int      `json:"total_from_api"`
	FailedCards  []string `json:"failed_cards,omitempty"`
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
		colorsStr := string(colorsJSON)

		card := cards.Card{
			MTGID:            ext.ID,
			Name:             cardName,
			Type:             ext.Type,
			CollectionNumber: ext.Number,
			Rarity:           ext.Rarity,
			SetCode:          ext.Set,
			ManaCost:         ext.ManaCost,
			Colors:           colorsStr,
			Color:            cards.ColorsJSONToDisplay(colorsStr),
			Language:         lang,
			Artist:           ext.Artist,
			Quantity:         1,
			Condition:        "Mint",
			DeckID:           int(deckID),
			PriceUSD:         parsePriceUSD(ext.Prices, false),
			ImageURL:         ext.ImageURL,
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

	result := s.importEntries(entries, deckID, input.SetCode, lang)
	result.DeckID = deckID
	result.DeckName = input.DeckName
	return result, nil
}

type ImportCardsToDeckInput struct {
	DeckList string `json:"deck_list" binding:"required"`
	SetCode  string `json:"set_code"`
	Language string `json:"language"`
}

func (s *Service) ImportCardsIntoDeck(deckID int64, input ImportCardsToDeckInput) (ImportResult, error) {
	lang := strings.ToUpper(strings.TrimSpace(input.Language))
	if lang == "" {
		lang = "EN"
	}

	entries := ParseDeckList(input.DeckList)
	if len(entries) == 0 {
		return ImportResult{}, fmt.Errorf("nenhuma carta encontrada na lista")
	}

	result := s.importEntries(entries, deckID, input.SetCode, lang)
	result.DeckID = deckID
	return result, nil
}

// importEntries busca cada carta na API e insere no banco. Reutilizado por ImportDeckList e ImportCardsIntoDeck.
func (s *Service) importEntries(entries []DeckListEntry, deckID int64, setCode, lang string) ImportResult {
	result := ImportResult{TotalFromAPI: len(entries)}

	for _, entry := range entries {
		time.Sleep(150 * time.Millisecond)

		// 3 tentativas na API com backoff curto; insert só ocorre após sucesso — sem risco de duplicata.
		var ext *mtgapi.ExternalCard
		for attempt := 1; attempt <= 3; attempt++ {
			if attempt > 1 {
				time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
			}
			ext, _ = s.mtgClient.SearchByName(entry.Name, setCode, lang)
			if ext != nil {
				break
			}
		}
		if ext == nil {
			result.Failed++
			result.FailedCards = append(result.FailedCards, entry.Name)
			continue
		}

		cardName := ext.PrintedName
		if cardName == "" {
			cardName = ext.Name
		}
		colorsJSON, _ := json.Marshal(ext.Colors)
		colorsStr := string(colorsJSON)

		card := cards.Card{
			MTGID:            ext.ID,
			Name:             cardName,
			Type:             ext.Type,
			CollectionNumber: ext.Number,
			Rarity:           ext.Rarity,
			SetCode:          ext.Set,
			ManaCost:         ext.ManaCost,
			Colors:           colorsStr,
			Color:            cards.ColorsJSONToDisplay(colorsStr),
			Language:         lang,
			Artist:           ext.Artist,
			Quantity:         entry.Quantity,
			Condition:        "Mint",
			DeckID:           int(deckID),
			PriceUSD:         parsePriceUSD(ext.Prices, false),
			ImageURL:         ext.ImageURL,
		}

		// database/sql já faz retry automático em ErrBadConn (broken pipe detectado antes do envio).
		if _, err := s.cardRepo.Create(card); err != nil {
			result.Failed++
			result.FailedCards = append(result.FailedCards, entry.Name+" (db: "+err.Error()+")")
		} else {
			result.Imported++
		}
	}

	return result
}
