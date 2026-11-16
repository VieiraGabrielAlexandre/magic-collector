package reports

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// brt is Brasília time. Brazil has had no DST since 2019, so a fixed zone
// avoids depending on tzdata in the Lambda image.
var brt = time.FixedZone("BRT", -3*60*60)

const (
	msgTopExpensive = 5
	msgRecentDecks  = 3
	msgRecentCards  = 5
	msgMaxName      = 40
)

var colorEmoji = map[string]string{
	"W": "⚪", "U": "🔵", "B": "⚫", "R": "🔴", "G": "🟢", "Multi": "🌈", "C": "💎",
}

var rarityOrder = []struct{ code, label string }{
	{"M", "Míticas"}, {"R", "Raras"}, {"U", "Incomuns"}, {"C", "Comuns"}, {"L", "Terrenos"}, {"T", "Tokens"},
}

var medals = []string{"🥇", "🥈", "🥉", "4️⃣", "5️⃣"}

// FormatWhatsApp renders the report as a ready-to-send WhatsApp message
// (WhatsApp markup: *bold*, _italic_). Consumed unchanged by ArcanePulse.
func FormatWhatsApp(r CollectionReport) string {
	var b strings.Builder
	s := r.Summary

	b.WriteString("🃏 *Magic Collector — Resumo da Coleção*\n")
	if t, err := time.Parse(time.RFC3339, r.GeneratedAt); err == nil {
		fmt.Fprintf(&b, "📅 %s\n", t.In(brt).Format("02/01/2006 às 15:04"))
	}

	b.WriteString("\n📊 *Visão geral*\n")
	fmt.Fprintf(&b, "🗂️ Cartas: *%s* (%s únicas)\n", fmtInt(s.TotalCards), fmtInt(s.TotalUnique))
	fmt.Fprintf(&b, "💰 Valor estimado: *%s*\n", fmtUSD(s.TotalValueUSD))
	fmt.Fprintf(&b, "🧱 Decks: *%s*\n", fmtInt(s.TotalDecks))
	fmt.Fprintf(&b, "✨ Foils: *%s*\n", fmtInt(s.FoilCount))

	b.WriteString("\n📦 *Organização*\n")
	fmt.Fprintf(&b, "✅ Em decks: %s únicas\n", fmtInt(s.CardsInDeck))
	fmt.Fprintf(&b, "📥 Sem deck: %s únicas\n", fmtInt(s.CardsWithoutDeck))

	var rar []string
	for _, ro := range rarityOrder {
		if n := s.RarityBreakdown[ro.code]; n > 0 {
			rar = append(rar, fmt.Sprintf("%s: %s", ro.label, fmtInt(n)))
		}
	}
	if len(rar) > 0 {
		b.WriteString("\n💎 *Raridades*\n")
		b.WriteString(strings.Join(rar, " · "))
		b.WriteString("\n")
	}

	if len(r.ByColor) > 0 {
		b.WriteString("\n🎨 *Sem deck, por cor*\n")
		for _, c := range r.ByColor {
			emoji := colorEmoji[c.Color]
			if emoji == "" {
				emoji = "▫️"
			}
			name := c.ColorPT
			if name == "" {
				name = c.Color
			}
			fmt.Fprintf(&b, "%s %s: %s (%s cópias)\n", emoji, name, fmtInt(c.Count), fmtInt(c.Quantity))
		}
	}

	if len(r.TopExpensive) > 0 {
		b.WriteString("\n🏆 *Mais valiosas*\n")
		for i, c := range r.TopExpensive {
			if i >= msgTopExpensive {
				break
			}
			fmt.Fprintf(&b, "%s %s%s%s — *%s*\n", medals[i], short(c.Name), setTag(c.SetCode), foilTag(c.Foil), fmtUSD(c.PriceUSD))
		}
	}

	if len(r.RecentDecks) > 0 {
		b.WriteString("\n🆕 *Decks recentes*\n")
		for i, d := range r.RecentDecks {
			if i >= msgRecentDecks {
				break
			}
			fmt.Fprintf(&b, "• %s — %s cartas · %s\n", short(d.Name), fmtInt(d.CardCount), fmtUSD(d.EstimatedValue))
		}
	}

	if len(r.RecentCards) > 0 {
		b.WriteString("\n📬 *Últimas adições*\n")
		for i, c := range r.RecentCards {
			if i >= msgRecentCards {
				break
			}
			fmt.Fprintf(&b, "• %s%s%s\n", short(c.Name), setTag(c.SetCode), foilTag(c.Foil))
		}
	}

	if len(r.PriceAlerts) > 0 {
		b.WriteString("\n📈 *Variações de preço*\n")
		for _, a := range r.PriceAlerts {
			arrow := "⬆️"
			if a.Direction == "down" {
				arrow = "⬇️"
			}
			fmt.Fprintf(&b, "%s %s%s%s: %s → *%s* (%+.1f%%)\n",
				arrow, short(a.Name), setTag(a.SetCode), foilTag(a.Foil),
				fmtUSD(a.PrevPrice), fmtUSD(a.CurrentPrice), a.ChangePct)
		}
	}

	b.WriteString("\n_Enviado automaticamente a cada 12h_ 🤖")
	return b.String()
}

func setTag(code string) string {
	if code == "" {
		return ""
	}
	return " (" + strings.ToUpper(code) + ")"
}

func foilTag(foil bool) string {
	if foil {
		return " ✨"
	}
	return ""
}

// short trims long names and neutralises WhatsApp markup chars so a card
// name cannot break the formatting.
func short(s string) string {
	s = strings.NewReplacer("*", "", "_", "", "~", "", "`", "'", "\n", " ").Replace(strings.TrimSpace(s))
	if utf8.RuneCountInString(s) > msgMaxName {
		r := []rune(s)
		s = strings.TrimSpace(string(r[:msgMaxName-1])) + "…"
	}
	return s
}

// fmtInt formats 1234567 as "1.234.567" (pt-BR).
func fmtInt(n int) string {
	neg := n < 0
	if neg {
		n = -n
	}
	digits := strconv.Itoa(n)
	var out []byte
	for i, d := range []byte(digits) {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, d)
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}

// fmtUSD formats 1247.5 as "US$ 1.247,50" (pt-BR).
func fmtUSD(v float64) string {
	cents := int64(v*100 + 0.5)
	if v < 0 {
		cents = int64(v*100 - 0.5)
	}
	whole, frac := cents/100, cents%100
	if frac < 0 {
		frac = -frac
	}
	return fmt.Sprintf("US$ %s,%02d", fmtInt(int(whole)), frac)
}
