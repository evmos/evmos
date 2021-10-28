package intrarelayer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/tharsis/evmos/x/intrarelayer/keeper"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

// NewIntrarelayerProposalHandler creates a governance handler to manage new proposal types.
// It enables RegisterTokenPairProposal to propose a registration of token mapping
func NewIntrarelayerProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.RegisterTokenPairProposal:
			return handleRegisterTokenPairProposal(ctx, k, c)
		case *types.ToggleTokenRelayProposal:
			return handleToggleRelayProposal(ctx, k, c)
		case *types.UpdateTokenPairERC20Proposal:
			return handleUpdateTokenPairERC20Proposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
}

func handleRegisterTokenPairProposal(ctx sdk.Context, k keeper.Keeper, p *types.RegisterTokenPairProposal) error {
	if err := k.RegisterTokenPair(ctx, p.TokenPair); err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRegisterTokenPair,
			sdk.NewAttribute(types.AttributeKeyCosmosCoin, p.TokenPair.Denom),
			sdk.NewAttribute(types.AttributeKeyERC20Token, p.TokenPair.Erc20Address),
		),
	)

	return nil
}

func handleToggleRelayProposal(ctx sdk.Context, k keeper.Keeper, p *types.ToggleTokenRelayProposal) error {
	pair, err := k.ToggleRelay(ctx, p.Token)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeToggleTokenRelay,
			sdk.NewAttribute(types.AttributeKeyCosmosCoin, pair.Denom),
			sdk.NewAttribute(types.AttributeKeyERC20Token, pair.Erc20Address),
		),
	)

	return nil
}

func handleUpdateTokenPairERC20Proposal(ctx sdk.Context, k keeper.Keeper, p *types.UpdateTokenPairERC20Proposal) error {
	pair, err := k.UpdateTokenPairERC20(ctx, p.GetERC20Address(), p.GetNewERC20Address())
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateTokenPairERC20,
			sdk.NewAttribute(types.AttributeKeyCosmosCoin, pair.Denom),
			sdk.NewAttribute(types.AttributeKeyERC20Token, pair.Erc20Address),
		),
	)

	return nil
}
