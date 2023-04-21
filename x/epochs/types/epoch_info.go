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
	"errors"
	"fmt"
	"strings"
)

// StartInitialEpoch sets the epoch info fields to their start values
func (ei *EpochInfo) StartInitialEpoch() {
	ei.EpochCountingStarted = true
	ei.CurrentEpoch = 1
	ei.CurrentEpochStartTime = ei.StartTime
}

// EndEpoch increments the epoch counter and resets the epoch start time
func (ei *EpochInfo) EndEpoch() {
	ei.CurrentEpoch++
	ei.CurrentEpochStartTime = ei.CurrentEpochStartTime.Add(ei.Duration)
}

// Validate performs a stateless validation of the epoch info fields
func (ei EpochInfo) Validate() error {
	if strings.TrimSpace(ei.Identifier) == "" {
		return errors.New("epoch identifier cannot be blank")
	}
	if ei.Duration == 0 {
		return errors.New("epoch duration cannot be 0")
	}
	if ei.CurrentEpoch < 0 {
		return fmt.Errorf("current epoch cannot be negative: %d", ei.CurrentEpochStartHeight)
	}
	if ei.CurrentEpochStartHeight < 0 {
		return fmt.Errorf("current epoch start height cannot be negative: %d", ei.CurrentEpochStartHeight)
	}
	return nil
}
