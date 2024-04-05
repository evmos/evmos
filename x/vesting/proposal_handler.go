// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package vesting

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/evmos/evmos/v17/x/vesting/keeper"
	"github.com/evmos/evmos/v17/x/vesting/types"
)

// NewVestingProposalHandler creates a governance handler to manage new proposal types.
func NewVestingProposalHandler(k *keeper.Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch c := content.(type) {
		case *types.ClawbackProposal:
			return HandleClawbackProposal(ctx, k, c)

		default:
			return errorsmod.Wrapf(errortypes.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
}

// HandleClawbackProposal handles the proposal for clawback
// of a vesting account that has this functionality enabled.
func HandleClawbackProposal(
	ctx sdk.Context,
	k *keeper.Keeper,
	p *types.ClawbackProposal,
) error {
	governanceAddr := authtypes.NewModuleAddress(govtypes.ModuleName)

	msg := &types.MsgClawback{
		FunderAddress:  governanceAddr.String(),
		AccountAddress: p.Address,
		DestAddress:    p.DestinationAddress,
	}

	if _, err := k.Clawback(ctx, msg); err != nil {
		return err
	}

	return nil
}
