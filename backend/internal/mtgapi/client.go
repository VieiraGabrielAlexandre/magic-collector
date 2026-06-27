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

// NormalizeRarity converte o valor de raridade (Scryfall lowercase ou variações)
// para o código de uma letra usado no banco: C U R M L T.
func NormalizeRarity(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "c", "common":
		return "C"
	case "u", "uncommon":
		return "U"
	case "r", "rare", "special", "bonus", "timeshifted":
		return "R"
	case "m", "mythic", "mythic rare":
		return "M"
	case "l", "land", "basic land":
		return "L"
	case "t", "token":
		return "T"
	default:
		v := strings.ToUpper(strings.TrimSpace(s))
		if v == "" {
			return ""
		}
		return v
	}
}

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
	FullArt     bool              `json:"full_art"`
	Year        int               `json:"year"`
}

type scryfallList struct {
	Data     []scryfallCard `json:"data"`
	HasMore  bool           `json:"has_more"`
	NextPage string         `json:"next_page"`
}

type scryfallCardFace struct {
	Name        string            `json:"name"`
	TypeLine    string            `json:"type_line"`
	OracleText  string            `json:"oracle_text"`
	Power       string            `json:"power"`
	Toughness   string            `json:"toughness"`
	ImageURIs   map[string]string `json:"image_uris"`
	PrintedName string            `json:"printed_name"`
	ManaCost    string            `json:"mana_cost"`
	Colors      []string          `json:"colors"`
}

// ExternalToken holds token data fetched from Scryfall (single or double-faced).
type ExternalToken struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	TypeLine         string   `json:"type_line"`
	OracleText       string   `json:"oracle_text"`
	Power            string   `json:"power"`
	Toughness        string   `json:"toughness"`
	Colors           []string `json:"colors"`
	SetCode          string   `json:"set_code"`
	CollectionNumber string   `json:"collection_number"`
	ImageURL         string   `json:"image_url"`
	DoubleFaced      bool     `json:"double_faced"`
	BackName         string   `json:"back_name"`
	BackTypeLine     string   `json:"back_type_line"`
	BackOracleText   string   `json:"back_oracle_text"`
	BackImageURL     string   `json:"back_image_url"`
	BackPower        string   `json:"back_power"`
	BackToughness    string   `json:"back_toughness"`
	Artist           string   `json:"artist"`
}

type scryfallCard struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	PrintedName     string             `json:"printed_name"`
	Set             string             `json:"set"`
	SetName         string             `json:"set_name"`
	Rarity          string             `json:"rarity"`
	TypeLine        string             `json:"type_line"`
	PrintedTypeLine string             `json:"printed_type_line"`
	ManaCost        string             `json:"mana_cost"`
	Colors          []string           `json:"colors"`
	ColorIdentity   []string           `json:"color_identity"`
	ImageURIs       map[string]string  `json:"image_uris"`
	CardFaces       []scryfallCardFace `json:"card_faces"`
	OracleText      string             `json:"oracle_text"`
	PrintedText     string             `json:"printed_text"`
	FlavorText      string             `json:"flavor_text"`
	Artist          string             `json:"artist"`
	CollectorNumber string             `json:"collector_number"`
	Power           string             `json:"power"`
	Toughness       string             `json:"toughness"`
	Prices          map[string]string  `json:"prices"`
	ScryfallURI     string             `json:"scryfall_uri"`
	Layout          string             `json:"layout"`
	FullArt         bool               `json:"full_art"`
	ReleasedAt      string             `json:"released_at"`
}

func yearFromReleasedAt(releasedAt string) int {
	if len(releasedAt) >= 4 {
		y := 0
		for _, ch := range releasedAt[:4] {
			if ch < '0' || ch > '9' {
				return 0
			}
			y = y*10 + int(ch-'0')
		}
		return y
	}
	return 0
}

func (s *scryfallCard) toExternal() *ExternalCard {
	imageURL := ""
	if s.ImageURIs != nil {
		if u, ok := s.ImageURIs["normal"]; ok {
			imageURL = u
		}
	} else if len(s.CardFaces) > 0 && s.CardFaces[0].ImageURIs != nil {
		if u, ok := s.CardFaces[0].ImageURIs["normal"]; ok {
			imageURL = u
		}
	}

	printedName := s.PrintedName
	if printedName == "" && len(s.CardFaces) > 0 {
		printedName = s.CardFaces[0].PrintedName
	}

	colors := s.Colors
	if len(colors) == 0 && len(s.ColorIdentity) > 0 {
		colors = s.ColorIdentity
	}

	return &ExternalCard{
		ID:          s.ID,
		Name:        s.Name,
		PrintedName: printedName,
		Set:         strings.ToUpper(s.Set),
		SetName:     s.SetName,
		Rarity:      NormalizeRarity(s.Rarity),
		Type:        s.TypeLine,
		PrintedType: s.PrintedTypeLine,
		ManaCost:    s.ManaCost,
		Colors:      colors,
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
		FullArt:     s.FullArt,
		Year:        yearFromReleasedAt(s.ReleasedAt),
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
	const maxAttempts = 3
	var delay time.Duration
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if delay > 0 {
			time.Sleep(delay)
		}

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

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			delay = 10 * time.Second // rate limit: aguarda 10s antes do retry
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, nil // 404 ou outro erro não-transitório
		}

		var card scryfallCard
		err = json.NewDecoder(resp.Body).Decode(&card)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if card.ID == "" {
			return nil, nil
		}
		return card.toExternal(), nil
	}
	return nil, nil // ainda em rate limit após todas as tentativas
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

