package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("não foi possível conectar ao banco de dados: %w", err)
	}

	// Evita broken pipe: fecha conexões ociosas antes que o servidor remoto as derrube.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cards (
			id                INT          NOT NULL AUTO_INCREMENT,
			mtg_id            VARCHAR(100) NOT NULL DEFAULT '',
			name              VARCHAR(500) NOT NULL,
			set_code          VARCHAR(20)  NOT NULL DEFAULT '',
			rarity            VARCHAR(10)  NOT NULL DEFAULT '',
			` + "`type`" + `  VARCHAR(200) NOT NULL DEFAULT '',
			mana_cost         VARCHAR(100) NOT NULL DEFAULT '',
			colors            TEXT         NOT NULL,
			quantity          INT          NOT NULL DEFAULT 1,
			` + "`condition`" + ` VARCHAR(50) NOT NULL DEFAULT '',
			language          VARCHAR(10)  NOT NULL DEFAULT '',
			notes             TEXT         NOT NULL,
			color             VARCHAR(100) NOT NULL DEFAULT '',
			subtitle          VARCHAR(200) NOT NULL DEFAULT '',
			collection_number VARCHAR(20)  NOT NULL DEFAULT '',
			year              INT          NOT NULL DEFAULT 0,
			artist            VARCHAR(200) NOT NULL DEFAULT '',
			company           VARCHAR(200) NOT NULL DEFAULT '',
			foil              TINYINT      NOT NULL DEFAULT 0,
			prerelease        TINYINT      NOT NULL DEFAULT 0,
			commander         TINYINT      NOT NULL DEFAULT 0,
			precon_deck       VARCHAR(200) NOT NULL DEFAULT '',
			created_at        DATETIME     DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS decks (
			id          INT          NOT NULL AUTO_INCREMENT,
			name        VARCHAR(200) NOT NULL,
			description TEXT,
			created_at  DATETIME     DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		return nil, err
	}

	// Migrações seguras: ignora erro se coluna já existe.
	db.Exec(`ALTER TABLE cards ADD COLUMN prerelease  TINYINT      NOT NULL DEFAULT 0`)
	db.Exec(`ALTER TABLE cards ADD COLUMN commander   TINYINT      NOT NULL DEFAULT 0`)
	db.Exec(`ALTER TABLE cards ADD COLUMN precon_deck VARCHAR(200) NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE cards ADD COLUMN deck_id     INT          NOT NULL DEFAULT 0`)
	db.Exec(`ALTER TABLE decks ADD COLUMN commander   TINYINT      NOT NULL DEFAULT 0`)
	db.Exec(`ALTER TABLE decks ADD COLUMN colors      VARCHAR(50)  NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE decks ADD COLUMN set_code    VARCHAR(20)  NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE decks ADD COLUMN icon_uri     VARCHAR(500) NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE decks ADD COLUMN theme_color  VARCHAR(30)  NOT NULL DEFAULT ''`)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS battles (
			id           INT         NOT NULL AUTO_INCREMENT,
			result       VARCHAR(10) NOT NULL,
			opponents    TEXT        NOT NULL,
			player_count INT         NOT NULL DEFAULT 2,
			game_style   VARCHAR(50) NOT NULL DEFAULT '',
			deck_id      INT         NOT NULL DEFAULT 0,
			deck_name    VARCHAR(255) NOT NULL DEFAULT '',
			deck_is_mine TINYINT     NOT NULL DEFAULT 1,
			notes        TEXT        NOT NULL,
			played_at    DATETIME    DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	db.Exec(`ALTER TABLE battles ADD COLUMN opponents TEXT`)
	db.Exec(`ALTER TABLE battles DROP COLUMN opponent`)
	if err != nil {
		return nil, err
	}

	return db, nil
}
