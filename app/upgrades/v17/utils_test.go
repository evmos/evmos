package v17_test

import (
	"fmt"
	"math/big"
	"testing"

	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	accounttypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/contracts"
	"github.com/evmos/evmos/v16/precompiles/werc20/testdata"
	testfactory "github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/utils"
	"github.com/evmos/evmos/v16/types"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	"github.com/stretchr/testify/require"
)

type ConvertERC20CoinsTestSuite struct {
	keyring testkeyring.Keyring
	network *network.UnitTestNetwork
	handler grpc.Handler
	factory testfactory.TxFactory

	expectedBalances   []banktypes.Balance
	tokenPair          erc20types.TokenPair
	nonNativeTokenPair erc20types.TokenPair
	wevmosContract     common.Address
}

const (
	// TODO: rename to AEVMOS_DENOM and XMPL_DENOM
	AEVMOS = "aevmos"
	XMPL   = "xmpl"

	testAccount   = 0
	erc20Deployer = testAccount + 1

	// baseAccountAddress is the address of the base account that is set in genesis.
	baseAccountAddress = "evmos1k5c2zrqmmyx4rfkfp3l98qxvrpcr3m4kgxd0dp"
	// erc20TokenPairHex is the string representation of the ERC-20 token pair address.
	erc20TokenPairHex = "0x8E06a16247A9ff081DA2c8CAB90B4E1f455A800e"
)

var (
	// mintAmount is the amount of tokens to be minted for a non-native ERC-20 contract.
	mintAmount = big.NewInt(5e18)
	// SmartContractAddress is the bech32 address of the Ethereum smart contract account that is set in genesis.
	SmartContractAddress = sdk.AccAddress(common.HexToAddress(erc20TokenPairHex).Bytes()).String()
)

