// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
)

// ERC20RegistrationData is the necessary data to provide in order to register an ERC20 token.
type ERC20RegistrationData struct {
	// Address is the address of the ERC20 token.
	Address common.Address
	// Denom is the ERC20 token denom.
	Denom string
	// ProposerPriv is the private key used to sign the proposal and voting transactions.
	ProposerPriv cryptotypes.PrivKey
}

// ValidateBasic does stateless validation of the data for the ERC20 registration.
func (ed ERC20RegistrationData) ValidateBasic() error {
	emptyAddr := common.Address{}
	if ed.Address.Hex() == emptyAddr.Hex() {
		return fmt.Errorf("address cannot be empty")
	}

	if ed.Denom == "" {
		return fmt.Errorf("denom cannot be empty")
	}

	if ed.ProposerPriv == nil {
		return fmt.Errorf("proposer private key cannot be nil")
	}

	return nil
}

// RegisterERC20 is a helper function to register ERC20 token through
// submitting a governance proposal and having it pass.
// It returns the registered token pair.
func RegisterERC20(tf factory.TxFactory, network network.Network, data ERC20RegistrationData) (erc20types.TokenPair, error) {
	err := data.ValidateBasic()
	if err != nil {
		return erc20types.TokenPair{}, errorsmod.Wrap(err, "failed to validate erc20 registration data")
	}

	proposal := erc20types.RegisterERC20Proposal{
		Title:          fmt.Sprintf("Register %s Token", data.Denom),
		Description:    fmt.Sprintf("This proposal registers the ERC20 token at address: %s", data.Address.Hex()),
		Erc20Addresses: []string{data.Address.Hex()},
	}

	// Submit the proposal
	proposalID, err := SubmitProposal(tf, network, data.ProposerPriv, &proposal)
	if err != nil {
		return erc20types.TokenPair{}, errorsmod.Wrap(err, "failed to submit proposal")
	}

	err = network.NextBlock()
	if err != nil {
		return erc20types.TokenPair{}, errorsmod.Wrap(err, "failed to commit block after proposal")
	}

	// Vote on proposal
	err = VoteOnProposal(tf, data.ProposerPriv, proposalID, govtypes.OptionYes)
	if err != nil {
		return erc20types.TokenPair{}, errorsmod.Wrap(err, "failed to vote on proposal")
	}

	gq := network.GetGovClient()
	params, err := gq.Params(network.GetContext(), &govtypes.QueryParamsRequest{ParamsType: "voting"})
	if err != nil {
		return erc20types.TokenPair{}, errorsmod.Wrap(err, "failed to query voting params")
	}

	err = network.NextBlockAfter(*params.Params.VotingPeriod) // commit after voting period is over
	if err != nil {
		return erc20types.TokenPair{}, errorsmod.Wrap(err, "failed to commit block after voting period ends")
	}

	err = network.NextBlock()
	if err != nil {
		return erc20types.TokenPair{}, errorsmod.Wrap(err, "failed to commit block after votes")
	}

	// NOTE: it's necessary to instantiate a new gov client here, because the previous one has now an outdated
	// context.
	gq = network.GetGovClient()
	proposalRes, err := gq.Proposal(network.GetContext(), &govtypes.QueryProposalRequest{ProposalId: proposalID})
	if err != nil {
		return erc20types.TokenPair{}, errorsmod.Wrap(err, "failed to query proposal")
	}

	if proposalRes.Proposal.Status != govtypes.StatusPassed {
		return erc20types.TokenPair{}, fmt.Errorf("proposal did not pass; got status: %s", proposalRes.Proposal.Status.String())
	}

	// Check if token pair is registered
	eq := network.GetERC20Client()
	tokenPairRes, err := eq.TokenPair(network.GetContext(), &erc20types.QueryTokenPairRequest{Token: data.Address.Hex()})
	if err != nil {
		return erc20types.TokenPair{}, errorsmod.Wrap(err, "failed to query token pair")
	}

	return tokenPairRes.TokenPair, nil
}
