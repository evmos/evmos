package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewClaimRecord creates a new claim record instance
func NewClaimRecord(initialClaimableAmt sdk.Int) ClaimRecord {
	return ClaimRecord{
		InitialClaimableAmount: initialClaimableAmt,
		ActionsCompleted:       []bool{false, false, false, false},
	}
}

// Validate
func (cr ClaimRecord) Validate() error {
	if cr.InitialClaimableAmount.IsNil() {
		return errors.New("initial claimable amount is nil")
	}
	if !cr.InitialClaimableAmount.IsPositive() {
		return fmt.Errorf("initial claimable amount is not positive, %s", cr.InitialClaimableAmount)
	}
	if len(cr.ActionsCompleted) == 0 || len(Action_value)-1 != len(cr.ActionsCompleted) {
		return fmt.Errorf("action length mismatch, expected %d, got %d", len(Action_value)-1, len(cr.ActionsCompleted))
	}

	return nil
}

func (cr *ClaimRecord) ClaimAction(action Action) {
	cr.ActionsCompleted[action-1] = true
}

// HasClaimedAction checks if the user has claimed a given action
func (cr ClaimRecord) HasClaimedAction(action Action) bool {
	return len(cr.ActionsCompleted) == len(Action_value)-1 && cr.ActionsCompleted[action-1]
}

// HasClaimedAll returns true if the user has claimed all the rewards from the
// available actions
func (cr ClaimRecord) HasClaimedAll() bool {
	if len(cr.ActionsCompleted) == 0 {
		return false
	}
	for _, completed := range cr.ActionsCompleted {
		if !completed {
			return false
		}
	}
	return true
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

	if cra.InitialClaimableAmount.IsNil() {
		return errors.New("initial claimable amount is nil")
	}

	if !cra.InitialClaimableAmount.IsPositive() {
		return fmt.Errorf("initial claimable amount is not positive, %s", cra.InitialClaimableAmount)
	}

	if len(Action_value)-1 != len(cra.ActionsCompleted) {
		return fmt.Errorf("action length mismatch, expected %d, got %d", len(Action_value)-1, len(cra.ActionsCompleted))
	}

	return nil
}
