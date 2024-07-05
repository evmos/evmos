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

// This test checks is a safeguard to avoid missing precompiles that should be blocked addresses.
//
// It does so, by initially expecting all precompiles available in the EVM to be blocked, and
// require developers to specify exactly which should be an exception to this rule.
func TestPrecompilesAreBlockedAddrs(t *testing.T) {
	keyring := testkeyring.New(1)
	signer := keyring.GetKey(0)
	network := testnetwork.NewUnitTestNetwork(
		testnetwork.WithPreFundedAccounts(signer.AccAddr),
	)
	handler := grpc.NewIntegrationHandler(network)
	factory := testfactory.New(network, handler)

	// NOTE: all precompiles that should NOT be blocked addresses need to go in here
	//
	// For now there are no exceptions, so this slice is empty.
	var precompilesAbleToReceiveFunds []ethcommon.Address

	availablePrecompiles := network.App.EvmKeeper.GetAvailablePrecompileAddrs()
	for _, precompileAddr := range availablePrecompiles {
		t.Run(fmt.Sprintf("Cosmos Tx to %s\n", precompileAddr), func(t *testing.T) {
			_, err := factory.ExecuteCosmosTx(signer.Priv, commontestfactory.CosmosTxArgs{
				Msgs: []sdk.Msg{
					banktypes.NewMsgSend(
						signer.AccAddr,
						precompileAddr.Bytes(),
						sdk.NewCoins(sdk.NewCoin(network.GetDenom(), sdk.NewInt(1e10))),
					),
				},
			})

			require.NoError(t, network.NextBlock(), "failed to advance block")

			if slices.Contains(precompilesAbleToReceiveFunds, precompileAddr) {
				require.NoError(t, err, "failed to send funds to precompile %s that should not be blocked", precompileAddr)
			} else {
				require.Error(t, err, "was able to send funds to precompile %s that should be blocked", precompileAddr)
			}
		})

		t.Run(fmt.Sprintf("EVM Tx to %s\n", precompileAddr), func(t *testing.T) {
			_, err := factory.ExecuteEthTx(signer.Priv, evmtypes.EvmTxArgs{
				To:     &precompileAddr,
				Amount: big.NewInt(1e10),
			})

			require.NoError(t, network.NextBlock(), "failed to advance block")

			if slices.Contains(precompilesAbleToReceiveFunds, precompileAddr) {
				require.NoError(t, err, "failed to send funds with Eth transaction to precompile %s that should not be blocked", precompileAddr)
			} else {
				require.Error(t, err, "was able to send funds with Eth transaction to precompile %s that should be blocked", precompileAddr)
			}
		})
	}
}
