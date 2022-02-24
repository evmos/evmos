package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	vestingtypes "github.com/tharsis/evmos/x/vesting/types"
)

// VestingDelegationDecorator validates delegation of vested coins
type VestingDelegationDecorator struct {
	ak evmtypes.AccountKeeper
}

// NewVestingDelegationDecorator creates a new VestingDelegationDecorator
func NewVestingDelegationDecorator(ak evmtypes.AccountKeeper) VestingDelegationDecorator {
	return VestingDelegationDecorator{
		ak: ak,
	}
}

// AnteHandle checks if the tx contains a staking delegation
// It error if the coins are still locked or the bond amount is greater than the coins already vested
func (vdd VestingDelegationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	//
	for _, msg := range tx.GetMsgs() {
		for _, addr := range msg.GetSigners() {

			// Continue only if delegation
			delegateMsg, isDelegation := msg.(*stakingtypes.MsgDelegate)
			if !isDelegation {
				continue
			}

			acc := vdd.ak.GetAccount(ctx, addr)
			if acc == nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
			}

			clawbackAccount, isPeriodicVesting := acc.(*vestingtypes.ClawbackVestingAccount)
			if !isPeriodicVesting {
				// continue to next decorator as this logic only applies to vesting
				return next(ctx, tx, simulate)
			}

			// error if bond amount is > vested coins
			coins := clawbackAccount.GetVestedOnly(ctx.BlockHeader().Time)
			vested := coins.AmountOf(stakingtypes.DefaultParams().BondDenom)
			if vested.LT(delegateMsg.Amount.Amount) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest,
					"cannot delegate unvested coins. Coins Vested: %x", vested,
				)
			}
		}
	}

	return next(ctx, tx, simulate)
}

// VestingGovernanceDecorator prevents voting with unvested coins
type VestingGovernanceDecorator struct {
	ak evmtypes.AccountKeeper
}

// NewVestingGovernanceDecorator creates a new VestingGovernanceDecorator
func NewVestingGovernanceDecorator(ak evmtypes.AccountKeeper) VestingGovernanceDecorator {
	return VestingGovernanceDecorator{
		ak: ak,
	}
}

// AnteHandle checks if the tx contains a vote for a gov proposal
// It error if there are no coins vested to the voter
func (vdd VestingGovernanceDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		for _, addr := range msg.GetSigners() {

			// Continue only if Vote
			_, isVote := msg.(*govtypes.MsgVote)
			if !isVote {
				continue
			}

			acc := vdd.ak.GetAccount(ctx, addr)
			if acc == nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
			}

			clawbackAccount, isPeriodicVesting := acc.(*vestingtypes.ClawbackVestingAccount)
			if !isPeriodicVesting {
				// continue to next decorator as this logic only applies to vesting
				return next(ctx, tx, simulate)
			}

			// error no vested coins
			coins := clawbackAccount.GetVestedOnly(ctx.BlockHeader().Time)
			vested := coins.AmountOf(stakingtypes.DefaultParams().BondDenom)
			if vested.LTE(sdk.ZeroInt()) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest,
					"cannot vote on proposals with a clawback vesting account with no vested coins",
				)
			}
		}
	}
	return next(ctx, tx, simulate)
}
