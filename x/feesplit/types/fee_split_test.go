package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/evmos/ethermint/tests"
)

type FeeSplitTestSuite struct {
	suite.Suite
	address1 sdk.AccAddress
	address2 sdk.AccAddress
}

func TestFeeSplitSuite(t *testing.T) {
	suite.Run(t, new(FeeSplitTestSuite))
}

func (suite *FeeSplitTestSuite) SetupTest() {
	suite.address1 = sdk.AccAddress(tests.GenerateAddress().Bytes())
	suite.address2 = sdk.AccAddress(tests.GenerateAddress().Bytes())
}

func (suite *FeeSplitTestSuite) TestFeeNew() {
	testCases := []struct {
		name       string
		contract   common.Address
		deployer   sdk.AccAddress
		withdraw   sdk.AccAddress
		expectPass bool
	}{
		{
			"Create fee split- pass",
			tests.GenerateAddress(),
			suite.address1,
			suite.address2,
			true,
		},
		{
			"Create fee, omit withdraw - pass",
			tests.GenerateAddress(),
			suite.address1,
			nil,
			true,
		},
		{
			"Create fee split- invalid contract address",
			common.Address{},
			suite.address1,
			suite.address2,
			false,
		},
		{
			"Create fee split- invalid deployer address",
			tests.GenerateAddress(),
			sdk.AccAddress{},
			suite.address2,
			false,
		},
	}

	for _, tc := range testCases {
		i := NewFeeSplit(tc.contract, tc.deployer, tc.withdraw)
		err := i.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *FeeSplitTestSuite) TestFee() {
	testCases := []struct {
		msg        string
		feeSplit   FeeSplit
		expectPass bool
	}{
		{
			"Create fee split- pass",
			FeeSplit{
				tests.GenerateAddress().String(),
				suite.address1.String(),
				suite.address2.String(),
			},
			true,
		},
		{
			"Create fee split- invalid contract address (not hex)",
			FeeSplit{
				"0x5dCA2483280D9727c80b5518faC4556617fb19ZZ",
				suite.address1.String(),
				suite.address2.String(),
			},
			false,
		},
		{
			"Create fee split- invalid contract address (invalid length 1)",
			FeeSplit{
				"0x5dCA2483280D9727c80b5518faC4556617fb19",
				suite.address1.String(),
				suite.address2.String(),
			},
			false,
		},
		{
			"Create fee split- invalid contract address (invalid length 2)",
			FeeSplit{
				"0x5dCA2483280D9727c80b5518faC4556617fb194FFF",
				suite.address1.String(),
				suite.address2.String(),
			},
			false,
		},
		{
			"Create fee split- invalid deployer address",
			FeeSplit{
				tests.GenerateAddress().String(),
				"evmos14mq5c8yn9jx295ahaxye2f0xw3tlell0lt542Z",
				suite.address2.String(),
			},
			false,
		},
		{
			"Create fee split- invalid withdraw address",
			FeeSplit{
				tests.GenerateAddress().String(),
				suite.address1.String(),
				"evmos14mq5c8yn9jx295ahaxye2f0xw3tlell0lt542Z",
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.feeSplit.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.msg)
		} else {
			suite.Require().Error(err, tc.msg)
		}
	}
}

func (suite *FeeSplitTestSuite) TestFeeSplitGetters() {
	contract := tests.GenerateAddress()
	fs := FeeSplit{
		contract.String(),
		suite.address1.String(),
		suite.address2.String(),
	}
	suite.Equal(fs.GetContractAddr(), contract)
	suite.Equal(fs.GetDeployerAddr(), suite.address1)
	suite.Equal(fs.GetWithdrawerAddr(), suite.address2)

	fs = FeeSplit{
		contract.String(),
		suite.address1.String(),
		"",
	}
	suite.Equal(fs.GetContractAddr(), contract)
	suite.Equal(fs.GetDeployerAddr(), suite.address1)
	suite.Equal(len(fs.GetWithdrawerAddr()), 0)
}
