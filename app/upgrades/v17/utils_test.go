package v17

import (
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/contracts"
	"github.com/evmos/evmos/v16/precompiles/werc20/testdata"
	"github.com/evmos/evmos/v16/testutil/integration/common/factory"
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
	tokenPair          *erc20types.TokenPair
	nonNativeTokenPair *erc20types.TokenPair
	wevmosContract     common.Address
}

const (
	AEVMOS = "aevmos"
	XMPL   = "xmpl"

	testAccount   = 0
	erc20Deployer = testAccount + 1
)

// mintAmount is the amount of tokens to be minted for a non-native ERC20 contract.
var mintAmount = big.NewInt(5e18)

// SetupConvertERC20CoinsTest sets up a test suite to test the conversion of ERC20 coins to native coins.
//
// It sets up a basic integration test suite, with two accounts, that contain balances in a native non-Evmos coin.
// This coin is registered as an ERC20 token pair, and a portion of the initial balance is converted to the ERC20
// representation.
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

	nw := network.NewUnitTestNetwork(
		// TODO: we'll need to merge main into the feature branch to update this
		network.WithBalances(fundedBalances...),
	)
	handler := grpc.NewIntegrationHandler(nw)
	txFactory := testfactory.New(nw, handler)

	govParamsRes, err := handler.GetGovParams("voting")
	require.NoError(t, err, "failed to get gov params")
	govParams := govParamsRes.GetParams()
	govParams.MinDeposit = sdk.Coins{}

	err = nw.UpdateGovParams(*govParams)
	require.NoError(t, err, "failed to update gov params")

	// Register the coins
	XMPLMetadata := banktypes.Metadata{
		Name:        XMPL,
		Symbol:      XMPL,
		Description: "Example coin",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: XMPL, Exponent: 0},
			{Denom: "u" + XMPL, Exponent: 6},
		},
		Base:    XMPL,
		Display: XMPL,
	}

	tokenPair, err := nw.App.Erc20Keeper.RegisterCoin(nw.GetContext(), XMPLMetadata)
	require.NoError(t, err, "failed to register coin")

	err = nw.NextBlock()
	require.NoError(t, err, "failed to execute block")

	// Call the token pair contract to check the balance
	balance, err := GetERC20Balance(txFactory, kr.GetAddr(testAccount), tokenPair.GetERC20Contract())
	require.NoError(t, err, "failed to query ERC20 balance")
	require.Equal(t, common.Big0.Int64(), balance.Int64(), "expected different balance initially")

	err = nw.NextBlock()
	require.NoError(t, err, "failed to execute block")

	// Convert the native coins to the ERC20 representation
	senderIdx := 0
	// FIXME: this method is removed on the feature branch -> use custom genesis instead?
	msgConvert := &erc20types.MsgConvertCoin{
		Coin:     sdk.NewCoin(XMPL, sdk.NewInt(100)),
		Receiver: kr.GetAddr(senderIdx).String(),
		Sender:   kr.GetAccAddr(senderIdx).String(),
	}
	res, err := txFactory.ExecuteCosmosTx(kr.GetPrivKey(senderIdx), factory.CosmosTxArgs{Msgs: []sdk.Msg{msgConvert}})
	require.NoError(t, err, "failed to execute tx")
	require.NotNil(t, res, "failed to execute tx")

	err = nw.NextBlock()
	require.NoError(t, err, "failed to execute block")

	// We check that the ERC20 contract for the token pair shows the correct balance
	balance, err = GetERC20Balance(txFactory, kr.GetAddr(testAccount), tokenPair.GetERC20Contract())
	require.NoError(t, err, "failed to query ERC20 balance")
	require.Equal(t, big.NewInt(100), balance, "expected different balance after converting ERC20")

	// NOTE: We check that the balances have been adjusted to remove 100 XMPL from the bank balance after
	// converting to ERC20s.
	err = utils.CheckBalances(handler, []banktypes.Balance{
		{Address: kr.GetAccAddr(testAccount).String(), Coins: sdk.NewCoins(sdk.NewInt64Coin(XMPL, 200))},
		{Address: kr.GetAccAddr(erc20Deployer).String(), Coins: sdk.NewCoins(sdk.NewInt64Coin(XMPL, 200))},
	})
	require.NoError(t, err, "failed to check balances")

	// NOTE: We deploy a standard ERC20 to show that non-native registered ERC20s are not converted and their balances still remain.
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
	balance, err = GetERC20Balance(txFactory, kr.GetAddr(erc20Deployer), erc20Addr)
	require.NoError(t, err, "failed to query ERC20 balance")
	require.Equal(t, mintAmount, balance, "expected different balance after minting ERC20")

	err = nw.NextBlock()
	require.NoError(t, err, "failed to execute block")

	// NOTE: We register the ERC20 token as a token pair.
	nonNativeTokenPair, err := utils.RegisterERC20(txFactory, nw, utils.ERC20RegistrationData{
		Address:      erc20Addr,
		Denom:        "MYTOKEN",
		ProposerPriv: kr.GetPrivKey(testAccount),
	})
	require.NoError(t, err, "failed to register ERC20 token")

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
	balance, err = GetERC20Balance(txFactory, kr.GetAddr(testAccount), wevmosAddr)
	require.NoError(t, err, "failed to query ERC20 balance")
	require.Equal(t, big.NewInt(1e18), balance, "expected different balance after minting ERC20")

	return ConvertERC20CoinsTestSuite{
		keyring:            kr,
		network:            nw,
		handler:            handler,
		factory:            txFactory,
		expectedBalances:   fundedBalances,
		tokenPair:          tokenPair,
		nonNativeTokenPair: &nonNativeTokenPair,
		wevmosContract:     wevmosAddr,
	}, nil
}

// GetERC20Balance is a helper method to return the balance of the given ERC20 contract for the given address.
func GetERC20Balance(txFactory testfactory.TxFactory, addr, erc20Addr common.Address) (*big.Int, error) {
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI

	txArgs := evmtypes.EvmTxArgs{
		To: &erc20Addr,
	}

	callArgs := testfactory.CallArgs{
		ContractABI: erc20ABI,
		MethodName:  "balanceOf",
		Args:        []interface{}{addr},
	}

	// TODO: should rather be done with EthCall
	res, err := txFactory.ExecuteContractCall(nil, txArgs, callArgs)
	if err != nil {
		return nil, err
	}

	ethRes, err := evmtypes.DecodeTxResponse(res.Data)
	if err != nil {
		return nil, err
	}

	balanceI, err := erc20ABI.Unpack("balanceOf", ethRes.Ret)
	if err != nil {
		return nil, err
	}

	balance, ok := balanceI[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("failed to convert balance to big.Int; got %T", balanceI[0])
	}

	return balance, nil
}
