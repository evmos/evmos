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
	"time"
)

// NewGenesisState creates a new genesis state instance
func NewGenesisState(epochs []EpochInfo) *GenesisState {
	return &GenesisState{Epochs: epochs}
}

// DefaultGenesisState returns the default epochs genesis state
func DefaultGenesisState() *GenesisState {
	epochs := []EpochInfo{
		{
			Identifier:              WeekEpochID,
			StartTime:               time.Time{},
			Duration:                time.Hour * 24 * 7,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
		{
			Identifier:              DayEpochID,
			StartTime:               time.Time{},
			Duration:                time.Hour * 24,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
	}
	return NewGenesisState(epochs)
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	epochIdentifiers := make(map[string]bool)

	for _, epoch := range gs.Epochs {
		if epochIdentifiers[epoch.Identifier] {
			return fmt.Errorf("duplicated epoch entry %s", epoch.Identifier)
		}
		if err := epoch.Validate(); err != nil {
			return err
		}
		epochIdentifiers[epoch.Identifier] = true
	}

	return nil
}
