// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keyring

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	utiltx "github.com/evmos/evmos/v17/testutil/tx"
)

type Key struct {
	Addr    common.Address
	AccAddr sdktypes.AccAddress
	Priv    cryptotypes.PrivKey
}

func NewKey() Key {
	addr, privKey := utiltx.NewAddrKey()
	return Key{
		Addr:    addr,
		AccAddr: sdktypes.AccAddress(addr.Bytes()),
		Priv:    privKey,
	}
}

type Keyring interface {
	// GetPrivKey returns the private key of the account at the given keyring index.
	GetPrivKey(index int) cryptotypes.PrivKey
	// GetAddr returns the address of the account at the given keyring index.
	GetAddr(index int) common.Address
	// GetAccAddr returns the SDK address of the account at the given keyring index.
	GetAccAddr(index int) sdktypes.AccAddress
	// GetAllAccAddrs returns all the SDK addresses of the accounts in the keyring.
	GetAllAccAddrs() []sdktypes.AccAddress
	// GetKey returns the key at the given keyring index
	GetKey(index int) Key

	// AddKey adds a new account to the keyring
	AddKey() int

	// Sign signs message with the specified account.
	Sign(index int, msg []byte) ([]byte, error)
}

// IntegrationKeyring is a keyring designed for integration tests.
type IntegrationKeyring struct {
	keys []Key
}

var _ Keyring = (*IntegrationKeyring)(nil)

// New returns a new keyring with nAccs accounts.
func New(nAccs int) Keyring {
	accs := make([]Key, 0, nAccs)
	for i := 0; i < nAccs; i++ {
		acc := NewKey()
		accs = append(accs, acc)
	}
	return &IntegrationKeyring{
		keys: accs,
	}
}

// GetPrivKey returns the private key of the specified account.
func (kr *IntegrationKeyring) GetPrivKey(index int) cryptotypes.PrivKey {
	return kr.keys[index].Priv
}

// GetAddr returns the address of the specified account.
func (kr *IntegrationKeyring) GetAddr(index int) common.Address {
	return kr.keys[index].Addr
}

// GetAccAddr returns the sdk address of the specified account.
func (kr *IntegrationKeyring) GetAccAddr(index int) sdktypes.AccAddress {
	return kr.keys[index].AccAddr
}

// GetAllAccAddrs returns all the sdk addresses of the accounts in the keyring.
func (kr *IntegrationKeyring) GetAllAccAddrs() []sdktypes.AccAddress {
	accs := make([]sdktypes.AccAddress, 0, len(kr.keys))
	for _, key := range kr.keys {
		accs = append(accs, key.AccAddr)
	}
	return accs
}

// GetKey returns the key specified by index
func (kr *IntegrationKeyring) GetKey(index int) Key {
	return kr.keys[index]
}

// AddKey adds a new account to the keyring. It returns the index for the key
func (kr *IntegrationKeyring) AddKey() int {
	acc := NewKey()
	index := len(kr.keys)
	kr.keys = append(kr.keys, acc)
	return index
}

// Sign signs message with the specified key.
func (kr *IntegrationKeyring) Sign(index int, msg []byte) ([]byte, error) {
	privKey := kr.GetPrivKey(index)
	if privKey == nil {
		return nil, fmt.Errorf("no private key for account %d", index)
	}
	return privKey.Sign(msg)
}
