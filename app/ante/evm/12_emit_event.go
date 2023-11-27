// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evm

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

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
