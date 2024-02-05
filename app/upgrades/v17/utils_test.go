package v17_test

import (
	"fmt"
	"math/big"
	"testing"

	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
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

	// nativeTokenPair is a registered token pair for a native Coin.
	nativeTokenPair erc20types.TokenPair
	// nonNativeTokenPair is a registered token pair for an ERC-20 native asset.
	nonNativeTokenPair erc20types.TokenPair
	// wevmosContract is the address of the deployed WEVMOS contract for testing purposes.
	wevmosContract common.Address
}

const (
	AEVMOS = "aevmos"
	XMPL   = "xmpl"

	// testAccount is the index for the main testing account that is sending
	testAccount = 0
	// erc20Deployer is the index for the account that deploys the ERC-20 contract.
	//
	// TODO: not really necessary??
	erc20Deployer = testAccount + 1
)

// SetupConvertERC20CoinsTest sets up a test suite to test the conversion of ERC-20 coins to native coins.
//
// It sets up a basic integration test suite with accounts, that contain balances in a native non-Evmos coin.
// This coin is registered as an ERC-20 token pair, and a portion of the initial balance is converted to the ERC-20
// representation.
//
// There is also another ERC-20 token pair (a NATIVE ERC-20) registered and some tokens minted to the test account.
// This is to show that non-native registered ERC20s are not converted and their balances still remain only in the EVM.
//
// TODO: set up the balance for the ERC-20 module address!
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
		{
			Address: bech32WithERC20s.String(),
			Coins: sdk.NewCoins(
				sdk.NewCoin(AEVMOS, network.PrefundedAccountInitialBalance),
				sdk.NewInt64Coin(XMPL, 500),
			),
		},
		// NOTE: Here, we are adding the ERC-20 module address to the funded balances,
		// since we have "converted coins to ERC-20s".
		// This initial balance is representative of the escrowed coins during this operation.
		{
			Address: types.NewModuleAddress(erc20types.ModuleName).String(),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(XMPL, 100)),
		},
	}

	genesisState := createGenesisWithTokenPairs(kr)

	nw := network.NewUnitTestNetwork(
		network.WithCustomGenesis(genesisState),
		network.WithBalances(fundedBalances...),
	)
	handler := grpc.NewIntegrationHandler(nw)
	txFactory := testfactory.New(nw, handler)

	// NOTE: we are adjusting the gov params to have a min deposit of 0 for the voting proposal.
	// This makes it simpler to register the token pair for the test.
	govParamsRes, err := handler.GetGovParams("voting")
	require.NoError(t, err, "failed to get gov params")
	govParams := govParamsRes.GetParams()
	govParams.MinDeposit = sdk.Coins{}

	err = nw.UpdateGovParams(*govParams)
	require.NoError(t, err, "failed to update gov params")

	res, err := handler.GetTokenPairs()
	require.NoError(t, err, "failed to get token pairs")
	require.NotNil(t, res, "failed to get token pairs")

	// nativeTokenPair is the token pair for the IBC native coin (XMPL).
	nativeTokenPair := res.TokenPairs[0]

	// NOTE: We deploy a standard ERC-20 to show that non-native registered ERC20s
	// are not converted and their balances still remain untouched.
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

	// We mint some tokens to the deployer address.
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

	// We check that the minting of tokens for the contract deployer has worked.
	balance, err := GetERC20Balance(txFactory, kr.GetPrivKey(erc20Deployer), erc20Addr)
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

	// NOTE: We deploy another smart contract. This is a wrapped token contract
	// as a representation of the WEVMOS token.
	wevmosAddr, err := txFactory.DeployContract(
		kr.GetPrivKey(testAccount),
		evmtypes.EvmTxArgs{},
		testfactory.ContractDeploymentData{Contract: testdata.WEVMOSContract},
	)
	require.NoError(t, err, "failed to deploy WEVMOS contract")

	err = nw.NextBlock()
	require.NoError(t, err, "failed to execute block")

	// send some coins to the wevmos address to deposit them.
	sentWEvmos := big.NewInt(1e18)
	_, err = txFactory.ExecuteEthTx(
		kr.GetPrivKey(testAccount),
		evmtypes.EvmTxArgs{
			To:     &wevmosAddr,
			Amount: sentWEvmos,
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
	require.Equal(t, sentWEvmos.String(), balance.String(), "expected different balance after minting ERC-20")

	return ConvertERC20CoinsTestSuite{
		keyring:            kr,
		network:            nw,
		handler:            handler,
		factory:            txFactory,
		nativeTokenPair:    nativeTokenPair,
		nonNativeTokenPair: nonNativeTokenPair,
		wevmosContract:     wevmosAddr,
	}, nil
}

// GetERC20Balance is a helper method to return the balance of the given ERC-20 contract for the given address.
func GetERC20Balance(txFactory testfactory.TxFactory, priv cryptotypes.PrivKey, erc20Addr common.Address) (*big.Int, error) {
	addr := common.BytesToAddress(priv.PubKey().Address().Bytes())

	return GetERC20BalanceForAddr(txFactory, priv, addr, erc20Addr)
}

// GetERC20BalanceForAddr is a helper method to return the balance of the given ERC-20 contract for the given address.
//
// NOTE: Under the hood this sends an actual EVM transaction instead of just querying the JSON-RPC.
// TODO: Use query instead of transaction in future.
func GetERC20BalanceForAddr(txFactory testfactory.TxFactory, priv cryptotypes.PrivKey, addr, erc20Addr common.Address) (*big.Int, error) {
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
