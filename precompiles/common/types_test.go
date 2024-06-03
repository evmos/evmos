package common_test

import (
	"testing"

	"github.com/evmos/evmos/v18/cmd/config"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/cmd/config"
	bankprecompile "github.com/evmos/evmos/v18/precompiles/bank"
	"github.com/evmos/evmos/v18/precompiles/bech32"
	"github.com/evmos/evmos/v18/precompiles/common"
	distprecompile "github.com/evmos/evmos/v18/precompiles/distribution"
	ics20precompile "github.com/evmos/evmos/v18/precompiles/ics20"
	"github.com/evmos/evmos/v18/precompiles/p256"
	stakingprecompile "github.com/evmos/evmos/v18/precompiles/staking"
	vestingprecompile "github.com/evmos/evmos/v18/precompiles/vesting"
	"github.com/evmos/evmos/v18/utils"
	"github.com/stretchr/testify/require"
)

var largeAmt, _ = math.NewIntFromString("1000000000000000000000000000000000000000")

func TestDefaultPrecompiles(t *testing.T) {
	// We need to update the bech32 prefixes to use evmos1... instead of the default cosmos1...
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)

	require.Equal(t, 7, len(common.DefaultPrecompilesBech32),
		"expected different number of default precompiles",
	)
	require.Equal(t, []string{
		sdk.AccAddress(ethcommon.HexToAddress(p256.PrecompileAddress).Bytes()).String(),
		sdk.AccAddress(ethcommon.HexToAddress(bech32.PrecompileAddress).Bytes()).String(),
		sdk.AccAddress(ethcommon.HexToAddress(stakingprecompile.PrecompileAddress).Bytes()).String(),
		sdk.AccAddress(ethcommon.HexToAddress(distprecompile.PrecompileAddress).Bytes()).String(),
		sdk.AccAddress(ethcommon.HexToAddress(ics20precompile.PrecompileAddress).Bytes()).String(),
		sdk.AccAddress(ethcommon.HexToAddress(vestingprecompile.PrecompileAddress).Bytes()).String(),
		sdk.AccAddress(ethcommon.HexToAddress(bankprecompile.PrecompileAddress).Bytes()).String(),
	}, common.DefaultPrecompilesBech32)
}

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
