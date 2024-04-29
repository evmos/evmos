// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package strv2_test

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/contracts"
	commonfactory "github.com/evmos/evmos/v18/testutil/integration/common/factory"
	testfactory "github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v18/utils"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
	"github.com/pkg/errors"
)

func CreateTestSuite(chainID string) (*STRv2TrackingSuite, error) {
	keyring := testkeyring.New(3)

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
		testnetwork.WithChainID(chainID),
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
		return &STRv2TrackingSuite{}, errorsmod.Wrap(err, "failed to register native IBC coin")
	}
	nativeCoinERC20Addr := common.HexToAddress(ibcNativeTokenPair.Erc20Address)

	if err := network.NextBlock(); err != nil {
		return &STRv2TrackingSuite{}, errorsmod.Wrap(err, "failed to advance block")
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
		return &STRv2TrackingSuite{}, errorsmod.Wrap(err, "failed to convert a balance")
	}

	if err := network.NextBlock(); err != nil {
		return &STRv2TrackingSuite{}, errorsmod.Wrap(err, "failed to advance block")
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
		return &STRv2TrackingSuite{}, errorsmod.Wrap(err, "failed to deploy ERC-20 contract")
	}

	if err := network.NextBlock(); err != nil {
		return &STRv2TrackingSuite{}, errorsmod.Wrap(err, "failed to advance block")
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
		return &STRv2TrackingSuite{}, errorsmod.Wrap(err, "failed to mint ERC-20 tokens")
	}

	_, err = network.App.Erc20Keeper.RegisterERC20(network.GetContext(), registeredERC20Addr)
	if err != nil {
		return &STRv2TrackingSuite{}, errorsmod.Wrap(err, "failed to register token pair")
	}

	if err := network.NextBlock(); err != nil {
		return &STRv2TrackingSuite{}, errorsmod.Wrap(err, "failed to advance block")
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
		return &STRv2TrackingSuite{}, errorsmod.Wrap(err, "failed to deploy ERC-20 contract")
	}

	if err := network.NextBlock(); err != nil {
		return &STRv2TrackingSuite{}, errorsmod.Wrap(err, "failed to advance block")
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
		return &STRv2TrackingSuite{}, errorsmod.Wrap(err, "failed to mint ERC-20 tokens")
	}

	network.App.Erc20Keeper.DeleteSTRv2Address(network.GetContext(), keyring.GetAccAddr(0))

	// NOTE: this is necessary to enable e.g. erc20Keeper.BalanceOf(...) to work
	// correctly internally.
	// Removing it will break a bunch of tests giving errors like: "failed to retrieve balance"
	if err = network.NextBlock(); err != nil {
		return nil, errors.Wrap(err, "failed to advance block")
	}

	return &STRv2TrackingSuite{
		keyring:               keyring,
		network:               network,
		handler:               handler,
		factory:               factory,
		nativeCoinERC20Addr:   nativeCoinERC20Addr,
		registeredERC20Addr:   registeredERC20Addr,
		unregisteredERC20Addr: unregisteredERC20Addr,
	}, nil
}
