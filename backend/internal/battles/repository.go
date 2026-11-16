package battles

import (
	"database/sql"
	"encoding/json"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List() ([]Battle, error) {
	rows, err := r.db.Query(`
		SELECT id, result, COALESCE(opponents, '[]'), player_count, game_style,
		       deck_id, deck_name, deck_is_mine, notes,
		       DATE_FORMAT(played_at, '%Y-%m-%dT%H:%i:%s')
		FROM battles
		ORDER BY played_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Battle
	for rows.Next() {
		var b Battle
		var mineInt int
		var opponentsJSON string
		err := rows.Scan(
			&b.ID, &b.Result, &opponentsJSON, &b.PlayerCount, &b.GameStyle,
			&b.DeckID, &b.DeckName, &mineInt, &b.Notes, &b.PlayedAt,
		)
		if err != nil {
			return nil, err
		}
		b.DeckIsMine = mineInt == 1
		if err := json.Unmarshal([]byte(opponentsJSON), &b.Opponents); err != nil {
			b.Opponents = []string{}
		}
		result = append(result, b)
	}
	return result, nil
}

func (r *Repository) Create(b BattleInput) (int64, error) {
	mineInt := 0
	if b.DeckIsMine {
		mineInt = 1
	}
	if b.PlayerCount <= 0 {
		b.PlayerCount = 2
	}
	if b.Opponents == nil {
		b.Opponents = []string{}
	}
	opponentsJSON, _ := json.Marshal(b.Opponents)

	res, err := r.db.Exec(
		`INSERT INTO battles (result, opponents, player_count, game_style, deck_id, deck_name, deck_is_mine, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		b.Result, string(opponentsJSON), b.PlayerCount, b.GameStyle,
		b.DeckID, b.DeckName, mineInt, b.Notes,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM battles WHERE id = ?`, id)
	return err
}

func (r *Repository) GetStats() (BattleStats, error) {
	var s BattleStats

	battles, err := r.List()
	if err != nil {
		return s, err
	}

	deckMap := map[string]*DeckStat{}
	monthMap := map[string]*MonthStat{}
	opponentMap := map[string]int{}

	for _, b := range battles {
		s.Total++
		switch b.Result {
		case "win":
			s.Wins++
		case "loss":
			s.Losses++
		case "draw":
			s.Draws++
		}

		// Por deck
		if b.DeckName != "" {
			if _, ok := deckMap[b.DeckName]; !ok {
				deckMap[b.DeckName] = &DeckStat{DeckName: b.DeckName}
			}
			deckMap[b.DeckName].Total++
			if b.Result == "win" {
				deckMap[b.DeckName].Wins++
			}
		}

		// Por mês
		month := ""
		if len(b.PlayedAt) >= 7 {
			month = b.PlayedAt[:7]
		}
		if month != "" {
			if _, ok := monthMap[month]; !ok {
				monthMap[month] = &MonthStat{Month: month}
			}
			monthMap[month].Total++
			if b.Result == "win" {
				monthMap[month].Wins++
			}
		}

		// Adversários derrotados
		if b.Result == "win" {
			for _, op := range b.Opponents {
				if op != "" {
					opponentMap[op]++
				}
			}
		}
	}

	if s.Total > 0 {
		s.WinRate = float64(s.Wins) / float64(s.Total) * 100
	}

	for _, d := range deckMap {
		if d.Total > 0 {
			d.WinRate = float64(d.Wins) / float64(d.Total) * 100
		}
		s.ByDeck = append(s.ByDeck, *d)
	}
	// sort ByDeck descending by Total
	for i := 0; i < len(s.ByDeck)-1; i++ {
		for j := i + 1; j < len(s.ByDeck); j++ {
			if s.ByDeck[j].Total > s.ByDeck[i].Total {
				s.ByDeck[i], s.ByDeck[j] = s.ByDeck[j], s.ByDeck[i]
			}
		}
	}

	// Sort months ascending
	var months []string
	for m := range monthMap {
		months = append(months, m)
	}
	for i := 0; i < len(months)-1; i++ {
		for j := i + 1; j < len(months); j++ {
			if months[j] < months[i] {
				months[i], months[j] = months[j], months[i]
			}
		}
	}
	for _, m := range months {
		s.ByMonth = append(s.ByMonth, *monthMap[m])
	}

	// Top 5 adversários mais derrotados
	type kv struct {
		name   string
		losses int
	}
	var opList []kv
	for name, count := range opponentMap {
		opList = append(opList, kv{name, count})
	}
	for i := 0; i < len(opList)-1; i++ {
		for j := i + 1; j < len(opList); j++ {
			if opList[j].losses > opList[i].losses {
				opList[i], opList[j] = opList[j], opList[i]
			}
		}
	}
	limit := 5
	if len(opList) < limit {
		limit = len(opList)
	}
	for _, item := range opList[:limit] {
		s.TopOpponents = append(s.TopOpponents, OpponentStat{Name: item.name, Losses: item.losses})
	}

	return s, nil
}
