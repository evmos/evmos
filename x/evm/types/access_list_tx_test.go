package types_test

import (
	"math/big"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v11/x/evm/types"
)

func (suite *TxDataTestSuite) TestAccessListTxCopy() {
	tx := &types.AccessListTx{}
	txCopy := tx.Copy()

	suite.Require().Equal(&types.AccessListTx{}, txCopy)
}

func (suite *TxDataTestSuite) TestAccessListTxGetGasTipCap() {
	testCases := []struct {
		name string
		tx   types.AccessListTx
		exp  *big.Int
	}{
		{
			"non-empty gasPrice",
			types.AccessListTx{
				GasPrice: &suite.sdkInt,
			},
			(&suite.sdkInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetGasTipCap()

		suite.Require().Equal(tc.exp, actual, tc.name)
	}
}

func (suite *TxDataTestSuite) TestAccessListTxGetGasFeeCap() {
	testCases := []struct {
		name string
		tx   types.AccessListTx
		exp  *big.Int
	}{
		{
			"non-empty gasPrice",
			types.AccessListTx{
				GasPrice: &suite.sdkInt,
			},
			(&suite.sdkInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetGasFeeCap()

		suite.Require().Equal(tc.exp, actual, tc.name)
	}
}

func (suite *TxDataTestSuite) TestEmptyAccessList() {
	testCases := []struct {
		name string
		tx   types.AccessListTx
	}{
		{
			"empty access list tx",
			types.AccessListTx{
				Accesses: nil,
			},
		},
	}
	for _, tc := range testCases {
		actual := tc.tx.GetAccessList()

		suite.Require().Nil(actual, tc.name)
	}
}

func (suite *TxDataTestSuite) TestAccessListTxCost() {
	testCases := []struct {
		name string
		tx   types.AccessListTx
		exp  *big.Int
	}{
		{
			"non-empty access list tx",
			types.AccessListTx{
				GasPrice: &suite.sdkInt,
				GasLimit: uint64(1),
				Amount:   &suite.sdkZeroInt,
			},
			(&suite.sdkInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.Cost()

		suite.Require().Equal(tc.exp, actual, tc.name)
	}
}

func (suite *TxDataTestSuite) TestAccessListEffectiveGasPrice() {
	testCases := []struct {
		name    string
		tx      types.AccessListTx
		baseFee *big.Int
	}{
		{
			"non-empty access list tx",
			types.AccessListTx{
				GasPrice: &suite.sdkInt,
			},
			(&suite.sdkInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.EffectiveGasPrice(tc.baseFee)

		suite.Require().Equal(tc.tx.GetGasPrice(), actual, tc.name)
	}
}

func (suite *TxDataTestSuite) TestAccessListTxEffectiveCost() {
	testCases := []struct {
		name    string
		tx      types.AccessListTx
		baseFee *big.Int
		exp     *big.Int
	}{
		{
			"non-empty access list tx",
			types.AccessListTx{
				GasPrice: &suite.sdkInt,
				GasLimit: uint64(1),
				Amount:   &suite.sdkZeroInt,
			},
			(&suite.sdkInt).BigInt(),
			(&suite.sdkInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.EffectiveCost(tc.baseFee)

		suite.Require().Equal(tc.exp, actual, tc.name)
	}
}

func (suite *TxDataTestSuite) TestAccessListTxType() {
	testCases := []struct {
		name string
		tx   types.AccessListTx
	}{
		{
			"non-empty access list tx",
			types.AccessListTx{},
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.TxType()

		suite.Require().Equal(uint8(ethtypes.AccessListTxType), actual, tc.name)
	}
}
