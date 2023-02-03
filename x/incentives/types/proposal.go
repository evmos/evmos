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
	"errors"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	ethermint "github.com/evmos/evmos/v11/types"
)

// constants
const (
	ProposalTypeRegisterIncentive string = "RegisterIncentive"
	ProposalTypeCancelIncentive   string = "CancelIncentive"
)

// Implements Proposal Interface
var (
	_ govv1beta1.Content = &RegisterIncentiveProposal{}
	_ govv1beta1.Content = &CancelIncentiveProposal{}
)

func init() {
	govv1beta1.RegisterProposalType(ProposalTypeRegisterIncentive)
	govv1beta1.RegisterProposalType(ProposalTypeCancelIncentive)
	govv1beta1.ModuleCdc.Amino.RegisterConcrete(&RegisterIncentiveProposal{}, "incentives/RegisterIncentiveProposal", nil)
	govv1beta1.ModuleCdc.Amino.RegisterConcrete(&CancelIncentiveProposal{}, "incentives/CancelIncentiveProposal", nil)
}

// NewRegisterIncentiveProposal returns new instance of RegisterIncentiveProposal
func NewRegisterIncentiveProposal(
	title, description, contract string,
	allocations sdk.DecCoins,
	epochs uint32,
) govv1beta1.Content {
	return &RegisterIncentiveProposal{
		Title:       title,
		Description: description,
		Contract:    contract,
		Allocations: allocations,
		Epochs:      epochs,
	}
}

// ProposalRoute returns router key for this proposal
func (*RegisterIncentiveProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*RegisterIncentiveProposal) ProposalType() string {
	return ProposalTypeRegisterIncentive
}

// ValidateBasic performs a stateless check of the proposal fields
func (rip *RegisterIncentiveProposal) ValidateBasic() error {
	if err := ethermint.ValidateAddress(rip.Contract); err != nil {
		return err
	}

	if err := validateAllocations(rip.Allocations); err != nil {
		return err
	}

	if err := validateEpochs(rip.Epochs); err != nil {
		return err
	}

	return govv1beta1.ValidateAbstract(rip)
}

// validateAllocations checks if each allocation has
// - a valid denom
// - a valid amount representing the percentage of allocation
func validateAllocations(allocations sdk.DecCoins) error {
	if allocations.Empty() {
		return errors.New("incentive allocations cannot be empty")
	}

	for _, al := range allocations {
		if err := validateAmount(al.Amount); err != nil {
			return err
		}
	}

	return allocations.Validate()
}

func validateAmount(amount sdk.Dec) error {
	if amount.GT(sdk.OneDec()) || amount.LTE(sdk.ZeroDec()) {
		return fmt.Errorf("invalid amount for allocation: %s", amount)
	}
	return nil
}

func validateEpochs(epochs uint32) error {
	if epochs == 0 {
		return fmt.Errorf("epochs value (%d) cannot be 0", epochs)
	}
	return nil
}

// NewCancelIncentiveProposal returns new instance of RegisterIncentiveProposal
func NewCancelIncentiveProposal(
	title, description, contract string,
) govv1beta1.Content {
	return &CancelIncentiveProposal{
		Title:       title,
		Description: description,
		Contract:    contract,
	}
}

// ProposalRoute returns router key for this proposal
func (*CancelIncentiveProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*CancelIncentiveProposal) ProposalType() string {
	return ProposalTypeCancelIncentive
}

// ValidateBasic performs a stateless check of the proposal fields
func (rip *CancelIncentiveProposal) ValidateBasic() error {
	if err := ethermint.ValidateAddress(rip.Contract); err != nil {
		return err
	}

	return govv1beta1.ValidateAbstract(rip)
}
