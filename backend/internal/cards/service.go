package cards

import (
	"encoding/json"
	"strconv"
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
	repository *Repository
	mtgClient  *mtgapi.Client
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
		Color:            input.Color,
		Type:             input.Type,
		Subtitle:         input.Subtitle,
		CollectionNumber: input.CollectionNumber,
		Rarity:           input.Rarity,
		SetCode:          input.SetCode,
		Language:         input.Language,
		Year:             input.Year,
		Artist:           input.Artist,
		Company:          input.Company,
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
		colors, _ := json.Marshal(ext.Colors)
		card.MTGID = ext.ID
		card.SetCode = ext.Set
		card.Rarity = ext.Rarity
		card.Type = ext.Type
		card.ManaCost = ext.ManaCost
		card.Colors = string(colors)
		card.PriceUSD = parsePriceUSD(ext.Prices, card.Foil)
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
		Name:             input.Name,
		Color:            input.Color,
		Type:             input.Type,
		Subtitle:         input.Subtitle,
		CollectionNumber: input.CollectionNumber,
		Rarity:           input.Rarity,
		SetCode:          input.SetCode,
		Language:         input.Language,
		Year:             input.Year,
		Artist:           input.Artist,
		Company:          input.Company,
		Foil:             input.Foil,
		PreRelease:       input.PreRelease,
		Commander:        input.Commander,
		PreconDeck:       input.PreconDeck,
		DeckID:           input.DeckID,
		Quantity:         input.Quantity,
		Condition:        input.Condition,
		Notes:            input.Notes,
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

func (s *Service) SetDeck(id string, deckID int) error {
	return s.repository.SetDeck(id, deckID)
}

func (s *Service) ExportAll() ([]Card, error) {
	return s.repository.ListAll()
}

func (s *Service) GetCardsForDeckBuilder() ([]DeckBuilderCard, error) {
	return s.repository.ListForDeckBuilder()
}

func (s *Service) GetStats() (CollectionStats, error) {
	return s.repository.GetStats()
}

// PriceRefreshResult resume o resultado da atualização em lote de preços.
type PriceRefreshResult struct {
	Updated int `json:"updated"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"` // sem mtg_id e sem set+number
	Total   int `json:"total"`
}

// RefreshPrices atualiza price_usd de todas as cartas buscando na Scryfall.
// Usa GetByMTGID quando possível; caso contrário, tenta Search por set+número.
func (s *Service) RefreshPrices() (PriceRefreshResult, error) {
	cards, err := s.repository.ListAllForPriceRefresh()
	if err != nil {
		return PriceRefreshResult{}, err
	}

	result := PriceRefreshResult{Total: len(cards)}

	for _, c := range cards {
		time.Sleep(80 * time.Millisecond)

		var ext *mtgapi.ExternalCard

		// 1ª tentativa: UUID Scryfall (mais rápido e exato)
		if c.MTGID != "" {
			ext, _ = s.mtgClient.GetByMTGID(c.MTGID)
		}

		// 2ª tentativa: set + coleção (para cartas inseridas manualmente)
		if ext == nil && c.SetCode != "" && c.CollectionNumber != "" {
			ext, _ = s.mtgClient.Search(c.SetCode, c.CollectionNumber, c.Language, c.Artist)
		}

		if ext == nil {
			result.Skipped++
			continue
		}

		price := parsePriceUSD(ext.Prices, c.Foil)

		var updateErr error
		if c.MTGID == "" && ext.ID != "" {
			// Aproveita e salva o mtg_id que estava faltando
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
