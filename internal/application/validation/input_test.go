package validation

import (
	"testing"

	"github.com/stretchr/testify/require"

	domainerror "github.com/handiism/infinita/internal/domain/error"
	"github.com/handiism/infinita/internal/testutil/assertdomain"
)

func TestNormalizeCategory(t *testing.T) {
	got, err := NormalizeCategory("  Food  ")
	require.NoError(t, err)
	require.Equal(t, "food", got)

	_, err = NormalizeCategory("")
	assertdomain.Code(t, err, domainerror.ErrInvalidCategory.Code)

	_, err = NormalizeCategory("   ")
	assertdomain.Code(t, err, domainerror.ErrInvalidCategory.Code)
}

func TestParseEntryType(t *testing.T) {
	got, err := ParseEntryType("income")
	require.NoError(t, err)
	require.Equal(t, "income", got)

	got, err = ParseEntryType(" expense ")
	require.NoError(t, err)
	require.Equal(t, "expense", got)

	_, err = ParseEntryType("transfer")
	assertdomain.Code(t, err, domainerror.ErrInvalidTransactionType.Code)

	_, err = ParseEntryType("")
	assertdomain.Code(t, err, domainerror.ErrInvalidTransactionType.Code)
}

func TestParseISODate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{name: "valid date", input: "2026-03-15", want: "2026-03-15"},
		{name: "trims whitespace", input: " 2026-03-15 ", want: "2026-03-15"},
		{name: "leap year valid", input: "2024-02-29", want: "2024-02-29"},
		{name: "rejects Feb 30", input: "2024-02-30", wantErr: domainerror.ErrInvalidDate.Code},
		{name: "rejects Feb 29 non-leap", input: "2025-02-29", wantErr: domainerror.ErrInvalidDate.Code},
		{name: "rejects Apr 31", input: "2026-04-31", wantErr: domainerror.ErrInvalidDate.Code},
		{name: "rejects month 13", input: "2026-13-01", wantErr: domainerror.ErrInvalidDate.Code},
		{name: "rejects month 00", input: "2026-00-15", wantErr: domainerror.ErrInvalidDate.Code},
		{name: "rejects day 00", input: "2026-03-00", wantErr: domainerror.ErrInvalidDate.Code},
		{name: "rejects empty", input: "", wantErr: domainerror.ErrInvalidDate.Code},
		{name: "rejects garbage", input: "not-a-date", wantErr: domainerror.ErrInvalidDate.Code},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseISODate(tt.input)
			if tt.wantErr != "" {
				assertdomain.Code(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseISOMonth(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{name: "valid month", input: "2026-03", want: "2026-03"},
		{name: "trims whitespace", input: " 2026-03 ", want: "2026-03"},
		{name: "rejects month 13", input: "2026-13", wantErr: domainerror.ErrInvalidMonth.Code},
		{name: "rejects month 00", input: "2026-00", wantErr: domainerror.ErrInvalidMonth.Code},
		{name: "rejects empty", input: "", wantErr: domainerror.ErrInvalidMonth.Code},
		{name: "rejects full date", input: "2026-03-15", wantErr: domainerror.ErrInvalidMonth.Code},
		{name: "rejects garbage", input: "abc", wantErr: domainerror.ErrInvalidMonth.Code},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseISOMonth(tt.input)
			if tt.wantErr != "" {
				assertdomain.Code(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseTimezone(t *testing.T) {
	got, err := ParseTimezone("Asia/Jakarta")
	require.NoError(t, err)
	require.Equal(t, "Asia/Jakarta", got)

	got, err = ParseTimezone(" UTC ")
	require.NoError(t, err)
	require.Equal(t, "UTC", got)

	_, err = ParseTimezone("")
	assertdomain.Code(t, err, domainerror.ErrInvalidTimezone.Code)

	_, err = ParseTimezone("Not/A/Zone")
	assertdomain.Code(t, err, domainerror.ErrInvalidTimezone.Code)
}
