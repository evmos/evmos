package types

import (
	"errors"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ethermint "github.com/tharsis/ethermint/types"
)

// constants
const (
	ProposalTypeRegisterContract string = "RegisterFeeContract"
	ProposalTypeCancelContract   string = "CancelFeeContract"
)

// Implements Proposal Interface
var (
	_ govtypes.Content = &RegisterContractProposal{}
	_ govtypes.Content = &CancelContractProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeRegisterContract)
	govtypes.RegisterProposalType(ProposalTypeCancelContract)
	govtypes.RegisterProposalTypeCodec(&RegisterContractProposal{}, "incentives/RegisterContractProposal")
	govtypes.RegisterProposalTypeCodec(&CancelContractProposal{}, "incentives/CancelContractProposal")
}

// NewRegisterContractProposal returns new instance of RegisterContractProposal
func NewRegisterContractProposal(
	title, description, contract string,
	allocations sdk.DecCoins,
	epochs uint32,
) govtypes.Content {
	return &RegisterContractProposal{
		Title:       title,
		Description: description,
		Contract:    contract,
		Allocations: allocations,
		Epochs:      epochs,
	}
}

// ProposalRoute returns router key for this proposal
func (*RegisterContractProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*RegisterContractProposal) ProposalType() string {
	return ProposalTypeRegisterContract
}

// ValidateBasic performs a stateless check of the proposal fields
func (rip *RegisterContractProposal) ValidateBasic() error {
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

// NewCancelContractProposal returns new instance of RegisterContractProposal
func NewCancelContractProposal(
	title, description, contract string,
) govtypes.Content {
	return &CancelContractProposal{
		Title:       title,
		Description: description,
		Contract:    contract,
	}
}

// ProposalRoute returns router key for this proposal
func (*CancelContractProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*CancelContractProposal) ProposalType() string {
	return ProposalTypeCancelContract
}

// ValidateBasic performs a stateless check of the proposal fields
func (rip *CancelContractProposal) ValidateBasic() error {
	if err := ethermint.ValidateAddress(rip.Contract); err != nil {
		return err
	}

	return govtypes.ValidateAbstract(rip)
}
