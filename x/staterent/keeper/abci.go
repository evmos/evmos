// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v19/x/staterent/types"
)

// BeginBlocker of staterent module
func (k Keeper) BeginBlocker(ctx sdk.Context) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	logger := k.Logger(ctx)

	p := k.GetParams(ctx)
	nextTic := p.CurrentTicBlock + int64(p.BlocksPerTic)

	// Array used to keep track of the empty flagged info objects
	entriesToRemove := []common.Address{}

	// Delete entries
	k.IterateFlaggedInfo(ctx, func(_ int64, info types.FlaggedInfo) bool {
		if info.StartDeletionTic >= p.CurrentTic {
			// Get the entries that needs to be deleted
			addr := common.HexToAddress(info.Contract)
			i := int64(0)
			entriesToDelete := []common.Hash{}
			k.evmKeeper.ForEachStorage(ctx, addr, func(key common.Hash, value common.Hash) bool {
				entriesToDelete = append(entriesToDelete, key)
				if i < p.EntriesToDeletePerBlock {
					return false
				}
				return true
			})

			// Delete the entries
			for _, v := range entriesToDelete {
				k.evmKeeper.DeleteState(ctx, addr, v)
			}

			logger.Debug("remove-contract-storage", "amount", len(entriesToDelete))
			// TODO: add event here

			if len(entriesToDelete) < int(p.EntriesToDeletePerBlock) {
				// Mark the entry to be deleted
				entriesToRemove = append(entriesToRemove, addr)
			} else {
				// Update info values
				info.CurrentDeletedEntries = info.CurrentDeletedEntries.Add(math.NewInt(p.EntriesToDeletePerBlock))
				k.SetFlaggedInfo(ctx, addr, info)
			}
		}

		return false
	})

	// Delete the flagged info for contracts that are empty
	for _, v := range entriesToRemove {
		k.DeleteFlaggedInfo(ctx, v)
		logger.Info("remove-flagged-info", "account", v.Hex())
		// TODO: add event here
	}

	// Logic for evey tic
	if nextTic == ctx.BlockHeight() {
		// TODO: find a fast to get storage usage by contract, in the meanwhile we add via governance the contracts with high usage of storage
		// TODO: until we have a good way to read the amount of entries of a contract, the amount of entries must be updated via governance
		k.IterateFlaggedInfo(ctx, func(_ int64, info types.FlaggedInfo) bool {
			// TODO: if the state was paid, keep the FlaggedInfo as active and burn the payment

			// TODO: if the state was not paid, mark the contract as inactive and mark the deleting tic as current + 1

			// TODO: emit event
			return false
		})

		// Update for next tic
		p.CurrentTic = p.CurrentTic + 1
		p.CurrentTicBlock = ctx.BlockHeight()
		k.SetParams(ctx, p)
	}
}
