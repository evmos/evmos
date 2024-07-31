// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package vesting_test

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/precompiles/authorization"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/testutil"
	"github.com/evmos/evmos/v18/precompiles/vesting"
	"github.com/evmos/evmos/v18/precompiles/vesting/testdata"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"

	// testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	testutils "github.com/evmos/evmos/v18/testutil/integration/evmos/utils"

	testutiltx "github.com/evmos/evmos/v18/testutil/tx"
	// "github.com/evmos/evmos/v18/utils"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
	vestingtypes "github.com/evmos/evmos/v18/x/vesting/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

var (
	// vestingCaller is the compiled contract of the smart contract that calls the vesting precompile.
	vestingCaller evmtypes.CompiledContract
	// vestingCallerAddr is the address of the smart contract that calls the vesting precompile.
	vestingCallerAddr common.Address

	// defaultPeriods is a slice of default periods used in testing
	defaultPeriods []vesting.Period
	// instantPeriods is a slice of instant periods used in testing (i.e. length = 0)
	instantPeriods []vesting.Period
	// doublePeriods is a slice of two default periods used in testing
	doublePeriods []vesting.Period
	// emptyPeriods is a empty slice of periods used in testing
	emptyPeriods []vesting.Period

	defaultFundingAmount int64

	// err is a basic error type
	err error

	// execRevertedCheck is a basic check for contract calls to the precompile, where only "execution reverted" is returned
	execRevertedCheck testutil.LogCheckArgs
	// passCheck is a basic check that is used to check if the transaction was successful
	passCheck testutil.LogCheckArgs
	// failCheck is the default setting to check execution logs for failing transactions
	failCheck testutil.LogCheckArgs

	// callTypes is a slice of testing configurations used to run the test cases for direct
	// contract calls as well as calls through a smart contract.
	callTypes = []CallType{
		{name: "directly", directCall: true},
		{name: "through a smart contract", directCall: false},
	}
	// // differentAddr is a new address used in testing
	// differentAddr = testutiltx.GenerateAddress()
	// // vestingAddr is a new address that is used to test the vesting extension.
	// vestingAddr = testutiltx.GenerateAddress()
	// gasPrice to be used on tests txs and calculate the fees.
	gasPrice = math.NewInt(1e9)

	// funderKey is the key used to represent the funder of the vesting account.
	funderKey keyring.Key
	// vestingAccKey is the key used to represent the vesting account.
	vestingAccKey keyring.Key
)

func TestPrecompileIntegrationTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Precompile Test Suite")
}

