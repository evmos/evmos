package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	vestingtypes "github.com/tharsis/evmos/x/vesting/types"
)

type VestingDelegationDecorator struct {
	ak evmtypes.AccountKeeper
}

func NewVestingDelegationDecorator(ak evmtypes.AccountKeeper) VestingDelegationDecorator {
	return VestingDelegationDecorator{
		ak: ak,
	}
}

func (vdd VestingDelegationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {

	// check if the tx contains a staking delegation and error if the tokens are still locked or the bond amount is greater than the tokens already vested
	for _, msg := range tx.GetMsgs() {
		for _, addr := range msg.GetSigners() {
			acc := vdd.ak.GetAccount(ctx, addr)
			if acc == nil {
				return ctx, err
			}

			clawbackAccount, isPeriodicVesting := acc.(*vestingtypes.ClawbackVestingAccount)
			if !isPeriodicVesting {
				// continue to next decorator as this logic only applies to vesting
				return next(ctx, tx, simulate)
			}

			delegateMsg, isDelegation := msg.(*stakingtypes.MsgDelegate)
			if !isDelegation {
				continue
			}

			// error if bond amount is > vested tokens
			coins := clawbackAccount.GetVestedCoins(ctx.BlockHeader().Time)
			vested := coins.AmountOf(delegateMsg.Amount.Denom)
			if vested.LT(delegateMsg.Amount.Amount) {
				// TODO Define error message
				return ctx, fmt.Errorf("coins are locked")
			}

		}

	}

	return next(ctx, tx, simulate)
}
