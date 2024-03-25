// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package strv2_test

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/contracts"
	commonfactory "github.com/evmos/evmos/v16/testutil/integration/common/factory"
	testfactory "github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v16/utils"
	"github.com/evmos/evmos/v16/x/erc20"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	"github.com/evmos/evmos/v16/x/evm"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

type GenesisSetup struct {
	keyring      testkeyring.Keyring
	genesisState *testnetwork.CustomGenesisState

	nativeCoinERC20Addr   common.Address
	registeredERC20Addr   common.Address
	unregisteredERC20Addr common.Address
}

// CreateGenesisSetup sets up a genesis state to base the tracking of interactions with
// ERC-20 token pairs on.
//
// NOTE: it sets up another test network and then exports the genesis state to be used in the actual tests,
// which is similar to how real upgrades work, where a previous network state is used to start with the
// new version.
func CreateGenesisSetup(keyring testkeyring.Keyring) (GenesisSetup, error) {
	genesisBalances := []banktypes.Balance{
		{
			Address: keyring.GetAccAddr(0).String(),
			Coins: sdk.NewCoins(
				sdk.Coin{Denom: utils.BaseDenom, Amount: testnetwork.PrefundedAccountInitialBalance},
				sdk.Coin{Denom: nativeIBCCoinDenom, Amount: testnetwork.PrefundedAccountInitialBalance},
			),
		},
		{
			Address: keyring.GetAccAddr(1).String(),
			Coins: sdk.NewCoins(
				sdk.Coin{Denom: utils.BaseDenom, Amount: testnetwork.PrefundedAccountInitialBalance},
				sdk.Coin{Denom: nativeIBCCoinDenom, Amount: testnetwork.PrefundedAccountInitialBalance.QuoRaw(2)},
			),
		},
	}

	network := testnetwork.NewUnitTestNetwork(
		testnetwork.WithBalances(genesisBalances...),
	)
	handler := grpc.NewIntegrationHandler(network)
	factory := testfactory.New(network, handler)

	deployer := keyring.GetKey(deployerIdx)

	// ------------------------------------------------------------------
	// Register the native IBC coin
	ibcCoinMetaData := banktypes.Metadata{
		Description: "The native IBC coin",
		Base:        nativeIBCCoinDenom,
		Display:     nativeIBCCoinDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: nativeIBCCoinDenom, Exponent: 0},
			{Denom: "u" + nativeIBCCoinDenom, Exponent: 6},
		},
	}

	ibcNativeTokenPair, err := network.App.Erc20Keeper.RegisterCoin(network.GetContext(), ibcCoinMetaData)
	if err != nil {
		return GenesisSetup{}, errorsmod.Wrap(err, "failed to register native IBC coin")
	}
	nativeCoinERC20Addr := common.HexToAddress(ibcNativeTokenPair.Erc20Address)

	if err := network.NextBlock(); err != nil {
		return GenesisSetup{}, errorsmod.Wrap(err, "failed to advance block")
	}

	// Convert a balance for the deployer
	msgConvertCoin := erc20types.MsgConvertCoin{
		Sender:   deployer.AccAddr.String(),
		Receiver: deployer.Addr.String(),
		Coin:     sdk.NewCoin(nativeIBCCoinDenom, convertAmount),
	}
	_, err = factory.ExecuteCosmosTx(deployer.Priv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&msgConvertCoin},
	})
	if err != nil {
		return GenesisSetup{}, errorsmod.Wrap(err, "failed to convert a balance")
	}

	if err := network.NextBlock(); err != nil {
		return GenesisSetup{}, errorsmod.Wrap(err, "failed to advance block")
	}

	// ------------------------------------------------------------------
	// Register an ERC-20 token pair
	registeredERC20Addr, err := factory.DeployContract(
		deployer.Priv,
		evmtypes.EvmTxArgs{},
		testfactory.ContractDeploymentData{
			Contract:        contracts.ERC20MinterBurnerDecimalsContract,
			ConstructorArgs: []interface{}{"TestToken", "TTK", uint8(18)},
		},
	)
	if err != nil {
		return GenesisSetup{}, errorsmod.Wrap(err, "failed to deploy ERC-20 contract")
	}

	if err := network.NextBlock(); err != nil {
		return GenesisSetup{}, errorsmod.Wrap(err, "failed to advance block")
	}

	// Mint some tokens for the deployer
	_, err = factory.ExecuteContractCall(
		deployer.Priv,
		evmtypes.EvmTxArgs{
			To: &registeredERC20Addr,
		},
		testfactory.CallArgs{
			ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
			MethodName:  "mint",
			Args: []interface{}{
				deployer.Addr,
				mintAmount,
			},
		},
	)
	if err != nil {
		return GenesisSetup{}, errorsmod.Wrap(err, "failed to mint ERC-20 tokens")
	}

	_, err = network.App.Erc20Keeper.RegisterERC20(network.GetContext(), registeredERC20Addr)
	if err != nil {
		return GenesisSetup{}, errorsmod.Wrap(err, "failed to register token pair")
	}

	if err := network.NextBlock(); err != nil {
		return GenesisSetup{}, errorsmod.Wrap(err, "failed to advance block")
	}

	// ------------------------------------------------------------------
	// Deploy an unregistered ERC-20 contract
	unregisteredERC20Addr, err := factory.DeployContract(
		deployer.Priv,
		evmtypes.EvmTxArgs{},
		testfactory.ContractDeploymentData{
			Contract:        contracts.ERC20MinterBurnerDecimalsContract,
			ConstructorArgs: []interface{}{"UnregisteredToken", "URT", uint8(18)},
		},
	)
	if err != nil {
		return GenesisSetup{}, errorsmod.Wrap(err, "failed to deploy ERC-20 contract")
	}

	if err := network.NextBlock(); err != nil {
		return GenesisSetup{}, errorsmod.Wrap(err, "failed to advance block")
	}

	// Mint some tokens for the deployer
	_, err = factory.ExecuteContractCall(
		deployer.Priv,
		evmtypes.EvmTxArgs{
			To: &unregisteredERC20Addr,
		},
		testfactory.CallArgs{
			ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
			MethodName:  "mint",
			Args: []interface{}{
				deployer.Addr,
				mintAmount,
			},
		},
	)
	if err != nil {
		return GenesisSetup{}, errorsmod.Wrap(err, "failed to mint ERC-20 tokens")
	}

	// ------------------------------------------------------------------
	// Export genesis states
	ag := network.App.AccountKeeper.ExportGenesis(network.GetContext())
	bg := network.App.BankKeeper.ExportGenesis(network.GetContext())
	dg := network.App.DistrKeeper.ExportGenesis(network.GetContext())
	eg := evm.ExportGenesis(network.GetContext(), network.App.EvmKeeper, network.App.AccountKeeper)
	ercg := erc20.ExportGenesis(network.GetContext(), network.App.Erc20Keeper)

	return GenesisSetup{
		keyring: keyring,
		genesisState: &testnetwork.CustomGenesisState{
			authtypes.ModuleName:         ag,
			banktypes.ModuleName:         bg,
			distributiontypes.ModuleName: dg,
			evmtypes.ModuleName:          eg,
			erc20types.ModuleName:        ercg,
		},
		nativeCoinERC20Addr:   nativeCoinERC20Addr,
		registeredERC20Addr:   registeredERC20Addr,
		unregisteredERC20Addr: unregisteredERC20Addr,
	}, nil
}
