package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewClaimsRecord creates a new claim record instance
func NewClaimsRecord(initialClaimableAmt sdk.Int) ClaimsRecord {
	return ClaimsRecord{
		InitialClaimableAmount: initialClaimableAmt,
		ActionsCompleted:       []bool{false, false, false, false},
	}
}

// Validate performs a stateless validation of the fields
func (cr ClaimsRecord) Validate() error {
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

// ClaimAction marks the given action as completed. It performs a no-op if the
// action is invalid or if the ActionsCompleted slice has an invalid length.
func (cr *ClaimsRecord) ClaimAction(action Action) {
	switch {
	case len(cr.ActionsCompleted) != len(Action_value)-1:
		return
	case action == ActionUnspecified || int(action) > len(Action_value)-1:
		return
	default:
		cr.ActionsCompleted[action-1] = true
	}
}

// HasClaimedAction checks if the user has claimed a given action. It also
// returns false if the action is invalid or if the ActionsCompleted slice has
// an invalid length.
func (cr ClaimsRecord) HasClaimedAction(action Action) bool {
	switch {
	case len(cr.ActionsCompleted) != len(Action_value)-1:
		return false
	case action == 0 || int(action) > len(Action_value)-1:
		return false
	default:
		return cr.ActionsCompleted[action-1]
	}
}

// HasClaimedAny returns true if the user has claimed at least one reward from the
// available actions
func (cr ClaimsRecord) HasClaimedAny() bool {
	for _, completed := range cr.ActionsCompleted {
		if completed {
			return true
		}
	}
	return false
}

// HasClaimedAll returns true if the user has claimed all the rewards from the
// available actions
func (cr ClaimsRecord) HasClaimedAll() bool {
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

// NewClaimsRecordAddress creates a new claim record instance
func NewClaimsRecordAddress(address sdk.AccAddress, initialClaimableAmt sdk.Int) ClaimsRecordAddress {
	return ClaimsRecordAddress{
		Address:                address.String(),
		InitialClaimableAmount: initialClaimableAmt,
		ActionsCompleted:       []bool{false, false, false, false},
	}
}

// Validate performs a stateless validation of the fields
func (cra ClaimsRecordAddress) Validate() error {
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
