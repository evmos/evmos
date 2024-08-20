package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/contracts/types"
	"github.com/evmos/evmos/v19/utils"
	"github.com/stretchr/testify/require"
)

var largeAmt, _ = math.NewIntFromString("1000000000000000000000000000000000000000")

func TestNewCoinsResponse(t *testing.T) {
	testCases := []struct {
		amount math.Int
	}{
		{amount: math.NewInt(1)},
		{amount: largeAmt},
	}

	for _, tc := range testCases {
		coin := sdk.NewCoin(utils.BaseDenom, tc.amount)
		coins := sdk.NewCoins(coin)
		res := types.NewCoinsResponse(coins)
		require.Equal(t, 1, len(res))
		require.Equal(t, tc.amount.BigInt(), res[0].Amount)
	}
}

func TestNewDecCoinsResponse(t *testing.T) {
	testCases := []struct {
		amount math.Int
	}{
		{amount: math.NewInt(1)},
		{amount: largeAmt},
	}

	for _, tc := range testCases {
		coin := sdk.NewDecCoin(utils.BaseDenom, tc.amount)
		coins := sdk.NewDecCoins(coin)
		res := types.NewDecCoinsResponse(coins)
		require.Equal(t, 1, len(res))
		require.Equal(t, tc.amount.BigInt(), res[0].Amount)
	}
}
