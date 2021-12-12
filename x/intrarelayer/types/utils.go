package types

import (
	"fmt"
	"strings"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// SanitizeERC20Name enforces snake_case and removes all "coin" and "token"
// strings from the ERC20 name.
func SanitizeERC20Name(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " token", "")
	name = strings.ReplaceAll(name, " coin", "")
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

// EqualMetadata checks if all the fields of the provided coin metadata are equal.
func EqualMetadata(a, b banktypes.Metadata) error {
	if a.Base == b.Base && a.Description == b.Description && a.Display == b.Display && a.Name == b.Name && a.Symbol == b.Symbol {
		if len(a.DenomUnits) != len(b.DenomUnits) {
			return fmt.Errorf("metadata provided has different denom units from stored, %d ≠ %d", len(a.DenomUnits), len(b.DenomUnits))
		}

		for i, v := range a.DenomUnits {
			if v != b.DenomUnits[i] {
				return fmt.Errorf("metadata provided has different denom unit from stored, %s ≠ %s", a.DenomUnits[i], b.DenomUnits[i])
			}
		}

		return nil
	}
	return fmt.Errorf("metadata provided is different from stored")
}
