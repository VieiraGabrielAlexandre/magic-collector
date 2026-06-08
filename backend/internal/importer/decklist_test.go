package importer

import (
	"testing"
)

func TestParseDeckList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []DeckListEntry
	}{
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "single card space separator",
			input: "4 Lightning Bolt",
			want:  []DeckListEntry{{Quantity: 4, Name: "Lightning Bolt"}},
		},
		{
			name:  "single card tab separator",
			input: "1\tSol Ring",
			want:  []DeckListEntry{{Quantity: 1, Name: "Sol Ring"}},
		},
		{
			name:  "multiple cards",
			input: "4 Counterspell\n2 Cancel\n1 Force of Will",
			want: []DeckListEntry{
				{Quantity: 4, Name: "Counterspell"},
				{Quantity: 2, Name: "Cancel"},
				{Quantity: 1, Name: "Force of Will"},
			},
		},
		{
			name:  "ignores section headers (no leading digit)",
			input: "Creatures\n4 Birds of Paradise\nLands\n10 Forest",
			want: []DeckListEntry{
				{Quantity: 4, Name: "Birds of Paradise"},
				{Quantity: 10, Name: "Forest"},
			},
		},
		{
			name:  "ignores empty lines",
			input: "2 Brainstorm\n\n3 Ponder\n",
			want: []DeckListEntry{
				{Quantity: 2, Name: "Brainstorm"},
				{Quantity: 3, Name: "Ponder"},
			},
		},
		{
			name:  "strips price column after tab",
			input: "1\tBlack Lotus\t\t$ 99999.00",
			want:  []DeckListEntry{{Quantity: 1, Name: "Black Lotus"}},
		},
		{
			name:  "ignores cards total footer line",
			input: "4 Thoughtseize\n100 Cards Total",
			want:  []DeckListEntry{{Quantity: 4, Name: "Thoughtseize"}},
		},
		{
			name:  "ignores 99 cards footer",
			input: "1 Mana Crypt\n99 Cards",
			want:  []DeckListEntry{{Quantity: 1, Name: "Mana Crypt"}},
		},
		{
			name:  "invalid quantity line skipped",
			input: "abc Lightning Bolt\n2 Dark Ritual",
			want:  []DeckListEntry{{Quantity: 2, Name: "Dark Ritual"}},
		},
		{
			name:  "zero quantity skipped",
			input: "0 Ancestral Recall\n1 Mox Sapphire",
			want:  []DeckListEntry{{Quantity: 1, Name: "Mox Sapphire"}},
		},
		{
			name:  "negative quantity skipped",
			input: "-1 Timetwister\n1 Mox Pearl",
			want:  []DeckListEntry{{Quantity: 1, Name: "Mox Pearl"}},
		},
		{
			name:  "card with no name after number skipped",
			input: "4\n3 Swords to Plowshares",
			want:  []DeckListEntry{{Quantity: 3, Name: "Swords to Plowshares"}},
		},
		{
			name:  "trimming leading/trailing spaces in card name",
			input: "2   Wrath of God",
			want:  []DeckListEntry{{Quantity: 2, Name: "Wrath of God"}},
		},
		{
			name:  "large quantity",
			input: "100 Basic Plains",
			want:  []DeckListEntry{{Quantity: 100, Name: "Basic Plains"}},
		},
		{
			name:  "multiword card names preserved",
			input: "1 Emrakul, the Aeons Torn",
			want:  []DeckListEntry{{Quantity: 1, Name: "Emrakul, the Aeons Torn"}},
		},
		{
			name:  "windows line endings (CRLF)",
			input: "2 Island\r\n3 Mountain",
			want: []DeckListEntry{
				{Quantity: 2, Name: "Island"},
				{Quantity: 3, Name: "Mountain"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseDeckList(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("ParseDeckList(%q) returned %d entries, want %d\ngot:  %v\nwant: %v",
					tc.input, len(got), len(tc.want), got, tc.want)
			}
			for i := range got {
				if got[i].Quantity != tc.want[i].Quantity || got[i].Name != tc.want[i].Name {
					t.Errorf("entry[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}
