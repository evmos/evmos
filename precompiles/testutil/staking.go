// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testutil

import (
	"testing"

	"cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	teststaking "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

// CreateValidator creates a validator with the given amount of staked tokens in the bond denomination set
// in the staking keeper.
func CreateValidator(ctx sdk.Context, t *testing.T, pubKey cryptotypes.PubKey, sk stakingkeeper.Keeper, stakeAmt math.Int) {
	zeroDec := math.LegacyZeroDec()
	stakingParams := sk.GetParams(ctx)
	stakingParams.BondDenom = sk.BondDenom(ctx)
	stakingParams.MinCommissionRate = zeroDec
	err := sk.SetParams(ctx, stakingParams)
	require.NoError(t, err)

	stakingHelper := teststaking.NewHelper(t, ctx, &sk)
	stakingHelper.Commission = stakingtypes.NewCommissionRates(zeroDec, zeroDec, zeroDec)
	stakingHelper.Denom = sk.BondDenom(ctx)

	valAddr := sdk.ValAddress(pubKey.Address())
	stakingHelper.CreateValidator(valAddr, pubKey, stakeAmt, true)
}
