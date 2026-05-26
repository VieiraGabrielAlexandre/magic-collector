package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"magic-collection-api/internal/cards"
	"magic-collection-api/internal/database"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	csvPath := "./data/cards.csv"
	if len(os.Args) > 1 {
		csvPath = os.Args[1]
	}

	f, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("erro ao abrir CSV: %v", err)
	}
	defer f.Close()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		getenv("DB_USER", "root"),
		getenv("DB_PASSWORD", ""),
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "3306"),
		getenv("DB_NAME", "magic_collector"),
	)

	db, err := database.Open(dsn)
	if err != nil {
		log.Fatalf("erro ao abrir banco: %v", err)
	}
	defer db.Close()

	repo := cards.NewRepository(db)

	reader := csv.NewReader(f)
	reader.LazyQuotes = true

	// lê cabeçalho
	if _, err := reader.Read(); err != nil {
		log.Fatalf("erro ao ler cabeçalho: %v", err)
	}

	// agrupa cartas idênticas (mesmo nome + coleção + número + idioma + foil)
	type key struct {
		name, setCode, collectionNumber, language string
		foil                                      bool
	}
	grouped := make(map[key]*cards.Card)
	order := []key{}

	lineNum := 1
	for {
		lineNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("linha %d ignorada: %v", lineNum, err)
			continue
		}
		if len(record) < 12 {
			log.Printf("linha %d ignorada: colunas insuficientes (%d)", lineNum, len(record))
			continue
		}

		name := strings.TrimSpace(record[0])
		color := strings.TrimSpace(record[1])
		cardType := strings.TrimSpace(record[2])
		subtitle := strings.TrimSpace(record[3])
		collectionNumber := strings.TrimSpace(record[4])
		rarity := strings.TrimSpace(record[5])
		setCode := strings.TrimSpace(record[6])
		language := strings.TrimSpace(record[7])
		yearStr := strings.TrimSpace(record[8])
		artist := strings.TrimSpace(record[9])
		company := strings.TrimSpace(record[10])
		foilStr := strings.ToLower(strings.TrimSpace(record[11]))

		if name == "" {
			continue
		}

		year, _ := strconv.Atoi(yearStr)
		foil := foilStr == "sim" || foilStr == "yes" || foilStr == "true"

		k := key{
			name:             name,
			setCode:          setCode,
			collectionNumber: collectionNumber,
			language:         language,
			foil:             foil,
		}

		if existing, ok := grouped[k]; ok {
			existing.Quantity++
		} else {
			grouped[k] = &cards.Card{
				Name:             name,
				Color:            color,
				Type:             cardType,
				Subtitle:         subtitle,
				CollectionNumber: collectionNumber,
				Rarity:           rarity,
				SetCode:          setCode,
				Language:         language,
				Year:             year,
				Artist:           artist,
				Company:          company,
				Foil:             foil,
				Quantity:         1,
				Condition:        "played",
			}
			order = append(order, k)
		}
	}

	inserted := 0
	for _, k := range order {
		card := grouped[k]
		_, err := repo.Create(*card)
		if err != nil {
			log.Printf("erro ao inserir '%s': %v", card.Name, err)
			continue
		}
		foilTag := ""
		if card.Foil {
			foilTag = " [FOIL]"
		}
		fmt.Printf("+ %s (%s #%s) x%d%s\n", card.Name, card.SetCode, card.CollectionNumber, card.Quantity, foilTag)
		inserted++
	}

	fmt.Printf("\n%d cartas inseridas (%d entradas únicas de %d linhas)\n", inserted, len(order), lineNum-1)
}
