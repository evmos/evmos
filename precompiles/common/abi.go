// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package common

import (
	"embed"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	contractutils "github.com/evmos/evmos/v19/contracts/utils"
)

// MakeTopic converts a filter query argument into a filter topic.
// NOTE: This was copied from accounts/abi/topics.go
func MakeTopic(rule interface{}) (common.Hash, error) {
	var topic common.Hash

	// Try to generate the topic based on simple types
	switch rule := rule.(type) {
	case common.Hash:
		copy(topic[:], rule[:])
	case common.Address:
		copy(topic[common.HashLength-common.AddressLength:], rule[:])
	case *big.Int:
		blob := rule.Bytes()
		copy(topic[common.HashLength-len(blob):], blob)
	case bool:
		if rule {
			topic[common.HashLength-1] = 1
		}
	case int8:
		copy(topic[:], genIntType(int64(rule), 1))
	case int16:
		copy(topic[:], genIntType(int64(rule), 2))
	case int32:
		copy(topic[:], genIntType(int64(rule), 4))
	case int64:
		copy(topic[:], genIntType(rule, 8))
	case uint8:
		blob := new(big.Int).SetUint64(uint64(rule)).Bytes()
		copy(topic[common.HashLength-len(blob):], blob)
	case uint16:
		blob := new(big.Int).SetUint64(uint64(rule)).Bytes()
		copy(topic[common.HashLength-len(blob):], blob)
	case uint32:
		blob := new(big.Int).SetUint64(uint64(rule)).Bytes()
		copy(topic[common.HashLength-len(blob):], blob)
	case uint64:
		blob := new(big.Int).SetUint64(rule).Bytes()
		copy(topic[common.HashLength-len(blob):], blob)
	case string:
		hash := crypto.Keccak256Hash([]byte(rule))
		copy(topic[:], hash[:])
	case []byte:
		hash := crypto.Keccak256Hash(rule)
		copy(topic[:], hash[:])

	default:
		// todo(rjl493456442) according solidity documentation, indexed event
		// parameters that are not value types i.e. arrays and structs are not
		// stored directly but instead a keccak256-hash of an encoding is stored.
		//
		// We only convert stringS and bytes to hash, still need to deal with
		// array(both fixed-size and dynamic-size) and struct.

		// Attempt to generate the topic from funky types
		val := reflect.ValueOf(rule)
		switch {
		// static byte array
		case val.Kind() == reflect.Array && reflect.TypeOf(rule).Elem().Kind() == reflect.Uint8:
			reflect.Copy(reflect.ValueOf(topic[:val.Len()]), val)
		default:
			return topic, fmt.Errorf("unsupported indexed type: %T", rule)
		}
	}

	return topic, nil
}

// UnpackLog unpacks a retrieved log into the provided output structure.
func UnpackLog(contractABI abi.ABI, out interface{}, event string, log ethtypes.Log) error {
	if log.Topics[0] != contractABI.Events[event].ID {
		return fmt.Errorf("event signature mismatch")
	}
	if len(log.Data) > 0 {
		if err := contractABI.UnpackIntoInterface(out, event, log.Data); err != nil {
			return err
		}
	}
	var indexed abi.Arguments
	for _, arg := range contractABI.Events[event].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return abi.ParseTopics(out, indexed, log.Topics[1:])
}

// NOTE: This was copied from accounts/abi/topics.go
func genIntType(rule int64, size uint) []byte {
	var topic [common.HashLength]byte
	if rule < 0 {
		// if a rule is negative, we need to put it into two's complement.
		// extended to common.HashLength bytes.
		topic = [common.HashLength]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}
	}
	for i := uint(0); i < size; i++ {
		topic[common.HashLength-i-1] = byte(rule >> (i * 8))
	}
	return topic[:]
}

// PackNum packs the given number (using the reflect value) and will cast it to appropriate number representation.
func PackNum(value reflect.Value) []byte {
	switch kind := value.Kind(); kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return math.U256Bytes(new(big.Int).SetUint64(value.Uint()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return math.U256Bytes(big.NewInt(value.Int()))
	case reflect.Ptr:
		return math.U256Bytes(new(big.Int).Set(value.Interface().(*big.Int)))
	default:
		panic("abi: fatal error")
	}
}

// LoadABI read the ABI file described by the path and parse it as JSON.
func LoadABI(fs embed.FS, path string) (abi.ABI, error) {
	abiBz, err := fs.ReadFile(path)
	if err != nil {
		return abi.ABI{}, fmt.Errorf("error loading the ABI %s", err)
	}

	contract, err := contractutils.ConvertPrecompileHardhatBytesToCompiledContract(abiBz)
	if err != nil {
		return abi.ABI{}, fmt.Errorf(ErrInvalidABI, err)
	}

	return contract.ABI, nil
}
