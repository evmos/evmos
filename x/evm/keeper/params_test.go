package keeper_test

import (
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/x/evm/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := types.DefaultParams()

	testCases := []struct {
		name      string
		paramsFun func() interface{}
		getFun    func() interface{}
		expected  bool
	}{
		{
			"success - Checks if the default params are set correctly",
			func() interface{} {
				return types.DefaultParams()
			},
			func() interface{} {
				return suite.app.EvmKeeper.GetParams(suite.ctx)
			},
			true,
		},
		{
			"success - EvmDenom param is set to \"inj\" and can be retrieved correctly",
			func() interface{} {
				params.EvmDenom = "inj"
				err := suite.app.EvmKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return params.EvmDenom
			},
			func() interface{} {
				evmParams := suite.app.EvmKeeper.GetParams(suite.ctx)
				return evmParams.GetEvmDenom()
			},
			true,
		},
		{
			"success - Check EnableCreate param is set to false and can be retrieved correctly",
			func() interface{} {
				params.EnableCreate = false
				err := suite.app.EvmKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return params.EnableCreate
			},
			func() interface{} {
				evmParams := suite.app.EvmKeeper.GetParams(suite.ctx)
				return evmParams.GetEnableCreate()
			},
			true,
		},
		{
			"success - Check EnableCall param is set to false and can be retrieved correctly",
			func() interface{} {
				params.EnableCall = false
				err := suite.app.EvmKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return params.EnableCall
			},
			func() interface{} {
				evmParams := suite.app.EvmKeeper.GetParams(suite.ctx)
				return evmParams.GetEnableCall()
			},
			true,
		},
		{
			"success - Check AllowUnprotectedTxs param is set to false and can be retrieved correctly",
			func() interface{} {
				params.AllowUnprotectedTxs = false
				err := suite.app.EvmKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return params.AllowUnprotectedTxs
			},
			func() interface{} {
				evmParams := suite.app.EvmKeeper.GetParams(suite.ctx)
				return evmParams.GetAllowUnprotectedTxs()
			},
			true,
		},
		{
			"success - Check ChainConfig param is set to the default value and can be retrieved correctly",
			func() interface{} {
				params.ChainConfig = types.DefaultChainConfig()
				err := suite.app.EvmKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return params.ChainConfig
			},
			func() interface{} {
				evmParams := suite.app.EvmKeeper.GetParams(suite.ctx)
				return evmParams.GetChainConfig()
			},
			true,
		},
		{
			name: "success - Active precompiles are sorted when setting params",
			paramsFun: func() interface{} {
				params.ActiveStaticPrecompiles = []string{
					"0x0000000000000000000000000000000000000801",
					"0x0000000000000000000000000000000000000800",
				}
				err := suite.app.EvmKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err, "expected no error when setting params")

				// NOTE: return sorted slice here because the precompiles should be sorted when setting the params
				return []string{
					"0x0000000000000000000000000000000000000800",
					"0x0000000000000000000000000000000000000801",
				}
			},
			getFun: func() interface{} {
				evmParams := suite.app.EvmKeeper.GetParams(suite.ctx)
				return evmParams.GetActiveStaticPrecompiles()
			},
			expected: true,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			outcome := reflect.DeepEqual(tc.paramsFun(), tc.getFun())
			suite.Require().Equal(tc.expected, outcome)
		})
	}
}

func (suite *KeeperTestSuite) TestEnableStaticPrecompiles() {
	params := types.DefaultParams()

	testCases := []struct {
		name              string
		addresses         []common.Address
		expectedaddresses []string
	}{
		{
			"success - default precompiles",
			[]common.Address{},
			params.ActiveStaticPrecompiles,
		},
		{
			"success - Add a static precompile",
			[]common.Address{common.HexToAddress("0xD4949664cD82660AaE99bEdc034a0deA8A0bd517")},
			append(params.ActiveStaticPrecompiles, "0xD4949664cD82660AaE99bEdc034a0deA8A0bd517"),
		},
		{
			"success - Add several static precompiles / and sort",
			[]common.Address{common.HexToAddress("0xD4949664cD82660AaE99bEdc034a0deA8A0bd517"), common.HexToAddress("0xA61808Fe40fEb8B3433778BBC2ecECCAA47c8c47")},
			append(params.ActiveStaticPrecompiles, "0xA61808Fe40fEb8B3433778BBC2ecECCAA47c8c47", "0xD4949664cD82660AaE99bEdc034a0deA8A0bd517"),
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			err := suite.app.EvmKeeper.EnableStaticPrecompiles(s.ctx, tc.addresses...)
			suite.Require().NoError(err)

			updated := suite.app.EvmKeeper.GetParams(s.ctx).ActiveStaticPrecompiles
			suite.Require().Equal(tc.expectedaddresses, updated)
		})
	}
}
