package v17_test

import (
	"fmt"
	"math/big"
	"testing"

	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/contracts"
	"github.com/evmos/evmos/v16/precompiles/werc20/testdata"
	testfactory "github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/utils"
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

	genesisState := createGenesisWithTokenPairs(kr)

	nw := network.NewUnitTestNetwork(
		network.WithCustomGenesis(genesisState),
		network.WithBalances(fundedBalances...),
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
