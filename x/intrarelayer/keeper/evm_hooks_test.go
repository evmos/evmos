package keeper_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

// Test
func (suite *KeeperTestSuite) TestEvmHooks() {
	suite.SetupTest()

	contractAddr := suite.DeployContract("coin", "token")
	suite.Commit()
	pair := types.NewTokenPair(contractAddr, "coinevm", true)
	err := suite.app.IntrarelayerKeeper.RegisterTokenPair(suite.ctx, pair)
	suite.Require().NoError(err)

	// Mint 10 tokens to suite.address (owner)
	_ = suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
	suite.Commit()

	// Burn the 10 tokens of suite.address (owner)
	msg := suite.BurnERC20Token(contractAddr, suite.address, big.NewInt(10))
	logs := suite.app.EvmKeeper.GetTxLogsTransient(msg.AsTransaction().Hash())

	// After this execution, the burned tokens will be available on the cosmos chain
	err = suite.app.IntrarelayerKeeper.PostTxProcessing(suite.ctx, msg.AsTransaction().Hash(), logs)

	// Check correct execution
	suite.Require().NoError(err)
	// Acc Address should have the 10 coins that were burned in the transaction
	suite.Require().Equal(suite.app.BankKeeper.GetBalance(suite.ctx, sdk.AccAddress(suite.address.Bytes()), "coinevm").Amount, sdk.NewIntFromBigInt(big.NewInt(10)))
}
