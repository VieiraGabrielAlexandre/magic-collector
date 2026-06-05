package game_sessions

type GameSession struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	Format       string   `json:"format"`
	Status       string   `json:"status"`
	StartingLife int      `json:"starting_life"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	EndedAt      *string  `json:"ended_at"`
	Players      []Player `json:"players"`
}

type Player struct {
	ID                      int64  `json:"id"`
	SessionID               int64  `json:"session_id"`
	Name                    string `json:"name"`
	ShortCode               string `json:"short_code"`
	Life                    int    `json:"life"`
	Poison                  int    `json:"poison"`
	CommanderDamageReceived int    `json:"commander_damage_received"`
	IsEliminated            bool   `json:"is_eliminated"`
	EliminatedReason        string `json:"eliminated_reason"`
	CreatedAt               string `json:"created_at"`
	UpdatedAt               string `json:"updated_at"`
}

type PlayerInput struct {
	Name      string `json:"name" binding:"required"`
	ShortCode string `json:"short_code" binding:"required"`
}

type CreateSessionInput struct {
	Name         string        `json:"name" binding:"required"`
	Format       string        `json:"format"`
	StartingLife int           `json:"starting_life"`
	Players      []PlayerInput `json:"players"`
}

type UpdatePlayerInput struct {
	Life                    int `json:"life"`
	Poison                  int `json:"poison"`
	CommanderDamageReceived int `json:"commander_damage_received"`
}
