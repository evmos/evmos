package osmosis_test

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ParseStringAsJSON parses the given string into a JSON object. Returns an error if the string is
// empty or if it is not a valid JSON formatted string.
//
// This function is a readjustment of the Osmosis' Wasm hook:
// https://github.com/osmosis-labs/osmosis/blob/6a28004ab7bf98f21ec22ad9f5e4dcbba78dfa76/x/ibc-hooks/wasm_hook.go#L158-L182
func parseStringAsJSON(memo string) (jsonObject map[string]interface{}, err error) {
	jsonObject = make(map[string]interface{})

	if len(memo) == 0 {
		return nil, fmt.Errorf("string cannot be empty")
	}

	// the jsonObject must be a valid JSON object
	if err := json.Unmarshal([]byte(memo), &jsonObject); err != nil {
		return nil, err
	}

	return jsonObject, nil
}

// ValidateAndParseWasmRoutedMemo check that the given memo is a JSON formatted string and that it
// contains the required keys and fields to be correctly routed by the Osmosis' Wasm hook.
//
// This function is a readjustment of the Osmosis' Wasm hook:
// https://github.com/osmosis-labs/osmosis/blob/6a28004ab7bf98f21ec22ad9f5e4dcbba78dfa76/x/ibc-hooks/wasm_hook.go#L184-L241
func ValidateAndParseWasmRoutedMemo(
	memo string,
	receiver string,
) (err error) {
	metadata, err := parseStringAsJSON(memo)
	if err != nil {
		return err
	}

	_, ok := metadata["wasm"]
	if !ok {
		return nil
	}

	wasmRaw := metadata["wasm"]

	// Make sure the wasm key is a map.
	wasm, ok := wasmRaw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("wasm metadata is not a valid JSON map object")
	}

	contract, ok := wasm["contract"].(string)
	if !ok {
		return fmt.Errorf(`Could not find key wasm["contract"]`)
	}

	_, err = sdk.AccAddressFromBech32(contract)
	if err != nil {
		return fmt.Errorf(`wasm["contract"] is not a valid bech32 address`)
	}

	// The contract and the receiver should be the same for the packet to be valid
	if contract != receiver {
		return fmt.Errorf(`wasm["contract"] should be the same as the receiver of the packet`)
	}

	// Ensure the message key is provided
	if wasm["msg"] == nil {
		return fmt.Errorf(`Could not find key wasm["msg"]`)
	}

	// Make sure the msg key is a map. If it isn't, return an error
	_, ok = wasm["msg"].(map[string]interface{})
	if !ok {
		return fmt.Errorf(`wasm["msg"] is not a map object`)
	}

	// Get the message string by serializing the map
	_, err = json.Marshal(wasm["msg"])
	if err != nil {
		return err
	}

	return nil
}
