package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate
func (cr ClaimRecord) Validate() error {
	if !cr.InitialClaimableAmount.IsPositive() {
		return fmt.Errorf("initial claimable amount is not positive, %s", cr.InitialClaimableAmount)
	}

	if len(Action_value)-1 != len(cr.ActionsCompleted) {
		return fmt.Errorf("action length mismatch, expected %d, got %d", len(Action_value)-1, len(cr.ActionsCompleted))
	}

	return nil
}

// NewClaimRecordAddress creates a new claim record instance
func NewClaimRecordAddress(address sdk.AccAddress, initialClaimableAmt sdk.Int) ClaimRecordAddress {
	return ClaimRecordAddress{
		Address:                address.String(),
		InitialClaimableAmount: initialClaimableAmt,
		ActionsCompleted:       []bool{false, false, false, false},
	}
}

func (cra ClaimRecordAddress) Validate() error {
	if _, err := sdk.AccAddressFromBech32(cra.Address); err != nil {
		return err
	}

	if !cra.InitialClaimableAmount.IsPositive() {
		return fmt.Errorf("initial claimable amount is not positive, %s", cra.InitialClaimableAmount)
	}

	if len(Action_value)-1 != len(cra.ActionsCompleted) {
		return fmt.Errorf("action length mismatch, expected %d, got %d", len(Action_value)-1, len(cra.ActionsCompleted))
	}

	return nil
}