// SetupConvertERC20CoinsTest sets up a test suite to test the conversion of ERC-20 coins to native coins.
//
// It sets up a basic integration test suite, with two accounts, that contain balances in a native non-Evmos coin.
// This coin is registered as an ERC-20 token pair, and a portion of the initial balance is converted to the ERC-20
// representation.
//
// FIXME: this method is removed on the feature branch -> use custom genesis instead?
//
// Things to add in custom genesis:
// TODO: add token contract to EVM genesis
// TODO: add token pair to ERC-20 genesis
// TODO: add some converted balances for the token pair (-> this should mean that there's a balance in the ERC-20 module account)
func SetupConvertERC20CoinsTest(t *testing.T) (ConvertERC20CoinsTestSuite, error) {
	kr := testkeyring.New(2)
	fundedBalances := []banktypes.Balance{
		{
			Address: kr.GetAccAddr(testAccount).String(),
			Coins: sdk.NewCoins(
				sdk.NewCoin(AEVMOS, network.PrefundedAccountInitialBalance),
				sdk.NewInt64Coin(XMPL, 300),
			),
		},
		{
			Address: kr.GetAccAddr(erc20Deployer).String(),
			Coins: sdk.NewCoins(
				sdk.NewCoin(AEVMOS, network.PrefundedAccountInitialBalance),
				sdk.NewInt64Coin(XMPL, 200),
			),
		},
	}

	genesisState := createGenesisWithTokenPairs(kr.GetKey(testAccount))

	nw := network.NewUnitTestNetwork(
		network.WithBalances(fundedBalances...),
		network.WithCustomGenesis(genesisState),
	)
	handler := grpc.NewIntegrationHandler(nw)
	txFactory := testfactory.New(nw, handler)

	// TODO: Can these whole gov params adjustments be removed? Currently, this is needed to register the ERC-20 as an IBC coin.
	govParamsRes, err := handler.GetGovParams("voting")
	require.NoError(t, err, "failed to get gov params")
	govParams := govParamsRes.GetParams()
	govParams.MinDeposit = sdk.Coins{}
	err = nw.UpdateGovParams(*govParams)
	require.NoError(t, err, "failed to update gov params")

	res, err := handler.GetTokenPairs()
	require.NoError(t, err, "failed to get token pairs")
	require.NotNil(t, res, "failed to get token pairs")

	tokenPair := res.TokenPairs[0]
	fmt.Println("Got token pair with denom: ", tokenPair.Denom)

	// We check that the ERC-20 contract for the token pair shows the correct balance after
	// converting a portion of the native coins to their ERC-20 representation.
	balance, err := GetERC20Balance(txFactory, kr.GetPrivKey(testAccount), tokenPair.GetERC20Contract())
	require.NoError(t, err, "failed to query ERC-20 balance")
	require.Equal(t, big.NewInt(100), balance, "expected different balance after converting ERC-20")

	// NOTE: We check that the balances have been adjusted to remove 100 XMPL from the bank balance after
	// converting to ERC20s.
	err = utils.CheckBalances(handler, []banktypes.Balance{
		{Address: kr.GetAccAddr(testAccount).String(), Coins: sdk.NewCoins(sdk.NewInt64Coin(XMPL, 200))},
		{Address: kr.GetAccAddr(erc20Deployer).String(), Coins: sdk.NewCoins(sdk.NewInt64Coin(XMPL, 200))},
	})
	require.NoError(t, err, "failed to check balances")

	// NOTE: We deploy a standard ERC-20 to show that non-native registered ERC20s are not converted and their balances still remain.
	erc20Addr, err := txFactory.DeployContract(kr.GetPrivKey(erc20Deployer),
		evmtypes.EvmTxArgs{},
		testfactory.ContractDeploymentData{
			Contract: contracts.ERC20MinterBurnerDecimalsContract,
			ConstructorArgs: []interface{}{
				"MYTOKEN", "TKN", uint8(18),
			},
		},
	)
	require.NoError(t, err, "failed to deploy contract")

	err = nw.NextBlock()
	require.NoError(t, err, "failed to execute block")

	// we mint some tokens to the deployer address
	_, err = txFactory.ExecuteContractCall(
		kr.GetPrivKey(erc20Deployer), evmtypes.EvmTxArgs{To: &erc20Addr}, testfactory.CallArgs{
			ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
			MethodName:  "mint",
			Args:        []interface{}{kr.GetAddr(erc20Deployer), mintAmount},
		},
	)
	require.NoError(t, err, "failed to execute contract call")

	err = nw.NextBlock()
	require.NoError(t, err, "failed to execute block")

	// we check that the balance of the deployer address is correct
	balance, err = GetERC20Balance(txFactory, kr.GetPrivKey(erc20Deployer), erc20Addr)
	require.NoError(t, err, "failed to query ERC-20 balance")
	require.Equal(t, mintAmount, balance, "expected different balance after minting ERC-20")

	err = nw.NextBlock()
	require.NoError(t, err, "failed to execute block")

	// NOTE: We register the ERC-20 token as a token pair.
	nonNativeTokenPair, err := utils.RegisterERC20(txFactory, nw, utils.ERC20RegistrationData{
		Address:      erc20Addr,
		Denom:        "MYTOKEN",
		ProposerPriv: kr.GetPrivKey(testAccount),
	})
	require.NoError(t, err, "failed to register ERC-20 token")

	// NOTE: We deploy another a wrapped token contract as a representation of the WEVMOS token.
	wevmosAddr, err := txFactory.DeployContract(
		kr.GetPrivKey(testAccount),
		evmtypes.EvmTxArgs{},
		testfactory.ContractDeploymentData{Contract: testdata.WEVMOSContract},
	)
	require.NoError(t, err, "failed to deploy WEVMOS contract")

	err = nw.NextBlock()
	require.NoError(t, err, "failed to execute block")

	// send some coins to the wevmos address to deposit them.
	_, err = txFactory.ExecuteEthTx(
		kr.GetPrivKey(testAccount),
		evmtypes.EvmTxArgs{
			To:     &wevmosAddr,
			Amount: big.NewInt(1e18),
			// FIXME: the gas simulation is not working correctly - otherwise results in out of gas
			GasLimit: 100_000,
		},
	)
	require.NoError(t, err, "failed to execute eth tx")

	err = nw.NextBlock()
	require.NoError(t, err, "failed to execute block")

	// check that the WEVMOS balance has been increased
	balance, err = GetERC20Balance(txFactory, kr.GetPrivKey(testAccount), wevmosAddr)
	require.NoError(t, err, "failed to query ERC-20 balance")
	require.Equal(t, big.NewInt(1e18), balance, "expected different balance after minting ERC-20")

	return ConvertERC20CoinsTestSuite{
		keyring:            kr,
		network:            nw,
		handler:            handler,
		factory:            txFactory,
		expectedBalances:   fundedBalances,
		tokenPair:          tokenPair,
		nonNativeTokenPair: nonNativeTokenPair,
		wevmosContract:     wevmosAddr,
	}, nil
}

