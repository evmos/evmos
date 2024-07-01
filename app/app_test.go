package app_test

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"slices"
	"testing"

	"cosmossdk.io/math"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/ibc-go/v7/testing/mock"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/app"
	"github.com/evmos/evmos/v18/encoding"
	commontestfactory "github.com/evmos/evmos/v18/testutil/integration/common/factory"
	testfactory "github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v18/utils"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestEvmosExport(t *testing.T) {
	// create public key
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err, "public key should be created without error")

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(100000000000000))),
	}

	db := dbm.NewMemDB()
	chainID := utils.MainnetChainID + "-1"
	evmosApp := app.NewEvmos(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
		db, nil, true, map[int64]bool{},
		app.DefaultNodeHome, 0,
		encoding.MakeConfig(app.ModuleBasics),
		simtestutil.NewAppOptionsWithFlagHome(app.DefaultNodeHome),
		baseapp.SetChainID(chainID),
	)

	genesisState := app.NewDefaultGenesisState()
	genesisState = app.GenesisStateWithValSet(evmosApp, genesisState, valSet, []authtypes.GenesisAccount{acc}, balance)
	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	// Initialize the chain
	evmosApp.InitChain(
		abci.RequestInitChain{
			ChainId:       chainID,
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	evmosApp.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := app.NewEvmos(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
		db, nil, true, map[int64]bool{},
		app.DefaultNodeHome, 0,
		encoding.MakeConfig(app.ModuleBasics),
		simtestutil.NewAppOptionsWithFlagHome(app.DefaultNodeHome),
		baseapp.SetChainID(chainID),
	)
	_, err = app2.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
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

	hexAvailablePrecompiles := network.App.EvmKeeper.GetParams(network.GetContext()).ActiveStaticPrecompiles
	availablePrecompiles := make([]ethcommon.Address, len(hexAvailablePrecompiles))
	for i, precompile := range hexAvailablePrecompiles {
		availablePrecompiles[i] = ethcommon.HexToAddress(precompile)
	}
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
