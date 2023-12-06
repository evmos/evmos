package osmosis_test

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// jsonStringHasKey parses the memo as a json object and checks if it contains the key.
//
// This function comes from the Osmosis' Wasm hook:
// https://github.com/osmosis-labs/osmosis/blob/6a28004ab7bf98f21ec22ad9f5e4dcbba78dfa76/x/ibc-hooks/wasm_hook.go#L158-L182
func jsonStringHasKey(memo, key string) (found bool, jsonObject map[string]interface{}) {
	jsonObject = make(map[string]interface{})

	// If there is no memo, the packet was either sent with an earlier version of IBC, or the memo was
	// intentionally left blank. Nothing to do here. Ignore the packet and pass it down the stack.
	if len(memo) == 0 {
		return false, jsonObject
	}

	// the jsonObject must be a valid JSON object
	err := json.Unmarshal([]byte(memo), &jsonObject)
	if err != nil {
		return false, jsonObject
	}

	// If the key doesn't exist, there's nothing to do on this hook. Continue by passing the packet
	// down the stack
	_, ok := jsonObject[key]
	if !ok {
		return false, jsonObject
	}

	return true, jsonObject
}

// ValidateAndParseWasmRoutedMemo check that the given memo is a JSON formatted string and that it
// contains the required keys and fields to be correctly routed by the Osmosis' Wasm hook.
//
// This function is a readjustment of the Osmosis' Wasm hook:
// https://github.com/osmosis-labs/osmosis/blob/6a28004ab7bf98f21ec22ad9f5e4dcbba78dfa76/x/ibc-hooks/wasm_hook.go#L184-L241
func ValidateAndParseWasmRoutedMemo(
	packet string,
	receiver string,
) (err error) {
	isWasm, metadata := jsonStringHasKey(packet, "wasm")
	if !isWasm {
		return fmt.Errorf("string is not a valid wasm targeted memo")
	}

	wasmRaw := metadata["wasm"]
	wasm, ok := wasmRaw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("error in getting the wasm field")
	}

	contract, ok := wasm["contract"].(string)
	if !ok {
		return fmt.Errorf(`could not find key wasm["contract"]`)
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
		return fmt.Errorf(`could not find key wasm["msg"]`)
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
