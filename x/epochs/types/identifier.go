package types

import (
	"fmt"
	"strings"
)

// ValidateEpochIdentifierInterface checks if the identifier is blank
func ValidateEpochIdentifierInterface(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return ValidateEpochIdentifierString(v)
}

// ValidateEpochIdentifierString checks if the identifier is blank
func ValidateEpochIdentifierString(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("blank epoch identifier")
	}
	return nil
}
