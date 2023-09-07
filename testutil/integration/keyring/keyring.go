package keyring

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	utiltx "github.com/evmos/evmos/v14/testutil/tx"
)

type Account struct {
	Addr    common.Address
	AccAddr sdktypes.AccAddress
	Priv    cryptotypes.PrivKey
}

func NewAccount() Account {
	addr, privKey := utiltx.NewAddrKey()
	return Account{
		Addr:    addr,
		AccAddr: sdktypes.AccAddress(addr.Bytes()),
		Priv:    privKey,
	}
}

type Keyring interface {
	// GetPrivKey returns the private key of the specified account.
	GetPrivKey(index int) cryptotypes.PrivKey
	// GetAddr returns the address of the specified account.
	GetAddr(index int) common.Address
	// GetAccAddr returns the sdk address of the specified account.
	GetAccAddr(index int) sdktypes.AccAddress
	// GetAccount returns the account of the specified address.
	GetAccount(index int) Account

	// Sign signs message with the specified account.
	Sign(index int, msg []byte) ([]byte, error)
}

type IntegrationKeyring struct {
	accounts []Account
}

var _ Keyring = (*IntegrationKeyring)(nil)

func NewKeyring(nAccs int) IntegrationKeyring {
	accs := make([]Account, 0, nAccs)
	for i := 0; i < nAccs; i++ {
		acc := NewAccount()
		accs = append(accs, acc)
	}
	return IntegrationKeyring{
		accounts: accs,
	}
}

func (kr *IntegrationKeyring) GetPrivKey(index int) cryptotypes.PrivKey {
	return kr.accounts[index].Priv
}

func (kr *IntegrationKeyring) GetAddr(index int) common.Address {
	return kr.accounts[index].Addr
}

func (kr *IntegrationKeyring) GetAccAddr(index int) sdktypes.AccAddress {
	return kr.accounts[index].AccAddr
}

func (kr *IntegrationKeyring) GetAccount(index int) Account {
	return kr.accounts[index]
}

func (kr *IntegrationKeyring) Sign(index int, msg []byte) ([]byte, error) {
	privKey := kr.GetPrivKey(index)
	if privKey == nil {
		return nil, fmt.Errorf("no private key for account %d", index)
	}
	return privKey.Sign(msg)
}
