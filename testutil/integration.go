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
	"strconv"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/encoding"

	"github.com/evmos/evmos/v10/app"
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
	denom := appEvmos.ClaimsKeeper.GetParams(ctx).ClaimsDenom

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(100_000_000)
	txBuilder.SetFeeAmount(sdk.Coins{{Denom: denom, Amount: sdk.NewInt(1)}})
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

	// bz are bytes to be broadcasted over the network
	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	req := abci.RequestDeliverTx{Tx: bz}
	res := appEvmos.BaseApp.DeliverTx(req)
	if res.Code != 0 {
		return abci.ResponseDeliverTx{}, errorsmod.Wrapf(errortypes.ErrInvalidRequest, res.Log)
	}

	return res, nil
}
