// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// This file contains all utility function that require the access to the unit
// test network and should only be used in unit tests.
package utils

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/evmos/evmos/v15/utils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	erc20types "github.com/evmos/evmos/v15/x/erc20/types"
	inflationtypes "github.com/evmos/evmos/v15/x/inflation/v1/types"
)

const (
	TokenToMint = 1e18
)

// RegisterEvmosERC20Coins uses the UnitNetwork to register the evmos token as an
// ERC20 token. The function performs all the required steps for the registration
// like registering the denom trace in the transfer keeper and minting the token
// with the bank. Returns the TokenPair or an error.
func RegisterEvmosERC20Coins(
	network network.UnitTestNetwork,
	tokenReceiver sdk.AccAddress,
) (erc20types.TokenPair, error) {
	bondDenom := network.App.StakingKeeper.BondDenom(network.GetContext())

	coin := sdk.NewCoin(utils.BaseDenom, math.NewInt(TokenToMint))
	err := network.App.BankKeeper.MintCoins(
		network.GetContext(),
		inflationtypes.ModuleName,
		sdk.NewCoins(coin),
	)
	if err != nil {
		return erc20types.TokenPair{}, err
	}
	err = network.App.BankKeeper.SendCoinsFromModuleToAccount(
		network.GetContext(),
		inflationtypes.ModuleName,
		tokenReceiver,
		sdk.NewCoins(coin),
	)
	if err != nil {
		return erc20types.TokenPair{}, err
	}

	evmosMetadata, found := network.App.BankKeeper.GetDenomMetaData(network.GetContext(), utils.BaseDenom)
	if !found {
		return erc20types.TokenPair{}, fmt.Errorf("expected evmos denom metadata")
	}

	_, err = network.App.Erc20Keeper.RegisterCoin(network.GetContext(), evmosMetadata)
	if err != nil {
		return erc20types.TokenPair{}, err
	}

	evmosDenomID := network.App.Erc20Keeper.GetDenomMap(network.GetContext(), bondDenom)
	tokenPair, ok := network.App.Erc20Keeper.GetTokenPair(network.GetContext(), evmosDenomID)
	if !ok {
		return erc20types.TokenPair{}, fmt.Errorf("expected evmos erc20 token pair")
	}

	return tokenPair, nil
}

// RegisterIBCERC20Coins uses the UnitNetwork to register the denomTrace as an
// ERC20 token. The function performs all the required steps for the registration
// like registering the denom trace in the transfer keeper and minting the token
// with the bank. Returns the TokenPair or an error.
func RegisterIBCERC20Coins(
	network network.UnitTestNetwork,
	tokenReceiver sdk.AccAddress,
	denomTrace transfertypes.DenomTrace,
) (erc20types.TokenPair, error) {
	ibcDenom := denomTrace.IBCDenom()
	network.App.TransferKeeper.SetDenomTrace(network.GetContext(), denomTrace)
	ibcMetadata := banktypes.Metadata{
		Name:        "Generic IBC name",
		Symbol:      "IBC",
		Description: "Generic IBC token description",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    ibcDenom,
				Exponent: 0,
				Aliases:  []string{ibcDenom},
			},
			{
				Denom:    ibcDenom,
				Exponent: 18,
			},
		},
		Display: ibcDenom,
		Base:    ibcDenom,
	}

	coin := sdk.NewCoin(ibcMetadata.Base, math.NewInt(TokenToMint))
	err := network.App.BankKeeper.MintCoins(
		network.GetContext(),
		inflationtypes.ModuleName,
		sdk.NewCoins(coin),
	)
	if err != nil {
		return erc20types.TokenPair{}, err
	}

	err = network.App.BankKeeper.SendCoinsFromModuleToAccount(
		network.GetContext(),
		inflationtypes.ModuleName,
		tokenReceiver,
		sdk.NewCoins(coin),
	)
	if err != nil {
		return erc20types.TokenPair{}, err
	}

	_, err = network.App.Erc20Keeper.RegisterCoin(network.GetContext(), ibcMetadata)
	if err != nil {
		return erc20types.TokenPair{}, err
	}

	ibcDenomID := network.App.Erc20Keeper.GetDenomMap(
		network.GetContext(),
		denomTrace.IBCDenom(),
	)
	tokenPair, ok := network.App.Erc20Keeper.GetTokenPair(network.GetContext(), ibcDenomID)
	if !ok {
		return erc20types.TokenPair{}, fmt.Errorf("expected %s erc20 token pair", ibcDenom)
	}

	return tokenPair, nil
}
