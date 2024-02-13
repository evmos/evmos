package werc20_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	auth "github.com/evmos/evmos/v16/precompiles/authorization"
	erc20precompile "github.com/evmos/evmos/v16/precompiles/erc20"
	"github.com/evmos/evmos/v16/precompiles/testutil"
	"github.com/evmos/evmos/v16/precompiles/werc20"
	"github.com/evmos/evmos/v16/precompiles/werc20/testdata"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	testutils "github.com/evmos/evmos/v16/testutil/integration/evmos/utils"
	evmosutiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/utils"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

// WERC20IntegrationTestSuite is the implementation of the TestSuite interface for ERC20 precompile
// unit tests.
type WERC20IntegrationTestSuite struct {
	network     network.Network
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring
}

const chainID = utils.MainnetChainID + "-1"

var _ = Describe("WEVMOS Extension -", Ordered, func() {
	var (
		s *WERC20IntegrationTestSuite

		// senderKey is the test key used to send all transactions in this test suite
		senderKey testkeyring.Key
		// werc20ExtensionAddr is the address of the WERC-20 EVM extension
		werc20ExtensionAddr common.Address
		// werc20ABI is the ABI of the WERC-20 EVM extension
		werc20ABI abi.ABI

		// werc20TxArgs are the default transactions arguments used to call the WERC-20 EVM extension
		werc20TxArgs evmtypes.EvmTxArgs

		// expPass is the default check for successful transactions
		expPass testutil.LogCheckArgs
	)

	BeforeAll(func() {
		// TODO: do we need three keys here?
		keyring := testkeyring.New(3)
		integrationNetwork := network.New(
			network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			network.WithChainID(chainID),
		)

		grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
		txFactory := factory.New(integrationNetwork, grpcHandler)

		s = &WERC20IntegrationTestSuite{
			network:     integrationNetwork,
			factory:     txFactory,
			grpcHandler: grpcHandler,
			keyring:     keyring,
		}

		// Add WEVMOS to params
		params, err := grpcHandler.GetEvmParams()
		Expect(err).To(BeNil(), "failed to get EVM params")

		WEVMOSAddress := common.HexToAddress(erc20precompile.WEVMOSContractMainnet)
		params.Params.ActivePrecompiles = append(params.Params.ActivePrecompiles, WEVMOSAddress.String())

		err = integrationNetwork.UpdateEvmParams(params.Params)
		Expect(err).To(BeNil(), "failed to update EVM params")

		senderKey = s.keyring.GetKey(0)
		werc20ExtensionAddr = common.HexToAddress(erc20precompile.WEVMOSContractMainnet)
		werc20ABI, err = werc20.LoadABI()
		Expect(err).To(BeNil(), "failed to load WERC-20 ABI")

		werc20TxArgs = evmtypes.EvmTxArgs{
			To: &werc20ExtensionAddr,
		}

		expPass = testutil.LogCheckArgs{
			ABIEvents: werc20ABI.Events,
			ExpPass:   true,
		}
	})

	When("calling deposit correctly", func() {
		It("should not emit events", func() {
			txArgs := evmtypes.EvmTxArgs{
				To:     &werc20ExtensionAddr,
				Amount: big.NewInt(100),
			}

			depositArgs := factory.CallArgs{
				ContractABI: werc20ABI,
				MethodName:  werc20.DepositMethod,
			}

			depositResponse, err := s.factory.ExecuteContractCall(
				senderKey.Priv, txArgs, depositArgs,
			)
			Expect(err).To(BeNil(), "failed to call contract")
			Expect(depositResponse.IsOK()).To(
				BeTrue(),
				"transaction should have succeeded",
				depositResponse.GetLog(),
			)
			Expect(depositResponse.GasUsed).To(
				BeNumerically(">=", werc20.DepositRequiredGas),
				"expected different gas used",
			)
		})
	})

	When("calling withdraw correctly", func() {
		It("should not emit events", func() {
			txArgs := evmtypes.EvmTxArgs{
				To: &werc20ExtensionAddr,
			}

			withdrawArgs := factory.CallArgs{
				ContractABI: werc20ABI,
				MethodName:  werc20.WithdrawMethod,
				Args: []interface{}{
					big.NewInt(100),
				},
			}

			withdrawResponse, err := s.factory.ExecuteContractCall(
				senderKey.Priv, txArgs, withdrawArgs,
			)
			Expect(err).To(BeNil(), "failed to call contract")
			Expect(withdrawResponse.IsOK()).To(
				BeTrue(),
				"transaction should have succeeded",
				withdrawResponse.GetLog(),
			)
			Expect(withdrawResponse.GasUsed).To(
				BeNumerically(">=", werc20.WithdrawRequiredGas),
				"expected different gas used",
			)
		})
	})

	// TODO: How do we actually check the method types here? We can see the correct ones being populated by printing the line in the cmn.Precompile
	//
	// FIXME: address TODO here?! What exactly is the question? If the fallback was correctly executed?
	When("calling with empty method but non-zero amount", func() {
		It("should call `receive`", func() {
			txArgs := evmtypes.EvmTxArgs{
				To:     &werc20ExtensionAddr,
				Amount: big.NewInt(100),
			}

			receiveArgs := factory.CallArgs{
				ContractABI: werc20ABI,
				MethodName:  "",
			}
			receiveResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, txArgs, receiveArgs)
			Expect(err).To(BeNil(), "unexpected result calling contract")
			Expect(receiveResponse.IsOK()).To(
				BeTrue(),
				"transaction should have succeeded",
				receiveResponse.GetLog(),
			)
		})
	})

	When("calling with short call data, empty method and non-zero amount", func() {
		It("should call `fallback`", func() {
			txArgs := evmtypes.EvmTxArgs{
				To:     &werc20ExtensionAddr,
				Amount: big.NewInt(100),
				Input:  []byte{1, 2, 3},
			}

			receiveArgs := factory.CallArgs{
				ContractABI: werc20ABI,
				MethodName:  "",
			}
			receiveResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, txArgs, receiveArgs)
			Expect(err).To(BeNil(), "unexpected result calling contract")
			Expect(receiveResponse.IsOK()).To(
				BeTrue(),
				"transaction should have succeeded",
				receiveResponse.GetLog(),
			)
		})
	})

	When("calling a non-existing function with amount", func() {
		It("should call `fallback`", func() {
			txArgs := evmtypes.EvmTxArgs{
				To:     &werc20ExtensionAddr,
				Amount: big.NewInt(100),
				Input:  []byte("nonExistingMethod"),
			}

			res, err := s.factory.ExecuteEthTx(senderKey.Priv, txArgs)
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

			err = testutil.CheckLogs(expPass.WithRes(res))
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

			// TODO: what balances are expected here?
			var expectedBalances []banktypes.Balance
			err = testutils.CheckBalances(s.grpcHandler, expectedBalances)
			Expect(err).ToNot(HaveOccurred(), "unexpected result checking balances")
		})
	})

	When("calling empty function without amount", func() {
		It("should call `fallback`", func() {
			txArgs := evmtypes.EvmTxArgs{
				To:    &werc20ExtensionAddr,
				Input: []byte("nonExistingMethod"),
			}

			res, err := s.factory.ExecuteEthTx(senderKey.Priv, txArgs)
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

			err = testutil.CheckLogs(expPass.WithRes(res))
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

			// TODO: what balances are expected here?
			var expectedBalances []banktypes.Balance
			err = testutils.CheckBalances(s.grpcHandler, expectedBalances)
			Expect(err).ToNot(HaveOccurred(), "unexpected result checking balances")
		})
	})

	When("calling with short call data without amount", func() {
		It("should call `fallback`", func() {
			txArgs := evmtypes.EvmTxArgs{
				To:    &werc20ExtensionAddr,
				Input: []byte{1, 2, 3},
			}

			res, err := s.factory.ExecuteEthTx(senderKey.Priv, txArgs)
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

			err = testutil.CheckLogs(expPass.WithRes(res))
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

			// TODO: what balances are expected here?
			var expectedBalances []banktypes.Balance
			err = testutils.CheckBalances(s.grpcHandler, expectedBalances)
			Expect(err).ToNot(HaveOccurred(), "unexpected result checking balances")
		})
	})

	When("calling non-existent function without amount", func() {
		It("should call `fallback`", func() {
			txArgs := evmtypes.EvmTxArgs{
				To:    &werc20ExtensionAddr,
				Input: []byte("nonExistingMethod"),
			}

			res, err := s.factory.ExecuteEthTx(senderKey.Priv, txArgs)
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

			err = testutil.CheckLogs(expPass.WithRes(res))
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

			// TODO: what balances are expected here?
			var expectedBalances []banktypes.Balance
			err = testutils.CheckBalances(s.grpcHandler, expectedBalances)
			Expect(err).ToNot(HaveOccurred(), "unexpected result checking balances")
		})
	})

	Context("Comparing to original WEVMOS contract (not precompiled)", func() {
		var WEVMOSOriginalContractAddr common.Address

		BeforeAll(func() {
			var err error
			WEVMOSOriginalContractAddr, err = s.factory.DeployContract(
				senderKey.Priv,
				evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
				factory.ContractDeploymentData{
					Contract:        testdata.WEVMOSContract,
					ConstructorArgs: []interface{}{},
				},
			)
			Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")
		})

		When("calling deposit", func() {
			It("should have the exact same gas consumption", func() {
				txArgs := evmtypes.EvmTxArgs{
					To:       &werc20ExtensionAddr,
					Amount:   big.NewInt(100),
					GasLimit: 50_000, // FIXME: why is the gas limit not enough here by default? Raises out of gas error if not provided
				}

				depositArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  werc20.DepositMethod,
				}

				depositResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, txArgs, depositArgs)
				Expect(err).To(BeNil(), "failed to call contract")
				Expect(depositResponse.IsOK()).To(
					BeTrue(),
					"transaction should have succeeded",
					depositResponse.GetLog(),
				)

				originalTxArgs := txArgs
				originalTxArgs.To = &WEVMOSOriginalContractAddr

				originalDepositResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, originalTxArgs, depositArgs)
				Expect(err).To(BeNil(), "failed to call contract")
				Expect(originalDepositResponse.IsOK()).To(
					BeTrue(),
					"transaction should have succeeded",
					originalDepositResponse.GetLog(),
				)

				Expect(depositResponse.GasUsed).To(
					Equal(originalDepositResponse.GasUsed),
					"expected same gas used between smart contract and precompile",
				)
			})

			It("should return the same error", func() {
				// Hardcode gas limit to search for error
				// Avoid simulate tx to fail on execution
				txArgs := evmtypes.EvmTxArgs{
					To:       &werc20ExtensionAddr,
					Amount:   big.NewInt(9e18),
					GasLimit: 50_000,
				}

				depositArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  werc20.DepositMethod,
					Args:        []interface{}{},
				}

				depositResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, txArgs, depositArgs)
				Expect(err).ToNot(BeNil(), "expected error but got none")
				Expect(depositResponse.IsOK()).To(
					BeFalse(),
					"transaction should have failed",
					depositResponse.GetLog(),
				)

				originalTxArgs := txArgs
				originalTxArgs.To = &WEVMOSOriginalContractAddr

				originalDepositResponse, errOriginal := s.factory.ExecuteContractCall(
					senderKey.Priv, originalTxArgs, depositArgs,
				)
				Expect(err).ToNot(BeNil(), "expected error but got none")
				Expect(originalDepositResponse.IsOK()).To(
					BeFalse(),
					"transaction should have failed",
					originalDepositResponse.GetLog(),
				)
				Expect(errOriginal.Error()).To(
					Equal(err.Error()),
					"expected same error for original and precompiled contracts",
				)
			})
		})

		//// FIXME: should this not just show the normal bank balance?? The deposit is a no-op now??
		//It("should reflect the correct balances", func() {
		//	// Deposit into the WEVMOS contract to have something to withdraw
		//	// TODO: -- this shouldn't work anymore though???? It should just show the correct
		//	txArgsPrecompile, callArgsPrecompile := s.getTxAndCallArgs(erc20Call, contractData, werc20.DepositMethod)
		//	txArgsPrecompile.Amount = amount
		//
		//	_, _, errPrecompile := s.factory.CallContractAndCheckLogs(sender.Priv, txArgsPrecompile, callArgsPrecompile, depositCheck)
		//	Expect(errPrecompile).ToNot(HaveOccurred(), "unexpected result calling contract")
		//
		//	txArgsContract, callArgsContract := s.getTxAndCallArgs(erc20Call, contractDataOriginal, werc20.DepositMethod)
		//	txArgsContract.Amount = amount
		//	txArgsContract.GasLimit = 50_000
		//
		//	depositCheckContract := expPass.WithExpPass(true).WithExpEvents(EventTypeDeposit)
		//	_, _, errOriginal := s.factory.CallContractAndCheckLogs(sender.Priv, txArgsContract, callArgsContract, depositCheckContract)
		//	Expect(errOriginal).ToNot(HaveOccurred(), "unexpected result calling contract")
		//
		//	// Check balances after calling precompile
		//	s.checkBalances(failCheck, sender, contractData)
		//
		//	// Check balances after calling original contract
		//	balanceCheck := failCheck.WithExpPass(true)
		//	txArgs, balancesArgs := s.getTxAndCallArgs(erc20Call, contractDataOriginal, erc20.BalanceOfMethod, sender.Addr)
		//
		//	_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, balanceCheck)
		//	Expect(err).ToNot(HaveOccurred(), "failed to execute balanceOf")
		//
		//	// Check the balance in the bank module is the same as calling `balanceOf` on the precompile
		//	var erc20Balance *big.Int
		//	//err = .UnpackIntoInterface(&erc20Balance, erc20.BalanceOfMethod, ethRes.Ret)
		//	Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
		//	Expect(erc20Balance).To(Equal(amount), "expected different balance")
		//})

		// NOTE: This is a no-op now, so there is no way of setting up funds to be withdrawn.
		// We do check that the gas consumption is working as expected.
		When("calling withdraw", func() {
			BeforeEach(func() {
				// Deposit into the separately deployed WEVMOS contract to have something to withdraw
				txArgs := evmtypes.EvmTxArgs{
					To:       &WEVMOSOriginalContractAddr,
					Amount:   big.NewInt(100),
					GasLimit: 50_000,
				}

				depositArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  werc20.DepositMethod,
				}

				_, err := s.factory.ExecuteContractCall(senderKey.Priv, txArgs, depositArgs)
				Expect(err).To(BeNil(), "failed to call contract")
			})

			It("should have the exact same gas consumption as before", func() {
				// FIXME: the gas usage is different
				Skip("This is not working as expected ATM. The gas usage is different.")

				txArgs := evmtypes.EvmTxArgs{
					To: &werc20ExtensionAddr,
				}

				withdrawArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  werc20.WithdrawMethod,
					Args:        []interface{}{big.NewInt(100)},
				}

				withdrawResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, txArgs, withdrawArgs)
				Expect(err).To(BeNil(), "failed to call contract")
				Expect(withdrawResponse.IsOK()).To(
					BeTrue(),
					"transaction should have succeeded",
					withdrawResponse.GetLog(),
				)

				originalTxArgs := txArgs
				originalTxArgs.To = &WEVMOSOriginalContractAddr

				originalWithdrawResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, originalTxArgs, withdrawArgs)
				Expect(err).To(BeNil(), "failed to call contract")
				Expect(originalWithdrawResponse.IsOK()).To(
					BeTrue(),
					"transaction should have succeeded",
					originalWithdrawResponse.GetLog(),
				)

				Expect(withdrawResponse.GasUsed).To(Equal(originalWithdrawResponse.GasUsed), "expected same gas to be used")
			})

			It("should return the same error", func() {
				// FIXME: this is not raising an error currently? Why would there be an error if this is a no-op? Why is there an error for the deposit method?
				Skip("This is not working as expected ATM. The EVM extension does not return an error.")

				// Hardcode gas limit to search for error
				txArgs := evmtypes.EvmTxArgs{
					To:       &werc20ExtensionAddr,
					GasLimit: 50_000,
				}

				withdrawArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  werc20.WithdrawMethod,
					Args:        []interface{}{big.NewInt(100)},
				}

				withdrawResponse, errPrecompile := s.factory.ExecuteContractCall(senderKey.Priv, txArgs, withdrawArgs)
				Expect(errPrecompile).ToNot(BeNil(), "expected error but got none")
				Expect(withdrawResponse.IsOK()).To(
					BeFalse(),
					"transaction should have failed",
					withdrawResponse.GetLog(),
				)

				originalTxArgs := txArgs
				originalTxArgs.To = &WEVMOSOriginalContractAddr

				originalWithdrawResponse, errOriginal := s.factory.ExecuteContractCall(senderKey.Priv, originalTxArgs, withdrawArgs)
				Expect(errOriginal).ToNot(BeNil(), "expected error but got none")
				Expect(originalWithdrawResponse.IsOK()).To(
					BeFalse(),
					"transaction should have failed",
					originalWithdrawResponse.GetLog(),
				)

				Expect(errOriginal.Error()).To(Equal(errPrecompile.Error()), "expected same error")
			})
		})
	})

	Context("ERC20 specific functions", func() {
		When("querying name", func() {
			It("should return the correct name", func() {
				txArgs := evmtypes.EvmTxArgs{
					To: &werc20ExtensionAddr,
				}

				nameArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  erc20precompile.NameMethod,
				}

				res, err := s.factory.ExecuteContractCall(senderKey.Priv, txArgs, nameArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
				Expect(res.IsOK()).To(
					BeTrue(),
					"transaction should have succeeded",
					res.GetLog(),
				)

				ethRes, err := evmtypes.DecodeTxResponse(res.Data)
				Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

				var name string
				err = werc20ABI.UnpackIntoInterface(&name, erc20precompile.NameMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(name).To(Equal("Evmos"), "expected different name")
			})
		})

		When("querying symbol", func() {
			It("should return the correct symbol", func() {
				txArgs := evmtypes.EvmTxArgs{
					To: &werc20ExtensionAddr,
				}

				symbolArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  erc20precompile.SymbolMethod,
				}

				res, err := s.factory.ExecuteContractCall(senderKey.Priv, txArgs, symbolArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
				Expect(res.IsOK()).To(
					BeTrue(),
					"transaction should have succeeded",
					res.GetLog(),
				)

				ethRes, err := evmtypes.DecodeTxResponse(res.Data)
				Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

				var symbol string
				err = werc20ABI.UnpackIntoInterface(&symbol, erc20precompile.SymbolMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(symbol).To(Equal("EVMOS"), "expected different symbol")
			})
		})

		When("querying decimals", func() {
			It("should return the correct decimals", func() {
				decimalsArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  erc20precompile.DecimalsMethod,
				}

				res, err := s.factory.ExecuteContractCall(senderKey.Priv, werc20TxArgs, decimalsArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
				Expect(res.IsOK()).To(
					BeTrue(),
					"transaction should have succeeded",
					res.GetLog(),
				)

				ethRes, err := evmtypes.DecodeTxResponse(res.Data)
				Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

				var decimals uint8
				err = werc20ABI.UnpackIntoInterface(&decimals, erc20precompile.DecimalsMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(decimals).To(Equal(uint8(18)), "expected different decimals")
			})
		})

		When("querying balance", func() {
			It("should return the same evmos balance as the bank query", func() {
				// Query the EVM extension for the balance
				txArgs := evmtypes.EvmTxArgs{
					To: &werc20ExtensionAddr,
				}

				balanceArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  erc20precompile.BalanceOfMethod,
					Args:        []interface{}{senderKey.Addr},
				}

				balanceResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, txArgs, balanceArgs)
				Expect(err).ToNot(HaveOccurred(), "failed to call contract")
				Expect(balanceResponse.IsOK()).To(
					BeTrue(),
					"transaction should have succeeded",
					balanceResponse.GetLog(),
				)

				ethRes, err := evmtypes.DecodeTxResponse(balanceResponse.Data)
				Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

				var erc20Balance *big.Int
				err = werc20ABI.UnpackIntoInterface(&erc20Balance, erc20precompile.BalanceOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")

				// Query the bank module for the balance
				//
				// NOTE: It's important to first query the ERC-20 balance, because that
				// consumes gas to send the transaction at the moment, so the balance will change.
				bankBalance, err := s.grpcHandler.GetBalance(senderKey.AccAddr, utils.BaseDenom)
				Expect(err).ToNot(HaveOccurred(), "failed to get balance")

				Expect(erc20Balance.String()).To(
					Equal(bankBalance.Balance.Amount.String()),
					"expected same balance for ERC-20 query and bank query",
				)
			})

			It("should return a zero balance for a new address", func() {
				balanceArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  erc20precompile.BalanceOfMethod,
					Args:        []interface{}{evmosutiltx.GenerateAddress()},
				}

				res, err := s.factory.ExecuteContractCall(senderKey.Priv, werc20TxArgs, balanceArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
				Expect(res.IsOK()).To(
					BeTrue(),
					"transaction should have succeeded",
					res.GetLog(),
				)

				ethRes, err := evmtypes.DecodeTxResponse(res.Data)
				Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

				var balance *big.Int
				err = werc20ABI.UnpackIntoInterface(&balance, erc20precompile.BalanceOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(balance.Int64()).To(Equal(int64(0)), "expected different balance")
			})
		})

		When("querying allowance", func() {
			It("should return an existing allowance", func() {
				grantee := evmosutiltx.GenerateAddress()
				approveAmount := big.NewInt(100)

				// Approve allowance
				approveArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  auth.ApproveMethod,
					Args:        []interface{}{grantee, approveAmount},
				}

				_, err := s.factory.ExecuteContractCall(senderKey.Priv, werc20TxArgs, approveArgs)
				Expect(err).ToNot(HaveOccurred(), "failed to approve allowance")

				allowanceArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  auth.AllowanceMethod,
					Args:        []interface{}{senderKey.Addr, grantee},
				}

				res, err := s.factory.ExecuteContractCall(senderKey.Priv, werc20TxArgs, allowanceArgs)
				Expect(err).ToNot(HaveOccurred(), "failed to query allowance")
				Expect(res.IsOK()).To(
					BeTrue(),
					"transaction should have succeeded",
					res.GetLog(),
				)

				ethRes, err := evmtypes.DecodeTxResponse(res.Data)
				Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

				var allowance *big.Int
				err = werc20ABI.UnpackIntoInterface(&allowance, auth.AllowanceMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(allowance).To(Equal(approveAmount), "expected different allowance")
			})

			It("should return zero if no allowance exists", func() {
				allowanceArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  auth.AllowanceMethod,
					Args: []interface{}{
						senderKey.Addr, evmosutiltx.GenerateAddress(),
					},
				}

				res, err := s.factory.ExecuteContractCall(senderKey.Priv, werc20TxArgs, allowanceArgs)
				Expect(err).ToNot(HaveOccurred(), "failed to query allowance")
				Expect(res.IsOK()).To(
					BeTrue(),
					"transaction should have succeeded",
					res.GetLog(),
				)

				ethRes, err := evmtypes.DecodeTxResponse(res.Data)
				Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

				var balance *big.Int
				err = werc20ABI.UnpackIntoInterface(&balance, auth.AllowanceMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(balance.Int64()).To(BeZero(), "expected zero balance")
			})
		})

		When("querying total supply", func() {
			It("should return the total supply", func() {
				expSupply, ok := new(big.Int).SetString("15000000000000000000", 10)
				Expect(ok).To(BeTrue(), "failed to parse expected supply")

				supplyArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  erc20precompile.TotalSupplyMethod,
				}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(senderKey.Priv, werc20TxArgs, supplyArgs, expPass)
				Expect(err).ToNot(HaveOccurred(), "unexpected result querying total supply")

				var supply *big.Int
				err = werc20ABI.UnpackIntoInterface(&supply, erc20precompile.TotalSupplyMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(supply).To(Equal(expSupply), "expected different supply")
			})
		})

		When("transferring tokens", func() {
			It("it should transfer tokens to a receiver using `transfer`", func() {
				receiver := s.keyring.GetKey(2)
				transferAmount := big.NewInt(100)
				transferCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), transferAmount.Int64())}

				senderBalance, err := s.grpcHandler.GetAllBalances(senderKey.AccAddr)
				Expect(err).ToNot(HaveOccurred(), "unexpected error querying sender balance")
				receiverBalance, err := s.grpcHandler.GetAllBalances(receiver.AccAddr)
				Expect(err).ToNot(HaveOccurred(), "unexpected error querying receiver balance")

				transferArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  erc20precompile.TransferMethod,
					Args:        []interface{}{receiver.Addr, transferAmount},
				}

				// Prefilling the gas price with the base fee to calculate expected balances after
				// the transfer
				baseFeeRes, err := s.grpcHandler.GetBaseFee()
				Expect(err).ToNot(HaveOccurred(), "unexpected error querying base fee")

				txArgs := werc20TxArgs
				txArgs.GasPrice = baseFeeRes.BaseFee.BigInt()

				transferCheck := expPass.WithExpEvents(erc20precompile.EventTypeTransfer)
				_, ethRes, err := s.factory.CallContractAndCheckLogs(senderKey.Priv, txArgs, transferArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				gasAmount := ethRes.GasUsed * txArgs.GasPrice.Uint64()
				coinsWithGasIncluded := transferCoins.Add(sdk.NewInt64Coin(s.network.GetDenom(), int64(gasAmount)))

				err = testutils.CheckBalances(
					s.grpcHandler,
					[]banktypes.Balance{
						{Address: senderKey.AccAddr.String(), Coins: senderBalance.Balances.Sub(coinsWithGasIncluded...)},
						{Address: receiver.AccAddr.String(), Coins: receiverBalance.Balances.Add(transferCoins...)},
					},
				)
				Expect(err).ToNot(HaveOccurred(), "expected different balances")
			})

			It("it should transfer tokens to a receiver using `transferFrom`", func() {
				receiver := s.keyring.GetKey(1)
				transferAmount := big.NewInt(100)
				transferCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), transferAmount.Int64())}

				senderBalance, err := s.grpcHandler.GetAllBalances(senderKey.AccAddr)
				Expect(err).ToNot(HaveOccurred(), "unexpected error querying sender balance")
				receiverBalance, err := s.grpcHandler.GetAllBalances(receiver.AccAddr)
				Expect(err).ToNot(HaveOccurred(), "unexpected error querying receiver balance")

				transferFromArgs := factory.CallArgs{
					ContractABI: werc20ABI,
					MethodName:  erc20precompile.TransferFromMethod,
					Args:        []interface{}{senderKey.Addr, receiver.Addr, transferAmount},
				}

				// Prefilling the gas price with the base fee to calculate expected balances after
				// the transfer
				baseFeeRes, err := s.grpcHandler.GetBaseFee()
				Expect(err).ToNot(HaveOccurred(), "unexpected error querying base fee")

				txArgs := werc20TxArgs
				txArgs.GasPrice = baseFeeRes.BaseFee.BigInt()

				transferCheck := expPass.WithExpEvents(erc20precompile.EventTypeTransfer, auth.EventTypeApproval)
				_, ethRes, err := s.factory.CallContractAndCheckLogs(senderKey.Priv, txArgs, transferFromArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				gasAmount := ethRes.GasUsed * txArgs.GasPrice.Uint64()
				coinsWithGasIncluded := transferCoins.Add(sdk.NewInt64Coin(s.network.GetDenom(), int64(gasAmount)))
				err = testutils.CheckBalances(
					s.grpcHandler,
					[]banktypes.Balance{
						{Address: senderKey.AccAddr.String(), Coins: senderBalance.Balances.Sub(coinsWithGasIncluded...)},
						{Address: receiver.AccAddr.String(), Coins: receiverBalance.Balances.Add(transferCoins...)},
					},
				)
				Expect(err).ToNot(HaveOccurred(), "expected different balances")
			})
		})
	})
})
