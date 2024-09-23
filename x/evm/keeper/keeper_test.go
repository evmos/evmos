package keeper_test

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/evmos/evmos/v20/utils"
	"github.com/evmos/evmos/v20/x/evm/config"
	"github.com/evmos/evmos/v20/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
)

func (suite *KeeperTestSuite) TestBaseFee() {
	testCases := []struct {
		name            string
		enableLondonHF  bool
		enableFeemarket bool
		expectBaseFee   *big.Int
	}{
		{"not enable london HF, not enable feemarket", false, false, nil},
		{"enable london HF, not enable feemarket", true, false, big.NewInt(0)},
		{"enable london HF, enable feemarket", true, true, big.NewInt(1000000000)},
		{"not enable london HF, enable feemarket", false, true, nil},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.enableFeemarket = tc.enableFeemarket
			suite.enableLondonHF = tc.enableLondonHF
			suite.SetupTest()
			ethCfg := config.GetChainConfig()
			baseFee := suite.network.App.EvmKeeper.GetBaseFee(suite.network.GetContext(), ethCfg)
			suite.Require().Equal(tc.expectBaseFee, baseFee)
		})
	}
	suite.enableFeemarket = false
	suite.enableLondonHF = true
}

func (suite *KeeperTestSuite) TestGetAccountStorage() {
	var ctx sdk.Context
	testCases := []struct {
		name     string
		malleate func() common.Address
	}{
		{
			name:     "Only accounts that are not a contract (no storage)",
			malleate: nil,
		},
		{
			name: "One contract (with storage) and other EOAs",
			malleate: func() common.Address {
				supply := big.NewInt(100)
				contractAddr := suite.DeployTestContract(suite.T(), ctx, suite.keyring.GetAddr(0), supply)
				return contractAddr
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			ctx = suite.network.GetContext()

			var contractAddr common.Address
			if tc.malleate != nil {
				contractAddr = tc.malleate()
			}

			i := 0
			suite.network.App.AccountKeeper.IterateAccounts(ctx, func(account sdk.AccountI) bool {
				acc, ok := account.(*authtypes.BaseAccount)
				if !ok {
					// Ignore e.g. module accounts
					return false
				}

				address, err := utils.Bech32ToHexAddr(acc.Address)
				if err != nil {
					// NOTE: we panic in the test to see any potential problems
					// instead of skipping to the next account
					panic(fmt.Sprintf("failed to convert %s to hex address", err))
				}

				storage := suite.network.App.EvmKeeper.GetAccountStorage(ctx, address)

				if address == contractAddr {
					suite.Require().NotEqual(0, len(storage),
						"expected account %d to have non-zero amount of storage slots, got %d",
						i, len(storage),
					)
				} else {
					suite.Require().Len(storage, 0,
						"expected account %d to have %d storage slots, got %d",
						i, 0, len(storage),
					)
				}

				i++
				return false
			})
		})
	}
}

func (suite *KeeperTestSuite) TestGetAccountOrEmpty() {
	ctx := suite.network.GetContext()
	empty := statedb.Account{
		Balance:  new(big.Int),
		CodeHash: evmtypes.EmptyCodeHash,
	}

	supply := big.NewInt(100)
	contractAddr := suite.DeployTestContract(suite.T(), ctx, suite.keyring.GetAddr(0), supply)

	testCases := []struct {
		name     string
		addr     common.Address
		expEmpty bool
	}{
		{
			"unexisting account - get empty",
			common.Address{},
			true,
		},
		{
			"existing contract account",
			contractAddr,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			res := suite.network.App.EvmKeeper.GetAccountOrEmpty(ctx, tc.addr)
			if tc.expEmpty {
				suite.Require().Equal(empty, res)
			} else {
				suite.Require().NotEqual(empty, res)
			}
		})
	}
}
