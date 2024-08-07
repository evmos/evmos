package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v19/testutil"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v19/x/vesting/types"
)

func TestBalances(t *testing.T) {
	var (
		ctx    sdk.Context
		nw     *network.UnitTestNetwork
		req    *types.QueryBalancesRequest
		expRes *types.QueryBalancesResponse
	)

	testCases := []struct {
		name        string
		malleate    func()
		expPass     bool
		errContains string
	}{
		{
			name: "nil req",
			malleate: func() {
				req = nil
			},
			expPass:     false,
			errContains: "empty address string is not allowed",
		},
		{
			name: "empty req",
			malleate: func() {
				req = &types.QueryBalancesRequest{}
			},
			expPass:     false,
			errContains: "empty address string is not allowed",
		},
		{
			name: "invalid address",
			malleate: func() {
				req = &types.QueryBalancesRequest{
					Address: "evmos1",
				}
			},
			expPass:     false,
			errContains: "decoding bech32 failed: invalid bech32 string length 6",
		},
		{
			name: "invalid account - not found",
			malleate: func() {
				req = &types.QueryBalancesRequest{
					Address: vestingAddr.String(),
				}
			},
			expPass:     false,
			errContains: "either does not exist or is not a vesting account",
		},
		{
			name: "invalid account - not clawback vesting account",
			malleate: func() {
				baseAccount := authtypes.NewBaseAccountWithAddress(vestingAddr)
				acc := nw.App.AccountKeeper.NewAccount(ctx, baseAccount)
				nw.App.AccountKeeper.SetAccount(ctx, acc)

				req = &types.QueryBalancesRequest{
					Address: vestingAddr.String(),
				}
			},
			expPass:     false,
			errContains: "either does not exist or is not a vesting account",
		},
		{
			name: "valid",
			malleate: func() {
				vestingStart := ctx.BlockTime()

				// fund the vesting account with coins to initialize it and
				// then send all balances to the funding account
				err := testutil.FundAccount(ctx, nw.App.BankKeeper, vestingAddr, balances)
				require.NoError(t, err, "error while funding the target account")
				err = nw.App.BankKeeper.SendCoins(ctx, vestingAddr, funder, balances)
				require.NoError(t, err, "error while sending coins to the funder account")

				msg := types.NewMsgCreateClawbackVestingAccount(
					funder,
					vestingAddr,
					false,
				)
				_, err = nw.App.VestingKeeper.CreateClawbackVestingAccount(ctx, msg)
				require.NoError(t, err, "error while creating the vesting account")

				msgFund := types.NewMsgFundVestingAccount(
					funder,
					vestingAddr,
					vestingStart,
					lockupPeriods,
					vestingPeriods,
				)
				_, err = nw.App.VestingKeeper.FundVestingAccount(ctx, msgFund)
				require.NoError(t, err, "error while funding the vesting account")

				req = &types.QueryBalancesRequest{
					Address: vestingAddr.String(),
				}
				expRes = &types.QueryBalancesResponse{
					Locked:   balances,
					Unvested: balances,
					Vested:   nil,
				}
			},
			expPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// reset
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()
			qc := nw.GetVestingClient()

			tc.malleate()

			res, err := qc.Balances(ctx, req)
			if tc.expPass {
				require.NoError(t, err)
				require.Equal(t, expRes, res)
			} else {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}
