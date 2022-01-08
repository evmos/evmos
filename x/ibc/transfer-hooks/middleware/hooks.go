package middleware

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/ibc/transfer-hooks/types"
)

var _ types.TransferHooks = &Middleware{}

func (m Middleware) AfterTransferAcked(
	ctx sdk.Context,
	sourcePort, sourceChannel string,
	token sdk.Coin,
	sender sdk.AccAddress,
	receiver string,
) error {
	if m.hooks != nil {
		return m.hooks.AfterTransferAcked(ctx, sourcePort, sourceChannel, token, sender, receiver)
	}
	return nil
}

func (m Middleware) AfterRecvTransfer(
	ctx sdk.Context,
	destPort, destChannel string,
	token sdk.Coin,
	receiver string,
) error {
	if m.hooks != nil {
		return m.hooks.AfterRecvTransfer(ctx, destPort, destChannel, token, receiver)
	}
	return nil
}
