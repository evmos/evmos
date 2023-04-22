// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package incentives

import (
	"strconv"

	errorsmod "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v12/x/incentives/keeper"
	"github.com/evmos/evmos/v12/x/incentives/types"
)

// NewIncentivesProposalHandler creates a governance handler to manage new
// proposal types.
func NewIncentivesProposalHandler(k *keeper.Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch c := content.(type) {
		case *types.RegisterIncentiveProposal:
			return handleRegisterIncentiveProposal(ctx, k, c)
		case *types.CancelIncentiveProposal:
			return handleCancelIncentiveProposal(ctx, k, c)
		default:
			return errorsmod.Wrapf(
				errortypes.ErrUnknownRequest,
				"unrecognized %s proposal content type: %T", types.ModuleName, c,
			)
		}
	}
}

func handleRegisterIncentiveProposal(ctx sdk.Context, k *keeper.Keeper, p *types.RegisterIncentiveProposal) error {
	in, err := k.RegisterIncentive(ctx, common.HexToAddress(p.Contract), p.Allocations, p.Epochs)
	if err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRegisterIncentive,
			sdk.NewAttribute(types.AttributeKeyContract, in.Contract),
			sdk.NewAttribute(
				types.AttributeKeyEpochs,
				strconv.FormatUint(uint64(in.Epochs), 10),
			),
		),
	)
	return nil
}

func handleCancelIncentiveProposal(ctx sdk.Context, k *keeper.Keeper, p *types.CancelIncentiveProposal) error {
	err := k.CancelIncentive(ctx, common.HexToAddress(p.Contract))
	if err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCancelIncentive,
			sdk.NewAttribute(types.AttributeKeyContract, p.Contract),
		),
	)
	return nil
}
