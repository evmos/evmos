package keeper_test

import (
	"math/big"
	"strings"
	"time"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"

	evmante "github.com/evmos/evmos/v19/app/ante/evm"
	"github.com/evmos/evmos/v19/contracts"
	"github.com/evmos/evmos/v19/testutil"
	"github.com/evmos/evmos/v19/x/vesting/types"
)

// Commit commits and starts a new block with an updated context.
func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	var err error
	suite.ctx, err = testutil.CommitAndCreateNewCtx(suite.ctx, suite.app, t, nil)
	suite.Require().NoError(err)
}

// MintFeeCollector mints coins with the bank modules and sends them to the fee
// collector.
func (suite *KeeperTestSuite) MintFeeCollector(coins sdk.Coins) {
	err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, types.ModuleName, authtypes.FeeCollectorName, coins)
	suite.Require().NoError(err)
}

// DeployContract deploys the ERC20MinterBurnerDecimalsContract.
func (suite *KeeperTestSuite) DeployContract(
	name, symbol string,
	decimals uint8,
) (common.Address, error) {
	suite.Commit()
	addr, err := testutil.DeployContract(
		suite.ctx,
		suite.app,
		suite.priv,
		suite.queryClientEvm,
		contracts.ERC20MinterBurnerDecimalsContract,
		name, symbol, decimals,
	)
	suite.Commit()
	return addr, err
}

// assertEthFails is a helper function that takes in 1 or more messages and checks
// that they can neither be validated nor delivered using the EthVesting.
func assertEthFails(msgs ...sdk.Msg) {
	insufficientUnlocked := "insufficient unlocked"

	err := validateEthVestingTransactionDecorator(msgs...)
	Expect(err).ToNot(BeNil())
	Expect(strings.Contains(err.Error(), insufficientUnlocked))

	// Sanity check that delivery fails as well
	_, err = testutil.DeliverEthTx(s.ctx, s.app, nil, msgs...)
	Expect(err).ToNot(BeNil())
	Expect(strings.Contains(err.Error(), insufficientUnlocked))
}

// assertEthSucceeds is a helper function, that checks if 1 or more messages
// can be validated and delivered.
func assertEthSucceeds(testAccounts []TestClawbackAccount, funder sdk.AccAddress, dest sdk.AccAddress, amount sdkmath.Int, denom string, msgs ...sdk.Msg) {
	numTestAccounts := len(testAccounts)

	// Track starting balances for all accounts
	granteeBalances := make(sdk.Coins, numTestAccounts)
	funderBalance := s.app.BankKeeper.GetBalance(s.ctx, funder, denom)
	destBalance := s.app.BankKeeper.GetBalance(s.ctx, dest, denom)

	for i, grantee := range testAccounts {
		granteeBalances[i] = s.app.BankKeeper.GetBalance(s.ctx, grantee.address, denom)
	}

	// Validate the AnteHandler passes without issue
	err := validateEthVestingTransactionDecorator(msgs...)
	Expect(err).To(BeNil())

	// Expect delivery to succeed, then compare balances
	_, err = testutil.DeliverEthTx(s.ctx, s.app, nil, msgs...)
	Expect(err).To(BeNil())

	fb := s.app.BankKeeper.GetBalance(s.ctx, funder, denom)
	db := s.app.BankKeeper.GetBalance(s.ctx, dest, denom)

	s.Require().Equal(funderBalance, fb)
	newAmount := sdk.NewIntFromBigInt(convert18To6Decimals(*amount.BigInt()))
	s.Require().Equal(destBalance.AddAmount(newAmount).Amount.Mul(sdkmath.NewInt(int64(numTestAccounts))), db.Amount)

	for i, account := range testAccounts {
		gb := s.app.BankKeeper.GetBalance(s.ctx, account.address, denom)
		// Use GreaterOrEqual because the gas fee is non-recoverable
		// fmt.Println("balance", gb.Amount, granteeBalances[i].SubAmount(newAmount).Amount)
		s.Require().GreaterOrEqual(granteeBalances[i].SubAmount(newAmount).Amount.Uint64(), gb.Amount.Uint64())
	}
}

// validateEthVestingTransactionDecorator is a helper function to execute the eth vesting transaction decorator
// with 1 or more given messages and return any occurring error.
func validateEthVestingTransactionDecorator(msgs ...sdk.Msg) error {
	dec := evmante.NewEthVestingTransactionDecorator(s.app.AccountKeeper, s.app.BankKeeper, s.app.EvmKeeper)
	err = testutil.ValidateAnteForMsgs(s.ctx, dec, msgs...)
	return err
}

// convert18To6Decimals converts a big.Int to 6 decimals from 18
func convert18To6Decimals(amount big.Int) *big.Int {
	return new(big.Int).Div(&amount, big.NewInt(1e12))
}
