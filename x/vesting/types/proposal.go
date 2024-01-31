// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// constants
const ProposalTypeClawback string = "Clawback"

// Implements Proposal Interface

var _ v1beta1.Content = &ClawbackProposal{}

func init() {
	v1beta1.RegisterProposalType(ProposalTypeClawback)
}

// NewClawbackProposal returns new instance of RegisterClawbackProposal
func NewClawbackProposal(title, description, address, destinationAddress string) v1beta1.Content {
	return &ClawbackProposal{
		Title:              title,
		Description:        description,
		Address:            address,
		DestinationAddress: destinationAddress,
	}
}

// ProposalRoute returns router key for this proposal
func (*ClawbackProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*ClawbackProposal) ProposalType() string {
	return ProposalTypeClawback
}

// ValidateBasic performs a stateless check of the proposal fields
func (cbp *ClawbackProposal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(cbp.Address); err != nil {
		return errorsmod.Wrap(err, "vesting account address")
	}

	if cbp.DestinationAddress != "" {
		if _, err := sdk.AccAddressFromBech32(cbp.DestinationAddress); err != nil {
			return errorsmod.Wrap(err, "vesting account destination address")
		}
	}

	return v1beta1.ValidateAbstract(cbp)
}
