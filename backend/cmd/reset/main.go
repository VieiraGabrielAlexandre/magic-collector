package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"magic-collection-api/internal/database"
)

// loadDotEnv lê o arquivo .env e seta as variáveis no ambiente do processo,
// sem sobrescrever variáveis já definidas externamente.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // sem .env é ok
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func confirm(prompt string) bool {
	fmt.Printf("%s [sim/não]: ", prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	ans := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return ans == "sim" || ans == "s"
}

func main() {
	// Carrega .env da raiz do projeto (../../.env relativo a cmd/reset)
	loadDotEnv("../../.env")
	loadDotEnv("../.env") // fallback se rodado de dentro de backend/
	loadDotEnv(".env")    // fallback se rodado da raiz

	cards    := flag.Bool("cards", false, "apagar todas as cartas")
	decks    := flag.Bool("decks", false, "apagar todos os decks (e suas cartas)")
	battles  := flag.Bool("battles", false, "apagar todo o histórico de batalhas")
	wishlist := flag.Bool("wishlist", false, "apagar wishlist de cartas e decks desejados")
	all      := flag.Bool("all", false, "apagar tudo (cartas + decks + batalhas + wishlist)")
	force    := flag.Bool("force", false, "pular confirmação")
	flag.Parse()

	if !*cards && !*decks && !*battles && !*wishlist && !*all {
		fmt.Println("Uso:")
		fmt.Println("  go run ./cmd/reset [flags]")
		fmt.Println()
		fmt.Println("Flags:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Exemplo: go run ./cmd/reset --cards")
		fmt.Println("Exemplo: go run ./cmd/reset --all --force")
		os.Exit(0)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		getenv("DB_USER", "root"),
		getenv("DB_PASSWORD", ""),
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "3306"),
		getenv("DB_NAME", "magic_collector"),
	)
	db, err := database.Open(dsn)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}
	defer db.Close()

	type op struct {
		label string
		query string
	}

	var ops []op

	if *all {
		ops = []op{
			{"cartas", "DELETE FROM cards"},
			{"decks", "DELETE FROM decks"},
			{"batalhas", "DELETE FROM battles"},
			{"wishlist de cartas", "DELETE FROM wishlist_cards"},
			{"decks desejados e suas cartas", "DELETE FROM wishlist_deck_cards; DELETE FROM wishlist_decks"},
		}
	} else {
		if *decks {
			ops = append(ops, op{"cartas (FK de decks)", "DELETE FROM cards WHERE deck_id > 0"})
			ops = append(ops, op{"decks", "DELETE FROM decks"})
		} else if *cards {
			ops = append(ops, op{"cartas", "DELETE FROM cards"})
		}
		if *battles {
			ops = append(ops, op{"batalhas", "DELETE FROM battles"})
		}
		if *wishlist {
			ops = append(ops, op{"wishlist de cartas", "DELETE FROM wishlist_cards"})
			ops = append(ops, op{"cartas dos decks desejados", "DELETE FROM wishlist_deck_cards"})
			ops = append(ops, op{"decks desejados", "DELETE FROM wishlist_decks"})
		}
	}

	fmt.Println("⚠️  As seguintes tabelas serão apagadas:")
	for _, o := range ops {
		fmt.Printf("  → %s\n", o.label)
	}
	fmt.Println()

	if !*force && !confirm("Tem certeza? Esta ação não pode ser desfeita") {
		fmt.Println("Cancelado.")
		os.Exit(0)
	}

	for _, o := range ops {
		// Suporta múltiplos statements separados por ";"
		for _, q := range strings.Split(o.query, ";") {
			q = strings.TrimSpace(q)
			if q == "" {
				continue
			}
			res, err := db.Exec(q)
			if err != nil {
				log.Printf("Erro ao apagar %s: %v", o.label, err)
				continue
			}
			n, _ := res.RowsAffected()
			fmt.Printf("✓ %s — %d registro(s) removido(s)\n", o.label, n)
		}
	}

	fmt.Println("\nConcluído.")
}
