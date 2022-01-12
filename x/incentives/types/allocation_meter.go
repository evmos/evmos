package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewAllocationMeter returns an instance of AllocationMeter
func NewAllocationMeter(allocation sdk.DecCoin) AllocationMeter {
	return AllocationMeter{
		Allocation: allocation,
	}
}

// Validate performs a stateless validation of a AllocationMeter
func (am AllocationMeter) Validate() error {
	if am.Allocation.IsZero() {
		return fmt.Errorf("total allocation cannot be empty: %s", am.Allocation)
	}

	if err := sdk.ValidateDenom(am.Allocation.Denom); err != nil {
		return err
	}

	if err := validateAmount(am.Allocation.Amount); err != nil {
		return err
	}

	return nil
}
