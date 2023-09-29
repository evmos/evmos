package v14_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	v14 "github.com/evmos/evmos/v14/app/upgrades/v14"
	"github.com/evmos/evmos/v14/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v14/testutil"
	testutiltx "github.com/evmos/evmos/v14/testutil/tx"
)

// zeroDec is a zero decimal value
var zeroDec = sdk.ZeroDec()

// MigrationTestAccount is a struct to hold the test account address, its private key
// as well as its balances and delegations before and after the migration.
type MigrationTestAccount struct {
	Addr            sdk.AccAddress
	PrivKey         *ethsecp256k1.PrivKey
	BalancePre      sdk.Coins
	BalancePost     sdk.Coins
	DelegationsPre  stakingtypes.Delegations
	DelegationsPost stakingtypes.Delegations
}

func GenerateMigrationTestAccount() MigrationTestAccount {
	addr, priv := testutiltx.NewAccAddressAndKey()
	return MigrationTestAccount{
		Addr:            addr,
		PrivKey:         priv,
		BalancePre:      sdk.Coins{},
		BalancePost:     sdk.Coins{},
		DelegationsPre:  stakingtypes.Delegations{},
		DelegationsPost: stakingtypes.Delegations{},
	}
}

// requireMigratedAccount checks that the account has no delegations, no unbonding delegations
func (s *UpgradesTestSuite) requireMigratedAccount(account MigrationTestAccount) {
	delegations := s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, account.Addr)
	s.Require().ElementsMatch(delegations, account.DelegationsPost, "expected different delegations after migration of account %s", account.Addr.String())
	unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, account.Addr)
	s.Require().Len(unbondingDelegations, 0, "expected no unbonding delegations after migration for account %s", account.Addr.String())
	balances := s.app.BankKeeper.GetAllBalances(s.ctx, account.Addr)
	s.Require().ElementsMatch(balances, account.BalancePost, "expected different balance after migration for account %s", account.Addr.String())
}

