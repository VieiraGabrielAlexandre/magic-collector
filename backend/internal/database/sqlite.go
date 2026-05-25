package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		mtg_id TEXT,
		name TEXT NOT NULL,
		set_code TEXT,
		rarity TEXT,
		type TEXT,
		mana_cost TEXT,
		colors TEXT,
		quantity INTEGER NOT NULL DEFAULT 1,
		condition TEXT,
		language TEXT,
		notes TEXT,
		color TEXT,
		subtitle TEXT,
		collection_number TEXT,
		year INTEGER,
		artist TEXT,
		company TEXT,
		foil INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	migrations := []string{
		`ALTER TABLE cards ADD COLUMN color TEXT`,
		`ALTER TABLE cards ADD COLUMN subtitle TEXT`,
		`ALTER TABLE cards ADD COLUMN collection_number TEXT`,
		`ALTER TABLE cards ADD COLUMN year INTEGER`,
		`ALTER TABLE cards ADD COLUMN artist TEXT`,
		`ALTER TABLE cards ADD COLUMN company TEXT`,
		`ALTER TABLE cards ADD COLUMN foil INTEGER NOT NULL DEFAULT 0`,
	}
	for _, m := range migrations {
		db.Exec(m) // ignora erro se coluna já existe
	}

	return db, nil
}
