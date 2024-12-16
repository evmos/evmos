// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package network

import (
	"math/big"

	"cosmossdk.io/math"
	"github.com/evmos/evmos/v20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type InitialAmounts struct {
	Base math.Int
	Evm  math.Int
}

func DefaultInitialAmounts() InitialAmounts {
	baseCoinInfo := evmtypes.ChainsCoinInfo[defaultChain]

	return InitialAmounts{
		Base: GetInitialAmount(baseCoinInfo.Decimals),
		Evm:  GetInitialAmount(baseCoinInfo.Decimals),
	}
}

func DefaultInitialBondedAmount() math.Int {
	baseCoinInfo := evmtypes.ChainsCoinInfo[defaultChain]

	return GetInitialBondedAmount(baseCoinInfo.Decimals)
}

func GetInitialAmount(decimals evmtypes.Decimals) math.Int {
	if err := decimals.Validate(); err != nil {
		panic("unsupported decimals")
	}

	// initialBalance defines the initial balance represented in 18 decimals.
	initialBalance, _ := math.NewIntFromString("100_000_000_000_000_000_000_000")

	// 18 decimals is the most precise representation we can have, for this
	// reason we have to divide the initial balance by the decimals value to
	// have the specific representation.
	return initialBalance.Quo(decimals.ConversionFactor())
}

func GetInitialBondedAmount(decimals evmtypes.Decimals) math.Int {
	if err := decimals.Validate(); err != nil {
		panic("unsupported decimals")
	}

	// initialBondedAmount represents the amount of tokens that each validator will
	// have initially bonded expressed in the 18 decimals representation.
	sdk.DefaultPowerReduction = math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	initialBondedAmount := sdk.TokensFromConsensusPower(1, types.PowerReduction)

	return initialBondedAmount.Quo(decimals.ConversionFactor())
}

func GetInitialBaseFeeAmount(decimals evmtypes.Decimals) math.LegacyDec {
	if err := decimals.Validate(); err != nil {
		panic("unsupported decimals")
	}

	switch decimals {
	case evmtypes.EighteenDecimals:
		return math.LegacyNewDec(1_000_000_000)
	case evmtypes.SixDecimals:
		return math.LegacyNewDecWithPrec(1, 3)
	default:
		panic("base fee not specified")
	}
}
