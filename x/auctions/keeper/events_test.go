package keeper_test

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v19/contracts/types"
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
		round   uint64
		burnAmt math.Int
		expPass bool
	}{
		{
			name:    "success - one coin",
			coins:   sdk.Coins{sdk.NewCoin(utils.BaseDenom, amt)},
			round:   1,
			burnAmt: amt,
			expPass: true,
		},
		{
			name:    "success - many coins",
			coins:   sdk.Coins{sdk.NewCoin(utils.BaseDenom, amt), sdk.NewCoin("atest", amt.SubRaw(5e5)), sdk.NewCoin("axmpl", amt.SubRaw(2e12))},
			round:   1,
			burnAmt: amt,
			expPass: true,
		},
		{
			name:    "success - no coins",
			coins:   nil,
			round:   1,
			burnAmt: amt,
			expPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			nw := network.NewUnitTestNetwork()
			ctx := nw.GetContext()
			winner, _ := testtx.NewAccAddressAndKey()
			bidWinnerHexAddr := common.BytesToAddress(winner.Bytes())

			err := keeper.EmitAuctionEndEvent(ctx, winner, tc.round, tc.coins, tc.burnAmt)
			events := ctx.EventManager().Events()
			if !tc.expPass {
				require.Error(t, err)
				require.Len(t, events, 0)
				return
			}
			require.NoError(t, err)
			require.Len(t, events, 1)
			require.Equal(t, events[0].Type, evmtypes.EventTypeTxLog)
			require.Len(t, events[0].Attributes, 1)

			var log evmtypes.Log
			err = json.Unmarshal([]byte(events[0].Attributes[0].Value), &log)
			require.NoError(t, err)

			ethLog := log.ToEthereum()
			require.Equal(t, common.HexToAddress(keeper.PrecompileAddress), ethLog.Address)

			require.Equal(t, keeper.EndAuctionEventABI.ID, ethLog.Topics[0])
			require.Equal(t, common.LeftPadBytes(bidWinnerHexAddr.Bytes(), 32), ethLog.Topics[1].Bytes())
			require.Equal(t, big.NewInt(int64(tc.round)), ethLog.Topics[2].Big()) //#nosec G115 -- gosec warning about integer overflow is not relevant here

			require.Len(t, ethLog.Topics, 3)

			require.Equal(t, uint64(ctx.BlockHeight()), ethLog.BlockNumber)
			require.Equal(t, common.BytesToHash(ctx.HeaderHash()), ethLog.BlockHash)

			logData, err := keeper.EndAuctionEventABI.Inputs.Unpack(ethLog.Data)
			require.NoError(t, err)
			// first arg of log data should be the coins
			coins, ok := logData[0].([]types.Coin)
			require.True(t, ok)
			for i, c := range coins {
				require.Equal(t, *tc.coins[i].Amount.BigInt(), *c.Amount)
				require.Equal(t, tc.coins[i].Denom, c.Denom)
			}

			// second arg is the burned amount
			amt, ok := logData[1].(*big.Int)
			require.True(t, ok)
			require.Equal(t, *tc.burnAmt.BigInt(), *amt)
		})
	}
}
