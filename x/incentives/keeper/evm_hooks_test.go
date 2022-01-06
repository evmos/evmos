package keeper_test

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/x/erc20/types"
)

// ensureHooksSet tries to set the hooks on EVMKeeper, this will fail if the
// incentives hook is already set
func (suite *KeeperTestSuite) ensureHooksSet() {
	defer func() {
		err := recover()
		suite.Require().NotNil(err)
	}()
	suite.app.EvmKeeper.SetHooks(suite.app.IncentivesKeeper)
}

func (suite *KeeperTestSuite) TestEvmHooksStoreTxGasUsed() {
	testCases := []struct {
		name     string
		malleate func(common.Address)
		expPass  bool
	}{
		{
			"correct execution",
			func(contract common.Address) {
				// _, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				// suite.Require().NoError(err)

				// // Mint 10 tokens to suite.address (owner)
				// _ = suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
				// suite.Commit()

				// // Burn the 10 tokens of suite.address (owner)
				// msg := suite.BurnERC20Token(contractAddr, suite.address, big.NewInt(10))
				// hash := msg.AsTransaction().Hash()
				// logs := suite.app.EvmKeeper.GetTxLogsTransient(hash)
				// suite.Require().NotEmpty(logs)
			},
			true,
		},
		// {
		// 	"unregistered pair",
		// 	func(contractAddr common.Address) {
		// 		// Mint 10 tokens to suite.address (owner)
		// 		_ = suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
		// 		suite.Commit()

		// 		// Burn the 10 tokens of suite.address (owner)
		// 		msg := suite.BurnERC20Token(contractAddr, suite.address, big.NewInt(10))
		// 		hash := msg.AsTransaction().Hash()
		// 		logs := suite.app.EvmKeeper.GetTxLogsTransient(hash)
		// 		suite.Require().NotEmpty(logs)
		// 	},
		// 	false,
		// },
		// {
		// 	"wrong event",
		// 	func(contractAddr common.Address) {
		// 		_, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
		// 		suite.Require().NoError(err)

		// 		// Mint 10 tokens to suite.address (owner)
		// 		msg := suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
		// 		hash := msg.AsTransaction().Hash()
		// 		logs := suite.app.EvmKeeper.GetTxLogsTransient(hash)
		// 		suite.Require().NotEmpty(logs)
		// 	},
		// 	false,
		// },
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			// suite.mintFeeCollector = true
			suite.SetupTest()
			suite.ensureHooksSet()

			// Deploy contract and create incentive
			contractAddr := suite.DeployContract("acoin", "COIN")
			suite.Commit()
			in, err := suite.app.IncentivesKeeper.RegisterIncentive(
				suite.ctx,
				contract,
				allocations,
				epochs,
			)
			suite.Require().NoError(err)
			suite.Commit()

			mint := big.NewInt(100)
			send := big.NewInt(10)

			// Submit contract Tx and make sure participant has enough tokens
			suite.MintERC20Token(contractAddr, suite.address, participant, mint)
			suite.Commit()
			res := suite.TransferERC20Token(contractAddr, participant, participant2, send)

			gasUsed := res.
			// check if gas is logged in Gas Meter
			gm, found := suite.app.IncentivesKeeper.GetIncentiveGasMeter(suite.ctx, contractAddr, participant)
			suite.Commit()

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(int64(10), balance.Amount.Int64())
			} else {
				suite.Require().Error(err)
				// Check that no changes were made to the account
				suite.Require().Equal(int64(0), balance.Amount.Int64())
			}
		})
	}
	suite.mintFeeCollector = false
}
