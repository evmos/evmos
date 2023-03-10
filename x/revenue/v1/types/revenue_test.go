package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"

	utiltx "github.com/evmos/evmos/v12/testutil/tx"
	"github.com/evmos/evmos/v12/x/revenue/v1/types"

	"github.com/stretchr/testify/suite"
)

type RevenueTestSuite struct {
	suite.Suite
	address1 sdk.AccAddress
	address2 sdk.AccAddress
}

func TestRevenueSuite(t *testing.T) {
	suite.Run(t, new(RevenueTestSuite))
}

func (suite *RevenueTestSuite) SetupTest() {
	suite.address1 = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	suite.address2 = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
}

func (suite *RevenueTestSuite) TestFeeNew() {
	testCases := []struct {
		name       string
		contract   common.Address
		deployer   sdk.AccAddress
		withdraw   sdk.AccAddress
		expectPass bool
	}{
		{
			"Create revenue- pass",
			utiltx.GenerateAddress(),
			suite.address1,
			suite.address2,
			true,
		},
		{
			"Create fee, omit withdraw - pass",
			utiltx.GenerateAddress(),
			suite.address1,
			nil,
			true,
		},
		{
			"Create revenue- invalid contract address",
			common.Address{},
			suite.address1,
			suite.address2,
			false,
		},
		{
			"Create revenue- invalid deployer address",
			utiltx.GenerateAddress(),
			sdk.AccAddress{},
			suite.address2,
			false,
		},
	}

	for _, tc := range testCases {
		i := types.NewRevenue(tc.contract, tc.deployer, tc.withdraw)
		err := i.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *RevenueTestSuite) TestFee() {
	testCases := []struct {
		msg        string
		revenue    types.Revenue
		expectPass bool
	}{
		{
			"Create revenue- pass",
			types.Revenue{
				utiltx.GenerateAddress().String(),
				suite.address1.String(),
				suite.address2.String(),
			},
			true,
		},
		{
			"Create revenue- invalid contract address (not hex)",
			types.Revenue{
				"0x5dCA2483280D9727c80b5518faC4556617fb19ZZ",
				suite.address1.String(),
				suite.address2.String(),
			},
			false,
		},
		{
			"Create revenue- invalid contract address (invalid length 1)",
			types.Revenue{
				"0x5dCA2483280D9727c80b5518faC4556617fb19",
				suite.address1.String(),
				suite.address2.String(),
			},
			false,
		},
		{
			"Create revenue- invalid contract address (invalid length 2)",
			types.Revenue{
				"0x5dCA2483280D9727c80b5518faC4556617fb194FFF",
				suite.address1.String(),
				suite.address2.String(),
			},
			false,
		},
		{
			"Create revenue- invalid deployer address",
			types.Revenue{
				utiltx.GenerateAddress().String(),
				"evmos14mq5c8yn9jx295ahaxye2f0xw3tlell0lt542Z",
				suite.address2.String(),
			},
			false,
		},
		{
			"Create revenue- invalid withdraw address",
			types.Revenue{
				utiltx.GenerateAddress().String(),
				suite.address1.String(),
				"evmos14mq5c8yn9jx295ahaxye2f0xw3tlell0lt542Z",
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.revenue.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.msg)
		} else {
			suite.Require().Error(err, tc.msg)
		}
	}
}

func (suite *RevenueTestSuite) TestRevenueGetters() {
	contract := utiltx.GenerateAddress()
	fs := types.Revenue{
		contract.String(),
		suite.address1.String(),
		suite.address2.String(),
	}
	suite.Equal(fs.GetContractAddr(), contract)
	suite.Equal(fs.GetDeployerAddr(), suite.address1)
	suite.Equal(fs.GetWithdrawerAddr(), suite.address2)

	fs = types.Revenue{
		contract.String(),
		suite.address1.String(),
		"",
	}
	suite.Equal(fs.GetContractAddr(), contract)
	suite.Equal(fs.GetDeployerAddr(), suite.address1)
	suite.Equal(len(fs.GetWithdrawerAddr()), 0)
}
