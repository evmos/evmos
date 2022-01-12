package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type AllocationMeterTestSuite struct {
	suite.Suite
}

func TestAllocationMeterSuite(t *testing.T) {
	suite.Run(t, new(AllocationMeterTestSuite))
}

func (suite *AllocationMeterTestSuite) TestAllocationMeterNew() {
	testCases := []struct {
		name       string
		allocation sdk.DecCoin
		expectPass bool
	}{
		{
			"Register AllocationMeter - pass",
			sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2)),
			true,
		},
		{
			"Register AllocationMeter - empty allocation",
			sdk.DecCoin{},
			false,
		},
		{
			"Register AllocationMeter - invalid allocation denom",
			sdk.DecCoin{Denom: "(photon", Amount: sdk.OneDec()},
			false,
		},
		{
			"Register AllocationMeter - invalid allocation amount (0)",
			sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(0, 2)),
			false,
		},
		{
			"Register AllocationMeter - invalid allocation amount (> 1)",
			sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(101, 2)),
			false,
		},
	}

	for _, tc := range testCases {
		i := NewAllocationMeter(tc.allocation)
		err := i.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *AllocationMeterTestSuite) TestAllocationMeter() {
	testCases := []struct {
		msg             string
		AllocationMeter AllocationMeter
		expectPass      bool
	}{
		{
			"pass",
			AllocationMeter{
				sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2)),
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.AllocationMeter.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.msg)
		} else {
			suite.Require().Error(err, tc.msg)
		}
	}
}
