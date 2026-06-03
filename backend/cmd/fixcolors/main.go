// fixcolors: recalcula a coluna `color` (PT legível) a partir de `colors` (JSON WUBRG).
// Uso: cd backend && go run ./cmd/fixcolors
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var colorCodeToPT = map[string]string{
	"W": "Branco", "U": "Azul", "B": "Preto",
	"R": "Vermelho", "G": "Verde", "C": "Incolor",
}

var wubrgIdx = map[string]int{"W": 0, "U": 1, "B": 2, "R": 3, "G": 4, "C": 5}

// colorsJSONToDisplay converte '["B","G"]' → "Preto/Verde" (ordem WUBRG).
func colorsJSONToDisplay(colorsJSON string) string {
	if colorsJSON == "" || colorsJSON == "null" || colorsJSON == "[]" {
		return ""
	}
	var codes []string
	if err := json.Unmarshal([]byte(colorsJSON), &codes); err != nil {
		return ""
	}
	sort.Slice(codes, func(i, j int) bool {
		oi, _ := wubrgIdx[codes[i]]
		oj, _ := wubrgIdx[codes[j]]
		return oi < oj
	})
	parts := make([]string, 0, len(codes))
	for _, c := range codes {
		if pt, ok := colorCodeToPT[c]; ok {
			parts = append(parts, pt)
		}
	}
	return strings.Join(parts, "/")
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		env("DB_USER", "root"),
		env("DB_PASSWORD", ""),
		env("DB_HOST", "localhost"),
		env("DB_PORT", "3306"),
		env("DB_NAME", "magic_collector"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Conexão: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("Ping: %v", err)
	}

	rows, err := db.Query("SELECT id, colors, color FROM cards ORDER BY id")
	if err != nil {
		log.Fatalf("Query: %v", err)
	}

	type row struct {
		id       int
		colors   string
		oldColor string
	}
	var cards []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.colors, &r.oldColor); err != nil {
			log.Printf("scan id desconhecido: %v", err)
			continue
		}
		cards = append(cards, r)
	}
	rows.Close()

	fmt.Printf("🃏  %d cartas carregadas\n\n", len(cards))

	updated, unchanged, errs := 0, 0, 0
	for _, c := range cards {
		newColor := colorsJSONToDisplay(c.colors)
		if newColor == c.oldColor {
			unchanged++
			continue
		}
		if _, err := db.Exec("UPDATE cards SET color = ? WHERE id = ?", newColor, c.id); err != nil {
			log.Printf("  ✗ id=%d: %v", c.id, err)
			errs++
			continue
		}
		fmt.Printf("  ✓ id=%-6d  %-25q → %q\n", c.id, c.oldColor, newColor)
		updated++
	}

	fmt.Printf("\n✅ %d atualizadas  |  %d sem mudança  |  %d erros\n", updated, unchanged, errs)
}
