package keeper_test

import (
	"math/big"

	"github.com/tharsis/evmos/x/intrarelayer/types"
)

// Test
func (suite *KeeperTestSuite) TestEvmHooks() {
	suite.SetupTest()

	// Module account is missing mint/burn permits

	contractAddr := suite.DeployContract("coin", "token")
	suite.Commit()
	pair := types.NewTokenPair(contractAddr, "coinevm", true)
	err := suite.app.IntrarelayerKeeper.RegisterTokenPair(suite.ctx, pair)
	suite.Require().NoError(err)
	_ = suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
	suite.Commit()

	msg := suite.BurnERC20Token(contractAddr, suite.address, big.NewInt(10))
	logs := suite.app.EvmKeeper.GetTxLogsTransient(msg.AsTransaction().Hash())
	err = suite.app.IntrarelayerKeeper.PostTxProcessing(suite.ctx, msg.AsTransaction().Hash(), logs)
	// Check correct execution
}
