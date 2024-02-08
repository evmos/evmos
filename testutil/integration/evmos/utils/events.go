// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package utils

import sdktypes "github.com/cosmos/cosmos-sdk/types"

// ContainsEventType returns true if the given events contain the given eventType.
func ContainsEventType(events sdktypes.Events, eventType string) bool {
	for _, event := range events {
		if event.Type == eventType {
			return true
		}
	}
	return false
}
