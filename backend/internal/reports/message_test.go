package reports

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func sampleReport() CollectionReport {
	return CollectionReport{
		GeneratedAt:   "2026-11-14T17:23:00Z",
		SchemaVersion: "1.0",
		Summary: Summary{
			TotalCards: 843, TotalUnique: 420, TotalDecks: 12, TotalValueUSD: 1247.50,
			CardsWithoutDeck: 180, CardsInDeck: 240, FoilCount: 35,
			RarityBreakdown: map[string]int{"M": 12, "R": 48, "U": 130, "C": 220},
		},
		ByColor: []ColorBucket{
			{Color: "W", ColorPT: "Branco", Count: 22, Quantity: 44},
			{Color: "U", ColorPT: "Azul", Count: 31, Quantity: 57},
			{Color: "Multi", ColorPT: "Multicolor", Count: 9, Quantity: 12},
		},
		TopExpensive: []ExpensiveCard{
			{Name: "The One Ring", SetCode: "ltr", PriceUSD: 45, Foil: true},
			{Name: "Sheoldred, the Apocalypse", SetCode: "dmu", PriceUSD: 38.9},
			{Name: "Ragavan, Nimble Pilferer", SetCode: "mh2", PriceUSD: 32.15},
		},
		RecentDecks: []RecentDeck{{Name: "Mono Red Aggro", CardCount: 60, EstimatedValue: 112.4}},
		RecentCards: []RecentCard{{Name: "Lightning Bolt", SetCode: "m10", Foil: true}, {Name: "Counterspell", SetCode: "mh2"}},
	}
}

func TestFormatWhatsApp(t *testing.T) {
	msg := FormatWhatsApp(sampleReport())
	for _, want := range []string{
		"📅 14/11/2026 às 14:23",
		"🗂️ Cartas: *843* (420 únicas)",
		"💰 Valor estimado: *US$ 1.247,50*",
		"Míticas: 12 · Raras: 48 · Incomuns: 130 · Comuns: 220",
		"⚪ Branco: 22 (44 cópias)",
		"🥇 The One Ring (LTR) ✨ — *US$ 45,00*",
		"• Mono Red Aggro — 60 cartas · US$ 112,40",
		"• Lightning Bolt (M10) ✨",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("missing %q in:\n%s", want, msg)
		}
	}
	if n := utf8.RuneCountInString(msg); n > 4096 {
		t.Fatalf("message too long: %d", n)
	}
}

func TestFormatWhatsAppEmptySections(t *testing.T) {
	msg := FormatWhatsApp(CollectionReport{GeneratedAt: "bad"})
	for _, absent := range []string{"Mais valiosas", "Decks recentes", "Últimas adições", "📅"} {
		if strings.Contains(msg, absent) {
			t.Errorf("unexpected %q", absent)
		}
	}
}

func TestFormatHelpers(t *testing.T) {
	cases := map[string]string{fmtInt(0): "0", fmtInt(999): "999", fmtInt(1000): "1.000", fmtInt(1234567): "1.234.567",
		fmtUSD(0): "US$ 0,00", fmtUSD(1247.5): "US$ 1.247,50", fmtUSD(0.005): "US$ 0,01"}
	for got, want := range cases {
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
	if s := short("*Bold_* ~x~ " + strings.Repeat("a", 60)); strings.ContainsAny(s, "*_~") || utf8.RuneCountInString(s) > msgMaxName {
		t.Errorf("short() = %q", s)
	}
}

func TestPrintSample(t *testing.T) { t.Log("\n" + FormatWhatsApp(sampleReport())) }
