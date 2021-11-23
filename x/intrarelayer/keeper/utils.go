package keeper

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func equalMetadata(a, b banktypes.Metadata) bool {
	if a.Base == b.Base && a.Description == b.Description && a.Display == b.Display && a.Name == b.Name && a.Symbol == b.Symbol {
		if len(a.DenomUnits) != len(b.DenomUnits) {
			return false
		}
		for i, v := range a.DenomUnits {
			if v != b.DenomUnits[i] {
				return false
			}
		}
		return true
	}
	return false
}
