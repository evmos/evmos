package keeper_test

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/ethermint/tests"
	evm "github.com/tharsis/ethermint/x/evm/types"
	"github.com/tharsis/evmos/x/incentives/types"
)

// ensureHooksSet tries to set the hooks on EVMKeeper, this will fail if the
// incentives hook is already set
func (suite *KeeperTestSuite) ensureHooksSet() {
	defer func() {
		err := recover()
		suite.Require().NotNil(err)
	}()
	suite.app.EvmKeeper.SetHooks(suite.app.IncentivesKeeper.Hooks())
}

func (suite *KeeperTestSuite) TestEvmHooksStoreTxGasUsed() {
	var expGasUsed uint64

	testCases := []struct {
		name     string
		malleate func(common.Address)

		expPass bool
	}{
		{
			"correct execution - one tx",
			func(contractAddr common.Address) {
				res := suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(1000))
				expGasUsed = res.AsTransaction().Gas()
			},
			true,
		},
		{
			"correct execution - two tx",
			func(contractAddr common.Address) {
				res := suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(500))
				res2 := suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(500))
				expGasUsed = res.AsTransaction().Gas() + res2.AsTransaction().Gas()
			},
			true,
		},
		{
			"tx with unincentivized contract",
			func(contractAddr common.Address) {
				suite.MintERC20Token(tests.GenerateAddress(), suite.address, suite.address, big.NewInt(1000))
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			suite.ensureHooksSet()

			// Deploy Contract
			contractAddr := suite.DeployContract(denomCoin, "COIN", erc20Decimals)
			suite.Commit()

			// Register Incentive
			_, err := suite.app.IncentivesKeeper.RegisterIncentive(
				suite.ctx,
				contractAddr,
				mintAllocations,
				epochs,
			)
			suite.Require().NoError(err)

			// Mint coins to pay gas fee
			coins := sdk.NewCoins(sdk.NewCoin(evm.DefaultEVMDenom, sdk.NewInt(30000000)))
			suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)

			err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(
				suite.ctx,
				types.ModuleName,
				sdk.AccAddress(suite.address.Bytes()),
				coins,
			)
			suite.Require().NoError(err)

			// Submit tx
			tc.malleate(contractAddr)

			incentive, _ := suite.app.IncentivesKeeper.GetIncentive(suite.ctx, contractAddr)
			totalGas := incentive.TotalGas
			gm, found := suite.app.IncentivesKeeper.GetGasMeter(
				suite.ctx,
				contractAddr,
				suite.address,
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().True(found)
				suite.Require().NotZero(gm)
				suite.Require().Equal(expGasUsed, gm)
				suite.Require().Equal(expGasUsed, totalGas)
			} else {
				suite.Require().NoError(err)
				suite.Require().Zero(gm)
				suite.Require().Zero(totalGas)
			}
		})
	}
	suite.mintFeeCollector = false
}
