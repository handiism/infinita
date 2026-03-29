package valueobject

import (
	"strings"
	"testing"
)

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

func TestNewID(t *testing.T) {
	id1 := NewID()
	id2 := NewID()

	if len(id1) != 26 {
		t.Fatalf("NewID() length = %d, want 26", len(id1))
	}
	if id1 == id2 {
		t.Fatalf("NewID() returned duplicate IDs")
	}
	for _, c := range id1 {
		if !strings.ContainsRune("0123456789ABCDEFGHJKMNPQRSTVWXYZ", c) {
			t.Fatalf("NewID() contains invalid character %q", c)
		}
	}
}

func TestNewIDSortability(t *testing.T) {
	ids := make([]string, 100)
	for i := range ids {
		ids[i] = NewID()
	}
	for i := 1; i < len(ids); i++ {
		if ids[i] <= ids[i-1] {
			t.Fatalf("IDs not monotonically increasing at index %d: %q <= %q", i, ids[i], ids[i-1])
		}
	}
}
