package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/ibc/transfer-hooks/types"
)

var _ types.TransferHooks = Keeper{}

func (k Keeper) AfterTransferAcked(
	ctx sdk.Context,
	sourcePort, sourceChannel string,
	token sdk.Coin,
	sender sdk.AccAddress,
	receiver string,
	isSource bool,
) error {
	if k.hooks != nil {
		return k.hooks.AfterTransferAcked(ctx, sourcePort, sourceChannel, token, sender, receiver, isSource)
	}
	return nil
}

func (k Keeper) AfterRecvTransfer(
	ctx sdk.Context,
	destPort, destChannel string,
	token sdk.Coin,
	receiver string,
	isSource bool,
) error {
	if k.hooks != nil {
		return k.hooks.AfterRecvTransfer(ctx, destPort, destChannel, token, receiver, isSource)
	}
	return nil
}
