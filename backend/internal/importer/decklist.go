package importer

import (
	"strconv"
	"strings"
)

type DeckListEntry struct {
	Quantity int
	Name     string
}

// ParseDeckList parseia uma lista de deck no formato padrão MTG.
// Suporta: "1 Card Name", "1\tCard Name", "1\tCard Name\t\t$ 1.23"
// Linhas que não começam com dígito são tratadas como cabeçalhos de seção e ignoradas.
func ParseDeckList(text string) []DeckListEntry {
	var entries []DeckListEntry

	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if len(line) == 0 {
			continue
		}
		if line[0] < '0' || line[0] > '9' {
			continue
		}

		sep := strings.IndexAny(line, " \t")
		if sep < 0 {
			continue
		}

		qty, err := strconv.Atoi(line[:sep])
		if err != nil || qty <= 0 {
			continue
		}

		rest := strings.TrimLeft(line[sep:], " \t")

		// Remove preço e texto após o primeiro tab: "Card Name\t\t$ 1.23"
		if tabIdx := strings.Index(rest, "\t"); tabIdx >= 0 {
			rest = rest[:tabIdx]
		}
		name := strings.TrimSpace(rest)
		if name == "" {
			continue
		}

		entries = append(entries, DeckListEntry{Quantity: qty, Name: name})
	}

	return entries
}
