// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v16/x/erc20/keeper"
	"github.com/evmos/evmos/v16/x/erc20/types"
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
		case *types.RegisterERC20Proposal:
			return handleRegisterERC20Proposal(ctx, k, c)
		case *types.ToggleTokenConversionProposal:
			return handleToggleConversionProposal(ctx, k, c)

		default:
			return errorsmod.Wrapf(errortypes.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
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

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRegisterERC20,
				sdk.NewAttribute(types.AttributeKeyCosmosCoin, pair.Denom),
				sdk.NewAttribute(types.AttributeKeyERC20Token, pair.Erc20Address),
			),
		)
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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeToggleTokenConversion,
			sdk.NewAttribute(types.AttributeKeyCosmosCoin, pair.Denom),
			sdk.NewAttribute(types.AttributeKeyERC20Token, pair.Erc20Address),
		),
	)

	return nil
}
