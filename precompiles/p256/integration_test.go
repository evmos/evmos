// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package p256_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/cometbft/cometbft/crypto"
	"github.com/ethereum/go-ethereum/common"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	"github.com/evmos/evmos/v19/precompiles/p256"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/utils"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

type IntegrationTestSuite struct {
	network           network.Network
	factory           factory.TxFactory
	keyring           testkeyring.Keyring
	precompileAddress common.Address
	p256Priv          *ecdsa.PrivateKey
}

var _ = Describe("Calling p256 precompile directly", Label("P256 Precompile"), Ordered, func() {
	var s *IntegrationTestSuite

	BeforeAll(func() {
		keyring := testkeyring.New(1)
		integrationNetwork := network.New(
			network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		)
		grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
		txFactory := factory.New(integrationNetwork, grpcHandler)
		p256Priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		Expect(err).To(BeNil())

		s = &IntegrationTestSuite{
			network:           integrationNetwork,
			factory:           txFactory,
			keyring:           keyring,
			precompileAddress: p256.Precompile{}.Address(),
			p256Priv:          p256Priv,
		}
	})

	AfterEach(func() {
		// Start each test with a fresh block
		err := s.network.NextBlock()
		Expect(err).To(BeNil())
	})

	When("the precompile is enabled in the EVM params", func() {
		BeforeAll(func() {
			s = setupIntegrationTestSuite(nil)
		})
		DescribeTable("execute contract call", func(inputFn func() (input, expOutput []byte, expErr string)) {
			senderKey := s.keyring.GetKey(0)

			input, expOutput, expErr := inputFn()
			args := evmtypes.EvmTxArgs{
				To:    &s.precompileAddress,
				Input: input,
			}

			txResult, err := s.factory.ExecuteEthTx(senderKey.Priv, args)
			Expect(err).To(BeNil())
			Expect(txResult.IsOK()).To(Equal(true), "transaction should have succeeded", txResult.GetLog())

			res, err := utils.DecodeExecTxResult(txResult)
			Expect(err).To(BeNil())
			Expect(res.VmError).To(Equal(expErr), "expected different vm error")
			Expect(res.Ret).To(Equal(expOutput))
		},
			Entry(
				"valid signature",
				func() (input, expOutput []byte, expErr string) {
					input = signMsg([]byte("hello world"), s.p256Priv)
					return input, trueValue, ""
				},
			),
			Entry(
				"invalid signature",
				func() (input, expOutput []byte, expErr string) {
					privB, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
					Expect(err).To(BeNil())

					hash := crypto.Sha256([]byte("hello world"))

					rInt, sInt, err := ecdsa.Sign(rand.Reader, s.p256Priv, hash)
					Expect(err).To(BeNil())

					input = make([]byte, p256.VerifyInputLength)
					copy(input[0:32], hash)
					copy(input[32:64], rInt.Bytes())
					copy(input[64:96], sInt.Bytes())
					copy(input[96:128], privB.PublicKey.X.Bytes())
					copy(input[128:160], privB.PublicKey.Y.Bytes())
					return input, nil, ""
				},
			),
		)
	})

	When("the precompile is not enabled in the EVM params", func() {
		BeforeAll(func() {
			customGenesis := evmtypes.DefaultGenesisState()
			params := customGenesis.Params
			addr := s.precompileAddress.String()
			var activePrecompiles []string
			for _, precompile := range params.ActiveStaticPrecompiles {
				if precompile != addr {
					activePrecompiles = append(activePrecompiles, precompile)
				}
			}
			params.ActiveStaticPrecompiles = activePrecompiles
			customGenesis.Params = params
			s = setupIntegrationTestSuite(customGenesis)
		})

		DescribeTable("execute contract call", func(inputFn func() (input []byte)) {
			senderKey := s.keyring.GetKey(0)

			input := inputFn()
			args := evmtypes.EvmTxArgs{
				To:    &s.precompileAddress,
				Input: input,
			}

			_, err := s.factory.ExecuteEthTx(senderKey.Priv, args)
			Expect(err).To(BeNil(), "expected no error since contract doesn't exists")
		},
			Entry(
				"valid signature",
				func() (input []byte) {
					input = signMsg([]byte("hello world"), s.p256Priv)
					return input
				},
			),
			Entry(
				"invalid signature",
				func() (input []byte) {
					privB, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
					Expect(err).To(BeNil())

					hash := crypto.Sha256([]byte("hello world"))

					rInt, sInt, err := ecdsa.Sign(rand.Reader, s.p256Priv, hash)
					Expect(err).To(BeNil())

					input = make([]byte, p256.VerifyInputLength)
					copy(input[0:32], hash)
					copy(input[32:64], rInt.Bytes())
					copy(input[64:96], sInt.Bytes())
					copy(input[96:128], privB.PublicKey.X.Bytes())
					copy(input[128:160], privB.PublicKey.Y.Bytes())
					return input
				},
			),
		)
	})
})

// setupIntegrationTestSuite is a helper function to setup a integration test suite
// with a network with a specified custom genesis state for the EVM module
func setupIntegrationTestSuite(customEVMGenesis *evmtypes.GenesisState) *IntegrationTestSuite {
	customGenesis := network.CustomGenesisState{}
	if customEVMGenesis != nil {
		customGenesis[evmtypes.ModuleName] = customEVMGenesis
	}
	keyring := testkeyring.New(1)
	integrationNetwork := network.New(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		network.WithCustomGenesis(customGenesis),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)
	p256Priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	Expect(err).To(BeNil())

	suite := &IntegrationTestSuite{
		network:           integrationNetwork,
		factory:           txFactory,
		keyring:           keyring,
		precompileAddress: p256.Precompile{}.Address(),
		p256Priv:          p256Priv,
	}

	return suite
}
