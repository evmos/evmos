package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	testtx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/auctions/keeper"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestEmitEndAuctionEvent(t *testing.T) {
	amt := math.NewInt(1e18)

	// TODO add more test cases
	testCases := []struct {
		name    string
		coins   sdk.Coins
		burnAmt math.Int
		expPass bool
	}{
		{
			name:    "success - one coin",
			coins:   sdk.Coins{sdk.NewCoin(utils.BaseDenom, amt)},
			burnAmt: amt,
			expPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			nw := network.NewUnitTestNetwork()
			ctx := nw.GetContext()
			winner, _ := testtx.NewAccAddressAndKey()

			err := keeper.EmitAuctionEndEvent(ctx, winner, tc.coins, tc.burnAmt)
			events := ctx.EventManager().Events()
			if !tc.expPass {
				require.Error(t, err)
				require.Len(t, events, 0)
				return
			}
			require.NoError(t, err)
			require.Len(t, events, 1)
			require.Equal(t, events[0].Type, evmtypes.EventTypeTxLog)
			// TODO check all event fields
		})
	}
}
