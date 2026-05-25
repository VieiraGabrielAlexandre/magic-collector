package cards

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

func (h *Handler) Create(c *gin.Context) {
	var input CreateCardInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos",
		})
		return
	}

	id, err := h.service.Create(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao cadastrar carta",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      id,
		"message": "Carta cadastrada com sucesso",
	})
}

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	params := ListParams{
		Search:   c.Query("q"),
		Page:     page,
		PageSize: pageSize,
		Sort:     c.DefaultQuery("sort", "name"),
		Order:    c.DefaultQuery("order", "asc"),
	}

	result, err := h.service.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao listar cartas",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        result.Cards,
		"total":       result.Total,
		"page":        result.Page,
		"page_size":   result.PageSize,
		"total_pages": result.TotalPages,
	})
}

func (h *Handler) GetByID(c *gin.Context) {
	card, err := h.service.GetByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Carta não encontrada",
		})
		return
	}

	c.JSON(http.StatusOK, card)
}

func (h *Handler) Delete(c *gin.Context) {
	err := h.service.Delete(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao remover carta",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Carta removida com sucesso",
	})
}
