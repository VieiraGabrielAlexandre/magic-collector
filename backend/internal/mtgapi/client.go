package mtgapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiBase = "https://api.scryfall.com/cards"

type Client struct {
	http *http.Client
}

type ExternalCard struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`         // nome em inglês
	PrintedName string            `json:"printed_name"` // nome no idioma da carta (quando não EN)
	Set         string            `json:"set"`
	SetName     string            `json:"set_name"`
	Rarity      string            `json:"rarity"`
	Type        string            `json:"type"`         // type line em inglês
	PrintedType string            `json:"printed_type"` // type line no idioma da carta
	ManaCost    string            `json:"mana_cost"`
	Colors      []string          `json:"colors"`
	ImageURL    string            `json:"image_url"`
	Text        string            `json:"text"`         // oracle text em inglês
	PrintedText string            `json:"printed_text"` // texto no idioma da carta
	FlavorText  string            `json:"flavor_text"`
	Artist      string            `json:"artist"`
	Number      string            `json:"number"`
	Power       string            `json:"power"`
	Toughness   string            `json:"toughness"`
	Prices      map[string]string `json:"prices"`
	ScryfallURI string            `json:"scryfall_uri"`
}

type scryfallList struct {
	Data []scryfallCard `json:"data"`
}

type scryfallCard struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	PrintedName     string            `json:"printed_name"`
	Set             string            `json:"set"`
	SetName         string            `json:"set_name"`
	Rarity          string            `json:"rarity"`
	TypeLine        string            `json:"type_line"`
	PrintedTypeLine string            `json:"printed_type_line"`
	ManaCost        string            `json:"mana_cost"`
	Colors          []string          `json:"colors"`
	ImageURIs       map[string]string `json:"image_uris"`
	OracleText      string            `json:"oracle_text"`
	PrintedText     string            `json:"printed_text"`
	FlavorText      string            `json:"flavor_text"`
	Artist          string            `json:"artist"`
	CollectorNumber string            `json:"collector_number"`
	Power           string            `json:"power"`
	Toughness       string            `json:"toughness"`
	Prices          map[string]string `json:"prices"`
	ScryfallURI     string            `json:"scryfall_uri"`
}

func (s *scryfallCard) toExternal() *ExternalCard {
	imageURL := ""
	if s.ImageURIs != nil {
		if u, ok := s.ImageURIs["normal"]; ok {
			imageURL = u
		}
	}
	return &ExternalCard{
		ID:          s.ID,
		Name:        s.Name,
		PrintedName: s.PrintedName,
		Set:         strings.ToUpper(s.Set),
		SetName:     s.SetName,
		Rarity:      s.Rarity,
		Type:        s.TypeLine,
		PrintedType: s.PrintedTypeLine,
		ManaCost:    s.ManaCost,
		Colors:      s.Colors,
		ImageURL:    imageURL,
		Text:        s.OracleText,
		PrintedText: s.PrintedText,
		FlavorText:  s.FlavorText,
		Artist:      s.Artist,
		Number:      s.CollectorNumber,
		Power:       s.Power,
		Toughness:   s.Toughness,
		Prices:      s.Prices,
		ScryfallURI: s.ScryfallURI,
	}
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 10 * time.Second}}
}

// Search busca uma carta exclusivamente por set + número.
// Tenta primeiro com o idioma da carta; se não disponível, busca sem idioma (EN).
// O artista é usado para validar o resultado e rejeitar correspondências incorretas.
func (c *Client) Search(setCode, number, lang, artist string) (*ExternalCard, error) {
	if setCode == "" || number == "" {
		return nil, nil
	}
	set := strings.ToLower(setCode)

	// tenta no idioma da carta (quando não EN)
	if langCode := toLangCode(lang); langCode != "en" {
		card, err := c.fetch(fmt.Sprintf("%s/%s/%s/%s", apiBase, set, number, langCode))
		if err != nil {
			return nil, err
		}
		if card != nil && artistsMatch(artist, card.Artist) {
			return card, nil
		}
	}

	// fallback para EN (ou idioma original quando lang era EN)
	return c.fetch(fmt.Sprintf("%s/%s/%s", apiBase, set, number))
}

// SearchPreRelease busca uma carta pré-release pelo nome usando o endpoint de busca do Scryfall.
// Cartas pré-release têm set codes diferentes (ex: "pgrn") e não são encontradas por set+número normal.
func (c *Client) SearchPreRelease(name, lang, artist string) (*ExternalCard, error) {
	if name == "" {
		return nil, nil
	}

	q := fmt.Sprintf(`is:prerelease name:"%s"`, name)
	endpoint := fmt.Sprintf("https://api.scryfall.com/cards/search?q=%s&order=released&dir=desc",
		url.QueryEscape(q))

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "magic-collector/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	var list scryfallList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil
	}

	// Prefere o resultado cujo artista bate com o cadastrado.
	best := &list.Data[0]
	if artist != "" {
		for i := range list.Data {
			if artistsMatch(artist, list.Data[i].Artist) {
				best = &list.Data[i]
				break
			}
		}
	}

	// Se não for EN, tenta buscar a versão localizada pelo set+número do resultado.
	if langCode := toLangCode(lang); langCode != "en" && best.Set != "" && best.CollectorNumber != "" {
		langEndpoint := fmt.Sprintf("%s/%s/%s/%s", apiBase, best.Set, best.CollectorNumber, langCode)
		if langCard, _ := c.fetch(langEndpoint); langCard != nil {
			return langCard, nil
		}
	}

	return best.toExternal(), nil
}

// GetByMTGID busca pelo UUID Scryfall armazenado em cache.
func (c *Client) GetByMTGID(id string) (*ExternalCard, error) {
	return c.fetch(fmt.Sprintf("%s/%s", apiBase, id))
}

func (c *Client) fetch(endpoint string) (*ExternalCard, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "magic-collector/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}
	var card scryfallCard
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, err
	}
	if card.ID == "" {
		return nil, nil
	}
	return card.toExternal(), nil
}

// SetInfo contém os dados relevantes de um set do Scryfall.
type SetInfo struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	IconSVGURI string `json:"icon_svg_uri"`
	ReleasedAt string `json:"released_at"`
	SetType    string `json:"set_type"`
	CardCount  int    `json:"card_count"`
}

// GetSetByCode busca dados de um set pelo seu código (ex: "dmu", "bro").
func (c *Client) GetSetByCode(code string) (*SetInfo, error) {
	if code == "" {
		return nil, nil
	}
	endpoint := "https://api.scryfall.com/sets/" + url.PathEscape(strings.ToLower(strings.TrimSpace(code)))
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "magic-collector/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}
	var set SetInfo
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		return nil, err
	}
	return &set, nil
}

// toLangCode converte os códigos de idioma do nosso DB para os do Scryfall.
func toLangCode(lang string) string {
	switch strings.ToUpper(lang) {
	case "PT":
		return "pt"
	case "ES":
		return "es"
	case "FR":
		return "fr"
	case "DE":
		return "de"
	case "JP":
		return "ja"
	case "IT":
		return "it"
	case "RU":
		return "ru"
	case "KO":
		return "ko"
	case "ZHS":
		return "zhs"
	case "ZHT":
		return "zht"
	default:
		return "en"
	}
}

// artistsMatch retorna true quando os nomes de artista são compatíveis.
// Aceita correspondência parcial para cobrir variações como
// "Dan Scott" ↔ "Dan Murayama Scott".
func artistsMatch(stored, fromAPI string) bool {
	if stored == "" || fromAPI == "" {
		return true
	}
	a := strings.ToLower(strings.TrimSpace(stored))
	b := strings.ToLower(strings.TrimSpace(fromAPI))
	return a == b || strings.Contains(b, a) || strings.Contains(a, b)
}
