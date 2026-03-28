package assertdomain

import (
	"testing"

	"github.com/stretchr/testify/require"

	domainerror "github.com/handiism/infinita/internal/domain/error"
)

func Code(t *testing.T, err error, code string) {
	t.Helper()
	require.Error(t, err)

	var domainErr domainerror.DomainError
	require.ErrorAs(t, err, &domainErr)
	require.Equal(t, code, domainErr.Code)
}
