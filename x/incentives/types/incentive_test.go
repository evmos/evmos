package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/evmos/evmos/v11/tests"
)

type IncentiveTestSuite struct {
	suite.Suite
}

func TestIncentiveSuite(t *testing.T) {
	suite.Run(t, new(IncentiveTestSuite))
}

func (suite *IncentiveTestSuite) TestIncentiveNew() {
	testCases := []struct {
		name        string
		contract    common.Address
		allocations sdk.DecCoins
		epochs      uint32
		expectPass  bool
	}{
		{
			"Register incentive - pass",
			tests.GenerateAddress(),
			sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
			10,
			true,
		},
		{
			"Register incentive - empty allocation",
			tests.GenerateAddress(),
			sdk.DecCoins{},
			10,
			false,
		},
		{
			"Register incentive - invalid allocation denom",
			tests.GenerateAddress(),
			sdk.DecCoins{{Denom: "(photon", Amount: sdk.OneDec()}},
			10,
			false,
		},
		{
			"Register incentive - invalid allocation amount (0)",
			tests.GenerateAddress(),
			sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(0, 2))},
			10,
			false,
		},
		{
			"Register incentive - invalid allocation amount (> 1)",
			tests.GenerateAddress(),
			sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(101, 2))},
			10,
			false,
		},
		{
			"Register incentive - zero epochs",
			tests.GenerateAddress(),
			sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
			0,
			false,
		},
	}

	for _, tc := range testCases {
		i := NewIncentive(tc.contract, tc.allocations, tc.epochs)
		err := i.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *IncentiveTestSuite) TestIncentive() {
	testCases := []struct {
		msg        string
		incentive  Incentive
		expectPass bool
	}{
		{
			"Register incentive - invalid address (no hex)",
			Incentive{
				"0x5dCA2483280D9727c80b5518faC4556617fb19ZZ",
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				10,
				time.Now(),
				0,
			},
			false,
		},
		{
			"Register incentive - invalid address (invalid length 1)",
			Incentive{
				"0x5dCA2483280D9727c80b5518faC4556617fb19",
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				10,
				time.Now(),
				0,
			},
			false,
		},
		{
			"Register incentive - invalid address (invalid length 2)",
			Incentive{
				"0x5dCA2483280D9727c80b5518faC4556617fb194FFF",
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				10,
				time.Now(),
				0,
			},
			false,
		},
		{
			"pass",
			Incentive{
				tests.GenerateAddress().String(),
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				10,
				time.Now(),
				0,
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.incentive.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.msg)
		} else {
			suite.Require().Error(err, tc.msg)
		}
	}
}

func (suite *IncentiveTestSuite) TestIsActive() {
	testCases := []struct {
		name       string
		incentive  Incentive
		expectPass bool
	}{
		{
			"pass",
			Incentive{
				tests.GenerateAddress().String(),
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				10,
				time.Now(),
				0,
			},
			true,
		},
		{
			"epoch is zero",
			Incentive{
				tests.GenerateAddress().String(),
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				0,
				time.Now(),
				0,
			},
			false,
		},
	}
	for _, tc := range testCases {
		res := tc.incentive.IsActive()
		if tc.expectPass {
			suite.Require().True(res, tc.name)
		} else {
			suite.Require().False(res, tc.name)
		}
	}
}
