package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TransferHooks interface {
	AfterSendTransferAcked(
		ctx sdk.Context,
		sourcePort,
		sourceChannel string,
		token sdk.Coin,
		sender sdk.AccAddress,
		receiver string,
		isSource bool,
	)
	AfterRecvTransfer(
		ctx sdk.Context,
		destPort,
		destChannel string,
		token sdk.Coin,
		receiver string,
		isSource bool,
	)
}

type MultiTransferHooks []TransferHooks

func NewMultiTransferHooks(hooks ...TransferHooks) MultiTransferHooks {
	return hooks
}

func (mths MultiTransferHooks) AfterSendTransferAcked(
	ctx sdk.Context,
	sourcePort, sourceChannel string,
	token sdk.Coin,
	sender sdk.AccAddress,
	receiver string,
	isSource bool) {
	for i := range mths {
		mths[i].AfterSendTransferAcked(ctx, sourcePort, sourceChannel, token, sender, receiver, isSource)
	}
}

func (mths MultiTransferHooks) AfterRecvTransfer(
	ctx sdk.Context,
	destPort, destChannel string,
	token sdk.Coin,
	receiver string,
	isSource bool) {
	for i := range mths {
		mths[i].AfterRecvTransfer(ctx, destPort, destChannel, token, receiver, isSource)
	}
}
