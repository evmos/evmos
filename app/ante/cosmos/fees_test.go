package cosmos_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	cosmosante "github.com/evmos/evmos/v12/app/ante/cosmos"
	"github.com/evmos/evmos/v12/testutil"
	testutiltx "github.com/evmos/evmos/v12/testutil/tx"
	"github.com/evmos/evmos/v12/utils"
)

func (suite *AnteTestSuite) TestDeductFeeDecorator() {
	var (
		dfd cosmosante.DeductFeeDecorator
		// General setup
		addr, priv = testutiltx.NewAccAddressAndKey()
		// fee granter
		fgAddr, _   = testutiltx.NewAccAddressAndKey()
		initBalance = sdk.NewInt(1e18)
		lowGasPrice = math.NewInt(1)
		zero        = sdk.ZeroInt()
	)

	// Testcase definitions
	testcases := []struct {
		name        string
		balance     math.Int
		rewards     math.Int
		gas         uint64
		gasPrice    *math.Int
		feeGranter  sdk.AccAddress
		checkTx     bool
		simulate    bool
		expPass     bool
		errContains string
		postCheck   func()
		malleate    func()
	}{
		{
			name:        "pass - sufficient balance to pay fees",
			balance:     initBalance,
			rewards:     zero,
			gas:         0,
			checkTx:     false,
			simulate:    true,
			expPass:     true,
			errContains: "",
		},
		{
			name:        "fail - zero gas limit in check tx mode",
			balance:     initBalance,
			rewards:     zero,
			gas:         0,
			checkTx:     true,
			simulate:    false,
			expPass:     false,
			errContains: "must provide positive gas",
		},
		{
			name:        "fail - checkTx - insufficient funds and no staking rewards",
			balance:     zero,
			rewards:     zero,
			gas:         10_000_000,
			checkTx:     true,
			simulate:    false,
			expPass:     false,
			errContains: "insufficient funds and failed to claim sufficient staking rewards",
			postCheck: func() {
				// the balance should not have changed
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, utils.BaseDenom)
				suite.Require().Equal(zero, balance.Amount, "expected balance to be zero")

				// there should be no rewards
				rewards, err := testutil.GetTotalDelegationRewards(suite.ctx, suite.app.DistrKeeper, addr)
				suite.Require().NoError(err, "failed to get total delegation rewards")
				suite.Require().Empty(rewards, "expected rewards to be zero")
			},
		},
		{
			name:        "pass - insufficient funds but sufficient staking rewards",
			balance:     zero,
			rewards:     initBalance,
			gas:         10_000_000,
			checkTx:     false,
			simulate:    false,
			expPass:     true,
			errContains: "",
			postCheck: func() {
				// the balance should have increased
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, utils.BaseDenom)
				suite.Require().False(
					balance.Amount.IsZero(),
					"expected balance to have increased after withdrawing a surplus amount of staking rewards",
				)

				// the rewards should all have been withdrawn
				rewards, err := testutil.GetTotalDelegationRewards(suite.ctx, suite.app.DistrKeeper, addr)
				suite.Require().NoError(err, "failed to get total delegation rewards")
				suite.Require().Empty(rewards, "expected all rewards to be withdrawn")
			},
		},
		{
			name:        "fail - insufficient funds and insufficient staking rewards",
			balance:     sdk.NewInt(1e5),
			rewards:     sdk.NewInt(1e5),
			gas:         10_000_000,
			checkTx:     false,
			simulate:    false,
			expPass:     false,
			errContains: "insufficient funds and failed to claim sufficient staking rewards",
			postCheck: func() {
				// the balance should not have changed
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, utils.BaseDenom)
				suite.Require().Equal(sdk.NewInt(1e5), balance.Amount, "expected balance to be unchanged")

				// the rewards should not have changed
				rewards, err := testutil.GetTotalDelegationRewards(suite.ctx, suite.app.DistrKeeper, addr)
				suite.Require().NoError(err, "failed to get total delegation rewards")
				suite.Require().Equal(
					sdk.NewDecCoins(sdk.NewDecCoin(utils.BaseDenom, sdk.NewInt(1e5))),
					rewards,
					"expected rewards to be unchanged")
			},
		},
		{
			name:        "fail - sufficient balance to pay fees but provided fees < required fees",
			balance:     initBalance,
			rewards:     zero,
			gas:         10_000_000,
			gasPrice:    &lowGasPrice,
			checkTx:     true,
			simulate:    false,
			expPass:     false,
			errContains: "insufficient fees",
			malleate: func() {
				suite.ctx = suite.ctx.WithMinGasPrices(
					sdk.NewDecCoins(
						sdk.NewDecCoin(utils.BaseDenom, sdk.NewInt(10_000)),
					),
				)
			},
		},
		{
			name:        "success - sufficient balance to pay fees & min gas prices is zero",
			balance:     initBalance,
			rewards:     zero,
			gas:         10_000_000,
			gasPrice:    &lowGasPrice,
			checkTx:     true,
			simulate:    false,
			expPass:     true,
			errContains: "",
			malleate: func() {
				suite.ctx = suite.ctx.WithMinGasPrices(
					sdk.NewDecCoins(
						sdk.NewDecCoin(utils.BaseDenom, zero),
					),
				)
			},
		},
		{
			name:        "success - sufficient balance to pay fees (fees > required fees)",
			balance:     initBalance,
			rewards:     zero,
			gas:         10_000_000,
			checkTx:     true,
			simulate:    false,
			expPass:     true,
			errContains: "",
			malleate: func() {
				suite.ctx = suite.ctx.WithMinGasPrices(
					sdk.NewDecCoins(
						sdk.NewDecCoin(utils.BaseDenom, sdk.NewInt(100)),
					),
				)
			},
		},
		{
			name:        "success - zero fees",
			balance:     initBalance,
			rewards:     zero,
			gas:         100,
			gasPrice:    &zero,
			checkTx:     true,
			simulate:    false,
			expPass:     true,
			errContains: "",
			malleate: func() {
				suite.ctx = suite.ctx.WithMinGasPrices(
					sdk.NewDecCoins(
						sdk.NewDecCoin(utils.BaseDenom, zero),
					),
				)
			},
			postCheck: func() {
				// the tx sender balance should not have changed
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, utils.BaseDenom)
				suite.Require().Equal(initBalance, balance.Amount, "expected balance to be unchanged")
			},
		},
		{
			name:        "fail - with not authorized fee granter",
			balance:     initBalance,
			rewards:     zero,
			gas:         10_000_000,
			feeGranter:  fgAddr,
			checkTx:     true,
			simulate:    false,
			expPass:     false,
			errContains: fmt.Sprintf("%s does not not allow to pay fees for %s", fgAddr, addr),
		},
		{
			name:        "success - with authorized fee granter",
			balance:     initBalance,
			rewards:     zero,
			gas:         10_000_000,
			feeGranter:  fgAddr,
			checkTx:     true,
			simulate:    false,
			expPass:     true,
			errContains: "",
			malleate: func() {
				// Fund the fee granter
				err := testutil.FundAccountWithBaseDenom(suite.ctx, suite.app.BankKeeper, fgAddr, initBalance.Int64())
				suite.Require().NoError(err)
				// grant the fees
				grant := sdk.NewCoins(sdk.NewCoin(
					utils.BaseDenom, initBalance,
				))
				err = suite.app.FeeGrantKeeper.GrantAllowance(suite.ctx, fgAddr, addr, &feegrant.BasicAllowance{
					SpendLimit: grant,
				})
				suite.Require().NoError(err)
			},
			postCheck: func() {
				// the tx sender balance should not have changed
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, utils.BaseDenom)
				suite.Require().Equal(initBalance, balance.Amount, "expected balance to be unchanged")
			},
		},
		{
			name:        "fail - authorized fee granter but no feegrant keeper on decorator",
			balance:     initBalance,
			rewards:     zero,
			gas:         10_000_000,
			feeGranter:  fgAddr,
			checkTx:     true,
			simulate:    false,
			expPass:     false,
			errContains: "fee grants are not enabled",
			malleate: func() {
				// Fund the fee granter
				err := testutil.FundAccountWithBaseDenom(suite.ctx, suite.app.BankKeeper, fgAddr, initBalance.Int64())
				suite.Require().NoError(err)
				// grant the fees
				grant := sdk.NewCoins(sdk.NewCoin(
					utils.BaseDenom, initBalance,
				))
				err = suite.app.FeeGrantKeeper.GrantAllowance(suite.ctx, fgAddr, addr, &feegrant.BasicAllowance{
					SpendLimit: grant,
				})
				suite.Require().NoError(err)

				// remove the feegrant keeper from the decorator
				dfd = cosmosante.NewDeductFeeDecorator(
					suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.DistrKeeper, nil, suite.app.StakingKeeper, nil,
				)
			},
		},
	}

	// Test execution
	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// Create a new DeductFeeDecorator
			dfd = cosmosante.NewDeductFeeDecorator(
				suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.DistrKeeper, suite.app.FeeGrantKeeper, suite.app.StakingKeeper, nil,
			)

			// prepare the testcase
			var err error
			suite.ctx, err = testutil.PrepareAccountsForDelegationRewards(suite.T(), suite.ctx, suite.app, addr, tc.balance, tc.rewards)
			suite.Require().NoError(err, "failed to prepare accounts for delegation rewards")
			suite.ctx, err = testutil.Commit(suite.ctx, suite.app, time.Second*0, nil)
			suite.Require().NoError(err)

			// Create an arbitrary message for testing purposes
			msg := sdktestutil.NewTestMsg(addr)

			// Set up the transaction arguments
			args := testutiltx.CosmosTxArgs{
				TxCfg:      suite.clientCtx.TxConfig,
				Priv:       priv,
				Gas:        tc.gas,
				GasPrice:   tc.gasPrice,
				FeeGranter: tc.feeGranter,
				Msgs:       []sdk.Msg{msg},
			}

			if tc.malleate != nil {
				tc.malleate()
			}
			suite.ctx = suite.ctx.WithIsCheckTx(tc.checkTx)

			// Create a transaction out of the message
			tx, err := testutiltx.PrepareCosmosTx(suite.ctx, suite.app, args)
			suite.Require().NoError(err, "failed to create transaction")

			// run the ante handler
			_, err = dfd.AnteHandle(suite.ctx, tx, tc.simulate, testutil.NextFn)

			// assert the resulting error
			if tc.expPass {
				suite.Require().NoError(err, "expected no error")
			} else {
				suite.Require().Error(err, "expected error")
				suite.Require().ErrorContains(err, tc.errContains)
			}

			// run the post check
			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}
