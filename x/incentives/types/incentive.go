// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmostypes "github.com/evmos/evmos/v14/types"
)

// NewIncentive returns an instance of Incentive
func NewIncentive(
	contract common.Address,
	allocations sdk.DecCoins,
	epochs uint32,
) Incentive {
	return Incentive{
		Contract:    contract.String(),
		Allocations: allocations,
		Epochs:      epochs,
		TotalGas:    0,
	}
}

// Validate performs a stateless validation of a Incentive
func (i Incentive) Validate() error {
	if err := evmostypes.ValidateAddress(i.Contract); err != nil {
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

// IsActive returns true if the Incentive has remaining Epochs
func (i Incentive) IsActive() bool {
	return i.Epochs > 0
}
