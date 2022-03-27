package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/tharsis/ethermint/types"
)

// NewFee returns an instance of Fee
func NewFee(
	contract common.Address,
) FeeContract {
	return FeeContract{
		Contract: contract.String(),
	}
}

// Validate performs a stateless validation of a FeeContract
func (i FeeContract) Validate() error {
	if err := ethermint.ValidateAddress(i.Contract); err != nil {
		return err
	}

	if i.Allocations.IsZero() {
		return fmt.Errorf("allocations cannot be empty: %s", i.Allocations)
	}

	for _, al := range i.Allocations {
		if err := sdk.ValidateDenom(al.Denom); err != nil {
			return err
		}
		if err := validateAmount(al.Amount); err != nil {
			return err
		}
	}

	if i.Epochs == 0 {
		return fmt.Errorf("epoch cannot be 0")
	}
	return nil
}

// IsActive returns true if the FeeContract has remaining Epochs
func (i FeeContract) IsActive() bool {
	return i.Epochs > 0
}
