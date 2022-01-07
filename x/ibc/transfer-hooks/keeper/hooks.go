package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/ibc/transfer-hooks/types"
)

var _ types.TransferHooks = Keeper{}

func (k Keeper) AfterSendTransferAcked(
	ctx sdk.Context,
	sourcePort, sourceChannel string,
	token sdk.Coin,
	sender sdk.AccAddress,
	receiver string,
	isSource bool) {
	if k.hooks != nil {
		k.hooks.AfterSendTransferAcked(ctx, sourcePort, sourceChannel, token, sender, receiver, isSource)
	}
}

func (k Keeper) AfterRecvTransfer(
	ctx sdk.Context,
	destPort, destChannel string,
	token sdk.Coin,
	receiver string,
	isSource bool) {
	if k.hooks != nil {
		k.hooks.AfterRecvTransfer(ctx, destPort, destChannel, token, receiver, isSource)
	}
}
