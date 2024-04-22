// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evm

import (
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	evmtypes "github.com/evmos/evmos/v17/x/evm/types"
)

// EthEmitEventDecorator emit events in ante handler in case of tx execution failed (out of block gas limit).
type EthEmitEventDecorator struct {
	evmKeeper EVMKeeper
}

// NewEthEmitEventDecorator creates a new EthEmitEventDecorator
func NewEthEmitEventDecorator(evmKeeper EVMKeeper) EthEmitEventDecorator {
	return EthEmitEventDecorator{evmKeeper}
}

// AnteHandle emits some basic events for the eth messages
func (eeed EthEmitEventDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// After eth tx passed ante handler, the fee is deducted and nonce increased, it shouldn't be ignored by json-rpc,
	// we need to emit some basic events at the very end of ante handler to be indexed by tendermint.
	blockTxIndex := eeed.evmKeeper.GetTxIndexTransient(ctx)

	for i, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*evmtypes.MsgEthereumTx)(nil))
		}

		txIdx := uint64(i) // nosec: G701
		EmitTxHashEvent(ctx, msgEthTx, blockTxIndex, txIdx)
	}

	return next(ctx, tx, simulate)
}

// EmitTxHashEvent emits the Ethereum tx
//
// FIXME: This is Technical debt. Ideally the sdk.Tx hash should be the Ethereum
// tx hash (msg.Hash) instead of using events for indexing Eth txs.
// TxIndex should be included in the header vote extension as part of ABCI++
func EmitTxHashEvent(ctx sdk.Context, msg *evmtypes.MsgEthereumTx, blockTxIndex, msgIndex uint64) {
	// emit ethereum tx hash as an event so that it can be indexed by Tendermint for query purposes
	// it's emitted in ante handler, so we can query failed transaction (out of block gas limit).
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			evmtypes.EventTypeEthereumTx,
			sdk.NewAttribute(evmtypes.AttributeKeyEthereumTxHash, msg.Hash),
			sdk.NewAttribute(evmtypes.AttributeKeyTxIndex, strconv.FormatUint(blockTxIndex+msgIndex, 10)), // #nosec G701
		),
	)
}
