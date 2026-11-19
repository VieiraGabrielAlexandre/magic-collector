package gamechangers

import (
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// List retorna todos os game changers com contagem de cópias na coleção do usuário.
func (r *Repository) List() ([]GameChanger, error) {
	rows, err := r.db.Query(`
		SELECT g.id, g.scryfall_id, g.name, g.set_code, g.rarity,
		       g.colors, COALESCE(g.image_url,''), g.mana_cost, g.type_line,
		       g.synced_at,
		       COALESCE(SUM(c.quantity), 0) AS owned_count
		FROM gamechanger_cards g
		LEFT JOIN cards c ON LOWER(c.name) = LOWER(g.name)
		GROUP BY g.id
		ORDER BY owned_count DESC, g.name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []GameChanger
	for rows.Next() {
		var gc GameChanger
		if err := rows.Scan(&gc.ID, &gc.ScryfallID, &gc.Name, &gc.SetCode, &gc.Rarity,
			&gc.Colors, &gc.ImageURL, &gc.ManaCost, &gc.TypeLine,
			&gc.SyncedAt, &gc.OwnedCount); err != nil {
			return nil, err
		}
		result = append(result, gc)
	}
	return result, nil
}

// CountSyncedAt retorna quando foi o último sync e quantas cartas há.
func (r *Repository) SyncMeta() (count int, syncedAt string) {
	r.db.QueryRow("SELECT COUNT(*), COALESCE(MAX(synced_at),'') FROM gamechanger_cards").
		Scan(&count, &syncedAt)
	return
}

// Upsert insere ou atualiza um game changer pelo scryfall_id.
func (r *Repository) Upsert(gc GameChanger) error {
	_, err := r.db.Exec(`
		INSERT INTO gamechanger_cards (scryfall_id, name, set_code, rarity, colors, image_url, mana_cost, type_line, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name=VALUES(name), set_code=VALUES(set_code), rarity=VALUES(rarity),
			colors=VALUES(colors), image_url=VALUES(image_url), mana_cost=VALUES(mana_cost),
			type_line=VALUES(type_line), synced_at=VALUES(synced_at)`,
		gc.ScryfallID, gc.Name, gc.SetCode, gc.Rarity, gc.Colors,
		gc.ImageURL, gc.ManaCost, gc.TypeLine, time.Now().Format("2006-01-02 15:04:05"))
	return err
}

// DeleteNotIn remove registros cujo scryfall_id não está na lista fornecida.
func (r *Repository) DeleteNotIn(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	// build placeholder list
	query := "DELETE FROM gamechanger_cards WHERE scryfall_id NOT IN (?"
	for i := 1; i < len(ids); i++ {
		query += ",?"
	}
	query += ")"
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	_, err := r.db.Exec(query, args...)
	return err
}