// GetERC20Balance is a helper method to return the balance of the given ERC-20 contract for the given address.
func GetERC20Balance(txFactory testfactory.TxFactory, priv cryptotypes.PrivKey, erc20Addr common.Address) (*big.Int, error) {
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI
	addr := common.BytesToAddress(priv.PubKey().Address().Bytes())

	txArgs := evmtypes.EvmTxArgs{
		To: &erc20Addr,
	}

	callArgs := testfactory.CallArgs{
		ContractABI: erc20ABI,
		MethodName:  "balanceOf",
		Args:        []interface{}{addr},
	}

	// TODO: should rather be done with EthCall
	res, err := txFactory.ExecuteContractCall(priv, txArgs, callArgs)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to execute contract call")
	}

	ethRes, err := evmtypes.DecodeTxResponse(res.Data)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to decode tx response")
	}
	if len(ethRes.Ret) == 0 {
		return nil, fmt.Errorf("got empty return value from contract call")
	}

	balanceI, err := erc20ABI.Unpack("balanceOf", ethRes.Ret)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to unpack balance")
	}

	balance, ok := balanceI[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("failed to convert balance to big.Int; got %T", balanceI[0])
	}

	return balance, nil
}

// createGenesisWithTokenPairs creates a genesis state that contains the state that cannot be created
// otherwise anymore. This is mainly the conversion of native Coins to their ERC-20 representation,
// which is now defunct with STR v2.
//
// That's why we need to create a custom genesis state that contains the ERC-20 balances and the token pairs
// for the native XMPL coin.
//
// The following elements have to be created:
//
//   - 1 token pair with IBC native denom
//   - The corresponding smart contract with its code and balances (ERC-20 balance)
//   - An ERC-20 balance for the xmpl denom which represents some native coins, that have been converted to the ERC-20 representation.
//
// NOTE: This assumes, that the SDK coin balances should be handled in the balances setup for
// the integration network.
func createGenesisWithTokenPairs(key testkeyring.Key) network.CustomGenesisState {
	// TODO: remove if not needed
	_ = key

	genesisAccounts := []accounttypes.AccountI{
		&accounttypes.BaseAccount{
			Address:       baseAccountAddress,
			PubKey:        nil,
			AccountNumber: 0,
			Sequence:      1,
		},
		&types.EthAccount{
			BaseAccount: &accounttypes.BaseAccount{
				Address:       SmartContractAddress,
				PubKey:        nil,
				AccountNumber: 9,
				Sequence:      1,
			},
			CodeHash: "307863313033656361656162303763323361303732336366653639633536376339383931643633346162363734313635666166396362333361303533616266656139",
		},
	}
	accGenesisState := accounttypes.DefaultGenesisState()
	for _, genesisAccount := range genesisAccounts {
		// NOTE: This type requires to be packed into a *types.Any as seen on SDK tests,
		// e.g. https://github.com/evmos/cosmos-sdk/blob/v0.47.5-evmos.2/x/auth/keeper/keeper_test.go#L193-L223
		accGenesisState.Accounts = append(accGenesisState.Accounts, codectypes.UnsafePackAny(genesisAccount))
	}

	// Add token pairs to genesis
	erc20GenesisState := erc20types.DefaultGenesisState()
	erc20GenesisState.TokenPairs = []erc20types.TokenPair{{
		Erc20Address:  erc20TokenPairHex,
		Denom:         XMPL,
		Enabled:       true,
		ContractOwner: erc20types.OWNER_MODULE, // NOTE: Owner is the module account since it's a native token and was registered through governance
	}}

	// Add the smart contracts to the EVM genesis
	evmGenesisState := evmtypes.DefaultGenesisState()
	evmGenesisState.Accounts = append(evmGenesisState.Accounts, evmtypes.GenesisAccount{
		// FIXME: This is currently not working -> panics with "account not found for address ..."
		Address: erc20TokenPairHex,
		// NOTE: This was generated using hexutil.Bytes(stateDBAccount.CodeHash) on the deployed contract
		Code: "0xc103ecaeab07c23a0723cfe69c567c9891d634ab674165faf9cb33a053abfea9",
		// TODO: how to get the correct contract storage that includes the address generated in the keyring?
		Storage: evmtypes.Storage{},
	})

	// Combine module genesis states
	return network.CustomGenesisState{
		accounttypes.ModuleName: accGenesisState,
		erc20types.ModuleName:   erc20GenesisState,
		// FIXME: Commented until account setup is fixed
		// evmtypes.ModuleName:     evmGenesisState,
	}
}

