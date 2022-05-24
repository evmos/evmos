package unigov

import (
	"github.com/Canto-Network/canto/v3/x/unigov/keeper"
	"github.com/Canto-Network/canto/v3/x/unigov/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"fmt"
)

//Return governance handler to process Compound Proposal
func NewUniGovProposalHandler(k *keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.LendingMarketProposal:
			return handleLendingMarketProposal(ctx, k, c)

		default:
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
}

func handleLendingMarketProposal(ctx sdk.Context, k *keeper.Keeper, p *types.LendingMarketProposal) error {
	_, err := k.AppendLendingMarketProposal(ctx, p) //Defined analogous to (erc20)k.RegisterCoin
	if err != nil {
		return err
	}

	fmt.Println("Proposal was here" + p.String() + "\n\n\n\n\n")
	
	// ctx.EventManager().EmitEvent(
	// 	sdk.NewEvent(
	// 		types.EventLendingMarketProposal,
			
	// 	)
	// )

	return nil
}
