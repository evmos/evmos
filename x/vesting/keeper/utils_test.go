package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v11/testutil"

	cosmosante "github.com/evmos/evmos/v11/app/ante/cosmos"
	evmante "github.com/evmos/evmos/v11/app/ante/evm"
	"github.com/evmos/evmos/v11/utils"
	"github.com/evmos/evmos/v11/x/vesting/types"
)

// delegate is a helper function which creates a message to delegate a given amount of tokens
// to a validator and checks if the Cosmos vesting delegation decorator returns no error.
func delegate(clawbackAccount *types.ClawbackVestingAccount, amount math.Int) error {
	addr, err := sdk.AccAddressFromBech32(clawbackAccount.Address)
	s.Require().NoError(err)

	val, err := sdk.ValAddressFromBech32("evmosvaloper1z3t55m0l9h0eupuz3dp5t5cypyv674jjn4d6nn")
	s.Require().NoError(err)
	delegateMsg := stakingtypes.NewMsgDelegate(addr, val, sdk.NewCoin(utils.BaseDenom, amount))

	dec := cosmosante.NewVestingDelegationDecorator(s.app.AccountKeeper, s.app.StakingKeeper, types.ModuleCdc)
	err = testutil.ValidateAnteForMsgs(s.ctx, dec, delegateMsg)
	return err
}

// validateEthVestingTransactionDecorator is a helper function to execute the eth vesting transaction decorator
// with 1 or more given messages and return any occurring error.
func validateEthVestingTransactionDecorator(msgs ...sdk.Msg) error {
	dec := evmante.NewEthVestingTransactionDecorator(s.app.AccountKeeper, s.app.BankKeeper, s.app.EvmKeeper)
	err = testutil.ValidateAnteForMsgs(s.ctx, dec, msgs...)
	return err
}
