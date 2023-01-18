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

package mocks

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mock "github.com/stretchr/testify/mock"
)

// ---------------------------------------
// Methods registrtry for SECP256k1

// original: Close() error
func MClose(s *SECP256K1) {
	s.On("Close").Return(nil)
}

// original: GetPublicKeySECP256K1([]uint32) ([]byte, error)
func MGetPublicKeySECP256K1(s *SECP256K1, pubKey types.PubKey) {
	s.On("GetPublicKeySECP256K1", mock.AnythingOfType("[]uint32")).
		Return(pubKey.Bytes(), nil)
}

// original: GetAddressPubKeySECP256K1([]uint32, string) ([]byte, string, error)
func MGetAddressPubKeySECP256K1(s *SECP256K1, accAddr sdk.AccAddress, pubKey types.PubKey) {
	s.On(
		"GetAddressPubKeySECP256K1",
		mock.AnythingOfType("[]uint32"),
		mock.AnythingOfType("string"),
	).Return(pubKey.Bytes(), accAddr.String(), nil)
}

// original: SignSECP256K1([]uint32, []byte) ([]byte, error)
func MSignSECP256K1(s *SECP256K1, f func([]uint32, []byte) ([]byte, error), e error) {
	s.On("SignSECP256K1", mock.AnythingOfType("[]uint32"), mock.AnythingOfType("[]uint8")).Return(f, e)
}

// ---------------------------------------
// Methods registrtry for AccountRetriever

// original:  GetAccount(_ client.Context, _ sdk.AccAddress) (client.Account, error)
func MGetAccount(m *AccountRetriever, acc client.Account, e error) {
	m.On("GetAccount", mock.Anything, mock.Anything).Return(acc, e)
}

// original: EnsureExists(client.Context, ypes.AccAddress) error
func MEnsureExist(m *AccountRetriever, e error) {
	m.On("EnsureExists", mock.Anything, mock.Anything).Return(e)
}

// original: GetAccountNumberSequence(client.Context, types.AccAddress) (uint64, uint64, error)
func MGetAccountNumberSequence(m *AccountRetriever, seq, num uint64, e error) {
	m.On("GetAccountNumberSequence", mock.Anything, mock.Anything).Return(seq, num, e)
}
