package v16_test

import (
	"math/big"
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	v16 "github.com/evmos/evmos/v16/app/upgrades/v16"
	"github.com/evmos/evmos/v16/contracts"
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

	expectedBalances   utils.ExpectedBalances
	tokenPair          *erc20types.TokenPair
	nonNativeTokenPair *erc20types.TokenPair
}

const (
	AEVMOS = "aevmos"
	XMPL   = "xmpl"

	testAccount   = 0
	erc20Deployer = testAccount + 1
)

// mintAmount is the amount of tokens to be minted for a non-native ERC20 contract.
var mintAmount = big.NewInt(5e18)

func TestConvertERC20Coins(t *testing.T) {
	ts, err := SetupConvertERC20CoinsTest(t)
	require.NoError(t, err, "failed to setup test")

	logger := ts.network.GetContext().Logger().With("upgrade")

	// Convert the coins back using the upgrade util
	err = v16.ConvertERC20Coins(ts.network.GetContext(), logger, ts.network.App.AccountKeeper, ts.network.App.BankKeeper, ts.network.App.Erc20Keeper)
	require.NoError(t, err, "failed to convert coins")

	err = ts.network.NextBlock()
	require.NoError(t, err, "failed to execute block")

	// NOTE: Here we check that the ERC20 converted coins have been added back to the bank balance.
	err = utils.CheckBalances(ts.handler, utils.ExpectedBalances{
		{Address: ts.keyring.GetAccAddr(testAccount), Coins: sdk.NewCoins(sdk.NewInt64Coin(XMPL, 300))},
		{Address: ts.keyring.GetAccAddr(erc20Deployer), Coins: sdk.NewCoins(sdk.NewInt64Coin(XMPL, 200))},
	})
	require.NoError(t, err, "failed to check balances")

	// NOTE: We check that the ERC20 contract for the token pair has been removed
	balance, err := GetERC20Balance(ts.factory, ts.keyring.GetPrivKey(testAccount), ts.tokenPair.GetERC20Contract())
	require.NoError(t, err, "failed to execute contract call")
	require.Equal(t, int64(0), balance.Int64(), "expected different balance after converting ERC20")

	// NOTE: We check that the balance of the module address is empty after converting native ERC20s
	balances := ts.network.App.BankKeeper.GetAllBalances(ts.network.GetContext(), authtypes.NewModuleAddress(erc20types.ModuleName))
	require.True(t, balances.IsZero(), "expected different balance for module account")

	// NOTE: We check that the erc20deployer account still has the minted balance after converting the native ERC20s only.
	balance, err = GetERC20Balance(ts.factory, ts.keyring.GetPrivKey(erc20Deployer), ts.nonNativeTokenPair.GetERC20Contract())
	require.NoError(t, err, "failed to execute contract call")
	require.Equal(t, mintAmount, balance, "expected different balance after converting ERC20")
}

// SetupConvertERC20CoinsTest sets up a test suite to test the conversion of ERC20 coins to native coins.
//
// It sets up a basic integration test suite, with two accounts, that contain balances in a native non-Evmos coin.
// This coin is registered as an ERC20 token pair, and a portion of the initial balance is converted to the ERC20
// representation.
func SetupConvertERC20CoinsTest(t *testing.T) (ConvertERC20CoinsTestSuite, error) {
	kr := testkeyring.New(2)
	fundedBalances := utils.ExpectedBalances{
		utils.ExpectedBalance{
			Address: kr.GetAccAddr(testAccount),
			Coins: sdk.NewCoins(
				sdk.NewCoin(AEVMOS, network.PrefundedAccountInitialBalance),
				sdk.NewInt64Coin(XMPL, 300),
			),
		},
		utils.ExpectedBalance{
			Address: kr.GetAccAddr(erc20Deployer),
			Coins: sdk.NewCoins(
				sdk.NewCoin(AEVMOS, network.PrefundedAccountInitialBalance),
				sdk.NewInt64Coin(XMPL, 200),
			),
		},
	}

	nw := network.NewUnitTestNetwork(
		network.WithBalances(fundedBalances.ToBalances()...),
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
	balance, err := GetERC20Balance(txFactory, kr.GetPrivKey(testAccount), tokenPair.GetERC20Contract())
	require.NoError(t, err, "failed to execute contract call")
	require.Equal(t, common.Big0.Int64(), balance.Int64(), "expected different balance initially")

	err = nw.NextBlock()
	require.NoError(t, err, "failed to execute block")

	// Convert the native coins to the ERC20 representation
	senderIdx := 0
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
	balance, err = GetERC20Balance(txFactory, kr.GetPrivKey(testAccount), tokenPair.GetERC20Contract())
	require.NoError(t, err, "failed to execute contract call")
	require.Equal(t, big.NewInt(100), balance, "expected different balance after converting ERC20")

	// NOTE: We check that the balances have been adjusted to remove 100 XMPL from the bank balance after
	// converting to ERC20s.
	err = utils.CheckBalances(handler, utils.ExpectedBalances{
		{Address: kr.GetAccAddr(testAccount), Coins: sdk.NewCoins(sdk.NewInt64Coin(XMPL, 200))},
		{Address: kr.GetAccAddr(erc20Deployer), Coins: sdk.NewCoins(sdk.NewInt64Coin(XMPL, 200))},
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
	balance, err = GetERC20Balance(txFactory, kr.GetPrivKey(erc20Deployer), erc20Addr)
	require.NoError(t, err, "failed to execute contract call")
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

	return ConvertERC20CoinsTestSuite{
		keyring:            kr,
		network:            nw,
		handler:            handler,
		factory:            txFactory,
		expectedBalances:   fundedBalances,
		tokenPair:          tokenPair,
		nonNativeTokenPair: &nonNativeTokenPair,
	}, nil
}

// GetERC20Balance is a helper method to return the balance of the given ERC20 contract for the given address.
func GetERC20Balance(txFactory testfactory.TxFactory, priv cryptotypes.PrivKey, contractAddr common.Address) (*big.Int, error) {
	addrBytes := priv.PubKey().Address().Bytes()
	addr := common.BytesToAddress(addrBytes)
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI

	balanceOfArgs := testfactory.CallArgs{
		ContractABI: erc20ABI,
		MethodName:  "balanceOf",
		Args:        []interface{}{addr},
	}

	// TODO: should be done with EthCall instead of transaction
	res, err := txFactory.ExecuteContractCall(
		priv,
		evmtypes.EvmTxArgs{To: &contractAddr},
		balanceOfArgs,
	)
	if err != nil {
		return nil, err
	}

	ethRes, err := evmtypes.DecodeTxResponse(res.Data)
	if err != nil {
		return nil, err
	}

	var balance *big.Int
	err = erc20ABI.UnpackIntoInterface(&balance, "balanceOf", ethRes.Ret)
	if err != nil {
		return nil, err
	}

	return balance, nil
}
