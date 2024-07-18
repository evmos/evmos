// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package utils

import (
	"fmt"
	"slices"

	"github.com/evmos/evmos/v19/testutil/integration/evmos/factory"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

// CheckTxTopics checks if all expected topics are present in the transaction response
func CheckTxTopics(res abcitypes.ResponseDeliverTx, expectedTopics []string) error {
	msgEthResponse, err := DecodeResponseDeliverTx(res)
	if err != nil {
		return err
	}

	// Collect all topics within the transaction
	availableLogs := make([]string, 0, len(msgEthResponse.Logs))
	for _, log := range msgEthResponse.Logs {
		availableLogs = append(availableLogs, log.Topics...)
	}

	// Check if all expected topics are present
	for _, expectedTopic := range expectedTopics {
		if !slices.Contains(availableLogs, expectedTopic) {
			return fmt.Errorf("expected topic %s not found in tx response", expectedTopic)
		}
	}
	return nil
}

// DecodeContractCallResponse decodes the response of a contract call query
func DecodeContractCallResponse(response interface{}, callArgs factory.CallArgs, res abcitypes.ResponseDeliverTx) error {
	msgEthResponse, err := DecodeResponseDeliverTx(res)
	if err != nil {
		return err
	}

	err = callArgs.ContractABI.UnpackIntoInterface(response, callArgs.MethodName, msgEthResponse.Ret)
	if err != nil {
		return err
	}
	return nil
}

func DecodeResponseDeliverTx(res abcitypes.ResponseDeliverTx) (*evmtypes.MsgEthereumTxResponse, error) {
	msgEthResponse, err := evmtypes.DecodeTxResponse(res.Data)
	if err != nil {
		return nil, err
	}
	return msgEthResponse, nil
}
