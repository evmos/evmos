package utils

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	errorsmod "cosmossdk.io/errors"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/ethereum/go-ethereum/common"
	commonfactory "github.com/evmos/evmos/v15/testutil/integration/common/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	erc20types "github.com/evmos/evmos/v15/x/erc20/types"
)

// RegisterERC20 is a helper function to register ERC20 token through
// submitting a governance proposal and having it pass.
func RegisterERC20(tf factory.TxFactory, network network.Network, erc20Addr common.Address, proposerPriv cryptotypes.PrivKey) error {
	proposal := erc20types.RegisterERC20Proposal{
		Title:          "Register ERC20 Token",
		Description:    fmt.Sprintf("This proposal registers the ERC20 token at address: %s", erc20Addr.Hex()),
		Erc20Addresses: []string{erc20Addr.Hex()},
	}

	proposerAccAddr := sdk.AccAddress(proposerPriv.PubKey().Address())

	// Submit the proposal
	msgSubmitProposal, err := govv1beta1.NewMsgSubmitProposal(
		&proposal,
		sdk.NewCoins(sdk.NewCoin(network.GetDenom(), sdk.NewInt(1e18))),
		proposerAccAddr,
	)
	if err != nil {
		return err
	}

	txArgs := commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{msgSubmitProposal},
	}

	res, err := tf.ExecuteCosmosTx(proposerPriv, txArgs)
	if err != nil {
		return err
	}

	proposalID, err := getProposalIDFromEvents(res.Events)
	if err != nil {
		return errorsmod.Wrap(err, "failed to get proposal ID from events")
	}
	fmt.Printf("Proposal submitted with ID: %d\n", proposalID)

	gq := network.GetGovClient()
	proposalRes, err := gq.Proposal(network.GetContext(), &govtypes.QueryProposalRequest{ProposalId: proposalID})
	if err != nil {
		return errorsmod.Wrap(err, "failed to query proposal")
	}
	fmt.Printf("Proposal status: %s\n", proposalRes.Proposal.Status)

	err = network.NextBlock()
	if err != nil {
		return errorsmod.Wrap(err, "failed to commit block after proposal")
	}

	// Vote on proposal
	msgVote := govtypes.NewMsgVote(
		proposerAccAddr,
		proposalID,
		govtypes.OptionYes,
		"",
	)

	res, err = tf.ExecuteCosmosTx(proposerPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{msgVote},
	})
	if err != nil {
		return errorsmod.Wrap(err, "failed to vote on proposal")
	}

	fmt.Printf("Found %d events voting\n", len(res.Events))

	err = network.NextBlockAfter(365 * 24 * time.Hour) // commit a year later
	if err != nil {
		return errorsmod.Wrap(err, "failed to commit block after voting period ends")
	}
	err = network.NextBlock()
	if err != nil {
		return errorsmod.Wrap(err, "failed to commit block after proposal")
	}

	// Check if proposal passed
	proposalRes, err = gq.Proposal(network.GetContext(), &govtypes.QueryProposalRequest{ProposalId: proposalID})
	if err != nil {
		return errorsmod.Wrap(err, "failed to query proposal")
	}

	if proposalRes.Proposal.Status != govtypes.StatusPassed {
		return fmt.Errorf("proposal did not pass; got status: %s", proposalRes.Proposal.Status.String())
	}
	fmt.Printf("Proposal status: %s\n", proposalRes.Proposal.Status)

	return nil
}

// getProposalIDFromEvents returns the proposal ID from the events in
// the ResponseDeliverTx.
func getProposalIDFromEvents(events []abcitypes.Event) (uint64, error) {
	var (
		err        error
		found      = false
		proposalID uint64
	)

	for _, event := range events {
		if event.Type != "proposal_deposit" {
			continue
		}

		for _, attr := range event.Attributes {
			if attr.Key != "proposal_id" {
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
