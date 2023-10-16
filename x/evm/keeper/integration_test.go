// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper_test

import (
	"math/big"

	"cosmossdk.io/math"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/contracts"
	"github.com/evmos/evmos/v15/precompiles/staking"

	"github.com/evmos/evmos/v15/testutil/integration/factory"
	"github.com/evmos/evmos/v15/testutil/integration/grpc"
	testkeyring "github.com/evmos/evmos/v15/testutil/integration/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/network"
	integrationutils "github.com/evmos/evmos/v15/testutil/integration/utils"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

type IntegrationTestSuite struct {
	network     network.Network
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring
}

// This test suite is meant to test the EVM module in the context of the EVMOS.
// It uses the integration test framework to spin up a local EVMOS network and
// perform transactions on it.
// The test suite focus on testing how the MsgEthereumTx message is handled under the
// different params configuration of the module while testing the different Tx types
// Ethereum supports (LegacyTx, AccessListTx, DynamicFeeTx) and the different types of
// transactions (transfer, contract deployment, contract call).
// Note that more in depth testing of the EVM and solidity execution is done through the
// hardhat and the nix setup.
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

		DescribeTable("Executes a transfer transaction", func(getTxArgs func() evmtypes.EvmTxArgs) {
			senderKey := s.keyring.GetKey(0)
			receiverKey := s.keyring.GetKey(1)
			denom := s.network.GetDenom()

			senderPrevBalanceResponse, err := s.grpcHandler.GetBalance(senderKey.AccAddr, denom)
			Expect(err).To(BeNil())
			senderPrevBalance := senderPrevBalanceResponse.GetBalance().Amount

			receiverPrevBalanceResponse, err := s.grpcHandler.GetBalance(receiverKey.AccAddr, denom)
			Expect(err).To(BeNil())
			receiverPrevBalance := receiverPrevBalanceResponse.GetBalance().Amount

			transferAmount := int64(1000)

			// Taking custom args from the table entry
			txArgs := getTxArgs()
			txArgs.Amount = big.NewInt(transferAmount)
			txArgs.To = &receiverKey.Addr

			res, err := s.factory.ExecuteEthTx(senderKey.Priv, txArgs)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())

			err = s.network.NextBlock()
			Expect(err).To(BeNil())

			// Check sender balance after transaction
			senderBalanceResultBeforeFees := senderPrevBalance.Sub(math.NewInt(transferAmount))
			senderAfterBalance, err := s.grpcHandler.GetBalance(senderKey.AccAddr, denom)
			Expect(err).To(BeNil())
			Expect(senderAfterBalance.GetBalance().Amount.LTE(senderBalanceResultBeforeFees)).To(BeTrue())

			// Check receiver balance after transaction
			receiverBalanceResult := receiverPrevBalance.Add(math.NewInt(transferAmount))
			receverAfterBalanceResponse, err := s.grpcHandler.GetBalance(receiverKey.AccAddr, denom)
			Expect(err).To(BeNil())
			Expect(receverAfterBalanceResponse.GetBalance().Amount).To(Equal(receiverBalanceResult))
		},
			Entry("as a DynamicFeeTx", func() evmtypes.EvmTxArgs { return evmtypes.EvmTxArgs{} }),
			Entry("as an AccessListTx",
				func() evmtypes.EvmTxArgs {
					return evmtypes.EvmTxArgs{
						Accesses: &ethtypes.AccessList{{
							Address:     s.keyring.GetAddr(1),
							StorageKeys: []common.Hash{{0}},
						}},
					}
				},
			),
			Entry("as a LegacyTx", func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{
					GasPrice: big.NewInt(1e9),
				}
			}),
		)

		DescribeTable("Executes a contract deployment", func(getTxArgs func() evmtypes.EvmTxArgs) {
			// Deploy contract
			senderPriv := s.keyring.GetPrivKey(0)
			constructorArgs := []interface{}{"coin", "token", uint8(18)}
			compiledContract := contracts.ERC20MinterBurnerDecimalsContract

			txArgs := getTxArgs()
			contractAddr, err := s.factory.DeployContract(
				senderPriv,
				txArgs,
				factory.ContractDeploymentData{
					Contract:        compiledContract,
					ConstructorArgs: constructorArgs,
				},
			)
			Expect(err).To(BeNil())
			Expect(contractAddr).ToNot(Equal(common.Address{}))

			err = s.network.NextBlock()
			Expect(err).To(BeNil())

			// Check contract account got created correctly
			contractBechAddr := sdktypes.AccAddress(contractAddr.Bytes()).String()
			contractAccount, err := s.grpcHandler.GetAccount(contractBechAddr)
			Expect(err).To(BeNil())
			err = integrationutils.IsContractAccount(contractAccount)
			Expect(err).To(BeNil())
		},
			Entry("as a DynamicFeeTx", func() evmtypes.EvmTxArgs { return evmtypes.EvmTxArgs{} }),
			Entry("as an AccessListTx",
				func() evmtypes.EvmTxArgs {
					return evmtypes.EvmTxArgs{
						Accesses: &ethtypes.AccessList{{
							Address:     s.keyring.GetAddr(1),
							StorageKeys: []common.Hash{{0}},
						}},
					}
				},
			),
			Entry("as a LegacyTx", func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{
					GasPrice: big.NewInt(1e9),
				}
			}),
		)

		Context("With a predeployed ERC20MinterBurnerDecimalsContract", func() {
			var contractAddr common.Address

			BeforeEach(func() {
				// Deploy contract
				senderPriv := s.keyring.GetPrivKey(0)
				constructorArgs := []interface{}{"coin", "token", uint8(18)}
				compiledContract := contracts.ERC20MinterBurnerDecimalsContract

				var err error // Avoid shadowing
				contractAddr, err = s.factory.DeployContract(
					senderPriv,
					evmtypes.EvmTxArgs{}, // Default values
					factory.ContractDeploymentData{
						Contract:        compiledContract,
						ConstructorArgs: constructorArgs,
					},
				)
				Expect(err).To(BeNil())
				Expect(contractAddr).ToNot(Equal(common.Address{}))

				err = s.network.NextBlock()
				Expect(err).To(BeNil())
			})

			DescribeTable("Executes a contract call", func(getTxArgs func() evmtypes.EvmTxArgs) {
				senderPriv := s.keyring.GetPrivKey(0)
				compiledContract := contracts.ERC20MinterBurnerDecimalsContract
				recipientKey := s.keyring.GetKey(1)

				// Execute contract call
				mintTxArgs := getTxArgs()
				mintTxArgs.To = &contractAddr

				amountToMint := big.NewInt(1e18)
				mintArgs := factory.CallArgs{
					ContractABI: compiledContract.ABI,
					MethodName:  "mint",
					Args:        []interface{}{recipientKey.Addr, amountToMint},
				}
				mintResponse, err := s.factory.ExecuteContractCall(senderPriv, mintTxArgs, mintArgs)
				Expect(err).To(BeNil())
				Expect(mintResponse.IsOK()).To(Equal(true), "transaction should have succeeded", mintResponse.GetLog())

				// Check contract call response has the expected topics for a mint
				// call within an ERC20 contract
				expectedTopics := []string{
					"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
					"0x0000000000000000000000000000000000000000000000000000000000000000",
				}
				err = integrationutils.CheckTxTopics(mintResponse, expectedTopics)
				Expect(err).To(BeNil())

				err = s.network.NextBlock()
				Expect(err).To(BeNil())

				totalSupplyTxArgs := evmtypes.EvmTxArgs{
					To: &contractAddr,
				}
				totalSupplyArgs := factory.CallArgs{
					ContractABI: compiledContract.ABI,
					MethodName:  "totalSupply",
					Args:        []interface{}{},
				}
				totalSupplyRes, err := s.factory.ExecuteContractCall(senderPriv, totalSupplyTxArgs, totalSupplyArgs)
				Expect(err).To(BeNil())
				Expect(totalSupplyRes.IsOK()).To(Equal(true), "transaction should have succeeded", totalSupplyRes.GetLog())

				var totalSupplyResponse *big.Int
				err = integrationutils.DecodeContractCallResponse(&totalSupplyResponse, totalSupplyArgs, totalSupplyRes)
				Expect(err).To(BeNil())
				Expect(totalSupplyResponse).To(Equal(amountToMint))
			},
				Entry("as a DynamicFeeTx", func() evmtypes.EvmTxArgs { return evmtypes.EvmTxArgs{} }),
				Entry("as an AccessListTx",
					func() evmtypes.EvmTxArgs {
						return evmtypes.EvmTxArgs{
							Accesses: &ethtypes.AccessList{{
								Address:     s.keyring.GetAddr(1),
								StorageKeys: []common.Hash{{0}},
							}},
						}
					},
				),
				Entry("as a LegacyTx", func() evmtypes.EvmTxArgs {
					return evmtypes.EvmTxArgs{
						GasPrice: big.NewInt(1e9),
					}
				}),
			)
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
			Expect(err.Error()).To(ContainSubstring("invalid chain id"))
			// Transaction fails before being broadcasted
			Expect(res).To(Equal(abcitypes.ResponseDeliverTx{}))
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
			senderKey := s.keyring.GetKey(0)
			receiverKey := s.keyring.GetKey(1)
			denom := s.network.GetDenom()

			senderPrevBalanceResponse, err := s.grpcHandler.GetBalance(senderKey.AccAddr, denom)
			Expect(err).To(BeNil())
			senderPrevBalance := senderPrevBalanceResponse.GetBalance().Amount

			receiverPrevBalanceResponse, err := s.grpcHandler.GetBalance(receiverKey.AccAddr, denom)
			Expect(err).To(BeNil())
			receiverPrevBalance := receiverPrevBalanceResponse.GetBalance().Amount

			transferAmount := int64(1000)

			txArgs := evmtypes.EvmTxArgs{
				To:     &receiverKey.Addr,
				Amount: big.NewInt(transferAmount),
			}

			res, err := s.factory.ExecuteEthTx(senderKey.Priv, txArgs)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())

			err = s.network.NextBlock()
			Expect(err).To(BeNil())

			// Check sender balance after transaction
			senderBalanceResultBeforeFees := senderPrevBalance.Sub(math.NewInt(transferAmount))
			senderAfterBalance, err := s.grpcHandler.GetBalance(senderKey.AccAddr, denom)
			Expect(err).To(BeNil())
			Expect(senderAfterBalance.GetBalance().Amount.LTE(senderBalanceResultBeforeFees)).To(BeTrue())

			// Check receiver balance after transaction
			receiverBalanceResult := receiverPrevBalance.Add(math.NewInt(transferAmount))
			receverAfterBalanceResponse, err := s.grpcHandler.GetBalance(receiverKey.AccAddr, denom)
			Expect(err).To(BeNil())
			Expect(receverAfterBalanceResponse.GetBalance().Amount).To(Equal(receiverBalanceResult))
		})

		It("fails when trying to perform contract deployment", func() {
			senderPriv := s.keyring.GetPrivKey(0)
			constructorArgs := []interface{}{"coin", "token", uint8(18)}
			compiledContract := contracts.ERC20MinterBurnerDecimalsContract
			contractAddr, err := s.factory.DeployContract(
				senderPriv,
				evmtypes.EvmTxArgs{}, // Default values
				factory.ContractDeploymentData{
					Contract:        compiledContract,
					ConstructorArgs: constructorArgs,
				},
			)

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("EVM Create operation is disabled"))
			Expect(contractAddr).To(Equal(common.Address{}))
		})

		It("performs a contract call to the staking precompile", func() {
			senderPriv := s.keyring.GetPrivKey(0)
			contractAddress := common.HexToAddress(staking.PrecompileAddress)
			contractABI, err := staking.LoadABI()
			Expect(err).To(BeNil())

			totalSupplyTxArgs := evmtypes.EvmTxArgs{
				To: &contractAddress,
			}

			validatorAddress := s.network.GetValidators()[0].OperatorAddress
			totalSupplyArgs := factory.CallArgs{
				ContractABI: contractABI,
				MethodName:  staking.ValidatorMethod,
				Args:        []interface{}{validatorAddress},
			}
			totalSupplyRes, err := s.factory.ExecuteContractCall(senderPriv, totalSupplyTxArgs, totalSupplyArgs)
			Expect(err).To(BeNil())
			Expect(totalSupplyRes.IsOK()).To(Equal(true), "transaction should have succeeded", totalSupplyRes.GetLog())

			var validatorResponse staking.ValidatorOutput
			err = integrationutils.DecodeContractCallResponse(&validatorResponse, totalSupplyArgs, totalSupplyRes)
			Expect(err).To(BeNil())
			Expect(validatorResponse.Validator.OperatorAddress).To(Equal(validatorAddress))
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
			Expect(err.Error()).To(ContainSubstring("EVM Call operation is disabled"))
			Expect(res.IsErr()).To(Equal(true), "transaction should have failed", res.GetLog())
		})

		It("performs a contract deployment and fails to perform a contract call", func() {
			senderPriv := s.keyring.GetPrivKey(0)
			constructorArgs := []interface{}{"coin", "token", uint8(18)}
			compiledContract := contracts.ERC20MinterBurnerDecimalsContract
			contractAddr, err := s.factory.DeployContract(
				senderPriv,
				evmtypes.EvmTxArgs{}, // Default values
				factory.ContractDeploymentData{
					Contract:        compiledContract,
					ConstructorArgs: constructorArgs,
				},
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
			Expect(err.Error()).To(ContainSubstring("EVM Call operation is disabled"))
			Expect(res.IsErr()).To(Equal(true), "transaction should have failed", res.GetLog())
		})
	})
})