var _ = Describe("Interacting with the vesting extension", Ordered, func() {
	var s *PrecompileTestSuite

	BeforeAll(func() {
		vestingCaller, err = testdata.LoadVestingCallerContract()
		Expect(err).ToNot(HaveOccurred(), "error while getting vestingCallerContract: %v", err)
	})

	BeforeEach(func() {
		// Setup the test suite with 3 pre-funded accounts.
		s = new(PrecompileTestSuite)
		s.SetupTest(3)

		// Set the default value for the vesting or lockup periods
		defaultFundingAmount = 100
		defaultPeriod := vesting.Period{
			Length: 10,
			Amount: []cmn.Coin{{Denom: s.bondDenom, Amount: big.NewInt(defaultFundingAmount)}},
		}
		instantPeriod := defaultPeriod
		instantPeriod.Length = 0
		defaultPeriods = []vesting.Period{defaultPeriod}
		doublePeriods = []vesting.Period{defaultPeriod, defaultPeriod}
		instantPeriods = []vesting.Period{instantPeriod}

		funderKey = s.keyring.GetKey(0)
		vestingAccKey = s.keyring.GetKey(1)

		// Deploy the smart contract that calls the vesting precompile.
		vestingCallerAddr, err = s.factory.DeployContract(
			funderKey.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: vestingCaller,
			},
		)
		Expect(err).ToNot(HaveOccurred(), "error while deploying the vesting caller smart contract: %v", err)
		Expect(s.network.NextBlock()).To(BeNil())

		// Set up the checks
		failCheck = testutil.LogCheckArgs{
			ABIEvents: s.precompile.Events,
			ExpPass:   false,
		}
		passCheck = failCheck.WithExpPass(true)
		execRevertedCheck = failCheck.WithErrContains("execution reverted")
	})

	// ---------------------------------------------
	//                   TRANSACTIONS
	//

	Context("to create a clawback vesting account", func() {
		for _, callType := range callTypes {
			callType := callType

			It(fmt.Sprintf("should succeed (%s)", callType.name), func() {
				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				callArgs.MethodName = vesting.CreateClawbackVestingAccountMethod
				callArgs.Args = []interface{}{
					funderKey.Addr,
					vestingAccKey.Addr,
					false,
				}

				createClawbackCheck := passCheck.WithExpEvents(vesting.EventTypeCreateClawbackVestingAccount)

				_, _, err = s.factory.CallContractAndCheckLogs(vestingAccKey.Priv, txArgs, callArgs, createClawbackCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				// Check the vesting account
				s.ExpectSimpleVestingAccount(vestingAccKey.Addr, funderKey.Addr)
			})

			It(fmt.Sprintf("should fail if the account is not initialized (%s)", callType.name), func() {
				nonExistentAddr, nonExistentPriv := testutiltx.NewAddrKey()

				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				callArgs.MethodName = vesting.CreateClawbackVestingAccountMethod
				callArgs.Args = []interface{}{
					funderKey.Addr,
					nonExistentAddr,
					false,
				}

				createClawbackCheck := failCheck.WithErrContains("account is not initialized")

				_, _, err = s.factory.CallContractAndCheckLogs(nonExistentPriv, txArgs, callArgs, createClawbackCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), nonExistentAddr.Bytes())
				Expect(acc).To(BeNil(), "account should not be created")
			})

			It(fmt.Sprintf("should fail if the vesting account is the zero address (%s)", callType.name), func() {
				sender := s.keyring.GetKey(2)

				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				callArgs.MethodName = vesting.CreateClawbackVestingAccountMethod
				callArgs.Args = []interface{}{
					funderKey.Addr,
					common.Address{},
					false,
				}

				createClawbackCheck := failCheck.WithErrContains("invalid address")

				if !callType.directCall {
					createClawbackCheck = failCheck.WithErrContains("execution reverted")
				}

				_, _, err = s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, callArgs, createClawbackCheck)
				Expect(err).To(BeNil(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should fail if the funder account is the zero address (%s)", callType.name), func() {
				funderAddr := common.Address{}

				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				callArgs.MethodName = vesting.CreateClawbackVestingAccountMethod
				callArgs.Args = []interface{}{
					funderAddr,
					vestingAccKey.Addr,
					false,
				}

				createClawbackCheck := failCheck.WithErrContains("invalid address")

				if !callType.directCall {
					createClawbackCheck = failCheck.WithErrContains("execution reverted")
				}

				_, _, err = s.factory.CallContractAndCheckLogs(vestingAccKey.Priv, txArgs, callArgs, createClawbackCheck)
				Expect(err).To(BeNil(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should fail if the origin is different than the vesting address (%s)", callType.name), func() {
				differentSender := s.keyring.GetKey(2)

				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				callArgs.MethodName = vesting.CreateClawbackVestingAccountMethod
				callArgs.Args = []interface{}{
					funderKey.Addr,
					vestingAccKey.Addr,
					false,
				}

				createClawbackCheck := failCheck.WithErrContains("origin is different than the vesting address")

				_, _, err = s.factory.CallContractAndCheckLogs(differentSender.Priv, txArgs, callArgs, createClawbackCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("does not match the from address"))
				}
			})

			It(fmt.Sprintf("should fail for a smart contract (%s)", callType.name), func() {
				if callType.directCall {
					Skip("this should only be run for smart contract calls")
				}

				// The vesting caller will try to convert its account into a
				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				callArgs.MethodName = "createClawbackVestingAccountForContract"

				_, _, err = s.factory.CallContractAndCheckLogs(funderKey.Priv, txArgs, callArgs, failCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(err.Error()).To(ContainSubstring("execution reverted"))
				Expect(s.network.NextBlock()).To(BeNil())

				// Check that the smart contract was not converted
				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingCallerAddr.Bytes())
				Expect(acc).ToNot(BeNil(), "smart contract should be found")
				_, ok := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(ok).To(BeFalse(), "smart contract should not be converted to a vesting account")
			})

			It(fmt.Sprintf("should fail if the account is already subjected to vesting (%s)", callType.name), func() {
				// Create a clawaback vesting account associated with the vestinfAccKey.
				err = s.factory.CreateClawbackVestingAccount(vestingAccKey.Priv, funderKey.AccAddr, false)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				callArgs.MethodName = vesting.CreateClawbackVestingAccountMethod
				callArgs.Args = []interface{}{
					funderKey.Addr,
					vestingAccKey.Addr,
					false,
				}

				createClawbackCheck := failCheck.WithErrContains("account is already subject to vesting")

				_, _, err = s.factory.CallContractAndCheckLogs(vestingAccKey.Priv, txArgs, callArgs, createClawbackCheck)
				Expect(s.network.NextBlock()).To(BeNil())
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("%s is already a clawback vesting account", vestingAccKey.AccAddr)))
				}
			})
		}
	})

	Context("to fund a clawback vesting account", func() {
		for _, callType := range callTypes {
			callType := callType

			BeforeEach(func() {
				if callType.directCall == false {
					approvalCallArgs := factory.CallArgs{
						ContractABI: s.precompile.ABI,
						MethodName:  "approve",
						Args: []interface{}{
							vestingCallerAddr,
							vesting.FundVestingAccountMsgURL,
						},
					}

					precompileAddr := s.precompile.Address()
					logCheck := passCheck.WithExpEvents(authorization.EventTypeApproval)

					_, _, err := s.factory.CallContractAndCheckLogs(funderKey.Priv, evmtypes.EvmTxArgs{To: &precompileAddr}, approvalCallArgs, logCheck)
					Expect(err).To(BeNil())
					Expect(s.network.NextBlock()).To(BeNil())

					auths, err := s.grpcHandler.GetAuthorizations(sdk.AccAddress(vestingCallerAddr.Bytes()).String(), funderKey.AccAddr.String())
					Expect(err).To(BeNil())
					Expect(auths).To(HaveLen(1))
				}
			})

			It(fmt.Sprintf("should succeed when defining only lockup and funder is an EOA (%s)", callType.name), func() { //nolint:dupl

				err = s.factory.CreateClawbackVestingAccount(vestingAccKey.Priv, funderKey.AccAddr, false)
				Expect(s.network.NextBlock()).To(BeNil())

				// Query balances before precompile call to compare final balances.
				initialBalances := s.GetBondBalances(vestingAccKey.AccAddr, funderKey.AccAddr, sdk.AccAddress(vestingCallerAddr.Bytes()))
				vestAccInitialBal, funderInitialBal, vestingCallerInitialBal := initialBalances[0], initialBalances[1], initialBalances[2]

				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				txArgs.GasPrice = gasPrice.BigInt()
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					funderKey.Addr,
					vestingAccKey.Addr,
					uint64(time.Now().Unix()),
					defaultPeriods,
					emptyPeriods,
				}

				fundClawbackVestingCheck := passCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				res, _, err := s.factory.CallContractAndCheckLogs(funderKey.Priv, txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				fees := gasPrice.MulRaw(res.GasUsed)

				// Check the vesting account
				//
				// NOTE: The vesting account is created with the lockup periods only, since the vesting periods are empty.
				// The vesting periods are defaulted to instant vesting, i.e. period length = 0.
				s.ExpectVestingAccount(vestingAccKey.Addr, defaultPeriods, instantPeriods)

				// Query balances after precompile call.
				finalBalances := s.GetBondBalances(vestingAccKey.AccAddr, funderKey.AccAddr, sdk.AccAddress(vestingCallerAddr.Bytes()))
				vestAccFinalBal, funderFinalBal, vestingCallerFinalBal := finalBalances[0], finalBalances[1], finalBalances[2]

				// Check balances after precompile call.
				Expect(vestAccFinalBal).To(Equal(vestAccInitialBal.Add(math.NewInt(defaultFundingAmount))))
				Expect(funderFinalBal).To(Equal(funderInitialBal.Sub(math.NewInt(defaultFundingAmount)).Sub(fees)))
				Expect(vestingCallerFinalBal).To(Equal(vestingCallerInitialBal))
			})

			It(fmt.Sprintf("should succeed when defining only lockup and funder is a smart contract (%s)", callType.name), func() {
				if callType.directCall {
					Skip("this should only be run for smart contract calls")
				}

				// We are not using the funder key here to avoid confusion so we need another key.
				txSenderKey := s.keyring.GetKey(2)

				// Note that in this case the funder is the vesting caller contract.
				err = s.factory.CreateClawbackVestingAccount(vestingAccKey.Priv, sdk.AccAddress(vestingCallerAddr.Bytes()), false)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				// Send some funds to the smart contract to allow it to fund the account.
				// The authorization to be able to fund from the smart contract is already in the setup
				err := testutils.FundAccountWithBaseDenom(s.factory, s.network, txSenderKey, sdk.AccAddress(vestingCallerAddr.Bytes()), math.NewInt(defaultFundingAmount))
				Expect(err).To(BeNil(), "error while sending coins")
				Expect(s.network.NextBlock()).To(BeNil())

				// Query balances before precompile call to compare final balances.
				initialBalances := s.GetBondBalances(vestingAccKey.AccAddr, txSenderKey.AccAddr, sdk.AccAddress(vestingCallerAddr.Bytes()))
				vestAccInitialBal, txSenderInitialBal, vestingCallerInitialBal := initialBalances[0], initialBalances[1], initialBalances[2]

				// Build and execute the tx to fund the vesting account from a smart contract.
				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				txArgs.GasPrice = gasPrice.BigInt()
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					vestingCallerAddr,
					vestingAccKey.Addr,
					uint64(time.Now().Unix()),
					defaultPeriods,
					emptyPeriods,
				}

				fundClawbackVestingCheck := passCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				res, _, err := s.factory.CallContractAndCheckLogs(txSenderKey.Priv, txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				fees := gasPrice.MulRaw(res.GasUsed)

				// Check the vesting account
				//
				// NOTE: The vesting account is created with the lockup periods only, since the vesting periods are empty.
				// The vesting periods are defaulted to instant vesting, i.e. period length = 0.
				s.ExpectVestingAccount(vestingAccKey.Addr, defaultPeriods, instantPeriods)

				// Query balances after precompile call.
				finalBalances := s.GetBondBalances(vestingAccKey.AccAddr, txSenderKey.AccAddr, sdk.AccAddress(vestingCallerAddr.Bytes()))
				vestAccFinalBal, txSenderFinalBal, vestingCallerFinalBal := finalBalances[0], finalBalances[1], finalBalances[2]

				// Check balances after precompile call.
				Expect(vestAccFinalBal).To(Equal(vestAccInitialBal.Add(math.NewInt(defaultFundingAmount))), "expected vesting account to have initial balance plus vesting")
				Expect(txSenderFinalBal).To(Equal(txSenderInitialBal.Sub(fees)), "expected tx sender to have initial balance minus fees")
				Expect(vestingCallerFinalBal.Int64()).To(Equal(vestingCallerInitialBal.Sub(math.NewInt(defaultFundingAmount)).Int64()), "expected vesting caller to have initial balance minus vesting")
			})

			It(fmt.Sprintf("should succeed when defining only vesting and funder is an EOA (%s)", callType.name), func() { //nolint:dupl

				err = s.factory.CreateClawbackVestingAccount(vestingAccKey.Priv, funderKey.AccAddr, false)
				Expect(s.network.NextBlock()).To(BeNil())

				// Query balances before precompile call to compare final balances.
				initialBalances := s.GetBondBalances(vestingAccKey.AccAddr, funderKey.AccAddr, sdk.AccAddress(vestingCallerAddr.Bytes()))
				vestAccInitialBal, funderInitialBal, vestingCallerInitialBal := initialBalances[0], initialBalances[1], initialBalances[2]

				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				txArgs.GasPrice = gasPrice.BigInt()
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					funderKey.Addr,
					vestingAccKey.Addr,
					uint64(time.Now().Unix()),
					emptyPeriods,
					defaultPeriods,
				}

				fundClawbackVestingCheck := passCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				res, _, err := s.factory.CallContractAndCheckLogs(funderKey.Priv, txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				fees := gasPrice.MulRaw(res.GasUsed)

				// Check the vesting account
				//
				// NOTE: The vesting account is created with the vesting periods only, since the lockup periods are empty.
				// The lockup periods are defaulted to instant unlocking, i.e. period length = 0.
				s.ExpectVestingAccount(vestingAccKey.Addr, instantPeriods, defaultPeriods)

				// Query balances after precompile call.
				finalBalances := s.GetBondBalances(vestingAccKey.AccAddr, funderKey.AccAddr, sdk.AccAddress(vestingCallerAddr.Bytes()))
				vestAccFinalBal, funderFinalBal, vestingCallerFinalBal := finalBalances[0], finalBalances[1], finalBalances[2]

				// Check balances after precompile call.
				Expect(vestAccFinalBal).To(Equal(vestAccInitialBal.Add(math.NewInt(defaultFundingAmount))))
				Expect(funderFinalBal).To(Equal(funderInitialBal.Sub(math.NewInt(defaultFundingAmount)).Sub(fees)))
				Expect(vestingCallerFinalBal).To(Equal(vestingCallerInitialBal))
			})

			// TODO: not clear the scope of this test

			// It(fmt.Sprintf("should fail with a smart contract different than the contract that calls the precompile (%s)", callType.name), func() {
			// 	if callType.directCall {
			// 		Skip("this should only be run for smart contract calls")
			// 	}
			//
			// 	counterContract, err := contracts.LoadCounterContract()
			// 	Expect(err).ToNot(HaveOccurred(), "error while getting the counter contract: %v", err)
			//
			// 	contractCounterAddr, err := s.factory.DeployContract(
			// 		funderKey.Priv,
			// 		evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			// 		factory.ContractDeploymentData{
			// 			Contract: counterContract,
			// 		},
			// 	)
			// 	Expect(err).ToNot(HaveOccurred(), "error while deploying the counter smart contract: %v", err)
			// 	Expect(s.network.NextBlock()).To(BeNil())
			//
			// 	err = s.factory.CreateClawbackVestingAccount(vestingAccKey.Priv, funderKey.AccAddr, false)
			// 	Expect(err).To(BeNil())
			// 	Expect(s.network.NextBlock()).To(BeNil())
			//
			// 	// Send some funds to the smart contract to allow it to fund the account.
			// 	err = testutils.FundAccountWithBaseDenom(s.factory, s.network, funderKey, sdk.AccAddress(contractCounterAddr.Bytes()), math.NewInt(defaultFundingAmount))
			// 	Expect(err).To(BeNil(), "error while sending coins to the contract")
			// 	Expect(s.network.NextBlock()).To(BeNil())
			//
			// 	// Create authorization for tx sender to use funder's balance to fund a vesting account.
			// 	approvalCallArgs := factory.CallArgs{
			// 		ContractABI: s.precompile.ABI,
			// 		MethodName:  "approve",
			// 		Args: []interface{}{
			// 			contractCounterAddr,
			// 			vesting.FundVestingAccountMsgURL,
			// 		},
			// 	}
			//
			// 	precompileAddr := s.precompile.Address()
			// 	logCheck := passCheck.WithExpEvents(authorization.EventTypeApproval)
			//
			// 	_, _, err = s.factory.CallContractAndCheckLogs(funderKey.Priv, evmtypes.EvmTxArgs{To: &precompileAddr}, approvalCallArgs, logCheck)
			// 	Expect(err).To(BeNil(), "error while creating the generic authorization: %v", err)
			// 	Expect(s.network.NextBlock()).To(BeNil())
			//
			// 	auths, err := s.grpcHandler.GetAuthorizations(sdk.AccAddress(contractCounterAddr.Bytes()).String(), funderKey.AccAddr.String())
			// 	Expect(err).To(BeNil())
			// 	Expect(auths).To(HaveLen(1))
			//
			// 	// Query balances before precompile call.
			// 	funderInitialBal, err := s.grpcHandler.GetBalance(funderKey.AccAddr, s.bondDenom)
			// 	Expect(err).To(BeNil())
			//
			// 	// Build and execute the tx to fund the vesting account from a smart contract
			// 	// Build and execute the tx to fund the vesting account from a smart contract.
			// 	callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
			// 	txArgs.GasPrice = gasPrice.BigInt()
			// 	callArgs.MethodName = vesting.FundVestingAccountMethod
			// 	callArgs.Args = []interface{}{
			// 		contractCounterAddr,
			// 		vestingAccKey.Addr,
			// 		uint64(time.Now().Unix()),
			// 		defaultPeriods,
			// 		emptyPeriods,
			// 	}
			//
			// 	res, _, err := s.factory.CallContractAndCheckLogs(funderKey.Priv, txArgs, callArgs, execRevertedCheck)
			// 	Expect(err).To(HaveOccurred())
			//
			// 	fees := gasPrice.MulRaw(res.GasUsed)
			//
			// 	funderFinalBal, err := s.grpcHandler.GetBalance(funderKey.AccAddr, s.bondDenom)
			// 	Expect(err).To(BeNil())
			//
			// 	Expect(funderFinalBal.Balance.Amount).To(Equal(funderInitialBal.Balance.Amount.Sub(fees)), "expected funder to have initial balance minus fees")
			// 	// // check that tx signer's balance is reduced by the fees paid
			// 	// txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
			// 	// Expect(txSenderFinalBal.Amount.LTE(txSenderInitialBal.Amount)).To(BeTrue())
			// 	//
			// 	// // the balance of the contract that calls the precompile should remain 0
			// 	// contractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
			// 	// Expect(contractFinalBal.Amount).To(Equal(sdk.ZeroInt()))
			// 	//
			// 	// // the balance of the funder contract should remain unchanged
			// 	// funderContractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, funderContractAddr.Bytes(), s.bondDenom)
			// 	// Expect(funderContractFinalBal.Amount).To(Equal(funderContractInitialAmt))
			// })
			//
			It(fmt.Sprintf("should fail using a third party EOA even if authorized by funder (%s)", callType.name), func() {
				if callType.directCall {
					Skip("this should only be run for smart contract calls")
				}

				txSenderKey := s.keyring.GetKey(2)

				// Create authorization for tx sender to use funder's balance to fund a vesting account.
				s.CreateFundVestingAccountAuthorization(funderKey, txSenderKey.Addr)

				// Create clawback vesting account. Not that the funder is not the transaction sender.
				err = s.factory.CreateClawbackVestingAccount(vestingAccKey.Priv, funderKey.AccAddr, false)

				// Query initialBalances before precompile call to compare final initialBalances.
				initialBalances := s.GetBondBalances(vestingAccKey.AccAddr, funderKey.AccAddr, txSenderKey.AccAddr)
				vestAccInitialBal, funderInitialBal, txSenderInitialBal := initialBalances[0], initialBalances[1], initialBalances[2]

				// Build and execute the tx to fund the vesting account.
				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				txArgs.GasPrice = gasPrice.BigInt()
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					funderKey.Addr,
					vestingAccKey.Addr,
					uint64(time.Now().Unix()),
					defaultPeriods,
					emptyPeriods,
				}

				fundClawbackVestingCheck := execRevertedCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				res, _, err := s.factory.CallContractAndCheckLogs(txSenderKey.Priv, txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "expected error in the fund vesting account execution")
				Expect(s.network.NextBlock()).To(BeNil())

				// NOTE: GasUsed is 0, is it normal?
				fees := gasPrice.MulRaw(res.GasUsed)

				// Query balances after precompile call.
				finalBalances := s.GetBondBalances(vestingAccKey.AccAddr, funderKey.AccAddr, txSenderKey.AccAddr)
				vestAccFinalBal, funderFinalBal, txSenderFinalBal := finalBalances[0], finalBalances[1], finalBalances[2]

				// Check balances after precompile call.
				Expect(vestAccFinalBal).To(Equal(vestAccInitialBal))
				Expect(funderFinalBal).To(Equal(funderInitialBal))
				Expect(txSenderFinalBal).To(Equal(txSenderInitialBal.Sub(fees))) // WROOONG
			})

			It(fmt.Sprintf("should fail when the contract calling the precompile is not authorized by the funder (%s)", callType.name), func() {
				if callType.directCall {
					Skip("this should only be run for smart contract calls")
				}

				// By default we are creating an authorization for the vesting caller from the funder key. For this
				// reason, we need to use another funder now.
				funderNoAuthKey := s.keyring.GetKey(2)
				err = s.factory.CreateClawbackVestingAccount(vestingAccKey.Priv, funderNoAuthKey.AccAddr, false)

				// Build and execute the tx to fund the vesting account.
				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				txArgs.GasPrice = gasPrice.BigInt()
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					funderNoAuthKey.Addr,
					vestingAccKey.Addr,
					uint64(time.Now().Unix()),
					emptyPeriods,
					defaultPeriods,
				}

				fundClawbackVestingCheck := execRevertedCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				_, _, err := s.factory.CallContractAndCheckLogs(funderNoAuthKey.Priv, txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "expected error in the fund vesting account execution")
				Expect(s.network.NextBlock()).To(BeNil())
			})

			It(fmt.Sprintf("should succeed when defining only vesting and funder is an EOA (%s)", callType.name), func() {
				err = s.factory.CreateClawbackVestingAccount(vestingAccKey.Priv, funderKey.AccAddr, false)
				Expect(s.network.NextBlock()).To(BeNil())

				// Query balances before precompile call to compare final balances.
				initialBalances := s.GetBondBalances(vestingAccKey.AccAddr, funderKey.AccAddr, sdk.AccAddress(vestingCallerAddr.Bytes()))
				vestAccInitialBal, funderInitialBal, vestingCallerInitialBal := initialBalances[0], initialBalances[1], initialBalances[2]

				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				txArgs.GasPrice = gasPrice.BigInt()
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					funderKey.Addr,
					vestingAccKey.Addr,
					uint64(time.Now().Unix()),
					emptyPeriods,
					defaultPeriods,
				}

				fundClawbackVestingCheck := passCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				res, _, err := s.factory.CallContractAndCheckLogs(funderKey.Priv, txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				fees := gasPrice.MulRaw(res.GasUsed)

				// Check the vesting account
				//
				// NOTE: The vesting account is created with the lockup periods only, since the vesting periods are empty.
				// The vesting periods are defaulted to instant vesting, i.e. period length = 0.
				s.ExpectVestingAccount(vestingAccKey.Addr, instantPeriods, defaultPeriods)

				// Query balances after precompile call.
				finalBalances := s.GetBondBalances(vestingAccKey.AccAddr, funderKey.AccAddr, sdk.AccAddress(vestingCallerAddr.Bytes()))
				vestAccFinalBal, funderFinalBal, vestingCallerFinalBal := finalBalances[0], finalBalances[1], finalBalances[2]

				// Check balances after precompile call.
				Expect(vestAccFinalBal).To(Equal(vestAccInitialBal.Add(math.NewInt(defaultFundingAmount))))
				Expect(funderFinalBal).To(Equal(funderInitialBal.Sub(math.NewInt(defaultFundingAmount)).Sub(fees)))
				Expect(vestingCallerFinalBal).To(Equal(vestingCallerInitialBal))
			})

			Context("using the vesting caller `fundVestingAccountAndTransfer` function", func() {
				// testCase is a struct used for cases of contracts calls that have some operation
				// performed before and/or after the precompile call.
				type testCase struct {
					transferTo *common.Address
					before     bool
					after      bool
				}

				var (
					callArgs                    factory.CallArgs
					txArgs                      evmtypes.EvmTxArgs
					funderInitialBal            math.Int
					vestingAccInitialBal        math.Int
					vestingCallerInitialBal     = math.NewInt(100)
					vestingCallerTransferAmount = int64(15)
				)

				BeforeEach(func() {
					callArgs, txArgs = s.BuildCallArgs(callType, vestingCallerAddr)
					txArgs.GasPrice = gasPrice.BigInt()
					callArgs.MethodName = "fundVestingAccountAndTransfer"

					// Send some funds to the contract calling into the vesting precompile to allow
					// it to send funds before and/or after calling the precompile.
					err := testutils.FundAccountWithBaseDenom(s.factory, s.network, funderKey, sdk.AccAddress(vestingCallerAddr.Bytes()), math.NewInt(vestingCallerInitialBal.Int64()))
					Expect(err).To(BeNil(), "error while sending coins")
					Expect(s.network.NextBlock()).To(BeNil())

					err = s.factory.CreateClawbackVestingAccount(vestingAccKey.Priv, funderKey.AccAddr, false)
					Expect(err).To(BeNil(), "error while creating clawback vesting account")
					Expect(s.network.NextBlock()).To(BeNil())

					// Query balances before precompile call to compare final balances.
					initialBalances := s.GetBondBalances(vestingAccKey.AccAddr, funderKey.AccAddr)
					vestingAccInitialBal, funderInitialBal = initialBalances[0], initialBalances[1]
				})

				DescribeTable("should fund the vesting account from tx origin when defining only vesting and ", func(tc testCase) {
					if callType.directCall {
						Skip("this should only be run for smart contract calls")
					}

					// If not specified, default the transferTo to the funder address.
					if tc.transferTo == nil {
						tc.transferTo = &funderKey.Addr
					}

					callArgs.Args = []interface{}{
						funderKey.Addr,
						vestingAccKey.Addr,
						*tc.transferTo,
						uint64(time.Now().Unix()),
						emptyPeriods,
						defaultPeriods,
						tc.before, tc.after, // transfer funds to the funder according to test case
					}

					fundClawbackVestingCheck := passCheck.
						WithExpEvents(vesting.EventTypeFundVestingAccount)

					res, _, err := s.factory.CallContractAndCheckLogs(funderKey.Priv, txArgs, callArgs, fundClawbackVestingCheck)
					Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
					Expect(s.network.NextBlock()).To(BeNil())

					fees := gasPrice.MulRaw(res.GasUsed)

					// The vesting caller can transfer funds before and after calling the precompile
					// depending on the call arguments. For this reason, we have to compute the total
					// amount sent.
					transferredToAmt := math.ZeroInt()
					for _, transferred := range []bool{tc.before, tc.after} {
						if transferred {
							transferredToAmt = transferredToAmt.AddRaw(vestingCallerTransferAmount)
						}
					}

					// Check the vesting account
					//
					// NOTE: The vesting account is created with the vesting periods only, since the lockup periods are empty.
					// The lockup periods are defaulted to instant unlocking, i.e. period length = 0.
					s.ExpectVestingAccount(vestingAccKey.Addr, instantPeriods, defaultPeriods)

					// Query balances after precompile call.
					finalBalances := s.GetBondBalances(vestingAccKey.AccAddr, funderKey.AccAddr, sdk.AccAddress(vestingCallerAddr.Bytes()))
					vestAccFinalBal, funderFinalBal, vestingCallerFinalBal := finalBalances[0], finalBalances[1], finalBalances[2]

					// Check the vesting caller's balance was deducted by the funds sent before and after calling
					// the precompile.
					Expect(vestingCallerFinalBal).To(Equal(vestingCallerInitialBal.Sub(transferredToAmt)), "expected the vesting caller to have initial balance minus transferred amount")

					// Check that the vesting account received the funds.
					expVestAccFinalBal := vestingAccInitialBal.Add(math.NewInt(defaultFundingAmount))
					if *tc.transferTo == vestingAccKey.Addr {
						expVestAccFinalBal = expVestAccFinalBal.Add(transferredToAmt)
					}

					Expect(vestAccFinalBal).To(Equal(expVestAccFinalBal), "expected the vesting account to have received the vesting plus any transfer")

					// check that funder balance is reduced by the fees paid, the amt to fund the vesting account,
					// but also got the funds sent from the contract (when corresponds)
					expFunderFinalBal := funderInitialBal.Sub(fees).Sub(math.NewInt(defaultFundingAmount))
					if *tc.transferTo == funderKey.Addr {
						expFunderFinalBal = expFunderFinalBal.Add(transferredToAmt)
					}
					Expect(funderFinalBal).To(Equal(expFunderFinalBal), "expected funder to have initial balance minus fee and vesting plus any transfer received")
				},
					Entry(fmt.Sprintf("funder balance change before & after precompile call (%s)", callType.name), testCase{
						before: true,
						after:  true,
					}),
					Entry(fmt.Sprintf("funder balance change before precompile call (%s)", callType.name), testCase{
						before: true,
						after:  false,
					}),
					Entry(fmt.Sprintf("funder balance change after precompile call (%s)", callType.name), testCase{
						before: false,
						after:  true,
					}),
					Entry(fmt.Sprintf("vesting acc balance change before & after precompile call (%s)", callType.name), testCase{
						transferTo: &vestingAccKey.Addr,
						before:     true,
						after:      true,
					}),
					Entry(fmt.Sprintf("vesting acc balance change before precompile call (%s)", callType.name), testCase{
						transferTo: &vestingAccKey.Addr,
						before:     true,
						after:      false,
					}),
					Entry(fmt.Sprintf("vesting acc balance change after precompile call (%s)", callType.name), testCase{
						transferTo: &vestingAccKey.Addr,
						before:     false,
						after:      true,
					}),
				)
			})

			It(fmt.Sprintf("should succeed when defining only vesting and funder is a smart contract (%s)", callType.name), func() { //nolint:dupl
				if callType.directCall {
					Skip("this should only be run for smart contract calls")
				}

				// We are not using the funder key here to avoid confusion so we need another key.
				txSenderKey := s.keyring.GetKey(2)

				// Note that in this case the funder is the vesting caller contract.
				err = s.factory.CreateClawbackVestingAccount(vestingAccKey.Priv, sdk.AccAddress(vestingCallerAddr.Bytes()), false)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				// Send some funds to the smart contract to allow it to fund the account.
				// The authorization to be able to fund from the smart contract is already in the setup
				err := testutils.FundAccountWithBaseDenom(s.factory, s.network, txSenderKey, sdk.AccAddress(vestingCallerAddr.Bytes()), math.NewInt(defaultFundingAmount))
				Expect(err).To(BeNil(), "error while sending coins")
				Expect(s.network.NextBlock()).To(BeNil())

				// Query balances before precompile call to compare final balances.
				initialBalances := s.GetBondBalances(vestingAccKey.AccAddr, txSenderKey.AccAddr, sdk.AccAddress(vestingCallerAddr.Bytes()))
				vestAccInitialBal, txSenderInitialBal, vestingCallerInitialBal := initialBalances[0], initialBalances[1], initialBalances[2]

				// Build and execute the tx to fund the vesting account from a smart contract.
				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				txArgs.GasPrice = gasPrice.BigInt()
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					vestingCallerAddr,
					vestingAccKey.Addr,
					uint64(time.Now().Unix()),
					emptyPeriods,
					defaultPeriods,
				}

				fundClawbackVestingCheck := passCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				res, _, err := s.factory.CallContractAndCheckLogs(txSenderKey.Priv, txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				fees := gasPrice.MulRaw(res.GasUsed)

				// Check the vesting account.
				s.ExpectVestingAccount(vestingAccKey.Addr, instantPeriods, defaultPeriods)

				// Query balances after precompile call.
				finalBalances := s.GetBondBalances(vestingAccKey.AccAddr, txSenderKey.AccAddr, sdk.AccAddress(vestingCallerAddr.Bytes()))
				vestAccFinalBal, txSenderFinalBal, vestingCallerFinalBal := finalBalances[0], finalBalances[1], finalBalances[2]

				// Check balances after precompile call.
				Expect(vestAccFinalBal).To(Equal(vestAccInitialBal.Add(math.NewInt(defaultFundingAmount))), "expected vesting account to have initial balance plus vesting")
				Expect(txSenderFinalBal).To(Equal(txSenderInitialBal.Sub(fees)), "expected tx sender to have initial balance minus fees")
				Expect(vestingCallerFinalBal.Int64()).To(Equal(vestingCallerInitialBal.Sub(math.NewInt(defaultFundingAmount)).Int64()), "expected vesting caller to have initial balance minus vesting")
			})
			//
			// 	It(fmt.Sprintf("should fund the vesting when defining both lockup and vesting (%s)", callType.name), func() { //nolint:dupl
			// 		funder := s.keyring.GetKey(0)
			// 		vestingKey := s.keyring.GetKey(1)
			//
			// 		err = s.factory.CreateClawbackVestingAccount(vestingKey.Priv, funder.AccAddr, false)
			// 		Expect(s.network.NextBlock()).To(BeNil())
			//
			// 		callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
			// 		callArgs.MethodName = vesting.FundVestingAccountMethod
			// 		callArgs.Args = []interface{}{
			// 			funder.Addr,
			// 			vestingKey.Addr,
			// 			uint64(time.Now().Unix()),
			// 			defaultPeriods,
			// 			defaultPeriods,
			// 		}
			//
			// 		fundClawbackVestingCheck := passCheck.
			// 			WithExpEvents(vesting.EventTypeFundVestingAccount)
			//
			// 		_, _, err := s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, fundClawbackVestingCheck)
			// 		Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
			// 		Expect(s.network.NextBlock()).To(BeNil())
			//
			// 		// Check the vesting account
			// 		s.ExpectVestingAccount(vestingKey.Addr, defaultPeriods, defaultPeriods)
			// 	})
			//
			// 	It(fmt.Sprintf("should fund the vesting account from a smart contract when defining both lockup and vesting  (%s)", callType.name), func() { //nolint:dupl
			// 		if callType.directCall {
			// 			Skip("this should only be run for smart contract calls")
			// 		}
			// 		s.CreateTestClawbackVestingAccount(contractAddr, toAddr)
			// 		// send some funds to the smart contract
			// 		err = evmosutil.FundAccount(s.ctx, s.app.BankKeeper, contractAddr.Bytes(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(100))))
			// 		Expect(err).ToNot(HaveOccurred(), "error while funding the contract: %v", err)
			//
			// 		txSenderInitialBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
			//
			// 		// Build and execute the tx to fund the vesting account from a smart contract
			// 		createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
			// 			WithMethodName(vesting.FundVestingAccountMethod).
			// 			WithArgs(
			// 				contractAddr,
			// 				toAddr,
			// 				uint64(time.Now().Unix()),
			// 				defaultPeriods,
			// 				defaultPeriods,
			// 			).
			// 			WithGasPrice(gasPrice.BigInt())
			//
			// 		fundClawbackVestingCheck := passCheck.
			// 			WithExpEvents(vesting.EventTypeFundVestingAccount)
			//
			// 		res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
			// 		Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
			// 		fees := gasPrice.MulRaw(res.GasUsed)
			//
			// 		// Check the vesting account
			// 		s.ExpectVestingAccount(toAddr, defaultPeriods, defaultPeriods)
			//
			// 		// check that tx signer's balance is reduced by the fees paid
			// 		txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
			// 		Expect(txSenderFinalBal.Amount).To(Equal(txSenderInitialBal.Amount.Sub(fees)))
			//
			// 		// check the contract's balance was deducted to fund the vesting account
			// 		contractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
			// 		Expect(contractFinalBal.Amount).To(Equal(sdk.ZeroInt()))
			// 	})
			//
			// 	It(fmt.Sprintf("should not fund the vesting when defining different total coins for lockup and vesting (%s)", callType.name), func() {
			// 		s.CreateTestClawbackVestingAccount(s.address, toAddr)
			//
			// 		// Build and execute the tx to fund the vesting account from a smart contract
			// 		createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
			// 			WithMethodName(vesting.FundVestingAccountMethod).
			// 			WithArgs(
			// 				contractAddr,
			// 				toAddr,
			// 				uint64(time.Now().Unix()),
			// 				defaultPeriods,
			// 				defaultPeriods,
			// 			).
			// 			WithGasPrice(gasPrice.BigInt())
			//
			// 		fundClawbackVestingCheck := passCheck.
			// 			WithExpEvents(vesting.EventTypeFundVestingAccount)
			//
			// 		res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
			// 		Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
			// 		fees := gasPrice.MulRaw(res.GasUsed)
			//
			// 		// Check the vesting account
			// 		s.ExpectVestingAccount(toAddr, defaultPeriods, defaultPeriods)
			//
			// 		// check that tx signer's balance is reduced by the fees paid
			// 		txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
			// 		Expect(txSenderFinalBal.Amount).To(Equal(txSenderInitialBal.Amount.Sub(fees)))
			//
			// 		// check the contract's balance was deducted to fund the vesting account
			// 		contractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
			// 		Expect(contractFinalBal.Amount).To(Equal(sdk.ZeroInt()))
			// 	})
			//
			// 	It(fmt.Sprintf("should not fund the vesting when defining different total coins for lockup and vesting (%s)", callType.name), func() {
			// 		funder := s.keyring.GetKey(0)
			// 		vestingKey := s.keyring.GetKey(1)
			//
			// 		err = s.factory.CreateClawbackVestingAccount(vestingKey.Priv, funder.AccAddr, false)
			// 		Expect(s.network.NextBlock()).To(BeNil())
			//
			// 		callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
			// 		callArgs.MethodName = vesting.FundVestingAccountMethod
			// 		callArgs.Args = []interface{}{
			// 			funder.Addr,
			// 			vestingKey.Addr,
			// 			uint64(time.Now().Unix()),
			// 			defaultPeriods,
			// 			doublePeriods,
			// 		}
			//
			// 		fundClawbackVestingCheck := execRevertedCheck
			// 		if callType.directCall {
			// 			fundClawbackVestingCheck = failCheck.WithErrContains("vesting and lockup schedules must have same total coins")
			// 		}
			//
			// 		_, _, err := s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, fundClawbackVestingCheck)
			// 		Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
			// 		Expect(s.network.NextBlock()).To(BeNil())
			//
			// 		// Check the vesting account
			// 		acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingKey.AccAddr)
			// 		Expect(acc).ToNot(BeNil(), "account should exist")
			// 		vestAcc, ok := acc.(*vestingtypes.ClawbackVestingAccount)
			// 		Expect(ok).To(BeTrue())
			// 		Expect(vestAcc.LockupPeriods).To(BeEmpty())
			// 		Expect(vestAcc.VestingPeriods).To(BeEmpty())
			// 	})
			//
			// 	It(fmt.Sprintf("should not fund the vesting when defining neither lockup nor vesting (%s)", callType.name), func() {
			// 		funder := s.keyring.GetKey(0)
			// 		vestingKey := s.keyring.GetKey(1)
			//
			// 		err = s.factory.CreateClawbackVestingAccount(vestingKey.Priv, funder.AccAddr, false)
			// 		Expect(err).To(BeNil())
			// 		Expect(s.network.NextBlock()).To(BeNil())
			//
			// 		callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
			// 		callArgs.MethodName = vesting.FundVestingAccountMethod
			// 		callArgs.Args = []interface{}{
			// 			funder.Addr,
			// 			vestingKey.Addr,
			// 			uint64(time.Now().Unix()),
			// 			emptyPeriods,
			// 			emptyPeriods,
			// 		}
			//
			// 		fundClawbackVestingCheck := execRevertedCheck
			// 		if callType.directCall {
			// 			fundClawbackVestingCheck = failCheck.WithErrContains("vesting and/or lockup schedules must be present")
			// 		}
			//
			// 		_, _, err := s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, fundClawbackVestingCheck)
			// 		Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
			// 		Expect(s.network.NextBlock()).To(BeNil())
			//
			// 		// Check the vesting account
			// 		acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingKey.AccAddr)
			// 		Expect(acc).ToNot(BeNil(), "account should exist")
			// 		vestAcc, ok := acc.(*vestingtypes.ClawbackVestingAccount)
			// 		Expect(ok).To(BeTrue())
			// 		Expect(vestAcc.LockupPeriods).To(BeEmpty())
			// 		Expect(vestAcc.VestingPeriods).To(BeEmpty())
			// 	})
			//
			// 	It(fmt.Sprintf("should not fund the vesting when exceeding the funder balance (%s)", callType.name), func() {
			// 		funder := s.keyring.GetKey(0)
			// 		vestingKey := s.keyring.GetKey(1)
			//
			// 		err = s.factory.CreateClawbackVestingAccount(vestingKey.Priv, funder.AccAddr, false)
			// 		Expect(err).To(BeNil())
			// 		Expect(s.network.NextBlock()).To(BeNil())
			//
			// 		res, err := s.grpcHandler.GetBalance(funder.AccAddr, s.bondDenom)
			// 		Expect(err).To(BeNil())
			// 		balance := res.Balance
			// 		exceededBalance := new(big.Int).Add(big.NewInt(1), balance.Amount.BigInt())
			//
			// 		exceedingVesting := []vesting.Period{{
			// 			Length: 10,
			// 			Amount: []cmn.Coin{{Denom: s.bondDenom, Amount: exceededBalance}},
			// 		}}
			//
			// 		callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
			// 		callArgs.MethodName = vesting.FundVestingAccountMethod
			// 		callArgs.Args = []interface{}{
			// 			funder.Addr,
			// 			vestingKey.Addr,
			// 			uint64(time.Now().Unix()),
			// 			exceedingVesting,
			// 			emptyPeriods,
			// 		}
			//
			// 		fundClawbackVestingCheck := execRevertedCheck
			// 		if callType.directCall {
			// 			fundClawbackVestingCheck = failCheck.WithErrContains("insufficient funds")
			// 		}
			//
			// 		_, _, err = s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, fundClawbackVestingCheck)
			// 		Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
			// 		Expect(s.network.NextBlock()).To(BeNil())
			//
			// 		// Check the vesting account
			// 		acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingKey.AccAddr)
			// 		va, ok := acc.(*vestingtypes.ClawbackVestingAccount)
			// 		Expect(ok).To(BeTrue())
			// 		Expect(va.LockupPeriods).To(BeNil(), "vesting account should not be funded")
			// 		Expect(va.VestingPeriods).To(BeNil(), "vesting account should not be funded")
			// 	})
			//
			// 	It(fmt.Sprintf("should not fund the vesting when not sending as the funder (%s)", callType.name), func() {
			// 		funder := s.keyring.GetKey(0)
			// 		vestingKey := s.keyring.GetKey(1)
			// 		differentSender := s.keyring.GetKey(2)
			//
			// 		err = s.factory.CreateClawbackVestingAccount(vestingKey.Priv, funder.AccAddr, false)
			// 		Expect(err).To(BeNil())
			// 		Expect(s.network.NextBlock()).To(BeNil())
			//
			// 		callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
			// 		callArgs.MethodName = vesting.FundVestingAccountMethod
			// 		callArgs.Args = []interface{}{
			// 			funder.Addr,
			// 			vestingKey.Addr,
			// 			uint64(time.Now().Unix()),
			// 			defaultPeriods,
			// 			defaultPeriods,
			// 		}
			//
			// 		fundClawbackVestingCheck := execRevertedCheck
			// 		if callType.directCall {
			// 			fundClawbackVestingCheck = failCheck.
			// 				WithErrContains(
			// 					fmt.Sprintf("tx origin address %s does not match the from address %s",
			// 						differentSender.Addr,
			// 						funder.Addr,
			// 					),
			// 				)
			// 		}
			//
			// 		_, _, err := s.factory.CallContractAndCheckLogs(differentSender.Priv, txArgs, callArgs, fundClawbackVestingCheck)
			// 		Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
			// 		Expect(s.network.NextBlock()).To(BeNil())
			//
			// 		// Check the vesting account
			// 		acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingKey.AccAddr)
			// 		Expect(acc).ToNot(BeNil(), "account should exist")
			// 		va, ok := acc.(*vestingtypes.ClawbackVestingAccount)
			// 		Expect(ok).To(BeTrue())
			// 		Expect(va.LockupPeriods).To(BeNil(), "vesting account should not be funded")
			// 		Expect(va.VestingPeriods).To(BeNil(), "vesting account should not be funded")
			// 	})
			//
			// 	It(fmt.Sprintf("should not fund the vesting when the address is blocked (%s)", callType.name), func() {
			// 		funder := s.keyring.GetKey(0)
			// 		moduleAddr := common.BytesToAddress(authtypes.NewModuleAddress("distribution").Bytes())
			//
			// 		callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
			// 		callArgs.MethodName = vesting.FundVestingAccountMethod
			// 		callArgs.Args = []interface{}{
			// 			funder.Addr,
			// 			moduleAddr,
			// 			uint64(time.Now().Unix()),
			// 			defaultPeriods,
			// 			defaultPeriods,
			// 		}
			//
			// 		fundClawbackVestingCheck := execRevertedCheck
			// 		if callType.directCall {
			// 			fundClawbackVestingCheck = failCheck.WithErrContains("is not allowed to receive funds")
			// 		}
			//
			// 		_, _, err := s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, fundClawbackVestingCheck)
			// 		Expect(err).NotTo(HaveOccurred(), "error while creating a clawback vesting account for a module address", err)
			//
			// 		// check that the module address is not a vesting account
			// 		acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), moduleAddr.Bytes())
			// 		Expect(acc).ToNot(BeNil(), "module account should be found")
			// 		_, ok := acc.(*vestingtypes.ClawbackVestingAccount)
			// 		Expect(ok).To(BeFalse(), "module account should not be a vesting account")
			// 	})
			//
			// 	It(fmt.Sprintf("should not fund the vesting when the address is blocked - a precompile address (%s)", callType.name), func() {
			// 		funder := s.keyring.GetKey(0)
			//
			// 		callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
			// 		callArgs.MethodName = vesting.FundVestingAccountMethod
			// 		callArgs.Args = []interface{}{
			// 			funder.Addr,
			// 			s.precompile.Address(),
			// 			uint64(time.Now().Unix()),
			// 			defaultPeriods,
			// 			defaultPeriods,
			// 		}
			//
			// 		fundClawbackVestingCheck := execRevertedCheck
			// 		if callType.directCall {
			// 			fundClawbackVestingCheck = failCheck.WithErrContains("is not allowed to receive funds")
			// 		}
			//
			// 		_, _, err := s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, fundClawbackVestingCheck)
			// 		Expect(err).NotTo(HaveOccurred(), "error while creating a clawback vesting account for a module address", err)
			// 	})
			//
			// 	It(fmt.Sprintf("should not fund the vesting when the address is uninitialized (%s)", callType.name), func() {
			// 		newAddr := testutiltx.GenerateAddress()
			// 		funder := s.keyring.GetKey(0)
			//
			// 		callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
			// 		callArgs.MethodName = vesting.FundVestingAccountMethod
			// 		callArgs.Args = []interface{}{
			// 			funder.Addr,
			// 			newAddr,
			// 			uint64(time.Now().Unix()),
			// 			defaultPeriods,
			// 			defaultPeriods,
			// 		}
			//
			// 		fundClawbackVestingCheck := execRevertedCheck
			// 		if callType.directCall {
			// 			fundClawbackVestingCheck = failCheck.WithErrContains("does not exist")
			// 		}
			//
			// 		_, _, err := s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, fundClawbackVestingCheck)
			// 		Expect(err).NotTo(HaveOccurred(), "error while creating a clawback vesting account for a module address", err)
			// 	})
			//
			// 	It(fmt.Sprintf("should not fund the vesting when the address is the zero address (%s)", callType.name), func() {
			// 		funder := s.keyring.GetKey(0)
			//
			// 		callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
			// 		callArgs.MethodName = vesting.FundVestingAccountMethod
			// 		callArgs.Args = []interface{}{
			// 			funder.Addr,
			// 			common.Address{},
			// 			uint64(time.Now().Unix()),
			// 			defaultPeriods,
			// 			defaultPeriods,
			// 		}
			//
			// 		fundClawbackVestingCheck := execRevertedCheck
			// 		if callType.directCall {
			// 			fundClawbackVestingCheck = failCheck.WithErrContains("invalid address")
			// 		}
			//
			// 		_, _, err := s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, fundClawbackVestingCheck)
			// 		Expect(err).NotTo(HaveOccurred(), "error while creating a clawback vesting account for the zero address", err)
			// 	})
		}
	})
	//
	// Context("to claw back from a vesting account", func() {
	// 	var (
	// 		clawbackReceiver   common.Address
	// 		funder, vestingKey testkeyring.Key
	// 		expClawbackAmt     = math.NewInt(1000)
	// 	)
	//
	// 	BeforeEach(func() {
	// 		funder = s.keyring.GetKey(0)
	// 		vestingKey = s.keyring.GetKey(1)
	// 		clawbackReceiver = testutiltx.GenerateAddress()
	//
	// 		err = s.factory.CreateClawbackVestingAccount(vestingKey.Priv, funder.AccAddr, false)
	// 		Expect(err).To(BeNil())
	// 		Expect(s.network.NextBlock()).To(BeNil())
	//
	// 		err = s.factory.FundVestingAccount(funder.Priv, vestingKey.AccAddr, time.Now(), sdkLockupPeriods, sdkVestingPeriods)
	// 		Expect(s.network.NextBlock()).To(BeNil())
	// 	})
	//
	// 	for _, callType := range callTypes {
	// 		callType := callType
	//
	// 		BeforeEach(func() {
	// 			if callType.directCall == false {
	// 				approvalCallArgs := factory.CallArgs{
	// 					ContractABI: s.precompile.ABI,
	// 					MethodName:  "approve",
	// 					Args: []interface{}{
	// 						contractAddr,
	// 						vesting.ClawbackMsgURL,
	// 					},
	// 				}
	//
	// 				precompileAddr := s.precompile.Address()
	// 				logCheck := passCheck.WithExpEvents(authorization.EventTypeApproval)
	//
	// 				_, _, err := s.factory.CallContractAndCheckLogs(funder.Priv, evmtypes.EvmTxArgs{To: &precompileAddr}, approvalCallArgs, logCheck)
	// 				Expect(err).To(BeNil())
	// 				Expect(s.network.NextBlock()).To(BeNil())
	// 			}
	// 		})
	//
	// 		Context("without authorization", func() {
	// 			It(fmt.Sprintf("should NOT claw back from the vesting when sending tx from the funder (%s)", callType.name), func() {
	// 				if callType.directCall {
	// 					Skip("this should only be run for smart contract calls")
	// 				}
	// 				balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")
	//
	// 				clawbackArgs := s.BuildCallArgs(callType, contractAddr).
	// 					WithMethodName(vesting.ClawbackMethod).
	// 					WithArgs(
	// 						s.address,
	// 						toAddr,
	// 						differentAddr,
	// 					)
	//
	// 				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, execRevertedCheck)
	// 				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
	//
	// 				balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePost.Amount).To(Equal(balancePre.Amount))
	// 				balanceReceiver := s.app.BankKeeper.GetBalance(s.ctx, differentAddr.Bytes(), s.bondDenom)
	// 				Expect(balanceReceiver.Amount).To(Equal(math.ZeroInt()))
	// 			})
	// 		})
	//
	// 		Context("with authorization", func() {
	// 			BeforeEach(func() {
	// 				if callType.directCall == false {
	// 					err = vesting.CreateGenericAuthz(s.ctx, s.app.AuthzKeeper, contractAddr, s.address, vesting.ClawbackMsgURL)
	// 					Expect(err).ToNot(HaveOccurred(), "error while creating the generic authorization: %v", err)
	// 				}
	// 			})
	//
	// 			It(fmt.Sprintf("should claw back from the vesting when sending as the funder (%s)", callType.name), func() {
	// 				res, err := s.grpcHandler.GetBalance(vestingKey.AccAddr, s.bondDenom)
	// 				Expect(err).To(BeNil())
	// 				balancePre := res.Balance
	//
	// 				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
	// 				callArgs.MethodName = vesting.ClawbackMethod
	// 				callArgs.Args = []interface{}{
	// 					funder.Addr,
	// 					vestingKey.Addr,
	// 					clawbackReceiver,
	// 				}
	//
	// 				clawbackCheck := passCheck.
	// 					WithExpEvents(vesting.EventTypeClawback)
	//
	// 				_, ethRes, err := s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, clawbackCheck)
	// 				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
	// 				Expect(s.network.NextBlock()).To(BeNil())
	//
	// 				var co vesting.ClawbackOutput
	// 				err = s.precompile.UnpackIntoInterface(&co, vesting.ClawbackMethod, ethRes.Ret)
	// 				Expect(err).ToNot(HaveOccurred(), "error while unpacking the clawback output: %v", err)
	// 				Expect(co.Coins).To(Equal(balances), "expected different clawback amount")
	//
	// 				balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePost.Amount.Int64()).To(Equal(int64(100)), "expected only initial balance after clawback")
	// 				balanceReceiver := s.app.BankKeeper.GetBalance(s.ctx, differentAddr.Bytes(), s.bondDenom)
	// 				Expect(balanceReceiver.Amount).To(Equal(expClawbackAmt), "expected receiver to show different balance after clawback")
	// 			})
	//
	// 			Context("table tests for clawback with state changes", func() {
	// 				type testCase struct {
	// 					dest       common.Address
	// 					transferTo *common.Address
	// 					before     bool
	// 					after      bool
	// 				}
	// 				DescribeTable(fmt.Sprintf("smart contract as funder - contract with state changes on destination address - should claw back from the vesting when sending as the funder (%s)", callType.name), func(tc testCase) {
	// 					if callType.directCall {
	// 						Skip("this should only be run for smart contract calls")
	// 					}
	// 					if tc.transferTo == nil {
	// 						tc.transferTo = &tc.dest
	// 					}
	// 					// change the vesting account funder to be the contract
	// 					_, err := s.app.VestingKeeper.UpdateVestingFunder(s.ctx, &vestingtypes.MsgUpdateVestingFunder{
	// 						FunderAddress:    sdk.AccAddress(s.address.Bytes()).String(),
	// 						NewFunderAddress: sdk.AccAddress(contractAddr.Bytes()).String(),
	// 						VestingAddress:   sdk.AccAddress(toAddr.Bytes()).String(),
	// 					})
	// 					Expect(err).ToNot(HaveOccurred())
	//
	// 					// fund the contract to make internal transfers
	// 					contractInitialBalance := math.NewInt(100)
	// 					err = evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, contractAddr.Bytes(), contractInitialBalance.Int64())
	// 					Expect(err).ToNot(HaveOccurred(), "error while funding the contract: %v", err)
	//
	// 					vestAccInitialBal := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 					Expect(vestAccInitialBal.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")
	//
	// 					clawbackArgs := s.BuildCallArgs(callType, contractAddr).
	// 						WithMethodName("clawbackWithTransfer").
	// 						WithArgs(
	// 							contractAddr,
	// 							toAddr,
	// 							tc.dest,
	// 							*tc.transferTo,
	// 							tc.before,
	// 							tc.after,
	// 						)
	//
	// 					clawbackCheck := passCheck.
	// 						WithExpEvents(vesting.EventTypeClawback)
	//
	// 					_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, clawbackCheck)
	// 					Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
	//
	// 					var co vesting.ClawbackOutput
	// 					err = s.precompile.UnpackIntoInterface(&co, vesting.ClawbackMethod, ethRes.Ret)
	// 					Expect(err).ToNot(HaveOccurred(), "error while unpacking the clawback output: %v", err)
	// 					Expect(co.Coins).To(Equal(balances), "expected different clawback amount")
	//
	// 					contractTransferredAmt := math.ZeroInt()
	// 					for _, transferred := range []bool{tc.before, tc.after} {
	// 						if transferred {
	// 							contractTransferredAmt = contractTransferredAmt.AddRaw(15)
	// 						}
	// 					}
	//
	// 					vestAccFinalBalance := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 					expVestAccFinalBal := vestAccInitialBal.Amount.Sub(expClawbackAmt)
	// 					if *tc.transferTo == toAddr {
	// 						expVestAccFinalBal = expVestAccFinalBal.Add(contractTransferredAmt)
	// 					}
	// 					Expect(vestAccFinalBalance.Amount).To(Equal(expVestAccFinalBal), "expected only initial balance after clawback")
	//
	// 					// contract transfers balances when it is not the destination
	// 					if tc.dest == contractAddr {
	// 						contractFinalBalance := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
	// 						Expect(contractFinalBalance.Amount).To(Equal(contractInitialBalance.Add(expClawbackAmt)))
	// 						return
	// 					}
	//
	// 					balanceDest := s.app.BankKeeper.GetBalance(s.ctx, tc.dest.Bytes(), s.bondDenom)
	// 					expBalDest := expClawbackAmt
	// 					if *tc.transferTo == tc.dest {
	// 						expBalDest = expBalDest.Add(contractTransferredAmt)
	// 					}
	// 					Expect(balanceDest.Amount).To(Equal(expBalDest), "expected receiver to show different balance after clawback")
	//
	// 					contractFinalBalance := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
	// 					Expect(contractFinalBalance.Amount).To(Equal(contractInitialBalance.Sub(contractTransferredAmt)))
	// 				},
	// 					Entry("funder is the destination address - state changes before & after precompile call", testCase{
	// 						dest:   contractAddr,
	// 						before: true,
	// 						after:  true,
	// 					}),
	// 					Entry("funder is the destination address - state changes before precompile call", testCase{
	// 						dest:   contractAddr,
	// 						before: true,
	// 						after:  false,
	// 					}),
	// 					Entry("funder is the destination address - state changes after precompile call", testCase{
	// 						dest:   contractAddr,
	// 						before: false,
	// 						after:  true,
	// 					}),
	// 					Entry("another address is the destination address - state changes before & after precompile", testCase{
	// 						dest:   differentAddr,
	// 						before: true,
	// 						after:  true,
	// 					}),
	// 					Entry("another address is the destination address - state changes before precompile", testCase{
	// 						dest:   differentAddr,
	// 						before: true,
	// 						after:  false,
	// 					}),
	// 					Entry("another address is the destination address - state changes after precompile", testCase{
	// 						dest:   differentAddr,
	// 						before: false,
	// 						after:  true,
	// 					}),
	// 					Entry("another address is the destination address - transfer to vest acc before & after precompile", testCase{
	// 						dest:       differentAddr,
	// 						transferTo: &toAddr,
	// 						before:     true,
	// 						after:      true,
	// 					}),
	// 					Entry("another address is the destination address - transfer to vest acc before precompile", testCase{
	// 						dest:       differentAddr,
	// 						transferTo: &toAddr,
	// 						before:     true,
	// 						after:      false,
	// 					}),
	// 					Entry("another address is the destination address - transfer to vest acc after precompile", testCase{
	// 						dest:       differentAddr,
	// 						transferTo: &toAddr,
	// 						before:     false,
	// 						after:      true,
	// 					}),
	// 				)
	// 			})
	//
	// 			It(fmt.Sprintf("should claw back from the vesting when sending as the funder with the caller smart contract as destination for the clawed back funds (%s)", callType.name), func() {
	// 				balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")
	//
	// 				// check the contract's (destination) initial balance. Should be 0
	// 				contractInitialBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
	// 				Expect(contractInitialBal.Amount).To(Equal(sdk.ZeroInt()))
	//
	// 				// get tx sender initial balance
	// 				txSenderInitialBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
	//
	// 				clawbackArgs := s.BuildCallArgs(callType, contractAddr).
	// 					WithMethodName(vesting.ClawbackMethod).
	// 					WithArgs(
	// 						s.address,
	// 						toAddr,
	// 						contractAddr,
	// 					).
	// 					WithGasPrice(gasPrice.BigInt())
	//
	// 				clawbackCheck := passCheck.
	// 					WithExpEvents(vesting.EventTypeClawback)
	//
	// 				res, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, clawbackCheck)
	// 				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
	//
	// 				fees := gasPrice.MulRaw(res.GasUsed)
	//
	// 				var co vesting.ClawbackOutput
	// 				err = s.precompile.UnpackIntoInterface(&co, vesting.ClawbackMethod, ethRes.Ret)
	// 				Expect(err).ToNot(HaveOccurred(), "error while unpacking the clawback output: %v", err)
	// 				Expect(co.Coins).To(Equal(balances), "expected different clawback amount")
	//
	// 				// check clawback account balance
	// 				balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePost.Amount.Int64()).To(Equal(int64(100)), "expected only initial balance after clawback")
	//
	// 				// check that tx signer's balance is reduced by the fees paid
	// 				txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
	// 				Expect(txSenderFinalBal.Amount).To(Equal(txSenderInitialBal.Amount.Sub(fees)))
	//
	// 				// check contract's final balance (clawback destination)
	// 				contractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
	// 				Expect(contractFinalBal.Amount).To(Equal(math.NewInt(1000)), "expected receiver to show different balance after clawback")
	// 			})
	//
	// 			It(fmt.Sprintf("clawback with revert after precompile call but before changing contract state - should NOT claw back and revert all balances to initial values (%s)", callType.name), func() { //nolint:dupl
	// 				if callType.directCall {
	// 					Skip("this should only be run for smart contract calls")
	// 				}
	// 				balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")
	//
	// 				clawbackArgs := s.BuildCallArgs(callType, contractAddr).
	// 					WithMethodName("clawbackWithRevert").
	// 					WithArgs(
	// 						s.address,
	// 						toAddr,
	// 						differentAddr,
	// 						true,
	// 					)
	//
	// 				err = s.precompile.UnpackIntoInterface(&co, vesting.ClawbackMethod, ethRes.Ret)
	// 				Expect(err).ToNot(HaveOccurred(), "error while unpacking the clawback output: %v", err)
	// 				Expect(co.Coins).To(Equal(balances), "expected different clawback amount")
	//
	// 				res, err = s.grpcHandler.GetBalance(vestingKey.AccAddr, s.bondDenom)
	// 				Expect(err).To(BeNil())
	// 				balancePost := res.Balance
	// 				Expect(balancePost.Amount).To(Equal(balancePre.Amount.Sub(expClawbackAmt)), "expected only initial balance after clawback")
	// 				res, err = s.grpcHandler.GetBalance(clawbackReceiver.Bytes(), s.bondDenom)
	// 				Expect(err).To(BeNil())
	// 				balanceReceiver := res.Balance
	// 				Expect(balanceReceiver.Amount).To(Equal(expClawbackAmt), "expected receiver to show different balance after clawback")
	// 			})
	//
	// 			Context("table tests for clawback with state changes", func() {
	// 				type testCase struct {
	// 					dest       common.Address
	// 					transferTo *common.Address
	// 					before     bool
	// 					after      bool
	// 				}
	// 				DescribeTable(fmt.Sprintf("smart contract as funder - contract with state changes on destination address - should claw back from the vesting when sending as the funder (%s)", callType.name), func(tc testCase) {
	// 					if callType.directCall {
	// 						Skip("this should only be run for smart contract calls")
	// 					}
	// 					if tc.transferTo == nil {
	// 						tc.transferTo = &tc.dest
	// 					}
	// 					// change the vesting account funder to be the contract
	// 					_, err := s.app.VestingKeeper.UpdateVestingFunder(s.ctx, &vestingtypes.MsgUpdateVestingFunder{
	// 						FunderAddress:    sdk.AccAddress(s.address.Bytes()).String(),
	// 						NewFunderAddress: sdk.AccAddress(contractAddr.Bytes()).String(),
	// 						VestingAddress:   sdk.AccAddress(toAddr.Bytes()).String(),
	// 					})
	// 					Expect(err).ToNot(HaveOccurred())
	//
	// 					// fund the contract to make internal transfers
	// 					contractInitialBalance := math.NewInt(100)
	// 					err = evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, contractAddr.Bytes(), contractInitialBalance.Int64())
	// 					Expect(err).ToNot(HaveOccurred(), "error while funding the contract: %v", err)
	//
	// 					vestAccInitialBal := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 					Expect(vestAccInitialBal.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")
	//
	// 					clawbackArgs := s.BuildCallArgs(callType, contractAddr).
	// 						WithMethodName("clawbackWithTransfer").
	// 						WithArgs(
	// 							contractAddr,
	// 							toAddr,
	// 							tc.dest,
	// 							*tc.transferTo,
	// 							tc.before,
	// 							tc.after,
	// 						)
	//
	// 					clawbackCheck := passCheck.
	// 						WithExpEvents(vesting.EventTypeClawback)
	//
	// 					_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, clawbackCheck)
	// 					Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
	//
	// 					var co vesting.ClawbackOutput
	// 					err = s.precompile.UnpackIntoInterface(&co, vesting.ClawbackMethod, ethRes.Ret)
	// 					Expect(err).ToNot(HaveOccurred(), "error while unpacking the clawback output: %v", err)
	// 					Expect(co.Coins).To(Equal(balances), "expected different clawback amount")
	//
	// 					contractTransferredAmt := math.ZeroInt()
	// 					for _, transferred := range []bool{tc.before, tc.after} {
	// 						if transferred {
	// 							contractTransferredAmt = contractTransferredAmt.AddRaw(15)
	// 						}
	// 					}
	//
	// 					vestAccFinalBalance := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 					expVestAccFinalBal := vestAccInitialBal.Amount.Sub(expClawbackAmt)
	// 					if *tc.transferTo == toAddr {
	// 						expVestAccFinalBal = expVestAccFinalBal.Add(contractTransferredAmt)
	// 					}
	// 					Expect(vestAccFinalBalance.Amount).To(Equal(expVestAccFinalBal), "expected only initial balance after clawback")
	//
	// 					// contract transfers balances when it is not the destination
	// 					if tc.dest == contractAddr {
	// 						contractFinalBalance := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
	// 						Expect(contractFinalBalance.Amount).To(Equal(contractInitialBalance.Add(expClawbackAmt)))
	// 						return
	// 					}
	//
	// 					balanceDest := s.app.BankKeeper.GetBalance(s.ctx, tc.dest.Bytes(), s.bondDenom)
	// 					expBalDest := expClawbackAmt
	// 					if *tc.transferTo == tc.dest {
	// 						expBalDest = expBalDest.Add(contractTransferredAmt)
	// 					}
	// 					Expect(balanceDest.Amount).To(Equal(expBalDest), "expected receiver to show different balance after clawback")
	//
	// 					contractFinalBalance := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
	// 					Expect(contractFinalBalance.Amount).To(Equal(contractInitialBalance.Sub(contractTransferredAmt)))
	// 				},
	// 					Entry("funder is the destination address - state changes before & after precompile call", testCase{
	// 						dest:   contractAddr,
	// 						before: true,
	// 						after:  true,
	// 					}),
	// 					Entry("funder is the destination address - state changes before precompile call", testCase{
	// 						dest:   contractAddr,
	// 						before: true,
	// 						after:  false,
	// 					}),
	// 					Entry("funder is the destination address - state changes after precompile call", testCase{
	// 						dest:   contractAddr,
	// 						before: false,
	// 						after:  true,
	// 					}),
	// 					Entry("another address is the destination address - state changes before & after precompile", testCase{
	// 						dest:   differentAddr,
	// 						before: true,
	// 						after:  true,
	// 					}),
	// 					Entry("another address is the destination address - state changes before precompile", testCase{
	// 						dest:   differentAddr,
	// 						before: true,
	// 						after:  false,
	// 					}),
	// 					Entry("another address is the destination address - state changes after precompile", testCase{
	// 						dest:   differentAddr,
	// 						before: false,
	// 						after:  true,
	// 					}),
	// 					Entry("another address is the destination address - transfer to vest acc before & after precompile", testCase{
	// 						dest:       differentAddr,
	// 						transferTo: &toAddr,
	// 						before:     true,
	// 						after:      true,
	// 					}),
	// 					Entry("another address is the destination address - transfer to vest acc before precompile", testCase{
	// 						dest:       differentAddr,
	// 						transferTo: &toAddr,
	// 						before:     true,
	// 						after:      false,
	// 					}),
	// 					Entry("another address is the destination address - transfer to vest acc after precompile", testCase{
	// 						dest:       differentAddr,
	// 						transferTo: &toAddr,
	// 						before:     false,
	// 						after:      true,
	// 					}),
	// 				)
	// 			})
	//
	// 			It(fmt.Sprintf("should claw back from the vesting when sending as the funder with the caller smart contract as destination for the clawed back funds (%s)", callType.name), func() {
	// 				// FIXME add here the corresponding args
	//
	// 				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, execRevertedCheck)
	// 				Expect(err).To(HaveOccurred())
	//
	// 				balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePost.Amount).To(Equal(balancePre.Amount), "expected no balance change")
	// 				balanceReceiver := s.app.BankKeeper.GetBalance(s.ctx, differentAddr.Bytes(), s.bondDenom)
	// 				Expect(balanceReceiver.Amount).To(Equal(math.ZeroInt()))
	// 			})
	//
	// 			It(fmt.Sprintf("clawback with revert after precompile after changing contract state - should NOT claw back and revert all balances to initial values (%s)", callType.name), func() { //nolint:dupl
	// 				if callType.directCall {
	// 					Skip("this should only be run for smart contract calls")
	// 				}
	// 				balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")
	//
	// 				clawbackArgs := s.BuildCallArgs(callType, contractAddr).
	// 					WithMethodName("clawbackWithRevert").
	// 					WithArgs(
	// 						s.address,
	// 						toAddr,
	// 						differentAddr,
	// 						false,
	// 					)
	//
	// 				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, execRevertedCheck)
	// 				Expect(err).To(HaveOccurred())
	//
	// 				balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePost.Amount).To(Equal(balancePre.Amount), "expected no balance change")
	// 				balanceReceiver := s.app.BankKeeper.GetBalance(s.ctx, differentAddr.Bytes(), s.bondDenom)
	// 				Expect(balanceReceiver.Amount).To(Equal(math.ZeroInt()))
	// 			})
	//
	// 			It(fmt.Sprintf("another contract as destination - should clawback from the vesting when sending as the funder with another smart contract as destination for the clawed back funds (%s)", callType.name), func() {
	// 				counterContract, err := contracts.LoadCounterContract()
	// 				Expect(err).ToNot(HaveOccurred())
	//
	// 				destContractAddr, err := s.DeployContract(counterContract)
	// 				Expect(err).ToNot(HaveOccurred(), "error while deploying the smart contract: %v", err)
	//
	// 				balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")
	//
	// 				// check the contract's (destination) initial balance. Should be 0
	// 				destContractInitialBal := s.app.BankKeeper.GetBalance(s.ctx, destContractAddr.Bytes(), s.bondDenom)
	// 				Expect(destContractInitialBal.Amount).To(Equal(sdk.ZeroInt()))
	//
	// 				// get tx sender initial balance
	// 				txSenderInitialBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
	//
	// 				clawbackCheck := passCheck.
	// 					WithExpEvents(vesting.EventTypeClawback)
	//
	// 				clawbackArgs := s.BuildCallArgs(callType, contractAddr).
	// 					WithMethodName(vesting.ClawbackMethod).
	// 					WithArgs(
	// 						s.address,
	// 						toAddr,
	// 						destContractAddr,
	// 					).
	// 					WithGasPrice(gasPrice.BigInt())
	//
	// 				res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, clawbackCheck)
	// 				Expect(err).NotTo(HaveOccurred())
	// 				fees := gasPrice.MulRaw(res.GasUsed)
	//
	// 				// check clawback account balance
	// 				balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePost.Amount).To(Equal(balancePre.Amount.Sub(expClawbackAmt)))
	//
	// 				// check that tx signer's balance is reduced by the fees paid
	// 				txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
	// 				Expect(txSenderFinalBal.Amount).To(Equal(txSenderInitialBal.Amount.Sub(fees)))
	//
	// 				// check caller contract's final balance should be zero
	// 				callerContractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
	// 				Expect(callerContractFinalBal.Amount).To(Equal(math.ZeroInt()))
	//
	// 				// check destination contract's final balance should
	// 				// have received the clawback amt
	// 				destContractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, destContractAddr.Bytes(), s.bondDenom)
	// 				Expect(destContractFinalBal.Amount).To(Equal(destContractInitialBal.Amount.Add(expClawbackAmt)))
	// 			})
	//
	// 			It(fmt.Sprintf("clawback with revert after precompile call but before changing contract state - should NOT claw back and revert all balances to initial values (%s)", callType.name), func() { //nolint:dupl
	// 				if callType.directCall {
	// 					Skip("this should only be run for smart contract calls")
	// 				}
	// 				balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")
	//
	// 				clawbackArgs := s.BuildCallArgs(callType, contractAddr).
	// 					WithMethodName("clawbackWithRevert").
	// 					WithArgs(
	// 						s.address,
	// 						toAddr,
	// 						differentAddr,
	// 						true,
	// 					)
	//
	// 				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, execRevertedCheck)
	// 				Expect(err).To(HaveOccurred())
	//
	// 				balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePost.Amount).To(Equal(balancePre.Amount), "expected no balance change")
	// 				balanceReceiver := s.app.BankKeeper.GetBalance(s.ctx, differentAddr.Bytes(), s.bondDenom)
	// 				Expect(balanceReceiver.Amount).To(Equal(math.ZeroInt()))
	// 			})
	//
	// 			It(fmt.Sprintf("clawback with revert after precompile after changing contract state - should NOT claw back and revert all balances to initial values (%s)", callType.name), func() { //nolint:dupl
	// 				if callType.directCall {
	// 					Skip("this should only be run for smart contract calls")
	// 				}
	// 				balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")
	//
	// 				clawbackArgs := s.BuildCallArgs(callType, contractAddr).
	// 					WithMethodName("clawbackWithRevert").
	// 					WithArgs(
	// 						s.address,
	// 						toAddr,
	// 						differentAddr,
	// 						false,
	// 					)
	//
	// 				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, execRevertedCheck)
	// 				Expect(err).To(HaveOccurred())
	//
	// 				balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePost.Amount).To(Equal(balancePre.Amount), "expected no balance change")
	// 				balanceReceiver := s.app.BankKeeper.GetBalance(s.ctx, differentAddr.Bytes(), s.bondDenom)
	// 				Expect(balanceReceiver.Amount).To(Equal(math.ZeroInt()))
	// 			})
	//
	// 			It(fmt.Sprintf("another contract as destination - should clawback from the vesting when sending as the funder with another smart contract as destination for the clawed back funds (%s)", callType.name), func() {
	// 				counterContract, err := contracts.LoadCounterContract()
	// 				Expect(err).ToNot(HaveOccurred())
	//
	// 				destContractAddr, err := s.DeployContract(counterContract)
	// 				Expect(err).ToNot(HaveOccurred(), "error while deploying the smart contract: %v", err)
	//
	// 				balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")
	//
	// 				// check the contract's (destination) initial balance. Should be 0
	// 				destContractInitialBal := s.app.BankKeeper.GetBalance(s.ctx, destContractAddr.Bytes(), s.bondDenom)
	// 				Expect(destContractInitialBal.Amount).To(Equal(sdk.ZeroInt()))
	//
	// 				// get tx sender initial balance
	// 				txSenderInitialBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
	//
	// 				clawbackCheck := passCheck.
	// 					WithExpEvents(vesting.EventTypeClawback)
	//
	// 				clawbackArgs := s.BuildCallArgs(callType, contractAddr).
	// 					WithMethodName(vesting.ClawbackMethod).
	// 					WithArgs(
	// 						s.address,
	// 						toAddr,
	// 						destContractAddr,
	// 					).
	// 					WithGasPrice(gasPrice.BigInt())
	//
	// 				res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, clawbackCheck)
	// 				Expect(err).NotTo(HaveOccurred())
	// 				fees := gasPrice.MulRaw(res.GasUsed)
	//
	// 				// check clawback account balance
	// 				balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePost.Amount).To(Equal(balancePre.Amount.Sub(expClawbackAmt)))
	//
	// 				// check that tx signer's balance is reduced by the fees paid
	// 				txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
	// 				Expect(txSenderFinalBal.Amount).To(Equal(txSenderInitialBal.Amount.Sub(fees)))
	//
	// 				// check caller contract's final balance should be zero
	// 				callerContractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
	// 				Expect(callerContractFinalBal.Amount).To(Equal(math.ZeroInt()))
	//
	// 				// check destination contract's final balance should
	// 				// have received the clawback amt
	// 				destContractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, destContractAddr.Bytes(), s.bondDenom)
	// 				Expect(destContractFinalBal.Amount).To(Equal(destContractInitialBal.Amount.Add(expClawbackAmt)))
	// 			})
	//
	// 			It(fmt.Sprintf("another contract as destination - should claw back from the vesting when sending as the funder with another smart contract as destination and triggering state change on destination contract (%s)", callType.name), func() {
	// 				if callType.directCall {
	// 					Skip("this should only be run for smart contract calls")
	// 				}
	// 				counterContract, err := contracts.LoadCounterContract()
	// 				Expect(err).ToNot(HaveOccurred())
	//
	// 				destContractAddr, err := s.DeployContract(counterContract)
	// 				Expect(err).ToNot(HaveOccurred(), "error while deploying the smart contract: %v", err)
	//
	// 				balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")
	//
	// 				// check the contract's (destination) initial balance. Should be 0
	// 				destContractInitialBal := s.app.BankKeeper.GetBalance(s.ctx, destContractAddr.Bytes(), s.bondDenom)
	// 				Expect(destContractInitialBal.Amount).To(Equal(sdk.ZeroInt()))
	//
	// 				// get tx sender initial balance
	// 				txSenderInitialBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
	//
	// 				clawbackArgs := s.BuildCallArgs(callType, contractAddr).
	// 					WithMethodName("clawbackWithCounterCall").
	// 					WithArgs(
	// 						s.address,
	// 						toAddr,
	// 						destContractAddr,
	// 					).
	// 					WithGasPrice(gasPrice.BigInt())
	//
	// 				// expect the vesting precompile events and the Counter
	// 				// contract's events
	// 				clawbackCheck := passCheck.
	// 					WithABIEvents(mergeEventMaps(
	// 						s.precompile.Events,
	// 						counterContract.ABI.Events,
	// 					)).
	// 					WithExpEvents([]string{
	// 						"Added", "Changed",
	// 						vesting.EventTypeClawback,
	// 						"Changed",
	// 					}...)
	//
	// 				res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, clawbackCheck)
	// 				Expect(err).NotTo(HaveOccurred())
	// 				fees := gasPrice.MulRaw(res.GasUsed)
	//
	// 				// check clawback account balance
	// 				balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
	// 				Expect(balancePost.Amount).To(Equal(balancePre.Amount.Sub(expClawbackAmt)))
	//
	// 				// check that tx signer's balance is reduced by the fees paid
	// 				txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
	// 				Expect(txSenderFinalBal.Amount).To(Equal(txSenderInitialBal.Amount.Sub(fees)))
	//
	// 				// check caller contract's final balance should be zero
	// 				callerContractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
	// 				Expect(callerContractFinalBal.Amount).To(Equal(math.ZeroInt()))
	//
	// 				// check destination contract's final balance should
	// 				// have received the clawback amt
	// 				destContractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, destContractAddr.Bytes(), s.bondDenom)
	// 				Expect(destContractFinalBal.Amount).To(Equal(destContractInitialBal.Amount.Add(expClawbackAmt)))
	// 			})
	//
	// 			It(fmt.Sprintf("should return an error when not sending as the funder (%s)", callType.name), func() {
	// 				differentSender := s.keyring.GetKey(2)
	//
	// 				res, err := s.grpcHandler.GetBalance(vestingKey.AccAddr, s.bondDenom)
	// 				Expect(err).To(BeNil())
	// 				balancePre := res.Balance
	//
	// 				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
	// 				callArgs.MethodName = vesting.ClawbackMethod
	// 				callArgs.Args = []interface{}{
	// 					funder.Addr,
	// 					vestingKey.Addr,
	// 					differentSender.Addr,
	// 				}
	//
	// 				clawbackCheck := execRevertedCheck
	// 				if callType.directCall {
	// 					clawbackCheck = failCheck.
	// 						WithErrContains(fmt.Sprintf(
	// 							"tx origin address %s does not match the funder address %s",
	// 							differentSender.Addr, funder.Addr,
	// 						))
	// 				}
	//
	// 				_, _, err = s.factory.CallContractAndCheckLogs(differentSender.Priv, txArgs, callArgs, clawbackCheck)
	// 				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
	//
	// 				res, err = s.grpcHandler.GetBalance(vestingKey.AccAddr, s.bondDenom)
	// 				Expect(err).To(BeNil())
	// 				balancePost := res.Balance
	// 				Expect(balancePost).To(Equal(balancePre), "expected balance not to have changed")
	// 			})
	//
	// 			It(fmt.Sprintf("should return an error when the vesting does not exist (%s)", callType.name), func() {
	// 				nonVestingKey := s.keyring.GetKey(2)
	//
	// 				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
	// 				callArgs.MethodName = vesting.ClawbackMethod
	// 				callArgs.Args = []interface{}{
	// 					funder.Addr,
	// 					nonVestingKey.Addr,
	// 					funder.Addr,
	// 				}
	//
	// 				clawbackCheck := execRevertedCheck
	// 				// FIXME: error messages in fail check now work differently!
	// 				if callType.directCall {
	// 					clawbackCheck = failCheck.
	// 						WithErrContains(vestingtypes.ErrNotSubjectToClawback.Error())
	// 				}
	//
	// 				_, _, err = s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, clawbackCheck)
	// 				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
	// 			})
	//
	// 			It(fmt.Sprintf("should succeed and return empty Coins when all tokens are vested (%s)", callType.name), func() {
	// 				// commit block with time so that vesting has ended
	// 				err := s.network.NextBlockAfter(time.Hour * 24)
	// 				Expect(err).ToNot(HaveOccurred(), "error while committing block: %v", err)
	//
	// 				res, err := s.grpcHandler.GetBalance(vestingKey.AccAddr, s.bondDenom)
	// 				Expect(err).To(BeNil())
	// 				balancePre := res.Balance
	//
	// 				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
	// 				callArgs.MethodName = vesting.ClawbackMethod
	// 				callArgs.Args = []interface{}{
	// 					funder.Addr,
	// 					vestingKey.Addr,
	// 					funder.Addr,
	// 				}
	//
	// 				_, ethRes, err := s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, passCheck)
	// 				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
	// 				Expect(s.network.NextBlock()).To(BeNil())
	//
	// 				var co vesting.ClawbackOutput
	// 				err = s.precompile.UnpackIntoInterface(&co, vesting.ClawbackMethod, ethRes.Ret)
	// 				Expect(err).ToNot(HaveOccurred(), "error while unpacking the clawback output: %v", err)
	// 				Expect(co.Coins).To(BeEmpty(), "expected empty clawback amount")
	//
	// 				res, err = s.grpcHandler.GetBalance(vestingKey.AccAddr, s.bondDenom)
	// 				Expect(err).To(BeNil())
	// 				balancePost := res.Balance
	// 				Expect(balancePost).To(Equal(balancePre), "expected balance not to have changed")
	// 			})
	// 		})
	// 	}
	// })
	//
	// Context("to update the vesting funder", func() {
	// 	var funder, newFunder, vestingKey testkeyring.Key
	//
	// 	BeforeEach(func() {
	// 		funder = s.keyring.GetKey(0)
	// 		vestingKey = s.keyring.GetKey(1)
	// 		newFunder = s.keyring.GetKey(2)
	//
	// 		err = s.factory.CreateClawbackVestingAccount(vestingKey.Priv, funder.AccAddr, false)
	// 		Expect(err).To(BeNil())
	// 		Expect(s.network.NextBlock()).To(BeNil())
	// 	})
	//
	// 	for _, callType := range callTypes {
	// 		callType := callType
	//
	// 		BeforeEach(func() {
	// 			if callType.directCall == false {
	// 				approvalCallArgs := factory.CallArgs{
	// 					ContractABI: s.precompile.ABI,
	// 					MethodName:  "approve",
	// 					Args: []interface{}{
	// 						contractAddr,
	// 						vesting.UpdateVestingFunderMsgURL,
	// 					},
	// 				}
	//
	// 				precompileAddr := s.precompile.Address()
	// 				logCheck := passCheck.WithExpEvents(authorization.EventTypeApproval)
	//
	// 				_, _, err := s.factory.CallContractAndCheckLogs(funder.Priv, evmtypes.EvmTxArgs{To: &precompileAddr}, approvalCallArgs, logCheck)
	// 				Expect(err).To(BeNil())
	// 				Expect(s.network.NextBlock()).To(BeNil())
	// 			}
	// 		})
	//
	// 		It(fmt.Sprintf("should update the vesting funder when sending as the funder (%s)", callType.name), func() {
	// 			callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
	// 			callArgs.MethodName = vesting.UpdateVestingFunderMethod
	// 			callArgs.Args = []interface{}{
	// 				funder.Addr,
	// 				newFunder.Addr,
	// 				vestingKey.Addr,
	// 			}
	//
	// 			updateFunderCheck := passCheck.
	// 				WithExpEvents(vesting.EventTypeUpdateVestingFunder)
	//
	// 			_, _, err := s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, updateFunderCheck)
	// 			Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
	// 			Expect(s.network.NextBlock()).To(BeNil())
	//
	// 			// Check that the vesting account has the new funder
	// 			s.ExpectVestingFunder(vestingKey.Addr, newFunder.Addr)
	// 		})
	//
	// 		It(fmt.Sprintf("should return an error when not sending as the funder (%s)", callType.name), func() {
	// 			differentSender := s.keyring.GetKey(2)
	//
	// 			callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
	// 			callArgs.MethodName = vesting.UpdateVestingFunderMethod
	// 			callArgs.Args = []interface{}{
	// 				funder.Addr,
	// 				differentSender.Addr,
	// 				vestingKey.Addr,
	// 			}
	//
	// 			updateFunderCheck := execRevertedCheck
	// 			if callType.directCall {
	// 				updateFunderCheck = failCheck.
	// 					WithErrContains(fmt.Sprintf(
	// 						"tx origin address %s does not match the funder address %s",
	// 						differentSender.Addr, funder.Addr.String(),
	// 					))
	// 			}
	//
	// 			_, _, err = s.factory.CallContractAndCheckLogs(differentSender.Priv, txArgs, callArgs, updateFunderCheck)
	// 			Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
	// 			Expect(s.network.NextBlock()).To(BeNil())
	//
	// 			// Check that the vesting account still has the same funder
	// 			s.ExpectVestingFunder(vestingKey.Addr, funder.Addr)
	// 		})
	//
	// 		It(fmt.Sprintf("should return an error when the account does not exist (%s)", callType.name), func() {
	// 			// Check that there's no account
	// 			nonExistentAddr := testutiltx.GenerateAddress()
	// 			acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), nonExistentAddr.Bytes())
	// 			Expect(acc).To(BeNil(), "expected no account to be found")
	//
	// 			callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
	// 			callArgs.MethodName = vesting.UpdateVestingFunderMethod
	// 			callArgs.Args = []interface{}{
	// 				funder.Addr,
	// 				newFunder.Addr,
	// 				nonExistentAddr, // the address of the vesting account
	// 			}
	//
	// 			updateFunderCheck := execRevertedCheck
	// 			if callType.directCall {
	// 				updateFunderCheck = failCheck.
	// 					WithErrContains(fmt.Sprintf(
	// 						"account at address '%s' does not exist",
	// 						sdk.AccAddress(nonExistentAddr.Bytes()).String(),
	// 					))
	// 			}
	//
	// 			_, _, err = s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, updateFunderCheck)
	// 			Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
	// 		})
	//
	// 		It(fmt.Sprintf("should return an error when the account is no vesting account (%s)", callType.name), func() {
	// 			KeyWithNoVesting := s.keyring.GetKey(2)
	//
	// 			acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), KeyWithNoVesting.AccAddr)
	// 			Expect(acc).ToNot(BeNil(), "expected account to be found")
	// 			_, ok := acc.(*vestingtypes.ClawbackVestingAccount)
	// 			Expect(ok).To(BeFalse(), "expected account not to be a vesting account")
	//
	// 			callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
	// 			callArgs.MethodName = vesting.UpdateVestingFunderMethod
	// 			callArgs.Args = []interface{}{
	// 				funder.Addr,
	// 				newFunder.Addr,
	// 				KeyWithNoVesting.Addr, // the address of the vesting account
	// 			}
	//
	// 			updateFunderCheck := execRevertedCheck
	// 			if callType.directCall {
	// 				updateFunderCheck = failCheck.
	// 					WithErrContains(fmt.Sprintf(
	// 						"%s: %s",
	// 						KeyWithNoVesting.AccAddr,
	// 						vestingtypes.ErrNotSubjectToClawback.Error(),
	// 					))
	// 			}
	//
	// 			_, _, err = s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, updateFunderCheck)
	// 			Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
	// 		})
	//
	// 		It(fmt.Sprintf("should return an error when the new funder is the zero address (%s)", callType.name), func() {
	// 			callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
	// 			callArgs.MethodName = vesting.UpdateVestingFunderMethod
	// 			callArgs.Args = []interface{}{
	// 				funder.Addr,
	// 				common.Address{},
	// 				vestingKey.Addr,
	// 			}
	//
	// 			updateFunderCheck := execRevertedCheck
	// 			if callType.directCall {
	// 				updateFunderCheck = failCheck.
	// 					WithErrContains("new funder address cannot be the zero address")
	// 			}
	//
	// 			_, _, err = s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, updateFunderCheck)
	// 			Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
	// 		})
	//
	// 		It(fmt.Sprintf("should return an error when the new funder is the same as the current funder (%s)", callType.name), func() {
	// 			callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
	// 			callArgs.MethodName = vesting.UpdateVestingFunderMethod
	// 			callArgs.Args = []interface{}{
	// 				funder.Addr,
	// 				funder.Addr,
	// 				vestingKey.Addr,
	// 			}
	//
	// 			updateFunderCheck := execRevertedCheck
	// 			if callType.directCall {
	// 				updateFunderCheck = failCheck.
	// 					WithErrContains("new funder address is equal to current funder address")
	// 			}
	//
	// 			_, _, err = s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, updateFunderCheck)
	// 			Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
	//
	// 			// Check that the vesting account still has the same funder
	// 			s.ExpectVestingFunder(vestingKey.Addr, funder.Addr)
	// 		})
	//
	// 		It(fmt.Sprintf("should return an error when the new funder is a blocked address (%s)", callType.name), func() {
	// 			moduleAddr := common.BytesToAddress(authtypes.NewModuleAddress("distribution").Bytes())
	//
	// 			callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
	// 			callArgs.MethodName = vesting.UpdateVestingFunderMethod
	// 			callArgs.Args = []interface{}{
	// 				funder.Addr,
	// 				moduleAddr,
	// 				vestingKey.Addr,
	// 			}
	//
	// 			updateFunderCheck := execRevertedCheck
	// 			if callType.directCall {
	// 				updateFunderCheck = failCheck.
	// 					WithErrContains("not allowed to fund vesting accounts")
	// 			}
	//
	// 			_, _, err = s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, updateFunderCheck)
	// 			Expect(err).NotTo(HaveOccurred(), "error while updating the funder to a module address: %v", err)
	// 		})
	// 	}
	// })
	//
	// Context("to convert a vesting account", func() {
	// 	var funder, vestingKey, KeyWithNoVesting testkeyring.Key
	//
	// 	BeforeEach(func() {
	// 		funder = s.keyring.GetKey(0)
	// 		vestingKey = s.keyring.GetKey(1)
	// 		KeyWithNoVesting = s.keyring.GetKey(2)
	//
	// 		// Create a vesting account
	// 		err = s.factory.CreateClawbackVestingAccount(vestingKey.Priv, funder.AccAddr, false)
	// 		Expect(err).To(BeNil())
	// 		Expect(s.network.NextBlock()).To(BeNil())
	//
	// 		// Fund vesting account
	// 		err = s.factory.FundVestingAccount(funder.Priv, vestingKey.AccAddr, time.Now(), sdkLockupPeriods, sdkVestingPeriods)
	// 		Expect(err).To(BeNil())
	// 		Expect(s.network.NextBlock()).To(BeNil())
	// 	})
	//
	// 	for _, callType := range callTypes {
	// 		callType := callType
	//
	// 		It(fmt.Sprintf("should convert the vesting account into a normal one after vesting has ended (%s)", callType.name), func() {
	// 			// commit block with new time so that the vesting period has ended
	// 			err = s.network.NextBlockAfter(time.Hour * 24)
	// 			Expect(err).To(BeNil(), "failed to commit block")
	//
	// 			callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
	// 			callArgs.MethodName = vesting.ConvertVestingAccountMethod
	// 			callArgs.Args = []interface{}{
	// 				vestingKey.Addr,
	// 			}
	//
	// 			convertClawbackCheck := passCheck.
	// 				WithExpEvents(vesting.EventTypeConvertVestingAccount)
	//
	// 			_, _, err := s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, convertClawbackCheck)
	// 			Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
	// 			Expect(s.network.NextBlock()).To(BeNil(), "failed to commit block")
	//
	// 			// Check that the vesting account has been converted
	// 			acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingKey.AccAddr)
	// 			Expect(acc).ToNot(BeNil(), "expected account to be found")
	// 			_, ok := acc.(*vestingtypes.ClawbackVestingAccount)
	// 			Expect(ok).To(BeFalse(), "expected account not to be a vesting account")
	// 		})
	//
	// 		It(fmt.Sprintf("should return an error when the vesting has not ended yet (%s)", callType.name), func() {
	// 			callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
	// 			callArgs.MethodName = vesting.ConvertVestingAccountMethod
	// 			callArgs.Args = []interface{}{
	// 				vestingKey.Addr,
	// 			}
	//
	// 			convertClawbackCheck := execRevertedCheck
	// 			if callType.directCall {
	// 				convertClawbackCheck = failCheck.WithErrContains("vesting coins still left in account")
	// 			}
	//
	// 			_, _, err := s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, convertClawbackCheck)
	// 			Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
	// 		})
	//
	// 		It(fmt.Sprintf("should return an error when the vesting does not exist (%s)", callType.name), func() {
	// 			callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
	// 			callArgs.MethodName = vesting.ConvertVestingAccountMethod
	// 			callArgs.Args = []interface{}{
	// 				KeyWithNoVesting.Addr, // this currently has no vesting
	// 			}
	//
	// 			convertClawbackCheck := execRevertedCheck
	// 			if callType.directCall {
	// 				convertClawbackCheck = failCheck.WithErrContains("account is not subject to clawback vesting")
	// 			}
	//
	// 			_, _, err = s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, convertClawbackCheck)
	// 			Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
	//
	// 			// Check that the account is no vesting account
	// 			acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), KeyWithNoVesting.AccAddr)
	// 			Expect(acc).ToNot(BeNil(), "expected account to be found")
	// 			_, ok := acc.(*vestingtypes.ClawbackVestingAccount)
	// 			Expect(ok).To(BeFalse(), "expected account not to be a vesting account")
	// 		})
	// 	}
	// })
	//
	// ---------------------------------------------
	//                     QUERIES
	//
	Context("to get vesting balances", func() {
		for _, callType := range callTypes {
			callType := callType

			It(fmt.Sprintf("should return the vesting when it exists (%s)", callType.name), func() {
				funder := s.keyring.GetKey(0)
				vestingKey := s.keyring.GetKey(1)

				err = s.factory.CreateClawbackVestingAccount(vestingKey.Priv, funder.AccAddr, false)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				err = s.factory.FundVestingAccount(funder.Priv, vestingKey.AccAddr, time.Now(), sdkLockupPeriods, sdkVestingPeriods)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				callArgs.MethodName = vesting.BalancesMethod
				callArgs.Args = []interface{}{
					vestingKey.Addr,
				}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				var res vesting.BalancesOutput
				err = s.precompile.UnpackIntoInterface(&res, vesting.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "error while unpacking response: %v", err)

				expectedCoins := []cmn.Coin{{Denom: s.bondDenom, Amount: big.NewInt(1000)}}
				Expect(res.Locked).To(Equal(expectedCoins), "expected different locked coins")
				Expect(res.Unvested).To(Equal(expectedCoins), "expected different unvested coins")
				Expect(res.Vested).To(BeEmpty(), "expected different vested coins")

				// Commit new block so that the vesting period is at the half and the lockup period is over
				err = s.network.NextBlockAfter(time.Second * 5000)
				Expect(err).To(BeNil(), "failed to commit block")

				// Recheck balances
				_, ethRes, err = s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				err = s.precompile.UnpackIntoInterface(&res, vesting.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "error while unpacking response: %v", err)

				halfCoins := []cmn.Coin{{Denom: s.bondDenom, Amount: big.NewInt(500)}}
				Expect(res.Locked).To(BeEmpty(), "expected no coins to be locked anymore")
				Expect(res.Unvested).To(Equal(halfCoins), "expected different unvested coins")
				Expect(res.Vested).To(Equal(halfCoins), "expected different vested coins")

				// Commit new block so that the vesting period is over
				err = s.network.NextBlockAfter(time.Second * 5000)
				Expect(err).To(BeNil(), "failed to commit block")

				// Recheck balances
				_, ethRes, err = s.factory.CallContractAndCheckLogs(funder.Priv, txArgs, callArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				err = s.precompile.UnpackIntoInterface(&res, vesting.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "error while unpacking response: %v", err)

				Expect(res.Locked).To(BeEmpty(), "expected no coins to be locked anymore")
				Expect(res.Unvested).To(BeEmpty(), "expected no coins to be unvested anymore")
				Expect(res.Vested).To(Equal(expectedCoins), "expected different vested coins")
			})

			It(fmt.Sprintf("should return an error when the account does not exist (%s)", callType.name), func() {
				sender := s.keyring.GetKey(0)
				nonExistentAddr := testutiltx.GenerateAddress()

				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				callArgs.MethodName = vesting.BalancesMethod
				callArgs.Args = []interface{}{
					nonExistentAddr,
				}

				balancesCheck := execRevertedCheck
				if callType.directCall {
					balancesCheck = failCheck.WithErrContains(fmt.Sprintf(
						"account at address '%s' either does not exist or is not a vesting account", sdk.AccAddress(nonExistentAddr.Bytes())))
				}

				_, _, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, callArgs, balancesCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should return an error when the account is not a vesting account (%s)", callType.name), func() {
				KeyWithNoVesting := s.keyring.GetKey(0)

				callArgs, txArgs := s.BuildCallArgs(callType, vestingCallerAddr)
				callArgs.MethodName = vesting.BalancesMethod
				callArgs.Args = []interface{}{
					KeyWithNoVesting.Addr,
				}

				balancesCheck := execRevertedCheck
				if callType.directCall {
					balancesCheck = failCheck.WithErrContains(fmt.Sprintf(
						"account at address '%s' either does not exist or is not a vesting account",
						KeyWithNoVesting.AccAddr,
					))
				}

				_, _, err = s.factory.CallContractAndCheckLogs(KeyWithNoVesting.Priv, txArgs, callArgs, balancesCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
			})
		}
	})
})
