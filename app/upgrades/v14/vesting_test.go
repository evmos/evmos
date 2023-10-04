package v14_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v14/app/upgrades/v14"
	"github.com/evmos/evmos/v14/testutil"
	"github.com/evmos/evmos/v14/x/vesting/types"
)

func (s *UpgradesTestSuite) TestUpdateVestingFunders() {
	s.SetupTest()

	// Fund the affected accounts to initialize them and then create vesting accounts
	s.prepareVestingAccount(v14.VestingAddrByFunder1, v14.OldFunder1)
	for _, address := range v14.VestingAddrsByFunder2 {
		s.prepareVestingAccount(address, v14.OldFunder2)
	}

	// Run the upgrade function
	err := v14.UpdateVestingFunders(s.ctx, s.app.VestingKeeper, v14.NewTeamPremintWalletAcc)
	s.Require().NoError(err, "failed to update vesting funders")

	// Check that the vesting accounts have been updated
	affectedAddrs := v14.VestingAddrsByFunder2
	affectedAddrs = append(affectedAddrs, v14.VestingAddrByFunder1)
	for _, address := range affectedAddrs {
		accAddr := sdk.MustAccAddressFromBech32(address)
		acc := s.app.AccountKeeper.GetAccount(s.ctx, accAddr)
		s.Require().NotNil(acc, "account not found for %s", address)
		vestingAcc, ok := acc.(*types.ClawbackVestingAccount)
		s.Require().True(ok, "account is not a vesting account for %s", address)
		s.Require().Equal(address, vestingAcc.Address, "expected different address in vesting account for %s", address)

		// Check that the funder has been updated
		s.Require().Equal(v14.NewTeamPremintWalletAcc.String(), vestingAcc.FunderAddress, "expected different funder address for %s", address)
	}
}

func (s *UpgradesTestSuite) prepareVestingAccount(address string, funder string) {
	accAddr := sdk.MustAccAddressFromBech32(address)
	err := testutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, accAddr, 1000)
	s.Require().NoError(err, "failed to fund account %s", address)

	// Create vesting account
	createMsg := &types.MsgCreateClawbackVestingAccount{
		FunderAddress:  funder,
		VestingAddress: address,
	}
	_, err = s.app.VestingKeeper.CreateClawbackVestingAccount(sdk.UnwrapSDKContext(s.ctx), createMsg)
	s.Require().NoError(err, "failed to create vesting account for %s", address)
}
