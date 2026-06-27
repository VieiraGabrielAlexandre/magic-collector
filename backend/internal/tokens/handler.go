package tokens

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	list, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar tokens"})
		return
	}
	if list == nil {
		list = []Token{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) Preview(c *gin.Context) {
	var input CreateTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}
	ext, err := h.service.Preview(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar token na Scryfall"})
		return
	}
	if ext == nil {
		c.JSON(http.StatusOK, gin.H{"found": false, "token": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "token": ext})
}

func (h *Handler) Create(c *gin.Context) {
	var input CreateTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}
	id, err := h.service.Create(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao cadastrar token"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Token cadastrado com sucesso"})
}

func (h *Handler) UpdateQuantity(c *gin.Context) {
	var body struct {
		Quantity int `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Quantity < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantidade inválida"})
		return
	}
	if err := h.service.UpdateQuantity(c.Param("id"), body.Quantity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar quantidade"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Quantidade atualizada"})
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if _, err := strconv.Atoi(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Token removido"})
}
