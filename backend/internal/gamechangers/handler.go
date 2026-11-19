package gamechangers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	cards, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar game changers"})
		return
	}
	if cards == nil {
		cards = []GameChanger{}
	}
	count, syncedAt := h.service.SyncMeta()
	c.JSON(http.StatusOK, gin.H{
		"data":      cards,
		"total":     count,
		"synced_at": syncedAt,
	})
}

func (h *Handler) Sync(c *gin.Context) {
	result, err := h.service.Sync()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
