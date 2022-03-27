package fees

// import (
// 	"strconv"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
// 	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
// 	"github.com/ethereum/go-ethereum/common"

// 	"github.com/tharsis/evmos/v3/x/fees/keeper"
// 	"github.com/tharsis/evmos/v3/x/fees/types"
// )

// // NewFeesProposalHandler creates a governance handler to manage new
// // proposal types.
// func NewContractProposalHandler(k *keeper.Keeper) govtypes.Handler {
// 	return func(ctx sdk.Context, content govtypes.Content) error {
// 		switch c := content.(type) {
// 		case *types.RegisterContractProposal:
// 			return handleRegisterIncentiveProposal(ctx, k, c)
// 		case *types.CancelContractProposal:
// 			return handleCancelIncentiveProposal(ctx, k, c)
// 		default:
// 			return sdkerrors.Wrapf(
// 				sdkerrors.ErrUnknownRequest,
// 				"unrecognized %s proposal content type: %T", types.ModuleName, c,
// 			)
// 		}
// 	}
// }

// func handleRegisterContractProposal(ctx sdk.Context, k *keeper.Keeper, p *types.RegisterContractProposal) error {
// 	in, err := k.RegisterContract(ctx, common.HexToAddress(p.Contract), p.Allocations, p.Epochs)
// 	if err != nil {
// 		return err
// 	}
// 	ctx.EventManager().EmitEvent(
// 		sdk.NewEvent(
// 			types.EventTypeRegisterContract,
// 			sdk.NewAttribute(types.AttributeKeyContract, in.Contract),
// 			sdk.NewAttribute(
// 				types.AttributeKeyEpochs,
// 				strconv.FormatUint(uint64(in.Epochs), 10),
// 			),
// 		),
// 	)
// 	return nil
// }

// func handleCancelContractProposal(ctx sdk.Context, k *keeper.Keeper, p *types.CancelContractProposal) error {
// 	err := k.CancelContract(ctx, common.HexToAddress(p.Contract))
// 	if err != nil {
// 		return err
// 	}
// 	ctx.EventManager().EmitEvent(
// 		sdk.NewEvent(
// 			types.EventTypeCancelContract,
// 			sdk.NewAttribute(types.AttributeKeyContract, p.Contract),
// 		),
// 	)
// 	return nil
// }
