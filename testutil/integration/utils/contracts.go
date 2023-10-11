package utils

import (
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/evmos/evmos/v14/testutil/integration/factory"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	evmostypes "github.com/evmos/evmos/v14/types"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"
)

// CheckTxTopics checks if all expected topics are present in the transaction response
func CheckTxTopics(res abcitypes.ResponseDeliverTx, expectedTopics []string) error {
	msgEthResponse, err := decodeResponseDeliverTx(res)
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

// IsContractAccount checks if the given account is a contract account
func IsContractAccount(acc authtypes.AccountI) error {
	contractETHAccount, ok := acc.(evmostypes.EthAccountI)
	if !ok {
		return fmt.Errorf("account is not an eth account")
	}

	emptyCodeHash := common.BytesToHash(crypto.Keccak256(nil))
	if contractETHAccount.GetCodeHash() == emptyCodeHash {
		return fmt.Errorf("account code hash is empty")
	}

	if contractETHAccount.Type() != evmostypes.AccountTypeContract {
		return fmt.Errorf("account is not a contract account")
	}
	return nil
}

// DecodeContractCallResponse decodes the response of a contract call query
func DecodeContractCallResponse(response interface{}, callArgs factory.CallArgs, res abcitypes.ResponseDeliverTx) error {
	msgEthResponse, err := decodeResponseDeliverTx(res)
	if err != nil {
		return err
	}

	err = callArgs.ContractABI.UnpackIntoInterface(response, callArgs.MethodName, msgEthResponse.Ret)
	if err != nil {
		return err
	}
	return nil
}

func decodeResponseDeliverTx(res abcitypes.ResponseDeliverTx) (*evmtypes.MsgEthereumTxResponse, error) {
	msgEthResponse, err := evmtypes.DecodeTxResponse(res.Data)
	if err != nil {
		return nil, err
	}
	return msgEthResponse, nil
}
