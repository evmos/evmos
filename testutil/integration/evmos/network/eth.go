package network

import (
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

// ExecuteEthCall runs a query call to the EVM.
func (n *IntegrationNetwork) ExecuteEthCall(req *evmtypes.EthCallRequest) (*evmtypes.MsgEthereumTxResponse, error) {
	return n.app.EvmKeeper.EthCall(n.ctx, req)
}
