package unigov

import(
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Canto-Network/canto/v3/unigov/keeper"
	"github.com/Canto-Network/canto/x/unigov/types"
)

//Return governance handler to process Compound Proposal
func NewUniGovProposalHandler(k *keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c:= govtypes.Content.(type) {
		case *types.CompoundProposal:
			return handleLendingMarketProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
}

func handleLendingMarketProposal(ctx sdk.Context, k *keeper.Keeper, p *types.LendingMarketProposal) error {
	addr, err := k.AppendLendingMarketProposal(ctx, p) //Defined analogous to (erc20)k.RegisterCoin 
	if err != nil {
		return err
	}
	//ctx.EventManager().EmitEvent(sdk.NewEvent(args))
	return nil
}
