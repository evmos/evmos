package distribution_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/ethereum/go-ethereum/common"

	evmosutil "github.com/evmos/evmos/v16/testutil"
	"github.com/evmos/evmos/v16/x/evm/statedb"
	inflationtypes "github.com/evmos/evmos/v16/x/inflation/v1/types"
)

type stakingRewards struct {
	Delegator sdk.AccAddress
	Validator stakingtypes.Validator
	RewardAmt math.Int
}

// prepareStakingRewards prepares the test suite for testing delegation rewards.
//
// Specified rewards amount are allocated to the specified validator using the distribution keeper,
// such that the given amount of tokens is outstanding as a staking reward for the account.
//
// The setup is done in the following way:
//   - Fund the account with the given address with the given rewards amount.
//   - Delegate the rewards amount to the validator specified
//   - Allocate rewards to the validator.
func (s *PrecompileTestSuite) prepareStakingRewards(ctx sdk.Context, stkRs ...stakingRewards) (sdk.Context, error) {
	for _, r := range stkRs {
		// fund account to make delegation
		if err := evmosutil.FundAccountWithBaseDenom(ctx, s.network.App.BankKeeper, r.Delegator, r.RewardAmt.Int64()); err != nil {
			return ctx, err
		}
		// set distribution module account balance which pays out the rewards
		distrAcc := s.network.App.DistrKeeper.GetDistributionAccount(ctx)
		if err := evmosutil.FundModuleAccount(ctx, s.network.App.BankKeeper, distrAcc.GetName(), sdk.NewCoins(sdk.NewCoin(s.bondDenom, r.RewardAmt))); err != nil {
			return ctx, err
		}

		// make a delegation
		if _, err := s.network.App.StakingKeeper.Delegate(ctx, r.Delegator, r.RewardAmt, stakingtypes.Unspecified, r.Validator, true); err != nil {
			return ctx, err
		}

		// end block to bond validator and increase block height
		if _, err := s.network.App.StakingKeeper.EndBlocker(ctx); err != nil {
			return ctx, err
		}
		// allocate rewards to validator (of these 50% will be paid out to the delegator)
		allocatedRewards := sdk.NewDecCoins(sdk.NewDecCoin(s.bondDenom, r.RewardAmt.Mul(math.NewInt(2))))
		if err := s.network.App.DistrKeeper.AllocateTokensToValidator(ctx, r.Validator, allocatedRewards); err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

// getStateDB is a helper function used in unit tests only
// to get a stateDB instance from the provided context
func (s *PrecompileTestSuite) getStateDB(ctx sdk.Context) *statedb.StateDB {
	headerHash := ctx.HeaderHash()
	return statedb.New(
		ctx,
		s.network.App.EvmKeeper,
		statedb.NewEmptyTxConfig(common.BytesToHash(headerHash)),
	)
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
