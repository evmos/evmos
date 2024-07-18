// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	"errors"
	"fmt"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	commonfactory "github.com/evmos/evmos/v19/testutil/integration/common/factory"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
)

// SubmitProposal is a helper function to submit a governance proposal and
// return the proposal ID.
func SubmitProposal(tf factory.TxFactory, network network.Network, proposerPriv cryptotypes.PrivKey, proposal govv1beta1.Content) (uint64, error) {
	proposerAccAddr := sdk.AccAddress(proposerPriv.PubKey().Address())

	msgSubmitProposal, err := govv1beta1.NewMsgSubmitProposal(
		proposal,
		sdk.NewCoins(sdk.NewCoin(network.GetDenom(), math.NewInt(1e18))),
		proposerAccAddr,
	)
	if err != nil {
		return 0, err
	}

	txArgs := commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{msgSubmitProposal},
	}

	res, err := tf.ExecuteCosmosTx(proposerPriv, txArgs)
	if err != nil {
		return 0, err
	}

	proposalID, err := getProposalIDFromEvents(res.Events)
	if err != nil {
		return 0, errorsmod.Wrap(err, "failed to get proposal ID from events")
	}

	err = network.NextBlock()
	if err != nil {
		return 0, errorsmod.Wrap(err, "failed to commit block after proposal")
	}

	gq := network.GetGovClient()
	proposalRes, err := gq.Proposal(network.GetContext(), &govv1.QueryProposalRequest{ProposalId: proposalID})
	if err != nil {
		return 0, errorsmod.Wrap(err, "failed to query proposal")
	}

	if proposalRes.Proposal.Status != govv1.StatusVotingPeriod {
		return 0, fmt.Errorf("expected proposal to be in voting period; got: %s", proposalRes.Proposal.Status.String())
	}

	return proposalRes.Proposal.GetId(), nil
}

// VoteOnProposal is a helper function to vote on a governance proposal given the private key of the voter and
// the option to vote.
func VoteOnProposal(tf factory.TxFactory, voterPriv cryptotypes.PrivKey, proposalID uint64, option govv1.VoteOption) error {
	voterAccAddr := sdk.AccAddress(voterPriv.PubKey().Address())

	msgVote := govv1.NewMsgVote(
		voterAccAddr,
		proposalID,
		option,
		"",
	)

	_, err := tf.ExecuteCosmosTx(voterPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{msgVote},
	})

	return err
}

// getProposalIDFromEvents returns the proposal ID from the events in
// the ResponseDeliverTx.
func getProposalIDFromEvents(events []abcitypes.Event) (uint64, error) {
	var (
		err        error
		found      bool
		proposalID uint64
	)

	for _, event := range events {
		if event.Type != govtypes.EventTypeProposalDeposit {
			continue
		}

		for _, attr := range event.Attributes {
			if attr.Key != govtypes.AttributeKeyProposalID {
				continue
			}

			proposalID, err = strconv.ParseUint(attr.Value, 10, 64)
			if err != nil {
				return 0, errorsmod.Wrap(err, "failed to parse proposal ID")
			}

			found = true
			break
		}

		if found {
			break
		}
	}

	if !found {
		return 0, errors.New("proposal deposit not found")
	}

	return proposalID, nil
}
