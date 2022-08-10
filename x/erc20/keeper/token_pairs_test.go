package keeper_test

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/tests"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/evmos/evmos/v8/x/erc20/types"
)

func (suite *KeeperTestSuite) TestGetTokenPairs() {
	var expRes []types.TokenPair

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"no pair registered", func() { expRes = []types.TokenPair{} },
		},
		{
			"1 pair registered",
			func() {
				pair := types.NewTokenPair(tests.GenerateAddress(), "coin", true, types.OWNER_MODULE)
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair)

				expRes = []types.TokenPair{pair}
			},
		},
		{
			"2 pairs registered",
			func() {
				pair := types.NewTokenPair(tests.GenerateAddress(), "coin", true, types.OWNER_MODULE)
				pair2 := types.NewTokenPair(tests.GenerateAddress(), "coin2", true, types.OWNER_MODULE)
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair)
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair2)

				expRes = []types.TokenPair{pair, pair2}
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()
			res := suite.app.Erc20Keeper.GetTokenPairs(suite.ctx)

			suite.Require().ElementsMatch(expRes, res, tc.name)
		})
	}
}

func (suite *KeeperTestSuite) TestGetTokenPairID() {
	pair := types.NewTokenPair(tests.GenerateAddress(), evmtypes.DefaultEVMDenom, true, types.OWNER_MODULE)
	suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair)

	testCases := []struct {
		name  string
		token string
		expId []byte
	}{
		{"nil token", "", nil},
		{"valid hex token", tests.GenerateAddress().Hex(), []byte{}},
		{"valid hex token", tests.GenerateAddress().String(), []byte{}},
	}
	for _, tc := range testCases {
		id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, tc.token)
		if id != nil {
			suite.Require().Equal(tc.expId, id, tc.name)
		} else {
			suite.Require().Nil(id)
		}
	}
}

func (suite *KeeperTestSuite) TestGetTokenPair() {
	pair := types.NewTokenPair(tests.GenerateAddress(), evmtypes.DefaultEVMDenom, true, types.OWNER_MODULE)
	suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair)

	testCases := []struct {
		name string
		id   []byte
		ok   bool
	}{
		{"nil id", nil, false},
		{"valid id", pair.GetID(), true},
		{"pair not found", []byte{}, false},
	}
	for _, tc := range testCases {
		p, found := suite.app.Erc20Keeper.GetTokenPair(suite.ctx, tc.id)
		if tc.ok {
			suite.Require().True(found, tc.name)
			suite.Require().Equal(pair, p, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestDeleteTokenPair() {
	pair := types.NewTokenPair(tests.GenerateAddress(), evmtypes.DefaultEVMDenom, true, types.OWNER_MODULE)
	id := pair.GetID()
	suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair)
	suite.app.Erc20Keeper.SetERC20Map(suite.ctx, pair.GetERC20Contract(), id)
	suite.app.Erc20Keeper.SetDenomMap(suite.ctx, pair.Denom, id)

	testCases := []struct {
		name     string
		id       []byte
		malleate func()
		ok       bool
	}{
		{"nil id", nil, func() {}, false},
		{"pair not found", []byte{}, func() {}, false},
		{"valid id", id, func() {}, true},
		{
			"delete tokenpair",
			id,
			func() {
				suite.app.Erc20Keeper.DeleteTokenPair(suite.ctx, pair)
			},
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate()
		p, found := suite.app.Erc20Keeper.GetTokenPair(suite.ctx, tc.id)
		if tc.ok {
			suite.Require().True(found, tc.name)
			suite.Require().Equal(pair, p, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestIsTokenPairRegistered() {
	pair := types.NewTokenPair(tests.GenerateAddress(), evmtypes.DefaultEVMDenom, true, types.OWNER_MODULE)
	suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair)

	testCases := []struct {
		name string
		id   []byte
		ok   bool
	}{
		{"valid id", pair.GetID(), true},
		{"pair not found", []byte{}, false},
	}
	for _, tc := range testCases {
		found := suite.app.Erc20Keeper.IsTokenPairRegistered(suite.ctx, tc.id)
		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestIsERC20Registered() {
	addr := tests.GenerateAddress()
	pair := types.NewTokenPair(addr, "coin", true, types.OWNER_MODULE)
	suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair)
	suite.app.Erc20Keeper.SetERC20Map(suite.ctx, addr, pair.GetID())
	suite.app.Erc20Keeper.SetDenomMap(suite.ctx, pair.Denom, pair.GetID())

	testCases := []struct {
		name     string
		erc20    common.Address
		malleate func()
		ok       bool
	}{
		{"nil erc20 address", common.Address{}, func() {}, false},
		{"valid erc20 address", pair.GetERC20Contract(), func() {}, true},
		{
			"deleted erc20 map",
			pair.GetERC20Contract(),
			func() {
				suite.app.Erc20Keeper.DeleteTokenPair(suite.ctx, pair)
			},
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate()

		found := suite.app.Erc20Keeper.IsERC20Registered(suite.ctx, tc.erc20)

		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestIsDenomRegistered() {
	addr := tests.GenerateAddress()
	pair := types.NewTokenPair(addr, "coin", true, types.OWNER_MODULE)
	suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair)
	suite.app.Erc20Keeper.SetERC20Map(suite.ctx, addr, pair.GetID())
	suite.app.Erc20Keeper.SetDenomMap(suite.ctx, pair.Denom, pair.GetID())

	testCases := []struct {
		name     string
		denom    string
		malleate func()
		ok       bool
	}{
		{"empty denom", "", func() {}, false},
		{"valid denom", pair.GetDenom(), func() {}, true},
		{
			"deleted denom map",
			pair.GetDenom(),
			func() {
				suite.app.Erc20Keeper.DeleteTokenPair(suite.ctx, pair)
			},
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate()

		found := suite.app.Erc20Keeper.IsDenomRegistered(suite.ctx, tc.denom)

		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}
