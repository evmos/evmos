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

package testutil

import (
	"math/big"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v11/encoding"
	"github.com/evmos/evmos/v11/utils"
	evmtypes "github.com/evmos/evmos/v11/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// SubmitProposal delivers a submit proposal tx for a given gov content.
// Depending on the content type, the eventNum needs to specify submit_proposal
// event.
func SubmitProposal(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	pk *ethsecp256k1.PrivKey,
	content govv1beta1.Content,
	eventNum int,
) (id uint64, err error) {
	accountAddress := sdk.AccAddress(pk.PubKey().Address().Bytes())
	stakeDenom := stakingtypes.DefaultParams().BondDenom

	deposit := sdk.NewCoins(sdk.NewCoin(stakeDenom, sdk.NewInt(100000000)))
	msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, accountAddress)
	if err != nil {
		return id, err
	}
	res, err := DeliverTx(ctx, appEvmos, pk, msg)
	if err != nil {
		return id, err
	}

	submitEvent := res.GetEvents()[eventNum]
	if submitEvent.Type != "submit_proposal" || string(submitEvent.Attributes[0].Key) != "proposal_id" {
		return id, errorsmod.Wrapf(errorsmod.Error{}, "eventNumber %d in SubmitProposal calls %s instead of submit_proposal", eventNum, submitEvent.Type)
	}

	return strconv.ParseUint(string(submitEvent.Attributes[0].Value), 10, 64)
}

// Delegate delivers a delegate tx
func Delegate(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	priv *ethsecp256k1.PrivKey,
	delegateAmount sdk.Coin,
	validator stakingtypes.Validator,
) (abci.ResponseDeliverTx, error) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	val, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	delegateMsg := stakingtypes.NewMsgDelegate(accountAddress, val, delegateAmount)
	return DeliverTx(ctx, appEvmos, priv, delegateMsg)
}

// Vote delivers a vote tx with the VoteOption "yes"
func Vote(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	priv *ethsecp256k1.PrivKey,
	proposalID uint64,
	voteOption govv1beta1.VoteOption,
) (abci.ResponseDeliverTx, error) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	voteMsg := govv1beta1.NewMsgVote(accountAddress, proposalID, voteOption)
	return DeliverTx(ctx, appEvmos, priv, voteMsg)
}

// DeliverTx delivers a tx for a given set of msgs
func DeliverTx(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	priv *ethsecp256k1.PrivKey,
	msgs ...sdk.Msg,
) (abci.ResponseDeliverTx, error) {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(100_000_000)
	amt, _ := sdk.NewIntFromString("1000000000000000000000")
	txBuilder.SetFeeAmount(sdk.Coins{{Denom: utils.BaseDenom, Amount: amt}})
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	seq, err := appEvmos.AccountKeeper.GetSequence(ctx, accountAddress)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	}

	sigsV2 := []signing.SignatureV2{sigV2}

	if err := txBuilder.SetSignatures(sigsV2...); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	accNumber := appEvmos.AccountKeeper.GetAccount(ctx, accountAddress).GetAccountNumber()
	signerData := authsigning.SignerData{
		ChainID:       ctx.ChainID(),
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = tx.SignWithPrivKey(
		encodingConfig.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, priv, encodingConfig.TxConfig,
		seq,
	)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	sigsV2 = []signing.SignatureV2{sigV2}
	if err = txBuilder.SetSignatures(sigsV2...); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return BroadcastTxBytes(appEvmos, encodingConfig.TxConfig.TxEncoder(), txBuilder.GetTx())
}

