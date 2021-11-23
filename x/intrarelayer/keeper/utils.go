package keeper

import (
	"fmt"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func equalMetadata(a, b banktypes.Metadata) error {
	if a.Base == b.Base && a.Description == b.Description && a.Display == b.Display && a.Name == b.Name && a.Symbol == b.Symbol {
		if len(a.DenomUnits) != len(b.DenomUnits) {
			return fmt.Errorf("metadata provided has different denom unit from stored")
		}
		for i, v := range a.DenomUnits {
			if v != b.DenomUnits[i] {
				return fmt.Errorf("metadata provided has different denom unit from stored")
			}
		}
		return nil
	}
	return fmt.Errorf("metadata provided is different from stored")
}
