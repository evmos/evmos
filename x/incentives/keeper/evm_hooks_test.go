package keeper_test

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/evmos/ethermint/tests"
	ethermint "github.com/evmos/ethermint/types"
	evm "github.com/evmos/ethermint/x/evm/types"

	"github.com/evmos/evmos/v9/testutil"
	"github.com/evmos/evmos/v9/x/incentives/types"
	vestingtypes "github.com/evmos/evmos/v9/x/vesting/types"
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
			"incentives are disabled globally",
			func(_ common.Address) {
				params := types.DefaultParams()
				params.EnableIncentives = false
				suite.app.IncentivesKeeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"from address is not an EOA",
			func(contractAddr common.Address) {
				// set a contract account for the address
				contract := &ethermint.EthAccount{
					BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
					CodeHash:    common.BytesToHash(crypto.Keccak256([]byte{0, 1, 2, 2})).String(),
				}

				suite.app.AccountKeeper.SetAccount(suite.ctx, contract)
				res := suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(1000))
				expGasUsed = res.AsTransaction().Gas()
			},
			false,
		},
		{
			"correct execution - one tx",
			func(contractAddr common.Address) {
				acc := &ethermint.EthAccount{
					BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
					CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
				}
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				res := suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(1000))
				expGasUsed = res.AsTransaction().Gas()
			},
			true,
		},
		{
			"correct execution with Base account - one tx",
			func(contractAddr common.Address) {
				acc := authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0)
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				res := suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(1000))
				expGasUsed = res.AsTransaction().Gas()
			},
			true,
		},
		{
			"correct execution with Vesting account - one tx",
			func(contractAddr common.Address) {
				acc := vestingtypes.NewClawbackVestingAccount(
					authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
					suite.address.Bytes(), nil, suite.ctx.BlockTime(), nil, nil,
				)

				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				res := suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(1000))
				expGasUsed = res.AsTransaction().Gas()
			},
			true,
		},
		{
			"correct execution - two tx",
			func(contractAddr common.Address) {
				acc := &ethermint.EthAccount{
					BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
					CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
				}
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				res := suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(500))
				res2 := suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(500))
				expGasUsed = res.AsTransaction().Gas() + res2.AsTransaction().Gas()
			},
			true,
		},
		{
			"tx with non-incentivized contract",
			func(_ common.Address) {
				_ = suite.MintERC20Token(tests.GenerateAddress(), suite.address, suite.address, big.NewInt(1000))
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
			contractAddr, err := suite.DeployContract(denomCoin, "COIN", erc20Decimals)
			suite.Require().NoError(err)
			suite.Commit()

			// Register Incentive
			_, err = suite.app.IncentivesKeeper.RegisterIncentive(
				suite.ctx,
				contractAddr,
				mintAllocations,
				epochs,
			)
			suite.Require().NoError(err)

			// Mint coins to pay gas fee
			coins := sdk.NewCoins(sdk.NewCoin(evm.DefaultEVMDenom, sdk.NewInt(30000000)))
			err = testutil.FundAccount(suite.app.BankKeeper, suite.ctx, sdk.AccAddress(suite.address.Bytes()), coins)
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
				suite.Require().NotZero(totalGas)
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
