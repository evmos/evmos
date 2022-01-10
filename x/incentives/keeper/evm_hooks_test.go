package keeper_test

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
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

// TODO Fix test when the ethermint version is updated
func (suite *KeeperTestSuite) TestEvmHooksStoreTxGasUsed() {
	testCases := []struct {
		name     string
		malleate func(common.Address)
		expPass  bool
	}{
		{
			"correct execution",
			func(contractAddr common.Address) {
				// Mint coins to pay for gas fee
				coins := sdk.NewCoins(sdk.NewCoin(evm.DefaultEVMDenom, sdk.NewInt(30000000)))
				suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
				err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(
					suite.ctx,
					types.ModuleName,
					sdk.AccAddress(participant.Bytes()),
					coins,
				)
				suite.Require().NoError(err)

				// Mint tokens to transfer
				suite.MintERC20Token(contractAddr, suite.address, participant, big.NewInt(1000000000))
				suite.Commit()

				balanceCoin := suite.BalanceOf(contractAddr, participant)
				fmt.Printf("balanceCoin: %v\n", balanceCoin)

				balanceToken := suite.app.EvmKeeper.GetBalance(suite.ctx, participant)
				fmt.Printf("balanceToken: %v\n", balanceToken)

				// submit contract Tx
				suite.TransferERC20Token(contractAddr, participant, participant2, big.NewInt(0))
			},
			true,
		},
		// {"unincentivized contract", func(contractAddr common.Address) {}, false},
		// {"wrong event", func(contractAddr common.Address) {}, false},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			suite.ensureHooksSet()

			// Deploy contract, nint and create incentive
			contractAddr := suite.DeployContract(denomCoin, "COIN", erc20Decimals)
			suite.Commit()

			in, err := suite.app.IncentivesKeeper.RegisterIncentive(
				suite.ctx,
				contractAddr,
				sdk.DecCoins{
					sdk.NewDecCoinFromDec(denomMint, sdk.NewDecWithPrec(allocationRate, 2)),
				},
				epochs,
			)
			suite.Require().NoError(err)
			suite.Commit()

			// submit Tx
			tc.malleate(contractAddr)
			suite.Commit()

			expGasUsed := int64(10)
			gm, found := suite.app.IncentivesKeeper.GetIncentiveGasMeter(
				suite.ctx,
				contractAddr,
				participant,
			)
			suite.Require().True(found)
			suite.Commit()

			// check if gasUsed is logged
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotZero(gm)
				suite.Require().Equal(expGasUsed, gm)
				suite.Require().Equal(expGasUsed, in.TotalGas)
			} else {
				suite.Require().Error(err)
				suite.Require().Zero(gm)
				suite.Require().Zero(in.TotalGas)
			}
		})
	}
	suite.mintFeeCollector = false
}
