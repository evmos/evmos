// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package erc20

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v10/x/erc20/keeper"
	"github.com/evmos/evmos/v10/x/erc20/types"
)

// NewErc20ProposalHandler creates a governance handler to manage new proposal types.
func NewErc20ProposalHandler(k *keeper.Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		// Check if the conversion is globally enabled
		if !k.IsERC20Enabled(ctx) {
			return errorsmod.Wrap(
				types.ErrERC20Disabled, "registration is currently disabled by governance",
			)
		}

		switch c := content.(type) {
		case *types.RegisterCoinProposal:
			return handleRegisterCoinProposal(ctx, k, c)
		case *types.RegisterERC20Proposal:
			return handleRegisterERC20Proposal(ctx, k, c)
		case *types.ToggleTokenConversionProposal:
			return handleToggleConversionProposal(ctx, k, c)

		default:
			return errorsmod.Wrapf(errortypes.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
}

// handleRegisterCoinProposal handles the registration proposal for multiple
// native Cosmos coins
func handleRegisterCoinProposal(
	ctx sdk.Context,
	k *keeper.Keeper,
	p *types.RegisterCoinProposal,
) error {
	for _, metadata := range p.Metadata {
		pair, err := k.RegisterCoin(ctx, metadata)
		if err != nil {
			return err
		}

		err = ctx.EventManager().EmitTypedEvent(&types.EventRegisterPair{
			Denom:        pair.Denom,
			Erc20Address: pair.Erc20Address,
		})

		if err != nil {
			k.Logger(ctx).Error(err.Error())
		}
	}

	return nil
}

// handleRegisterERC20Proposal handles the registration proposal for multiple
// ERC20 tokens
func handleRegisterERC20Proposal(
	ctx sdk.Context,
	k *keeper.Keeper,
	p *types.RegisterERC20Proposal,
) error {
	for _, address := range p.Erc20Addresses {
		pair, err := k.RegisterERC20(ctx, common.HexToAddress(address))
		if err != nil {
			return err
		}

		err = ctx.EventManager().EmitTypedEvent(&types.EventRegisterPair{
			Denom:        pair.Denom,
			Erc20Address: pair.Erc20Address,
		})

		if err != nil {
			k.Logger(ctx).Error(err.Error())
		}
	}

	return nil
}

// handleToggleConversionProposal handles the toggle proposal for a token pair
func handleToggleConversionProposal(
	ctx sdk.Context,
	k *keeper.Keeper,
	p *types.ToggleTokenConversionProposal,
) error {
	pair, err := k.ToggleConversion(ctx, p.Token)
	if err != nil {
		return err
	}

	err = ctx.EventManager().EmitTypedEvent(&types.EventToggleTokenConversion{
		Denom:        pair.Denom,
		Erc20Address: pair.Erc20Address,
	})

	if err != nil {
		k.Logger(ctx).Error(err.Error())
	}

	return nil
}
