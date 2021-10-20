package keeper_test

import (
	"github.com/tharsis/ethermint/tests"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

func (suite *KeeperTestSuite) TestGetTokenPairID() {
	pair := types.NewTokenPair(tests.GenerateAddress(), evmtypes.DefaultEVMDenom, true)
	suite.app.IntrarelayerKeeper.SetTokenPair(suite.ctx, pair)

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
		id := suite.app.IntrarelayerKeeper.GetTokenPairID(suite.ctx, tc.token)
		if id != nil {
			suite.Require().Equal(tc.expId, id, tc.name)
		} else {
			suite.Require().Nil(id)
		}
	}
}

func (suite *KeeperTestSuite) TestGetTokenPair() {
	pair := types.NewTokenPair(tests.GenerateAddress(), evmtypes.DefaultEVMDenom, true)
	suite.app.IntrarelayerKeeper.SetTokenPair(suite.ctx, pair)

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
		p, found := suite.app.IntrarelayerKeeper.GetTokenPair(suite.ctx, tc.id)
		if tc.ok {
			suite.Require().True(found, tc.name)
			suite.Require().Equal(pair, p, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}
