package types_test

import (
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v15/x/evm/types"
)

func (suite *TxDataTestSuite) TestTestNewAccessList() {
	testCases := []struct {
		name          string
		ethAccessList *ethtypes.AccessList
		expAl         types.AccessList
	}{
		{
			"ethAccessList is nil",
			nil,
			nil,
		},
		{
			"non-empty ethAccessList",
			&ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
			types.AccessList{{Address: suite.hexAddr, StorageKeys: []string{common.Hash{}.Hex()}}},
		},
	}
	for _, tc := range testCases {
		al := types.NewAccessList(tc.ethAccessList)

		suite.Require().Equal(tc.expAl, al)
	}
}

func (suite *TxDataTestSuite) TestAccessListToEthAccessList() {
	ethAccessList := ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}}
	al := types.NewAccessList(&ethAccessList)
	actual := al.ToEthAccessList()

	suite.Require().Equal(&ethAccessList, actual)
}
