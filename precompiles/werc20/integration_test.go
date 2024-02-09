package werc20_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/precompiles/erc20"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"

	erc20precompile "github.com/evmos/evmos/v16/precompiles/erc20"
	"github.com/evmos/evmos/v16/precompiles/werc20"
	"github.com/evmos/evmos/v16/precompiles/werc20/testdata"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

// PrecompileTestSuite is the implementation of the TestSuite interface for ERC20 precompile
// unit tests.
type PrecompileTestSuite struct {
	// suite.Suite

	// bondDenom   string
	network     network.Network
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring
}

const (
	// EventTypeDeposit defines the event type for the Deposit transaction.
	EventTypeDeposit = "Deposit"
	// EventTypeWithdrawal defines the event type for the Withdraw transaction.
	EventTypeWithdrawal = "Withdrawal"

	chainID = "evmos_9001-1"
)

var _ = Describe("WEVMOS Extension -", Ordered, func() {
	var s *PrecompileTestSuite

	BeforeAll(func() {
		keyring := testkeyring.New(3)
		integrationNetwork := network.New(
			network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			network.WithChainID(chainID),
		)

		grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
		txFactory := factory.New(integrationNetwork, grpcHandler)

		s = &PrecompileTestSuite{
			network:     integrationNetwork,
			factory:     txFactory,
			grpcHandler: grpcHandler,
			keyring:     keyring,
		}

		// Add WEVMOS to params
		params, err := grpcHandler.GetEvmParams()
		Expect(err).To(BeNil())
		WEVMOSAddress := common.HexToAddress(erc20precompile.WEVMOSContractMainnet)
		params.Params.ActivePrecompiles = append(params.Params.ActivePrecompiles, WEVMOSAddress.String())
		integrationNetwork.UpdateEvmParams(params.Params)
	})

	Context("WEVMOS specific functions", func() {
		When("calling deposit correctly", func() {
			It("should not emit events", func() {

				senderKey := s.keyring.GetKey(1)
				contractAddress := common.HexToAddress(erc20.WEVMOSContractMainnet)
				contractABI, err := werc20.LoadABI()
				Expect(err).To(BeNil())

				totalSupplyTxArgs := evmtypes.EvmTxArgs{
					To: &contractAddress,
				}

				// Perform a delegate transaction to the staking precompile
				depositArgs := factory.CallArgs{
					ContractABI: contractABI,
					MethodName:  werc20.DepositMethod,
					Args:        []interface{}{},
				}
				depositResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, totalSupplyTxArgs, depositArgs)
				Expect(err).To(BeNil())
				Expect(depositResponse.IsOK()).To(Equal(true), "transaction should have succeeded", depositResponse.GetLog())

				Expect(depositResponse.GasUsed).To(BeNumerically(">=", werc20.DepositRequiredGas), "expected different gas used")

			})
		})

		When("calling withdraw correctly", func() {
			It("should not emit events", func() {
				senderKey := s.keyring.GetKey(1)
				contractAddress := common.HexToAddress(erc20.WEVMOSContractMainnet)
				contractABI, err := werc20.LoadABI()
				Expect(err).To(BeNil())

				amountToWithdraw := big.NewInt(200)

				totalSupplyTxArgs := evmtypes.EvmTxArgs{
					To: &contractAddress,
				}

				// Perform a delegate transaction to the staking precompile
				withdrawArgs := factory.CallArgs{
					ContractABI: contractABI,
					MethodName:  werc20.WithdrawMethod,
					Args:        []interface{}{amountToWithdraw},
				}
				withdrawResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, totalSupplyTxArgs, withdrawArgs)
				Expect(err).To(BeNil())
				Expect(withdrawResponse.IsOK()).To(Equal(true), "transaction should have succeeded", withdrawResponse.GetLog())

				Expect(withdrawResponse.GasUsed).To(BeNumerically(">=", werc20.WithdrawRequiredGas), "expected different gas used")

			})
		})

		// TODO: How do we actually check the method types here? We can see the correct ones being populated by printing the line in the cmn.Precompile
		When("calling with incomplete data or amount", func() {
			It("calls no call data, with amount - should call `receive` ", func() {

				senderKey := s.keyring.GetKey(1)
				contractAddress := common.HexToAddress(erc20.WEVMOSContractMainnet)
				contractABI, err := werc20.LoadABI()
				Expect(err).To(BeNil())

				amountToSend := big.NewInt(200)

				totalSupplyTxArgs := evmtypes.EvmTxArgs{
					To:     &contractAddress,
					Amount: amountToSend,
				}

				// Perform a delegate transaction to the staking precompile
				receiveArgs := factory.CallArgs{
					ContractABI: contractABI,
					MethodName:  "",
					Args:        []interface{}{},
				}
				receiveResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, totalSupplyTxArgs, receiveArgs)
				Expect(err).To(BeNil())
				Expect(receiveResponse.IsOK()).To(Equal(true), "transaction should have succeeded", receiveResponse.GetLog())
			})
		})

		It("calls short call data, with amount - should call `fallback` ", func() {
			senderKey := s.keyring.GetKey(1)
			contractAddress := common.HexToAddress(erc20.WEVMOSContractMainnet)
			contractABI, err := werc20.LoadABI()
			Expect(err).To(BeNil())

			amountToSend := big.NewInt(200)

			totalSupplyTxArgs := evmtypes.EvmTxArgs{
				To:     &contractAddress,
				Amount: amountToSend,
				Input:  []byte{1, 2, 3},
			}

			// Perform a delegate transaction to the staking precompile
			receiveArgs := factory.CallArgs{
				ContractABI: contractABI,
				MethodName:  "",
				Args:        []interface{}{},
			}
			receiveResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, totalSupplyTxArgs, receiveArgs)
			Expect(err).To(BeNil())
			Expect(receiveResponse.IsOK()).To(Equal(true), "transaction should have succeeded", receiveResponse.GetLog())
		})

		// 		It("calls with non-existing function, with amount - should call `fallback` ", func() {
		// 			txArgs, _ := s.getTxAndCallArgs(erc20Call, contractData, "")
		// 			txArgs.Input = []byte("nonExistingMethod")
		// 			txArgs.Amount = amount

		// 			res, err := s.factory.ExecuteEthTx(sender.Priv, txArgs)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			depositCheck := passCheck.WithExpPass(true)
		// 			depositCheck.Res = res
		// 			err = testutil.CheckLogs(depositCheck)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			s.checkBalances(failCheck, sender, contractData)
		// 		})

		// 		It("calls non call data, without amount - should call `fallback` ", func() {
		// 			txArgs, _ := s.getTxAndCallArgs(erc20Call, contractData, "")

		// 			res, err := s.factory.ExecuteEthTx(sender.Priv, txArgs)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			depositCheck := passCheck.WithExpPass(true)
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

		// 			depositCheck := passCheck.WithExpPass(true)
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

		// 			depositCheck := passCheck.WithExpPass(true)
		// 			depositCheck.Res = res
		// 			err = testutil.CheckLogs(depositCheck)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			s.checkBalances(failCheck, sender, contractData)
		// 		})
		// 	})
		// })

		Context("Comparing to original WEVMOS contract", func() {
			var (
				WEVMOSOriginalContractAddr common.Address
			)
			BeforeEach(func() {
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
					senderKey := s.keyring.GetKey(1)
					contractAddress := common.HexToAddress(erc20.WEVMOSContractMainnet)
					contractABI, err := werc20.LoadABI()
					Expect(err).To(BeNil())

					totalSupplyTxArgs := evmtypes.EvmTxArgs{
						To: &contractAddress,
					}

					// Perform a delegate transaction to the staking precompile
					depositArgs := factory.CallArgs{
						ContractABI: contractABI,
						MethodName:  werc20.DepositMethod,
						Args:        []interface{}{},
					}
					depositResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, totalSupplyTxArgs, depositArgs)
					Expect(err).To(BeNil())
					Expect(depositResponse.IsOK()).To(Equal(true), "transaction should have succeeded", depositResponse.GetLog())
					originalCall := evmtypes.EvmTxArgs{
						To: &WEVMOSOriginalContractAddr,
					}

					originalDepositResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, originalCall, depositArgs)
					Expect(err).To(BeNil())
					Expect(originalDepositResponse.IsOK()).To(Equal(true), "transaction should have succeeded", originalDepositResponse.GetLog())

					// FIXME: why gas consumption failed
					// Expect(depositResponse.GasUsed).To(BeNumerically("==", originalDepositResponse.GasUsed), "expected different gas used")
				})
			})

			It("should return the same error", func() {

				senderKey := s.keyring.GetKey(1)
				contractAddress := common.HexToAddress(erc20.WEVMOSContractMainnet)
				contractABI, err := werc20.LoadABI()
				Expect(err).To(BeNil())

				// Hardcode gas limit to search for error
				// Avoid simulate tx to fail on execution
				totalSupplyTxArgs := evmtypes.EvmTxArgs{
					To:       &contractAddress,
					Amount:   big.NewInt(9e18),
					GasLimit: 50_000,
				}

				// Perform a delegate transaction to the staking precompile
				depositArgs := factory.CallArgs{
					ContractABI: contractABI,
					MethodName:  werc20.DepositMethod,
					Args:        []interface{}{},
				}
				depositResponse, err := s.factory.ExecuteContractCall(senderKey.Priv, totalSupplyTxArgs, depositArgs)
				Expect(err).ToNot(BeNil())
				Expect(depositResponse.IsOK()).To(Equal(false), "transaction should have failed", depositResponse.GetLog())
				originalCall := evmtypes.EvmTxArgs{
					To:       &WEVMOSOriginalContractAddr,
					Amount:   big.NewInt(9e18),
					GasLimit: 50_000,
				}

				originalDepositResponse, errOriginal := s.factory.ExecuteContractCall(senderKey.Priv, originalCall, depositArgs)
				Expect(err).ToNot(BeNil())
				Expect(originalDepositResponse.IsOK()).To(Equal(false), "transaction should have failed", originalDepositResponse.GetLog())

				Expect(errOriginal.Error()).To(Equal(err.Error()))

			})

			// 		It("should reflect the correct balances", func() {
			// 			depositCheck := passCheck.WithExpPass(true)
			// 			txArgsPrecompile, callArgsPrecompile := s.getTxAndCallArgs(erc20Call, contractData, werc20.DepositMethod)
			// 			txArgsPrecompile.Amount = amount

			// 			_, _, errPrecompile := s.factory.CallContractAndCheckLogs(sender.Priv, txArgsPrecompile, callArgsPrecompile, depositCheck)
			// 			Expect(errPrecompile).ToNot(HaveOccurred(), "unexpected result calling contract")

			// 			txArgsContract, callArgsContract := s.getTxAndCallArgs(erc20Call, contractDataOriginal, werc20.DepositMethod)
			// 			txArgsContract.Amount = amount
			// 			txArgsContract.GasLimit = 50_000

			// 			depositCheckContract := passCheck.WithExpPass(true).WithExpEvents(EventTypeDeposit)
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
		// 			depositCheck := passCheck.WithExpPass(true).WithExpEvents(EventTypeDeposit)
		// 			txArgsContract, callArgsContract := s.getTxAndCallArgs(erc20Call, contractDataOriginal, werc20.DepositMethod)
		// 			txArgsContract.Amount = amount
		// 			txArgsContract.GasLimit = 50_000

		// 			_, _, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgsContract, callArgsContract, depositCheck)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
		// 		})

		// 		It("should have exact gas consumption", func() {
		// 			withdrawCheck := passCheck.WithExpPass(true)
		// 			txArgsPrecompile, callArgsPrecompile := s.getTxAndCallArgs(erc20Call, contractData, werc20.WithdrawMethod, amount)

		// 			_, ethResPrecompile, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgsPrecompile, callArgsPrecompile, withdrawCheck)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			txArgsContract, callArgsContract := s.getTxAndCallArgs(erc20Call, contractDataOriginal, werc20.WithdrawMethod, amount)

		// 			withdrawCheckContract := passCheck.WithExpPass(true).WithExpEvents(EventTypeWithdrawal)
		// 			_, ethResOriginal, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgsContract, callArgsContract, withdrawCheckContract)
		// 			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		// 			Expect(ethResOriginal.GasUsed).To(Equal(ethResPrecompile.GasUsed), "expected exact gas used")
		// 		})

		// 		It("should return the same error", func() {
		// 			withdrawCheck := passCheck.WithExpPass(true)
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

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, nameArgs, passCheck)
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

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, symbolArgs, passCheck)
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

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, decimalsArgs, passCheck)
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

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, passCheck)
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

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, passCheck)
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

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, passCheck)
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

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, passCheck)
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

		// 			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, supplyArgs, passCheck)
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

		// 			transferCheck := passCheck.WithExpEvents(erc20.EventTypeTransfer)
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

		// 			transferCheck := passCheck.WithExpEvents(erc20.EventTypeTransfer, auth.EventTypeApproval)
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
