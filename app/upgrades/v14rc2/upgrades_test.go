package v14rc2_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	s.prepareVestingAccount(v14rc2.VestingAddrByFunder1, v14rc2.OldFunder1)
	for _, address := range v14rc2.VestingAddrsByFunder2 {
		s.prepareVestingAccount(address, v14rc2.OldFunder2)
	}

	// Run the upgrade function
	err := v14rc2.UpdateVestingFunders(s.ctx, s.app.VestingKeeper)
	s.Require().NoError(err, "failed to update vesting funders")

	// Check that the vesting accounts have been updated
	affectedAddrs := append(v14rc2.VestingAddrsByFunder2, v14rc2.VestingAddrByFunder1)
	for _, address := range affectedAddrs {
		accAddr := sdk.MustAccAddressFromBech32(address)
		acc := s.app.AccountKeeper.GetAccount(s.ctx, accAddr)
		s.Require().NotNil(acc, "account not found for %s", address)
		vestingAcc, ok := acc.(*types.ClawbackVestingAccount)
		s.Require().True(ok, "account is not a vesting account for %s", address)
		s.Require().Equal(address, vestingAcc.Address, "expected different address in vesting account for %s", address)

		// Check that the funder has been updated
		s.Require().Equal(v14rc2.NewTeamPremintWalletAcc.String(), vestingAcc.FunderAddress, "expected different funder address for %s", address)
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

func (s *UpgradesTestSuite) TestUpdateMigrateNativeMultisigs() {
	s.SetupTest()

	stakeDenom := s.app.StakingKeeper.BondDenom(s.ctx)
	stakeAmount := int64(1e17)
	stakeInt := sdk.NewInt(stakeAmount)
	stakeCoin := sdk.NewCoin(stakeDenom, stakeInt)
	doubleStakeCoin := sdk.NewCoin(stakeDenom, stakeInt.MulRaw(2))
	nAccounts := 3

	affectedAccounts := make(map[*ethsecp256k1.PrivKey]sdk.AccAddress, nAccounts)
	for idx := 0; idx < nAccounts; idx++ {
		accAddr, priv := testutiltx.NewAccAddressAndKey()
		affectedAccounts[priv] = accAddr
	}

	s.NextBlock()

	var (
		migratedBalances sdk.Coins
		migrationTarget  = v14rc2.NewTeamPremintWalletAcc
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
	delegations := s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, migrationTarget)
	s.Require().Len(delegations, 0, "expected no delegations for account %s", migrationTarget.String())

	// Check validator shares before migration
	expectedSharesMap := s.getDelegationSharesMap()

	err := v14rc2.MigrateNativeMultisigs(s.ctx, s.app.BankKeeper, s.app.StakingKeeper, oldMultisigs, migrationTarget)
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
	delegations = s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, migrationTarget)
	s.Require().Len(delegations, 2, "expected two delegations after migration for account %s", migrationTarget.String())
	totalBalances := s.app.BankKeeper.GetAllBalances(s.ctx, migrationTarget)
	s.Require().Equal(migratedBalances, totalBalances, "expected different balance for target account %s", migrationTarget.String())

	// Check validator shares after migration.
	// NOTE: They must be equal to guarantee that the voting power is unchanged before and after the migration.
	sharesMap := s.getDelegationSharesMap()
	s.Require().Equal(expectedSharesMap, sharesMap, "expected different validator shares after migration")
}

func (s *UpgradesTestSuite) TestInstantUnbonding() {
	balancePre := s.app.BankKeeper.GetAllBalances(s.ctx, s.address.Bytes())
	notBondedPool := s.app.AccountKeeper.GetModuleAccount(s.ctx, stakingtypes.NotBondedPoolName)
	poolBalancePre := s.app.BankKeeper.GetAllBalances(s.ctx, notBondedPool.GetAddress())
	delegation, found := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), s.validators[0].GetOperator())
	s.Require().True(found, "delegation not found")

	unbondAmount, err := v14rc2.InstantUnbonding(s.ctx, s.app.BankKeeper, s.app.StakingKeeper, delegation, s.bondDenom)
	s.Require().NoError(err, "failed to unbond")
	s.Require().Equal(unbondAmount, math.NewInt(1e18), "expected different unbond amount")

	expectedDiff := sdk.Coins{{Denom: s.bondDenom, Amount: unbondAmount}}
	balancePost := s.app.BankKeeper.GetAllBalances(s.ctx, s.address.Bytes())
	diff := balancePost.Sub(balancePre...)
	s.Require().Equal(expectedDiff, diff, "expected different balance diff")

	_, found = s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), s.validators[0].GetOperator())
	s.Require().False(found, "delegation should not be found")

	poolBalancePost := s.app.BankKeeper.GetAllBalances(s.ctx, notBondedPool.GetAddress())
	s.Require().Equal(poolBalancePre, poolBalancePost, "expected no change in pool balance")
}

// getDelegationSharesMap returns a map of validator operator addresses to the
// total shares delegated to them.
func (s *UpgradesTestSuite) getDelegationSharesMap() map[string]sdk.Dec {
	allValidators := s.app.StakingKeeper.GetAllValidators(s.ctx)
	sharesMap := make(map[string]sdk.Dec, len(allValidators))
	for _, validator := range allValidators {
		sharesMap[validator.OperatorAddress] = validator.DelegatorShares
	}
	return sharesMap
}