// DeliverEthTx generates and broadcasts a Cosmos Tx populated with MsgEthereumTx messages.
// If a private key is provided, it will attempt to sign all messages with the given private key,
// otherwise, it will assume the messages have already been signed.
func DeliverEthTx(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	priv *ethsecp256k1.PrivKey,
	msgs ...sdk.Msg,
) (abci.ResponseDeliverTx, error) {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	denom := appEvmos.ClaimsKeeper.GetParams(ctx).ClaimsDenom

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	signer := ethtypes.LatestSignerForChainID(appEvmos.EvmKeeper.ChainID())
	txFee := sdk.Coins{}
	txGasLimit := uint64(0)

	// Sign messages and compute gas/fees.
	for _, m := range msgs {
		msg, ok := m.(*evmtypes.MsgEthereumTx)
		if !ok {
			return abci.ResponseDeliverTx{}, errorsmod.Wrapf(errorsmod.Error{}, "cannot mix Ethereum and Cosmos messages in one Tx")
		}

		if priv != nil {
			err := msg.Sign(signer, NewSigner(priv))
			if err != nil {
				return abci.ResponseDeliverTx{}, err
			}
		}

		msg.From = ""

		txGasLimit += msg.GetGas()
		txFee = txFee.Add(sdk.Coin{Denom: denom, Amount: math.NewIntFromBigInt(msg.GetFee())})
	}

	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	// Set the extension
	var option *codectypes.Any
	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return abci.ResponseDeliverTx{}, errorsmod.Wrapf(errorsmod.Error{}, "could not set extensions for Ethereum tx")
	}

	builder.SetExtensionOptions(option)

	txBuilder.SetGasLimit(txGasLimit)
	txBuilder.SetFeeAmount(txFee)

	return BroadcastTxBytes(appEvmos, encodingConfig.TxConfig.TxEncoder(), txBuilder.GetTx())
}

// CreateEthTx is a helper function to create and sign an Ethereum transaction.
//
// If the given private key is not nil, it will be used to sign the transaction.
//
// It offers the ability to increment the nonce by a given amount in case one wants to set up
// multiple transactions that are supposed to be executed one after another.
// Should this not be the case, just pass in zero.
func CreateEthTx(ctx sdk.Context, appEvmos *app.Evmos, privKey *ethsecp256k1.PrivKey, from sdk.AccAddress, dest sdk.AccAddress, amount *big.Int, nonceIncrement int) (*evmtypes.MsgEthereumTx, error) {
	toAddr := common.BytesToAddress(dest.Bytes())
	fromAddr := common.BytesToAddress(from.Bytes())
	chainID := appEvmos.EvmKeeper.ChainID()

	// When we send multiple Ethereum Tx's in one Cosmos Tx, we need to increment the nonce for each one.
	nonce := appEvmos.EvmKeeper.GetNonce(ctx, fromAddr) + uint64(nonceIncrement)
	msgEthereumTx := evmtypes.NewTx(chainID, nonce, &toAddr, amount, 100000, nil, appEvmos.FeeMarketKeeper.GetBaseFee(ctx), big.NewInt(1), nil, &ethtypes.AccessList{})
	msgEthereumTx.From = fromAddr.String()

	// If we are creating multiple eth Tx's with different senders, we need to sign here rather than later.
	if privKey != nil {
		signer := ethtypes.LatestSignerForChainID(appEvmos.EvmKeeper.ChainID())
		err := msgEthereumTx.Sign(signer, NewSigner(privKey))
		if err != nil {
			return nil, err
		}
	}

	return msgEthereumTx, nil
}

// BroadcastTxBytes encodes a transaction and calls DeliverTx on the app.
func BroadcastTxBytes(app *app.Evmos, txEncoder sdk.TxEncoder, tx sdk.Tx) (abci.ResponseDeliverTx, error) {
	// bz are bytes to be broadcasted over the network
	bz, err := txEncoder(tx)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	req := abci.RequestDeliverTx{Tx: bz}
	res := app.BaseApp.DeliverTx(req)
	if res.Code != 0 {
		return abci.ResponseDeliverTx{}, errorsmod.Wrapf(errortypes.ErrInvalidRequest, res.Log)
	}

	return res, nil
}
