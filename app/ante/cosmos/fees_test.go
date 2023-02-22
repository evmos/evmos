package cosmos_test

import (
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
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
		malleate    func() signing.Tx
		checkTx     bool
		simulate    bool
		expPass     bool
		errContains string
	}{
		{
			name: "pass - sufficient balance to pay fees",
			malleate: func() signing.Tx {
				// Generate new account
				err := testutil.FundAccountWithBaseDenom(suite.ctx, suite.app.BankKeeper, addr, 1e18)
				suite.Require().NoError(err, "failed to fund account")

				// Create an arbitrary message for testing purposes
				msg := sdktestutil.NewTestMsg(addr)

				msgArgs := testutiltx.CosmosTxArgs{
					TxCfg: suite.clientCtx.TxConfig,
					Priv:  priv,
					Gas:   0,
					Fees:  sdk.Coins{sdk.Coin{"", sdk.ZeroInt()}},
					Msgs:  []sdk.Msg{msg},
				}

				// Create a transaction out of the message
				tx, err := testutiltx.PrepareCosmosTx(suite.ctx, suite.app, msgArgs)
				suite.Require().NoError(err, "failed to create transaction")

				return tx
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
				msgArgs := testutiltx.CosmosTxArgs{
					TxCfg:    suite.clientCtx.TxConfig,
					Priv:     priv,
					Gas:      0,
					GasPrice: nil,
					Fees:     sdk.Coins{sdk.Coin{"", sdk.ZeroInt()}},
					Msgs:     []sdk.Msg{msg},
				}

				// Create a transaction out of the message
				tx, err := testutiltx.PrepareCosmosTx(suite.ctx, suite.app, msgArgs)
				suite.Require().NoError(err, "failed to create transaction")

				return tx
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
				msgArgs := testutiltx.CosmosTxArgs{
					TxCfg: suite.clientCtx.TxConfig,
					Priv:  priv,
					Gas:   10_000_000,
					Fees:  sdk.Coins{sdk.Coin{"", sdk.ZeroInt()}},
					Msgs:  []sdk.Msg{msg},
				}
				tx, err := testutiltx.PrepareCosmosTx(suite.ctx, suite.app, msgArgs)
				suite.Require().NoError(err, "failed to create transaction")

				return tx
			},
			checkTx:     true,
			simulate:    false,
			expPass:     false,
			errContains: "insufficient funds and failed to claim sufficient staking rewards",
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
