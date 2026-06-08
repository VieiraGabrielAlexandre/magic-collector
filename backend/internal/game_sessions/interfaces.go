package game_sessions

// gsRepository abstrai o acesso ao banco para o Service.
type gsRepository interface {
	List() ([]GameSession, error)
	GetByID(id int64) (*GameSession, error)
	Create(input CreateSessionInput) (*GameSession, error)
	Delete(id int64) error
	AddPlayer(sessionID int64, input PlayerInput, startingLife int) (*Player, error)
	UpdatePlayer(sessionID, playerID int64, input UpdatePlayerInput) (*Player, error)
	DeletePlayer(sessionID, playerID int64) error
	Reset(sessionID int64) (*GameSession, error)
	Finish(sessionID int64) (*GameSession, error)
	Restore(sessionID int64) (*GameSession, error)
}

// gameSessionService abstrai o Service para o Handler.
type gameSessionService interface {
	List() ([]GameSession, error)
	GetByID(id int64) (*GameSession, error)
	Create(input CreateSessionInput) (*GameSession, error)
	Delete(id int64) error
	AddPlayer(sessionID int64, input PlayerInput) (*Player, error)
	UpdatePlayer(sessionID, playerID int64, input UpdatePlayerInput) (*Player, error)
	DeletePlayer(sessionID, playerID int64) error
	Reset(sessionID int64) (*GameSession, error)
	Finish(sessionID int64) (*GameSession, error)
	Restore(sessionID int64) (*GameSession, error)
}
