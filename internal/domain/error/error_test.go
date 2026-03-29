package domainerror

import "testing"

func TestDomainErrorErrorFormats(t *testing.T) {
	tests := []struct {
		name string
		err  DomainError
		want string
	}{
		{
			name: "code and message only",
			err:  New("INVALID", "bad input"),
			want: "INVALID: bad input",
		},
		{
			name: "with field",
			err:  New("INVALID", "bad input").WithField("amount"),
			want: "INVALID: bad input (field=amount)",
		},
		{
			name: "with hint",
			err:  New("INVALID", "bad input").WithHint("must be numeric"),
			want: "INVALID: bad input (hint=must be numeric)",
		},
		{
			name: "with field and hint",
			err:  New("INVALID", "bad input").WithField("amount").WithHint("must be numeric"),
			want: "INVALID: bad input (field=amount, hint=must be numeric)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Fatalf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDomainErrorCloneHelpers(t *testing.T) {
	base := New("INVALID", "bad input")
	withField := base.WithField("amount")
	withHint := base.WithHint("must be numeric")

	if base.Field != "" || base.Hint != "" {
		t.Fatalf("base error mutated: %+v", base)
	}
	if withField.Field != "amount" || withField.Hint != "" {
		t.Fatalf("WithField() = %+v", withField)
	}
	if withHint.Field != "" || withHint.Hint != "must be numeric" {
		t.Fatalf("WithHint() = %+v", withHint)
	}
}
