// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

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
