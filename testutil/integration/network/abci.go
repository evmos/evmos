package network

import (
	abci "github.com/cometbft/cometbft/abci/types"
)

// CommitBlock is a private helper function that runs the EndBlocker logic, commits the changes,
// updates the header and runs the BeginBlocker
func (n *Network) CommitBlock() error {
	// End block and commit
	header := n.ctx.BlockHeader()
	n.App.EndBlocker(n.ctx, abci.RequestEndBlock{Height: header.Height})
	n.App.Commit()

	// Update block header and BeginBlock
	header.Height++
	header.AppHash = n.App.LastCommitID().Hash
	n.App.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// Update context header
	newCtx := n.App.BaseApp.NewContext(false, header)
	newCtx = newCtx.WithMinGasPrices(n.ctx.MinGasPrices())
	newCtx = newCtx.WithEventManager(n.ctx.EventManager())
	newCtx = newCtx.WithKVGasConfig(n.ctx.KVGasConfig())
	newCtx = newCtx.WithTransientKVGasConfig(n.ctx.TransientKVGasConfig())

	n.ctx = newCtx
	return nil
}
