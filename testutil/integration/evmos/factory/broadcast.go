// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package factory

import (
	errorsmod "cosmossdk.io/errors"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v19/precompiles/testutil"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

// ExecuteEthTx executes an Ethereum transaction - contract call with the provided private key and txArgs
// It first builds a MsgEthereumTx and then broadcasts it to the network.
func (tf *IntegrationTxFactory) ExecuteEthTx(
	priv cryptotypes.PrivKey,
	txArgs evmtypes.EvmTxArgs,
) (abcitypes.ResponseDeliverTx, error) {
	signedMsg, err := tf.GenerateSignedEthTx(priv, txArgs)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, errorsmod.Wrap(err, "failed to generate signed ethereum tx")
	}

	txBytes, err := tf.encodeTx(signedMsg)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, errorsmod.Wrap(err, "failed to encode ethereum tx")
	}

	res, err := tf.network.BroadcastTxSync(txBytes)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, errorsmod.Wrap(err, "failed to broadcast ethereum tx")
	}

	if err := tf.checkEthTxResponse(&res); err != nil {
		return res, errorsmod.Wrap(err, "failed ETH tx")
	}
	return res, nil
}

// ExecuteContractCall executes a contract call with the provided private key.
func (tf *IntegrationTxFactory) ExecuteContractCall(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs, callArgs CallArgs) (abcitypes.ResponseDeliverTx, error) {
	completeTxArgs, err := tf.GenerateContractCallArgs(txArgs, callArgs)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, errorsmod.Wrap(err, "failed to generate contract call args")
	}

	return tf.ExecuteEthTx(privKey, completeTxArgs)
}

// DeployContract deploys a contract with the provided private key,
// compiled contract data and constructor arguments.
// TxArgs Input and Nonce fields are overwritten.
func (tf *IntegrationTxFactory) DeployContract(
	priv cryptotypes.PrivKey,
	txArgs evmtypes.EvmTxArgs,
	deploymentData ContractDeploymentData,
) (common.Address, error) {
	// Get account's nonce to create contract hash
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	completeTxArgs, err := tf.GenerateDeployContractArgs(from, txArgs, deploymentData)
	if err != nil {
		return common.Address{}, errorsmod.Wrap(err, "failed to generate contract call args")
	}

	res, err := tf.ExecuteEthTx(priv, completeTxArgs)
	if err != nil || !res.IsOK() {
		return common.Address{}, errorsmod.Wrap(err, "failed to execute eth tx")
	}
	return crypto.CreateAddress(from, completeTxArgs.Nonce), nil
}

// CallContractAndCheckLogs is a helper function to call a contract and check the logs using
// the integration test utilities.
//
// It returns the Cosmos Tx response, the decoded Ethereum Tx response and an error. This error value
// is nil, if the expected logs are found and the VM error is the expected one, should one be expected.
func (tf *IntegrationTxFactory) CallContractAndCheckLogs(
	priv cryptotypes.PrivKey,
	txArgs evmtypes.EvmTxArgs,
	callArgs CallArgs,
	logCheckArgs testutil.LogCheckArgs,
) (abcitypes.ResponseDeliverTx, *evmtypes.MsgEthereumTxResponse, error) {
	res, err := tf.ExecuteContractCall(priv, txArgs, callArgs)
	logCheckArgs.Res = res
	if err != nil {
		// NOTE: here we are still passing the response to the log check function,
		// because we want to check the logs and expected error in case of a VM error.
		return abcitypes.ResponseDeliverTx{}, nil, CheckError(err, logCheckArgs)
	}

	ethRes, err := evmtypes.DecodeTxResponse(res.Data)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, nil, err
	}

	return res, ethRes, testutil.CheckLogs(logCheckArgs)
}
