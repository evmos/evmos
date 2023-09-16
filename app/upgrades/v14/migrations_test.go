package v14_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v14/app/upgrades/v14"
	"github.com/evmos/evmos/v14/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v14/testutil"
	testutiltx "github.com/evmos/evmos/v14/testutil/tx"
)

// zeroDec is a zero decimal value
var zeroDec = sdk.ZeroDec()

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
		migrationTarget  = v14.NewTeamPremintWalletAcc
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

	err := v14.MigrateNativeMultisigs(s.ctx, s.app.BankKeeper, s.app.StakingKeeper, migrationTarget, oldMultisigs...)
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

	unbondAmount, err := v14.InstantUnbonding(s.ctx, s.app.BankKeeper, s.app.StakingKeeper, delegation, s.bondDenom)
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
