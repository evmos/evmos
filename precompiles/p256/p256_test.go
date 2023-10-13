package p256_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/cometbft/cometbft/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/precompiles/p256"
)

func (suite *PrecompileTestSuite) TestRun() {
	testCases := []struct {
		name     string
		sign     func() []byte
		expError bool
		expPass  bool
	}{
		{
			"pass - Sign",
			func() []byte {
				msg := []byte("hello world")
				hash := crypto.Sha256(msg)

				r, s, err := ecdsa.Sign(rand.Reader, suite.p256Priv, hash)
				suite.Require().NoError(err)

				input := make([]byte, p256.VerifyInputLength)
				copy(input[0:32], hash)
				copy(input[32:64], r.Bytes())
				copy(input[64:96], s.Bytes())
				copy(input[96:128], suite.p256Priv.PublicKey.X.Bytes())
				copy(input[128:160], suite.p256Priv.PublicKey.Y.Bytes())

				return input
			},
			false,
			true,
		},
		{
			"pass - sign ASN.1 encoded signature",
			func() []byte {
				msg := []byte("hello world")
				hash := crypto.Sha256(msg)

				sig, err := ecdsa.SignASN1(rand.Reader, s.p256Priv, hash)
				suite.Require().NoError(err)

				r, s, err := parseSignature(sig)
				suite.Require().NoError(err)

				input := make([]byte, p256.VerifyInputLength)
				copy(input[0:32], hash)
				copy(input[32:64], r)
				copy(input[64:96], s)
				copy(input[96:128], suite.p256Priv.PublicKey.X.Bytes())
				copy(input[128:160], suite.p256Priv.PublicKey.Y.Bytes())

				return input
			},
			false,
			false,
		},
		{
			"fail - invalid signature",
			func() []byte {
				privB, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				s.Require().NoError(err)

				bz := elliptic.MarshalCompressed(elliptic.P256(), suite.p256Priv.X, suite.p256Priv.Y)
				s.Require().NotEmpty(bz)

				msg := []byte("hello world")
				hash := crypto.Sha256(msg)

				r, s, err := ecdsa.Sign(rand.Reader, suite.p256Priv, hash)
				suite.Require().NoError(err)

				input := make([]byte, p256.VerifyInputLength)
				copy(input[0:32], hash)
				copy(input[32:64], r.Bytes())
				copy(input[64:96], s.Bytes())
				copy(input[96:128], privB.PublicKey.X.Bytes())
				copy(input[128:160], privB.PublicKey.Y.Bytes())

				return input
			},
			true,
			false,
		},
		{
			"fail - invalid length",
			func() []byte {
				msg := []byte("hello world")
				hash := crypto.Sha256(msg)

				input := make([]byte, 32)
				copy(input[0:32], hash)

				return input
			},
			true,
			false,
		},
	}

	trueValue := common.LeftPadBytes(common.Big1.Bytes(), 32)

	for _, tc := range testCases {
		input := tc.sign()
		bz, err := suite.precompile.Run(nil, &vm.Contract{Input: input}, false)
		if !tc.expError {
			suite.Require().NoError(err)
			if tc.expPass {
				suite.Require().Equal(trueValue, bz, tc.name)
			}
		} else {
			suite.Require().NoError(err)
			suite.Require().Empty(bz, tc.name)
		}
	}
}
