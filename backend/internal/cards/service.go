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
		Quantity:         input.Quantity,
		Condition:        input.Condition,
		Notes:            input.Notes,
	}

	if card.Quantity <= 0 {
		card.Quantity = 1
	}

	if ext, _ := s.mtgClient.Search(input.SetCode, input.CollectionNumber, input.Language, input.Artist); ext != nil {
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

	// fallback: busca por set+número ou nome+set
	if external == nil {
		external, _ = s.mtgClient.Search(card.SetCode, card.CollectionNumber, card.Language, card.Artist)
		if external != nil {
			_ = s.repository.UpdateMTGID(id, external.ID)
		}
	}

	if external != nil {
		result["external"] = external
	}

	return result, nil
}

func (s *Service) Delete(id string) error {
	return s.repository.Delete(id)
}
