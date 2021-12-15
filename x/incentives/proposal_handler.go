package incentives

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/tharsis/evmos/x/incentives/keeper"
	"github.com/tharsis/evmos/x/incentives/types"
)

// NewIncentivesProposalHandler creates a governance handler to manage new
// proposal types.
func NewIncentivesProposalHandler(k *keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.RegisterIncentiveProposal:
			return handleRegisterIncentiveProposal(ctx, k, c)
		case *types.CancelIncentiveProposal:
			return handleCancelIncentiveProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
}

func handleRegisterIncentiveProposal(ctx sdk.Context, k *keeper.Keeper, p *types.RegisterIncentiveProposal) error {
	_, err := k.RegisterIncentive(ctx, p.Allocations, p.Contract, p.Epochs)
	if err != nil {
		return err
	}
	// TODO events
	return nil
}

func handleCancelIncentiveProposal(ctx sdk.Context, k *keeper.Keeper, p *types.CancelIncentiveProposal) error {
	err := k.CancelIncentive(ctx, p.Contract)
	if err != nil {
		return err
	}
	// TODO events

	return nil
}
