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
	address1 sdk.AccAddress
	address2 sdk.AccAddress
}

func TestFeeSuite(t *testing.T) {
	suite.Run(t, new(FeeTestSuite))
}

func (suite *FeeTestSuite) SetupTest() {
	suite.address1 = sdk.AccAddress(tests.GenerateAddress().Bytes())
	suite.address2 = sdk.AccAddress(tests.GenerateAddress().Bytes())
}

func (suite *FeeTestSuite) TestFeeNew() {
	testCases := []struct {
		name       string
		contract   common.Address
		deployer   sdk.AccAddress
		withdraw   sdk.AccAddress
		expectPass bool
	}{
		{
			"Create fee - pass",
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
			"Create fee - invalid contract address",
			common.Address{},
			suite.address1,
			suite.address2,
			false,
		},
		{
			"Create fee - invalid deployer address",
			tests.GenerateAddress(),
			sdk.AccAddress{},
			suite.address2,
			false,
		},
	}

	for _, tc := range testCases {
		i := NewFee(tc.contract, tc.deployer, tc.withdraw)
		err := i.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *FeeTestSuite) TestFee() {
	testCases := []struct {
		msg        string
		fee        Fee
		expectPass bool
	}{
		{
			"Create fee - pass",
			Fee{
				tests.GenerateAddress().String(),
				suite.address1.String(),
				suite.address2.String(),
			},
			true,
		},
		{
			"Create fee - invalid contract address (not hex)",
			Fee{
				"0x5dCA2483280D9727c80b5518faC4556617fb19ZZ",
				suite.address1.String(),
				suite.address2.String(),
			},
			false,
		},
		{
			"Create fee - invalid contract address (invalid length 1)",
			Fee{
				"0x5dCA2483280D9727c80b5518faC4556617fb19",
				suite.address1.String(),
				suite.address2.String(),
			},
			false,
		},
		{
			"Create fee - invalid contract address (invalid length 2)",
			Fee{
				"0x5dCA2483280D9727c80b5518faC4556617fb194FFF",
				suite.address1.String(),
				suite.address2.String(),
			},
			false,
		},
		{
			"Create fee - invalid deployer address",
			Fee{
				tests.GenerateAddress().String(),
				"evmos14mq5c8yn9jx295ahaxye2f0xw3tlell0lt542Z",
				suite.address2.String(),
			},
			false,
		},
		{
			"Create fee - invalid withdraw address",
			Fee{
				tests.GenerateAddress().String(),
				suite.address1.String(),
				"evmos14mq5c8yn9jx295ahaxye2f0xw3tlell0lt542Z",
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.fee.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.msg)
		} else {
			suite.Require().Error(err, tc.msg)
		}
	}
}
