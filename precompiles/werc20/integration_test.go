package werc20_test

import (
	"math/big"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	erc20precompile "github.com/evmos/evmos/v16/precompiles/erc20"
	"github.com/evmos/evmos/v16/precompiles/testutil"
	"github.com/evmos/evmos/v16/precompiles/werc20"
	"github.com/evmos/evmos/v16/precompiles/werc20/testdata"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	testutils "github.com/evmos/evmos/v16/testutil/integration/evmos/utils"
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

const chainID = "evmos_9001-1"

var _ = Describe("WEVMOS Extension -", Ordered, func() {
	var (
		s *WERC20IntegrationTestSuite

		senderKey       testkeyring.Key
		contractAddress common.Address
		contractABI     abi.ABI

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

		senderKey = s.keyring.GetKey(1)
		contractAddress = common.HexToAddress(erc20precompile.WEVMOSContractMainnet)
		contractABI, err = werc20.LoadABI()
		Expect(err).To(BeNil(), "failed to load WERC-20 ABI")

		expPass = testutil.LogCheckArgs{
			ABIEvents: contractABI.Events,
			ExpPass:   false,
		}
	})

	// TODO: remove this level of nesting, unnecessary
	Context("WEVMOS specific functions", func() {
		When("calling deposit correctly", func() {
			It("should not emit events", func() {
				txArgs := evmtypes.EvmTxArgs{
					To:     &contractAddress,
					Amount: big.NewInt(100),
				}

				depositArgs := factory.CallArgs{
					ContractABI: contractABI,
					MethodName:  werc20.DepositMethod,
				}

				depositResponse, err := s.factory.ExecuteContractCall(
					senderKey.Priv, txArgs, depositArgs,
				)
				Expect(err).To(BeNil(), "failed to call contract")
				Expect(depositResponse.IsOK()).To(Equal(true), "transaction should have succeeded", depositResponse.GetLog())
				Expect(depositResponse.GasUsed).To(BeNumerically(">=", werc20.DepositRequiredGas), "expected different gas used")
			})
		})

		When("calling withdraw correctly", func() {
			It("should not emit events", func() {
				txArgs := evmtypes.EvmTxArgs{
					To: &contractAddress,
				}

				withdrawArgs := factory.CallArgs{
					ContractABI: contractABI,
					MethodName:  werc20.WithdrawMethod,
					Args: []interface{}{
						big.NewInt(100),
					},
				}

				withdrawResponse, err := s.factory.ExecuteContractCall(
					senderKey.Priv, txArgs, withdrawArgs,
				)
				Expect(err).To(BeNil(), "failed to call contract")
				Expect(withdrawResponse.IsOK()).To(Equal(true), "transaction should have succeeded", withdrawResponse.GetLog())
				Expect(withdrawResponse.GasUsed).To(BeNumerically(">=", werc20.WithdrawRequiredGas), "expected different gas used")
			})
		})

		// TODO: How do we actually check the method types here? We can see the correct ones being populated by printing the line in the cmn.Precompile
		//
		// FIXME: address TODO here?! What exactly is the question? If the fallback was correctly executed?
		When("calling with empty method but non-zero amount", func() {
			It("should call `receive`", func() {
				txArgs := evmtypes.EvmTxArgs{
					To:     &contractAddress,
					Amount: big.NewInt(100),
				}

				receiveArgs := factory.CallArgs{
					ContractABI: contractABI,
					MethodName:  "",
				}
				receiveResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, txArgs, receiveArgs)
				Expect(err).To(BeNil(), "unexpected result calling contract")
				Expect(receiveResponse.IsOK()).To(Equal(true), "transaction should have succeeded", receiveResponse.GetLog())
			})
		})

		When("calling with short call data, empty method and non-zero amount", func() {
			It("should call `fallback` ", func() {
				txArgs := evmtypes.EvmTxArgs{
					To:     &contractAddress,
					Amount: big.NewInt(100),
					Input:  []byte{1, 2, 3},
				}

				receiveArgs := factory.CallArgs{
					ContractABI: contractABI,
					MethodName:  "",
				}
				receiveResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, txArgs, receiveArgs)
				Expect(err).To(BeNil(), "unexpected result calling contract")
				Expect(receiveResponse.IsOK()).To(Equal(true), "transaction should have succeeded", receiveResponse.GetLog())
			})
		})

		When("calling a non-existing function with amount", func() {
			It("should call `fallback` ", func() {
				txArgs := evmtypes.EvmTxArgs{
					To:     &contractAddress,
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

		// 		It("calls non call data, without amount - should call `fallback` ", func() {
		// 			txArgs, _ := s.getTxAndCallArgs(erc20Call, contractData, "")

		// 			res, err := s.factory.ExecuteEthTx(sender.Priv, txArgs)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			depositCheck := expPass.WithExpPass(true)
		// 			depositCheck.Res = res
		// 			err = testutil.CheckLogs(depositCheck)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			s.checkBalances(failCheck, sender, contractData)
		// 		})

		// 		It("calls short call data, without amount - should call `fallback` ", func() {
		// 			txArgs, _ := s.getTxAndCallArgs(erc20Call, contractData, "")
		// 			txArgs.Input = []byte{1, 2, 3} // 3 dummy bytes

		// 			res, err := s.factory.ExecuteEthTx(sender.Priv, txArgs)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			depositCheck := expPass.WithExpPass(true)
		// 			depositCheck.Res = res
		// 			err = testutil.CheckLogs(depositCheck)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			s.checkBalances(failCheck, sender, contractData)
		// 		})

		// 		It("calls with non-existing function, without amount -  should call `fallback` ", func() {
		// 			txArgs, _ := s.getTxAndCallArgs(erc20Call, contractData, "")
		// 			txArgs.Input = []byte("nonExistingMethod")

		// 			res, err := s.factory.ExecuteEthTx(sender.Priv, txArgs)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			depositCheck := expPass.WithExpPass(true)
		// 			depositCheck.Res = res
		// 			err = testutil.CheckLogs(depositCheck)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			s.checkBalances(failCheck, sender, contractData)
		// 		})
		// 	})
		// })

		Context("Comparing to original WEVMOS contract (not precompiled)", func() {
			var WEVMOSOriginalContractAddr common.Address

			BeforeAll(func() {
				var err error
				WEVMOSOriginalContractAddr, err = s.factory.DeployContract(
					s.keyring.GetKey(1).Priv,
					evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
					factory.ContractDeploymentData{
						Contract:        testdata.WEVMOSContract,
						ConstructorArgs: []interface{}{},
					},
				)
				Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")
			})

			When("calling deposit", func() {
				It("should have exact gas consumption", func() {
					txArgs := evmtypes.EvmTxArgs{
						To:       &contractAddress,
						Amount:   big.NewInt(100),
						GasLimit: 50_000, // FIXME: why is the gas limit not enough here by default? Raises out of gas error if not provided
					}

					depositArgs := factory.CallArgs{
						ContractABI: contractABI,
						MethodName:  werc20.DepositMethod,
					}

					depositResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, txArgs, depositArgs)
					Expect(err).To(BeNil(), "failed to call contract")
					Expect(depositResponse.IsOK()).To(Equal(true), "transaction should have succeeded", depositResponse.GetLog())

					originalTxArgs := txArgs
					originalTxArgs.To = &WEVMOSOriginalContractAddr

					originalDepositResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, originalTxArgs, depositArgs)
					Expect(err).To(BeNil(), "failed to call contract")
					Expect(originalDepositResponse.IsOK()).To(
						Equal(true),
						"transaction should have succeeded",
						originalDepositResponse.GetLog(),
					)

					// FIXME: why is the gas consumption not equal?
					Expect(depositResponse.GasUsed).To(
						Equal(originalDepositResponse.GasUsed),
						"expected same gas used between smart contract and precompile",
					)
				})
			})

			It("should return the same error", func() {
				// Hardcode gas limit to search for error
				// Avoid simulate tx to fail on execution
				txArgs := evmtypes.EvmTxArgs{
					To:       &contractAddress,
					Amount:   big.NewInt(9e18),
					GasLimit: 50_000,
				}

				depositArgs := factory.CallArgs{
					ContractABI: contractABI,
					MethodName:  werc20.DepositMethod,
					Args:        []interface{}{},
				}

				depositResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, txArgs, depositArgs)
				Expect(err).ToNot(BeNil(), "expected error but got none")
				Expect(depositResponse.IsOK()).To(
					Equal(false),
					"transaction should have failed",
					depositResponse.GetLog(),
				)

				originalTxArgs := txArgs
				originalTxArgs.To = &WEVMOSOriginalContractAddr

				originalDepositResponse, errOriginal := s.factory.ExecuteContractCall(senderKey.Priv, originalTxArgs, depositArgs)
				Expect(err).ToNot(BeNil(), "expected error but got none")
				Expect(originalDepositResponse.IsOK()).To(
					Equal(false),
					"transaction should have failed",
					originalDepositResponse.GetLog(),
				)
				Expect(errOriginal.Error()).To(
					Equal(err.Error()),
					"expected same error for original and precompiled contracts",
				)
			})

			// 		It("should reflect the correct balances", func() {
			// 			depositCheck := expPass.WithExpPass(true)
			// 			txArgsPrecompile, callArgsPrecompile := s.getTxAndCallArgs(erc20Call, contractData, werc20.DepositMethod)
			// 			txArgsPrecompile.Amount = amount

			// 			_, _, errPrecompile := s.factory.CallContractAndCheckLogs(sender.Priv, txArgsPrecompile, callArgsPrecompile, depositCheck)
			// 			Expect(errPrecompile).ToNot(HaveOccurred(), "unexpected result calling contract")

			// 			txArgsContract, callArgsContract := s.getTxAndCallArgs(erc20Call, contractDataOriginal, werc20.DepositMethod)
			// 			txArgsContract.Amount = amount
			// 			txArgsContract.GasLimit = 50_000

			// 			depositCheckContract := expPass.WithExpPass(true).WithExpEvents(EventTypeDeposit)
			// 			_, _, errOriginal := s.factory.CallContractAndCheckLogs(sender.Priv, txArgsContract, callArgsContract, depositCheckContract)
			// 			Expect(errOriginal).ToNot(HaveOccurred(), "unexpected result calling contract")

			// 			// Check balances after calling precompile
			// 			s.checkBalances(failCheck, sender, contractData)

			// 			// Check balances after calling original contract
			// 			balanceCheck := failCheck.WithExpPass(true)
			// 			txArgs, balancesArgs := s.getTxAndCallArgs(erc20Call, contractDataOriginal, erc20.BalanceOfMethod, sender.Addr)

			// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, balanceCheck)
			// 			Expect(err).ToNot(HaveOccurred(), "failed to execute balanceOf")

			// 			// Check the balance in the bank module is the same as calling `balanceOf` on the precompile
			// 			var erc20Balance *big.Int
			// 			//err = .UnpackIntoInterface(&erc20Balance, erc20.BalanceOfMethod, ethRes.Ret)
			// 			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			// 			Expect(erc20Balance).To(Equal(amount), "expected different balance")
			// 		})
		})

		// 	When("calling withdraw", func() {
		// 		BeforeEach(func() {
		// 			// Deposit into the WEVMOS contract to have something to withdraw
		// 			depositCheck := expPass.WithExpPass(true).WithExpEvents(EventTypeDeposit)
		// 			txArgsContract, callArgsContract := s.getTxAndCallArgs(erc20Call, contractDataOriginal, werc20.DepositMethod)
		// 			txArgsContract.Amount = amount
		// 			txArgsContract.GasLimit = 50_000

		// 			_, _, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgsContract, callArgsContract, depositCheck)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
		// 		})

		// 		It("should have exact gas consumption", func() {
		// 			withdrawCheck := expPass.WithExpPass(true)
		// 			txArgsPrecompile, callArgsPrecompile := s.getTxAndCallArgs(erc20Call, contractData, werc20.WithdrawMethod, amount)

		// 			_, ethResPrecompile, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgsPrecompile, callArgsPrecompile, withdrawCheck)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			txArgsContract, callArgsContract := s.getTxAndCallArgs(erc20Call, contractDataOriginal, werc20.WithdrawMethod, amount)

		// 			withdrawCheckContract := expPass.WithExpPass(true).WithExpEvents(EventTypeWithdrawal)
		// 			_, ethResOriginal, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgsContract, callArgsContract, withdrawCheckContract)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			Expect(ethResOriginal.GasUsed).To(Equal(ethResPrecompile.GasUsed), "expected exact gas used")
		// 		})

		// 		It("should return the same error", func() {
		// 			withdrawCheck := expPass.WithExpPass(true)
		// 			txArgsPrecompile, callArgsPrecompile := s.getTxAndCallArgs(erc20Call, contractData, werc20.WithdrawMethod)

		// 			_, _, errPrecompile := s.factory.CallContractAndCheckLogs(sender.Priv, txArgsPrecompile, callArgsPrecompile, withdrawCheck)
		// 			Expect(errPrecompile).To(HaveOccurred(), "unexpected result calling contract")

		// 			txArgsContract, callArgsContract := s.getTxAndCallArgs(erc20Call, contractDataOriginal, werc20.WithdrawMethod)
		// 			txArgsContract.GasLimit = 50_000

		// 			_, _, errOriginal := s.factory.CallContractAndCheckLogs(sender.Priv, txArgsContract, callArgsContract, withdrawCheck)
		// 			Expect(errOriginal).To(HaveOccurred(), "unexpected result calling contract")

		// 			Expect(errOriginal.Error()).To(Equal(errPrecompile.Error()), "expected same error")
		// 		})
		// 	})
		// })

		// Context("ERC20 specific functions", func() {
		// 	When("querying name", func() {
		// 		It("should return the correct name", func() {
		// 			// Query the name
		// 			txArgs, nameArgs := s.getTxAndCallArgs(directCall, contractData, erc20.NameMethod)

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, nameArgs, expPass)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			var name string
		// 			err = s.precompile.UnpackIntoInterface(&name, erc20.NameMethod, ethRes.Ret)
		// 			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
		// 			Expect(name).To(Equal("Evmos"), "expected different name")
		// 		})
		// 	})

		// 	When("querying symbol", func() {
		// 		It("should return the correct symbol", func() {
		// 			// Query the symbol
		// 			txArgs, symbolArgs := s.getTxAndCallArgs(directCall, contractData, erc20.SymbolMethod)

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, symbolArgs, expPass)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			var symbol string
		// 			err = s.precompile.UnpackIntoInterface(&symbol, erc20.SymbolMethod, ethRes.Ret)
		// 			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
		// 			Expect(symbol).To(Equal("EVMOS"), "expected different symbol")
		// 		})
		// 	})

		// 	When("querying decimals", func() {
		// 		It("should return the correct decimals", func() {
		// 			// Query the decimals
		// 			txArgs, decimalsArgs := s.getTxAndCallArgs(directCall, contractData, erc20.DecimalsMethod)

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, decimalsArgs, expPass)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			var decimals uint8
		// 			err = s.precompile.UnpackIntoInterface(&decimals, erc20.DecimalsMethod, ethRes.Ret)
		// 			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
		// 			Expect(decimals).To(Equal(uint8(18)), "expected different decimals")
		// 		})
		// 	})

		// 	When("querying balance", func() {
		// 		It("should return an existing balance", func() {
		// 			// Query the balance
		// 			txArgs, balancesArgs := s.getTxAndCallArgs(directCall, contractData, erc20.BalanceOfMethod, sender.Addr)

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, expPass)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			expBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), sender.AccAddr, s.network.GetDenom())

		// 			var balance *big.Int
		// 			err = s.precompile.UnpackIntoInterface(&balance, erc20.BalanceOfMethod, ethRes.Ret)
		// 			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
		// 			Expect(balance).To(Equal(expBalance.Amount.BigInt()), "expected different balance")
		// 		})

		// 		It("should return a 0 balance new address", func() {
		// 			// Query the balance
		// 			txArgs, balancesArgs := s.getTxAndCallArgs(directCall, contractData, erc20.BalanceOfMethod, evmosutiltx.GenerateAddress())

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, expPass)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			var balance *big.Int
		// 			err = s.precompile.UnpackIntoInterface(&balance, erc20.BalanceOfMethod, ethRes.Ret)
		// 			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
		// 			Expect(balance.Int64()).To(Equal(int64(0)), "expected different balance")
		// 		})
		// 	})

		// 	When("querying allowance", func() {
		// 		It("should return an existing allowance", func() {
		// 			grantee := evmosutiltx.GenerateAddress()
		// 			granter := sender
		// 			authzCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 100)}

		// 			s.setupSendAuthzForContract(directCall, grantee, granter.Priv, authzCoins)

		// 			txArgs, allowanceArgs := s.getTxAndCallArgs(directCall, contractData, auth.AllowanceMethod, granter.Addr, grantee)

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, expPass)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			var allowance *big.Int
		// 			err = s.precompile.UnpackIntoInterface(&allowance, auth.AllowanceMethod, ethRes.Ret)
		// 			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
		// 			Expect(allowance).To(Equal(authzCoins[0].Amount.BigInt()), "expected different allowance")
		// 		})

		// 		It("should return zero if no balance exists", func() {
		// 			address := evmosutiltx.GenerateAddress()

		// 			// Query the balance
		// 			txArgs, balancesArgs := s.getTxAndCallArgs(directCall, contractData, erc20.BalanceOfMethod, address)

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, expPass)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			var balance *big.Int
		// 			err = s.precompile.UnpackIntoInterface(&balance, erc20.BalanceOfMethod, ethRes.Ret)
		// 			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
		// 			Expect(balance.Int64()).To(BeZero(), "expected zero balance")
		// 		})
		// 	})

		// 	When("querying total supply", func() {
		// 		It("should return the total supply", func() {
		// 			expSupply, ok := new(big.Int).SetString("11000000000000000000", 10)
		// 			Expect(ok).To(BeTrue(), "failed to parse expected supply")

		// 			// Query the balance
		// 			txArgs, supplyArgs := s.getTxAndCallArgs(directCall, contractData, erc20.TotalSupplyMethod)

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, supplyArgs, expPass)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			var supply *big.Int
		// 			err = s.precompile.UnpackIntoInterface(&supply, erc20.TotalSupplyMethod, ethRes.Ret)
		// 			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
		// 			Expect(supply).To(Equal(expSupply), "expected different supply")
		// 		})
		// 	})

		// 	When("transferring tokens", func() {
		// 		It("it should transfer tokens to a receiver using `transfer`", func() {
		// 			// Get receiver address
		// 			receiver := s.keyring.GetKey(1)

		// 			senderBalance := s.network.App.BankKeeper.GetAllBalances(s.network.GetContext(), sender.AccAddr)
		// 			receiverBalance := s.network.App.BankKeeper.GetAllBalances(s.network.GetContext(), receiver.AccAddr)

		// 			// Transfer tokens
		// 			txArgs, transferArgs := s.getTxAndCallArgs(directCall, contractData, erc20.TransferMethod, receiver.Addr, amount)
		// 			// Prefilling the gas price with the base fee to calculate expected balances after
		// 			// the transfer
		// 			baseFeeRes, err := s.grpcHandler.GetBaseFee()
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected error querying base fee")
		// 			txArgs.GasPrice = baseFeeRes.BaseFee.BigInt()

		// 			transferCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), amount.Int64())}

		// 			transferCheck := expPass.WithExpEvents(erc20.EventTypeTransfer)
		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, transferCheck)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			gasAmount := ethRes.GasUsed * txArgs.GasPrice.Uint64()
		// 			coinsWithGasIncluded := transferCoins.Add(sdk.NewInt64Coin(s.network.GetDenom(), int64(gasAmount)))
		// 			s.ExpectBalances(
		// 				[]ExpectedBalance{
		// 					{address: sender.AccAddr, expCoins: senderBalance.Sub(coinsWithGasIncluded...)},
		// 					{address: receiver.AccAddr, expCoins: receiverBalance.Add(transferCoins...)},
		// 				},
		// 			)
		// 		})

		// 		It("it should transfer tokens to a receiver using `transferFrom`", func() {
		// 			// Get receiver address
		// 			receiver := s.keyring.GetKey(1)

		// 			senderBalance := s.network.App.BankKeeper.GetAllBalances(s.network.GetContext(), sender.AccAddr)
		// 			receiverBalance := s.network.App.BankKeeper.GetAllBalances(s.network.GetContext(), receiver.AccAddr)

		// 			// Transfer tokens
		// 			txArgs, transferArgs := s.getTxAndCallArgs(directCall, contractData, erc20.TransferFromMethod, sender.Addr, receiver.Addr, amount)
		// 			// Prefilling the gas price with the base fee to calculate expected balances after
		// 			// the transfer
		// 			baseFeeRes, err := s.grpcHandler.GetBaseFee()
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected error querying base fee")
		// 			txArgs.GasPrice = baseFeeRes.BaseFee.BigInt()

		// 			transferCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), amount.Int64())}

		// 			transferCheck := expPass.WithExpEvents(erc20.EventTypeTransfer, auth.EventTypeApproval)
		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, transferCheck)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			gasAmount := ethRes.GasUsed * txArgs.GasPrice.Uint64()
		// 			coinsWithGasIncluded := transferCoins.Add(sdk.NewInt64Coin(s.network.GetDenom(), int64(gasAmount)))
		// 			s.ExpectBalances(
		// 				[]ExpectedBalance{
		// 					{address: sender.AccAddr, expCoins: senderBalance.Sub(coinsWithGasIncluded...)},
		// 					{address: receiver.AccAddr, expCoins: receiverBalance.Add(transferCoins...)},
		// 				},
		// 			)
		// 		})
		// })
	})
})
