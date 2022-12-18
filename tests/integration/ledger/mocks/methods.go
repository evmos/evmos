package mocks

import (
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mock "github.com/stretchr/testify/mock"
)

// original: Close() error
func RegisterClose(s *SECP256K1) {
	s.On("Close").Return(nil)
}

// original: GetPublicKeySECP256K1([]uint32) ([]byte, error)
func RegisterGetPublicKeySECP256K1(s *SECP256K1, pubKey types.PubKey) {
	s.On("GetPublicKeySECP256K1", mock.AnythingOfType("[]uint32")).
		Return(pubKey.Bytes(), nil)
}

// original: GetAddressPubKeySECP256K1([]uint32, string) ([]byte, string, error)
func RegisterGetAddressPubKeySECP256K1(s *SECP256K1, accAddr sdk.AccAddress, pubKey types.PubKey) {
	s.On(
		"GetAddressPubKeySECP256K1",
		mock.AnythingOfType("[]uint32"),
		mock.AnythingOfType("string"),
	).Return(pubKey.Bytes(), accAddr.String(), nil)
}

// original: SignSECP256K1([]uint32, []byte) ([]byte, error)
func RegisterSignSECP256K1(s *SECP256K1, f func([]uint32, []byte) ([]byte, error), e error) {
	s.On("SignSECP256K1", mock.AnythingOfType("[]uint32"), mock.AnythingOfType("[]uint8")).Return(f, e)
}

func RegisterSignSECP256K1Error(s *SECP256K1) {
	s.On("SignSECP256K1", mock.AnythingOfType("[]uint32"), mock.AnythingOfType("[]uint8")).Return(nil, ErrMockedSigning)
}
