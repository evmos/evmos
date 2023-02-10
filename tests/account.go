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

package tests

import (
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v6/testing/mock"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

type AccountType int8

const (
	AccountTypeEOA = iota
	AccountTypeContract
	AccountTypeValidator
)

type Account struct {
	Address    sdk.AccAddress
	AddressHex common.Address
	PubKey     cryptotypes.PubKey
	TmPubKey   tmcrypto.PubKey
	PrivKey    cryptotypes.PrivKey
	Type       AccountType
}

func NewEOAAccount(_ testing.TB) Account {
	addr, privKey := NewAddrKey()
	return Account{
		Address:    addr.Bytes(),
		AddressHex: addr,
		PubKey:     privKey.PubKey(),
		PrivKey:    privKey,
		Type:       AccountTypeEOA,
	}
}

func NewValidatorAccount(t testing.TB) Account {
	privVal := mock.NewPV()
	pubKey := privVal.PrivKey.PubKey()
	tmPubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	addr := sdk.AccAddress(pubKey.Address())
	return Account{
		Address:    addr,
		AddressHex: common.BytesToAddress(addr.Bytes()),
		PubKey:     privVal.PrivKey.PubKey(),
		PrivKey:    privVal.PrivKey,
		Type:       AccountTypeValidator,
		TmPubKey:   tmPubKey,
	}
}
