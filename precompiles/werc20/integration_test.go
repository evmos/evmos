// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package werc20_test

import (
	"math/big"
	"strings"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	auth "github.com/evmos/evmos/v20/precompiles/authorization"
	"github.com/evmos/evmos/v20/precompiles/erc20"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	"github.com/evmos/evmos/v20/precompiles/werc20"
	"github.com/evmos/evmos/v20/precompiles/werc20/testdata"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"

	"github.com/ethereum/go-ethereum/common"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

// -------------------------------------------------------------------------------------------------
// Integration test suite
// -------------------------------------------------------------------------------------------------

type PrecompileIntegrationTestSuite struct {
	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     keyring.Keyring

	wrappedCoinDenom string

	// WEVMOS related fields
	precompile        *werc20.Precompile
	precompileAddrHex string
}

func TestPrecompileIntegrationTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "WEVMOS precompile test suite")
}

// checkAndReturnBalance check that the balance of the address is the same in
// the smart contract and in the balance and returns the amount.
func (is *PrecompileIntegrationTestSuite) checkAndReturnBalance(
	balanceCheck testutil.LogCheckArgs,
	callsData CallsData,
	address common.Address,
) *big.Int {
	txArgs, balancesArgs := callsData.getTxAndCallArgs(directCall, erc20.BalanceOfMethod, address)
	txArgs.GasLimit = 1_000_000_000_000

	_, ethRes, err := is.factory.CallContractAndCheckLogs(callsData.sender.Priv, txArgs, balancesArgs, balanceCheck)
	Expect(err).ToNot(HaveOccurred(), "failed to execute balanceOf")
	var erc20Balance *big.Int
	err = is.precompile.UnpackIntoInterface(&erc20Balance, erc20.BalanceOfMethod, ethRes.Ret)
	Expect(err).ToNot(HaveOccurred(), "failed to unpack result")

	addressAcc := sdk.AccAddress(address.Bytes())
	balanceAfter, err := is.grpcHandler.GetBalance(addressAcc, is.wrappedCoinDenom)
	Expect(err).ToNot(HaveOccurred(), "expected no error getting balance")

	Expect(erc20Balance.String()).To(Equal(balanceAfter.Balance.Amount.BigInt().String()), "expected return balance from contract equal to bank")
	return erc20Balance
}

// -------------------------------------------------------------------------------------------------
// Integration tests
// -------------------------------------------------------------------------------------------------

