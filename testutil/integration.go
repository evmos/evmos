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
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/evmos/evmos/v12/app"
	"github.com/evmos/evmos/v12/crypto/ethsecp256k1"
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
	res, err := DeliverTx(ctx, appEvmos, pk, nil, msg)
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
	return DeliverTx(ctx, appEvmos, priv, nil, delegateMsg)
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
	return DeliverTx(ctx, appEvmos, priv, nil, voteMsg)
}
