package types

import (
	"errors"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ethermint "github.com/evmos/ethermint/types"
)

// constants
const (
	ProposalTypeRegisterIncentive string = "RegisterIncentive"
	ProposalTypeCancelIncentive   string = "CancelIncentive"
)

// Implements Proposal Interface
var (
	_ govtypes.Content = &RegisterIncentiveProposal{}
	_ govtypes.Content = &CancelIncentiveProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeRegisterIncentive)
	govtypes.RegisterProposalType(ProposalTypeCancelIncentive)
	govtypes.RegisterProposalTypeCodec(&RegisterIncentiveProposal{}, "incentives/RegisterIncentiveProposal")
	govtypes.RegisterProposalTypeCodec(&CancelIncentiveProposal{}, "incentives/CancelIncentiveProposal")
}

// NewRegisterIncentiveProposal returns new instance of RegisterIncentiveProposal
func NewRegisterIncentiveProposal(
	title, description, contract string,
	allocations sdk.DecCoins,
	epochs uint32,
) govtypes.Content {
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

	return govtypes.ValidateAbstract(rip)
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
) govtypes.Content {
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

	return govtypes.ValidateAbstract(rip)
}
