package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v20/x/feemarket/types"
	"github.com/stretchr/testify/require"
)

func TestUpdateParams(t *testing.T) {
	var (
		nw  *network.UnitTestNetwork
		ctx sdk.Context
	)

	testCases := []struct {
		name      string
		request   *types.MsgUpdateParams
		expectErr bool
	}{
		{
			name:      "fail - invalid authority",
			request:   &types.MsgUpdateParams{Authority: "foobar"},
			expectErr: true,
		},
		{
			name: "pass - valid Update msg",
			request: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.DefaultParams(),
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// reset network and context
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			_, err := nw.App.FeeMarketKeeper.UpdateParams(ctx, tc.request)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
