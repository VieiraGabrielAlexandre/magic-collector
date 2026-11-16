package trade

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Gainers(c *gin.Context) {
	minPct := 10.0
	if v := c.Query("min_pct"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			minPct = f
		}
	}
	data, err := h.repo.Gainers(minPct, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if data == nil {
		data = []CardMovement{}
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) Losers(c *gin.Context) {
	maxPct := -10.0
	if v := c.Query("max_pct"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			maxPct = f
		}
	}
	data, err := h.repo.Losers(maxPct, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if data == nil {
		data = []CardMovement{}
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) Stable(c *gin.Context) {
	data, err := h.repo.Stable(50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if data == nil {
		data = []CardMovement{}
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) Suggestions(c *gin.Context) {
	minPrice := 2.0
	if v := c.Query("min_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			minPrice = f
		}
	}
	data, err := h.repo.TradeSuggestions(minPrice, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if data == nil {
		data = []TradeSuggestion{}
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) PriceSpikes(c *gin.Context) {
	threshold := 20.0
	if v := c.Query("threshold"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			threshold = f
		}
	}
	data, err := h.repo.PriceSpikes(threshold)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if data == nil {
		data = []PriceSpike{}
	}
	c.JSON(http.StatusOK, gin.H{"spikes": data, "threshold": threshold, "count": len(data)})
}

func (h *Handler) PortfolioHistory(c *gin.Context) {
	data, err := h.repo.PortfolioHistory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if data == nil {
		data = []PortfolioPoint{}
	}
	c.JSON(http.StatusOK, gin.H{"points": data})
}
