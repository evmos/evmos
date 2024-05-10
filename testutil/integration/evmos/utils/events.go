// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package utils

import (
	"errors"
	"fmt"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// ContainsEventType returns true if the given events contain the given eventType.
func ContainsEventType(events []abcitypes.Event, eventType string) bool {
	event := GetEventType(events, eventType)
	return event != nil
}

// GetEventType returns the given events if found
// Otherwise returns nil
func GetEventType(events []abcitypes.Event, eventType string) *abcitypes.Event {
	for _, event := range events {
		if event.Type == eventType {
			return &event
		}
	}
	return nil
}

// GetEventAttributeValue returns the value for the required
// attribute key
func GetEventAttributeValue(event abcitypes.Event, attrKey string) string {
	for _, attr := range event.Attributes {
		if attr.Key == attrKey {
			return attr.Value
		}
	}
	return ""
}

// GetFeesFromEvents returns the fees value for the
// specified events
func GetFeesFromEvents(events []abcitypes.Event) (sdktypes.DecCoins, error) {
	event := GetEventType(events, sdktypes.EventTypeTx)
	if event == nil {
		return sdktypes.DecCoins{}, errors.New("tx event not found")
	}
	feeStr := GetEventAttributeValue(*event, sdktypes.AttributeKeyFee)
	feeCoins, err := sdktypes.ParseDecCoins(feeStr)
	if err != nil {
		return sdktypes.DecCoins{}, fmt.Errorf("invalid fees: %v. got %s", err, feeStr)
	}
	return feeCoins, nil
}
