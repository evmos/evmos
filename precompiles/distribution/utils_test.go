package distribution_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/evmos/v20/precompiles/staking"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	inflationtypes "github.com/evmos/evmos/v20/x/inflation/v1/types"
)

type stakingRewards struct {
	Delegator sdk.AccAddress
	Validator stakingtypes.Validator
	RewardAmt math.Int
}

var (
	testRewardsAmt, _       = math.NewIntFromString("1000000000000000000")
	validatorCommPercentage = math.LegacyNewDecWithPrec(5, 2) // 5% commission
	validatorCommAmt        = math.LegacyNewDecFromInt(testRewardsAmt).Mul(validatorCommPercentage).TruncateInt()
	expRewardsAmt           = testRewardsAmt.Sub(validatorCommAmt) // testRewardsAmt - commission
)

// prepareStakingRewards prepares the test suite for testing delegation rewards.
//
// Specified rewards amount are allocated to the specified validator using the distribution keeper,
// such that the given amount of tokens is outstanding as a staking reward for the account.
//
// The setup is done in the following way:
//   - Fund distribution module to pay for rewards.
//   - Allocate rewards to the validator.
func (s *PrecompileTestSuite) prepareStakingRewards(ctx sdk.Context, stkRs ...stakingRewards) (sdk.Context, error) {
	for _, r := range stkRs {
		// set distribution module account balance which pays out the rewards
		coins := sdk.NewCoins(sdk.NewCoin(s.bondDenom, r.RewardAmt))
		if err := s.mintCoinsForDistrMod(ctx, coins); err != nil {
			return ctx, err
		}

		// allocate rewards to validator
		allocatedRewards := sdk.NewDecCoins(sdk.NewDecCoin(s.bondDenom, r.RewardAmt))
		if err := s.network.App.DistrKeeper.AllocateTokensToValidator(ctx, r.Validator, allocatedRewards); err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

// mintCoinsForDistrMod is a helper function to mint a specific amount of coins from the
// distribution module to pay for staking rewards.
func (s *PrecompileTestSuite) mintCoinsForDistrMod(ctx sdk.Context, amount sdk.Coins) error {
	// Minting tokens for the FeeCollector to simulate fee accrued.
	if err := s.network.App.BankKeeper.MintCoins(
		ctx,
		inflationtypes.ModuleName,
		amount,
	); err != nil {
		return err
	}

	return s.network.App.BankKeeper.SendCoinsFromModuleToModule(
		ctx,
		inflationtypes.ModuleName,
		distrtypes.ModuleName,
		amount,
	)
}

// fundAccountWithBaseDenom is a helper function to fund a given address with the chain's
// base denomination.
func (s *PrecompileTestSuite) fundAccountWithBaseDenom(ctx sdk.Context, addr sdk.AccAddress, amount math.Int) error {
	coins := sdk.NewCoins(sdk.NewCoin(s.bondDenom, amount))
	if err := s.network.App.BankKeeper.MintCoins(ctx, inflationtypes.ModuleName, coins); err != nil {
		return err
	}
	return s.network.App.BankKeeper.SendCoinsFromModuleToAccount(ctx, inflationtypes.ModuleName, addr, coins)
}

func (s *PrecompileTestSuite) getStakingPrecompile() (*staking.Precompile, error) {
	return staking.NewPrecompile(
		s.network.App.StakingKeeper,
		s.network.App.AuthzKeeper,
	)
}

func generateKeys(count int) []keyring.Key {
	accs := make([]keyring.Key, 0, count)
	for i := 0; i < count; i++ {
		acc := keyring.NewKey()
		accs = append(accs, acc)
	}
	return accs
}
