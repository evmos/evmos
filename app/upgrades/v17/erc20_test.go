package v17_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v17 "github.com/evmos/evmos/v16/app/upgrades/v17"
	erc20precompile "github.com/evmos/evmos/v16/precompiles/erc20"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/utils"
	"github.com/evmos/evmos/v16/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v16/x/evm/keeper"
	"github.com/stretchr/testify/require"
	"testing"
)

// SetupNetwork sets up a new test network and returns the network and the grpc handler.
func SetupNetwork() *testnetwork.UnitTestNetwork {
	keyring := testkeyring.New(1)

	return testnetwork.NewUnitTestNetwork(
		testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
}

func TestRegisterERC20Extensions(t *testing.T) {
	var network *testnetwork.UnitTestNetwork

	ibcDenom := utils.ComputeIBCDenom("transfer", "channel-0", "uosmo")
	ibcTokenPair := types.NewTokenPair(utiltx.GenerateAddress(), ibcDenom, types.OWNER_MODULE)
	nativeTokenPair := types.NewTokenPair(utiltx.GenerateAddress(), "test", types.OWNER_MODULE)
	otherNativeTokenPair := types.NewTokenPair(utiltx.GenerateAddress(), "other", types.OWNER_MODULE)
	externalTokenPair := types.NewTokenPair(utiltx.GenerateAddress(), "uext", types.OWNER_EXTERNAL)

	testcases := []struct {
		name        string
		malleate    func()
		expPass     bool
		errContains string
		postCheck   func()
	}{
		{
			name:    "pass - no token pairs in ERC20 keeper",
			expPass: true,
			postCheck: func() {
				requireActiveDynamicPrecompiles(t, network.GetContext(), network.App.EvmKeeper, []string{})
			},
		},
		{
			name: "pass - native token pair",
			malleate: func() {
				network.App.Erc20Keeper.SetTokenPair(network.GetContext(), nativeTokenPair)
			},
			expPass: true,
			postCheck: func() {
				// Check that the precompile was registered
				activeDynamicPrecompiles := network.App.EvmKeeper.GetParams(network.GetContext()).ActiveDynamicPrecompiles
				require.Contains(t, activeDynamicPrecompiles, nativeTokenPair.GetERC20Contract().String(), "expected precompile to be registered")

				// Check that the precompile is set as active
				requireActiveDynamicPrecompiles(
					t, network.GetContext(), network.App.EvmKeeper,
					[]string{nativeTokenPair.Erc20Address},
				)
			},
		},
		{
			name: "pass - IBC token pair",
			malleate: func() {
				network.App.Erc20Keeper.SetTokenPair(network.GetContext(), ibcTokenPair)
			},
			expPass: true,
			postCheck: func() {
				// Check that the precompile was registered
				activeDynamicPrecompiles := network.App.EvmKeeper.GetParams(network.GetContext()).ActiveDynamicPrecompiles
				require.Contains(t, activeDynamicPrecompiles, ibcTokenPair.GetERC20Contract().String(), "expected precompile to be registered")

				// Check that the precompile is set as active
				requireActiveDynamicPrecompiles(
					t, network.GetContext(), network.App.EvmKeeper,
					[]string{ibcTokenPair.Erc20Address},
				)
			},
		},
		{
			name: "pass - external token pair is skipped",
			malleate: func() {
				network.App.Erc20Keeper.SetTokenPair(network.GetContext(), externalTokenPair)
				network.App.Erc20Keeper.SetTokenPair(network.GetContext(), otherNativeTokenPair)
			},
			expPass: true,
			postCheck: func() {
				// Check that active precompiles are unchanged
				requireActiveDynamicPrecompiles(
					t, network.GetContext(), network.App.EvmKeeper,
					[]string{otherNativeTokenPair.Erc20Address},
				)
			},
		},
		{
			// TODO: this is not failing anymore in the new implementation? Is that what we want?
			name: "no-op - already registered precompile",
			malleate: func() {
				network.App.Erc20Keeper.SetTokenPair(network.GetContext(), nativeTokenPair)
				network.App.Erc20Keeper.SetTokenPair(network.GetContext(), otherNativeTokenPair)

				tokenPrecompile, err := erc20precompile.NewPrecompile(
					nativeTokenPair, network.App.BankKeeper, network.App.AuthzKeeper, network.App.TransferKeeper,
				)
				require.NoError(t, err, "expected no error creating precompile")

				// NOTE: We are adding this in the malleate function, so it is already present before
				// the real test case.
				err = network.App.EvmKeeper.AddDynamicPrecompiles(network.GetContext(), tokenPrecompile)
				require.NoError(t, err, "expected no error adding precompile")
			},
			expPass: true,
			postCheck: func() {
				// Check that active precompiles contain the already registered precompile only once
				requireActiveDynamicPrecompiles(
					t, network.GetContext(), network.App.EvmKeeper,
					[]string{nativeTokenPair.Erc20Address, otherNativeTokenPair.Erc20Address},
				)
			},
		},
		{
			name: "pass - evm denomination deploys erc20 contract",
			malleate: func() {
				params := network.App.EvmKeeper.GetParams(network.GetContext())
				params.EvmDenom = nativeTokenPair.Denom
				err := network.App.EvmKeeper.SetParams(network.GetContext(), params)
				require.NoError(t, err, "expected no error setting EVM params")

				network.App.Erc20Keeper.SetTokenPair(network.GetContext(), nativeTokenPair)
			},
			expPass: true,
			postCheck: func() {
				// Check that the precompile was registered
				requireActiveDynamicPrecompiles(
					t, network.GetContext(), network.App.EvmKeeper,
					[]string{nativeTokenPair.Erc20Address},
				)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			network = SetupNetwork()

			if tc.malleate != nil {
				tc.malleate()
			}

			err := v17.RegisterERC20Extensions(
				network.GetContext(),
				network.App.AuthzKeeper,
				network.App.BankKeeper,
				network.App.Erc20Keeper,
				network.App.EvmKeeper,
				network.App.TransferKeeper,
			)

			if tc.expPass {
				require.NoError(t, err, "expected no error registering ERC20 extensions")
			} else {
				require.Error(t, err, "expected an error registering ERC20 extensions")
				require.ErrorContains(t, err, tc.errContains, "expected different error message")
			}

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}

// requireActiveDynamicPrecompiles checks that the active dynamic precompiles
// in the EVM keeper match the expected precompiles.
func requireActiveDynamicPrecompiles(
	t *testing.T, ctx sdk.Context, evmKeeper *evmkeeper.Keeper, expPrecompiles []string,
) {
	activeDynamicPrecompiles := evmKeeper.GetParams(ctx).ActiveDynamicPrecompiles
	require.ElementsMatch(
		t, expPrecompiles, activeDynamicPrecompiles,
		"expected active precompiles to match",
	)
}
