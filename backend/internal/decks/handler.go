package decks

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *gin.Context) {
	decks, err := h.svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar decks"})
		return
	}
	if decks == nil {
		decks = []Deck{}
	}
	c.JSON(http.StatusOK, decks)
}

func (h *Handler) Create(c *gin.Context) {
	var input DeckInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}
	id, err := h.svc.Create(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar deck"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Deck criado com sucesso"})
}

func (h *Handler) Update(c *gin.Context) {
	var input DeckInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}
	if err := h.svc.Update(c.Param("id"), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar deck"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deck atualizado com sucesso"})
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover deck"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deck removido com sucesso"})
}

func (h *Handler) FetchIcon(c *gin.Context) {
	iconURI, err := h.svc.FetchIcon(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar ícone"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"icon_uri": iconURI})
}

func (h *Handler) Evaluate(c *gin.Context) {
	deck, err := h.svc.EvaluateDeck(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deck)
}
