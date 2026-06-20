package cards

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"magic-collection-api/internal/mtgapi"
)

func parsePriceUSD(prices map[string]string, foil bool) float64 {
	if prices == nil {
		return 0
	}
	key := "usd"
	if foil {
		if v, ok := prices["usd_foil"]; ok && v != "" {
			f, _ := strconv.ParseFloat(v, 64)
			return f
		}
	}
	if v, ok := prices[key]; ok && v != "" {
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}
	return 0
}

type Service struct {
	repository cardRepository
	mtgClient  mtgAPIClient
}

func NewService(repository *Repository, mtgClient *mtgapi.Client) *Service {
	return &Service{
		repository: repository,
		mtgClient:  mtgClient,
	}
}

func (s *Service) Create(input CreateCardInput) (int64, error) {
	card := Card{
		Name:             input.Name,
		Colors:           input.Colors,
		Color:            ColorsJSONToDisplay(input.Colors),
		Type:             input.Type,
		Subtitle:         input.Subtitle,
		CollectionNumber: input.CollectionNumber,
		Rarity:           input.Rarity,
		SetCode:          input.SetCode,
		Language:         input.Language,
		Year:             input.Year,
		Artist:           input.Artist,
		Foil:             input.Foil,
		PreRelease:       input.PreRelease,
		Commander:        input.Commander,
		PreconDeck:       input.PreconDeck,
		DeckID:           input.DeckID,
		Quantity:         input.Quantity,
		Condition:        input.Condition,
		Notes:            input.Notes,
	}

	card.PreRelease = input.PreRelease

	if card.Quantity <= 0 {
		card.Quantity = 1
	}

	var fetchExt func() (*mtgapi.ExternalCard, error)
	if input.PreRelease {
		fetchExt = func() (*mtgapi.ExternalCard, error) {
			return s.mtgClient.SearchPreRelease(input.Name, input.Language, input.Artist)
		}
	} else {
		fetchExt = func() (*mtgapi.ExternalCard, error) {
			return s.mtgClient.Search(input.SetCode, input.CollectionNumber, input.Language, input.Artist)
		}
	}

	if ext, _ := fetchExt(); ext != nil {
		colorsJSON, _ := json.Marshal(ext.Colors)
		card.MTGID = ext.ID
		card.SetCode = ext.Set
		card.Rarity = ext.Rarity
		card.Type = ext.Type
		card.ManaCost = ext.ManaCost
		card.Colors = string(colorsJSON)
		card.Color = ColorsJSONToDisplay(card.Colors)
		card.PriceUSD = parsePriceUSD(ext.Prices, card.Foil)
		card.ImageURL = ext.ImageURL
		card.FullArt = ext.FullArt
		if ext.Year > 0 {
			card.Year = ext.Year
		}
		if ext.Artist != "" && card.Artist == "" {
			card.Artist = ext.Artist
		}
	}

	return s.repository.Create(card)
}

func (s *Service) List(params ListParams) (ListResult, error) {
	return s.repository.List(params)
}

func (s *Service) GetByID(id string) (map[string]any, error) {
	card, err := s.repository.GetByID(id)
	if err != nil {
		return nil, err
	}

	result := map[string]any{
		"local": card,
	}

	var external *mtgapi.ExternalCard

	// tenta pelo ID cacheado (rápido, sem nova busca)
	if card.MTGID != "" {
		external, _ = s.mtgClient.GetByMTGID(card.MTGID)
	}

	// fallback: busca por set+número (normal) ou por nome com is:prerelease (pré-release)
	if external == nil {
		if card.PreRelease {
			external, _ = s.mtgClient.SearchPreRelease(card.Name, card.Language, card.Artist)
		} else {
			external, _ = s.mtgClient.Search(card.SetCode, card.CollectionNumber, card.Language, card.Artist)
		}
		if external != nil {
			_ = s.repository.UpdateMTGID(id, external.ID)
		}
	}

	if external != nil {
		result["external"] = external
	}

	return result, nil
}

