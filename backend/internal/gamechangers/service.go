package gamechangers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List() ([]GameChanger, error) {
	return s.repo.List()
}

func (s *Service) SyncMeta() (int, string) {
	return s.repo.SyncMeta()
}

type scryfallSearchResp struct {
	Object   string           `json:"object"`
	HasMore  bool             `json:"has_more"`
	NextPage string           `json:"next_page"`
	Data     []scryfallCardGC `json:"data"`
}

type scryfallCardGC struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Set       string            `json:"set"`
	Rarity    string            `json:"rarity"`
	Colors    []string          `json:"colors"`
	ManaCost  string            `json:"mana_cost"`
	TypeLine  string            `json:"type_line"`
	ImageURIs map[string]string `json:"image_uris"`
	CardFaces []struct {
		ImageURIs map[string]string `json:"image_uris"`
	} `json:"card_faces"`
}

func (c *scryfallCardGC) imageURL() string {
	if c.ImageURIs != nil {
		if u, ok := c.ImageURIs["normal"]; ok {
			return u
		}
	}
	if len(c.CardFaces) > 0 && c.CardFaces[0].ImageURIs != nil {
		if u, ok := c.CardFaces[0].ImageURIs["normal"]; ok {
			return u
		}
	}
	return ""
}

func (s *Service) Sync() (SyncResult, error) {
	url := "https://api.scryfall.com/cards/search?q=is%3Agamechanger&order=color&unique=cards"
	client := &http.Client{Timeout: 30 * time.Second}

	var allCards []scryfallCardGC

	for url != "" {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return SyncResult{}, err
		}
		req.Header.Set("User-Agent", "magic-collector/1.0")
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return SyncResult{}, fmt.Errorf("erro ao buscar Scryfall: %w", err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != 200 {
			return SyncResult{}, fmt.Errorf("scryfall retornou %d: %s", resp.StatusCode, string(body))
		}

		var page scryfallSearchResp
		if err := json.Unmarshal(body, &page); err != nil {
			return SyncResult{}, fmt.Errorf("erro ao decodificar resposta: %w", err)
		}

		allCards = append(allCards, page.Data...)

		if page.HasMore && page.NextPage != "" {
			url = page.NextPage
			time.Sleep(100 * time.Millisecond)
		} else {
			url = ""
		}
	}

	var inserted, updated int
	scryfallIDs := make([]string, 0, len(allCards))

	for _, c := range allCards {
		rarityCode := rarityLetter(c.Rarity)
		colorsJSON := "[" + strings.Join(quoteSlice(c.Colors), ",") + "]"

		gc := GameChanger{
			ScryfallID: c.ID,
			Name:       c.Name,
			SetCode:    strings.ToUpper(c.Set),
			Rarity:     rarityCode,
			Colors:     colorsJSON,
			ImageURL:   c.imageURL(),
			ManaCost:   c.ManaCost,
			TypeLine:   c.TypeLine,
		}

		if err := s.repo.Upsert(gc); err == nil {
			inserted++
		}
		scryfallIDs = append(scryfallIDs, c.ID)
	}

	// remove cartas que já não são game changers
	s.repo.DeleteNotIn(scryfallIDs)

	return SyncResult{Inserted: inserted, Updated: updated, Total: len(allCards)}, nil
}

func rarityLetter(r string) string {
	switch strings.ToLower(r) {
	case "mythic":
		return "M"
	case "rare":
		return "R"
	case "uncommon":
		return "U"
	case "common":
		return "C"
	}
	return "R"
}

func quoteSlice(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = `"` + s + `"`
	}
	return out
}
