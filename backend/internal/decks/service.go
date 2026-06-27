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
	repo      deckRepository
	mtgClient deckMtgClient
	cardRepo  deckCardRepo
	aiClient  deckAiClient
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
	raw, err := s.aiClient.Complete(prompt)
	if err != nil {
		return nil, fmt.Errorf("erro na API de IA: %w", err)
	}

	evaluation := extractJSON(raw)

	if err := s.repo.UpdateEvaluation(id, evaluation); err != nil {
		return nil, fmt.Errorf("erro ao salvar avaliação: %w", err)
	}

	deck.Evaluation = evaluation
	deck.EvaluatedAt = time.Now().Format("2006-01-02T15:04:05")
	return deck, nil
}

// extractJSON remove fences markdown (```json ... ```) que modelos às vezes adicionam.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.SplitN(s, "\n", 2)
		if len(lines) == 2 {
			s = lines[1]
		}
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
	}
	return strings.TrimSpace(s)
}

func buildEvalPrompt(deck *Deck, evalCards []cards.EvalCardInfo) string {
	colors := deck.Colors
	if colors == "" {
		colors = "Incolor"
	}
	format := "Deck personalizado"
	if deck.Commander {
		format = "Commander (EDH)"
	}

	// Separa comandantes e computa resumo de tipos
	var commanders []string
	var cardLines []string

	type typeSummary struct {
		lands, legendaryCreatures, creatures, instants, sorceries, artifacts, enchantments, planeswalkers, other int
	}
	var ts typeSummary

	for _, c := range evalCards {
		qty := c.Quantity
		if qty < 1 {
			qty = 1
		}
		set := strings.ToUpper(c.SetCode)
		if set == "" {
			set = "???"
		}
		num := c.CollectionNumber
		if num == "" {
			num = "?"
		}

		if c.IsCommander {
			commanders = append(commanders, fmt.Sprintf("%s (%s #%s)", c.Name, set, num))
			continue
		}

		line := fmt.Sprintf("%s (%s #%s)", c.Name, set, num)
		if qty > 1 {
			line += fmt.Sprintf(" ×%d", qty)
		}
		cardLines = append(cardLines, line)

		t := strings.ToLower(c.Type)
		switch {
		case strings.Contains(t, "land") || strings.Contains(t, "terreno"):
			ts.lands += qty
		case (strings.Contains(t, "legendary") || strings.Contains(t, "lendária")) &&
			(strings.Contains(t, "creature") || strings.Contains(t, "criatura")):
			ts.legendaryCreatures += qty
			ts.creatures += qty
		case strings.Contains(t, "creature") || strings.Contains(t, "criatura"):
			ts.creatures += qty
		case strings.Contains(t, "instant") || strings.Contains(t, "imediata"):
			ts.instants += qty
		case strings.Contains(t, "sorcery") || strings.Contains(t, "feitiço"):
			ts.sorceries += qty
		case strings.Contains(t, "artifact") || strings.Contains(t, "artefato"):
			ts.artifacts += qty
		case strings.Contains(t, "enchantment") || strings.Contains(t, "encantamento"):
			ts.enchantments += qty
		case strings.Contains(t, "planeswalker"):
			ts.planeswalkers += qty
		default:
			ts.other += qty
		}
	}
	total := ts.lands + ts.creatures + ts.instants + ts.sorceries + ts.artifacts + ts.enchantments + ts.planeswalkers + ts.other + len(commanders)

	cmdSection := "Nenhum"
	if len(commanders) > 0 {
		cmdSection = strings.Join(commanders, "\n")
	}

	cardSection := strings.Join(cardLines, "\n")

	return fmt.Sprintf(`Você é um deck builder profissional de Magic.

Seu objetivo é identificar as características estratégicas do deck.

Não conte cartas.
Não calcule estatísticas.
Utilize as estatísticas fornecidas.
Responda apenas com base no deck recebido.

══════════════════════════════════════════════

DECK

Nome: %s
Formato: %s
Identidade de cores: %s

COMANDANTES

%s

RESUMO DE TIPOS
Terrenos: %d
Criaturas: %d (sendo %d lendárias)
Instantâneas: %d
Feitiços: %d
Artefatos: %d
Encantamentos: %d
Planeswalkers: %d
Outros: %d
Total: %d

LISTA DE CARTAS

%s

══════════════════════════════════════════════

Responda EXCLUSIVAMENTE com um objeto JSON válido — sem explicações, sem markdown, sem texto fora do JSON.

Use exatamente esta estrutura:

{
  "arch_principal": "string — arquétipo principal",
  "arch_secundary": "string — arquétipo secundário ou null",
  "game_plan": "string — descrição do plano de jogo",
  "win_condition": "string — como o deck vence",
  "resource_recovery": "string — como recupera recursos",
  "has_explosion": true,
  "has_engine": true,
  "commander_dependency": "string — Não / Pouco / Médio / Alto / Totalmente",
  "has_plan_b": true,
  "plan_b": "string — descrição do plano B ou null se não houver",
  "speed": {
    "early": "string — o que faz nos turnos 1-3",
    "mid": "string — o que faz nos turnos 4-7",
    "late": "string — o que faz no turno 8+"
  },
  "mechanics": ["string"],
  "keywords": ["string"],
  "tribes": ["string"],
  "core_cards": ["string"],
  "displaced_cards": ["string"],
  "win_condition_cards": ["string"],
  "advantage_cards": ["string"],
  "acceleration_cards": ["string"],
  "protection_cards": ["string"],
  "removal_cards": ["string"],
  "engine_cards": ["string"],
  "never_remove_cards": ["string"],
  "optional_cards": ["string"],
  "weak_cards": ["string"],
  "bracket": 1,
  "bracket_explanation": "string"
}

Use apenas cartas da lista recebida. Nunca invente cartas. Responda em português brasileiro.`,
		deck.Name, format, colors,
		cmdSection,
		ts.lands, ts.creatures, ts.legendaryCreatures,
		ts.instants, ts.sorceries, ts.artifacts, ts.enchantments, ts.planeswalkers, ts.other,
		total,
		cardSection,
	)
}
