package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
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
	db.Exec(`ALTER TABLE decks ADD COLUMN evaluation   LONGTEXT`)
	db.Exec(`ALTER TABLE decks ADD COLUMN evaluated_at DATETIME NULL`)
	db.Exec(`ALTER TABLE cards ADD COLUMN price_usd  DECIMAL(10,2)  NOT NULL DEFAULT 0`)
	db.Exec(`ALTER TABLE cards ADD COLUMN image_url  VARCHAR(500)   NOT NULL DEFAULT ''`)

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

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id            INT          NOT NULL AUTO_INCREMENT,
			username      VARCHAR(50)  NOT NULL UNIQUE,
			display_name  VARCHAR(100) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at    DATETIME     DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			token      VARCHAR(64)  NOT NULL,
			user_id    INT          NOT NULL,
			created_at DATETIME     DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME     NOT NULL,
			PRIMARY KEY (token),
			INDEX idx_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS wishlist_cards (
			id                BIGINT        NOT NULL AUTO_INCREMENT,
			mtg_id            VARCHAR(64)   NOT NULL DEFAULT '',
			set_code          VARCHAR(10)   NOT NULL,
			collection_number VARCHAR(20)   NOT NULL,
			name              VARCHAR(255)  NOT NULL DEFAULT '',
			printed_name      VARCHAR(255)  NOT NULL DEFAULT '',
			image_uri         TEXT,
			artist            VARCHAR(255)  NOT NULL DEFAULT '',
			rarity            VARCHAR(10)   NOT NULL DEFAULT '',
			colors            TEXT,
			color             VARCHAR(100)  NOT NULL DEFAULT '',
			price_usd         DECIMAL(10,2) NULL,
			price_usd_foil    DECIMAL(10,2) NULL,
			foil              TINYINT(1)    NOT NULL DEFAULT 0,
			reason            TEXT,
			acquired          TINYINT(1)    NOT NULL DEFAULT 0,
			created_at        DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at        DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS game_sessions (
			id            BIGINT       NOT NULL AUTO_INCREMENT,
			name          VARCHAR(255) NOT NULL,
			format        VARCHAR(50)  NOT NULL DEFAULT 'Commander',
			status        VARCHAR(20)  NOT NULL DEFAULT 'active',
			starting_life INT          NOT NULL DEFAULT 40,
			created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			ended_at      DATETIME     NULL,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS game_session_players (
			id                        BIGINT       NOT NULL AUTO_INCREMENT,
			session_id                BIGINT       NOT NULL,
			name                      VARCHAR(255) NOT NULL,
			short_code                VARCHAR(3)   NOT NULL,
			life                      INT          NOT NULL DEFAULT 40,
			poison                    INT          NOT NULL DEFAULT 0,
			commander_damage_received INT          NOT NULL DEFAULT 0,
			is_eliminated             TINYINT(1)   NOT NULL DEFAULT 0,
			eliminated_reason         VARCHAR(50)  NULL,
			created_at                DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at                DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			INDEX idx_gsp_session_id (session_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		return nil, err
	}

	seedUsers(db)

	return db, nil
}

// seedUsers insere os usuários iniciais se ainda não existirem.
func seedUsers(db *sql.DB) {
	initial := []struct{ username, displayName, password string }{
		{"gabriel", "Gabriel", "gabriel123"},
		{"juliana", "Juliana", "juliana123"},
	}
	for _, u := range initial {
		var count int
		db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", u.username).Scan(&count)
		if count == 0 {
			hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("seed: erro ao gerar hash para %s: %v", u.username, err)
				continue
			}
			db.Exec(
				"INSERT INTO users (username, display_name, password_hash) VALUES (?, ?, ?)",
				u.username, u.displayName, string(hash),
			)
			log.Printf("seed: usuário '%s' criado (senha: %s)", u.username, u.password)
		}
	}
}
