// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper_test

import (
	"math/big"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v14/contracts"

	"github.com/evmos/evmos/v14/testutil/integration/factory"
	"github.com/evmos/evmos/v14/testutil/integration/grpc"
	testkeyring "github.com/evmos/evmos/v14/testutil/integration/keyring"
	"github.com/evmos/evmos/v14/testutil/integration/network"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"
)

type IntegrationTestSuite struct {
	network     network.Network
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring
}

var _ = Describe("Handling a MsgEthereumTx message", Label("EVM"), Ordered, func() {
	var s *IntegrationTestSuite

	BeforeAll(func() {
		keyring := testkeyring.New(3)
		integrationNetwork := network.New(
			network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		)
		grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
		txFactory := factory.New(integrationNetwork, grpcHandler)
		s = &IntegrationTestSuite{
			network:     integrationNetwork,
			factory:     txFactory,
			grpcHandler: grpcHandler,
			keyring:     keyring,
		}
	})

	AfterEach(func() {
		// Start each test with a fresh block
		err := s.network.NextBlock()
		Expect(err).To(BeNil())
	})

	When("the params have default values", Ordered, func() {
		BeforeAll(func() {
			// Set params to default values
			defaultParams := evmtypes.DefaultParams()
			err := s.network.UpdateEvmParams(defaultParams)
			Expect(err).To(BeNil())

			err = s.network.NextBlock()
			Expect(err).To(BeNil())
		})

		It("performs a transfer transaction", func() {
			senderPriv := s.keyring.GetPrivKey(0)
			receiver := s.keyring.GetKey(1)
			txArgs := evmtypes.EvmTxArgs{
				To:     &receiver.Addr,
				Amount: big.NewInt(1000),
			}

			res, err := s.factory.ExecuteEthTx(senderPriv, txArgs)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
		})

		It("performs a contract deployment and contract call", func() {
			senderPriv := s.keyring.GetPrivKey(0)
			constructorArgs := []interface{}{"coin", "token", uint8(18)}
			compiledContract := contracts.ERC20MinterBurnerDecimalsContract
			contractAddr, err := s.factory.DeployContract(
				senderPriv,
				compiledContract,
				constructorArgs...,
			)
			Expect(err).To(BeNil())
			Expect(contractAddr).ToNot(Equal(common.Address{}))

			err = s.network.NextBlock()
			Expect(err).To(BeNil())

			txArgs := evmtypes.EvmTxArgs{
				To: &contractAddr,
			}
			callArgs := factory.CallArgs{
				ContractABI: compiledContract.ABI,
				MethodName:  "mint",
				Args:        []interface{}{s.keyring.GetAddr(1), big.NewInt(1e18)},
			}
			res, err := s.factory.ExecuteContractCall(senderPriv, txArgs, callArgs)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
		})

		It("should fail when ChainID is wrong", func() {
			senderPriv := s.keyring.GetPrivKey(0)
			receiver := s.keyring.GetKey(1)
			txArgs := evmtypes.EvmTxArgs{
				To:      &receiver.Addr,
				Amount:  big.NewInt(1000),
				ChainID: big.NewInt(1),
			}

			res, err := s.factory.ExecuteEthTx(senderPriv, txArgs)
			Expect(err).NotTo(BeNil())
			// Transaction fails before being broadcasted
			Expect(res).To(Equal(abcitypes.ResponseDeliverTx{}))
		})

		It("performs an AccessListTx", func() {
			senderPriv := s.keyring.GetPrivKey(0)
			receiver := s.keyring.GetKey(1)
			accessList := &ethtypes.AccessList{{Address: receiver.Addr, StorageKeys: []common.Hash{{0}}}}
			// GasFeeCap and GasTipCap are populated by default by the factory
			txArgs := evmtypes.EvmTxArgs{
				To:       &receiver.Addr,
				Amount:   big.NewInt(1000),
				Accesses: accessList,
			}

			res, err := s.factory.ExecuteEthTx(senderPriv, txArgs)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
		})

		It("performs a LegacyTx", func() {
			senderPriv := s.keyring.GetPrivKey(0)
			receiver := s.keyring.GetKey(1)
			// GasFeeCap and GasTipCap are populated by default by the factory
			txArgs := evmtypes.EvmTxArgs{
				To:       &receiver.Addr,
				Amount:   big.NewInt(1000),
				GasPrice: big.NewInt(1e9),
			}

			res, err := s.factory.ExecuteEthTx(senderPriv, txArgs)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
		})
	})

	When("EnableCreate param is set to false", Ordered, func() {
		BeforeAll(func() {
			// Set params to default values
			defaultParams := evmtypes.DefaultParams()
			defaultParams.EnableCreate = false
			err := s.network.UpdateEvmParams(defaultParams)
			Expect(err).To(BeNil())

			err = s.network.NextBlock()
			Expect(err).To(BeNil())
		})

		It("performs a transfer transaction", func() {
			senderPriv := s.keyring.GetPrivKey(0)
			receiver := s.keyring.GetKey(1)
			txArgs := evmtypes.EvmTxArgs{
				To:     &receiver.Addr,
				Amount: big.NewInt(1000),
				// Hard coded gas limit to avoid failure on gas estimation because
				// of the param
				GasLimit: 100000,
			}

			res, err := s.factory.ExecuteEthTx(senderPriv, txArgs)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
		})

		It("fails when trying to perform contract deployment", func() {
			senderPriv := s.keyring.GetPrivKey(0)
			constructorArgs := []interface{}{"coin", "token", uint8(18)}
			compiledContract := contracts.ERC20MinterBurnerDecimalsContract
			contractAddr, err := s.factory.DeployContract(
				senderPriv,
				compiledContract,
				constructorArgs...,
			)
			Expect(err).NotTo(BeNil())
			Expect(contractAddr).To(Equal(common.Address{}))
		})
	})

	When("EnableCall param is set to false", Ordered, func() {
		BeforeAll(func() {
			// Set params to default values
			defaultParams := evmtypes.DefaultParams()
			defaultParams.EnableCall = false
			err := s.network.UpdateEvmParams(defaultParams)
			Expect(err).To(BeNil())

			err = s.network.NextBlock()
			Expect(err).To(BeNil())
		})

		It("fails when performing a transfer transaction", func() {
			senderPriv := s.keyring.GetPrivKey(0)
			receiver := s.keyring.GetKey(1)
			txArgs := evmtypes.EvmTxArgs{
				To:     &receiver.Addr,
				Amount: big.NewInt(1000),
				// Hard coded gas limit to avoid failure on gas estimation because
				// of the param
				GasLimit: 100000,
			}

			res, err := s.factory.ExecuteEthTx(senderPriv, txArgs)
			Expect(err).NotTo(BeNil())
			Expect(res.IsErr()).To(Equal(true), "transaction should have failed", res.GetLog())
		})

		It("performs a contract deployment and fails to perform a contract call", func() {
			senderPriv := s.keyring.GetPrivKey(0)
			constructorArgs := []interface{}{"coin", "token", uint8(18)}
			compiledContract := contracts.ERC20MinterBurnerDecimalsContract
			contractAddr, err := s.factory.DeployContract(
				senderPriv,
				compiledContract,
				constructorArgs...,
			)
			Expect(err).To(BeNil())
			Expect(contractAddr).ToNot(Equal(common.Address{}))

			txArgs := evmtypes.EvmTxArgs{
				To: &contractAddr,
				// Hard coded gas limit to avoid failure on gas estimation because
				// of the param
				GasLimit: 100000,
			}
			callArgs := factory.CallArgs{
				ContractABI: compiledContract.ABI,
				MethodName:  "mint",
				Args:        []interface{}{s.keyring.GetAddr(1), big.NewInt(1e18)},
			}
			res, err := s.factory.ExecuteContractCall(senderPriv, txArgs, callArgs)
			Expect(err).NotTo(BeNil())
			Expect(res.IsErr()).To(Equal(true), "transaction should have failed", res.GetLog())
		})
	})
})
