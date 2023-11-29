// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package network

import (
	"time"

	storetypes "cosmossdk.io/store/types"
)

// NextBlock is a private helper function that runs the EndBlocker logic, commits the changes,
// updates the header and runs the BeginBlocker
func (n *IntegrationNetwork) NextBlock() error {
	return n.NextBlockAfter(time.Second)
}

// NextBlockAfter is a private helper function that runs the EndBlocker logic, commits the changes,
// updates the header to have a block time after the given duration and runs the BeginBlocker.
func (n *IntegrationNetwork) NextBlockAfter(duration time.Duration) error {
	// End block and commit
	header := n.ctx.BlockHeader()
	if _, err := n.app.EndBlocker(n.ctx); err != nil {
		return err
	}

	n.app.Commit()

	// Calculate new block time after duration
	newBlockTime := header.Time.Add(duration)

	// Update block header and BeginBlock
	header.Height++
	header.AppHash = n.app.LastCommitID().Hash
	header.Time = newBlockTime

	// Update context header
	newCtx := n.app.BaseApp.NewContextLegacy(false, header)
	newCtx = newCtx.WithMinGasPrices(n.ctx.MinGasPrices())
	newCtx = newCtx.WithEventManager(n.ctx.EventManager())
	newCtx = newCtx.WithKVGasConfig(n.ctx.KVGasConfig())
	newCtx = newCtx.WithTransientKVGasConfig(n.ctx.TransientKVGasConfig())
	newCtx = newCtx.WithConsensusParams(n.ctx.ConsensusParams())
	// This might have to be changed with time if we want to test gas limits
	newCtx = newCtx.WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	if _, err := n.app.BeginBlocker(newCtx); err != nil {
		return err
	}

	n.ctx = newCtx
	return nil
}
