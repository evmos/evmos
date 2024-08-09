package types

import (
	"math/big"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

// Coin defines a struct that stores all needed information about a coin
// in types native to the EVM.
type Coin struct {
	Denom  string
	Amount *big.Int
}

// DecCoin defines a struct that stores all needed information about a decimal coin
// in types native to the EVM.
type DecCoin struct {
	Denom     string
	Amount    *big.Int
	Precision uint8
}

// Dec defines a struct that represents a decimal number of a given precision
// in types native to the EVM.
type Dec struct {
	Value     *big.Int
	Precision uint8
}

// ToSDKType converts the Coin to the Cosmos SDK representation.
func (c Coin) ToSDKType() sdk.Coin {
	return sdk.NewCoin(c.Denom, math.NewIntFromBigInt(c.Amount))
}

// NewCoinsResponse converts a response to an array of Coin.
func NewCoinsResponse(amount sdk.Coins) []Coin {
	// Create a new output for each coin and add it to the output array.
	outputs := make([]Coin, len(amount))
	for i, coin := range amount {
		outputs[i] = Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount.BigInt(),
		}
	}
	return outputs
}

// NewDecCoinsResponse converts a response to an array of DecCoin.
func NewDecCoinsResponse(amount sdk.DecCoins) []DecCoin {
	// Create a new output for each coin and add it to the output array.
	outputs := make([]DecCoin, len(amount))
	for i, coin := range amount {
		outputs[i] = DecCoin{
			Denom:     coin.Denom,
			Amount:    coin.Amount.TruncateInt().BigInt(),
			Precision: math.LegacyPrecision,
		}
	}
	return outputs
}

// HexAddressFromBech32String converts a hex address to a bech32 encoded address.
func HexAddressFromBech32String(addr string) (res common.Address, err error) {
	if strings.Contains(addr, sdk.PrefixValidator) {
		valAddr, err := sdk.ValAddressFromBech32(addr)
		if err != nil {
			return res, err
		}
		return common.BytesToAddress(valAddr.Bytes()), nil
	}
	return common.BytesToAddress(sdk.MustAccAddressFromBech32(addr)), nil
}

// SafeAdd adds two integers and returns a boolean if an overflow occurs to avoid panic.
// TODO: Upstream this to the SDK math package.
func SafeAdd(a, b math.Int) (res *big.Int, overflow bool) {
	res = a.BigInt().Add(a.BigInt(), b.BigInt())
	return res, res.BitLen() > math.MaxBitLen
}
