// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package post

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

var _ sdk.PostDecorator = &BurnDecorator{}

// BurnDecorator is the decorator that burns all the transaction fees from Cosmos transactions.
type BurnDecorator struct {
	feeCollectorName string
	bankKeeper       bankkeeper.Keeper
}

// NewBurnDecorator creates a new instance of the BurnDecorator.
func NewBurnDecorator(feeCollector string, bankKeeper bankkeeper.Keeper) sdk.PostDecorator {
	return &BurnDecorator{
		feeCollectorName: feeCollector,
		bankKeeper:       bankKeeper,
	}
}

// PostHandle burns all the transaction fees from Cosmos transactions. If an Ethereum transaction is present, this logic
// is skipped.
func (bd BurnDecorator) PostHandle(ctx sdk.Context, tx sdk.Tx, simulate, success bool, next sdk.PostHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrapf(errortypes.ErrInvalidType, "invalid transaction type %T, expected sdk.FeeTx", tx)
	}

	// skip logic if there is an Ethereum transaction
	for _, msg := range tx.GetMsgs() {
		if _, ok := msg.(*evmtypes.MsgEthereumTx); ok {
			return next(ctx, tx, simulate, success)
		}
	}

	fees := feeTx.GetFee()

	// safety check: ensure the fees are not empty and with positive amounts
	// before burning
	if len(fees) == 0 || !fees.IsAllPositive() {
		return next(ctx, tx, simulate, success)
	}

	// NOTE: since all Cosmos tx fees are pooled by the fee collector module account,
	// we burn them directly from it
	if err := bd.bankKeeper.BurnCoins(ctx, bd.feeCollectorName, fees); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate, success)
}
