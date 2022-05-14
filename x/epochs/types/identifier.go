package types

import (
	"fmt"
	"strings"
	"time"
)

const (
	// WeekEpochID defines the identifier for weekly epochs
	WeekEpochID = "week"
	// DayEpochID defines the identifier for daily epochs
	DayEpochID = "day"
	// HourEpochID defines the identifier for hourly epochs
	HourEpochID = "hour"
	// WeekEpochDuration defines the duration for weekly epochs
	WeekEpochDuration = time.Hour * 24 * 7
	// DayEpochDuration defines the duration for daily epochs
	DayEpochDuration = time.Hour * 24
	// HourEpochDuration defines the duration for hourly epochs
	HourEpochDuration = time.Hour
)

// IdentifierToDuration maps a human-readable identifier to the appropriate duration
// If genesis is changed, this map needs to be updated
var IdentifierToDuration = map[string]time.Duration{
	WeekEpochID: WeekEpochDuration,
	DayEpochID:  DayEpochDuration,
	HourEpochID: HourEpochDuration,
}

// DurationToIdentifier maps a duration to a human-readable identifier
// If genesis is changed, this map needs to be updated
var DurationToIdentifier = map[time.Duration]string{
	WeekEpochDuration: WeekEpochID,
	DayEpochDuration:  DayEpochID,
	HourEpochDuration: HourEpochID,
}

// ValidateEpochIdentifierInterface performs a stateless
// validation of the epoch ID interface.
func ValidateEpochIdentifierInterface(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if err := ValidateEpochIdentifierString(v); err != nil {
		return err
	}

	return nil
}

// ValidateEpochIdentifierString performs a stateless
// validation of the epoch ID.
func ValidateEpochIdentifierString(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("blank epoch identifier: %s", s)
	}
	return nil
}
