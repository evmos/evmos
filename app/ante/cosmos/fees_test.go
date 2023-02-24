package cosmos_test

import (
	"cosmossdk.io/math"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosante "github.com/evmos/evmos/v11/app/ante/cosmos"
	"github.com/evmos/evmos/v11/testutil"
	testutiltx "github.com/evmos/evmos/v11/testutil/tx"
)

func (suite *AnteTestSuite) TestDeductFeeDecorator() {
	// General setup
	addr, priv := testutiltx.NewAccAddressAndKey()

	// Testcase definitions
	testcases := []struct {
		name        string
		balance     math.Int
		rewards     math.Int
		gas         uint64
		checkTx     bool
		simulate    bool
		expPass     bool
		errContains string
	}{
		{
			name:        "pass - sufficient balance to pay fees",
			balance:     sdk.NewInt(1e18),
			rewards:     sdk.NewInt(0),
			gas:         0,
			checkTx:     false,
			simulate:    true,
			expPass:     true,
			errContains: "",
		},
		{
			name:        "fail - zero gas limit in check tx mode",
			balance:     sdk.NewInt(1e18),
			rewards:     sdk.NewInt(0),
			gas:         0,
			checkTx:     true,
			simulate:    false,
			expPass:     false,
			errContains: "must provide positive gas",
		},
		{
			name:        "fail - checkTx - insufficient funds and no staking rewards",
			balance:     sdk.NewInt(0),
			rewards:     sdk.NewInt(0),
			gas:         10_000_000,
			checkTx:     true,
			simulate:    false,
			expPass:     false,
			errContains: "insufficient funds and failed to claim sufficient staking rewards",
		},
		{
			name:        "pass - insufficient funds but sufficient staking rewards",
			balance:     sdk.NewInt(1e18),
			rewards:     sdk.NewInt(1e18),
			gas:         10_000_000,
			checkTx:     false,
			simulate:    false,
			expPass:     true,
			errContains: "",
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
		},
	}

	// Test execution
	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			suite.ctx = suite.ctx.WithIsCheckTx(tc.checkTx)

			// Create a new DeductFeeDecorator
			dfd := cosmosante.NewDeductFeeDecorator(
				suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.DistrKeeper, suite.app.FeeGrantKeeper, suite.app.StakingKeeper, nil,
			)

			// prepare the testcase
			ctx, err := testutil.PrepareAccountsForDelegationRewards(suite.T(), suite.ctx, suite.app, addr, tc.balance, tc.rewards)
			suite.Require().NoError(err, "failed to prepare accounts for delegation rewards")
			suite.ctx = ctx

			// Create an arbitrary message for testing purposes
			msg := sdktestutil.NewTestMsg(addr)

			// Set up the transaction arguments
			args := testutiltx.CosmosTxArgs{
				TxCfg: suite.clientCtx.TxConfig,
				Priv:  priv,
				Gas:   tc.gas,
				Msgs:  []sdk.Msg{msg},
			}

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
		})
	}
}
