// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package factory

import (
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	testutiltypes "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/gogoproto/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/evmos/evmos/v19/precompiles/testutil"
	commonfactory "github.com/evmos/evmos/v19/testutil/integration/common/factory"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

// TxFactory defines a struct that can build and broadcast transactions for the Evmos
// network.
// Methods are organized by build sign and broadcast type methods.
type TxFactory interface {
	commonfactory.CoreTxFactory
	VestingTxFactory

	// GenerateDefaultTxTypeArgs generates a default ETH tx args for the desired tx type
	GenerateDefaultTxTypeArgs(sender common.Address, txType int) (evmtypes.EvmTxArgs, error)
	// GenerateSignedEthTx generates an Ethereum tx with the provided private key and txArgs but does not broadcast it.
	GenerateSignedEthTx(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs) (signing.Tx, error)
	// GenerateSignedMsgEthereumTx generates an MsgEthereumTx signed with the provided private key and txArgs.
	GenerateSignedMsgEthereumTx(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs) (evmtypes.MsgEthereumTx, error)

	// SignMsgEthereumTx signs a MsgEthereumTx with the provided private key.
	SignMsgEthereumTx(privKey cryptotypes.PrivKey, msgEthereumTx evmtypes.MsgEthereumTx) (evmtypes.MsgEthereumTx, error)

	// ExecuteEthTx builds, signs and broadcasts an Ethereum tx with the provided private key and txArgs.
	// If the txArgs are not provided, they will be populated with default values or gas estimations.
	ExecuteEthTx(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs) (abcitypes.ExecTxResult, error)
	// ExecuteContractCall executes a contract call with the provided private key
	ExecuteContractCall(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs, callArgs CallArgs) (abcitypes.ExecTxResult, error)
	// DeployContract deploys a contract with the provided private key,
	// compiled contract data and constructor arguments
	DeployContract(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs, deploymentData ContractDeploymentData) (common.Address, error)
	// CallContractAndCheckLogs is a helper function to call a contract and check the logs using
	// the integration test utilities.
	//
	// It returns the Cosmos Tx response, the decoded Ethereum Tx response and an error. This error value
	// is nil, if the expected logs are found and the VM error is the expected one, should one be expected.
	CallContractAndCheckLogs(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs, callArgs CallArgs, logCheckArgs testutil.LogCheckArgs) (abcitypes.ExecTxResult, *evmtypes.MsgEthereumTxResponse, error)
	// GenerateDeployContractArgs generates the txArgs for a contract deployment.
	GenerateDeployContractArgs(from common.Address, txArgs evmtypes.EvmTxArgs, deploymentData ContractDeploymentData) (evmtypes.EvmTxArgs, error)
	// GenerateContractCallArgs generates the txArgs for a contract call.
	GenerateContractCallArgs(txArgs evmtypes.EvmTxArgs, callArgs CallArgs) (evmtypes.EvmTxArgs, error)
	// GenerateMsgEthereumTx creates a new MsgEthereumTx with the provided arguments.
	GenerateMsgEthereumTx(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs) (evmtypes.MsgEthereumTx, error)
	// GenerateGethCoreMsg creates a new GethCoreMsg with the provided arguments.
	GenerateGethCoreMsg(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs) (core.Message, error)
	// EstimateGasLimit estimates the gas limit for a tx with the provided address and txArgs
	EstimateGasLimit(from *common.Address, txArgs *evmtypes.EvmTxArgs) (uint64, error)
	// GetEvmTransactionResponseFromTxResult returns the MsgEthereumTxResponse from the provided txResult
	GetEvmTransactionResponseFromTxResult(txResult abcitypes.ExecTxResult) (*evmtypes.MsgEthereumTxResponse, error)
}

var _ TxFactory = (*IntegrationTxFactory)(nil)

// IntegrationTxFactory is a helper struct to build and broadcast transactions
// to the network on integration tests. This is to simulate the behavior of a real user.
type IntegrationTxFactory struct {
	commonfactory.CoreTxFactory
	VestingTxFactory

	grpcHandler grpc.Handler
	network     network.Network
	ec          testutiltypes.TestEncodingConfig
}

// New creates a new IntegrationTxFactory instance
func New(
	network network.Network,
	grpcHandler grpc.Handler,
) TxFactory {
	cf := commonfactory.New(network, grpcHandler)
	return &IntegrationTxFactory{
		CoreTxFactory:    cf,
		VestingTxFactory: newVestingTxFactory(cf),
		grpcHandler:      grpcHandler,
		network:          network,
		ec:               network.GetEncodingConfig(),
	}
}

// GetEvmTransactionResponseFromTxResult returns the MsgEthereumTxResponse from the provided txResult.
func (tf *IntegrationTxFactory) GetEvmTransactionResponseFromTxResult(
	txResult abcitypes.ExecTxResult,
) (*evmtypes.MsgEthereumTxResponse, error) {
	var txData sdktypes.TxMsgData
	if err := tf.ec.Codec.Unmarshal(txResult.Data, &txData); err != nil {
		return nil, errorsmod.Wrap(err, "failed to unmarshal tx data")
	}

	if len(txData.MsgResponses) != 1 {
		return nil, fmt.Errorf("expected 1 message response, got %d", len(txData.MsgResponses))
	}

	var evmRes evmtypes.MsgEthereumTxResponse
	if err := proto.Unmarshal(txData.MsgResponses[0].Value, &evmRes); err != nil {
		return nil, errorsmod.Wrap(err, "failed to unmarshal evm tx response")
	}

	return &evmRes, nil
}

// populateEvmTxArgsWithDefault populates the missing fields in the provided EvmTxArgs with default values.
// If no GasLimit is present it will estimate the gas needed for the transaction.
func (tf *IntegrationTxFactory) populateEvmTxArgsWithDefault(
	fromAddr common.Address,
	txArgs evmtypes.EvmTxArgs,
) (evmtypes.EvmTxArgs, error) {
	if txArgs.ChainID == nil {
		txArgs.ChainID = tf.network.GetEIP155ChainID()
	}

	if txArgs.Nonce == 0 {
		accountResp, err := tf.grpcHandler.GetEvmAccount(fromAddr)
		if err != nil {
			return evmtypes.EvmTxArgs{}, errorsmod.Wrapf(err, "failed to get evm account: %s", fromAddr.String())
		}
		txArgs.Nonce = accountResp.GetNonce()
	}

	// If there is no GasPrice it is assumed this is a DynamicFeeTx.
	// If fields are empty they are populated with current dynamic values.
	if txArgs.GasPrice == nil {
		if txArgs.GasTipCap == nil {
			txArgs.GasTipCap = big.NewInt(1)
		}
		if txArgs.GasFeeCap == nil {
			baseFeeResp, err := tf.grpcHandler.GetBaseFee()
			if err != nil {
				return evmtypes.EvmTxArgs{}, errorsmod.Wrap(err, "failed to get base fee")
			}
			txArgs.GasFeeCap = baseFeeResp.BaseFee.BigInt()
		}
	}

	// If the gas limit is not set, estimate it
	// through the /simulate endpoint.
	if txArgs.GasLimit == 0 {
		gasLimit, err := tf.EstimateGasLimit(&fromAddr, &txArgs)
		if err != nil {
			return evmtypes.EvmTxArgs{}, errorsmod.Wrap(err, "failed to estimate gas limit")
		}
		txArgs.GasLimit = gasLimit
	}

	return txArgs, nil
}

func (tf *IntegrationTxFactory) encodeTx(tx sdktypes.Tx) ([]byte, error) {
	txConfig := tf.ec.TxConfig
	txBytes, err := txConfig.TxEncoder()(tx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to encode tx")
	}
	return txBytes, nil
}

func (tf *IntegrationTxFactory) buildSignedTx(msg evmtypes.MsgEthereumTx) (signing.Tx, error) {
	txConfig := tf.ec.TxConfig
	txBuilder := txConfig.NewTxBuilder()
	return msg.BuildTx(txBuilder, tf.network.GetDenom())
}

// checkEthTxResponse checks if the response is valid and returns the MsgEthereumTxResponse
func (tf *IntegrationTxFactory) checkEthTxResponse(res *abcitypes.ExecTxResult) error {
	var txData sdktypes.TxMsgData
	if !res.IsOK() {
		return fmt.Errorf("tx failed with Code: %d, Logs: %s", res.Code, res.Log)
	}

	cdc := tf.ec.Codec
	if err := cdc.Unmarshal(res.Data, &txData); err != nil {
		return errorsmod.Wrap(err, "failed to unmarshal tx data")
	}

	if len(txData.MsgResponses) != 1 {
		return fmt.Errorf("expected 1 message response, got %d", len(txData.MsgResponses))
	}

	var evmRes evmtypes.MsgEthereumTxResponse
	if err := proto.Unmarshal(txData.MsgResponses[0].Value, &evmRes); err != nil {
		return errorsmod.Wrap(err, "failed to unmarshal evm tx response")
	}

	if evmRes.Failed() {
		return fmt.Errorf("tx failed with VmError: %v, Logs: %s", evmRes.VmError, res.GetLog())
	}
	return nil
}
