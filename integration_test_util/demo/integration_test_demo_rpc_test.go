package demo

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	rpctypes "github.com/evmos/evmos/v16/rpc/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

//goland:noinspection SpellCheckingInspection

func (suite *DemoTestSuite) Test_QC_Rpc_Balance() {
	address := suite.CITS.WalletAccounts.Number(1).GetEthAddress().String()

	res, err := suite.CITS.QueryClients.Rpc.Balance(
		rpctypes.ContextWithHeight(0),
		&evmtypes.QueryBalanceRequest{
			Address: address,
		},
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	balance, ok := sdkmath.NewIntFromString(res.Balance)
	suite.Require().True(ok)
	suite.True(balance.GT(sdk.ZeroInt()))
	suite.Equal(suite.CITS.TestConfig.InitBalanceAmount, balance)
}

func (suite *DemoTestSuite) Test_QC_Rpc_Balance_At_Different_Blocks() {
	sender := suite.CITS.WalletAccounts.Number(1)
	receiver := suite.CITS.WalletAccounts.Number(2)

	res, err := suite.CITS.QueryClients.Rpc.Balance(
		rpctypes.ContextWithHeight(0),
		&evmtypes.QueryBalanceRequest{
			Address: sender.GetEthAddress().String(),
		},
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	senderBalanceBefore, _ := sdkmath.NewIntFromString(res.Balance)

	suite.Require().Truef(senderBalanceBefore.GT(sdk.ZeroInt()), "sender must have balance")

	res, err = suite.CITS.QueryClients.Rpc.Balance(
		rpctypes.ContextWithHeight(0),
		&evmtypes.QueryBalanceRequest{
			Address: receiver.GetEthAddress().String(),
		},
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	receiverBalanceBefore, _ := sdkmath.NewIntFromString(res.Balance)

	err = suite.CITS.TxSend(sender, receiver, 0.1)
	suite.Commit()
	suite.Require().NoError(err)

	res, err = suite.CITS.QueryClients.Rpc.Balance(
		rpctypes.ContextWithHeight(0),
		&evmtypes.QueryBalanceRequest{
			Address: sender.GetEthAddress().String(),
		},
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	senderBalanceAfter, _ := sdkmath.NewIntFromString(res.Balance)

	res, err = suite.CITS.QueryClients.Rpc.Balance(
		rpctypes.ContextWithHeight(0),
		&evmtypes.QueryBalanceRequest{
			Address: receiver.GetEthAddress().String(),
		},
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	receiverBalanceAfter, _ := sdkmath.NewIntFromString(res.Balance)

	suite.NotEqualf(senderBalanceBefore.String(), senderBalanceAfter.String(), "sender balance must be reduced")
	suite.Require().Truef(senderBalanceAfter.LT(senderBalanceBefore), "sender balance must be reduced")

	suite.NotEqualf(receiverBalanceBefore.String(), receiverBalanceAfter.String(), "receiver balance must be increased")
	suite.Require().Truef(receiverBalanceBefore.LT(receiverBalanceAfter), "receiver balance must be increased")
}
