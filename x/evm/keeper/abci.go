// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper

import (
	"encoding/json"
	"math/big"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/evmos/evmos/v19/x/evm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// BeginBlock sets the sdk Context and EIP155 chain id to the Keeper.
func (k *Keeper) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	k.WithChainID(ctx)
	// check for eth_tx_log events at BeginBlock
	// and update block bloom filter with them (if any)
	k.updateBlockBloom(ctx)
}

// EndBlock also retrieves the bloom filter value from the transient store and commits it to the
// KVStore. The EVM end block logic doesn't update the validator set, thus it returns
// an empty slice.
func (k *Keeper) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	// Gas costs are handled within msg handler so costs should be ignored
	infCtx := ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	bloom := ethtypes.BytesToBloom(k.GetBlockBloomTransient(infCtx).Bytes())
	k.EmitBlockBloomEvent(infCtx, bloom)

	return []abci.ValidatorUpdate{}
}

// updateBlockBloom checks for eth_tx_log events at BeginBlock
// only those on the modules BeginBlocker should be included (eg. epochs)
// and update block bloom filter with these
func (k *Keeper) updateBlockBloom(ctx sdk.Context) {
	logger := ctx.Logger().With("begin_block", "evm")
	var logs []*ethtypes.Log
	for _, event := range ctx.EventManager().Events() {
		if event.Type != types.EventTypeTxLog {
			continue
		}
		ls, err := parseLog(event)
		if err != nil {
			logger.Error("error when parsing logs", "error", err.Error())
			continue
		}
		logs = append(logs, ls...)
	}

	// Update block bloom filter
	logsCount := len(logs)
	if logsCount > 0 {
		logger.Debug("updating block bloom filter", "logs_count", logsCount, "block_height", ctx.BlockHeight())
		bloom := k.GetBlockBloomTransient(ctx)
		bloom.Or(bloom, big.NewInt(0).SetBytes(ethtypes.LogsBloom(logs)))
		k.SetBlockBloomTransient(ctx, bloom)
		k.SetLogSizeTransient(ctx, uint64(logsCount))
	}
}

func parseLog(event sdk.Event) (logs []*ethtypes.Log, err error) {
	for _, attr := range event.Attributes {
		if attr.Key != types.AttributeKeyTxLog {
			continue
		}

		var log ethtypes.Log
		if err = json.Unmarshal([]byte(attr.Value), &log); err != nil {
			return
		}

		logs = append(logs, &log)
	}
	return
}
