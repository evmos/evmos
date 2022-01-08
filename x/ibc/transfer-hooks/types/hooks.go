package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ModuleName = "transfer-hooks"

type TransferHooks interface {
	AfterTransferAcked(
		ctx sdk.Context,
		sourcePort,
		sourceChannel string,
		token sdk.Coin,
		sender sdk.AccAddress,
		receiver string,
	) error
	AfterRecvTransfer(
		ctx sdk.Context,
		destPort,
		destChannel string,
		token sdk.Coin,
		receiver string,
	) error
}

var _ TransferHooks = &MultiTransferHooks{}

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
) error {
	for i := range mths {
		err := mths[i].AfterTransferAcked(ctx, sourcePort, sourceChannel, token, sender, receiver)
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
) error {
	for i := range mths {
		err := mths[i].AfterRecvTransfer(ctx, destPort, destChannel, token, receiver)
		if err != nil {
			return err
		}
	}
	return nil
}
