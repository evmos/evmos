// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package types

import (
	fmt "fmt"
	math "math"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
)

const maxBitLen = 256

// SafeInt64 checks for overflows while casting a uint64 to int64 value.
func SafeInt64(value uint64) (int64, error) {
	if value > uint64(math.MaxInt64) {
		return 0, errorsmod.Wrapf(errortypes.ErrInvalidHeight, "uint64 value %v cannot exceed %v", value, int64(math.MaxInt64))
	}

	return int64(value), nil // #nosec G701 -- checked for int overflow already
}

// SafeNewIntFromBigInt constructs Int from big.Int, return error if more than 256bits
func SafeNewIntFromBigInt(i *big.Int) (sdkmath.Int, error) {
	if !IsValidInt256(i) {
		return sdkmath.NewInt(0), fmt.Errorf("big int out of bound: %s", i)
	}
	return sdkmath.NewIntFromBigInt(i), nil
}

// IsValidInt256 check the bound of 256 bit number
func IsValidInt256(i *big.Int) bool {
	return i == nil || i.BitLen() <= maxBitLen
}
