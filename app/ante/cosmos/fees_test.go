package cosmos_test

import (
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/evmos/evmos/v11/testutil"
	"github.com/evmos/evmos/v11/utils"
)

func (suite *AnteTestSuite) TestDeductFeeDecorator_ZeroGas() {
	suite.SetupTest()

	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()

	dfd := ante.NewDeductFeeDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.FeeGrantKeeper, nil)
	antehandler := sdk.ChainAnteDecorators(dfd)

	// Generate new account
	addr, priv := testutil.NewAccAddressAndKey()
	coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, sdk.NewInt(300)))
	err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr, coins)
	suite.Require().NoError(err, "failed to fund account")

	// Create an arbitrary message for testing purposes
	msg := sdktestutil.NewTestMsg(addr)
	err = txBuilder.SetMsgs(msg)
	suite.Require().NoError(err, "failed to set message")

	// Set gas limit to zero for testing purposes
	txBuilder.SetGasLimit(0)

	// Create a transaction out of the message
	txBuilder, err = testutil.CreateTxInTxBuilder(suite.ctx, suite.app, txBuilder, priv, msg)
	suite.Require().NoError(err, "failed to create transaction")

	// Set IsCheckTx to true and check if the antehandler returns an error
	// because a gas limit of zero should not be accepted
	suite.ctx = suite.ctx.WithIsCheckTx(true)
	_, err = antehandler(suite.ctx, txBuilder.GetTx(), false)
	suite.Require().Error(err, "expected error when gas limit is zero")

	// Zero gas is accepted in simulation mode
	_, err = antehandler(suite.ctx, txBuilder.GetTx(), true)
	suite.Require().NoError(err, "expected no error when gas limit is zero in simulation mode")
}
