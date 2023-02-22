// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package tx

import (
	"math"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/utils"
)

var (
	feeAmt     = math.Pow10(16)
	DefaultFee = sdk.NewCoin(utils.BaseDenom, sdk.NewIntFromUint64(uint64(feeAmt))) // 0.01 EVMOS
)

// CosmosTxInput contains the input parameters required for preparing
// an EIP712 cosmos tx
type CosmosTxInput struct {
	// TxCfg is the client transaction config
	TxCfg client.TxConfig
	// Priv is the private key that will be used to sign the tx
	Priv cryptotypes.PrivKey
	// ChainID is the chain's id on cosmos format, e.g. 'evmos_9000-1'
	ChainID string
	// Gas to be used on the tx
	Gas uint64
	// GasPrice to use on tx
	GasPrice *sdkmath.Int
	// Fees is the fee to be used on the tx (amount and denom)
	Fees sdk.Coins
	// Msgs slice of messages to include on the tx
	Msgs []sdk.Msg
}

// PrepareCosmosTx creates a cosmos tx and signs it with the provided messages and private key.
// It returns the signed transaction and an error
func PrepareCosmosTx(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	input CosmosTxInput,
) (authsigning.Tx, error) {
	txBuilder := input.TxCfg.NewTxBuilder()

	txBuilder.SetGasLimit(input.Gas)

	var fees sdk.Coins
	if input.GasPrice != nil {
		fees = sdk.Coins{{Denom: utils.BaseDenom, Amount: input.GasPrice.MulRaw(int64(input.Gas))}}
	} else {
		fees = sdk.Coins{DefaultFee}
	}

	txBuilder.SetFeeAmount(fees)
	if err := txBuilder.SetMsgs(input.Msgs...); err != nil {
		return nil, err
	}

	return signCosmosTx(
		ctx,
		appEvmos,
		input,
		txBuilder,
	)
}

// signCosmosTx signs the cosmos transaction on the txBuilder provided using
// the provided private key
func signCosmosTx(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	input CosmosTxInput,
	txBuilder client.TxBuilder,
) (authsigning.Tx, error) {
	addr := sdk.AccAddress(input.Priv.PubKey().Address().Bytes())
	seq, err := appEvmos.AccountKeeper.GetSequence(ctx, addr)
	if err != nil {
		return nil, err
	}

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: input.Priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  input.TxCfg.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	}

	sigsV2 := []signing.SignatureV2{sigV2}

	if err := txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	accNumber := appEvmos.AccountKeeper.GetAccount(ctx, addr).GetAccountNumber()
	signerData := authsigning.SignerData{
		ChainID:       input.ChainID,
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = tx.SignWithPrivKey(
		input.TxCfg.SignModeHandler().DefaultMode(),
		signerData,
		txBuilder, input.Priv, input.TxCfg,
		seq,
	)
	if err != nil {
		return nil, err
	}

	sigsV2 = []signing.SignatureV2{sigV2}
	if err = txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, err
	}
	return txBuilder.GetTx(), nil
}

var _ sdk.Tx = &InvalidTx{}

// InvalidTx defines a type, which satisfies the sdk.Tx interface, but
// holds no valid transaction information.
//
// NOTE: This is used for testing purposes, to serve the edge case of invalid data being passed to functions.
type InvalidTx struct{}

func (InvalidTx) GetMsgs() []sdk.Msg { return []sdk.Msg{nil} }

func (InvalidTx) ValidateBasic() error { return nil }
