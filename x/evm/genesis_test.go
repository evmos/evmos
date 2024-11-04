package evm_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v20/contracts"
	"github.com/evmos/evmos/v20/crypto/ethsecp256k1"
	testfactory "github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	testhandler "github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v20/x/evm"
	"github.com/evmos/evmos/v20/x/evm/statedb"
	"github.com/evmos/evmos/v20/x/evm/types"
	"github.com/stretchr/testify/require"
)

type GenesisTestSuite struct {
	keyring testkeyring.Keyring
	network *testnetwork.UnitTestNetwork
	handler testhandler.Handler
	factory testfactory.TxFactory
}

func SetupTest() *GenesisTestSuite {
	keyring := testkeyring.New(1)
	network := testnetwork.NewUnitTestNetwork(
		testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	handler := testhandler.NewIntegrationHandler(network)
	factory := testfactory.New(network, handler)

	return &GenesisTestSuite{
		keyring: keyring,
		network: network,
		handler: handler,
		factory: factory,
	}
}

func TestInitGenesis(t *testing.T) {
	privkey, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err, "failed to generate private key")

	address := common.HexToAddress(privkey.PubKey().Address().String())

	var (
		vmdb *statedb.StateDB
		ctx  sdk.Context
	)

	testCases := []struct {
		name     string
		malleate func(*testnetwork.UnitTestNetwork)
		genState *types.GenesisState
		code     common.Hash
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
			name: "ignore empty account code checking",
			malleate: func(network *testnetwork.UnitTestNetwork) {
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
			name: "valid account with code",
			malleate: func(network *testnetwork.UnitTestNetwork) {
				acc := network.App.AccountKeeper.NewAccountWithAddress(ctx, address.Bytes())
				network.App.AccountKeeper.SetAccount(ctx, acc)
			},
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Accounts: []types.GenesisAccount{
					{
						Address: address.String(),
						Code:    "1234",
					},
				},
			},
			expPanic: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := SetupTest()
			ctx = ts.network.GetContext()

			vmdb = statedb.New(
				ctx,
				ts.network.App.EvmKeeper,
				statedb.NewEmptyTxConfig(common.BytesToHash(ctx.HeaderHash())),
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

				// If the initialization has not panicked we're checking the state
				for _, account := range tc.genState.Accounts {
					acc := ts.network.App.AccountKeeper.GetAccount(ctx, common.HexToAddress(account.Address).Bytes())
					require.NotNil(t, acc, "account not found in account keeper")

					expHash := crypto.Keccak256Hash(common.Hex2Bytes(account.Code))
					if account.Code == "" {
						expHash = common.BytesToHash(types.EmptyCodeHash)
					}

					require.Equal(t,
						expHash.String(),
						ts.network.App.EvmKeeper.GetCodeHash(
							ts.network.GetContext(),
							common.HexToAddress(account.Address),
						).String(),
						"code hash mismatch",
					)

					require.Equal(t,
						account.Code,
						common.Bytes2Hex(
							ts.network.App.EvmKeeper.GetCode(
								ts.network.GetContext(),
								expHash,
							),
						),
						"code mismatch",
					)

					for _, storage := range account.Storage {
						key := common.HexToHash(storage.Key)
						value := common.HexToHash(storage.Value)
						require.Equal(t, value, vmdb.GetState(common.HexToAddress(account.Address), key), "storage mismatch")
					}
				}
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	ts := SetupTest()

	contractAddr, err := ts.factory.DeployContract(
		ts.keyring.GetPrivKey(0),
		types.EvmTxArgs{},
		testfactory.ContractDeploymentData{
			Contract:        contracts.ERC20MinterBurnerDecimalsContract,
			ConstructorArgs: []interface{}{"TestToken", "TTK", uint8(18)},
		},
	)
	require.NoError(t, err, "failed to deploy contract")
	require.NoError(t, ts.network.NextBlock(), "failed to advance block")

	contractAddr2, err := ts.factory.DeployContract(
		ts.keyring.GetPrivKey(0),
		types.EvmTxArgs{},
		testfactory.ContractDeploymentData{
			Contract:        contracts.ERC20MinterBurnerDecimalsContract,
			ConstructorArgs: []interface{}{"AnotherToken", "ATK", uint8(18)},
		},
	)
	require.NoError(t, err, "failed to deploy contract")
	require.NoError(t, ts.network.NextBlock(), "failed to advance block")

	genState := evm.ExportGenesis(ts.network.GetContext(), ts.network.App.EvmKeeper)
	require.Len(t, genState.Accounts, 2, "expected only one smart contract in the exported genesis")

	genAddresses := make([]string, 0, len(genState.Accounts))
	for _, acc := range genState.Accounts {
		genAddresses = append(genAddresses, acc.Address)
	}
	require.Contains(t, genAddresses, contractAddr.Hex(), "expected contract 1 address in exported genesis")
	require.Contains(t, genAddresses, contractAddr2.Hex(), "expected contract 2 address in exported genesis")
}
