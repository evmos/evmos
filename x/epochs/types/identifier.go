package types

import (
	"fmt"
	"strings"
)

<<<<<<< HEAD
// ValidateEpochIdentifierInterface checks if the identifier is blank
=======
const (
	// WeekEpochID defines the identifier for weekly epochs
	WeekEpochID = "week"
	// DayEpochID defines the identifier for daily epochs
	DayEpochID = "day"
	// HourEpochID defines the identifier for hourly epochs
	HourEpochID = "hour"
)

// ValidateEpochIdentifierInterface performs a stateless
// validation of the epoch ID interface.
>>>>>>> ffff90cc83bb057af68ca2f5d9b6007df3161298
func ValidateEpochIdentifierInterface(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return ValidateEpochIdentifierString(v)
}

<<<<<<< HEAD
// ValidateEpochIdentifierString checks if the identifier is blank
=======
// ValidateEpochIdentifierInterface performs a stateless
// validation of the epoch ID.
>>>>>>> ffff90cc83bb057af68ca2f5d9b6007df3161298
func ValidateEpochIdentifierString(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("blank epoch identifier")
	}
	return nil
}
