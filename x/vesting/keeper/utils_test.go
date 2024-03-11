package keeper_test

// import (
// 	"strings"

// 	//nolint:revive // dot imports are fine for Ginkgo
// 	. "github.com/onsi/gomega"

// 	sdkmath "cosmossdk.io/math"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

// 	cosmosante "github.com/evmos/evmos/v16/app/ante/cosmos"
// 	evmante "github.com/evmos/evmos/v16/app/ante/evm"
// 	"github.com/evmos/evmos/v16/testutil"
// 	"github.com/evmos/evmos/v16/x/vesting/types"
// )

// // assertEthFails is a helper function that takes in 1 or more messages and checks
// // that they can neither be validated nor delivered using the EthVesting.
// func assertEthFails(msgs ...sdk.Msg) {
// 	insufficientUnlocked := "insufficient unlocked"

// 	err := validateEthVestingTransactionDecorator(msgs...)
// 	Expect(err).ToNot(BeNil())
// 	Expect(strings.Contains(err.Error(), insufficientUnlocked))

// 	// Sanity check that delivery fails as well
// 	_, err = testutil.DeliverEthTx(suite.app, nil, msgs...)
// 	Expect(err).ToNot(BeNil())
// 	Expect(strings.Contains(err.Error(), insufficientUnlocked))
// }

// assertEthSucceeds is a helper function, that checks if 1 or more messages
// can be validated and delivered.
func assertEthSucceeds(testAccounts []TestClawbackAccount, funder sdk.AccAddress, dest sdk.AccAddress, amount sdkmath.Int, denom string, msgs ...sdk.Msg) {
	numTestAccounts := len(testAccounts)

	// Track starting balances for all accounts
	granteeBalances := make(sdk.Coins, numTestAccounts)
	funderBalance := suite.app.BankKeeper.GetBalance(suite.ctx, funder, denom)
	destBalance := suite.app.BankKeeper.GetBalance(suite.ctx, dest, denom)

	for i, grantee := range testAccounts {
		granteeBalances[i] = suite.app.BankKeeper.GetBalance(suite.ctx, grantee.address, denom)
	}

	// Validate the AnteHandler passes without issue
	err := validateEthVestingTransactionDecorator(msgs...)
	Expect(err).To(BeNil())

	// Expect delivery to succeed, then compare balances
	_, err = testutil.DeliverEthTx(suite.app, nil, msgs...)
	Expect(err).To(BeNil())

	fb := suite.app.BankKeeper.GetBalance(suite.ctx, funder, denom)
	db := suite.app.BankKeeper.GetBalance(suite.ctx, dest, denom)

	s.Require().Equal(funderBalance, fb)
	s.Require().Equal(destBalance.AddAmount(amount).Amount.Mul(sdkmath.NewInt(int64(numTestAccounts))), db.Amount)

	for i, account := range testAccounts {
		gb := suite.app.BankKeeper.GetBalance(suite.ctx, account.address, denom)
		// Use GreaterOrEqual because the gas fee is non-recoverable
		s.Require().GreaterOrEqual(granteeBalances[i].SubAmount(amount).Amount.Uint64(), gb.Amount.Uint64())
	}
}

// // delegate is a helper function which creates a message to delegate a given amount of tokens
// // to a validator and checks if the Cosmos vesting delegation decorator returns no error.
// func delegate(account TestClawbackAccount, coins sdk.Coins) (*stakingtypes.MsgDelegate, error) {
// 	msg := stakingtypes.NewMsgDelegate(account.address.String(), s.validator.GetOperator(), coins[0])
// 	dec := cosmosante.NewVestingDelegationDecorator(suite.app.AccountKeeper, suite.app.StakingKeeper, suite.app.BankKeeper, types.ModuleCdc)
// 	err = testutil.ValidateAnteForMsgs(suite.ctx, dec, msg)
// 	return msg, err
// }

// // validateEthVestingTransactionDecorator is a helper function to execute the eth vesting transaction decorator
// // with 1 or more given messages and return any occurring error.
// func validateEthVestingTransactionDecorator(msgs ...sdk.Msg) error {
// 	dec := evmante.NewEthVestingTransactionDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.EvmKeeper)
// 	err = testutil.ValidateAnteForMsgs(suite.ctx, dec, msgs...)
// 	return err
// }
