package cards

import "testing"

func TestColorsJSONToDisplay(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"null literal", "null", ""},
		{"empty array literal", "[]", ""},
		{"white only", `["W"]`, "Branco"},
		{"blue only", `["U"]`, "Azul"},
		{"black only", `["B"]`, "Preto"},
		{"red only", `["R"]`, "Vermelho"},
		{"green only", `["G"]`, "Verde"},
		{"colorless only", `["C"]`, "Incolor"},
		{"white/blue (azorius)", `["W","U"]`, "Branco/Azul"},
		{"blue/red (izzet)", `["U","R"]`, "Azul/Vermelho"},
		{"black/green (golgari)", `["B","G"]`, "Preto/Verde"},
		{"white/blue/black (esper)", `["W","U","B"]`, "Branco/Azul/Preto"},
		{"four-color (sans-green)", `["W","U","B","R"]`, "Branco/Azul/Preto/Vermelho"},
		{"five-color (WUBRG)", `["W","U","B","R","G"]`, "Branco/Azul/Preto/Vermelho/Verde"},
		{"invalid json", "not-json", ""},
		{"unknown code ignored", `["X","W"]`, "Branco"},
		{"only unknown codes", `["X","Y"]`, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ColorsJSONToDisplay(tc.input)
			if got != tc.want {
				t.Errorf("ColorsJSONToDisplay(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
