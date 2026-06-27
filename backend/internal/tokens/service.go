package tokens

import (
	"encoding/json"

	"magic-collection-api/internal/mtgapi"
)

type Service struct {
	repo      *Repository
	mtgClient *mtgapi.Client
}

func NewService(repo *Repository, mtgClient *mtgapi.Client) *Service {
	return &Service{repo: repo, mtgClient: mtgClient}
}

func (s *Service) List() ([]Token, error) {
	return s.repo.List()
}

func (s *Service) Preview(input CreateTokenInput) (*mtgapi.ExternalToken, error) {
	return s.mtgClient.SearchToken(input.SetCode, input.CollectionNumber)
}

func (s *Service) Create(input CreateTokenInput) (int64, error) {
	ext, err := s.mtgClient.SearchToken(input.SetCode, input.CollectionNumber)
	if err != nil {
		return 0, err
	}

	t := Token{
		SetCode:          input.SetCode,
		CollectionNumber: input.CollectionNumber,
	}

	if ext != nil {
		colorsJSON, _ := json.Marshal(ext.Colors)
		t.Name = ext.Name
		t.TypeLine = ext.TypeLine
		t.OracleText = ext.OracleText
		t.Power = ext.Power
		t.Toughness = ext.Toughness
		t.Colors = string(colorsJSON)
		t.SetCode = ext.SetCode
		t.CollectionNumber = ext.CollectionNumber
		t.MtgID = ext.ID
		t.ImageURL = ext.ImageURL
		t.DoubleFaced = ext.DoubleFaced
		t.BackName = ext.BackName
		t.BackTypeLine = ext.BackTypeLine
		t.BackOracleText = ext.BackOracleText
		t.BackImageURL = ext.BackImageURL
		t.BackPower = ext.BackPower
		t.BackToughness = ext.BackToughness
		t.Artist = ext.Artist
	}

	t.Quantity = input.Quantity
	if t.Quantity <= 0 {
		t.Quantity = 1
	}
	t.Foil = input.Foil

	// Manual back face: overrides any auto-detected back face from Scryfall
	if input.BackSetCode != "" && input.BackCollectionNumber != "" {
		backExt, err2 := s.mtgClient.SearchToken(input.BackSetCode, input.BackCollectionNumber)
		if err2 == nil && backExt != nil {
			t.DoubleFaced = true
			t.BackName = backExt.Name
			t.BackTypeLine = backExt.TypeLine
			t.BackOracleText = backExt.OracleText
			t.BackPower = backExt.Power
			t.BackToughness = backExt.Toughness
			t.BackImageURL = backExt.ImageURL
		}
	}

	return s.repo.Create(t)
}

func (s *Service) UpdateQuantity(id string, quantity int) error {
	return s.repo.UpdateQuantity(id, quantity)
}

func (s *Service) Delete(id string) error {
	return s.repo.Delete(id)
}