var _ = When("a user interact with the WEVMOS precompiled contract", func() {
	var (
		is                                         *PrecompileIntegrationTestSuite
		passCheck, failCheck                       testutil.LogCheckArgs
		transferCheck, depositCheck, withdrawCheck testutil.LogCheckArgs

		callsData CallsData

		txSender, user keyring.Key

		revertContractAddr common.Address
	)

	depositAmount := big.NewInt(1e18)
	withdrawAmount := depositAmount
	transferAmount := depositAmount

	BeforeEach(func() {
		is = new(PrecompileIntegrationTestSuite)
		keyring := keyring.New(2)

		txSender = keyring.GetKey(0)
		user = keyring.GetKey(1)

		// Set the base fee to zero to allow for zero cost tx. The final gas cost is
		// not part of the logic tested here so this makes testing more easy.
		customGenesis := network.CustomGenesisState{}
		feemarketGenesis := feemarkettypes.DefaultGenesisState()
		feemarketGenesis.Params.NoBaseFee = true
		customGenesis[feemarkettypes.ModuleName] = feemarketGenesis

		integrationNetwork := network.NewUnitTestNetwork(
			network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			network.WithCustomGenesis(customGenesis),
		)
		grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
		txFactory := factory.New(integrationNetwork, grpcHandler)

		is.network = integrationNetwork
		is.factory = txFactory
		is.grpcHandler = grpcHandler
		is.keyring = keyring

		is.wrappedCoinDenom = evmtypes.GetEVMCoinDenom()
		is.precompileAddrHex = erc20types.GetWEVMOSContractHex(is.network.GetChainID())

		ctx := integrationNetwork.GetContext()

		// Perform some check before adding the precompile to the suite.

		// Check that WEVMOS is part of the native precompiles.
		erc20Params := is.network.App.Erc20Keeper.GetParams(ctx)
		Expect(erc20Params.NativePrecompiles).To(
			ContainElement(is.precompileAddrHex),
			"expected wevmos to be in the native precompiles",
		)
		_, found := is.network.App.BankKeeper.GetDenomMetaData(ctx, evmtypes.GetEVMCoinDenom())
		Expect(found).To(BeTrue(), "expected native token metadata to be registered")

		// Check that WEVMOS is registered in the token pairs map.
		tokenPairID := is.network.App.Erc20Keeper.GetTokenPairID(ctx, is.wrappedCoinDenom)
		tokenPair, found := is.network.App.Erc20Keeper.GetTokenPair(ctx, tokenPairID)
		Expect(found).To(BeTrue(), "expected wevmos precompile to be registered in the tokens map")
		Expect(tokenPair.Erc20Address).To(Equal(is.precompileAddrHex))

		precompileAddr := common.HexToAddress(is.precompileAddrHex)
		tokenPair = erc20types.NewTokenPair(
			precompileAddr,
			evmtypes.GetEVMCoinDenom(),
			erc20types.OWNER_MODULE,
		)
		precompile, err := werc20.NewPrecompile(
			tokenPair,
			is.network.App.BankKeeper,
			is.network.App.AuthzKeeper,
			is.network.App.TransferKeeper,
		)
		Expect(err).ToNot(HaveOccurred(), "failed to instantiate the werc20 precompile")
		is.precompile = precompile

		// Setup of the contract calling into the precompile to tests revert
		// edge cases and proper handling of snapshots.
		revertCallerContract, err := testdata.LoadWEVMOS9TestCaller()
		Expect(err).ToNot(HaveOccurred(), "failed to load werc20 reverter caller contract")

		txArgs := evmtypes.EvmTxArgs{}
		txArgs.GasTipCap = new(big.Int).SetInt64(0)
		txArgs.GasLimit = 1_000_000_000_000
		revertContractAddr, err = is.factory.DeployContract(
			txSender.Priv,
			txArgs,
			factory.ContractDeploymentData{
				Contract: revertCallerContract,
				ConstructorArgs: []interface{}{
					common.HexToAddress(is.precompileAddrHex),
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy werc20 reverter contract")
		Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

		// Support struct used to simplify transactions creation.
		callsData = CallsData{
			sender: txSender,

			precompileAddr: precompileAddr,
			precompileABI:  precompile.ABI,

			precompileReverterAddr: revertContractAddr,
			precompileReverterABI:  revertCallerContract.ABI,
		}

		// Utility types used to check the different events emitted.
		failCheck = testutil.LogCheckArgs{ABIEvents: is.precompile.Events}
		passCheck = failCheck.WithExpPass(true)
		withdrawCheck = passCheck.WithExpEvents(werc20.EventTypeWithdrawal)
		depositCheck = passCheck.WithExpEvents(werc20.EventTypeDeposit)
		transferCheck = passCheck.WithExpEvents(erc20.EventTypeTransfer)
	})
	Context("calling a specific wrapped coin method", func() {
		Context("and funds are part of the transaction", func() {
			When("the method is deposit", func() {
				It("it should return funds to sender and emit the event", func() {
					// Store initial balance to verify that sender
					// balance remains the same after the contract call.
					initBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, werc20.DepositMethod)
					txArgs.Amount = depositAmount

					_, _, err := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, depositCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected error calling the precompile")
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					finalBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(finalBalance.String()).To(Equal(initBalance.String()))
				})
				It("it should consume at least the deposit requested gas", func() {
					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, werc20.DepositMethod)
					txArgs.Amount = depositAmount

					_, ethRes, _ := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, depositCheck)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					Expect(ethRes.GasUsed).To(BeNumerically(">=", werc20.DepositRequiredGas), "expected different gas used for deposit")
				})
			})
			//nolint:dupl
			When("no calldata is provided", func() {
				It("it should call the receive which behave like deposit", func() {
					initBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
					txArgs.Amount = depositAmount

					_, _, err := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, depositCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected error calling the precompile")
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					finalBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(finalBalance).To(Equal(initBalance))
				})
				It("it should consume at least the deposit requested gas", func() {
					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, werc20.DepositMethod)
					txArgs.Amount = depositAmount

					_, ethRes, _ := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, depositCheck)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					Expect(ethRes.GasUsed).To(BeNumerically(">=", werc20.DepositRequiredGas), "expected different gas used for receive")
				})
			})
			When("the specified method is too short", func() {
				It("it should call the fallback which behave like deposit", func() {
					initBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
					txArgs.Amount = depositAmount
					// Short method is directly set in the input to skip ABI validation
					txArgs.Input = []byte{1, 2, 3}

					_, _, err := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, depositCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected error calling the precompile")
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					finalBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(finalBalance).To(Equal(initBalance))
				})
				It("it should consume at least the deposit requested gas", func() {
					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
					txArgs.Amount = depositAmount
					// Short method is directly set in the input to skip ABI validation
					txArgs.Input = []byte{1, 2, 3}

					_, ethRes, _ := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, depositCheck)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					Expect(ethRes.GasUsed).To(BeNumerically(">=", werc20.DepositRequiredGas), "expected different gas used for fallback")
				})
			})
			When("the specified method does not exist", func() {
				It("it should call the fallback which behave like deposit", func() {
					initBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
					txArgs.Amount = depositAmount
					// Wrong method is directly set in the input to skip ABI validation
					txArgs.Input = []byte("nonExistingMethod")

					_, _, err := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, depositCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected error calling the precompile")
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					finalBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(finalBalance).To(Equal(initBalance))
				})
				It("it should consume at least the deposit requested gas", func() {
					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
					txArgs.Amount = depositAmount
					// Wrong method is directly set in the input to skip ABI validation
					txArgs.Input = []byte("nonExistingMethod")

					_, ethRes, _ := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, depositCheck)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					Expect(ethRes.GasUsed).To(BeNumerically(">=", werc20.DepositRequiredGas), "expected different gas used for fallback")
				})
			})
		})
		Context("and funds are NOT part of the transaction", func() {
			When("the method is withdraw", func() {
				It("it should fail if user doesn't have enough funds", func() {
					// Store initial balance to verify withdraw is a no-op and sender
					// balance remains the same after the contract call.
					initBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					newUserAcc, newUserPriv := utiltx.NewAccAddressAndKey()
					newUserBalance := sdk.Coins{sdk.Coin{
						Denom:  evmtypes.GetEVMCoinDenom(),
						Amount: math.NewIntFromBigInt(withdrawAmount).SubRaw(1),
					}}
					err := is.network.App.BankKeeper.SendCoins(is.network.GetContext(), user.AccAddr, newUserAcc, newUserBalance)
					Expect(err).ToNot(HaveOccurred(), "expected no error sending tokens")
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, werc20.WithdrawMethod, withdrawAmount)

					_, _, err = is.factory.CallContractAndCheckLogs(newUserPriv, txArgs, callArgs, withdrawCheck)
					Expect(err).To(HaveOccurred(), "expected an error because not enough funds")
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					finalBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(finalBalance).To(Equal(initBalance))
				})
				It("it should be a no-op and emit the event", func() {
					initBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, werc20.WithdrawMethod, withdrawAmount)

					_, _, err := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, withdrawCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected error calling the precompile")
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					finalBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(finalBalance).To(Equal(initBalance))
				})
				It("it should consume at least the withdraw requested gas", func() {
					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, werc20.WithdrawMethod, withdrawAmount)

					_, ethRes, _ := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, withdrawCheck)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					Expect(ethRes.GasUsed).To(BeNumerically(">=", werc20.WithdrawRequiredGas), "expected different gas used for withdraw")
				})
			})
			//nolint:dupl
			When("no calldata is provided", func() {
				It("it should call the fallback which behave like deposit", func() {
					initBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
					txArgs.Amount = depositAmount

					_, _, err := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, depositCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected error calling the precompile")
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					finalBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(finalBalance).To(Equal(initBalance))
				})
				It("it should consume at least the deposit requested gas", func() {
					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, werc20.DepositMethod)
					txArgs.Amount = depositAmount

					_, ethRes, _ := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, depositCheck)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					Expect(ethRes.GasUsed).To(BeNumerically(">=", werc20.DepositRequiredGas), "expected different gas used for receive")
				})
			})
			When("the specified method is too short", func() {
				It("it should call the fallback which behave like deposit", func() {
					initBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
					txArgs.Amount = depositAmount
					// Short method is directly set in the input to skip ABI validation
					txArgs.Input = []byte{1, 2, 3}

					_, _, err := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, depositCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected error calling the precompile")
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					finalBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(finalBalance).To(Equal(initBalance))
				})
				It("it should consume at least the deposit requested gas", func() {
					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
					txArgs.Amount = depositAmount
					// Short method is directly set in the input to skip ABI validation
					txArgs.Input = []byte{1, 2, 3}

					_, ethRes, _ := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, depositCheck)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					Expect(ethRes.GasUsed).To(BeNumerically(">=", werc20.DepositRequiredGas), "expected different gas used for fallback")
				})
			})
			When("the specified method does not exist", func() {
				It("it should call the fallback which behave like deposit", func() {
					initBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
					txArgs.Amount = depositAmount
					// Wrong method is directly set in the input to skip ABI validation
					txArgs.Input = []byte("nonExistingMethod")

					_, _, err := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, depositCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected error calling the precompile")
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					finalBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
					Expect(finalBalance).To(Equal(initBalance))
				})
				It("it should consume at least the deposit requested gas", func() {
					txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
					txArgs.Amount = depositAmount
					// Wrong method is directly set in the input to skip ABI validation
					txArgs.Input = []byte("nonExistingMethod")

					_, ethRes, _ := is.factory.CallContractAndCheckLogs(user.Priv, txArgs, callArgs, depositCheck)
					Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

					Expect(ethRes.GasUsed).To(BeNumerically(">=", werc20.DepositRequiredGas), "expected different gas used for fallback")
				})
			})
		})
	})
	Context("calling a reverter contract", func() {
		When("to call the deposit", func() {
			It("it should return funds to the last sender and emit the event", func() {
				ctx := is.network.GetContext()

				txArgs, callArgs := callsData.getTxAndCallArgs(contractCall, "depositWithRevert", false, false)
				txArgs.Amount = depositAmount

				_, _, err := is.factory.CallContractAndCheckLogs(txSender.Priv, txArgs, callArgs, depositCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected error calling the precompile")
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				finalBalance := is.network.App.BankKeeper.GetAllBalances(ctx, revertContractAddr.Bytes())
				Expect(finalBalance.AmountOf(evmtypes.GetEVMCoinDenom()).String()).To(Equal(depositAmount.String()), "expected final balance equal to deposit")
			})
		})
		DescribeTable("to call the deposit", func(before, after bool) {
			ctx := is.network.GetContext()

			initBalance := is.network.App.BankKeeper.GetAllBalances(ctx, txSender.AccAddr)

			txArgs, callArgs := callsData.getTxAndCallArgs(contractCall, "depositWithRevert", before, after)
			txArgs.Amount = depositAmount

			_, _, err := is.factory.CallContractAndCheckLogs(txSender.Priv, txArgs, callArgs, depositCheck)
			Expect(err).To(HaveOccurred(), "execution should have reverted")
			Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

			finalBalance := is.network.App.BankKeeper.GetAllBalances(ctx, txSender.AccAddr)
			Expect(finalBalance.String()).To(Equal(initBalance.String()), "expected final balance equal to initial")
		},
			Entry("it should not move funds and dont emit the event reverting before changing state", true, false),
			Entry("it should not move funds and dont emit the event reverting after changing state", false, true),
		)
	})
	Context("calling an erc20 method", func() {
		When("transferring tokens", func() {
			It("it should transfer tokens to a receiver using `transfer`", func() {
				ctx := is.network.GetContext()

				senderBalance := is.network.App.BankKeeper.GetAllBalances(ctx, txSender.AccAddr)
				receiverBalance := is.network.App.BankKeeper.GetAllBalances(ctx, user.AccAddr)

				txArgs, transferArgs := callsData.getTxAndCallArgs(directCall, erc20.TransferMethod, user.Addr, transferAmount)
				transferCoins := sdk.Coins{sdk.NewInt64Coin(is.wrappedCoinDenom, transferAmount.Int64())}

				_, _, err := is.factory.CallContractAndCheckLogs(txSender.Priv, txArgs, transferArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				senderBalanceAfter := is.network.App.BankKeeper.GetAllBalances(ctx, txSender.AccAddr)
				receiverBalanceAfter := is.network.App.BankKeeper.GetAllBalances(ctx, user.AccAddr)
				Expect(senderBalanceAfter).To(Equal(senderBalance.Sub(transferCoins...)))
				Expect(receiverBalanceAfter).To(Equal(receiverBalance.Add(transferCoins...)))
			})
			It("it should transfer tokens to a receiver using `transferFrom`", func() {
				ctx := is.network.GetContext()

				senderBalance := is.network.App.BankKeeper.GetAllBalances(ctx, txSender.AccAddr)
				receiverBalance := is.network.App.BankKeeper.GetAllBalances(ctx, user.AccAddr)

				txArgs, transferArgs := callsData.getTxAndCallArgs(directCall, erc20.TransferFromMethod, txSender.Addr, user.Addr, transferAmount)
				transferCoins := sdk.Coins{sdk.NewInt64Coin(is.wrappedCoinDenom, transferAmount.Int64())}

				transferCheck := passCheck.WithExpEvents(erc20.EventTypeTransfer, auth.EventTypeApproval)
				_, _, err := is.factory.CallContractAndCheckLogs(txSender.Priv, txArgs, transferArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				senderBalanceAfter := is.network.App.BankKeeper.GetAllBalances(ctx, txSender.AccAddr)
				receiverBalanceAfter := is.network.App.BankKeeper.GetAllBalances(ctx, user.AccAddr)
				Expect(senderBalanceAfter).To(Equal(senderBalance.Sub(transferCoins...)))
				Expect(receiverBalanceAfter).To(Equal(receiverBalance.Add(transferCoins...)))
			})
		})
		When("querying information", func() {
			Context("to retrieve a balance", func() {
				It("should return the correct balance for an existing account", func() {
					// Query the balance
					txArgs, balancesArgs := callsData.getTxAndCallArgs(directCall, erc20.BalanceOfMethod, txSender.Addr)

					_, ethRes, err := is.factory.CallContractAndCheckLogs(txSender.Priv, txArgs, balancesArgs, passCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					expBalance := is.network.App.BankKeeper.GetBalance(is.network.GetContext(), txSender.AccAddr, is.wrappedCoinDenom)

					var balance *big.Int
					err = is.precompile.UnpackIntoInterface(&balance, erc20.BalanceOfMethod, ethRes.Ret)
					Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
					Expect(balance).To(Equal(expBalance.Amount.BigInt()), "expected different balance")
				})
				It("should return 0 for a new account", func() {
					// Query the balance
					txArgs, balancesArgs := callsData.getTxAndCallArgs(directCall, erc20.BalanceOfMethod, utiltx.GenerateAddress())

					_, ethRes, err := is.factory.CallContractAndCheckLogs(txSender.Priv, txArgs, balancesArgs, passCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					var balance *big.Int
					err = is.precompile.UnpackIntoInterface(&balance, erc20.BalanceOfMethod, ethRes.Ret)
					Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
					Expect(balance.Int64()).To(Equal(int64(0)), "expected different balance")
				})
			})
			It("should return the correct name", func() {
				txArgs, nameArgs := callsData.getTxAndCallArgs(directCall, erc20.NameMethod)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(txSender.Priv, txArgs, nameArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var name string
				err = is.precompile.UnpackIntoInterface(&name, erc20.NameMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(name).To(ContainSubstring("Evmos"), "expected different name")
			})

			It("should return the correct symbol", func() {
				txArgs, symbolArgs := callsData.getTxAndCallArgs(directCall, erc20.SymbolMethod)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(txSender.Priv, txArgs, symbolArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var symbol string
				err = is.precompile.UnpackIntoInterface(&symbol, erc20.SymbolMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(symbol).To(ContainSubstring("EVMOS"), "expected different symbol")
			})

			It("should return the decimals", func() {
				txArgs, decimalsArgs := callsData.getTxAndCallArgs(directCall, erc20.DecimalsMethod)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(txSender.Priv, txArgs, decimalsArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var decimals uint8
				err = is.precompile.UnpackIntoInterface(&decimals, erc20.DecimalsMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")

				chainID := strings.Split(is.network.GetChainID(), "-")[0]
				coinInfo := evmtypes.ChainsCoinInfo[chainID]
				Expect(decimals).To(Equal(uint8(coinInfo.Decimals)), "expected different decimals")
			},
			)
		})
	})
})
