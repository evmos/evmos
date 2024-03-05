// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package network

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	sdktypes "github.com/cosmos/cosmos-sdk/store/types"
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
	n.app.EndBlocker(n.ctx, abci.RequestEndBlock{Height: header.Height})
	n.app.Commit()

	// Calculate new block time after duration
	newBlockTime := header.Time.Add(duration)

	// Update block header and BeginBlock
	header.Height++
	header.AppHash = n.app.LastCommitID().Hash
	header.Time = newBlockTime
	n.app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// Update context header
	newCtx := n.app.BaseApp.NewContext(false, header)
	newCtx = newCtx.WithMinGasPrices(n.ctx.MinGasPrices())
	newCtx = newCtx.WithKVGasConfig(n.ctx.KVGasConfig())
	newCtx = newCtx.WithTransientKVGasConfig(n.ctx.TransientKVGasConfig())
	newCtx = newCtx.WithConsensusParams(n.ctx.ConsensusParams())
	// This might have to be changed with time if we want to test gas limits
	newCtx = newCtx.WithBlockGasMeter(sdktypes.NewInfiniteGasMeter())

	n.ctx = newCtx
	return nil
}
