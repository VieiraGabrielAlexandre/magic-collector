package cards

import (
	"encoding/json"

	"magic-collection-api/internal/mtgapi"
)

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
