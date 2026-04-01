package validation

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	domainerror "github.com/handiism/infinita/internal/domain/error"
	"github.com/handiism/infinita/internal/testutil/assertdomain"
)

func TestParseAmount(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		allowZero bool
		want      int64
		wantErr   bool
	}{
		{name: "parses whole number", input: "123", want: 12300},
		{name: "parses fractional amount", input: "45.67", want: 4567},
		{name: "parses single fractional digit", input: "10.5", want: 1050},
		{name: "parses with leading spaces", input: "  100  ", want: 10000},
		{name: "parses zero with allowZero", input: "0", allowZero: true, want: 0},
		{name: "parses 0.00 with allowZero", input: "0.00", allowZero: true, want: 0},
		{name: "parses leading dot", input: ".50", want: 50},
		{name: "accepts max int64 boundary", input: "92233720368547758.07", want: math.MaxInt64},
		{name: "rejects empty string", input: "", wantErr: true},
		{name: "rejects whitespace only", input: "   ", wantErr: true},
		{name: "rejects decimal point only", input: ".", allowZero: true, wantErr: true},
		{name: "rejects more than 2 decimals", input: "10.123", wantErr: true},
		{name: "rejects three decimal digits", input: "1.001", wantErr: true},
		{name: "rejects multiple dots", input: "1.2.3", wantErr: true},
		{name: "rejects thousand separator", input: "1,000", wantErr: true},
		{name: "rejects alpha characters", input: "abc", wantErr: true},
		{name: "rejects mixed alpha-numeric", input: "12a", wantErr: true},
		{name: "rejects zero without allowZero", input: "0", wantErr: true},
		{name: "rejects negative sign", input: "-100", wantErr: true},
		{name: "rejects overflow whole number", input: "92233720368547759", wantErr: true},
		{name: "rejects overflow with fraction", input: "92233720368547758.08", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAmount(tt.input, tt.allowZero)
			if tt.wantErr {
				assertdomain.Code(t, err, domainerror.ErrInvalidAmount.Code)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
