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
package eip712

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type eip712MessagePayload struct {
	payload        gjson.Result
	numPayloadMsgs int
	message        map[string]interface{}
}

const (
	PAYLOAD_MSGS = "msgs"
)

// createEIP712MessagePayload generates the EIP-712 message payload corresponding to the input data.
func createEIP712MessagePayload(data []byte) (eip712MessagePayload, error) {
	rawPayload, err := unmarshalBytesToJSONObject(data)
	if err != nil {
		return eip712MessagePayload{}, err
	}

	payload, numPayloadMsgs, err := FlattenPayloadMessages(rawPayload)
	if err != nil {
		return eip712MessagePayload{}, errorsmod.Wrap(err, "failed to flatten payload JSON messages")
	}

	message, ok := payload.Value().(map[string]interface{})
	if !ok {
		return eip712MessagePayload{}, errorsmod.Wrap(errortypes.ErrInvalidType, "failed to parse JSON as map")
	}

	messagePayload := eip712MessagePayload{
		payload:        payload,
		numPayloadMsgs: numPayloadMsgs,
		message:        message,
	}

	return messagePayload, nil
}

// unmarshalBytesToJSONObject converts a bytestream into
// a JSON object, then makes sure the JSON is an object.
func unmarshalBytesToJSONObject(data []byte) (gjson.Result, error) {
	if !gjson.ValidBytes(data) {
		return gjson.Result{}, errorsmod.Wrap(errortypes.ErrJSONUnmarshal, "invalid JSON received")
	}

	payload := gjson.ParseBytes(data)

	if !payload.IsObject() {
		return gjson.Result{}, errorsmod.Wrap(errortypes.ErrJSONUnmarshal, "failed to JSON unmarshal data as object")
	}

	return payload, nil
}

// FlattenPayloadMessages flattens the input payload's messages, representing
// them as key-value pairs of "msg{i}": {Msg}, rather than as an array of Msgs.
// We do this to support messages with different schemas.
func FlattenPayloadMessages(payload gjson.Result) (gjson.Result, int, error) {
	var err error
	flattenedRaw := payload.Raw

	msgs, err := getPayloadMsgs(payload)
	if err != nil {
		return gjson.Result{}, 0, err
	}

	numMsgs := len(msgs)

	for i, msg := range msgs {
		if !msg.IsObject() {
			return gjson.Result{}, 0, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "msg at index %d is not valid JSON: %v", i, msg)
		}

		msgField := flattenedMsgFieldForIndex(i)

		if gjson.Get(flattenedRaw, msgField).Exists() {
			return gjson.Result{}, 0, errorsmod.Wrapf(
				errortypes.ErrInvalidRequest,
				"malformed payload received, did not expect to find key with field %v", msgField,
			)
		}

		flattenedRaw, err = sjson.SetRaw(flattenedRaw, msgField, msg.Raw)
		if err != nil {
			return gjson.Result{}, 0, err
		}
	}

	flattenedRaw, err = sjson.Delete(flattenedRaw, PAYLOAD_MSGS)
	if err != nil {
		return gjson.Result{}, 0, err
	}

	flattenedJSON := gjson.Parse(flattenedRaw)
	return flattenedJSON, numMsgs, nil
}

// getPayloadMsgs processes and returns the payload messages as a JSON array.
func getPayloadMsgs(payload gjson.Result) ([]gjson.Result, error) {
	rawMsgs := payload.Get(PAYLOAD_MSGS)

	if !rawMsgs.Exists() {
		return nil, errorsmod.Wrap(errortypes.ErrInvalidRequest, "no messages found in payload, unable to parse")
	}

	if !rawMsgs.IsArray() {
		return nil, errorsmod.Wrap(errortypes.ErrInvalidRequest, "expected type array of messages, cannot parse")
	}

	return rawMsgs.Array(), nil
}

// flattenedMsgFieldForIndex returns the payload field for a given message post-flattening.
// e.g. msgs[2] is moved to 'msg2'
func flattenedMsgFieldForIndex(i int) string {
	return fmt.Sprintf("msg%d", i)
}
