package app_test

import (
	"fmt"
	"math/big"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/evmos/evmos/v20/app"
	cmnfactory "github.com/evmos/evmos/v20/testutil/integration/common/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
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
	keyring := keyring.New(1)
	signer := keyring.GetKey(0)
	network := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(signer.AccAddr),
	)
	handler := grpc.NewIntegrationHandler(network)
	factory := factory.New(network, handler)

	// NOTE: all precompiles that should NOT be blocked addresses need to go in here
	//
	// For now there are no exceptions, so this slice is empty.
	var precompilesAbleToReceiveFunds []ethcommon.Address

	hexAvailablePrecompiles := network.App.EvmKeeper.GetParams(network.GetContext()).ActiveStaticPrecompiles
	availablePrecompiles := make([]ethcommon.Address, len(hexAvailablePrecompiles))
	for i, precompile := range hexAvailablePrecompiles {
		availablePrecompiles[i] = ethcommon.HexToAddress(precompile)
	}
	for _, precompileAddr := range availablePrecompiles {
		t.Run(fmt.Sprintf("Cosmos Tx to %s\n", precompileAddr), func(t *testing.T) {
			_, err := factory.ExecuteCosmosTx(signer.Priv, cmnfactory.CosmosTxArgs{
				Msgs: []sdk.Msg{
					banktypes.NewMsgSend(
						signer.AccAddr,
						precompileAddr.Bytes(),
						sdk.NewCoins(sdk.NewCoin(network.GetDenom(), math.NewInt(1e10))),
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
