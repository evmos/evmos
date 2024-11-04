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
	commonfactory "github.com/evmos/evmos/v20/testutil/integration/common/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
)

// SubmitProposal is a helper function to submit a governance proposal and
// return the proposal ID.
func SubmitProposal(tf factory.TxFactory, network network.Network, proposerPriv cryptotypes.PrivKey, title string, msgs ...sdk.Msg) (uint64, error) {
	proposerAccAddr := sdk.AccAddress(proposerPriv.PubKey().Address()).String()
	proposal, err := govv1.NewMsgSubmitProposal(
		msgs,
		sdk.NewCoins(sdk.NewCoin(network.GetDenom(), math.NewInt(1e18))),
		proposerAccAddr,
		"",
		title,
		title,
		false,
	)
	if err != nil {
		return 0, err
	}

	txArgs := commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{proposal},
	}

	return submitProposal(tf, network, proposerPriv, txArgs)
}

// SubmitLegacyProposal is a helper function to submit a governance proposal and
// return the proposal ID.
func SubmitLegacyProposal(tf factory.TxFactory, network network.Network, proposerPriv cryptotypes.PrivKey, proposal govv1beta1.Content) (uint64, error) {
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

	return submitProposal(tf, network, proposerPriv, txArgs)
}

// VoteOnProposal is a helper function to vote on a governance proposal given the private key of the voter and
// the option to vote.
func VoteOnProposal(tf factory.TxFactory, voterPriv cryptotypes.PrivKey, proposalID uint64, option govv1.VoteOption) (abcitypes.ExecTxResult, error) {
	voterAccAddr := sdk.AccAddress(voterPriv.PubKey().Address())

	msgVote := govv1.NewMsgVote(
		voterAccAddr,
		proposalID,
		option,
		"",
	)

	res, err := tf.CommitCosmosTx(voterPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{msgVote},
	})

	return res, err
}

// ApproveProposal is a helper function to vote 'yes'
// for it and wait till it passes.
func ApproveProposal(tf factory.TxFactory, network network.Network, proposerPriv cryptotypes.PrivKey, proposalID uint64) error {
	// Vote on proposal
	if _, err := VoteOnProposal(tf, proposerPriv, proposalID, govv1.OptionYes); err != nil {
		return errorsmod.Wrap(err, "failed to vote on proposal")
	}

	if err := waitVotingPeriod(network); err != nil {
		return errorsmod.Wrap(err, "failed to wait for voting period to pass")
	}

	return checkProposalStatus(network, proposalID, govv1.StatusPassed)
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

func submitProposal(tf factory.TxFactory, network network.Network, proposerPriv cryptotypes.PrivKey, txArgs commonfactory.CosmosTxArgs) (uint64, error) {
	res, err := tf.CommitCosmosTx(proposerPriv, txArgs)
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

	if err := checkProposalStatus(network, proposalID, govv1.StatusVotingPeriod); err != nil {
		return 0, errorsmod.Wrap(err, "error while checking proposal")
	}

	return proposalID, nil
}

// waitVotingPeriod is a helper function that waits for the current voting period
// defined in the gov module params to pass
func waitVotingPeriod(n network.Network) error {
	gq := n.GetGovClient()
	params, err := gq.Params(n.GetContext(), &govv1.QueryParamsRequest{ParamsType: "voting"})
	if err != nil {
		return errorsmod.Wrap(err, "failed to query voting params")
	}

	err = n.NextBlockAfter(*params.Params.VotingPeriod) // commit after voting period is over
	if err != nil {
		return errorsmod.Wrap(err, "failed to commit block after voting period ends")
	}

	return n.NextBlock()
}

// checkProposalStatus is a helper function to check for a specific proposal status
func checkProposalStatus(n network.Network, proposalID uint64, expStatus govv1.ProposalStatus) error {
	gq := n.GetGovClient()
	proposalRes, err := gq.Proposal(n.GetContext(), &govv1.QueryProposalRequest{ProposalId: proposalID})
	if err != nil {
		return errorsmod.Wrap(err, "failed to query proposal")
	}

	if proposalRes.Proposal.Status != expStatus {
		return fmt.Errorf("proposal status different than expected. Expected %s; got: %s", expStatus.String(), proposalRes.Proposal.Status.String())
	}
	return nil
}
