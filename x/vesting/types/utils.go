package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// ScaleCoins scales the given coins, rounding down.
func ScaleCoins(coins sdk.Coins, scale sdk.Dec) sdk.Coins {
	scaledCoins := sdk.NewCoins()
	for _, coin := range coins {
		var amt sdk.Int
		if !coin.IsZero() {
			amt = coin.Amount.ToDec().Mul(scale).TruncateInt() // round down
		}
		scaledCoins = scaledCoins.Add(sdk.NewCoin(coin.Denom, amt))
	}
	return scaledCoins
}

// CoinsMin returns the minimum of its inputs for all denominations.
func CoinsMin(a, b sdk.Coins) sdk.Coins {
	min := sdk.NewCoins()
	for _, coinA := range a {
		denom := coinA.Denom
		bAmt := b.AmountOfNoDenomValidation(denom)
		minAmt := coinA.Amount
		if minAmt.GT(bAmt) {
			minAmt = bAmt
		}
		if minAmt.IsPositive() {
			min = min.Add(sdk.NewCoin(denom, minAmt))
		}
	}
	return min
}

// coinEq returns whether two Coins are equal.
// The IsEqual() method can panic.
func coinEq(a, b sdk.Coins) bool {
	return a.IsAllLTE(b) && b.IsAllLTE(a)
}

// max64 returns the maximum of its inputs.
func Max64(i, j int64) int64 {
	if i > j {
		return i
	}
	return j
}

// Min64 returns the minimum of its inputs.
func Min64(i, j int64) int64 {
	if i < j {
		return i
	}
	return j
}