// FetchSetCards busca todas as cartas de um set, paginando automaticamente.
// Se lang != "en", busca a versão localizada de cada carta individualmente.
func (c *Client) FetchSetCards(setCode, lang string) ([]*ExternalCard, error) {
	setCode = strings.ToLower(strings.TrimSpace(setCode))
	langCode := toLangCode(lang)

	searchURL := fmt.Sprintf(
		"https://api.scryfall.com/cards/search?include_extras=true&include_variations=true&order=set&q=%s&unique=prints",
		url.QueryEscape("e:"+setCode),
	)

	var enCards []scryfallCard
	for searchURL != "" {
		page, err := c.fetchPage(searchURL)
		if err != nil {
			return nil, err
		}
		enCards = append(enCards, page.Data...)
		if page.HasMore && page.NextPage != "" {
			searchURL = page.NextPage
			time.Sleep(100 * time.Millisecond)
		} else {
			searchURL = ""
		}
	}

	result := make([]*ExternalCard, 0, len(enCards))
	for _, sc := range enCards {
		if langCode != "en" && sc.Set != "" && sc.CollectorNumber != "" {
			time.Sleep(75 * time.Millisecond)
			langEndpoint := fmt.Sprintf("%s/%s/%s/%s", apiBase, sc.Set, sc.CollectorNumber, langCode)
			if langCard, _ := c.fetch(langEndpoint); langCard != nil {
				result = append(result, langCard)
				continue
			}
		}
		result = append(result, sc.toExternal())
	}

	return result, nil
}

func (c *Client) fetchPage(endpoint string) (*scryfallList, error) {
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
		return nil, fmt.Errorf("scryfall retornou %d para %s", resp.StatusCode, endpoint)
	}

	var list scryfallList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	return &list, nil
}

// SearchByName busca uma carta pelo nome no Scryfall.
// Tenta exact match primeiro (com set hint se fornecido), depois fuzzy como fallback.
// Se lang != "en", tenta buscar a versão localizada.
func (c *Client) SearchByName(name, setCode, lang string) (*ExternalCard, error) {
	langCode := toLangCode(lang)
	exactBase := "https://api.scryfall.com/cards/named?exact=" + url.QueryEscape(name)

	var enCard *ExternalCard
	if setCode != "" {
		enCard, _ = c.fetch(exactBase + "&set=" + url.QueryEscape(strings.ToLower(strings.TrimSpace(setCode))))
	}
	if enCard == nil {
		enCard, _ = c.fetch(exactBase)
	}
	if enCard == nil {
		// Pausa antes do fuzzy para não ultrapassar o rate limit do Scryfall.
		time.Sleep(110 * time.Millisecond)
		fuzzy := "https://api.scryfall.com/cards/named?fuzzy=" + url.QueryEscape(name)
		var err error
		enCard, err = c.fetch(fuzzy)
		if err != nil || enCard == nil {
			return nil, err
		}
	}

	if langCode != "en" {
		time.Sleep(75 * time.Millisecond)
		langEndpoint := fmt.Sprintf("%s/%s/%s/%s", apiBase, strings.ToLower(enCard.Set), enCard.Number, langCode)
		if langCard, _ := c.fetch(langEndpoint); langCard != nil {
			return langCard, nil
		}
	}

	return enCard, nil
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

// SearchToken busca um token da Scryfall pelo set base (sem o 't') e número.
// A URL de busca é construída como: /cards/t{setCode}/{number}
func (c *Client) SearchToken(setCode, number string) (*ExternalToken, error) {
	if setCode == "" || number == "" {
		return nil, nil
	}
	endpoint := fmt.Sprintf("%s/t%s/%s", apiBase, strings.ToLower(strings.TrimSpace(setCode)), strings.TrimSpace(number))

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

	var sc scryfallCard
	if err := json.NewDecoder(resp.Body).Decode(&sc); err != nil {
		return nil, err
	}
	if sc.ID == "" {
		return nil, nil
	}

	tok := &ExternalToken{
		ID:               sc.ID,
		Name:             sc.Name,
		TypeLine:         sc.TypeLine,
		OracleText:       sc.OracleText,
		Power:            sc.Power,
		Toughness:        sc.Toughness,
		Colors:           sc.Colors,
		SetCode:          strings.ToUpper(sc.Set),
		CollectionNumber: sc.CollectorNumber,
		Artist:           sc.Artist,
	}
	if sc.ImageURIs != nil {
		tok.ImageURL = sc.ImageURIs["normal"]
	}

	if len(sc.CardFaces) >= 2 {
		tok.DoubleFaced = true
		f0, f1 := sc.CardFaces[0], sc.CardFaces[1]
		if f0.Name != "" {
			tok.Name = f0.Name
		}
		if len(f0.Colors) > 0 {
			tok.Colors = f0.Colors
		}
		tok.TypeLine = f0.TypeLine
		tok.OracleText = f0.OracleText
		tok.Power = f0.Power
		tok.Toughness = f0.Toughness
		if f0.ImageURIs != nil {
			tok.ImageURL = f0.ImageURIs["normal"]
		}
		tok.BackName = f1.Name
		tok.BackTypeLine = f1.TypeLine
		tok.BackOracleText = f1.OracleText
		tok.BackPower = f1.Power
		tok.BackToughness = f1.Toughness
		if f1.ImageURIs != nil {
			tok.BackImageURL = f1.ImageURIs["normal"]
		}
	}

	return tok, nil
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
