package game_sessions

import "encoding/json"

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
	ID                          int64  `json:"id"`
	SessionID                   int64  `json:"session_id"`
	Name                        string `json:"name"`
	ShortCode                   string `json:"short_code"`
	CommanderName               string `json:"commander_name"`
	CommanderSetCode            string `json:"commander_set_code"`
	CommanderCollectionNumber   string `json:"commander_collection_number"`
	CommanderImageURL           string `json:"commander_image_url"`
	Life                        int    `json:"life"`
	Poison                      int    `json:"poison"`
	CommanderDamageReceived     int    `json:"commander_damage_received"`
	Experience                  int    `json:"experience"`
	Energy                      int    `json:"energy"`
	IsEliminated                bool           `json:"is_eliminated"`
	EliminatedReason            string         `json:"eliminated_reason"`
	CmdDamageBySource           map[string]int `json:"cmd_damage_by_source"`
	CreatedAt                   string         `json:"created_at"`
	UpdatedAt                   string         `json:"updated_at"`
}

// parseCmdDamageBySource safely decodes a JSON string into a map.
func parseCmdDamageBySource(raw string) map[string]int {
	m := map[string]int{}
	if raw != "" && raw != "{}" {
		_ = json.Unmarshal([]byte(raw), &m)
	}
	return m
}

type PlayerInput struct {
	Name                      string `json:"name" binding:"required"`
	ShortCode                 string `json:"short_code" binding:"required"`
	CommanderSetCode          string `json:"commander_set_code"`
	CommanderCollectionNumber string `json:"commander_collection_number"`
	// Filled by service after Scryfall lookup — not sent by client
	CommanderName     string `json:"-"`
	CommanderImageURL string `json:"-"`
}

type CreateSessionInput struct {
	Name         string        `json:"name" binding:"required"`
	Format       string        `json:"format"`
	StartingLife int           `json:"starting_life"`
	Players      []PlayerInput `json:"players"`
}

type UpdatePlayerInput struct {
	Life                    int            `json:"life"`
	Poison                  int            `json:"poison"`
	CommanderDamageReceived int            `json:"commander_damage_received"`
	Experience              int            `json:"experience"`
	Energy                  int            `json:"energy"`
	CmdDamageBySource       map[string]int `json:"cmd_damage_by_source"`
}
