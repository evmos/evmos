package mocks

import (
	"crypto/ecdsa"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	mock "github.com/stretchr/testify/mock"
)

var ErrMockedSigning = errors.New("catched signing")

// original: Close() error
func RegisterClose(s *SECP256K1) {
	s.On("Close").Return(nil)
}

// original: GetPublicKeySECP256K1([]uint32) ([]byte, error)
func RegisterGetPublicKeySECP256K1(s *SECP256K1, pubKey *ecdsa.PublicKey) {
	s.On("GetPublicKeySECP256K1", mock.AnythingOfType("[]uint32")).
		Return(crypto.FromECDSAPub(pubKey), nil)
}

// original: GetAddressPubKeySECP256K1([]uint32, string) ([]byte, string, error)
func RegisterGetAddressPubKeySECP256K1(s *SECP256K1, accAddr sdk.AccAddress, pubKey *ecdsa.PublicKey) {
	s.On(
		"GetAddressPubKeySECP256K1",
		mock.AnythingOfType("[]uint32"),
		mock.AnythingOfType("string"),
	).Return(crypto.FromECDSAPub(pubKey), accAddr.String(), nil)
}

// original: SignSECP256K1([]uint32, []byte) ([]byte, error)
func RegisterSignSECP256K1(s *SECP256K1) {
	s.On("SignSECP256K1", mock.AnythingOfType("[]uint32"), mock.AnythingOfType("[]uint8")).
		Return(func(_ []uint32, msg []byte) ([]byte, error) {
			return msg, ErrMockedSigning
		})
}
