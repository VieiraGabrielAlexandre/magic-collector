package wishlist_decks

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"magic-collection-api/internal/cards"
	"magic-collection-api/internal/decks"
	"magic-collection-api/internal/mtgapi"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo      *Repository
	mtgClient *mtgapi.Client
	cardRepo  *cards.Repository
	deckRepo  *decks.Repository
}

func NewHandler(repo *Repository, mtgClient *mtgapi.Client, cardRepo *cards.Repository, deckRepo *decks.Repository) *Handler {
	return &Handler{repo: repo, mtgClient: mtgClient, cardRepo: cardRepo, deckRepo: deckRepo}
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.repo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) Create(c *gin.Context) {
	var input CreateWishlistDeckInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nome é obrigatório"})
		return
	}
	if input.Format == "" {
		input.Format = "Commander"
	}
	item, err := h.repo.Create(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) Acquire(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	item, err := h.repo.Acquire(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}

func (h *Handler) SearchCards(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parâmetro q é obrigatório"})
		return
	}
	cards, err := h.mtgClient.SearchPrintings(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if cards == nil {
		cards = []*mtgapi.ExternalCard{}
	}
	c.JSON(http.StatusOK, cards)
}

func (h *Handler) ListCards(c *gin.Context) {
	deckID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	cards, err := h.repo.ListCards(deckID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cards)
}

func (h *Handler) AddCard(c *gin.Context) {
	deckID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	var input AddDeckCardInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.SetCode == "" || input.CollectionNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "set_code e collection_number são obrigatórios"})
		return
	}

	// Busca na Scryfall
	card, err := h.mtgClient.Search(input.SetCode, input.CollectionNumber, "", "")
	if err != nil || card == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "carta não encontrada na Scryfall"})
		return
	}

	imageURI := card.ImageURL
	printedName := card.PrintedName
	if printedName == "" {
		printedName = card.Name
	}

	result, err := h.repo.AddCard(deckID, input, card.Name, printedName, card.Type, imageURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

type deckListLine struct {
	quantity int
	name     string
}

func parseDeckList(raw string) []deckListLine {
	var result []deckListLine
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		qtyStr := strings.TrimRight(strings.TrimRight(parts[0], "x"), "X")
		qty, err := strconv.Atoi(qtyStr)
		if err != nil || qty <= 0 {
			continue
		}
		name := strings.TrimSpace(parts[1])
		// Remove set/collector annotations: "Forest (M19) 266" → "Forest"
		if idx := strings.Index(name, " ("); idx != -1 {
			name = strings.TrimSpace(name[:idx])
		}
		if name != "" {
			result = append(result, deckListLine{quantity: qty, name: name})
		}
	}
	return result
}

func (h *Handler) ImportList(c *gin.Context) {
	deckID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	var input struct {
		DeckList string `json:"deck_list"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.DeckList) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deck_list é obrigatório"})
		return
	}
	lines := parseDeckList(input.DeckList)
	if len(lines) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nenhuma carta encontrada na lista"})
		return
	}

	type importResult struct {
		Added   int      `json:"added"`
		Skipped []string `json:"skipped"`
	}
	res := importResult{Skipped: []string{}}

	for i, line := range lines {
		if i > 0 {
			time.Sleep(100 * time.Millisecond)
		}
		card, err := h.mtgClient.SearchOldestPrinting(line.name)
		if err != nil || card == nil {
			res.Skipped = append(res.Skipped, line.name)
			continue
		}
		printedName := card.PrintedName
		if printedName == "" {
			printedName = card.Name
		}
		dci := AddDeckCardInput{SetCode: card.Set, CollectionNumber: card.Number}
		failed := false
		for j := 0; j < line.quantity; j++ {
			if _, err := h.repo.AddCard(deckID, dci, card.Name, printedName, card.Type, card.ImageURL); err != nil {
				if !failed {
					res.Skipped = append(res.Skipped, line.name)
					failed = true
				}
				break
			}
			res.Added++
		}
	}

	c.JSON(http.StatusOK, res)
}

func rarityToCode(r string) string {
	switch strings.ToLower(r) {
	case "common":
		return "C"
	case "uncommon":
		return "U"
	case "rare":
		return "R"
	case "mythic":
		return "M"
	case "land":
		return "L"
	case "token":
		return "T"
	default:
		return "C"
	}
}

func (h *Handler) AcquireCard(c *gin.Context) {
	deckID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	cardID, err := strconv.ParseInt(c.Param("card_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "card_id inválido"})
		return
	}

	// Busca o deck desejado
	wishDeck, err := h.repo.getByID(deckID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "deck desejado não encontrado"})
		return
	}

	// Busca a carta desejada
	wishCard, err := h.repo.GetCard(deckID, cardID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "carta não encontrada"})
		return
	}

	// Cria o deck real se ainda não existe
	linkedDeckID := wishDeck.LinkedDeckID
	if linkedDeckID == 0 {
		isCommander := strings.EqualFold(wishDeck.Format, "Commander")
		newDeckID, err := h.deckRepo.Create(decks.DeckInput{
			Name:      wishDeck.Name,
			Colors:    wishDeck.Colors,
			Commander: isCommander,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao criar deck: " + err.Error()})
			return
		}
		linkedDeckID = newDeckID
		if err := h.repo.SetLinkedDeck(deckID, linkedDeckID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao vincular deck: " + err.Error()})
			return
		}
	}

	// Busca dados da carta na Scryfall
	ext, err := h.mtgClient.Search(wishCard.SetCode, wishCard.CollectionNumber, "", "")
	if err != nil || ext == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "carta não encontrada na Scryfall"})
		return
	}

	// Monta JSON de cores
	colorsJSON, _ := json.Marshal(ext.Colors)
	colorsStr := string(colorsJSON)
	colorDisplay := cards.ColorsJSONToDisplay(colorsStr)

	// Cria a carta na coleção
	newCard := cards.Card{
		MTGID:            ext.ID,
		Name:             ext.Name,
		Color:            colorDisplay,
		Colors:           colorsStr,
		Type:             ext.Type,
		CollectionNumber: ext.Number,
		SetCode:          strings.ToUpper(ext.Set),
		Rarity:           rarityToCode(ext.Rarity),
		ManaCost:         ext.ManaCost,
		Artist:           ext.Artist,
		Foil:             wishCard.Foil,
		Quantity:         1,
		Condition:        "near_mint",
		DeckID:           int(linkedDeckID),
		ImageURL:         ext.ImageURL,
		Year:             ext.Year,
		DoubleFaced:      ext.DoubleFaced,
		ImageURLBack:     ext.BackImageURL,
		FullArt:          ext.FullArt,
	}
	newCardID, err := h.cardRepo.Create(newCard)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao criar carta: " + err.Error()})
		return
	}

	// Marca a carta do deck desejado como adquirida
	_ = h.repo.MarkCardAcquired(deckID, cardID)

	c.JSON(http.StatusOK, AcquireCardResult{DeckID: linkedDeckID, CardID: newCardID})
}

func (h *Handler) ToggleCardAcquired(c *gin.Context) {
	deckID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	cardID, err := strconv.ParseInt(c.Param("card_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "card_id inválido"})
		return
	}
	card, err := h.repo.ToggleCardAcquired(deckID, cardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, card)
}

func (h *Handler) DeleteCard(c *gin.Context) {
	deckID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	cardID, err := strconv.ParseInt(c.Param("card_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "card_id inválido"})
		return
	}
	if err := h.repo.DeleteCard(deckID, cardID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": cardID})
}
