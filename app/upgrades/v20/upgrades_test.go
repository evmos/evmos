package v20_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	v20 "github.com/evmos/evmos/v20/app/upgrades/v20"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v20/x/evm/types"
)

func TestEnableGovPrecompile(t *testing.T) {
	var (
		nw  *network.UnitTestNetwork
		ctx sdk.Context
	)

	testCases := []struct {
		name        string
		setup       func()
		expFail     bool
		errContains string
	}{
		{
			name:        "fail - duplicated",
			setup:       func() {},
			expFail:     true,
			errContains: "duplicate precompile",
		},
		{
			name: "pass - enable gov precompile",
			setup: func() {
				params := nw.App.EvmKeeper.GetParams(ctx)
				params.ActiveStaticPrecompiles = []string{
					types.StakingPrecompileAddress,
					types.DistributionPrecompileAddress,
					types.ICS20PrecompileAddress,
					types.VestingPrecompileAddress,
				}
				require.NoError(t, nw.App.EvmKeeper.SetParams(ctx, params))
			},
			expFail: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()
			tc.setup()

			err := v20.EnableGovPrecompile(ctx, nw.App.EvmKeeper)
			if tc.expFail {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
				return
			}
			require.NoError(t, err)
			updatedParams := nw.App.EvmKeeper.GetParams(ctx)
			require.Contains(t, updatedParams.ActiveStaticPrecompiles, types.GovPrecompileAddress)
		})
	}
}

func TestUpdateExpeditedPropsParams(t *testing.T) {
	var (
		nw            *network.UnitTestNetwork
		ctx           sdk.Context
		initialParams govv1.Params
		err           error
	)

	testCases := []struct {
		name      string
		setup     func()
		postCheck func()
	}{
		{
			name:  "pass - default params, no-op",
			setup: func() {},
			postCheck: func() {
				params, err := nw.App.GovKeeper.Params.Get(ctx)
				require.NoError(t, err)
				require.Equal(t, initialParams, params)
			},
		},
		{
			name: "pass - expedited has 'stake' denom",
			setup: func() {
				// wrong exp min deposit denom
				initialParams.ExpeditedMinDeposit[0].Denom = "stake"
				err := nw.App.GovKeeper.Params.Set(ctx, initialParams)
				require.NoError(t, err)
			},
			postCheck: func() {
				params, err := nw.App.GovKeeper.Params.Get(ctx)
				require.NoError(t, err)
				require.Equal(t, "aevmos", params.ExpeditedMinDeposit[0].Denom)
				require.Equal(t, initialParams.ExpeditedMinDeposit[0].Amount, params.ExpeditedMinDeposit[0].Amount)
			},
		},
		{
			name: "pass - updates denom, amount and period",
			setup: func() {
				// wrong exp min deposit denom
				initialParams.ExpeditedMinDeposit[0].Denom = "stake"
				// wrong exp min deposit amount (< than min_deposit amt)
				initialParams.ExpeditedMinDeposit[0].Amount = initialParams.MinDeposit[0].Amount.SubRaw(1)
				// wrong exp voting period (> than voting period)
				expPeriod := *initialParams.VotingPeriod * 2
				initialParams.ExpeditedVotingPeriod = &expPeriod
				err := nw.App.GovKeeper.Params.Set(ctx, initialParams)
				require.NoError(t, err)
			},
			postCheck: func() {
				params, err := nw.App.GovKeeper.Params.Get(ctx)
				require.NoError(t, err)
				require.Equal(t, "aevmos", params.ExpeditedMinDeposit[0].Denom)
				require.Equal(t, initialParams.MinDeposit[0].Amount.MulRaw(govv1.DefaultMinExpeditedDepositTokensRatio), params.ExpeditedMinDeposit[0].Amount)
				require.Equal(t, *initialParams.VotingPeriod/2, *params.ExpeditedVotingPeriod)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()
			initialParams, err = nw.App.GovKeeper.Params.Get(ctx)
			require.NoError(t, err)
			// setup for testcase
			tc.setup()

			err = v20.UpdateExpeditedPropsParams(ctx, nw.App.GovKeeper)
			require.NoError(t, err)

			tc.postCheck()
		})
	}
}
