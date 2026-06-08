package battles

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo battleRepo
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) List(c *gin.Context) {
	list, err := h.repo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar batalhas"})
		return
	}
	if list == nil {
		list = []Battle{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) Create(c *gin.Context) {
	var input BattleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}
	id, err := h.repo.Create(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao registrar batalha"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Batalha registrada"})
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.repo.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover batalha"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Batalha removida"})
}
