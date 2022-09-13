package ante

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var minCommission = sdk.NewDecWithPrec(5, 2) // 5%

// TODO: remove once Cosmos SDK is upgraded to v0.46

// ValidatorCommissionDecorator validates that the validator commission is always
// greater or equal than the min commission rate
type ValidatorCommissionDecorator struct {
	cdc codec.BinaryCodec
}

// NewValidatorCommissionDecorator creates a new NewValidatorCommissionDecorator
func NewValidatorCommissionDecorator(cdc codec.BinaryCodec) ValidatorCommissionDecorator {
	return ValidatorCommissionDecorator{
		cdc: cdc,
	}
}

// AnteHandle checks if the tx contains a staking create validator or edit validator.
// It errors if the the commission rate is below the min threshold.
func (vcd ValidatorCommissionDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		switch msg := msg.(type) {
		case *authz.MsgExec:
			// Check for bypassing authorization
			if err := vcd.validateAuthz(ctx, msg); err != nil {
				return ctx, err
			}

		default:
			if err := vcd.validateMsg(ctx, msg); err != nil {
				return ctx, err
			}
		}
	}

	return next(ctx, tx, simulate)
}

// validateAuthz validates the authorization internal message
func (vcd ValidatorCommissionDecorator) validateAuthz(ctx sdk.Context, execMsg *authz.MsgExec) error {
	for _, v := range execMsg.Msgs {
		var innerMsg sdk.Msg
		err := vcd.cdc.UnpackAny(v, &innerMsg)
		if err != nil {
			return sdkerrors.Wrap(err, "cannot unmarshal authz exec msgs")
		}

		if err := vcd.validateMsg(ctx, innerMsg); err != nil {
			return err
		}
	}

	return nil
}

// validateMsg checks that the commission rate is over 5% for create and edit validator msgs
func (vcd ValidatorCommissionDecorator) validateMsg(_ sdk.Context, msg sdk.Msg) error {
	switch msg := msg.(type) {
	case *stakingtypes.MsgCreateValidator:
		if msg.Commission.Rate.LT(minCommission) {
			return sdkerrors.Wrapf(
				errortypes.ErrInvalidRequest,
				"validator commission %s be lower than minimum of %s", msg.Commission.Rate, minCommission)
		}
	case *stakingtypes.MsgEditValidator:
		if msg.CommissionRate != nil && msg.CommissionRate.LT(minCommission) {
			return sdkerrors.Wrapf(
				errortypes.ErrInvalidRequest,
				"validator commission %s be lower than minimum of %s", msg.CommissionRate, minCommission)
		}
	}
	return nil
}
