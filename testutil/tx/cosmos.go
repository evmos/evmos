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

// CosmosTxArgs contains the params to create a cosmos tx
type CosmosTxArgs struct {
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
	args CosmosTxArgs,
) (authsigning.Tx, error) {
	txBuilder := args.TxCfg.NewTxBuilder()

	txBuilder.SetGasLimit(args.Gas)

	var fees sdk.Coins
	if args.GasPrice != nil {
		fees = sdk.Coins{{Denom: utils.BaseDenom, Amount: args.GasPrice.MulRaw(int64(args.Gas))}}
	} else {
		fees = sdk.Coins{DefaultFee}
	}

	txBuilder.SetFeeAmount(fees)
	if err := txBuilder.SetMsgs(args.Msgs...); err != nil {
		return nil, err
	}

	return signCosmosTx(
		ctx,
		appEvmos,
		args,
		txBuilder,
	)
}

// signCosmosTx signs the cosmos transaction on the txBuilder provided using
// the provided private key
func signCosmosTx(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	args CosmosTxArgs,
	txBuilder client.TxBuilder,
) (authsigning.Tx, error) {
	addr := sdk.AccAddress(args.Priv.PubKey().Address().Bytes())
	seq, err := appEvmos.AccountKeeper.GetSequence(ctx, addr)
	if err != nil {
		return nil, err
	}

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: args.Priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  args.TxCfg.SignModeHandler().DefaultMode(),
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
		ChainID:       args.ChainID,
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = tx.SignWithPrivKey(
		args.TxCfg.SignModeHandler().DefaultMode(),
		signerData,
		txBuilder, args.Priv, args.TxCfg,
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
