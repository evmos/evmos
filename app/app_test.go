package app_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v18/app"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
)

func TestEvmosExport(t *testing.T) {
	nw := network.NewUnitTestNetwork()
	exported, err := nw.App.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")

	require.NotEmpty(t, exported.AppState)
	require.NotEmpty(t, exported.Validators)
	require.Equal(t, int64(2), exported.Height)
	require.Equal(t, *app.DefaultConsensusParams, exported.ConsensusParams)
}
