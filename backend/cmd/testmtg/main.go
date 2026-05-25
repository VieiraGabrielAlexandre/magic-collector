package main

import (
	"fmt"
	"magic-collection-api/internal/mtgapi"
)

func main() {
	c := mtgapi.NewClient()

	type testCase struct {
		desc, set, number, lang, artist string
		wantName                        string
	}

	tests := []testCase{
		{
			desc:     "carta PT (Aeroesquife = Sky Skiff KLD #233)",
			set: "KLD", number: "233", lang: "PT", artist: "Richard Wright",
			wantName: "Sky Skiff / Aeroesquife",
		},
		{
			desc:     "carta PT com artista inválido deve rejeitar PT e retornar EN",
			set: "KLD", number: "233", lang: "PT", artist: "Artista Errado",
			wantName: "Sky Skiff (EN fallback)",
		},
		{
			desc:     "carta EN (Janjeet Sentry KLD #53)",
			set: "KLD", number: "53", lang: "EN", artist: "Dan Scott",
			wantName: "Janjeet Sentry",
		},
		{
			desc:     "carta artista parcial (Dan Scott ↔ Dan Murayama Scott)",
			set: "KLD", number: "53", lang: "EN", artist: "Dan Scott",
			wantName: "Janjeet Sentry",
		},
		{
			desc:     "set/número inexistente deve retornar nil",
			set: "KLD", number: "9999", lang: "EN", artist: "",
			wantName: "(sem resultado)",
		},
	}

	for _, tt := range tests {
		card, err := c.Search(tt.set, tt.number, tt.lang, tt.artist)
		status := "NÃO ENCONTROU"
		if err != nil {
			status = "ERRO: " + err.Error()
		} else if card != nil {
			en := card.Name
			pt := card.PrintedName
			if pt != "" {
				status = fmt.Sprintf("encontrou: %s / %s (artista: %s, lang: %s)", en, pt, card.Artist, tt.lang)
			} else {
				status = fmt.Sprintf("encontrou: %s (artista: %s)", en, card.Artist)
			}
		}
		fmt.Printf("[%s]\n   → %s\n\n", tt.desc, status)
	}
}
