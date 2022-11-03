package keeper

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"

	erc20types "github.com/evmos/evmos/v9/x/erc20/types"
)

var _ types.MsgServer = Keeper{}

// Transfer defines a gRPC msg server method for MsgTransfer.
// This implementation overrides the default ICS20 transfer's by converting
// the ERC20 tokens to their Cosmos representation if the token pair has been
// registered through governance.
func (k Keeper) Transfer(goCtx context.Context, msg *types.MsgTransfer) (*types.MsgTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if IBC transfer denom is a valid Ethereum contract address
	if !common.IsHexAddress(msg.Token.Denom) {
		// no-op: continue with regular transfer
		return k.Keeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	}

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		// NOTE: shouldn't happen as the receiving address has already
		// been validated on ICS20 transfer logic
		return nil, sdkerrors.Wrap(err, "invalid sender")
	}

	// Return acknowledgement and continue with the next layer of the IBC middleware
	// stack if if:
	// - ERC20s are disabled
	// - The ERC20 contract is not registered as Cosmos coin
	erc20Params := k.erc20Keeper.GetParams(ctx)
	if !erc20Params.EnableErc20 {
		// no-op: continue with regular transfer
		return k.Keeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	}

	pairID := k.erc20Keeper.GetTokenPairID(ctx, msg.Token.Denom)
	if len(pairID) == 0 {
		// no-op: token is not registered so we can proceed with regular transfer
		return k.Keeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	}

	// NOTE: no need to check if the token pair is found
	tokenPair, _ := k.erc20Keeper.GetTokenPair(ctx, pairID)

	// if the user has enough balance of the Cosmos representation, then we don't need to Convert
	if k.bankKeeper.HasBalance(ctx, sender, sdk.Coin{Denom: tokenPair.Denom, Amount: msg.Token.Amount}) {
		// update the denom and proceed with regular transfer
		msg.Token.Denom = tokenPair.Denom
		return k.Keeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	}

	contractAddr := common.HexToAddress(msg.Token.Denom)

	msgConvertERC20 := erc20types.NewMsgConvertERC20(
		msg.Token.Amount,
		sender,
		contractAddr,
		common.BytesToAddress(sender.Bytes()),
	)

	if err := msgConvertERC20.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to validate MsgConvertERC20")
	}

	// Use MsgConvertERC20 to convert the ERC20 to a Cosmos IBC Coin
	if _, err := k.erc20Keeper.ConvertERC20(sdk.WrapSDKContext(ctx), msgConvertERC20); err != nil {
		return nil, err
	}
        msg.Token.Denom = tokenPair.Denom
	return k.Keeper.Transfer(sdk.WrapSDKContext(ctx), msg)
}
