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
	"bytes"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	errorsmod "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/tidwall/gjson"
)

const (
	ROOT_PREFIX = "_"
	TYPE_PREFIX = "Type"

	TX_FIELD   = "Tx"
	ETH_BOOL   = "bool"
	ETH_INT64  = "int64"
	ETH_STRING = "string"

	MSG_TYPE = "type"

	MAX_TYPEDEF_DUPLICATES = 1000
)

// getEIP712Types creates and returns the EIP-712 types
// for the given message payload.
func createEIP712Types(messagePayload eip712MessagePayload) (apitypes.Types, error) {
	eip712Types := apitypes.Types{
		"EIP712Domain": {
			{
				Name: "name",
				Type: "string",
			},
			{
				Name: "version",
				Type: "string",
			},
			{
				Name: "chainId",
				Type: "uint256",
			},
			{
				Name: "verifyingContract",
				Type: "string",
			},
			{
				Name: "salt",
				Type: "string",
			},
		},
		"Tx": {
			{Name: "account_number", Type: "string"},
			{Name: "chain_id", Type: "string"},
			{Name: "fee", Type: "Fee"},
			{Name: "memo", Type: "string"},
			{Name: "sequence", Type: "string"},
			// Note timeout_height was removed because it was not getting filled with the legacyTx
		},
		"Fee": {
			{Name: "amount", Type: "Coin[]"},
			{Name: "gas", Type: "string"},
		},
		"Coin": {
			{Name: "denom", Type: "string"},
			{Name: "amount", Type: "string"},
		},
	}

	for i := 0; i < messagePayload.numPayloadMsgs; i++ {
		field := msgFieldForIndex(i)
		msg := messagePayload.payload.Get(field)

		if !msg.IsObject() {
			return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "message is not valid JSON, cannot parse types")
		}

		if err := addMsgTypesToRoot(eip712Types, field, msg); err != nil {
			return nil, err
		}
	}

	return eip712Types, nil
}

// addMsgTypesToRoot adds all types for the given message
// to eip712Types, recursively handling fields as necessary.
func addMsgTypesToRoot(eip712Types apitypes.Types, msgField string, msgJSON gjson.Result) (err error) {
	defer doRecover(&err)

	msgRootType, err := msgRootType(msgJSON)
	if err != nil {
		return err
	}

	msgTypeDef, err := recursivelyAddTypesToRoot(eip712Types, msgRootType, ROOT_PREFIX, msgJSON)
	if err != nil {
		return err
	}

	addMsgTypeDefToTxSchema(eip712Types, msgField, msgTypeDef)

	return nil
}

// msgRootType parses the message and returns the formatted
// type signature corresponding to the message type.
func msgRootType(msgJSON gjson.Result) (string, error) {
	msgType := msgJSON.Get(MSG_TYPE).Str
	if msgType == "" {
		// .Str is empty for arrays and objects
		return "", errorsmod.Wrap(errortypes.ErrInvalidType, "malformed message type value, expected type string")
	}

	// For example, convert cosmos-sdk/MsgSend to TypeMsgSend
	typeTokens := strings.Split(msgType, "/")
	msgSignature := typeTokens[len(typeTokens)-1]
	msgType = fmt.Sprintf("%v%v", TYPE_PREFIX, msgSignature)

	return msgType, nil
}

// addMsgTypeDefToTxSchema adds the message's field-type pairing
// to the Tx schema.
func addMsgTypeDefToTxSchema(eip712Types apitypes.Types, msgField string, msgTypeDef string) {
	eip712Types[TX_FIELD] = append(eip712Types[TX_FIELD], apitypes.Type{
		Name: msgField,
		Type: msgTypeDef,
	})
}

// recursivelyAddTypesToRoot walks all types in the given map
// and recursively adds sub-maps as new types when necessary.
// It adds all type definitions to typeMap, then returns a key
// to the json object's type definition within the map.
func recursivelyAddTypesToRoot(
	typeMap apitypes.Types,
	rootType string,
	prefix string,
	payload gjson.Result,
) (string, error) {
	var typeDef string

	// Must sort JSON keys for deterministic type generation
	jsonFieldNames, err := sortedJSONKeys(payload)
	if err != nil {
		return "", errorsmod.Wrap(err, "unable to sort object keys")
	}

	typesToAdd := []apitypes.Type{}

	for _, fieldName := range jsonFieldNames {
		field := payload.Get(fieldName)
		if !field.Exists() {
			continue
		}

		// Handle array types.
		isCollection := false
		if field.IsArray() {
			if len(field.Array()) == 0 {
				// Arbitrarily add string[] type to handle empty arrays,
				// since we cannot access the underlying object.
				typesToAdd = append(typesToAdd, apitypes.Type{
					Name: fieldName,
					Type: "string[]",
				})

				continue
			}

			field = field.Array()[0]
			isCollection = true
		}

		ethType := getEthTypeForJSON(field)

		// Handle JSON primitive types.
		if ethType != "" {
			if isCollection {
				ethType += "[]"
			}
			typesToAdd = append(typesToAdd, apitypes.Type{
				Name: fieldName,
				Type: ethType,
			})

			continue
		}

		// Handle object types. Note that nested array types are not supported
		// in EIP-712, so we must exclude that case.
		if field.IsObject() {
			// Recursively parse and update root types for the sub-field.
			subFieldPrefix := fmt.Sprintf("%s.%s", prefix, fieldName)
			fieldTypeDef, err := recursivelyAddTypesToRoot(typeMap, rootType, subFieldPrefix, field)
			if err != nil {
				return "", err
			}

			fieldTypeDef = sanitizeTypedef(fieldTypeDef)
			if isCollection {
				fieldTypeDef += "[]"
			}

			typesToAdd = append(typesToAdd, apitypes.Type{
				Name: fieldName,
				Type: fieldTypeDef,
			})

			continue
		}
	}

	if prefix == ROOT_PREFIX {
		typeDef = rootType
	} else {
		typeDef = sanitizeTypedef(prefix)
	}

	return addTypesToRoot(typeMap, typeDef, typesToAdd)
}

