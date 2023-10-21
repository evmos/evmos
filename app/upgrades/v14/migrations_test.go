package v14_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	v14 "github.com/evmos/evmos/v15/app/upgrades/v14"
	"github.com/evmos/evmos/v15/testutil"
	testutiltx "github.com/evmos/evmos/v15/testutil/tx"
)

// TestUpdateMigrateNativeMultisigs is the main test for the migration of the strategic reserves and the premint wallet.
// This test is where the actual mainnet data is being replicated and the full migration tested.
func (s *UpgradesTestSuite) TestUpdateMigrateNativeMultisigs() {
	s.SetupTest()

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

		// We are delegating one token to each of the validators in the test setup
		delegateAmount = int64(1)
		// One delegated token equals to 1e-18 share issued for the delegator
		delegateShares = math.LegacyNewDecWithPrec(1, 18)
	)

	oldStrategicReserves := make([]MigrationTestAccount, 0, 5)
	for idx := 0; idx < 5; idx++ {
		oldStrategicReserves = append(oldStrategicReserves, GenerateMigrationTestAccount())
		s.T().Logf("Old Strategic Reserve %d: %q\n", idx+1, oldStrategicReserves[idx].Addr.String())
	}
	// assign pre-balances
	oldStrategicReserves[0].BalancePre = sdk.NewCoins(stratRes1EvmosCoin, stratRes1IBCCoin)
	oldStrategicReserves[1].BalancePre = sdk.NewCoins(stratRes2Coin)
	oldStrategicReserves[2].BalancePre = sdk.NewCoins(stratRes3Coin)
	oldStrategicReserves[3].BalancePre = sdk.NewCoins(stratRes4Coin)
	oldStrategicReserves[4].BalancePre = sdk.NewCoins(stratRes5Coin)

	// the new strategic reserve should hold the sum of all old strategic reserves
	newStrategicReserve := GenerateMigrationTestAccount()
	s.T().Logf("New Strategic Reserve: %q\n", newStrategicReserve.Addr.String())
	newStrategicReserve.BalancePost = sdk.NewCoins(
		stratRes1IBCCoin,
		stratRes1EvmosCoin.Add(stratRes2Coin).Add(stratRes3Coin).Add(stratRes4Coin).Add(stratRes5Coin),
	)
	// NOTE: after the migration the delegation that returns zero tokens should be removed / not newly delegated to
	newStrategicReserve.DelegationsPost = stakingtypes.Delegations{
		stakingtypes.Delegation{
			DelegatorAddress: newStrategicReserve.Addr.String(),
			ValidatorAddress: s.validators[1].OperatorAddress,
			Shares:           delegateShares,
		},
	}

	// premint wallets
	oldPremintWallet := GenerateMigrationTestAccount()
	s.T().Logf("Old Premint Wallet: %q\n", oldPremintWallet.Addr.String())
	oldPremintWallet.BalancePre = sdk.Coins{oldPremintCoin}

	// the new premint wallet should have the same balance as the old premint wallet before the migration
	newPremintWallet := GenerateMigrationTestAccount()
	s.T().Logf("New Premint Wallet: %q\n", newPremintWallet.Addr.String())
	newPremintWallet.BalancePost = sdk.Coins{oldPremintCoin}

	// Fund the accounts to be migrated
	affectedAccounts := oldStrategicReserves
	affectedAccounts = append(affectedAccounts, oldPremintWallet)
	for _, affectedAccount := range affectedAccounts {
		err := testutil.FundAccount(s.ctx, s.app.BankKeeper, affectedAccount.Addr, affectedAccount.BalancePre)
		s.Require().NoError(err, "failed to fund account %s", affectedAccount.Addr.String())
	}

	// delegation to validator 0 with zero tokens being returned because of the slashing
	_, err := CreateDelegationWithZeroTokens(
		s.ctx,
		s.app,
		oldStrategicReserves[0].PrivKey,
		oldStrategicReserves[0].Addr,
		s.validators[0],
		delegateAmount,
	)
	s.Require().NoError(err, "failed to create delegation with zero tokens")

	// delegation to validator 1
	_, err = Delegate(
		s.ctx,
		s.app,
		oldStrategicReserves[0].PrivKey,
		oldStrategicReserves[0].Addr,
		s.validators[1],
		delegateAmount,
	)
	s.Require().NoError(err, "failed to create delegation")

	// NOTE: We send twice the delegate amount to the old strategic reserve here, because that is
	// the amount that was just spent on delegations and thus removed from the balance. We need to fill it up again,
	// because the expected post-migration balance does not include the delegated amount.
	err = testutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, oldStrategicReserves[0].Addr, 2*delegateAmount)
	s.Require().NoError(err, "failed to fund account %s to even out delegated amount", oldStrategicReserves[0].Addr.String())
	// Additionally, we need to send twice the default fee that is charged for the delegation transaction.
	feeAmt := testutiltx.DefaultFee.Amount.MulRaw(2)
	err = testutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, oldStrategicReserves[0].Addr, feeAmt.Int64())
	s.Require().NoError(err, "failed to fund account %s to account for delegation fees", oldStrategicReserves[0].Addr.String())

	// Store addresses in a slice
	oldStrategicReservesAddrs := make([]string, 0, len(oldStrategicReserves))
	for _, oldStrategicReserve := range oldStrategicReserves {
		oldStrategicReservesAddrs = append(oldStrategicReservesAddrs, oldStrategicReserve.Addr.String())
	}

	// Check validator shares before migration, which are stored as the expected shares map.
	//
	// NOTE: There is a minor difference expected between the pre- and post-migration shares. This is because
	// the zero-token delegation is being unbonded and the corresponding shares are removed. Since zero tokens
	// would be delegated to the validator after the migration, the shares are not added again, creating a reduction
	// in the total shares of s.validators[0] of 1e-18.
	expectedSharesMap := s.getDelegationSharesMap()
	expectedSharesMap[s.validators[0].OperatorAddress] = expectedSharesMap[s.validators[0].OperatorAddress].Sub(delegateShares)

	// Migrate strategic reserves
	err = v14.MigrateNativeMultisigs(s.ctx, s.app.BankKeeper, s.app.StakingKeeper, newStrategicReserve.Addr, oldStrategicReservesAddrs...)
	s.Require().NoError(err, "failed to migrate strategic reserves")

	// Migrate premint wallet
	err = v14.MigrateNativeMultisigs(s.ctx, s.app.BankKeeper, s.app.StakingKeeper, newPremintWallet.Addr, oldPremintWallet.Addr.String())
	s.Require().NoError(err, "failed to migrate premint wallet")

	// Check that the multisigs have been updated
	expectedAccounts := oldStrategicReserves
	expectedAccounts = append(expectedAccounts, newStrategicReserve, oldPremintWallet, newPremintWallet)
	for _, account := range expectedAccounts {
		s.requireMigratedAccount(account)
	}

	// Check validator shares after migration.
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

func (s *UpgradesTestSuite) TestCreateDelegationWithZeroTokens() {
	s.SetupTest()

	targetValidator := s.validators[1]

	// Create new account and fund it
	addr, priv := testutiltx.NewAccAddressAndKey()
	err := testutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, addr, 2e18)
	s.Require().NoError(err, "failed to fund account")

	s.NextBlock()

	delegation, err := CreateDelegationWithZeroTokens(s.ctx, s.app, priv, addr, targetValidator, 1)
	s.Require().NoError(err, "failed to create delegation with zero tokens")
	s.Require().NotEqual(sdk.ZeroDec(), delegation.Shares, "delegation shares should not be zero")

	// Check that the validators tokenFromShares method returns zero tokens when truncated to an int
	valAfterSlashing := s.app.StakingKeeper.Validator(s.ctx, targetValidator.GetOperator())
	tokens := valAfterSlashing.TokensFromShares(delegation.Shares).TruncateInt()
	s.Require().Equal(int64(0), tokens.Int64(), "expected zero tokens to be returned")
}
