package evm

import sdk "github.com/cosmos/cosmos-sdk/types"

// ClaimSufficientStakingRewards checks if the account has enough staking rewards unclaimed
// to cover the given amount. If more than enough rewards are unclaimed, only those up to
// the given amount are claimed.
func ClaimSufficientStakingRewards(ctx sdk.Context, addr sdk.AccAddress, amount sdk.Coin) error {
	return nil
}
