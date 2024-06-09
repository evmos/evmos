// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// ModuleName is the name of the module
	// FIXME: Need another name that doesn't start with `access` otherwise there is a collision in KVSTore keys
	ModuleName = "factory_access_control"
	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName
)

var (
	KeyPrefixRole      = []byte{0x01}
	KeyPrefixRoleAdmin = []byte{0x02}
	KeyPrefixOwner     = []byte{0x03}
	KeyPrefixPauser    = []byte{0x04}
	KeyPrefixPaused    = []byte{0x05}
)

var RoleDefaultAdmin = crypto.Keccak256Hash([]byte("DEFAULT_ADMIN_ROLE"))

func KeyRole(contract common.Address, role common.Hash, account common.Address) []byte {
	key := make([]byte, 2*common.AddressLength+common.HashLength)
	copy(key[:common.AddressLength], contract.Bytes())
	copy(key[common.AddressLength:common.AddressLength+common.HashLength], role.Bytes())
	copy(key[common.AddressLength+common.HashLength:], account.Bytes())
	return key
}

func KeyRoleAdmin(contract common.Address, role common.Hash) []byte {
	key := make([]byte, common.AddressLength+common.HashLength)
	copy(key[:common.AddressLength], contract.Bytes())
	copy(key[common.AddressLength:common.AddressLength+common.HashLength], role.Bytes())
	return key
}
