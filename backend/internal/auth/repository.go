package auth

import (
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUserByUsername(username string) (*userWithHash, error) {
	var u userWithHash
	err := r.db.QueryRow(
		"SELECT id, username, display_name, password_hash, created_at FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

func (r *Repository) CreateSession(token string, userID int) error {
	_, err := r.db.Exec(
		"INSERT INTO sessions (token, user_id, created_at, expires_at) VALUES (?, ?, NOW(), DATE_ADD(NOW(), INTERVAL 30 DAY))",
		token, userID,
	)
	return err
}

func (r *Repository) GetSessionWithUser(token string) (*Session, *User, error) {
	var s Session
	var u User
	err := r.db.QueryRow(`
		SELECT s.token, s.user_id, s.created_at, s.expires_at,
		       u.id, u.username, u.display_name, u.created_at
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.token = ? AND s.expires_at > NOW()
	`, token).Scan(
		&s.Token, &s.UserID, &s.CreatedAt, &s.ExpiresAt,
		&u.ID, &u.Username, &u.DisplayName, &u.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	return &s, &u, nil
}

func (r *Repository) DeleteSession(token string) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}
