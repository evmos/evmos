// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcdc "github.com/cosmos/cosmos-sdk/x/gov/codec"
	v1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	evmostypes "github.com/evmos/evmos/v16/types"
)

// constants
const (
	ProposalTypeRegisterERC20         string = "RegisterERC20"
	ProposalTypeToggleTokenConversion string = "ToggleTokenConversion" // #nosec
)

// Implements Proposal Interface
var (
	_ v1beta1.Content = &RegisterERC20Proposal{}
	_ v1beta1.Content = &ToggleTokenConversionProposal{}
)

func init() {
	v1beta1.RegisterProposalType(ProposalTypeRegisterERC20)
	v1beta1.RegisterProposalType(ProposalTypeToggleTokenConversion)
	govcdc.ModuleCdc.Amino.RegisterConcrete(&RegisterERC20Proposal{}, "erc20/RegisterERC20Proposal", nil)
	govcdc.ModuleCdc.Amino.RegisterConcrete(&ToggleTokenConversionProposal{}, "erc20/ToggleTokenConversionProposal", nil)
}

// CreateDenomDescription generates a string with the coin description
func CreateDenomDescription(address string) string {
	return fmt.Sprintf("Cosmos coin token representation of %s", address)
}

// CreateDenom generates a string the module name plus the address to avoid conflicts with names staring with a number
func CreateDenom(address string) string {
	return fmt.Sprintf("%s/%s", ModuleName, address)
}

// ValidateErc20Denom checks if a denom is a valid erc20/
// denomination
func ValidateErc20Denom(denom string) error {
	denomSplit := strings.SplitN(denom, "/", 2)

	if len(denomSplit) != 2 || denomSplit[0] != ModuleName {
		return fmt.Errorf("invalid denom. %s denomination should be prefixed with the format 'erc20/", denom)
	}

	return evmostypes.ValidateAddress(denomSplit[1])
}

// NewRegisterERC20Proposal returns new instance of RegisterERC20Proposal
func NewRegisterERC20Proposal(title, description string, erc20Addreses ...string) v1beta1.Content {
	return &RegisterERC20Proposal{
		Title:          title,
		Description:    description,
		Erc20Addresses: erc20Addreses,
	}
}

// ProposalRoute returns router key for this proposal
func (*RegisterERC20Proposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*RegisterERC20Proposal) ProposalType() string {
	return ProposalTypeRegisterERC20
}

// ValidateBasic performs a stateless check of the proposal fields
func (rtbp *RegisterERC20Proposal) ValidateBasic() error {
	for _, address := range rtbp.Erc20Addresses {
		if err := evmostypes.ValidateAddress(address); err != nil {
			return errorsmod.Wrap(err, "ERC20 address")
		}
	}

	return v1beta1.ValidateAbstract(rtbp)
}

// NewToggleTokenConversionProposal returns new instance of ToggleTokenConversionProposal
func NewToggleTokenConversionProposal(title, description string, token string) v1beta1.Content {
	return &ToggleTokenConversionProposal{
		Title:       title,
		Description: description,
		Token:       token,
	}
}

// ProposalRoute returns router key for this proposal
func (*ToggleTokenConversionProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*ToggleTokenConversionProposal) ProposalType() string {
	return ProposalTypeToggleTokenConversion
}

// ValidateBasic performs a stateless check of the proposal fields
func (ttcp *ToggleTokenConversionProposal) ValidateBasic() error {
	// check if the token is a hex address, if not, check if it is a valid SDK
	// denom
	if err := evmostypes.ValidateAddress(ttcp.Token); err != nil {
		if err := sdk.ValidateDenom(ttcp.Token); err != nil {
			return err
		}
	}

	return v1beta1.ValidateAbstract(ttcp)
}
