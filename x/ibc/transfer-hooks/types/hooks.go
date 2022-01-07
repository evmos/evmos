package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TransferHooks interface {
	AfterTransferAcked(
		ctx sdk.Context,
		sourcePort,
		sourceChannel string,
		token sdk.Coin,
		sender sdk.AccAddress,
		receiver string,
		isSource bool,
	) error
	AfterRecvTransfer(
		ctx sdk.Context,
		destPort,
		destChannel string,
		token sdk.Coin,
		receiver string,
		isSource bool,
	) error
}

type MultiTransferHooks []TransferHooks

func NewMultiTransferHooks(hooks ...TransferHooks) MultiTransferHooks {
	return hooks
}

func (mths MultiTransferHooks) AfterTransferAcked(
	ctx sdk.Context,
	sourcePort, sourceChannel string,
	token sdk.Coin,
	sender sdk.AccAddress,
	receiver string,
	isSource bool,
) error {
	for i := range mths {
		err := mths[i].AfterTransferAcked(ctx, sourcePort, sourceChannel, token, sender, receiver, isSource)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mths MultiTransferHooks) AfterRecvTransfer(
	ctx sdk.Context,
	destPort, destChannel string,
	token sdk.Coin,
	receiver string,
	isSource bool,
) error {
	for i := range mths {
		err := mths[i].AfterRecvTransfer(ctx, destPort, destChannel, token, receiver, isSource)
		if err != nil {
			return err
		}
	}
	return nil
}
