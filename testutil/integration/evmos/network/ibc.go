// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package network

import (
	"testing"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"
)

// GetIBCChain returns a TestChain instance for the given network.
// Note: the sender accounts are not populated. Do not use this accounts to send transactions during tests.
// The keyring should be used instead.
func (n *IntegrationNetwork) GetIBCChain(t *testing.T, coord *ibctesting.Coordinator) *ibctesting.TestChain {
	return &ibctesting.TestChain{
		T:             t,
		Coordinator:   coord,
		ChainID:       n.GetChainID(),
		App:           n.app,
		CurrentHeader: n.ctx.BlockHeader(),
		QueryServer:   n.app.GetIBCKeeper(),
		TxConfig:      n.app.GetTxConfig(),
		Codec:         n.app.AppCodec(),
		Vals:          n.valSet,
		NextVals:      n.valSet,
	}
}
