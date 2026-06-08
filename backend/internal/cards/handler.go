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
	service  cardService
	aiClient aiCompleter
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
		FullArtOnly:  c.Query("full_art") == "1",
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

type CardRole struct {
	Nome  string `json:"nome"`
	Papel string `json:"papel"`
}

type TerrenoInfo struct {
	Total  int    `json:"total"`
	Motivo string `json:"motivo"`
}

type CardRoles struct {
	NaoTerrenos []CardRole  `json:"nao_terrenos"`
	Terrenos    TerrenoInfo `json:"terrenos"`
}

// parseDeckBuilderOutput extrai o bloco JSON da IA e separa a análise em markdown.
// Retorna (suggestion, analysis, errIA, roles): errIA é preenchido quando a IA declara impossibilidade.
func parseDeckBuilderOutput(raw string) (*deckSuggestion, string, string, *CardRoles) {
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
				return nil, "", errObj.Motivo, nil
			}
			return nil, "", strings.TrimSpace(raw[es+len(errStart) : ee]), nil
		}
	}

	const startTag = "<<<DECK>>>"
	const endTag = "<<<FIM_DECK>>>"
	start := strings.Index(raw, startTag)
	end := strings.Index(raw, endTag)
	if start == -1 || end == -1 || end <= start {
		return nil, raw, "", nil
	}
	jsonStr := strings.TrimSpace(raw[start+len(startTag) : end])
	var s deckSuggestion
	if err := json.Unmarshal([]byte(jsonStr), &s); err != nil {
		return nil, raw, "", nil
	}

	afterDeck := raw[end+len(endTag):]

	// Extrai o bloco de papéis das cartas, se presente
	const cartasStart = "<<<CARTAS>>>"
	const cartasEnd = "<<<FIM_CARTAS>>>"
	var roles *CardRoles
	if cs := strings.Index(afterDeck, cartasStart); cs != -1 {
		if ce := strings.Index(afterDeck, cartasEnd); ce > cs {
			cartasJSON := strings.TrimSpace(afterDeck[cs+len(cartasStart) : ce])
			var r CardRoles
			if json.Unmarshal([]byte(cartasJSON), &r) == nil {
				roles = &r
			}
			// Remove o bloco CARTAS da análise
			afterDeck = afterDeck[:cs] + afterDeck[ce+len(cartasEnd):]
		}
	}

	analysis := strings.TrimSpace(afterDeck)
	return &s, analysis, "", roles
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

	suggestion, analysis, errIA, roles := parseDeckBuilderOutput(raw)

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
	if roles != nil {
		resp["card_roles"] = roles
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

		group := "Outros"
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
		if len(cats[cat]) == 0 {
			continue
		}
		cardList.WriteString(fmt.Sprintf("\n### %s (%d únicos)\n", cat, len(cats[cat])))
		for _, entry := range cats[cat] {
			cardList.WriteString("- " + entry + "\n")
		}
	}

	var formatInstr string
	var outputExample string

	switch input.Format {
	case "commander":
		formatInstr = `Monte UM deck COMMANDER válido.

REGRAS DURAS DO COMMANDER:
- O deck deve ter EXATAMENTE 100 cartas no total, incluindo o comandante.
- Deve existir uma seção "Commander" com EXATAMENTE 1 carta.
- O comandante deve ser uma criatura lendária, ou uma carta que diga explicitamente que pode ser comandante.
- Todas as cartas devem respeitar a IDENTIDADE DE COR do comandante.
- A identidade de cor considera custo de mana, símbolos de mana no texto da carta e faces alternativas.
- Nenhuma carta fora da identidade de cor do comandante pode entrar.
- Cartas não-básicas são singleton: no máximo 1 cópia de cada nome.
- Terrenos básicos podem repetir livremente.
- Terrenos básicos devem respeitar as cores do comandante.
- Não use terrenos básicos de cores que o comandante não tenha.
- Use normalmente entre 35 e 40 terrenos.
- O campo "commander" no JSON deve ser true.

VALIDAÇÃO OBRIGATÓRIA:
1. Escolha primeiro o comandante.
2. Defina as cores exclusivamente pela identidade de cor dele.
3. Remova qualquer carta fora dessas cores.
4. Garanta singleton para todas as cartas não-básicas.
5. Conte todas as cartas.
6. O total deve ser exatamente 100.
7. Se não conseguir montar Commander válido, emita <<<ERRO>>>.`

		outputExample = `{
  "nome": "Legião de Tartarugas",
  "cores": "W,U,B,R,G",
  "commander": true,
  "descricao": "Deck Commander focado em sinergia tribal, valor incremental e finalizações em mesa cheia.",
  "lista": "Commander\n1 Leonardo, the Balance\n\nCriaturas\n1 Donatello, the Brains\n1 Raphael, the Muscle\n1 Michelangelo, the Heart\n\nArtefatos\n1 Sol Ring\n1 Arcane Signet\n\nTerrenos\n8 Plains\n8 Island\n8 Swamp\n8 Mountain\n8 Forest"
}`

	case "casual60":
		formatInstr = `Monte UM deck CASUAL DE 60 CARTAS válido.

REGRAS DURAS DO CASUAL 60:
- O deck deve ter EXATAMENTE 60 cartas.
- Não existe comandante.
- Não crie seção "Commander".
- O campo "commander" no JSON deve ser false.
- Cartas não-básicas podem ter até 4 cópias.
- Nunca use mais cópias de uma carta do que a quantidade disponível na coleção.
- Terrenos básicos podem ser adicionados livremente.
- Use normalmente entre 22 e 24 terrenos.
- O deck deve ter uma estratégia clara.
- A curva de mana deve ser jogável, com foco em custos 1, 2 e 3.
- Use poucas cartas de custo 5 ou maior.

VALIDAÇÃO OBRIGATÓRIA:
1. Defina uma estratégia central.
2. Escolha as cores de acordo com as melhores sinergias.
3. Respeite até 4 cópias por carta não-básica.
4. Respeite a quantidade disponível da coleção.
5. Conte todas as cartas.
6. O total deve ser exatamente 60.
7. Se não conseguir montar 60 cartas coerentes, emita <<<ERRO>>>.`

		outputExample = `{
  "nome": "Ataque Rápido",
  "cores": "R,G",
  "commander": false,
  "descricao": "Deck casual de 60 cartas focado em criaturas eficientes, pressão inicial e remoções simples.",
  "lista": "Criaturas\n4 Carta Um\n4 Carta Dois\n\nFeitiços\n4 Carta Três\n\nMágicas Imediatas\n4 Carta Quatro\n\nTerrenos\n10 Mountain\n10 Forest"
}`

	default:
		formatInstr = `Escolha automaticamente entre CASUAL 60 ou COMMANDER.

CRITÉRIO DE ESCOLHA:
- Escolha COMMANDER apenas se houver um comandante válido e cartas suficientes que respeitem sua identidade de cor.
- Escolha CASUAL 60 se as cartas disponíveis formarem melhor um deck comum de 60 cartas.
- Se escolher Commander, siga TODAS as regras de Commander.
- Se escolher Casual 60, siga TODAS as regras de Casual 60.

VALIDAÇÃO:
- Commander = exatamente 100 cartas, com comandante, singleton e identidade de cor.
- Casual 60 = exatamente 60 cartas, sem comandante, até 4 cópias por carta não-básica.
- Se nenhum formato ficar válido, emita <<<ERRO>>>.`

		outputExample = `{
  "nome": "Ataque Rápido",
  "cores": "R,G",
  "commander": false,
  "descricao": "Deck casual de 60 cartas focado em criaturas eficientes, pressão inicial e remoções simples.",
  "lista": "Criaturas\n4 Carta Um\n4 Carta Dois\n\nFeitiços\n4 Carta Três\n\nMágicas Imediatas\n4 Carta Quatro\n\nTerrenos\n10 Mountain\n10 Forest"
}`
	}

	var goalInstr string
	if input.Goal == "competitive" {
		goalInstr = `**Objetivo: COMPETITIVO** — maximize consistência, eficiência, curva de mana, remoções e vantagem de cartas.`
	} else {
		goalInstr = `**Objetivo: DIVERSÃO** — priorize sinergias temáticas, combos interessantes e plano de jogo divertido, mantendo o deck funcional.`
	}

	var colorInstr string
	if input.Colors != "" {
		colorInstr = fmt.Sprintf(`
**Cores preferidas:** %s.
- Em Casual 60, tente respeitar essas cores se houver cartas suficientes.
- Em Commander, essas cores só podem ser usadas se forem compatíveis com a identidade de cor do comandante.
- Nunca viole identidade de cor em Commander por causa da preferência de cores.
`, input.Colors)
	}

	var revaluateInstr string
	if input.Revaluate {
		revaluateInstr = `
**RE-AVALIAÇÃO:** sugira uma estratégia diferente da anterior. Explore outro arquétipo, outra combinação de cores ou outro comandante, mas sem violar as regras do formato.`
	}

	return fmt.Sprintf(`Você é um especialista em Magic: The Gathering e deck-building.

## CARTAS DISPONÍVEIS (%d cópias totais, %d únicas sem deck)
%s

## PARÂMETROS
%s

%s
%s
%s

## REGRAS GERAIS OBRIGATÓRIAS
1. Use somente cartas da lista acima, respeitando as quantidades disponíveis.
2. Exceção: terrenos básicos podem ser adicionados livremente.
3. Terrenos básicos válidos: Plains, Island, Swamp, Mountain, Forest.
4. Em Commander, cartas não-básicas são limitadas a 1 cópia, mesmo se houver mais cópias disponíveis.
5. Em Casual 60, cartas não-básicas podem usar até 4 cópias, mas nunca acima da quantidade disponível.
6. Não invente cartas não-básicas.
7. Não misture regras de Commander com Casual 60.
8. Monte apenas UM deck.
9. A lista final precisa bater exatamente com a contagem exigida pelo formato.
10. Se não conseguir montar um deck válido, use o bloco <<<ERRO>>>.

## FORMATO DE SAÍDA — EMITA APENAS UMA DAS OPÇÕES

OPÇÃO A — Deck válido:

<<<DECK>>>
%s
<<<FIM_DECK>>>

Regras do campo "lista":
- Use seções sem número.
- Seções permitidas: "Commander", "Criaturas", "Planeswalkers", "Feitiços", "Mágicas Imediatas", "Artefatos", "Encantamentos", "Terrenos".
- Use a seção "Commander" somente quando o formato for Commander.
- Cada carta deve seguir o formato: "N Nome Exato da Carta".
- Não coloque comentários, preços ou códigos de set dentro da lista.
- Terrenos básicos devem ficar no final.
- A soma das quantidades deve bater exatamente com o formato escolhido.

OPÇÃO B — Impossível montar deck válido:

<<<ERRO>>>
{"motivo": "Explique em detalhes por que não foi possível montar um deck válido com as cartas e parâmetros fornecidos."}
<<<FIM_ERRO>>>

Depois do bloco escolhido, escreva a análise em markdown:

## Análise
Explique a estratégia central do deck.

## Sinergias Principais
Liste 3 a 5 sinergias importantes.

## Como Jogar
Explique início, meio e fim de jogo.

## Pontos Fortes

## Limitações e o que comprar para completar

Se emitiu OPÇÃO A, emita também:

<<<CARTAS>>>
{
  "nao_terrenos": [
    {"nome": "Nome Exato da Carta", "papel": "O que a carta faz e por que está neste deck."}
  ],
  "terrenos": {
    "total": 0,
    "motivo": "Explique a quantidade de terrenos e a distribuição por cor."
  }
}
<<<FIM_CARTAS>>>

Regras do bloco CARTAS:
- Liste todas as cartas não-terreno do deck.
- Use os nomes exatamente como aparecem na lista.
- "terrenos.total" deve bater com a soma de terrenos da lista.`,
		totalQty,
		len(cards),
		cardList.String(),
		formatInstr,
		goalInstr,
		colorInstr,
		revaluateInstr,
		outputExample,
	)
}
