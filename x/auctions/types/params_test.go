package types_test

import (
	"testing"

	"github.com/evmos/evmos/v19/x/auctions/types"
	"github.com/stretchr/testify/require"
)

func TestParamsValidate(t *testing.T) {
	params := types.DefaultParams()

	require.NoError(t, params.Validate(), "expected no error with default values")
}
