// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package factory

import (
	"encoding/json"
	"errors"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v19/server/config"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

func (tf *IntegrationTxFactory) GenerateDefaultTxTypeArgs(sender common.Address, txType int) (evmtypes.EvmTxArgs, error) {
	defaultArgs := evmtypes.EvmTxArgs{}
	switch txType {
	case gethtypes.DynamicFeeTxType:
		return tf.populateEvmTxArgsWithDefault(sender, defaultArgs)
	case gethtypes.AccessListTxType:
		defaultArgs.Accesses = &gethtypes.AccessList{{
			Address:     sender,
			StorageKeys: []common.Hash{{0}},
		}}
		defaultArgs.GasPrice = big.NewInt(1e9)
		return tf.populateEvmTxArgsWithDefault(sender, defaultArgs)
	case gethtypes.LegacyTxType:
		defaultArgs.GasPrice = big.NewInt(1e9)
		return tf.populateEvmTxArgsWithDefault(sender, defaultArgs)
	default:
		return evmtypes.EvmTxArgs{}, errors.New("tx type not supported")
	}
}

// EstimateGasLimit estimates the gas limit for a tx with the provided address and txArgs
func (tf *IntegrationTxFactory) EstimateGasLimit(from *common.Address, txArgs *evmtypes.EvmTxArgs) (uint64, error) {
	args, err := json.Marshal(evmtypes.TransactionArgs{
		Data:       (*hexutil.Bytes)(&txArgs.Input),
		From:       from,
		To:         txArgs.To,
		AccessList: txArgs.Accesses,
	})
	if err != nil {
		return 0, errorsmod.Wrap(err, "failed to marshal tx args")
	}

	res, err := tf.grpcHandler.EstimateGas(args, config.DefaultGasCap)
	if err != nil {
		return 0, errorsmod.Wrap(err, "failed to estimate gas")
	}
	gas := res.Gas
	return gas, nil
}

// GenerateSignedEthTx generates an Ethereum tx with the provided private key and txArgs but does not broadcast it.
func (tf *IntegrationTxFactory) GenerateSignedEthTx(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs) (signing.Tx, error) {
	msgEthereumTx, err := tf.GenerateMsgEthereumTx(privKey, txArgs)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to create ethereum tx")
	}

	signedMsg, err := tf.SignMsgEthereumTx(privKey, msgEthereumTx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to sign ethereum tx")
	}

	// Validate the transaction to avoid unrealistic behavior
	if err = signedMsg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "failed to validate transaction")
	}

	return tf.buildSignedTx(signedMsg)
}

// GenerateMsgEthereumTx creates a new MsgEthereumTx with the provided arguments.
// If any of the arguments are not provided, they will be populated with default values.
func (tf *IntegrationTxFactory) GenerateMsgEthereumTx(
	privKey cryptotypes.PrivKey,
	txArgs evmtypes.EvmTxArgs,
) (evmtypes.MsgEthereumTx, error) {
	fromAddr := common.BytesToAddress(privKey.PubKey().Address().Bytes())
	// Fill TxArgs with default values
	txArgs, err := tf.populateEvmTxArgsWithDefault(fromAddr, txArgs)
	if err != nil {
		return evmtypes.MsgEthereumTx{}, errorsmod.Wrap(err, "failed to populate tx args")
	}
	msg := buildMsgEthereumTx(txArgs, fromAddr)

	return msg, nil
}

// GenerateGethCoreMsg creates a new GethCoreMsg with the provided arguments.
func (tf *IntegrationTxFactory) GenerateGethCoreMsg(
	privKey cryptotypes.PrivKey,
	txArgs evmtypes.EvmTxArgs,
) (core.Message, error) {
	msg, err := tf.GenerateMsgEthereumTx(privKey, txArgs)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to generate ethereum tx")
	}

	signedMsg, err := tf.SignMsgEthereumTx(privKey, msg)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to sign ethereum tx")
	}

	baseFeeResp, err := tf.grpcHandler.GetBaseFee()
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get base fee")
	}
	signer := gethtypes.LatestSignerForChainID(
		tf.network.GetEIP155ChainID(),
	)
	return signedMsg.AsMessage(signer, baseFeeResp.BaseFee.BigInt())
}

// GenerateContractCallArgs generates the txArgs for a contract call.
func (tf *IntegrationTxFactory) GenerateContractCallArgs(
	txArgs evmtypes.EvmTxArgs,
	callArgs CallArgs,
) (evmtypes.EvmTxArgs, error) {
	input, err := callArgs.ContractABI.Pack(callArgs.MethodName, callArgs.Args...)
	if err != nil {
		return evmtypes.EvmTxArgs{}, errorsmod.Wrap(err, "failed to pack contract arguments")
	}
	txArgs.Input = input
	return txArgs, nil
}

// GenerateDeployContractArgs generates the txArgs for a contract deployment.
func (tf *IntegrationTxFactory) GenerateDeployContractArgs(
	from common.Address,
	txArgs evmtypes.EvmTxArgs,
	deploymentData ContractDeploymentData,
) (evmtypes.EvmTxArgs, error) {
	account, err := tf.grpcHandler.GetEvmAccount(from)
	if err != nil {
		return evmtypes.EvmTxArgs{}, errorsmod.Wrapf(err, "failed to get evm account: %s", from.String())
	}
	txArgs.Nonce = account.GetNonce()

	ctorArgs, err := deploymentData.Contract.ABI.Pack("", deploymentData.ConstructorArgs...)
	if err != nil {
		return evmtypes.EvmTxArgs{}, errorsmod.Wrap(err, "failed to pack constructor arguments")
	}
	data := deploymentData.Contract.Bin
	data = append(data, ctorArgs...)

	txArgs.Input = data
	return txArgs, nil
}
