package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/utils"
)

// removeBaseCoinFromCoins returns an sdk.Coins removing the
// base denom from the input.
func removeBaseCoinFromCoins(coins sdk.Coins) sdk.Coins {
	remainingCoins := sdk.NewCoins()
	// TODO: check if evmos tokens are accumulated when the module has more than the bid.
	for _, coin := range coins {
		if coin.Denom != utils.BaseDenom {
			remainingCoins = remainingCoins.Add(coin)
		}
	}
	return remainingCoins
}
