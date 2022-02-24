package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*evmtypes.MsgEthereumTx)(nil))
		}

		acc := vtd.ak.GetAccount(ctx, msgEthTx.GetFrom())
		if acc == nil {
			return ctx, err
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
		if vested == nil {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest,
				"cannot perform Ethereum tx with clawback vesting account, that has no vested coins: %x", vested,
			)
		}

		// Error if account has locked coins (before surpassing all lockup periods)
		islocked := clawbackAccount.HasLockedCoins(ctx.BlockTime())
		if islocked {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest,
				"cannot perform Ethereum tx with clawback vesting account, that has locked coins: %x", vested,
			)
		}
	}

	return next(ctx, tx, simulate)
}
