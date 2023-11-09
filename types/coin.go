// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

import (
	"math/big"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// AttoEvmos defines the default coin denomination used in Evmos in:
	//
	// - Staking parameters: denomination used as stake in the dPoS chain
	// - Mint parameters: denomination minted due to fee distribution rewards
	// - Governance parameters: denomination used for spam prevention in proposal deposits
	// - Crisis parameters: constant fee denomination used for spam prevention to check broken invariant
	// - EVM parameters: denomination used for running EVM state transitions in Evmos.
	AttoEvmos string = "aevmos"

	// BaseDenomUnit defines the base denomination unit for Evmos.
	// 1 evmos = 1x10^{BaseDenomUnit} aevmos
	BaseDenomUnit = 18

	// DefaultGasPrice is default gas price for evm transactions
	DefaultGasPrice = 20
)

// PowerReduction defines the default power reduction value for staking
var PowerReduction = sdkmath.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(BaseDenomUnit), nil))

// NewEvmosCoin is a utility function that returns an "aevmos" coin with the given sdkmath.Int amount.
// The function will panic if the provided amount is negative.
func NewEvmosCoin(amount sdkmath.Int) sdk.Coin {
	return sdk.NewCoin(AttoEvmos, amount)
}

// NewEvmosDecCoin is a utility function that returns an "aevmos" decimal coin with the given sdkmath.Int amount.
// The function will panic if the provided amount is negative.
func NewEvmosDecCoin(amount sdkmath.Int) sdk.DecCoin {
	return sdk.NewDecCoin(AttoEvmos, amount)
}

// NewEvmosCoinInt64 is a utility function that returns an "aevmos" coin with the given int64 amount.
// The function will panic if the provided amount is negative.
func NewEvmosCoinInt64(amount int64) sdk.Coin {
	return sdk.NewInt64Coin(AttoEvmos, amount)
}
