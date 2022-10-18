package incentives

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v9/x/incentives/keeper"
	"github.com/evmos/evmos/v9/x/incentives/types"
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
			return sdkerrors.Wrapf(
				sdkerrors.ErrUnknownRequest,
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
