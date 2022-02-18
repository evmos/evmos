package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewTestCoins coins to more than cover the fee
func NewTestCoins() sdk.Coins {
	return sdk.Coins{
		sdk.NewInt64Coin("atom", 10000000),
	}
}
