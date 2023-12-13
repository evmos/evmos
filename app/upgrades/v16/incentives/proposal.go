// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package incentives

import (
	"errors"
	fmt "fmt"

	evmostypes "github.com/evmos/evmos/v16/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govcdc "github.com/cosmos/cosmos-sdk/x/gov/codec"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// constants
const (
	RouterKey                            = "incentives"
	ProposalTypeRegisterIncentive string = "RegisterIncentive"
)

// Implements Proposal Interface
var (
	_ govv1beta1.Content = &RegisterIncentiveProposal{}
)

func init() {
	govv1beta1.RegisterProposalType(ProposalTypeRegisterIncentive)
	govcdc.ModuleCdc.Amino.RegisterConcrete(&RegisterIncentiveProposal{}, "incentives/RegisterIncentiveProposal", nil)
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
	if err := evmostypes.ValidateAddress(rip.Contract); err != nil {
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