func (s *UpgradesTestSuite) TestUpdateMigrateNativeMultisigs() {
	s.SetupTest()

	// Check validator shares before migration
	expectedSharesMap := s.getDelegationSharesMap()

	amountPremint, ok := sdk.NewIntFromString("64699999994000000000000000")
	s.Require().True(ok, "failed to parse premint amount")
	amount1, ok := sdk.NewIntFromString("13824747333293928482487986")
	s.Require().True(ok, "failed to parse amount1")
	amount1IBC, ok := sdk.NewIntFromString("421720500000000000000")
	s.Require().True(ok, "failed to parse amount2")
	amount2, ok := sdk.NewIntFromString("494000000000000000")
	s.Require().True(ok, "failed to parse amount3")
	amount3 := amount2
	amount4 := amount2
	amount5, ok := sdk.NewIntFromString("1013699976000000000000000")
	s.Require().True(ok, "failed to parse amount6")

	var (
		oldPremintCoin     = sdk.Coin{Denom: s.bondDenom, Amount: amountPremint}
		stratRes1EvmosCoin = sdk.Coin{Denom: s.bondDenom, Amount: amount1}
		stratRes1IBCCoin   = sdk.Coin{Denom: "someIBCdenom", Amount: amount1IBC}
		stratRes2Coin      = sdk.Coin{Denom: s.bondDenom, Amount: amount2}
		stratRes3Coin      = sdk.Coin{Denom: s.bondDenom, Amount: amount3}
		stratRes4Coin      = sdk.Coin{Denom: s.bondDenom, Amount: amount4}
		stratRes5Coin      = sdk.Coin{Denom: s.bondDenom, Amount: amount5}
	)

	oldStrategicReserves := make([]MigrationTestAccount, 0, 5)
	for idx := 0; idx < 5; idx++ {
		oldStrategicReserves = append(oldStrategicReserves, GenerateMigrationTestAccount())
	}
	// assign pre-balances
	oldStrategicReserves[0].BalancePre = sdk.Coins{stratRes1EvmosCoin, stratRes1IBCCoin}
	oldStrategicReserves[1].BalancePre = sdk.Coins{stratRes2Coin}
	oldStrategicReserves[2].BalancePre = sdk.Coins{stratRes3Coin}
	oldStrategicReserves[3].BalancePre = sdk.Coins{stratRes4Coin}
	oldStrategicReserves[4].BalancePre = sdk.Coins{stratRes5Coin}

	oldStrategicReserves[0].DelegationsPre = stakingtypes.Delegations{
		stakingtypes.Delegation{
			DelegatorAddress: oldStrategicReserves[0].Addr.String(),
			ValidatorAddress: s.validators[0].OperatorAddress,
			Shares:           sdk.NewDecWithPrec(2452009409460295636, 18),
		},
		stakingtypes.Delegation{
			DelegatorAddress: oldStrategicReserves[0].Addr.String(),
			ValidatorAddress: s.validators[1].OperatorAddress,
			Shares:           sdk.NewDecWithPrec(173554344899830220, 18),
		},
	}

	// the new strategic reserve should hold the sum of all old strategic reserves
	newStrategicReserve := GenerateMigrationTestAccount()
	newStrategicReserve.BalancePost = sdk.Coins{
		stratRes1IBCCoin,
		stratRes1EvmosCoin.Add(stratRes2Coin).Add(stratRes3Coin).Add(stratRes4Coin).Add(stratRes5Coin),
	}
	newStrategicReserve.DelegationsPost = stakingtypes.Delegations{
		stakingtypes.Delegation{
			DelegatorAddress: newStrategicReserve.Addr.String(),
			ValidatorAddress: s.validators[0].OperatorAddress,
			Shares:           sdk.NewDecWithPrec(2452009409460295636, 18),
		},
		stakingtypes.Delegation{
			DelegatorAddress: newStrategicReserve.Addr.String(),
			ValidatorAddress: s.validators[1].OperatorAddress,
			Shares:           sdk.NewDecWithPrec(173554344899830220, 18),
		},
	}

	// premint wallets
	oldPremintWallet := GenerateMigrationTestAccount()
	oldPremintWallet.BalancePre = sdk.Coins{oldPremintCoin}

	// the new premint wallet should have the same balance as the old premint wallet before the migration
	newPremintWallet := GenerateMigrationTestAccount()
	newPremintWallet.BalancePost = sdk.Coins{oldPremintCoin}

	// Prepare the accounts to be migrated
	affectedAccounts := append(oldStrategicReserves, oldPremintWallet)
	for _, affectedAccount := range affectedAccounts {
		err := testutil.FundAccount(s.ctx, s.app.BankKeeper, affectedAccount.Addr, affectedAccount.BalancePre)
		s.Require().NoError(err, "failed to fund account %s", affectedAccount.Addr.String())

		if len(affectedAccount.DelegationsPre) == 0 {
			continue
		}

		for _, delegation := range affectedAccount.DelegationsPre {
			s.app.StakingKeeper.SetDelegation(s.ctx, delegation)
		}
	}

	oldStrategicReservesAddrs := make([]string, 0, len(oldStrategicReserves))
	for _, oldStrategicReserve := range oldStrategicReserves {
		oldStrategicReservesAddrs = append(oldStrategicReservesAddrs, oldStrategicReserve.Addr.String())
	}

	err := v14.MigrateNativeMultisigs(s.ctx, s.app.BankKeeper, s.app.EvmKeeper, s.app.StakingKeeper, newStrategicReserve.Addr, oldStrategicReservesAddrs...)
	s.Require().NoError(err, "failed to migrate native multisigs")

	// Check that the multisigs have been updated
	expectedAccounts := append(oldStrategicReserves, newStrategicReserve, oldPremintWallet, newPremintWallet)
	for _, account := range expectedAccounts {
		s.requireMigratedAccount(account)
	}

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
