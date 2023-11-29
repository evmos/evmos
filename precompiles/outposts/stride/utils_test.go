// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package stride_test

import (
	"fmt"

	"cosmossdk.io/math"
	commonnetwork "github.com/evmos/evmos/v15/testutil/integration/common/network"
	"github.com/evmos/evmos/v15/testutil/integration/ibc/coordinator"

	erc20types "github.com/evmos/evmos/v15/x/erc20/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	inflationtypes "github.com/evmos/evmos/v15/x/inflation/v1/types"
)

const (
	portID    = "transfer"
	channelID = "channel-0"
)

// registerStrideCoinERC20 registers stEvmos and Evmos coin as an ERC20 token
func (s *PrecompileTestSuite) registerStrideCoinERC20() {
	// Register EVMOS ERC20 equivalent
	ctx := s.network.GetContext()
	bondDenom := s.network.App.StakingKeeper.BondDenom(ctx)
	evmosMetadata, found := s.network.App.BankKeeper.GetDenomMetaData(ctx, bondDenom)
	s.Require().True(found, "expected evmos denom metadata")

	coin := sdk.NewCoin(evmosMetadata.Base, math.NewInt(2e18))
	err := s.network.App.BankKeeper.MintCoins(ctx, inflationtypes.ModuleName, sdk.NewCoins(coin))
	s.Require().NoError(err)

	// Register some Token Pairs
	_, err = s.network.App.Erc20Keeper.RegisterCoin(ctx, evmosMetadata)
	s.Require().NoError(err)

	// Register stEvmos Token Pair
	denomTrace := transfertypes.DenomTrace{
		Path:      fmt.Sprintf("%s/%s", portID, channelID),
		BaseDenom: "st" + bondDenom,
	}
	s.network.App.TransferKeeper.SetDenomTrace(ctx, denomTrace)
	stEvmosMetadata := banktypes.Metadata{
		Description: "The native token of Evmos",
		Base:        denomTrace.IBCDenom(),
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denomTrace.IBCDenom(),
				Exponent: 0,
				Aliases:  []string{denomTrace.BaseDenom},
			},
			{
				Denom:    denomTrace.BaseDenom,
				Exponent: 18,
			},
		},
		Name:    "stEvmos",
		Symbol:  "STEVMOS",
		Display: denomTrace.BaseDenom,
	}

	stEvmos := sdk.NewCoin(stEvmosMetadata.Base, math.NewInt(9e18))
	err = s.network.App.BankKeeper.MintCoins(ctx, inflationtypes.ModuleName, sdk.NewCoins(stEvmos))
	s.Require().NoError(err)
	err = s.network.App.BankKeeper.SendCoinsFromModuleToAccount(ctx, inflationtypes.ModuleName, s.keyring.GetAccAddr(0), sdk.NewCoins(stEvmos))
	s.Require().NoError(err)

	// Register some Token Pairs
	_, err = s.network.App.Erc20Keeper.RegisterCoin(ctx, stEvmosMetadata)
	s.Require().NoError(err)

	convertCoin := erc20types.NewMsgConvertCoin(
		stEvmos,
		s.keyring.GetAddr(0),
		s.keyring.GetAccAddr(0),
	)

	_, err = s.network.App.Erc20Keeper.ConvertCoin(ctx, convertCoin)
	s.Require().NoError(err)
}

// setupIBCCoordinator sets up the IBC coordinator
func (s *PrecompileTestSuite) setupIBCCoordinator() {
	ibcSender, ibcSenderPrivKey := s.keyring.GetAccAddr(0), s.keyring.GetPrivKey(0)
	ibcAcc, err := s.grpcHandler.GetAccount(ibcSender.String())
	s.Require().NoError(err)

	IBCCoordinator := coordinator.NewIntegrationCoordinator(
		s.T(),
		[]commonnetwork.Network{s.network},
	)

	IBCCoordinator.SetDefaultSignerForChain(s.network.GetChainID(), ibcSenderPrivKey, ibcAcc)
	IBCCoordinator.Setup(s.network.GetChainID(), IBCCoordinator.GetDummyChainsIds()[0])

	err = IBCCoordinator.CommitAll()
	s.Require().NoError(err)
}
