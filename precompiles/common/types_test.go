package common_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v15/precompiles/common"
	"github.com/evmos/evmos/v15/utils"
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
		res := common.NewCoinsResponse(coins)
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
		res := common.NewDecCoinsResponse(coins)
		require.Equal(t, 1, len(res))
		require.Equal(t, tc.amount.BigInt(), res[0].Amount)
	}
}
