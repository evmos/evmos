package v20_test

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v18/contracts"
	testfactory "github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/utils"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

const (
	AEVMOS = "aevmos"
	XMPL   = "xmpl"

	// erc20Deployer is the index for the account that deploys the ERC-20 contract.
	erc20Deployer = 0
)

// sentWEVMOS is the amount of WEVMOS sent to the WEVMOS contract during testing.
var sentWEVMOS = sdk.NewInt(1e18)

// NewConvertERC20CoinsTestSuite sets up a test suite to test the conversion of ERC-20 coins to native coins.
//
// It sets up a basic integration test suite with accounts, that contain balances in a native non-Evmos coin.
// This coin is registered as an ERC-20 token pair in the ERC-20 module keeper,
// and a portion of the initial balance is converted to the ERC-20 representation upon genesis already.
//
// This also means, that the ERC-20 module address has a balance of the escrowed ERC-20 token pair coins.
func NewConvertERC20CoinsTestSuite() (*ConvertERC20CoinsTestSuite, error) {
	kr := testkeyring.New(1)
	fundedBalances := []banktypes.Balance{
		{
			Address: kr.GetAccAddr(erc20Deployer).String(),
			Coins: sdk.NewCoins(
				sdk.NewCoin(AEVMOS, network.PrefundedAccountInitialBalance),
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

	return &ConvertERC20CoinsTestSuite{
		keyring: kr,
		network: nw,
		handler: handler,
		factory: txFactory,
	}, nil
}

// PrepareNetwork is a helper method to take care of the following steps to prepare the test suite for
// testing the STR v2 migrations:
//
//   - deploy an ERC-20 token contract (EVM-native!)
//   - register a token pair for this smart contract
//   - mint some tokens to the deployer account
//   - This is to show that non-native registered ERC20s are not converted and their balances still remain only in the EVM.
func PrepareNetwork(ts *ConvertERC20CoinsTestSuite) (*ConvertERC20CoinsTestSuite, error) {
	// NOTE: we are adjusting the gov params to have a min deposit of 0 for the voting proposal.
	// This makes it simpler to register the token pair for the test.
	govParamsRes, err := ts.handler.GetGovParams("voting")
	if err != nil {
		return &ConvertERC20CoinsTestSuite{}, errorsmod.Wrap(err, "failed to get gov params")
	}

	govParams := govParamsRes.GetParams()
	govParams.MinDeposit = sdk.Coins{}
	err = ts.network.UpdateGovParams(*govParams)
	if err != nil {
		return &ConvertERC20CoinsTestSuite{}, errorsmod.Wrap(err, "failed to update gov params")
	}

	// NOTE: We deploy a standard ERC-20 to show that non-native registered ERC20s
	// are not converted and their balances still remain untouched.
	erc20Addr, err := ts.factory.DeployContract(ts.keyring.GetPrivKey(erc20Deployer),
		evmtypes.EvmTxArgs{},
		testfactory.ContractDeploymentData{
			Contract: contracts.ERC20MinterBurnerDecimalsContract,
			ConstructorArgs: []interface{}{
				"MYTOKEN", "TKN", uint8(18),
			},
		},
	)
	if err != nil {
		return &ConvertERC20CoinsTestSuite{}, errorsmod.Wrap(err, "failed to deploy ERC-20 contract")
	}

	err = ts.network.NextBlock()
	if err != nil {
		return &ConvertERC20CoinsTestSuite{}, errorsmod.Wrap(err, "failed to execute block")
	}

	// We mint some tokens to the deployer address.
	_, err = ts.factory.ExecuteContractCall(
		ts.keyring.GetPrivKey(erc20Deployer), evmtypes.EvmTxArgs{To: &erc20Addr}, testfactory.CallArgs{
			ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
			MethodName:  "mint",
			Args:        []interface{}{ts.keyring.GetAddr(erc20Deployer), mintAmount},
		},
	)
	if err != nil {
		return &ConvertERC20CoinsTestSuite{}, errorsmod.Wrap(err, "failed to mint ERC-20 tokens")
	}

	err = ts.network.NextBlock()
	if err != nil {
		return &ConvertERC20CoinsTestSuite{}, errorsmod.Wrap(err, "failed to execute block")
	}

	// NOTE: We register the ERC-20 token as a token pair.
	nonNativeTokenPair, err := utils.RegisterERC20(ts.factory, ts.network, utils.ERC20RegistrationData{
		Address:      erc20Addr,
		Denom:        "MYTOKEN",
		ProposerPriv: ts.keyring.GetPrivKey(erc20Deployer),
	})
	if err != nil {
		return &ConvertERC20CoinsTestSuite{}, errorsmod.Wrap(err, "failed to register ERC-20 token")
	}

	// NOTE: We deploy another smart contract. This is a wrapped token contract
	// as a representation of the WEVMOS token.
	wevmosAddr, err := ts.factory.DeployContract(
		ts.keyring.GetPrivKey(erc20Deployer),
		evmtypes.EvmTxArgs{},
		testfactory.ContractDeploymentData{Contract: contracts.WEVMOSContract},
	)
	if err != nil {
		return &ConvertERC20CoinsTestSuite{}, errorsmod.Wrap(err, "failed to deploy WEVMOS contract")
	}

	err = ts.network.NextBlock()
	if err != nil {
		return &ConvertERC20CoinsTestSuite{}, errorsmod.Wrap(err, "failed to execute block")
	}

	// send some coins to the wevmos address to deposit them.
	_, err = ts.factory.ExecuteEthTx(
		ts.keyring.GetPrivKey(erc20Deployer),
		evmtypes.EvmTxArgs{
			To:     &wevmosAddr,
			Amount: sentWEVMOS.BigInt(),
			// FIXME: the gas simulation is not working correctly - otherwise results in out of gas
			GasLimit: 100_000,
		},
	)
	if err != nil {
		return &ConvertERC20CoinsTestSuite{}, errorsmod.Wrap(err, "failed to send WEVMOS to contract")
	}

	err = ts.network.NextBlock()
	if err != nil {
		return &ConvertERC20CoinsTestSuite{}, errorsmod.Wrap(err, "failed to execute block")
	}

	// Assign dynamic values to the test suite (=those that were not included in genesis).
	ts.erc20Contract = erc20Addr
	ts.nonNativeTokenPair = nonNativeTokenPair
	ts.wevmosContract = wevmosAddr

	return ts, nil
}
