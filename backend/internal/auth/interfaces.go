package auth

// authRepository abstrai o acesso ao banco para o Service.
type authRepository interface {
	GetUserByUsername(username string) (*userWithHash, error)
	CreateSession(token string, userID int) error
	GetSessionWithUser(token string) (*Session, *User, error)
	DeleteSession(token string) error
}

// authService abstrai o Service para o Handler.
type authService interface {
	Login(username, password string) (*Session, *User, error)
	GetSession(token string) (*Session, *User, error)
	Logout(token string) error
}
