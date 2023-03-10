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

package types

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	v1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	evmostypes "github.com/evmos/evmos/v12/types"
)

// constants
const (
	ProposalTypeRegisterCoin          string = "RegisterCoin"
	ProposalTypeRegisterERC20         string = "RegisterERC20"
	ProposalTypeToggleTokenConversion string = "ToggleTokenConversion" // #nosec
)

// Implements Proposal Interface
var (
	_ v1beta1.Content = &RegisterCoinProposal{}
	_ v1beta1.Content = &RegisterERC20Proposal{}
	_ v1beta1.Content = &ToggleTokenConversionProposal{}
)

func init() {
	v1beta1.RegisterProposalType(ProposalTypeRegisterCoin)
	v1beta1.RegisterProposalType(ProposalTypeRegisterERC20)
	v1beta1.RegisterProposalType(ProposalTypeToggleTokenConversion)
	v1beta1.ModuleCdc.Amino.RegisterConcrete(&RegisterCoinProposal{}, "erc20/RegisterCoinProposal", nil)
	v1beta1.ModuleCdc.Amino.RegisterConcrete(&RegisterERC20Proposal{}, "erc20/RegisterERC20Proposal", nil)
	v1beta1.ModuleCdc.Amino.RegisterConcrete(&ToggleTokenConversionProposal{}, "erc20/ToggleTokenConversionProposal", nil)
}

// CreateDenomDescription generates a string with the coin description
func CreateDenomDescription(address string) string {
	return fmt.Sprintf("Cosmos coin token representation of %s", address)
}

// CreateDenom generates a string the module name plus the address to avoid conflicts with names staring with a number
func CreateDenom(address string) string {
	return fmt.Sprintf("%s/%s", ModuleName, address)
}

// NewRegisterCoinProposal returns new instance of RegisterCoinProposal
func NewRegisterCoinProposal(title, description string, coinMetadata ...banktypes.Metadata) v1beta1.Content {
	return &RegisterCoinProposal{
		Title:       title,
		Description: description,
		Metadata:    coinMetadata,
	}
}

// ProposalRoute returns router key for this proposal
func (*RegisterCoinProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*RegisterCoinProposal) ProposalType() string {
	return ProposalTypeRegisterCoin
}

// ValidateBasic performs a stateless check of the proposal fields
func (rtbp *RegisterCoinProposal) ValidateBasic() error {
	for _, metadata := range rtbp.Metadata {
		if err := metadata.Validate(); err != nil {
			return err
		}

		// Prohibit denominations that contain the evm denom
		if strings.Contains(metadata.Base, "evm") {
			return errorsmod.Wrapf(
				ErrEVMDenom, "cannot register the EVM denomination %s", metadata.Base,
			)
		}

		if err := ibctransfertypes.ValidateIBCDenom(metadata.Base); err != nil {
			return err
		}

		if err := validateIBCVoucherMetadata(metadata); err != nil {
			return err
		}
	}

	return v1beta1.ValidateAbstract(rtbp)
}

// validateIBCVoucherMetadata checks that the coin metadata fields are consistent
// with an IBC voucher denomination.
func validateIBCVoucherMetadata(metadata banktypes.Metadata) error {
	// Check ibc/ denom
	denomSplit := strings.SplitN(metadata.Base, "/", 2)

	if denomSplit[0] == metadata.Base && strings.TrimSpace(metadata.Base) != "" {
		// Not IBC
		return nil
	}

	if len(denomSplit) != 2 || denomSplit[0] != ibctransfertypes.DenomPrefix {
		// NOTE: should be unaccessible (covered on ValidateIBCDenom)
		return fmt.Errorf("invalid metadata. %s denomination should be prefixed with the format 'ibc/", metadata.Base)
	}

	return nil
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
