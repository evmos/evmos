// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package factory

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/client"
	cosmostx "github.com/cosmos/cosmos-sdk/client/tx"

	sdkmath "cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
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
	account, err := tf.grpcHandler.GetAccount(senderAddress.String())
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to get account: %s", senderAddress.String())
	}

	sequence := account.GetSequence()
	signMode := txConfig.SignModeHandler().DefaultMode()
	signerData := xauthsigning.SignerData{
		ChainID:       tf.network.GetChainID(),
		AccountNumber: account.GetAccountNumber(),
		Sequence:      sequence,
		Address:       senderAddress.String(),
	}

	// sign tx
	sigsV2 := signing.SignatureV2{
		PubKey: privKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signMode,
			Signature: nil,
		},
		Sequence: sequence,
	}

	err = txBuilder.SetSignatures(sigsV2)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to set tx signatures")
	}

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

	signature, err := cosmostx.SignWithPrivKey(signMode, signerData, txBuilder, privKey, txConfig, sequence)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to sign tx")
	}

	err = txBuilder.SetSignatures(signature)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to set tx signatures")
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
