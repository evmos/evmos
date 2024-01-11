package types

import (
	abci "github.com/cometbft/cometbft/abci/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

type ResponseDeliverEthTx struct {
	CosmosTxHash         string
	EthTxHash            string
	EvmError             string
	ResponseDeliverEthTx *abci.ResponseDeliverTx
}

func NewResponseDeliverEthTx(responseDeliverTx *abci.ResponseDeliverTx) *ResponseDeliverEthTx {
	if responseDeliverTx == nil {
		return nil
	}

	response := &ResponseDeliverEthTx{
		ResponseDeliverEthTx: responseDeliverTx,
	}

	for _, event := range responseDeliverTx.Events {
		if event.Type == evmtypes.TypeMsgEthereumTx {
			for _, attribute := range event.Attributes {
				//fmt.Println(evmtypes.TypeMsgEthereumTx, "attribute.Key", attribute.Key, "attribute.Value", attribute.Value)
				if attribute.Key == evmtypes.AttributeKeyTxHash {
					if len(attribute.Value) > 0 && response.CosmosTxHash == "" {
						response.CosmosTxHash = attribute.Value
					}
				} else if attribute.Key == evmtypes.AttributeKeyEthereumTxHash {
					if len(attribute.Value) > 0 && response.EthTxHash == "" {
						response.EthTxHash = attribute.Value
					}
				} else if attribute.Key == evmtypes.AttributeKeyEthereumTxFailed {
					if len(attribute.Value) > 0 && response.EvmError == "" {
						response.EvmError = attribute.Value
					}
				}
			}
		}
	}

	return response
}
