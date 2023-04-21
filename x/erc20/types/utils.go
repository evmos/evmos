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
	"fmt"
	"regexp"
	"strings"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	// (?m)^(\d+) remove leading numbers
	reLeadingNumbers = `(?m)^(\d+)`
	// ^[^A-Za-z] forces first chars to be letters
	// [^a-zA-Z0-9/-] deletes special characters
	reDnmString = `^[^A-Za-z]|[^a-zA-Z0-9/-]`
)

func removeLeadingNumbers(str string) string {
	re := regexp.MustCompile(reLeadingNumbers)
	return re.ReplaceAllString(str, "")
}

func removeSpecialChars(str string) string {
	re := regexp.MustCompile(reDnmString)
	return re.ReplaceAllString(str, "")
}

// recursively remove every invalid prefix
func removeInvalidPrefixes(str string) string {
	if strings.HasPrefix(str, "ibc/") {
		return removeInvalidPrefixes(str[4:])
	}
	if strings.HasPrefix(str, "erc20/") {
		return removeInvalidPrefixes(str[6:])
	}
	return str
}

// SanitizeERC20Name enforces 128 max string length, deletes leading numbers
// removes special characters  (except /)  and spaces from the ERC20 name
func SanitizeERC20Name(name string) string {
	name = removeLeadingNumbers(name)
	name = removeSpecialChars(name)
	if len(name) > 128 {
		name = name[:128]
	}
	name = removeInvalidPrefixes(name)
	return name
}

// EqualMetadata checks if all the fields of the provided coin metadata are equal.
func EqualMetadata(a, b banktypes.Metadata) error {
	if a.Base == b.Base && a.Description == b.Description && a.Display == b.Display && a.Name == b.Name && a.Symbol == b.Symbol {
		if len(a.DenomUnits) != len(b.DenomUnits) {
			return fmt.Errorf("metadata provided has different denom units from stored, %d ≠ %d", len(a.DenomUnits), len(b.DenomUnits))
		}

		for i, v := range a.DenomUnits {
			if (v.Exponent != b.DenomUnits[i].Exponent) || (v.Denom != b.DenomUnits[i].Denom) || !EqualStringSlice(v.Aliases, b.DenomUnits[i].Aliases) {
				return fmt.Errorf("metadata provided has different denom unit from stored, %s ≠ %s", a.DenomUnits[i], b.DenomUnits[i])
			}
		}

		return nil
	}
	return fmt.Errorf("metadata provided is different from stored")
}

// EqualStringSlice checks if two string slices are equal.
func EqualStringSlice(aliasesA, aliasesB []string) bool {
	if len(aliasesA) != len(aliasesB) {
		return false
	}

	for i := 0; i < len(aliasesA); i++ {
		if aliasesA[i] != aliasesB[i] {
			return false
		}
	}

	return true
}

// IsModuleAccount returns true if the given account is a module account
func IsModuleAccount(acc authtypes.AccountI) bool {
	_, isModuleAccount := acc.(authtypes.ModuleAccountI)
	return isModuleAccount
}
