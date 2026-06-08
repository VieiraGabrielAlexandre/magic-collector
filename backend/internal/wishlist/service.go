package wishlist

import (
	"encoding/json"
	"strconv"
	"strings"

	"magic-collection-api/internal/mtgapi"
)

type Service struct {
	repo      wishlistRepository
	mtgClient wishlistMtgClient
}

func NewService(repo *Repository, mtgClient *mtgapi.Client) *Service {
	return &Service{repo: repo, mtgClient: mtgClient}
}

func (s *Service) List() ([]WishlistCard, error) {
	return s.repo.List()
}

func (s *Service) GetByID(id string) (*WishlistCard, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(input WishlistCardInput) (int64, error) {
	setCode := strings.ToUpper(strings.TrimSpace(input.SetCode))
	number := strings.TrimSpace(input.CollectionNumber)

	card := WishlistCard{
		SetCode:          setCode,
		CollectionNumber: number,
		Foil:             input.Foil,
		Reason:           strings.TrimSpace(input.Reason),
		Colors:           "[]",
	}

	if ext, _ := s.mtgClient.Search(setCode, number, "EN", ""); ext != nil {
		colorsJSON, _ := json.Marshal(ext.Colors)
		card.MTGID = ext.ID
		card.Name = ext.Name
		card.PrintedName = ext.PrintedName
		card.ImageURI = ext.ImageURL
		card.Artist = ext.Artist
		card.Rarity = ext.Rarity
		card.Colors = string(colorsJSON)
		card.Color = colorsJSONToDisplay(card.Colors)
		card.PriceUSD = parsePrice(ext.Prices, "usd")
		card.PriceUSDFoil = parsePrice(ext.Prices, "usd_foil")
	}

	return s.repo.Create(card)
}

func (s *Service) Update(id string, input WishlistUpdateInput) error {
	return s.repo.Update(id, input)
}

func (s *Service) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *Service) Acquire(id string, input AcquireInput) (int64, error) {
	return s.repo.Acquire(id, input)
}

func parsePrice(prices map[string]string, key string) float64 {
	if prices == nil {
		return 0
	}
	v, ok := prices[key]
	if !ok || v == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(v, 64)
	return f
}