// addTypesToRoot attempts to add the types to the root at key
// typeDef and returns the key at which the types are present,
// or an error if they cannot be added. If the typeDef key is a
// duplicate, we return the key corresponding to an identical copy
// if present (without modifying the structure). Otherwise, we insert
// the types at the next available typeDef-{n} field. We do this to
// support identically named payloads with different schemas.
func addTypesToRoot(rootTypes apitypes.Types, typeDef string, types []apitypes.Type) (string, error) {
	var indexedTypeDef string

	numDuplicates := 0

	for {
		indexedTypeDef = fmt.Sprintf("%v%d", typeDef, numDuplicates)
		existingTypes, ok := rootTypes[indexedTypeDef]

		// Found identical duplicate
		if ok && typesAreEqual(types, existingTypes) {
			return indexedTypeDef, nil
		}

		// Found no element
		if !ok {
			break
		}

		numDuplicates++

		if numDuplicates == MAX_TYPEDEF_DUPLICATES {
			return "", errorsmod.Wrap(errortypes.ErrInvalidRequest, "exceeded maximum number of duplicates for a single type definition")
		}
	}

	// Add new type to root at current duplicate index
	rootTypes[indexedTypeDef] = types
	return indexedTypeDef, nil
}

// typesAreEqual compares two apitypes.Type arrays
// and returns a boolean indicating whether they have
// the same values.
// It assumes both arrays are in the same sorted order.
func typesAreEqual(types1 []apitypes.Type, types2 []apitypes.Type) bool {
	if len(types1) != len(types2) {
		return false
	}

	for i := 0; i < len(types1); i++ {
		if types1[i].Name != types2[i].Name || types1[i].Type != types2[i].Type {
			return false
		}
	}

	return true
}

// sortedJSONKeys returns the sorted JSON keys for the input object.
func sortedJSONKeys(json gjson.Result) ([]string, error) {
	if !json.IsObject() {
		return nil, errorsmod.Wrap(errortypes.ErrInvalidType, "expected JSON map to parse")
	}

	jsonMap := json.Map()

	keys := make([]string, len(jsonMap))
	i := 0
	for k := range jsonMap {
		keys[i] = k
		i++
	}

	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(keys[i], keys[j]) > 0
	})

	return keys, nil
}

// _.foo_bar.baz -> TypeFooBarBaz
//
// Since Geth does not tolerate complex EIP-712 type names, we need to sanitize
// the inputs.
func sanitizeTypedef(str string) string {
	buf := new(bytes.Buffer)
	caser := cases.Title(language.English, cases.NoLower)
	parts := strings.Split(str, ".")

	for _, part := range parts {
		if part == ROOT_PREFIX {
			buf.WriteString(TYPE_PREFIX)
			continue
		}

		subparts := strings.Split(part, ROOT_PREFIX)
		for _, subpart := range subparts {
			buf.WriteString(caser.String(subpart))
		}
	}

	return buf.String()
}

// getEthTypeForJSON converts a JSON type to an Ethereum type.
// It returns an empty string for Objects, Arrays, or Null.
// See https://github.com/ethereum/EIPs/blob/master/EIPS/eip-712.md for more.
func getEthTypeForJSON(json gjson.Result) string {
	switch json.Type {
	case gjson.True, gjson.False:
		return ETH_BOOL
	case gjson.Number:
		return ETH_INT64
	case gjson.String:
		return ETH_STRING
	case gjson.JSON:
		// Array or Object type
		return ""
	default:
		return ""
	}
}

func doRecover(err *error) {
	if r := recover(); r != nil {
		if e, ok := r.(error); ok {
			e = errorsmod.Wrap(e, "panicked with error")
			*err = e
			return
		}

		*err = fmt.Errorf("%v", r)
	}
}
