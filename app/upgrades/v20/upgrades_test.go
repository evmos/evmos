package v20_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

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
