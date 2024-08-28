// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/utils"
)

// removeBaseCoinFromCoins returns an sdk.Coins removing the
// base denom from the input.
func removeBaseCoinFromCoins(coins sdk.Coins) sdk.Coins {
	remainingCoins := sdk.NewCoins()
	for _, coin := range coins {
		if coin.Denom != utils.BaseDenom {
			remainingCoins = remainingCoins.Add(coin)
		}
	}
	return remainingCoins
}
