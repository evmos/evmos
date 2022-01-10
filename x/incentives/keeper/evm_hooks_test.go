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
	suite.app.EvmKeeper.SetHooks(suite.app.IncentivesKeeper)
}

func (suite *KeeperTestSuite) TestEvmHooksStoreTxGasUsed() {
	testCases := []struct {
		name       string
		malleate   func(common.Address)
		expGasUsed uint64
		expPass    bool
	}{
		{
			"correct execution - one tx",
			func(contractAddr common.Address) {
				suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(1000))
			},
			uint64(73820),
			true,
		},
		{
			"correct execution - two tx",
			func(contractAddr common.Address) {
				suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(500))
				suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(500))
			},
			uint64(113440),
			true,
		},
		{
			"tx with unincentivized contract",
			func(contractAddr common.Address) {
				suite.MintERC20Token(tests.GenerateAddress(), suite.address, suite.address, big.NewInt(1000))
			},
			uint64(0),
			false,
		},
		// {"wrong event", func(contractAddr common.Address) {}, false},
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
			incentive, err := suite.app.IncentivesKeeper.RegisterIncentive(
				suite.ctx,
				contractAddr,
				sdk.DecCoins{
					sdk.NewDecCoinFromDec(denomMint, sdk.NewDecWithPrec(allocationRate, 2)),
				},
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

			totalGas := suite.app.IncentivesKeeper.GetIncentiveTotalGas(suite.ctx, *incentive)
			gm, found := suite.app.IncentivesKeeper.GetIncentiveGasMeter(
				suite.ctx,
				contractAddr,
				suite.address,
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().True(found)
				suite.Require().NotZero(gm)
				suite.Require().Equal(tc.expGasUsed, gm)
				suite.Require().Equal(tc.expGasUsed, totalGas)
			} else {
				suite.Require().NoError(err)
				suite.Require().Zero(gm)
				suite.Require().Zero(totalGas)
			}
		})
	}
	suite.mintFeeCollector = false
}
