package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	utiltx "github.com/evmos/evmos/v11/testutil/tx"
	"github.com/evmos/evmos/v11/x/incentives/types"
)

type ProposalTestSuite struct {
	suite.Suite
}

func TestProposalTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalTestSuite))
}

func (suite *ProposalTestSuite) TestKeysTypes() {
	suite.Require().Equal("incentives", (&types.RegisterIncentiveProposal{}).ProposalRoute())
	suite.Require().Equal("RegisterIncentive", (&types.RegisterIncentiveProposal{}).ProposalType())
	suite.Require().Equal("incentives", (&types.CancelIncentiveProposal{}).ProposalRoute())
	suite.Require().Equal("CancelIncentive", (&types.CancelIncentiveProposal{}).ProposalType())
}

func (suite *ProposalTestSuite) TestRegisterIncentiveProposal() {
	testCases := []struct {
		name        string
		title       string
		description string
		incentive   types.Incentive
		expectPass  bool
	}{
		{
			"Register incentive - valid",
			"test",
			"test desc",
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
			"Register incentive - empty allocations",
			"test",
			"test desc",
			types.Incentive{
				utiltx.GenerateAddress().String(),
				sdk.DecCoins{},
				10,
				time.Now(),
				0,
			},
			false,
		},
		{
			"Register incentive - invalid missing title ",
			"",
			"test desc",
			types.Incentive{
				utiltx.GenerateAddress().String(),
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				10,
				time.Now(),
				0,
			},
			false,
		},
		{
			"Register incentive - invalid missing description ",
			"test",
			"",
			types.Incentive{
				utiltx.GenerateAddress().String(),
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				10,
				time.Now(),
				0,
			},
			false,
		},
		{
			"Register incentive - invalid address (no hex)",
			"test",
			"test desc",
			types.Incentive{
				"",
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				10,
				time.Now(),
				0,
			},
			false,
		},
		{
			"Register incentive - invalid address (invalid length 1)",
			"test",
			"test desc",
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
			"test",
			"test desc",
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
			"Register incentive - invalid allocation amount >100% ",
			"test",
			"test desc",
			types.Incentive{
				utiltx.GenerateAddress().String(),
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(101, 2))},
				10,
				time.Now(),
				0,
			},
			false,
		},
		{
			"Register incentive - invalid allocation amount 0%",
			"test",
			"test desc",
			types.Incentive{
				utiltx.GenerateAddress().String(),
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(0, 2))},
				10,
				time.Now(),
				0,
			},
			false,
		},
		{
			"Register incentive - zero epochs",
			"test",
			"test desc",
			types.Incentive{
				utiltx.GenerateAddress().String(),
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				0,
				time.Now(),
				0,
			},
			false,
		},
		{
			"Register incentive - invalid allocation amount 0%",
			"test",
			"test desc",
			types.Incentive{
				utiltx.GenerateAddress().String(),
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(0, 2))},
				10,
				time.Now(),
				0,
			},
			false,
		},
	}
	for _, tc := range testCases {
		tx := types.NewRegisterIncentiveProposal(
			tc.title,
			tc.description,
			tc.incentive.Contract,
			tc.incentive.Allocations,
			tc.incentive.Epochs,
		)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *ProposalTestSuite) TestCancelIncentiveProposal() {
	testCases := []struct {
		name        string
		title       string
		description string
		incentive   types.Incentive
		expectPass  bool
	}{
		{
			"Cancel incentive - valid",
			"test",
			"test desc",
			types.Incentive{
				utiltx.GenerateAddress().String(),
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				5,
				time.Now(),
				0,
			},
			true,
		},
		{
			"Cancel incentive - invalid missing title ",
			"",
			"test desc",
			types.Incentive{
				utiltx.GenerateAddress().String(),
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				10,
				time.Now(),
				0,
			},
			false,
		},
		{
			"Cancel incentive - invalid missing description ",
			"test",
			"",
			types.Incentive{
				utiltx.GenerateAddress().String(),
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				10,
				time.Now(),
				0,
			},
			false,
		},
		{
			"Cancel incentive - invalid address (no hex)",
			"test",
			"test desc",
			types.Incentive{
				"035dCA2483280D9727c80b5518faC4556617fb19ZZ",
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				10,
				time.Now(),
				0,
			},
			false,
		},
		{
			"Cancel incentive - invalid address (invalid length 1)",
			"test",
			"test desc",
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
			"Cancel incentive - invalid address (invalid length 2)",
			"test",
			"test desc",
			types.Incentive{
				"0x5dCA2483280D9727c80b5518faC4556617fb194FFF",
				sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
				10,
				time.Now(),
				0,
			},
			false,
		},
	}
	for _, tc := range testCases {
		tx := types.NewCancelIncentiveProposal(
			tc.title,
			tc.description,
			tc.incentive.Contract,
		)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
