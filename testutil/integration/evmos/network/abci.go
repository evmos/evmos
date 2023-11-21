// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package network

// NextBlock is a private helper function that runs the EndBlocker logic, commits the changes,
// updates the header and runs the BeginBlocker
func (n *IntegrationNetwork) NextBlock() error {
	// End block and commit
	header := n.ctx.BlockHeader()
	n.app.EndBlocker(n.ctx)
	n.app.Commit()

	// Update block header and BeginBlock
	header.Height++
	header.AppHash = n.app.LastCommitID().Hash

	// Update context header
	newCtx := n.app.BaseApp.NewContextLegacy(false, header)
	newCtx = newCtx.WithMinGasPrices(n.ctx.MinGasPrices())
	newCtx = newCtx.WithEventManager(n.ctx.EventManager())
	newCtx = newCtx.WithKVGasConfig(n.ctx.KVGasConfig())
	newCtx = newCtx.WithTransientKVGasConfig(n.ctx.TransientKVGasConfig())
	n.app.BeginBlocker(newCtx)

	n.ctx = newCtx
	return nil
}
