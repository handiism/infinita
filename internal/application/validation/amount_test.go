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
		{
			name:      "parses whole number",
			input:     "123",
			allowZero: false,
			want:      12300,
		},
		{
			name:      "parses fractional amount",
			input:     "45.67",
			allowZero: false,
			want:      4567,
		},
		{
			name:      "rejects decimal point without digits",
			input:     ".",
			allowZero: true,
			wantErr:   true,
		},
		{
			name:      "accepts max int64 boundary",
			input:     "92233720368547758.07",
			allowZero: false,
			want:      math.MaxInt64,
		},
		{
			name:      "rejects overflow whole number",
			input:     "92233720368547759",
			allowZero: false,
			wantErr:   true,
		},
		{
			name:      "rejects overflow with fraction",
			input:     "92233720368547758.08",
			allowZero: false,
			wantErr:   true,
		},
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
