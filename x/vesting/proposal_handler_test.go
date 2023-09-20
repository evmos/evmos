package vesting_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/evmos/evmos/v14/testutil"
	testutiltx "github.com/evmos/evmos/v14/testutil/tx"
	"github.com/evmos/evmos/v14/x/vesting"
	vestingtypes "github.com/evmos/evmos/v14/x/vesting/types"
)

var (
	vestingAmount  = sdk.Coins{sdk.NewInt64Coin("test", 1000)}
	examplePeriods = sdkvesting.Periods{
		{Length: 100, Amount: vestingAmount},
	}
)

func (s *VestingTestSuite) TestHandleClawbackProposal() {
	funderAddr := sdk.AccAddress(testutiltx.GenerateAddress().Bytes())
	notFoundAddr := sdk.AccAddress(testutiltx.GenerateAddress().Bytes())
	vestingAddr := sdk.AccAddress(testutiltx.GenerateAddress().Bytes())

	testcases := []struct {
		name        string
		vestingAddr string
		destAddr    string
		// skipCreating specifies if the creation of the vesting account should be skipped
		skipCreating bool
		// skipFunding specifies if the funding of the vesting account should be skipped
		skipFunding       bool
		enableGovClawback bool
		// postCheck will be executed after the test logic is run to check the correct
		// state changes have been made
		postCheck   func(proposal vestingtypes.ClawbackProposal)
		expPass     bool
		errContains string
	}{
		{
			name:              "pass - claw back without destination defined",
			vestingAddr:       vestingAddr.String(),
			enableGovClawback: true,
			expPass:           true,
			postCheck: func(proposal vestingtypes.ClawbackProposal) {
				_, err := s.app.VestingKeeper.Balances(s.ctx, &vestingtypes.QueryBalancesRequest{Address: proposal.Address})
				s.Require().Error(err, "expected error when querying balances of vesting account")
				s.Require().ErrorContains(err, "either does not exist or is not a vesting account", "account should not be a vesting account anymore")
			},
		},
		{
			name:              "pass - claw back with destination defined",
			vestingAddr:       vestingAddr.String(),
			destAddr:          testutiltx.GenerateAddress().String(),
			enableGovClawback: true,
			expPass:           true,
			postCheck: func(proposal vestingtypes.ClawbackProposal) {
				_, err := s.app.VestingKeeper.Balances(s.ctx, &vestingtypes.QueryBalancesRequest{Address: proposal.Address})
				s.Require().Error(err, "expected error when querying balances of vesting account")
				s.Require().ErrorContains(err, "either does not exist or is not a vesting account", "account should not be a vesting account anymore")
			},
		},
		{
			name:         "fail - not a vesting account",
			vestingAddr:  vestingAddr.String(),
			skipCreating: true,
			errContains:  vestingtypes.ErrNotSubjectToClawback.Error(),
			postCheck:    func(proposal vestingtypes.ClawbackProposal) {},
		},
		{
			name:              "fail - vesting account only initialized (no schedules)",
			vestingAddr:       vestingAddr.String(),
			skipFunding:       true,
			enableGovClawback: true,
			errContains:       "has no vesting or lockup periods",
			postCheck:         func(proposal vestingtypes.ClawbackProposal) {},
		},
		{
			name:              "fail - account does not exist",
			vestingAddr:       notFoundAddr.String(),
			skipCreating:      true,
			enableGovClawback: true,
			errContains:       fmt.Sprintf("account at address '%s' does not exist", notFoundAddr.String()),
			postCheck:         func(proposal vestingtypes.ClawbackProposal) {},
		},
		{
			name:              "fail - gov clawback not enabled",
			vestingAddr:       vestingAddr.String(),
			destAddr:          "",
			enableGovClawback: false,
			errContains:       vestingtypes.ErrNotSubjectToGovClawback.Error(),
			postCheck: func(proposal vestingtypes.ClawbackProposal) {
				balances, err := s.app.VestingKeeper.Balances(s.ctx, &vestingtypes.QueryBalancesRequest{Address: proposal.Address})
				s.Require().NoError(err, "expected no error when querying balances of vesting account")
				s.Require().Equal(vestingAmount, balances.GetUnvested(), "expected full vesting amount still unvested")
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Fund vesting account to initialize it and then send all coins to funder account
			err := testutil.FundAccount(s.ctx, s.app.BankKeeper, vestingAddr, vestingAmount)
			s.Require().NoError(err, "failed to fund account")
			sentBalances := s.app.BankKeeper.GetAllBalances(s.ctx, vestingAddr)
			err = s.app.BankKeeper.SendCoins(s.ctx, vestingAddr, funderAddr, sentBalances)
			s.Require().NoError(err, "failed to send coins to funder")

			if !tc.skipCreating {
				msgCreate := &vestingtypes.MsgCreateClawbackVestingAccount{
					FunderAddress:     funderAddr.String(),
					VestingAddress:    vestingAddr.String(),
					EnableGovClawback: tc.enableGovClawback,
				}
				_, err = s.app.VestingKeeper.CreateClawbackVestingAccount(s.ctx, msgCreate)
				s.Require().NoError(err, "failed to create vesting account")

				if !tc.skipFunding {
					msgFund := &vestingtypes.MsgFundVestingAccount{
						FunderAddress:  funderAddr.String(),
						VestingAddress: vestingAddr.String(),
						StartTime:      s.ctx.BlockTime(),
						LockupPeriods:  examplePeriods,
						VestingPeriods: examplePeriods,
					}
					_, err = s.app.VestingKeeper.FundVestingAccount(s.ctx, msgFund)
					s.Require().NoError(err, "failed to fund vesting account")
				}
			}

			proposal := vestingtypes.ClawbackProposal{
				Title:              "Clawback Test Proposal",
				Description:        "Test Description",
				Address:            tc.vestingAddr,
				DestinationAddress: tc.destAddr,
			}
			err = vesting.HandleClawbackProposal(s.ctx, &s.app.VestingKeeper, &proposal)
			if tc.expPass {
				s.Require().NoError(err, "failed to handle clawback proposal")
				tc.postCheck(proposal)
			} else {
				s.Require().Error(err, "expected to fail handling clawback proposal")
				s.Require().ErrorContains(err, tc.errContains, "expected error message to contain")
			}
		})
	}
}