func (s *Service) Update(id string, input UpdateCardInput) error {
	current, err := s.repository.GetByID(id)
	if err != nil {
		return err
	}

	card := Card{
		// Campos do usuário
		Name:             input.Name,
		Subtitle:         input.Subtitle,
		CollectionNumber: input.CollectionNumber,
		SetCode:          input.SetCode,
		Language:         input.Language,
		Year:             input.Year,
		Artist:           input.Artist,
		Foil:             input.Foil,
		PreRelease:       input.PreRelease,
		Commander:        input.Commander,
		PreconDeck:       input.PreconDeck,
		DeckID:           input.DeckID,
		Quantity:         input.Quantity,
		Condition:        input.Condition,
		Notes:            input.Notes,
		// Campos externos: preserva os valores do banco até o re-fetch
		MTGID:    current.MTGID,
		Type:     current.Type,
		ManaCost: current.ManaCost,
		Colors:   current.Colors,
		Color:    current.Color,
		Rarity:   current.Rarity,
		PriceUSD: current.PriceUSD,
		ImageURL: current.ImageURL,
		FullArt:  current.FullArt,
	}

	// Re-fetch do Scryfall para atualizar campos externos (type, rarity, image, colors, mana_cost, price)
	var ext *mtgapi.ExternalCard
	if current.MTGID != "" {
		ext, _ = s.mtgClient.GetByMTGID(current.MTGID)
	}
	if ext == nil {
		if card.PreRelease {
			ext, _ = s.mtgClient.SearchPreRelease(card.Name, card.Language, card.Artist)
		} else {
			ext, _ = s.mtgClient.Search(card.SetCode, card.CollectionNumber, card.Language, card.Artist)
		}
	}
	if ext != nil {
		colorsJSON, _ := json.Marshal(ext.Colors)
		card.MTGID = ext.ID
		card.Type = ext.Type
		card.ManaCost = ext.ManaCost
		card.Colors = string(colorsJSON)
		card.Color = ColorsJSONToDisplay(card.Colors)
		card.Rarity = ext.Rarity
		card.PriceUSD = parsePriceUSD(ext.Prices, card.Foil)
		card.ImageURL = ext.ImageURL
		card.FullArt = ext.FullArt
		if ext.Year > 0 {
			card.Year = ext.Year
		}
		if ext.Artist != "" && card.Artist == "" {
			card.Artist = ext.Artist
		}
	}

	if input.Propagate {
		if err := s.repository.UpdateSharedByIdentity(
			current.Name, current.SetCode, current.CollectionNumber,
			current.Language, current.Foil, card,
		); err != nil {
			return err
		}
	}

	return s.repository.Update(id, card)
}

func (s *Service) Delete(id string) error {
	return s.repository.Delete(id)
}

func (s *Service) SetQuantity(id string, quantity int) error {
	return s.repository.SetQuantity(id, quantity)
}

func (s *Service) SetDeck(id string, deckID int) error {
	return s.repository.SetDeck(id, deckID)
}

func (s *Service) ExportAll() ([]Card, error) {
	return s.repository.ListAll()
}

func (s *Service) GetCardsForDeckBuilder() ([]DeckBuilderCard, error) {
	return s.repository.ListForDeckBuilder()
}

func (s *Service) ListColorCombos() ([]ColorCombo, error) {
	return s.repository.ListColorCombos()
}

func (s *Service) NormalizeRarities() (NormalizeRarityResult, error) {
	return s.repository.NormalizeRarities()
}

// RefreshColorsResult resume o resultado da atualização em lote de cores.
type RefreshColorsResult struct {
	Updated int `json:"updated"`
	Skipped int `json:"skipped"` // sem set_code+number nem mtg_id, ou Scryfall não retornou cores
	Failed  int `json:"failed"`
	Total   int `json:"total"`
}

// RefreshColors busca cores na Scryfall para todas as cartas sem `colors` definido.
func (s *Service) RefreshColors() (RefreshColorsResult, error) {
	cards, err := s.repository.ListCardsWithoutColors()
	if err != nil {
		return RefreshColorsResult{}, err
	}

	result := RefreshColorsResult{Total: len(cards)}

	for _, c := range cards {
		time.Sleep(80 * time.Millisecond)

		var ext *mtgapi.ExternalCard
		if c.MTGID != "" {
			ext, _ = s.mtgClient.GetByMTGID(c.MTGID)
		}
		if ext == nil && c.SetCode != "" && c.CollectionNumber != "" {
			ext, _ = s.mtgClient.Search(c.SetCode, c.CollectionNumber, c.Language, c.Artist)
		}

		if ext == nil {
			result.Skipped++
			continue
		}

		colorsJSON, _ := json.Marshal(ext.Colors)
		colorsStr := string(colorsJSON)
		colorDisplay := ColorsJSONToDisplay(colorsStr)

		var updateErr error
		if c.MTGID == "" && ext.ID != "" {
			updateErr = s.repository.UpdateColorsAndMTGID(c.ID, ext.ID, colorsStr, colorDisplay)
		} else {
			updateErr = s.repository.UpdateColors(c.ID, colorsStr, colorDisplay)
		}

		if updateErr != nil {
			result.Failed++
		} else {
			result.Updated++
		}
	}

	return result, nil
}

