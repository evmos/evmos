package keeper_test

import (
	"fmt"

	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

// Test
func (suite *KeeperTestSuite) TestRegisterTokenPairWithContract() {
	suite.SetupTest()
	contractAddr := suite.DeployContract("coin", "token")
	suite.Commit()
	pair := types.NewTokenPair(contractAddr, "coinevm", true)
	err := suite.app.IntrarelayerKeeper.RegisterTokenPair(suite.ctx, pair)
	suite.Require().NoError(err)
	// TODO: check in the banking module if the Denom was created
}

func (suite KeeperTestSuite) TestRegisterTokenPair() {
	pair := types.NewTokenPair(tests.GenerateAddress(), "coin", true)
	id := pair.GetID()

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"intrarelaying is disabled globally",
			func() {
				params := types.DefaultParams()
				params.EnableIntrarelayer = false
				suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"token ERC20 already registered",
			func() {
				suite.app.IntrarelayerKeeper.SetERC20Map(suite.ctx, pair.GetERC20Contract(), id)
			},
			false,
		},
		{
			"denom already registered",
			func() {
				suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, pair.Denom, id)
			},
			false,
		},
		{
			"meta data already stored",
			func() {
				suite.app.IntrarelayerKeeper.CreateMetadata(suite.ctx, pair)
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			err := suite.app.IntrarelayerKeeper.RegisterTokenPair(suite.ctx, pair)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite KeeperTestSuite) TestEnableRelay() {
	var (
		pair types.TokenPair
		id   []byte
		err  error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"token not registered",
			func() {},
			false,
		},
		{
			"registered, disabled pair",
			func() {
				pair = types.NewTokenPair(tests.GenerateAddress(), "coin", true)
				id = pair.GetID()
				suite.app.IntrarelayerKeeper.SetTokenPair(suite.ctx, pair)
				suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, pair.Denom, id)
				suite.app.IntrarelayerKeeper.SetERC20Map(suite.ctx, pair.GetERC20Contract(), id)
				pair.Enabled = false
			},
			true,
		},
		{
			"registered, enabled pair",
			func() {
				pair = types.NewTokenPair(tests.GenerateAddress(), "coin", true)
				id = pair.GetID()
				suite.app.IntrarelayerKeeper.SetTokenPair(suite.ctx, pair)
				suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, pair.Denom, id)
				suite.app.IntrarelayerKeeper.SetERC20Map(suite.ctx, pair.GetERC20Contract(), id)
				pair.Enabled = true
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			pair, err = suite.app.IntrarelayerKeeper.EnableRelay(suite.ctx, "coin")
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().True(pair.Enabled)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite KeeperTestSuite) TestUpdateTokenPairERC20() {
	var (
		pair types.TokenPair
		id   []byte
		err  error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"token not registered",
			func() {},
			false,
		},
		{
			"registered pair",
			func() {
				pair = types.NewTokenPair(tests.GenerateAddress(), "coin", true)
				id = pair.GetID()
				suite.app.IntrarelayerKeeper.SetTokenPair(suite.ctx, pair)
				suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, pair.Denom, id)
				suite.app.IntrarelayerKeeper.SetERC20Map(suite.ctx, pair.GetERC20Contract(), id)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			erc20 := pair.GetERC20Contract()
			newErc20 := tests.GenerateAddress()

			pair, err = suite.app.IntrarelayerKeeper.UpdateTokenPairERC20(suite.ctx, erc20, newErc20)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(newErc20.Hex(), pair.Erc20Address)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}
