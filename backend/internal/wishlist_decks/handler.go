package wishlist_decks

import (
	"net/http"
	"strconv"

	"magic-collection-api/internal/mtgapi"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo      *Repository
	mtgClient *mtgapi.Client
}

func NewHandler(repo *Repository, mtgClient *mtgapi.Client) *Handler {
	return &Handler{repo: repo, mtgClient: mtgClient}
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

	result, err := h.repo.AddCard(deckID, input, card.Name, printedName, imageURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
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
