package keeper_test

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

func (suite *KeeperTestSuite) TestEvmHooks() {

	testCases := []struct {
		name     string
		malleate func(common.Address) error
		result   bool
	}{
		{
			"correct execution",
			func(contractAddr common.Address) error {
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
				return suite.app.IntrarelayerKeeper.PostTxProcessing(suite.ctx, msg.AsTransaction().Hash(), logs)

			},
			true,
		},
		{
			"Unregistered pair",
			func(contractAddr common.Address) error {
				// Mint 10 tokens to suite.address (owner)
				_ = suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
				suite.Commit()

				// Burn the 10 tokens of suite.address (owner)
				msg := suite.BurnERC20Token(contractAddr, suite.address, big.NewInt(10))
				logs := suite.app.EvmKeeper.GetTxLogsTransient(msg.AsTransaction().Hash())

				// Since theres no pair registered, no coins should be minted
				return suite.app.IntrarelayerKeeper.PostTxProcessing(suite.ctx, msg.AsTransaction().Hash(), logs)
			},
			false,
		},
		{
			"Wrong event",
			func(contractAddr common.Address) error {
				pair := types.NewTokenPair(contractAddr, "coinevm", true)
				err := suite.app.IntrarelayerKeeper.RegisterTokenPair(suite.ctx, pair)
				suite.Require().NoError(err)

				// Mint 10 tokens to suite.address (owner)
				msg := suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
				logs := suite.app.EvmKeeper.GetTxLogsTransient(msg.AsTransaction().Hash())

				// No coins should be minted on cosmos after a mint of the erc20 token
				return suite.app.IntrarelayerKeeper.PostTxProcessing(suite.ctx, msg.AsTransaction().Hash(), logs)
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			contractAddr := suite.DeployContract("coin", "token")
			suite.Commit()

			err := tc.malleate(contractAddr)
			//None of this test should error
			suite.Require().NoError(err)

			if tc.result {
				// Check if the execution was successfull
				suite.Require().Equal(suite.app.BankKeeper.GetBalance(suite.ctx, sdk.AccAddress(suite.address.Bytes()), "coinevm").Amount, sdk.NewIntFromBigInt(big.NewInt(10)))
			} else {
				// Check that no changes were made to the account
				suite.Require().Equal(suite.app.BankKeeper.GetBalance(suite.ctx, sdk.AccAddress(suite.address.Bytes()), "coinevm").Amount, sdk.NewIntFromBigInt(big.NewInt(0)))
			}
		})
	}
}
