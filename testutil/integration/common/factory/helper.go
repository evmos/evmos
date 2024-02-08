// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package factory

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// buildTx builds a tx with the provided private key and txArgs
func (tf *IntegrationTxFactory) buildTx(privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (client.TxBuilder, error) {
	txConfig := tf.ec.TxConfig
	txBuilder := txConfig.NewTxBuilder()

	if err := txBuilder.SetMsgs(txArgs.Msgs...); err != nil {
		return nil, errorsmod.Wrap(err, "failed to set tx msgs")
	}

	if txArgs.FeeGranter != nil {
		txBuilder.SetFeeGranter(txArgs.FeeGranter)
	}

	senderAddress := sdktypes.AccAddress(privKey.PubKey().Address().Bytes())

	if txArgs.FeeGranter != nil {
		txBuilder.SetFeeGranter(txArgs.FeeGranter)
	}

	txBuilder.SetFeePayer(senderAddress)
	gasLimit, err := tf.estimateGas(txArgs, txBuilder)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to estimate gas")
	}
	txBuilder.SetGasLimit(gasLimit)

	fees, err := tf.calculateFees(txArgs.GasPrice, gasLimit)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to calculate fees")
	}
	txBuilder.SetFeeAmount(fees)

	if err := tf.SignCosmosTx(privKey, txBuilder); err != nil {
		return nil, errorsmod.Wrap(err, "failed to sign Cosmos Tx")
	}

	return txBuilder, nil
}

// calculateFees calculates the fees for the transaction.
func (tf *IntegrationTxFactory) calculateFees(gasPrice *sdkmath.Int, gasLimit uint64) (sdktypes.Coins, error) {
	denom := tf.network.GetDenom()
	var fees sdktypes.Coins
	if gasPrice != nil {
		fees = sdktypes.Coins{{Denom: denom, Amount: gasPrice.MulRaw(int64(gasLimit))}}
	} else {
		baseFee, err := tf.grpcHandler.GetBaseFee()
		if err != nil {
			return sdktypes.Coins{}, errorsmod.Wrap(err, "failed to get base fee")
		}
		price := baseFee.BaseFee
		fees = sdktypes.Coins{{Denom: denom, Amount: price.MulRaw(int64(gasLimit))}}
	}
	return fees, nil
}

// estimateGas estimates the gas needed for the transaction.
func (tf *IntegrationTxFactory) estimateGas(txArgs CosmosTxArgs, txBuilder client.TxBuilder) (uint64, error) {
	txConfig := tf.ec.TxConfig
	simulateBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return 0, errorsmod.Wrap(err, "failed to encode tx")
	}

	var gasLimit uint64
	if txArgs.Gas == 0 {
		simulateRes, err := tf.network.Simulate(simulateBytes)
		if err != nil {
			return 0, errorsmod.Wrap(err, "failed to simulate tx")
		}

		gasAdj := new(big.Float).SetFloat64(GasAdjustment)
		gasUsed := new(big.Float).SetUint64(simulateRes.GasInfo.GasUsed)
		gasLimit, _ = gasAdj.Mul(gasAdj, gasUsed).Uint64()
	} else {
		gasLimit = txArgs.Gas
	}
	return gasLimit, nil
}

// encodeTx encodes the tx using the txConfig's encoder.
func (tf *IntegrationTxFactory) encodeTx(tx sdktypes.Tx) ([]byte, error) {
	txConfig := tf.ec.TxConfig
	txBytes, err := txConfig.TxEncoder()(tx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to encode tx")
	}
	return txBytes, nil
}
