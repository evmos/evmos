package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// DeductFeeDecorator deducts fees from the first signer of the tx
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next AnteHandler if fees successfully deducted
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
type DeductFeeDecorator struct {
	ak             ante.AccountKeeper
	bankKeeper     types.BankKeeper
	feegrantKeeper ante.FeegrantKeeper
	feesplitKeeper FeesplitKeeper
}

func NewDeductFeeDecorator(
	ak ante.AccountKeeper,
	bk types.BankKeeper,
	fk ante.FeegrantKeeper,
	fsk FeesplitKeeper,
) DeductFeeDecorator {
	return DeductFeeDecorator{
		ak:             ak,
		bankKeeper:     bk,
		feegrantKeeper: fk,
		feesplitKeeper: fsk,
	}
}

func (dfd DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if addr := dfd.ak.GetModuleAddress(types.FeeCollectorName); addr == nil {
		return ctx, fmt.Errorf("Fee collector module account (%s) has not been set", types.FeeCollectorName)
	}

	fees := feeTx.GetFee()
	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()

	deductFeesFrom := feePayer

	// if feegranter set deduct fee from feegranter account.
	// this works with only when feegrant enabled.
	if feeGranter != nil {
		if dfd.feegrantKeeper == nil {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fees, tx.GetMsgs())
			if err != nil {
				return ctx, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := dfd.ak.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", deductFeesFrom)
	}

	ibcMsgs := int64(0)
	params := dfd.feesplitKeeper.GetParams(ctx)

	// only iterate over messages if the discount is not 0%
	if !params.FeeDiscount.Equal(sdk.ZeroDec()) {
		// iterate over messages to calculate the total discount on fees
		for _, msg := range tx.GetMsgs() {
			// increment counter for eligible transactions.
			url := sdk.MsgTypeURL(msg)
			if params.IsEligibleMsg(url) {
				ibcMsgs++
			}
		}

		// discount is 50% * (IBC Msgs / total number of messages )
		totalFeeDiscount := params.FeeDiscount.MulInt64(ibcMsgs).QuoInt64(int64(len(tx.GetMsgs())))

		for _, fee := range fees {
			discount := fee.Amount.ToDec().Mul(totalFeeDiscount).TruncateInt()
			fee.Amount = fee.Amount.Sub(discount)
		}
	}

	// deduct the fees
	if err := DeductFees(dfd.bankKeeper, ctx, deductFeesFromAcc, fees); err != nil {
		return ctx, err
	}

	events := sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeTx, sdk.NewAttribute(sdk.AttributeKeyFee, fees.String()),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return next(ctx, tx, simulate)
}

// DeductFees deducts fees from the given account.
func DeductFees(bankKeeper types.BankKeeper, ctx sdk.Context, acc types.AccountI, fees sdk.Coins) error {
	if !fees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, fees)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}
