package demo

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v16/integration_test_util"
)

//goland:noinspection SpellCheckingInspection

func (suite *DemoTestSuite) Test_MintCoins() {
	newAccount := integration_test_util.NewTestAccount(suite.T(), nil)

	balance := suite.CITS.QueryBalance(0, newAccount.GetCosmosAddress().String())
	suite.Require().True(balance.Amount.Equal(sdk.ZeroInt()))

	mintCoin := sdk.NewCoin(suite.CITS.ChainConstantsConfig.GetMinDenom(), suite.CITS.TestConfig.InitBalanceAmount)
	suite.CITS.MintCoin(newAccount, mintCoin)
	suite.Commit()

	balance = suite.CITS.QueryBalance(0, newAccount.GetCosmosAddress().String())
	suite.Require().True(balance.Amount.Equal(mintCoin.Amount))
}

func (suite *DemoTestSuite) Test_QC_Bank_Balance() {
	balance := suite.CITS.QueryBalance(0, suite.CITS.WalletAccounts.Number(1).GetCosmosAddress().String())

	suite.Require().True(balance.Amount.GT(sdk.ZeroInt()))
	suite.Equal(suite.CITS.TestConfig.InitBalanceAmount, balance.Amount)

	secondaryBalance := suite.CITS.QueryBalanceByDenom(
		0,
		suite.CITS.WalletAccounts.Number(1).GetCosmosAddress().String(),
		suite.CITS.TestConfig.SecondaryDenomUnits[0].Denom,
	)

	suite.Require().True(secondaryBalance.Amount.GT(sdk.ZeroInt()))
	suite.Equal(suite.CITS.TestConfig.InitBalanceAmount, secondaryBalance.Amount)
}

func (suite *DemoTestSuite) Test_QC_Bank_Balance_At_Different_Blocks() {
	sender := suite.CITS.WalletAccounts.Number(1)
	receiver := suite.CITS.WalletAccounts.Number(2)

	senderBalanceBefore := suite.CITS.QueryBalance(0, sender.GetCosmosAddress().String())
	receiverBalanceBefore := suite.CITS.QueryBalance(0, receiver.GetCosmosAddress().String())

	suite.Require().Truef(senderBalanceBefore.Amount.GT(sdk.ZeroInt()), "sender must have balance")

	contextHeightBeforeSend := suite.CITS.CurrentContext.BlockHeight()
	suite.Commit()

	err := suite.CITS.TxSend(sender, receiver, 0.1)
	suite.Commit()
	suite.Require().NoError(err)

	senderBalanceAfter := suite.CITS.QueryBalance(0, sender.GetCosmosAddress().String())
	receiverBalanceAfter := suite.CITS.QueryBalance(0, receiver.GetCosmosAddress().String())

	suite.NotEqualf(senderBalanceBefore.String(), senderBalanceAfter.String(), "sender balance must be reduced")
	suite.Require().Truef(senderBalanceAfter.IsLT(*senderBalanceBefore), "sender balance must be reduced")

	suite.NotEqualf(receiverBalanceBefore.String(), receiverBalanceAfter.String(), "receiver balance must be increased")
	suite.Require().Truef(receiverBalanceBefore.IsLT(*receiverBalanceAfter), "receiver balance must be increased")

	// Historical block height
	historicalSenderBalance := suite.CITS.QueryBalance(contextHeightBeforeSend, sender.GetCosmosAddress().String())
	historicalReceiverBalanceAfter := suite.CITS.QueryBalance(contextHeightBeforeSend, receiver.GetCosmosAddress().String())
	suite.Equal(senderBalanceBefore.String(), historicalSenderBalance.String(), "mis-match sender balance at historical height")
	suite.Equal(receiverBalanceBefore.String(), historicalReceiverBalanceAfter.String(), "mis-match sender balance at historical height")
}
