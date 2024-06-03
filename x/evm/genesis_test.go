package evm_test

import (
	"math/big"
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/crypto/ethsecp256k1"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	evmostypes "github.com/evmos/evmos/v18/types"
	"github.com/evmos/evmos/v18/x/evm"
	"github.com/evmos/evmos/v18/x/evm/statedb"
	"github.com/evmos/evmos/v18/x/evm/types"
	"github.com/stretchr/testify/require"
)

type GenesisTestSuite struct {
	keyring testkeyring.Keyring
	network *testnetwork.UnitTestNetwork
}

func SetupTest() *GenesisTestSuite {
	keyring := testkeyring.New(1)
	network := testnetwork.NewUnitTestNetwork(
		testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	return &GenesisTestSuite{
		keyring: keyring,
		network: network,
	}
}

func TestInitGenesis(t *testing.T) {
	privkey, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err, "failed to generate private key")

	address := common.HexToAddress(privkey.PubKey().Address().String())

	var vmdb *statedb.StateDB

	testCases := []struct {
		name     string
		malleate func(*testnetwork.UnitTestNetwork)
		genState *types.GenesisState
		expPanic bool
	}{
		{
			name:     "default",
			malleate: func(_ *testnetwork.UnitTestNetwork) {},
			genState: types.DefaultGenesisState(),
			expPanic: false,
		},
		{
			name: "valid account",
			malleate: func(_ *testnetwork.UnitTestNetwork) {
				vmdb.AddBalance(address, big.NewInt(1))
			},
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Accounts: []types.GenesisAccount{
					{
						Address: address.String(),
						Storage: types.Storage{
							{Key: common.BytesToHash([]byte("key")).String(), Value: common.BytesToHash([]byte("value")).String()},
						},
					},
				},
			},
			expPanic: false,
		},
		{
			name:     "account not found",
			malleate: func(_ *testnetwork.UnitTestNetwork) {},
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Accounts: []types.GenesisAccount{
					{
						Address: address.String(),
					},
				},
			},
			expPanic: true,
		},
		{
			name: "invalid account type",
			malleate: func(network *testnetwork.UnitTestNetwork) {
				acc := authtypes.NewBaseAccountWithAddress(address.Bytes())
				network.App.AccountKeeper.SetAccount(network.GetContext(), acc)
			},
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Accounts: []types.GenesisAccount{
					{
						Address: address.String(),
					},
				},
			},
			expPanic: true,
		},
		{
			name: "invalid code hash",
			malleate: func(network *testnetwork.UnitTestNetwork) {
				ctx := network.GetContext()
				acc := network.App.AccountKeeper.NewAccountWithAddress(ctx, address.Bytes())
				network.App.AccountKeeper.SetAccount(ctx, acc)
			},
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Accounts: []types.GenesisAccount{
					{
						Address: address.String(),
						Code:    "ffffffff",
					},
				},
			},
			expPanic: true,
		},
		{
			name: "ignore empty account code checking",
			malleate: func(network *testnetwork.UnitTestNetwork) {
				ctx := network.GetContext()
				acc := network.App.AccountKeeper.NewAccountWithAddress(ctx, address.Bytes())
				network.App.AccountKeeper.SetAccount(ctx, acc)
			},
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Accounts: []types.GenesisAccount{
					{
						Address: address.String(),
						Code:    "",
					},
				},
			},
			expPanic: false,
		},
		{
			name: "ignore empty account code checking with non-empty codehash",
			malleate: func(network *testnetwork.UnitTestNetwork) {
				ethAcc := &evmostypes.EthAccount{
					BaseAccount: authtypes.NewBaseAccount(address.Bytes(), nil, 0, 0),
					CodeHash:    common.BytesToHash([]byte{1, 2, 3}).Hex(),
				}

				network.App.AccountKeeper.SetAccount(network.GetContext(), ethAcc)
			},
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Accounts: []types.GenesisAccount{
					{
						Address: address.String(),
						Code:    "",
					},
				},
			},
			expPanic: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := SetupTest()
			ctx := ts.network.GetContext()

			vmdb = statedb.New(
				ctx,
				ts.network.App.EvmKeeper,
				statedb.NewEmptyTxConfig(common.BytesToHash(ctx.HeaderHash().Bytes())),
			)

			tc.malleate(ts.network)
			err := vmdb.Commit()
			require.NoError(t, err, "failed to commit to state db")

			if tc.expPanic {
				require.Panics(t, func() {
					_ = evm.InitGenesis(
						ts.network.GetContext(),
						ts.network.App.EvmKeeper,
						ts.network.App.AccountKeeper,
						*tc.genState,
					)
				})
			} else {
				require.NotPanics(t, func() {
					_ = evm.InitGenesis(
						ctx,
						ts.network.App.EvmKeeper,
						ts.network.App.AccountKeeper,
						*tc.genState,
					)
				})
			}
		})
	}
}
