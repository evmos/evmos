package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	vestingtypes "github.com/tharsis/evmos/x/vesting/types"
)

// EthVestingTransactionDecorator validates if clawback vesting accounts are
// permitted to perform Ethereum Tx.
type EthVestingTransactionDecorator struct {
	ak evmtypes.AccountKeeper
}

func NewEthVestingTransactionDecorator(ak evmtypes.AccountKeeper) EthVestingTransactionDecorator {
	return EthVestingTransactionDecorator{
		ak: ak,
	}
}

// AnteHandle validates that a clawback vesting account has surpassed the
// vesting cliff and lockup period.
//
// This AnteHandler decorator will fail if:
//  - the message is not a MsgEthereumTx
//  - sender account cannot be found
//  - sender account is not a ClawbackvestingAccount
//  - blocktime is before surpassing vesting cliff end (with zero vested coins) AND
//  - blocktime is before surpassing all lockup periods (with non-zero locked coins)
func (vtd EthVestingTransactionDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest,
				"invalid message type %T, expected %T", msg, (*evmtypes.MsgEthereumTx)(nil),
			)
		}

		acc := vtd.ak.GetAccount(ctx, msgEthTx.GetFrom())
		if acc == nil {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress,
				"account %s does not exist", acc)
		}

		// Check that this decorator only applies to clawback vesting accounts
		clawbackAccount, isClawback := acc.(*vestingtypes.ClawbackVestingAccount)
		if !isClawback {
			return next(ctx, tx, simulate)
		}

		// Error if vesting cliff has not passed (with zero vested coins). This
		// rule does not apply for existing clawback accounts that receive a new
		// grant while there are already vested coins on the account.
		vested := clawbackAccount.GetVestedCoins(ctx.BlockTime())
		if len(vested) == 0 {
			return ctx, sdkerrors.Wrapf(vestingtypes.ErrInsufficientVestedCoins,
				"cannot perform Ethereum tx with clawback vesting account, that has no vested coins: %s", vested,
			)
		}

		// Error if account has locked coins (before surpassing all lockup periods)
		islocked := clawbackAccount.HasLockedCoins(ctx.BlockTime())
		if islocked {
			return ctx, sdkerrors.Wrapf(vestingtypes.ErrVestingLockup,
				"cannot perform Ethereum tx with clawback vesting account, that has locked coins: %s", vested,
			)
		}
	}

	return next(ctx, tx, simulate)
}

// VestingDelegationDecorator validates delegation of vested coins
type VestingDelegationDecorator struct {
	ak evmtypes.AccountKeeper
	sk vestingtypes.StakingKeeper
}

// NewVestingDelegationDecorator creates a new VestingDelegationDecorator
func NewVestingDelegationDecorator(ak evmtypes.AccountKeeper, sk vestingtypes.StakingKeeper) VestingDelegationDecorator {
	return VestingDelegationDecorator{
		ak: ak,
		sk: sk,
	}
}

// AnteHandle checks if the tx contains a staking delegation.
// It errors if the coins are still locked or the bond amount is greater than
// the coins already vested
func (vdd VestingDelegationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		// Continue only if delegation
		delegateMsg, isDelegation := msg.(*stakingtypes.MsgDelegate)
		if !isDelegation {
			continue
		}

		for _, addr := range msg.GetSigners() {

			acc := vdd.ak.GetAccount(ctx, addr)
			if acc == nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress,
					"account %s does not exist", addr,
				)
			}

			clawbackAccount, isPeriodicVesting := acc.(*vestingtypes.ClawbackVestingAccount)
			if !isPeriodicVesting {
				// continue to next decorator as this logic only applies to vesting
				return next(ctx, tx, simulate)
			}

			// error if bond amount is > vested coins
			coins := clawbackAccount.GetVestedOnly(ctx.BlockHeader().Time)
			vested := coins.AmountOf(vdd.sk.BondDenom(ctx))
			if vested.LT(delegateMsg.Amount.Amount) {
				return ctx, sdkerrors.Wrapf(vestingtypes.ErrInsufficientVestedCoins,
					"cannot delegate unvested coins. Coins Vested: %x", vested,
				)
			}
		}
	}

	return next(ctx, tx, simulate)
}
