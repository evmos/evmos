package v14rc2_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	doubleStakeCoin := sdk.NewCoin(stakeDenom, stakeInt.MulRaw(2))
	nAccounts := 1

	affectedAccounts := make(map[*ethsecp256k1.PrivKey]sdk.AccAddress, nAccounts)
	for idx := 0; idx < nAccounts; idx++ {
		accAddr, priv := testutiltx.NewAccAddressAndKey()
		affectedAccounts[priv] = accAddr
	}

	s.NextBlock()

	var (
		migratedBalances sdk.Coins
		oldMultisigs     = make([]string, 0, len(affectedAccounts))
	)

	// Fund the affected accounts to initialize them and then create delegations
	for priv, oldMultisig := range affectedAccounts {
		oldMultisigs = append(oldMultisigs, oldMultisig.String())
		err := testutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, oldMultisig, 10*stakeAmount)
		s.Require().NoError(err, "failed to fund account %s", oldMultisig.String())

		_, err = testutil.Delegate(s.ctx, s.app, priv, stakeCoin, s.validators[0])
		s.Require().NoError(err, "failed to delegate to validator %s", s.validators[0].GetOperator())
		_, err = testutil.Delegate(s.ctx, s.app, priv, doubleStakeCoin, s.validators[1])
		s.Require().NoError(err, "failed to delegate to validator %s", s.validators[1].GetOperator())

		balances := s.app.BankKeeper.GetAllBalances(s.ctx, oldMultisig)
		migratedBalances = migratedBalances.Add(balances...)
	}

	// Check there are no prior delegations for new team multisig
	delegations := s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, v14rc2.NewTeamMultisigAcc)
	s.Require().Len(delegations, 0, "expected no delegations for account %s", v14rc2.NewTeamMultisigAcc.String())

	// Check validator shares before migration
	allValidators := s.app.StakingKeeper.GetAllValidators(s.ctx)
	expectedSharesMap := make(map[string]sdk.Dec, len(allValidators))
	for _, validator := range allValidators {
		expectedSharesMap[validator.OperatorAddress] = validator.DelegatorShares
	}

	err := v14rc2.MigrateNativeMultisigs(s.ctx, s.app.BankKeeper, s.app.StakingKeeper, oldMultisigs)
	s.Require().NoError(err, "failed to migrate native multisigs")

	// Check that the multisigs have been updated
	for _, oldMultisig := range affectedAccounts {
		delegations := s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, oldMultisig)
		s.Require().Len(delegations, 0, "expected no delegations after migration for account %s", oldMultisig.String())
		unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, oldMultisig)
		s.Require().Len(unbondingDelegations, 0, "expected no unbonding delegations after migration for account %s", oldMultisig.String())
		balances := s.app.BankKeeper.GetAllBalances(s.ctx, oldMultisig)
		s.Require().Len(balances, 0, "expected no balance after migration for account %s", oldMultisig.String())
	}

	// Check that the new multisig has the corresponding delegations
	delegations = s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, v14rc2.NewTeamMultisigAcc)
	s.Require().True(len(delegations) > 0, "expected delegations after migration for account %s", v14rc2.NewTeamMultisigAcc.String())
	totalBalances := s.app.BankKeeper.GetAllBalances(s.ctx, v14rc2.NewTeamMultisigAcc)
	s.Require().Equal(migratedBalances, totalBalances, "expected different balance for target account %s", v14rc2.NewTeamMultisigAcc.String())

	// Check validator shares after migration
	allValidators = s.app.StakingKeeper.GetAllValidators(s.ctx)
	sharesMap := make(map[string]sdk.Dec, len(allValidators))
	for _, validator := range allValidators {
		sharesMap[validator.OperatorAddress] = validator.DelegatorShares
	}
	s.Require().Equal(expectedSharesMap, sharesMap, "expected different validator shares after migration")
}
