package decks

import (
	"strconv"

	"magic-collection-api/internal/mtgapi"
)

type Service struct {
	repo      *Repository
	mtgClient *mtgapi.Client
}

func NewService(repo *Repository, mtgClient *mtgapi.Client) *Service {
	return &Service{repo: repo, mtgClient: mtgClient}
}

func (s *Service) List() ([]Deck, error) {
	return s.repo.List()
}

func (s *Service) Create(input DeckInput) (int64, error) {
	id, err := s.repo.Create(input)
	if err != nil {
		return 0, err
	}
	if input.SetCode != "" {
		if set, _ := s.mtgClient.GetSetByCode(input.SetCode); set != nil && set.IconSVGURI != "" {
			_ = s.repo.UpdateIcon(strconv.FormatInt(id, 10), set.IconSVGURI)
		}
	}
	return id, nil
}

func (s *Service) Update(id string, input DeckInput) error {
	if err := s.repo.Update(id, input); err != nil {
		return err
	}
	if input.SetCode != "" {
		if set, _ := s.mtgClient.GetSetByCode(input.SetCode); set != nil && set.IconSVGURI != "" {
			_ = s.repo.UpdateIcon(id, set.IconSVGURI)
		}
	}
	return nil
}

func (s *Service) Delete(id string) error {
	return s.repo.Delete(id)
}

// FetchIcon busca e persiste o ícone do set para um deck que ainda não tem ícone.
func (s *Service) FetchIcon(id string) (string, error) {
	deck, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	if deck.SetCode == "" {
		return "", nil
	}
	set, err := s.mtgClient.GetSetByCode(deck.SetCode)
	if err != nil || set == nil || set.IconSVGURI == "" {
		return "", err
	}
	if err := s.repo.UpdateIcon(id, set.IconSVGURI); err != nil {
		return "", err
	}
	return set.IconSVGURI, nil
}
