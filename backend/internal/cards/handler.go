package cards

import (
	"encoding/json"
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
		ColorsFilter: c.Query("colors"),
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

func (h *Handler) NormalizeRarities(c *gin.Context) {
	result, err := h.service.NormalizeRarities()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) RefreshColors(c *gin.Context) {
	result, err := h.service.RefreshColors()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListColors(c *gin.Context) {
	combos, err := h.service.ListColorCombos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, combos)
}

func (h *Handler) UpdateQuantity(c *gin.Context) {
	var body struct {
		Quantity int `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Quantity < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantidade inválida"})
		return
	}
	if err := h.service.SetQuantity(c.Param("id"), body.Quantity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Preview(c *gin.Context) {
	var input PreviewCardInput
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

// SuggestDecksInput parametriza a geração de deck pela IA.
type SuggestDecksInput struct {
	Format    string `json:"format"`    // "auto" | "casual60" | "commander"
	Goal      string `json:"goal"`      // "fun" | "competitive"
	Colors    string `json:"colors"`    // "W,U,B" ou ""
	Revaluate bool   `json:"revaluate"` // true = sugerir cartas/estratégia diferentes
}

// deckSuggestion é o bloco estruturado que a IA devolve dentro dos delimitadores.
type deckSuggestion struct {
	Nome      string `json:"nome"`
	Cores     string `json:"cores"`
	Commander bool   `json:"commander"`
	Descricao string `json:"descricao"`
	Lista     string `json:"lista"`
}

// parseDeckBuilderOutput extrai o bloco JSON da IA e separa a análise em markdown.
// Retorna (suggestion, analysis, errIA): errIA é preenchido quando a IA declara impossibilidade.
func parseDeckBuilderOutput(raw string) (*deckSuggestion, string, string) {
	// Verifica se a IA emitiu um bloco de erro (impossível montar deck)
	const errStart = "<<<ERRO>>>"
	const errEnd = "<<<FIM_ERRO>>>"
	if es := strings.Index(raw, errStart); es != -1 {
		if ee := strings.Index(raw, errEnd); ee > es {
			jsonStr := strings.TrimSpace(raw[es+len(errStart) : ee])
			var errObj struct {
				Motivo string `json:"motivo"`
			}
			if json.Unmarshal([]byte(jsonStr), &errObj) == nil && errObj.Motivo != "" {
				return nil, "", errObj.Motivo
			}
			return nil, "", strings.TrimSpace(raw[es+len(errStart) : ee])
		}
	}

	const startTag = "<<<DECK>>>"
	const endTag = "<<<FIM_DECK>>>"
	start := strings.Index(raw, startTag)
	end := strings.Index(raw, endTag)
	if start == -1 || end == -1 || end <= start {
		return nil, raw, ""
	}
	jsonStr := strings.TrimSpace(raw[start+len(startTag) : end])
	var s deckSuggestion
	if err := json.Unmarshal([]byte(jsonStr), &s); err != nil {
		return nil, raw, ""
	}
	analysis := strings.TrimSpace(raw[end+len(endTag):])
	return &s, analysis, ""
}

func (h *Handler) SuggestDecks(c *gin.Context) {
	var input SuggestDecksInput
	_ = c.ShouldBindJSON(&input) // campos opcionais — ignora erro de bind vazio
	if input.Format == "" {
		input.Format = "auto"
	}
	if input.Goal == "" {
		input.Goal = "fun"
	}

	cards, err := h.service.GetCardsForDeckBuilder()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar cartas: " + err.Error()})
		return
	}
	if len(cards) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhuma carta sem deck encontrada"})
		return
	}

	prompt := buildDeckBuilderPrompt(cards, input)
	raw, err := h.aiClient.Complete(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro na API de IA: " + err.Error()})
		return
	}

	suggestion, analysis, errIA := parseDeckBuilderOutput(raw)

	resp := gin.H{"analysis": analysis, "card_count": len(cards)}
	if errIA != "" {
		resp["error_ia"] = errIA
	}
	if suggestion != nil {
		resp["deck_name"] = suggestion.Nome
		resp["deck_colors"] = suggestion.Cores
		resp["deck_commander"] = suggestion.Commander
		resp["deck_list"] = suggestion.Lista
		resp["deck_description"] = suggestion.Descricao
	}
	c.JSON(http.StatusOK, resp)
}

func buildDeckBuilderPrompt(cards []DeckBuilderCard, input SuggestDecksInput) string {
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
			entry += " (" + c.ManaCost + ")"
		}
		if c.Colors != "" && c.Colors != "[]" && c.Colors != "null" {
			entry += " {" + c.Colors + "}"
		}
		if c.Rarity != "" {
			entry += " [" + c.Rarity + "]"
		}
		cats[group] = append(cats[group], entry)
	}

	var cardList strings.Builder
	for _, cat := range order {
		if len(cats[cat]) > 0 {
			cardList.WriteString(fmt.Sprintf("\n### %s (%d únicos)\n", cat, len(cats[cat])))
			for _, entry := range cats[cat] {
				cardList.WriteString("- " + entry + "\n")
			}
		}
	}

	// ── Instruções de formato ─────────────────────────────────────
	var formatInstr string
	switch input.Format {
	case "commander":
		formatInstr = "Monte UM deck **Commander de EXATAMENTE 100 cartas** (o Comandante é incluído nessa contagem).\n" +
			"- Declare a carta Comandante na primeira linha da lista (ex: `Commander\\n1 [Nome]`)\n" +
			"- Máximo 1 cópia de cada carta não-básica (singleton)\n" +
			"- Inclua 37-40 terrenos (pode adicionar terrenos básicos não listados)\n" +
			"- As cores do deck são definidas pela identidade de cor do Comandante\n" +
			"- **ANTES DE GERAR:** conte mentalmente quantas cartas você escolherá; o total DEVE ser exatamente 100. Se for impossível atingir 100 com as cartas disponíveis, emita `<<<ERRO>>>` em vez de `<<<DECK>>>`.\n" +
			`- campo "commander": true no JSON`
	case "casual60":
		formatInstr = "Monte UM deck **de EXATAMENTE 60 cartas** (Casual / Standard / Modern). Máximo tolerado: 62.\n" +
			"- Até 4 cópias de cartas não-básicas\n" +
			"- Inclua 20-24 terrenos básicos (pode adicionar básicos não listados)\n" +
			"- Curva de mana: pico em 2-3 mana, poucas cartas com custo 5+\n" +
			"- **ANTES DE GERAR:** conte mentalmente quantas cartas você escolherá; o total DEVE ser 60, 61 ou 62. Se for impossível atingir esse intervalo, emita `<<<ERRO>>>` em vez de `<<<DECK>>>`.\n" +
			`- campo "commander": false no JSON`
	default: // auto
		formatInstr = "Escolha o formato mais adequado para as cartas disponíveis:\n" +
			"- Prefira **60 cartas** (Casual/Modern) se houver sinergias suficientes\n" +
			"- Use **Commander (100 cartas)** apenas se as cartas claramente favorecem esse formato\n" +
			"- Inclua terrenos básicos necessários mesmo que não estejam na lista\n" +
			"- **ANTES DE GERAR:** verifique a contagem: Commander = exatamente 100, Casual60 = 60 a 62. Se for impossível atingir esses valores, emita `<<<ERRO>>>` em vez de `<<<DECK>>>`."
	}

	// ── Instruções de objetivo ────────────────────────────────────
	var goalInstr string
	if input.Goal == "competitive" {
		goalInstr = "**Objetivo: COMPETITIVO** — maximize consistência e eficiência. " +
			"Priorize cartas com baixo custo de mana, remoções e geração de vantagem. " +
			"O deck deve ser o mais forte possível com as cartas disponíveis."
	} else {
		goalInstr = "**Objetivo: DIVERSÃO** — priorize combos interessantes, sinergias temáticas e " +
			"interações criativas. O deck não precisa ser o mais eficiente, mas deve ser divertido de jogar."
	}

	// ── Preferência de cores ──────────────────────────────────────
	var colorInstr string
	if input.Colors != "" {
		colorInstr = fmt.Sprintf("\n**Cores preferidas:** %s — monte o deck preferencialmente nessas cores. "+
			"Ignore cartas de outras cores, exceto se forem essenciais para a estratégia.\n", input.Colors)
	}

	// ── Instrução de re-avaliação ─────────────────────────────────
	var revaluateInstr string
	if input.Revaluate {
		revaluateInstr = "\n**⚠️ RE-AVALIAÇÃO:** Sugira uma estratégia DIFERENTE da avaliação anterior. " +
			"Explore outras sinergias, outro archetype ou uma combinação de cores diferente.\n"
	}

	return fmt.Sprintf(`Você é um especialista em Magic: The Gathering com conhecimento profundo em deck-building.

## 🃏 CARTAS DISPONÍVEIS (%d cópias totais, %d únicas sem deck)
%s

## ⚙️ PARÂMETROS
%s
%s%s%s
## 📐 REGRAS OBRIGATÓRIAS
1. Use SOMENTE cartas da lista acima, respeitando as quantidades disponíveis
2. **Terrenos básicos** (Mountain, Island, Swamp, Plains, Forest) podem ser adicionados livremente mesmo sem estarem na lista
3. **Coerência de mana:** calcule a proporção de terrenos por cor baseado no custo de mana das cartas; não inclua terrenos de cores desnecessárias
4. **Sinergia obrigatória:** cada carta não-terreno deve contribuir diretamente para a estratégia central
5. **Monte APENAS UM deck** — o melhor possível com os parâmetros dados

## 📤 FORMATO DE SAÍDA — DUAS OPÇÕES (emita APENAS UMA)

**OPÇÃO A — Deck válido (contagem correta):**

<<<DECK>>>
{
  "nome": "Nome criativo e temático em português",
  "cores": "X,Y",
  "commander": false,
  "descricao": "2-3 frases descrevendo estilo e estratégia do deck.",
  "lista": "Commander\n1 Nome do Comandante\n\nCriaturas\n4 Carta Um\n3 Carta Dois\n\nFeitiços\n4 Carta Três\n\nTerrenos\n20 Forest\n4 Mountain"
}
<<<FIM_DECK>>>

Regras do campo "lista":
- Linhas de seção sem número: "Commander", "Criaturas", "Feitiços", "Mágicas Imediatas", "Artefatos", "Encantamentos", "Terrenos"
- Cartas: "N Nome Exato da Carta" (ex: "4 Lightning Bolt", "1 Sol Ring")
- Terrenos básicos no final; use os nomes exatos em inglês: Mountain, Island, Swamp, Plains, Forest
- Sem comentários, sem preços, sem códigos de set dentro da lista
- **A soma de todas as quantidades N da lista deve bater exatamente com o total exigido pelo formato**

**OPÇÃO B — Impossível montar deck válido:**

<<<ERRO>>>
{"motivo": "Explique em detalhes por que não é possível montar um deck com a contagem exigida usando as cartas disponíveis e os parâmetros fornecidos"}
<<<FIM_ERRO>>>

Depois do bloco (A ou B), escreva a análise em markdown:

## 📊 Análise
Estratégia central do deck em 2-3 parágrafos.

## 🔗 Sinergias Principais
As 3-5 combinações de cartas mais importantes e por quê funcionam.

## 🎮 Como Jogar
Turnos 1-3, mid-game e como fechar o jogo.

## ✅ Pontos Fortes

## ❌ Limitações e o que comprar para completar`,
		totalQty, len(cards),
		cardList.String(),
		formatInstr, goalInstr, colorInstr, revaluateInstr)
}
