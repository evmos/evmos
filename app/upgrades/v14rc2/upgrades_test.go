package v14rc2_test

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v13/app/upgrades/v14rc2"
	"github.com/evmos/evmos/v13/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v13/testutil"
	testutiltx "github.com/evmos/evmos/v13/testutil/tx"
	"github.com/evmos/evmos/v13/x/vesting/types"
)

var (
	// zeroDec is a zero decimal value
	zeroDec = sdk.ZeroDec()
)

func (s *UpgradesTestSuite) TestUpdateVestingFunders() {
	s.SetupTest()

	// Fund the affected accounts to initialize them and then create vesting accounts
	for address, oldFunder := range v14rc2.AffectedAddresses {
		accAddr := sdk.MustAccAddressFromBech32(address)
		err := testutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, accAddr, 1000)
		s.Require().NoError(err, "failed to fund account %s", address)

		// Create vesting account
		createMsg := &types.MsgCreateClawbackVestingAccount{
			FunderAddress:  oldFunder,
			VestingAddress: address,
		}
		_, err = s.app.VestingKeeper.CreateClawbackVestingAccount(sdk.UnwrapSDKContext(s.ctx), createMsg)
		s.Require().NoError(err, "failed to create vesting account for %s", address)
	}

	// Run the upgrade function
	err := v14rc2.UpdateVestingFunders(s.ctx, s.app.VestingKeeper)
	s.Require().NoError(err, "failed to update vesting funders")

	// Check that the vesting accounts have been updated
	for address := range v14rc2.AffectedAddresses {
		accAddr := sdk.MustAccAddressFromBech32(address)
		acc := s.app.AccountKeeper.GetAccount(s.ctx, accAddr)
		s.Require().NotNil(acc, "account not found for %s", address)
		vestingAcc, ok := acc.(*types.ClawbackVestingAccount)
		s.Require().True(ok, "account is not a vesting account for %s", address)
		s.Require().Equal(address, vestingAcc.Address, "expected different address in vesting account for %s", address)

		// Check that the funder has been updated
		s.Require().Equal(v14rc2.NewTeamMultisigAcc.String(), vestingAcc.FunderAddress, "expected different funder address for %s", address)
	}
}

func (s *UpgradesTestSuite) TestUpdateMigrateNativeMultisigs() {
	s.SetupTest()

	stakeDenom := s.app.StakingKeeper.BondDenom(s.ctx)
	stakeAmount := int64(1e17)
	stakeInt := sdk.NewInt(stakeAmount)
	stakeCoin := sdk.NewCoin(stakeDenom, stakeInt)

	// Create staking helper
	stakingHelper := teststaking.NewHelper(s.T(), s.ctx, s.app.StakingKeeper)
	stakingHelper.Commission = stakingtypes.NewCommissionRates(zeroDec, zeroDec, zeroDec)
	stakingHelper.Denom = stakeDenom

	// Create validator
	valAccAddr, valPriv := testutiltx.NewAccAddressAndKey()
	valAddr := sdk.ValAddress(valAccAddr)
	err := testutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, valAccAddr, stakeAmount)
	s.Require().NoError(err, "failed to fund validator account")
	stakingHelper.CreateValidator(valAddr, valPriv.PubKey(), stakeInt, true)

	val := s.app.StakingKeeper.Validator(s.ctx, sdk.ValAddress(valPriv.PubKey().Address()))
	s.Require().NotNil(val, "validator not found")
	validator, ok := val.(stakingtypes.Validator)
	s.Require().True(ok, "validator is not a staking validator")

	var affectedAccounts = make(map[*ethsecp256k1.PrivKey]sdk.AccAddress, 3)
	for idx := 0; idx < 3; idx++ {
		accAddr, priv := testutiltx.NewAccAddressAndKey()
		affectedAccounts[priv] = accAddr
	}

	fmt.Println("chain ID: ", s.ctx.ChainID())

	// Fund the affected accounts to initialize them and then create delegations
	var oldMultisigs = make([]string, 0, len(affectedAccounts))
	for priv, oldMultisig := range affectedAccounts {
		oldMultisigs = append(oldMultisigs, oldMultisig.String())
		fmt.Println("Current addr: ", oldMultisig.String())
		err := testutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, oldMultisig, stakeAmount)
		s.Require().NoError(err, "failed to fund account %s", oldMultisig.String())

		acc := s.app.AccountKeeper.GetAccount(s.ctx, oldMultisig)
		s.Require().NotNil(acc, "account not found for %s", oldMultisig.String())

		res, err := testutil.Delegate(s.ctx, s.app, priv, stakeCoin, validator)
		s.Require().NoError(err, "failed to delegate to validator %s", val.GetOperator())
		s.Require().True(res.IsOK(), "failed to delegate to validator %s", val.GetOperator())
	}

	err = v14rc2.MigrateNativeMultisigs(s.ctx, s.app.StakingKeeper, oldMultisigs)
	s.Require().NoError(err, "failed to migrate native multisigs")

	// Check that the multisigs have been updated
	for _, oldMultisig := range v14rc2.OldMultisigs {
		fmt.Println("old multisig: ", oldMultisig)
	}
}
