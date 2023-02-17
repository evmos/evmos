package cosmos_test

import (
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	cosmosante "github.com/evmos/evmos/v11/app/ante/cosmos"
	"github.com/evmos/evmos/v11/testutil"
	testutiltx "github.com/evmos/evmos/v11/testutil/tx"
)

func (suite *AnteTestSuite) TestDeductFeeDecorator() {
	// General setup
	addr, priv := testutiltx.NewAccAddressAndKey()
	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()

	// Testcase definitions
	testcases := []struct {
		name        string
		malleate    func() signing.Tx
		checkTx     bool
		simulate    bool
		expPass     bool
		errContains string
	}{
		{
			name: "pass - zero gas limit in simulation mode",
			malleate: func() signing.Tx {
				// Generate new account
				err := testutil.FundAccountWithBaseDenom(suite.ctx, suite.app.BankKeeper, addr, 300)
				suite.Require().NoError(err, "failed to fund account")

				// Create an arbitrary message for testing purposes
				msg := sdktestutil.NewTestMsg(addr)

				// Set gas limit to zero for testing purposes
				txBuilder.SetGasLimit(0)

				// Create a transaction out of the message
				txBuilder, err = testutiltx.CreateTxInTxBuilder(suite.ctx, suite.app, txBuilder, priv, msg)
				suite.Require().NoError(err, "failed to create transaction")

				return txBuilder.GetTx()
			},
			checkTx:     false,
			simulate:    true,
			expPass:     true,
			errContains: "",
		},
		{
			name: "fail - zero gas limit in check tx mode",
			malleate: func() signing.Tx {
				// TODO: refactor this
				// Generate new account
				err := testutil.FundAccountWithBaseDenom(suite.ctx, suite.app.BankKeeper, addr, 300)
				suite.Require().NoError(err, "failed to fund account")

				// Create an arbitrary message for testing purposes
				msg := sdktestutil.NewTestMsg(addr)

				// Set gas limit to zero for testing purposes
				txBuilder.SetGasLimit(0)

				// Create a transaction out of the message
				txBuilder, err = testutiltx.CreateTxInTxBuilder(suite.ctx, suite.app, txBuilder, priv, msg)
				suite.Require().NoError(err, "failed to create transaction")

				return txBuilder.GetTx()
			},
			checkTx:     true,
			simulate:    false,
			expPass:     false,
			errContains: "must provide positive gas",
		},
		{
			name: "fail - checkTx - insufficient funds and no staking rewards",
			malleate: func() signing.Tx {
				suite.app.AccountKeeper.SetAccount(suite.ctx, suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr))

				// Create an arbitrary message for testing purposes
				msg := sdktestutil.NewTestMsg(addr)

				// Create a transaction out of the message
				tx, err := testutiltx.PrepareCosmosTx(suite.ctx, suite.clientCtx.TxConfig, suite.app, priv, nil, msg)
				suite.Require().NoError(err, "failed to create transaction")

				return tx
			},
			checkTx:     true,
			simulate:    false,
			expPass:     false,
			errContains: "insufficient funds",
		},
	}

	// Test execution
	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			dfd := cosmosante.NewDeductFeeDecorator(
				suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.DistrKeeper, suite.app.FeeGrantKeeper, suite.app.StakingKeeper, nil,
			)

			// set up the testcase
			tx := tc.malleate()
			suite.ctx = suite.ctx.WithIsCheckTx(tc.checkTx)

			// run the ante handler
			_, err := dfd.AnteHandle(suite.ctx, tx, tc.simulate, testutil.NextFn)

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
