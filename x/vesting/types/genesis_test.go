package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	defaultParams := DefaultParams()
	defaultGenesis := DefaultGenesisState()
	require.Equal(t, defaultParams, defaultGenesis.Params, "expected default parameters to be in genesis state")

	gs := NewGenesisState(defaultParams)
	require.Equal(t, defaultGenesis, gs, "expected genesis state to be the default")
	require.NoError(t, gs.Validate(), "expected genesis state to pass validation")
}
