// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"fmt"
	"time"
)

// ParamsKey params store key
var ParamsKey = []byte("Params")

// DefaultPacketTimeoutDuration defines the default packet timeout for outgoing
// IBC transfers
var (
	DefaultEnableRecovery        = true
	DefaultPacketTimeoutDuration = 4 * time.Hour
)

// NewParams creates a new Params instance
func NewParams(
	enableRecovery bool, timeoutDuration time.Duration,
) Params {
	return Params{
		EnableRecovery:        enableRecovery,
		PacketTimeoutDuration: timeoutDuration,
	}
}

// DefaultParams defines the default params for the recovery module
func DefaultParams() Params {
	return Params{
		EnableRecovery:        DefaultEnableRecovery,
		PacketTimeoutDuration: DefaultPacketTimeoutDuration,
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateDuration(i interface{}) error {
	duration, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if duration < 0 {
		return fmt.Errorf("packet timout duration cannot be negative")
	}

	return nil
}

// Validate checks that the fields have valid values
func (p Params) Validate() error {
	if err := validateDuration(p.PacketTimeoutDuration); err != nil {
		return err
	}

	return validateBool(p.EnableRecovery)
}
