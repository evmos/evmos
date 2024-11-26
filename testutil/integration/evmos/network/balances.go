// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package network

import (
	"cosmossdk.io/math"
)

type InitialBalanaces struct {
	baseCoinAmount math.Int
	evmCoinAmount  math.Int
}

func DefaultInitialBalances() InitialBalanaces {
	// prefundedAccountInitialBalance is the amount of tokens that each
	// prefunded account has at genesis. It represents a 100k amount expressed
	// in the 18 decimals representation.
	prefundedAccountInitialBalance, _ := math.NewIntFromString("100_000_000_000_000_000_000_000")
	return InitialBalanaces{
		baseCoinAmount: prefundedAccountInitialBalance,
		evmCoinAmount:  prefundedAccountInitialBalance,
	}
}
