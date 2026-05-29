package decks

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"magic-collection-api/internal/ai"
	"magic-collection-api/internal/cards"
	"magic-collection-api/internal/mtgapi"
)

type Service struct {
	repo      *Repository
	mtgClient *mtgapi.Client
	cardRepo  *cards.Repository
	aiClient  *ai.Client
}

func NewService(repo *Repository, mtgClient *mtgapi.Client, cardRepo *cards.Repository, aiClient *ai.Client) *Service {
	return &Service{repo: repo, mtgClient: mtgClient, cardRepo: cardRepo, aiClient: aiClient}
}

func (s *Service) List() ([]Deck, error) {
	return s.repo.List()
}

func (s *Service) Create(input DeckInput) (int64, error) {
	id, err := s.repo.Create(input)
	if err != nil {
		return 0, err
	}
	if input.SetCode != "" {
		if set, _ := s.mtgClient.GetSetByCode(input.SetCode); set != nil && set.IconSVGURI != "" {
			_ = s.repo.UpdateIcon(strconv.FormatInt(id, 10), set.IconSVGURI)
		}
	}
	return id, nil
}

func (s *Service) Update(id string, input DeckInput) error {
	if err := s.repo.Update(id, input); err != nil {
		return err
	}
	if input.SetCode != "" {
		if set, _ := s.mtgClient.GetSetByCode(input.SetCode); set != nil && set.IconSVGURI != "" {
			_ = s.repo.UpdateIcon(id, set.IconSVGURI)
		}
	}
	return nil
}

func (s *Service) Delete(id string) error {
	return s.repo.Delete(id)
}

// FetchIcon busca e persiste o ícone do set para um deck que ainda não tem ícone.
func (s *Service) FetchIcon(id string) (string, error) {
	deck, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	if deck.SetCode == "" {
		return "", nil
	}
	set, err := s.mtgClient.GetSetByCode(deck.SetCode)
	if err != nil || set == nil || set.IconSVGURI == "" {
		return "", err
	}
	if err := s.repo.UpdateIcon(id, set.IconSVGURI); err != nil {
		return "", err
	}
	return set.IconSVGURI, nil
}

// EvaluateDeck gera uma avaliação estratégica do deck usando a Claude API e persiste o resultado.
func (s *Service) EvaluateDeck(id string) (*Deck, error) {
	deck, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("deck não encontrado: %w", err)
	}

	evalCards, err := s.cardRepo.ListForEval(deck.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar cartas: %w", err)
	}
	if len(evalCards) == 0 {
		return nil, fmt.Errorf("deck sem cartas para avaliar")
	}

	prompt := buildEvalPrompt(deck, evalCards)
	evaluation, err := s.aiClient.Complete(prompt)
	if err != nil {
		return nil, fmt.Errorf("erro na API de IA: %w", err)
	}

	if err := s.repo.UpdateEvaluation(id, evaluation); err != nil {
		return nil, fmt.Errorf("erro ao salvar avaliação: %w", err)
	}

	deck.Evaluation = evaluation
	deck.EvaluatedAt = time.Now().Format("2006-01-02T15:04:05")
	return deck, nil
}

func buildEvalPrompt(deck *Deck, evalCards []cards.EvalCardInfo) string {
	cats := map[string][]string{
		"Criatura": {}, "Planeswalker": {}, "Feitiço": {},
		"Mágica Imediata": {}, "Artefato": {}, "Encantamento": {},
		"Terreno": {}, "Outros": {},
	}
	order := []string{"Criatura", "Planeswalker", "Feitiço", "Mágica Imediata", "Artefato", "Encantamento", "Terreno", "Outros"}

	for _, c := range evalCards {
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
		entry := c.Name
		if c.ManaCost != "" {
			entry += " " + c.ManaCost
		}
		cats[group] = append(cats[group], entry)
	}

	var cardList strings.Builder
	for _, cat := range order {
		if len(cats[cat]) > 0 {
			cardList.WriteString(fmt.Sprintf("\n**%s (%d):**\n", cat, len(cats[cat])))
			for _, name := range cats[cat] {
				cardList.WriteString("- " + name + "\n")
			}
		}
	}

	colors := deck.Colors
	if colors == "" {
		colors = "Incolor"
	}
	format := "Deck personalizado"
	if deck.Commander {
		format = "Commander (EDH) — 100 cartas singleton, multiplayer 4 jogadores"
	}

	return fmt.Sprintf(`Você é um especialista em Magic: The Gathering com amplo conhecimento de EDH/Commander e formatos competitivos.

Analise o deck abaixo e forneça uma avaliação estratégica completa em português brasileiro. Seja detalhado, útil e direto.

**Nome do Deck:** %s
**Formato:** %s
**Identidade de Cores:** %s
**Total de Cartas:** %d
%s
Forneça a avaliação em markdown com exatamente estas seções:

## 🎯 Estratégia Principal
Qual o objetivo central do deck? Como ele pretende vencer? Descreva o plano de jogo em alto nível.

## ✨ Destaques e Cartas-Chave
Liste as cartas mais importantes do deck, explique o papel de cada uma e por que são fundamentais.

## 🔗 Sinergias e Combos
Identifique combinações poderosas entre cartas. Seja específico: "Carta A + Carta B = efeito X". Liste os principais combos se houver.

## 🎮 Como Jogar
Guia prático: quais cartas manter no mulligan, como montar a mesa nos primeiros turnos, mid-game e como fechar o jogo. Prioridades de cada fase.

## ⚡ Efeitos Principais
Mecânicas centrais que o deck utiliza (ex: ramp, draw, tokens, counters, reanimação, etc). Explique como cada mecânica serve a estratégia.

## 🔄 Efeitos Secundários e Plano B
Mecânicas de apoio, respostas, remoção, e o que fazer quando o plano principal falha.

## ⚠️ Ameaças e Pontos Fracos
O que contra-ataca esse deck. Tipos de estratégias adversárias perigosas. Cartas específicas que o ameaçam. Onde o deck é vulnerável.

## 💡 Dicas Avançadas
Timing de jogo, como lidar com situações difíceis, erros comuns a evitar, como reagir a diferentes composições de mesa.

## 📊 Avaliação Final
Dê uma nota de 1-10 para: Sinergia, Consistência, Potencial Competitivo e Curva de Mana. Finalize com um parágrafo de resumo geral do deck.`,
		deck.Name, format, colors, len(evalCards), cardList.String())
}