// This test asserts that we are generating the correct genesis state for the STR v2 migration logic tests.
// Specifically, this test should enable the scenario where we have token pairs registered in the ERC-20 module
// and some users with balances in the native denom and the token pair denoms.
//
// The users should have balances in both representations (SDK coins and ERC-20s).
// NOTE: This could also be done after genesis by interacting with the smart contracts, but would be good
// to already set up the scenario in the genesis state.
func TestCreateGenesisWithTokenPairs(t *testing.T) {
	// Create the custom genesis
	keyring := testkeyring.New(2)
	genesisState := createGenesisWithTokenPairs(keyring.GetKey(testAccount))

	// Instantiate the network
	unitNetwork := network.NewUnitTestNetwork(
		network.WithCustomGenesis(genesisState),
	)
	handler := grpc.NewIntegrationHandler(unitNetwork)
	tf := testfactory.New(unitNetwork, handler)
	_ = tf

	// Test that the token pairs are registered correctly
	res, err := handler.GetTokenPairs()
	require.NoError(t, err, "failed to get token pairs")
	require.NotNil(t, res, "failed to get token pairs")
	require.Equal(t, 1, len(res.TokenPairs), "expected different number of token pairs")
	require.Equal(t, XMPL, res.TokenPairs[0].Denom, fmt.Sprintf("expected different denom for %q token pair", XMPL))

	// Check that the base account exists
	expectedAccounts := []string{baseAccountAddress, SmartContractAddress}
	for i, expectedAccount := range expectedAccounts {
		acc, err := handler.GetAccount(expectedAccount)
		require.NoError(t, err, "failed to get account %d", i)
		require.NotNil(t, acc, "expected account %d to be not nil after genesis: %s", i, expectedAccount)
	}

	//// Test that the ERC-20 contract for the IBC native coin has the correct user balance after genesis.
	//balance, err := GetERC20Balance(tf, keyring.GetPrivKey(testAccount), res.TokenPairs[0].GetERC20Contract())
	//require.NoError(t, err, "failed to query ERC-20 balance")
	//require.Equal(t, big.NewInt(100), balance, "expected different ERC-20 balance after genesis")
	//
	//// NOTE: We check that the balances have been adjusted to remove 100 XMPL from the bank balance after
	//// converting to ERC20s.
	//err = utils.CheckBalances(handler, []banktypes.Balance{
	//	{Address: keyring.GetAccAddr(testAccount).String(), Coins: sdk.NewCoins(sdk.NewInt64Coin(XMPL, 200))},
	//	{Address: keyring.GetAccAddr(erc20Deployer).String(), Coins: sdk.NewCoins(sdk.NewInt64Coin(XMPL, 200))},
	//})
	//require.NoError(t, err, "failed to check balances")
}
