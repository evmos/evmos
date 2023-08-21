package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	utiltx "github.com/evmos/evmos/v14/testutil/tx"
	"github.com/evmos/evmos/v14/x/incentives/types"
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
			utiltx.GenerateAddress(),
			sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
			10,
			true,
		},
		{
			"Register incentive - empty allocation",
			utiltx.GenerateAddress(),
			sdk.DecCoins{},
			10,
			false,
		},
		{
			"Register incentive - invalid allocation denom",
			utiltx.GenerateAddress(),
			sdk.DecCoins{{Denom: "(evmos", Amount: sdk.OneDec()}},
			10,
			false,
		},
		{
			"Register incentive - invalid allocation amount (0)",
			utiltx.GenerateAddress(),
			sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(0, 2))},
			10,
			false,
		},
		{
			"Register incentive - invalid allocation amount (> 1)",
			utiltx.GenerateAddress(),
			sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(101, 2))},
			10,
			false,
		},
		{
			"Register incentive - zero epochs",
			utiltx.GenerateAddress(),
			sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
			0,
			false,
		},
	}

	for _, tc := range testCases {
		i := types.NewIncentive(tc.contract, tc.allocations, tc.epochs)
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
		incentive  types.Incentive
		expectPass bool
	}{
		{
			"Register incentive - invalid address (no hex)",
			types.Incentive{
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
			types.Incentive{
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
			types.Incentive{
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
			types.Incentive{
				utiltx.GenerateAddress().String(),
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
		incentive  types.Incentive
		expectPass bool
	}{
		{
			"pass",
			types.Incentive{
				utiltx.GenerateAddress().String(),
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				10,
				time.Now(),
				0,
			},
			true,
		},
		{
			"epoch is zero",
			types.Incentive{
				utiltx.GenerateAddress().String(),
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
