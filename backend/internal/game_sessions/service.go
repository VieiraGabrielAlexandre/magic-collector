package game_sessions

import (
	"errors"
	"fmt"
	"strings"

	"magic-collection-api/internal/battles"
)

type Service struct {
	repo        gsRepository
	battleRepo  *battles.Repository
}

func NewService(repo *Repository, battleRepo *battles.Repository) *Service {
	return &Service{repo: repo, battleRepo: battleRepo}
}

func (s *Service) List() ([]GameSession, error) {
	return s.repo.List()
}

func (s *Service) GetByID(id int64) (*GameSession, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(input CreateSessionInput) (*GameSession, error) {
	if len(input.Players) < 2 {
		return nil, errors.New("mínimo de 2 jogadores")
	}
	if len(input.Players) > 8 {
		return nil, errors.New("máximo de 8 jogadores")
	}
	for _, p := range input.Players {
		if strings.TrimSpace(p.Name) == "" {
			return nil, errors.New("nome do jogador é obrigatório")
		}
		if len([]rune(strings.TrimSpace(p.ShortCode))) > 3 {
			return nil, errors.New("sigla deve ter no máximo 3 caracteres")
		}
		if strings.TrimSpace(p.ShortCode) == "" {
			return nil, errors.New("sigla do jogador é obrigatória")
		}
	}

	if input.Format == "" {
		input.Format = "Commander"
	}
	if input.StartingLife <= 0 {
		if strings.EqualFold(input.Format, "casual") {
			input.StartingLife = 20
		} else {
			input.StartingLife = 40
		}
	}

	return s.repo.Create(input)
}

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}

func (s *Service) AddPlayer(sessionID int64, input PlayerInput) (*Player, error) {
	session, err := s.repo.GetByID(sessionID)
	if err != nil || session == nil {
		return nil, errors.New("sessão não encontrada")
	}
	if session.Status == "finished" {
		return nil, errors.New("sessão encerrada não pode ser alterada")
	}
	if len(session.Players) >= 8 {
		return nil, errors.New("máximo de 8 jogadores")
	}
	if len([]rune(strings.TrimSpace(input.ShortCode))) > 3 {
		return nil, errors.New("sigla deve ter no máximo 3 caracteres")
	}
	return s.repo.AddPlayer(sessionID, input, session.StartingLife)
}

func (s *Service) UpdatePlayer(sessionID, playerID int64, input UpdatePlayerInput) (*Player, error) {
	session, err := s.repo.GetByID(sessionID)
	if err != nil || session == nil {
		return nil, errors.New("sessão não encontrada")
	}
	if session.Status == "finished" {
		return nil, errors.New("sessão encerrada não pode ser alterada")
	}
	return s.repo.UpdatePlayer(sessionID, playerID, input)
}

func (s *Service) DeletePlayer(sessionID, playerID int64) error {
	session, err := s.repo.GetByID(sessionID)
	if err != nil || session == nil {
		return errors.New("sessão não encontrada")
	}
	if session.Status == "finished" {
		return errors.New("sessão encerrada não pode ser alterada")
	}
	remaining := 0
	for _, p := range session.Players {
		if p.ID != playerID {
			remaining++
		}
	}
	if remaining < 2 {
		return errors.New("mínimo de 2 jogadores por sessão")
	}
	return s.repo.DeletePlayer(sessionID, playerID)
}

func (s *Service) Reset(sessionID int64) (*GameSession, error) {
	session, err := s.repo.GetByID(sessionID)
	if err != nil || session == nil {
		return nil, errors.New("sessão não encontrada")
	}
	if session.Status == "finished" {
		return nil, errors.New("sessão encerrada não pode ser resetada")
	}
	return s.repo.Reset(sessionID)
}

func (s *Service) Finish(sessionID int64) (*GameSession, error) {
	session, err := s.repo.GetByID(sessionID)
	if err != nil || session == nil {
		return nil, errors.New("sessão não encontrada")
	}
	if session.Status == "finished" {
		return nil, errors.New("sessão já encerrada")
	}
	finished, err := s.repo.Finish(sessionID)
	if err != nil {
		return nil, err
	}
	s.autoCreateBattle(finished)
	return finished, nil
}

func (s *Service) autoCreateBattle(session *GameSession) {
	if s.battleRepo == nil || session == nil {
		return
	}

	var survivors []Player
	for _, p := range session.Players {
		if !p.IsEliminated {
			survivors = append(survivors, p)
		}
	}

	result := "draw"
	winnerName := ""
	var opponents []string

	if len(survivors) == 1 {
		result = "win"
		winnerName = survivors[0].Name
		for _, p := range session.Players {
			if p.ID != survivors[0].ID {
				opponents = append(opponents, p.Name)
			}
		}
	} else {
		for _, p := range session.Players {
			opponents = append(opponents, p.Name)
		}
	}

	notes := fmt.Sprintf("Sessão: %s", session.Name)
	if winnerName != "" {
		notes += fmt.Sprintf(" | Vencedor: %s", winnerName)
	}

	_, _ = s.battleRepo.Create(battles.BattleInput{
		Result:      result,
		Opponents:   opponents,
		PlayerCount: len(session.Players),
		GameStyle:   session.Format,
		DeckName:    winnerName,
		Notes:       notes,
	})
}

func (s *Service) Restore(sessionID int64) (*GameSession, error) {
	session, err := s.repo.GetByID(sessionID)
	if err != nil || session == nil {
		return nil, errors.New("sessão não encontrada")
	}
	return s.repo.Restore(sessionID)
}
