package cards

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"magic-collection-api/internal/ai"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service  *Service
	aiClient *ai.Client
}

func NewHandler(service *Service, aiClient *ai.Client) *Handler {
	return &Handler{service: service, aiClient: aiClient}
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

	var deckIDFilter *int
	if deckStr := c.Query("deck_id"); deckStr != "" {
		if v, err := strconv.Atoi(deckStr); err == nil {
			deckIDFilter = &v
		}
	}

	params := ListParams{
		Search:       c.Query("q"),
		Page:         page,
		PageSize:     pageSize,
		Sort:         c.DefaultQuery("sort", "name"),
		Order:        c.DefaultQuery("order", "asc"),
		DeckIDFilter: deckIDFilter,
		FoilOnly:     c.Query("foil") == "1",
		RarityFilter: c.Query("rarity"),
	}

	result, err := h.service.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao listar cartas",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":           result.Cards,
		"total":          result.Total,
		"total_quantity": result.TotalQuantity,
		"page":           result.Page,
		"page_size":      result.PageSize,
		"total_pages":    result.TotalPages,
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

func (h *Handler) Update(c *gin.Context) {
	var input UpdateCardInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}
	if err := h.service.Update(c.Param("id"), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar carta"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Carta atualizada com sucesso"})
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

func (h *Handler) SetDeck(c *gin.Context) {
	var input struct {
		DeckID int `json:"deck_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}
	if err := h.service.SetDeck(c.Param("id"), input.DeckID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atribuir deck"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) Preview(c *gin.Context) {
	var input CreateCardInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}
	ext, err := h.service.Preview(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ext == nil {
		c.JSON(http.StatusOK, gin.H{"found": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "card": ext})
}

func (h *Handler) RefreshImages(c *gin.Context) {
	result, err := h.service.RefreshImages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) RefreshPrices(c *gin.Context) {
	emptyOnly := c.Query("empty_only") == "1"
	result, err := h.service.RefreshPrices(emptyOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) Stats(c *gin.Context) {
	stats, err := h.service.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular estatísticas"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *Handler) Export(c *gin.Context) {
	cards, err := h.service.ExportAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao exportar"})
		return
	}
	c.JSON(http.StatusOK, cards)
}

func (h *Handler) SuggestDecks(c *gin.Context) {
	cards, err := h.service.GetCardsForDeckBuilder()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar cartas: " + err.Error()})
		return
	}
	if len(cards) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhuma carta sem deck encontrada"})
		return
	}

	prompt := buildDeckBuilderPrompt(cards)
	analysis, err := h.aiClient.Complete(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro na API de IA: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analysis": analysis, "card_count": len(cards)})
}

func buildDeckBuilderPrompt(cards []DeckBuilderCard) string {
	cats := map[string][]string{
		"Criatura": {}, "Planeswalker": {}, "Feitiço": {},
		"Mágica Imediata": {}, "Artefato": {}, "Encantamento": {},
		"Terreno": {}, "Outros": {},
	}
	order := []string{"Criatura", "Planeswalker", "Feitiço", "Mágica Imediata", "Artefato", "Encantamento", "Terreno", "Outros"}

	totalQty := 0
	for _, c := range cards {
		totalQty += c.Quantity
		t := strings.ToLower(c.Type)
		var group string
		switch {
		case strings.Contains(t, "creature") || strings.Contains(t, "criatura"):
			group = "Criatura"
		case strings.Contains(t, "planeswalker"):
			group = "Planeswalker"
		case strings.Contains(t, "sorcery") || strings.Contains(t, "feitiço"):
			group = "Feitiço"
		case strings.Contains(t, "instant") || strings.Contains(t, "imediata"):
			group = "Mágica Imediata"
		case strings.Contains(t, "artifact") || strings.Contains(t, "artefato"):
			group = "Artefato"
		case strings.Contains(t, "enchantment") || strings.Contains(t, "encantamento"):
			group = "Encantamento"
		case strings.Contains(t, "land") || strings.Contains(t, "terreno"):
			group = "Terreno"
		default:
			group = "Outros"
		}
		entry := fmt.Sprintf("%dx %s", c.Quantity, c.Name)
		if c.ManaCost != "" {
			entry += " " + c.ManaCost
		}
		if c.Rarity != "" {
			entry += " [" + c.Rarity + "]"
		}
		cats[group] = append(cats[group], entry)
	}

	var cardList strings.Builder
	for _, cat := range order {
		if len(cats[cat]) > 0 {
			cardList.WriteString(fmt.Sprintf("\n**%s (%d únicos):**\n", cat, len(cats[cat])))
			for _, entry := range cats[cat] {
				cardList.WriteString("- " + entry + "\n")
			}
		}
	}

	return fmt.Sprintf(`Você é um especialista em Magic: The Gathering com profundo conhecimento de deck-building.

O jogador possui %d cartas (%d únicas) SEM DECK atribuído. Analise-as e sugira como montar o(s) melhor(es) deck(s) possível(is).

**REGRA PRINCIPAL:** Priorize decks de 60 cartas (Casual, Standard, Pioneer, Modern ou Legacy). Só sugira Commander (100 cartas) se for claramente o melhor aproveitamento do que o jogador tem.
**RESTRIÇÃO IMPORTANTE:** Use APENAS as cartas da lista abaixo, nas quantidades disponíveis.

**Cartas disponíveis:**
%s

Forneça uma análise completa em markdown com as seguintes seções:

## 📊 Análise Geral
Quantos decks de 60 cartas é possível montar? Avalie a viabilidade geral. Se não for possível montar decks completos, explique o que falta e por quê.

---
(Para cada deck sugerido, use o template abaixo:)

## 🃏 Deck [número]: [Nome do Archetype]
**Formato:** [formato ideal — ex: Casual 60, Modern, Standard, Commander]
**Identidade de Cores:** [cores do deck]
**Viabilidade:** [nota 1-10] — [frase justificando]

### 📋 Lista de Cartas
(liste apenas cartas da lista disponível com as quantidades usadas, ex: 4x Lightning Bolt)

### 🎮 Como Jogar
Passo a passo da estratégia: mulligan ideal, primeiros turnos, mid-game e como fechar o jogo.

### 🔗 Sinergias Principais
Quais cartas combinam entre si e por quê.

### ✅ Vantagens
Pontos fortes do deck.

### ❌ Desvantagens e Limitações
O que falta, pontos fracos, o que comprar para completar.

---

## 💡 Recomendações Finais
O que o jogador deveria adquirir para completar ou melhorar os decks sugeridos.`,
		totalQty, len(cards), cardList.String())
}
