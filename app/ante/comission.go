package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var minCommission = sdk.NewDecWithPrec(5, 2) // 5%

// TODO: remove once Cosmos SDK is upgraded to v0.46

// ValidatorCommissionDecorator validates that the validator commission is always
// greater or equal than the min commission rate
type ValidatorCommissionDecorator struct{}

// NewValidatorCommissionDecorator creates a new NewValidatorCommissionDecorator
func NewValidatorCommissionDecorator() ValidatorCommissionDecorator {
	return ValidatorCommissionDecorator{}
}

// AnteHandle checks if the tx contains a staking create validator or edit validator.
// It errors if the the commission rate is below the min threshold.
func (vcd ValidatorCommissionDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		switch msg := msg.(type) {
		case *stakingtypes.MsgCreateValidator:
			if msg.Commission.Rate.LT(minCommission) {
				return ctx, sdkerrors.Wrapf(
					sdkerrors.ErrInvalidRequest,
					"validator commission %s be lower than minimum of %s", msg.Commission.Rate, minCommission)
			}
		case *stakingtypes.MsgEditValidator:
			if msg.CommissionRate != nil && msg.CommissionRate.LT(minCommission) {
				return ctx, sdkerrors.Wrapf(
					sdkerrors.ErrInvalidRequest,
					"validator commission %s be lower than minimum of %s", msg.CommissionRate, minCommission)
			}
		default:
			continue
		}
	}

	return next(ctx, tx, simulate)
}
