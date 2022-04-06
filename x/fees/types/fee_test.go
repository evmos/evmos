package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/tharsis/ethermint/tests"
)

type FeeTestSuite struct {
	suite.Suite
}

func TestFeeSuite(t *testing.T) {
	suite.Run(t, new(FeeTestSuite))
}

func (suite *FeeTestSuite) TestDevFeeInfoNew() {
	address1 := sdk.AccAddress(tests.GenerateAddress().Bytes())
	address2 := sdk.AccAddress(tests.GenerateAddress().Bytes())
	testCases := []struct {
		name       string
		contract   common.Address
		deployer   sdk.AccAddress
		withdraw   sdk.AccAddress
		expectPass bool
	}{
		{
			"Create fee info - pass",
			tests.GenerateAddress(),
			address1,
			address2,
			true,
		},
		{
			"Create fee info, omit withdraw - pass",
			tests.GenerateAddress(),
			address1,
			nil,
			true,
		},
		{
			"Create fee info - invalid contract address",
			common.Address{},
			address1,
			address2,
			false,
		},
		{
			"Create fee info - invalid deployer address",
			tests.GenerateAddress(),
			sdk.AccAddress{},
			address2,
			false,
		},
	}

	for _, tc := range testCases {
		i := NewDevFeeInfo(tc.contract, tc.deployer, tc.withdraw)
		err := i.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *FeeTestSuite) TestFee() {
	address1 := sdk.AccAddress(tests.GenerateAddress().Bytes()).String()
	address2 := sdk.AccAddress(tests.GenerateAddress().Bytes()).String()
	testCases := []struct {
		msg        string
		feeInfo    DevFeeInfo
		expectPass bool
	}{
		{
			"Create fee info - pass",
			DevFeeInfo{
				tests.GenerateAddress().String(),
				address1,
				address2,
			},
			true,
		},
		{
			"Create fee info - invalid contract address (not hex)",
			DevFeeInfo{
				"0x5dCA2483280D9727c80b5518faC4556617fb19ZZ",
				address1,
				address2,
			},
			false,
		},
		{
			"Create fee info - invalid contract address (invalid length 1)",
			DevFeeInfo{
				"0x5dCA2483280D9727c80b5518faC4556617fb19",
				address1,
				address2,
			},
			false,
		},
		{
			"Create fee info - invalid contract address (invalid length 2)",
			DevFeeInfo{
				"0x5dCA2483280D9727c80b5518faC4556617fb194FFF",
				address1,
				address2,
			},
			false,
		},
		{
			"Create fee info - invalid deployer address",
			DevFeeInfo{
				tests.GenerateAddress().String(),
				"evmos14mq5c8yn9jx295ahaxye2f0xw3tlell0lt542Z",
				address2,
			},
			false,
		},
		{
			"Create fee info - invalid withdraw address",
			DevFeeInfo{
				tests.GenerateAddress().String(),
				address1,
				"evmos14mq5c8yn9jx295ahaxye2f0xw3tlell0lt542Z",
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.feeInfo.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.msg)
		} else {
			suite.Require().Error(err, tc.msg)
		}
	}
}
