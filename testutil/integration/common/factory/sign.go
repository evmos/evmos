// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package factory

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	cosmostx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// setSignatures is a helper function that sets the signature for
// the transaction in the tx builder. It returns the signerData to be used
// when signing the transaction (e.g. when calling signWithPrivKey)
func (tf *baseTxFactory) setSignatures(privKey cryptotypes.PrivKey, txBuilder client.TxBuilder, signMode signing.SignMode) (signerData authsigning.SignerData, err error) {
	senderAddress := sdktypes.AccAddress(privKey.PubKey().Address().Bytes())
	account, err := tf.grpcHandler.GetAccount(senderAddress.String())
	if err != nil {
		return signerData, err
	}
	sequence := account.GetSequence()
	signerData = authsigning.SignerData{
		ChainID:       tf.network.GetChainID(),
		AccountNumber: account.GetAccountNumber(),
		Sequence:      sequence,
		Address:       senderAddress.String(),
		PubKey:        privKey.PubKey(),
	}

	sigsV2 := signing.SignatureV2{
		PubKey: privKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signMode,
			Signature: nil,
		},
		Sequence: sequence,
	}

	return signerData, txBuilder.SetSignatures(sigsV2)
}

// signWithPrivKey is a helper function that signs a transaction
// with the provided private key
func (tf *baseTxFactory) signWithPrivKey(privKey cryptotypes.PrivKey, txBuilder client.TxBuilder, signerData authsigning.SignerData, signMode signing.SignMode) error {
	txConfig := tf.ec.TxConfig
	signature, err := cosmostx.SignWithPrivKey(context.TODO(), signMode, signerData, txBuilder, privKey, txConfig, signerData.Sequence)
	if err != nil {
		return errorsmod.Wrap(err, "failed to sign tx")
	}

	return txBuilder.SetSignatures(signature)
}
