package valueobject

import "testing"

func TestNormalizeCategoryKey(t *testing.T) {
	if got := NormalizeCategoryKey("  Food And Drink  "); got != "food and drink" {
		t.Fatalf("NormalizeCategoryKey() = %q, want %q", got, "food and drink")
	}
}

func TestParseEntryType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  EntryType
		ok    bool
	}{
		{name: "income", input: "income", want: EntryTypeIncome, ok: true},
		{name: "expense", input: "expense", want: EntryTypeExpense, ok: true},
		{name: "invalid", input: "transfer", want: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseEntryType(tt.input)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("ParseEntryType() = (%q, %t), want (%q, %t)", got, ok, tt.want, tt.ok)
			}
		})
	}
}
