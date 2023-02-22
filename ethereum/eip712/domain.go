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
package eip712

import (
	"strconv"

	errorsmod "cosmossdk.io/errors"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// createEIP712Domain creates the Domain object for the given ChainID.
func createEIP712Domain(chainID uint64) (apitypes.TypedDataDomain, error) {
	chainIDAsInt64, err := strconv.ParseInt(strconv.FormatUint(chainID, 10), 10, 64)
	if err != nil {
		return apitypes.TypedDataDomain{}, errorsmod.Wrap(err, "invalid chainID")
	}

	domain := apitypes.TypedDataDomain{
		Name:              "Cosmos Web3",
		Version:           "1.0.0",
		ChainId:           math.NewHexOrDecimal256(chainIDAsInt64),
		VerifyingContract: "cosmos",
		Salt:              "0",
	}

	return domain, nil
}
