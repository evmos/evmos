package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/ibc/evm/types"
)

var _ types.MsgServer = &Keeper{}

func (k Keeper) CallEVM(goCtx context.Context, msg *types.MsgCallEVM) (*types.MsgCallEVMResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Check if EVM send param is enabled
	if !k.GetSendEvmTxEnabled(ctx) {
		return nil, types.ErrReceiveDisabled
	}

	// Convert our Bech32 to AccAddress
	accAddress, err := sdk.AccAddressFromBech32(msg.Packet.Sender)
	if err != nil {
		return nil, err
	}

	// Convert string amount to integer
	amount, ok := sdk.NewIntFromString(msg.Amount)
	if !ok {
		return nil, fmt.Errorf("invalid amount %s", msg.Amount)
	}

	// Check if the account has the amount to process this transaction
	coin := sdk.Coin{Denom: msg.Denom, Amount: amount}
	coinBalance := k.bankKeeper.GetBalance(ctx, accAddress, coin.Denom)
	if coinBalance.Amount.LTE(sdk.NewInt(0)) {
		return nil, fmt.Errorf("not enough amount for gas - %s", coinBalance.Amount)
	}

	k.sendEvmTx(ctx, msg.SourcePort, msg.SourceChannel, coin, accAddress, msg.TimeoutHeight, msg.TimeoutTimestamp, msg.Packet.Data.Value, nil)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTransfer,
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Packet.Sender),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	})

	res := &types.MsgCallEVMResponse{}
	return res, nil

}
