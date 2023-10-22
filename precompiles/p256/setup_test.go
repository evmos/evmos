// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package p256_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/cometbft/cometbft/crypto"
	"github.com/evmos/evmos/v15/precompiles/p256"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"
)

var s *PrecompileTestSuite

type PrecompileTestSuite struct {
	suite.Suite
	p256Priv   *ecdsa.PrivateKey
	precompile *p256.Precompile
}

func TestPrecompileTestSuite(t *testing.T) {
	s = new(PrecompileTestSuite)
	suite.Run(t, s)
}

func (s *PrecompileTestSuite) SetupTest() {
	p256Priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	s.Require().NoError(err)
	s.p256Priv = p256Priv
	s.precompile = &p256.Precompile{}
}

func signMsg(msg []byte, priv *ecdsa.PrivateKey) []byte {
	hash := crypto.Sha256(msg)

	rInt, sInt, err := ecdsa.Sign(rand.Reader, priv, hash)
	s.Require().NoError(err)

	input := make([]byte, p256.VerifyInputLength)
	copy(input[0:32], hash)
	copy(input[32:64], rInt.Bytes())
	copy(input[64:96], sInt.Bytes())
	copy(input[96:128], priv.PublicKey.X.Bytes())
	copy(input[128:160], priv.PublicKey.Y.Bytes())

	return input
}

func parseSignature(sig []byte) (r, s []byte, err error) {
	var inner cryptobyte.String
	input := cryptobyte.String(sig)
	if !input.ReadASN1(&inner, asn1.SEQUENCE) ||
		!input.Empty() ||
		!inner.ReadASN1Integer(&r) ||
		!inner.ReadASN1Integer(&s) ||
		!inner.Empty() {
		return nil, nil, errors.New("invalid ASN.1")
	}
	return r, s, nil
}