func (s *Service) GetStats() (CollectionStats, error) {
	return s.repository.GetStats()
}

// Preview busca a carta na Scryfall sem salvar no banco — usado para confirmar antes de cadastrar.
func (s *Service) Preview(input PreviewCardInput) (*mtgapi.ExternalCard, error) {
	if input.PreRelease {
		return s.mtgClient.SearchPreRelease(input.Name, input.Language, input.Artist)
	}
	return s.mtgClient.Search(input.SetCode, input.CollectionNumber, input.Language, input.Artist)
}

// ImageRefreshResult resume o resultado da atualização em lote de imagens.
type ImageRefreshResult struct {
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
	Total   int `json:"total"`
}

// RefreshImages atualiza image_url de todas as cartas via Scryfall.
func (s *Service) RefreshImages() (ImageRefreshResult, error) {
	cards, err := s.repository.ListAllForPriceRefresh()
	if err != nil {
		return ImageRefreshResult{}, err
	}

	result := ImageRefreshResult{Total: len(cards)}

	for _, c := range cards {
		time.Sleep(80 * time.Millisecond)

		var ext *mtgapi.ExternalCard
		if c.MTGID != "" {
			ext, _ = s.mtgClient.GetByMTGID(c.MTGID)
		}
		if ext == nil && c.SetCode != "" && c.CollectionNumber != "" {
			ext, _ = s.mtgClient.Search(c.SetCode, c.CollectionNumber, c.Language, c.Artist)
		}

		if ext == nil || ext.ImageURL == "" {
			result.Skipped++
			continue
		}

		var updateErr error
		if c.MTGID == "" && ext.ID != "" {
			updateErr = s.repository.UpdateImageURLAndMTGID(c.ID, ext.ID, ext.ImageURL)
		} else {
			updateErr = s.repository.UpdateImageURL(c.ID, ext.ImageURL)
		}

		if updateErr != nil {
			result.Failed++
		} else {
			result.Updated++
		}
	}

	return result, nil
}

// PriceRefreshResult resume o resultado da atualização em lote de preços.
type PriceRefreshResult struct {
	Updated int `json:"updated"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"` // sem mtg_id e sem set+number
	Total   int `json:"total"`
}

// RefreshPrices atualiza price_usd de todas as cartas buscando na Scryfall.
// Se emptyOnly=true, processa apenas cartas com price_usd = 0.
// Para cartas não-EN sem preço, tenta automaticamente a versão EN como fallback.
func (s *Service) RefreshPrices(emptyOnly bool) (PriceRefreshResult, error) {
	var cards []CardForPriceRefresh
	var err error
	if emptyOnly {
		cards, err = s.repository.ListEmptyPricesForRefresh()
	} else {
		cards, err = s.repository.ListAllForPriceRefresh()
	}
	if err != nil {
		return PriceRefreshResult{}, err
	}

	result := PriceRefreshResult{Total: len(cards)}

	for _, c := range cards {
		time.Sleep(80 * time.Millisecond)

		var ext *mtgapi.ExternalCard

		// 1ª tentativa: UUID Scryfall (rápido e exato)
		if c.MTGID != "" {
			ext, _ = s.mtgClient.GetByMTGID(c.MTGID)
		}
		// 2ª tentativa: set + número (cartas sem mtg_id)
		if ext == nil && c.SetCode != "" && c.CollectionNumber != "" {
			ext, _ = s.mtgClient.Search(c.SetCode, c.CollectionNumber, c.Language, c.Artist)
		}

		if ext == nil {
			result.Skipped++
			continue
		}

		price := parsePriceUSD(ext.Prices, c.Foil)

		// Fallback EN: carta não-EN sem preço → busca versão EN para pegar o preço
		if price == 0 && strings.ToUpper(c.Language) != "EN" && ext.Set != "" && ext.Number != "" {
			time.Sleep(80 * time.Millisecond)
			enExt, _ := s.mtgClient.Search(ext.Set, ext.Number, "EN", "")
			if enExt != nil {
				price = parsePriceUSD(enExt.Prices, c.Foil)
			}
		}

		var updateErr error
		if c.MTGID == "" && ext.ID != "" {
			updateErr = s.repository.UpdatePriceAndMTGID(c.ID, ext.ID, price)
		} else {
			updateErr = s.repository.UpdatePrice(c.ID, price)
		}

		if updateErr != nil {
			result.Failed++
		} else {
			result.Updated++
		}
	}

	return result, nil
}
