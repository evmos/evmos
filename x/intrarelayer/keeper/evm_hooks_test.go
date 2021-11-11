package keeper_test

import (
	"math/big"

	"github.com/tharsis/evmos/x/intrarelayer/types"
	"github.com/tharsis/evmos/x/intrarelayer/types/contracts"
)

// func (suite *KeeperTestSuite) TestEvmHooks() {

// 	testCases := []struct {
// 		name     string
// 		malleate func(common.Address)
// 		result   bool
// 	}{
// 		{
// 			"correct execution",
// 			func(contractAddr common.Address) {
// 				pair := types.NewTokenPair(contractAddr, "coinevm", true, types.MODULE_OWNER)
// 				err := suite.app.IntrarelayerKeeper.RegisterTokenPair(suite.ctx, pair)
// 				suite.Require().NoError(err)

// 				// Mint 10 tokens to suite.address (owner)
// 				_ = suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
// 				suite.Commit()

// 				// Burn the 10 tokens of suite.address (owner)
// 				msg := suite.BurnERC20Token(contractAddr, suite.address, big.NewInt(10))
// 				logs := suite.app.EvmKeeper.GetTxLogsTransient(msg.AsTransaction().Hash())

// 				// After this execution, the burned tokens will be available on the cosmos chain
// 				err = suite.app.IntrarelayerKeeper.PostTxProcessing(suite.ctx, msg.AsTransaction().Hash(), logs)
// 				suite.Require().NoError(err)
// 			},
// 			true,
// 		},
// 		{
// 			"Unregistered pair",
// 			func(contractAddr common.Address) {
// 				// Mint 10 tokens to suite.address (owner)
// 				_ = suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
// 				suite.Commit()

// 				// Burn the 10 tokens of suite.address (owner)
// 				msg := suite.BurnERC20Token(contractAddr, suite.address, big.NewInt(10))
// 				logs := suite.app.EvmKeeper.GetTxLogsTransient(msg.AsTransaction().Hash())

// 				// Since theres no pair registered, no coins should be minted
// 				err := suite.app.IntrarelayerKeeper.PostTxProcessing(suite.ctx, msg.AsTransaction().Hash(), logs)
// 				suite.Require().NoError(err)
// 			},
// 			false,
// 		},
// 		{
// 			"Wrong event",
// 			func(contractAddr common.Address) {
// 				pair := types.NewTokenPair(contractAddr, "coinevm", true, types.MODULE_OWNER)
// 				err := suite.app.IntrarelayerKeeper.RegisterTokenPair(suite.ctx, pair)
// 				suite.Require().NoError(err)

// 				// Mint 10 tokens to suite.address (owner)
// 				msg := suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
// 				logs := suite.app.EvmKeeper.GetTxLogsTransient(msg.AsTransaction().Hash())

// 				// No coins should be minted on cosmos after a mint of the erc20 token
// 				err = suite.app.IntrarelayerKeeper.PostTxProcessing(suite.ctx, msg.AsTransaction().Hash(), logs)
// 				suite.Require().NoError(err)
// 			},
// 			false,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
// 			suite.mintFeeCollector = true
// 			suite.SetupTest()

// 			contractAddr := suite.DeployContract("coin", "token")
// 			suite.Commit()

// 			tc.malleate(contractAddr)

// 			balance := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.AccAddress(suite.address.Bytes()), "coinevm")

// 			if tc.result {
// 				// Check if the execution was successfull
// 				suite.Require().Equal(balance.Amount, sdk.NewInt(10))
// 			} else {
// 				// Check that no changes were made to the account
// 				suite.Require().Equal(balance.Amount, sdk.NewInt(0))
// 			}
// 		})
// 	}
// 	suite.mintFeeCollector = false
// }

func (suite *KeeperTestSuite) TestTransferBurn() {
	suite.mintFeeCollector = true
	contractAddr := suite.setupNewTokenPair()
	suite.Require().NotNil(contractAddr)

	// Mint 10 tokens to suite.address (owner)
	_ = suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
	suite.Commit()

	rsp := suite.TransferERC20Token(contractAddr, suite.address, types.ModuleAddress, big.NewInt(10))
	suite.Commit()

	unpackedRet, err := contracts.ERC20BurnableAndMintableContract.ABI.Unpack("transfer", rsp.Ret)
	suite.Require().True(unpackedRet[0].(bool))

	balance := suite.BalanceOf(contractAddr, types.ModuleAddress)
	suite.Require().Equal(balance, big.NewInt(10))

	balance = suite.BalanceOf(contractAddr, suite.address)
	// zero := big.NewInt(int64(0))
	// suite.Require().Equal(balance.(*big.Int), zero)

	transferData, err := contracts.ERC20BurnableAndMintableContract.ABI.Pack("transfer", suite.address, big.NewInt(10))
	suite.Require().NoError(err)

	suite.app.IntrarelayerKeeper.CallEVMWithPayload(suite.ctx, contractAddr, transferData)
	// _ = suite.TransferERC20Token(contractAddr, types.ModuleAddress, suite.address, big.NewInt(10))
	suite.Commit()

	balance = suite.BalanceOf(contractAddr, types.ModuleAddress)
	// zero := big.NewInt(int64(0))
	// suite.Require().Equal(balance.(*big.Int), zero)

	balance = suite.BalanceOf(contractAddr, suite.address)
	suite.Require().Equal(balance, big.NewInt(10))
	// zero := big.NewInt(int64(0))
	// suite.Require().Equal(balance.(*big.Int), zero)

	suite.mintFeeCollector = false

}
