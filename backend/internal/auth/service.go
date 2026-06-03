package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repository *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repository: repo}
}

func (s *Service) Login(username, password string) (*Session, *User, error) {
	u, err := s.repository.GetUserByUsername(username)
	if err != nil {
		return nil, nil, err
	}
	if u == nil {
		return nil, nil, fmt.Errorf("credenciais inválidas")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, nil, fmt.Errorf("credenciais inválidas")
	}

	token, err := generateToken()
	if err != nil {
		return nil, nil, err
	}
	if err := s.repository.CreateSession(token, u.ID); err != nil {
		return nil, nil, err
	}

	sess, usr, err := s.repository.GetSessionWithUser(token)
	return sess, usr, err
}

func (s *Service) GetSession(token string) (*Session, *User, error) {
	return s.repository.GetSessionWithUser(token)
}

func (s *Service) Logout(token string) error {
	return s.repository.DeleteSession(token)
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashPassword é exportado para uso no seed de usuários iniciais.
func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h), err
}
