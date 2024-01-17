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
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v16/server/config"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

func (tf *IntegrationTxFactory) GenerateDefaultTxTypeArgs(sender common.Address, txType int) (evmtypes.EvmTxArgs, error) {
	defaultArgs := evmtypes.EvmTxArgs{}
	switch txType {
	case gethtypes.DynamicFeeTxType:
		return tf.populateEvmTxArgs(sender, defaultArgs)
	case gethtypes.AccessListTxType:
		defaultArgs.Accesses = &gethtypes.AccessList{{
			Address:     sender,
			StorageKeys: []common.Hash{{0}},
		}}
		defaultArgs.GasPrice = big.NewInt(1e9)
		return tf.populateEvmTxArgs(sender, defaultArgs)
	case gethtypes.LegacyTxType:
		defaultArgs.GasPrice = big.NewInt(1e9)
		return tf.populateEvmTxArgs(sender, defaultArgs)
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
	msgEthereumTx, err := tf.createMsgEthereumTx(privKey, txArgs)
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
