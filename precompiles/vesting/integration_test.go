// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package vesting_test

import (
	"fmt"
	"math/big"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/testutil"
	"github.com/evmos/evmos/v18/precompiles/testutil/contracts"
	"github.com/evmos/evmos/v18/precompiles/vesting"
	evmosutil "github.com/evmos/evmos/v18/testutil"
	testutiltx "github.com/evmos/evmos/v18/testutil/tx"
	"github.com/evmos/evmos/v18/utils"
	vestingtypes "github.com/evmos/evmos/v18/x/vesting/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

var (
	// contractAddr is the address of the smart contract that calls the vesting extension
	contractAddr common.Address
	// failCheck is the default setting to check execution logs for failing transactions
	failCheck testutil.LogCheckArgs
	// defaultPeriods is a slice of default periods used in testing
	defaultPeriods []vesting.Period
	// doublePeriods is a slice of two default periods used in testing
	doublePeriods []vesting.Period
	// emptyPeriods is a empty slice of periods used in testing
	emptyPeriods []vesting.Period
	// err is a basic error type
	err error
	// instantPeriods is a slice of instant periods used in testing (i.e. length = 0)
	instantPeriods []vesting.Period
	// execRevertedCheck is a basic check for contract calls to the precompile, where only "execution reverted" is returned
	execRevertedCheck testutil.LogCheckArgs
	// passCheck is a basic check that is used to check if the transaction was successful
	passCheck testutil.LogCheckArgs

	// callTypes is a slice of testing configurations used to run the test cases for direct
	// contract calls as well as calls through a smart contract.
	callTypes = []CallType{
		{name: "directly", directCall: true},
		{name: "through a smart contract", directCall: false},
	}
	// differentAddr is a new address used in testing
	differentAddr = testutiltx.GenerateAddress()
	// vestingAddr is a new address that is used to test the vesting extension.
	vestingAddr = testutiltx.GenerateAddress()
	// gasPrice to be used on tests txs and calculate the fees
	gasPrice = math.NewInt(1e9)
)

var _ = Describe("Interacting with the vesting extension", func() {
	BeforeEach(func() {
		// Setup the test suite
		s.SetupTest()
		s.NextBlock()

		// Set the default value for the vesting or lockup periods
		defaultPeriod := vesting.Period{
			Length: 10,
			Amount: []cmn.Coin{{Denom: s.bondDenom, Amount: big.NewInt(100)}},
		}
		instantPeriod := defaultPeriod
		instantPeriod.Length = 0
		defaultPeriods = []vesting.Period{defaultPeriod}
		doublePeriods = []vesting.Period{defaultPeriod, defaultPeriod}
		instantPeriods = []vesting.Period{instantPeriod}

		// Deploy the smart contract that calls the vesting extension
		contractAddr, err = s.DeployContract(s.vestingCallerContract)
		Expect(err).ToNot(HaveOccurred(), "error while deploying the smart contract: %v", err)

		// Set up the checks
		failCheck = testutil.LogCheckArgs{
			ABIEvents: s.precompile.Events,
			ExpPass:   false,
		}
		execRevertedCheck = failCheck.WithErrContains("execution reverted")
		passCheck = failCheck.WithExpPass(true)
	})

	// ---------------------------------------------
	//                   TRANSACTIONS
	//

	Context("to create a clawback vesting account", func() {
		for _, callType := range callTypes {
			callType := callType

			It(fmt.Sprintf("should create a clawback vesting account (%s)", callType.name), func() {
				err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, toAddr.Bytes(), 10000)
				Expect(err).ToNot(HaveOccurred(), "error while funding the account: %v", err)
				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.CreateClawbackVestingAccountMethod).
					WithArgs(
						funderAddr,
						s.address,
						false,
					)

				createClawbackCheck := passCheck.WithExpEvents(vesting.EventTypeCreateClawbackVestingAccount)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, createClawbackCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				// Check the vesting account
				s.ExpectSimpleVestingAccount(s.address, funderAddr)
			})

			It(fmt.Sprintf("should not create a clawback vesting account if the account is not initialized (%s)", callType.name), func() {
				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.CreateClawbackVestingAccountMethod).
					WithArgs(
						s.address,
						toAddr,
					)

				createClawbackCheck := failCheck.WithErrContains("account is not initialized")

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, createClawbackCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)

				acc := s.app.AccountKeeper.GetAccount(s.ctx, toAddr.Bytes())
				Expect(acc).To(BeNil(), "account should not be created")
			})

			It(fmt.Sprintf("should not create a clawback vesting account if the vesting account is the zero address (%s)", callType.name), func() {
				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.CreateClawbackVestingAccountMethod).
					WithArgs(
						s.address,
						common.Address{},
					)

				createClawbackCheck := failCheck.WithErrContains("account is not initialized")

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, createClawbackCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should not create a clawback vesting account if the funder account is the zero address (%s)", callType.name), func() {
				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.CreateClawbackVestingAccountMethod).
					WithArgs(
						common.Address{},
						s.address,
					)

				createClawbackCheck := failCheck.WithErrContains("account is not initialized")

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, createClawbackCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should not create a clawback vesting account if the origin is different than the vesting address (%s)", callType.name), func() {
				err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, differentAddr.Bytes(), 10000)
				Expect(err).ToNot(HaveOccurred(), "error while funding the account: %v", err)
				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.CreateClawbackVestingAccountMethod).
					WithArgs(
						s.address,
						differentAddr,
						false,
					)

				createClawbackCheck := failCheck.WithErrContains("origin is different than the vesting address")

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, createClawbackCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("does not match the from address"))
				}
			})

			It(fmt.Sprintf("should not create a clawback vesting account for a smart contract (%s)", callType.name), func() {
				if callType.directCall {
					Skip("this should only be run for smart contract calls")
				}

				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName("createClawbackVestingAccountForContract").
					WithArgs()

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, failCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(err.Error()).To(ContainSubstring("execution reverted"))

				// Check that the smart contract was not converted
				acc := s.app.AccountKeeper.GetAccount(s.ctx, contractAddr.Bytes())
				Expect(acc).ToNot(BeNil(), "smart contract should be found")
				_, ok := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(ok).To(BeFalse(), "smart contract should not be converted to a vesting account")
			})

			It(fmt.Sprintf("should not create a clawback vesting account if the account already is subject to vesting (%s)", callType.name), func() {
				addr, priv := testutiltx.NewAddrKey()
				err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, addr.Bytes(), 1e18)
				Expect(err).ToNot(HaveOccurred(), "error while funding the account: %v", err)

				s.CreateTestClawbackVestingAccount(s.address, addr)

				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.CreateClawbackVestingAccountMethod).
					WithPrivKey(priv). // send from the vesting account
					WithArgs(
						s.address,
						addr,
						false,
					)

				createClawbackCheck := failCheck.WithErrContains("account is already subject to vesting")

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, createClawbackCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("%s is already a clawback vesting account", sdk.AccAddress(addr.Bytes()))))
				}
			})
		}
	})

	Context("to fund a clawback vesting account", func() {
		for _, callType := range callTypes {
			callType := callType

			It(fmt.Sprintf("should fund the vesting when defining only lockup (%s)", callType.name), func() {
				if !callType.directCall {
					// create authorization to allow contract to spend the funder's (s.address) balance
					// when funding a vesting account
					err = vesting.CreateGenericAuthz(s.ctx, s.app.AuthzKeeper, contractAddr, s.address, vesting.FundVestingAccountMsgURL)
					Expect(err).ToNot(HaveOccurred(), "error while creating the generic authorization: %v", err)
				}
				s.CreateTestClawbackVestingAccount(s.address, toAddr)

				vestAccInitialBal := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)

				fundVestingAccArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						s.address,
						toAddr,
						uint64(time.Now().Unix()),
						defaultPeriods,
						emptyPeriods,
					)

				fundClawbackVestingCheck := passCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, fundVestingAccArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				// Check the vesting account
				//
				// NOTE: The vesting account is created with the lockup periods only, since the vesting periods are empty.
				// The vesting periods are defaulted to instant vesting, i.e. period length = 0.
				s.ExpectVestingAccount(toAddr, defaultPeriods, instantPeriods)

				vestCoinsAmt := math.NewIntFromBigInt(defaultPeriods[0].Amount[0].Amount)
				vestAccFinalBal := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
				Expect(vestAccFinalBal.Amount).To(Equal(vestAccInitialBal.Amount.Add(vestCoinsAmt)))
			})

			It(fmt.Sprintf("should fund the vesting account from a smart contract when defining only lockup (%s)", callType.name), func() {
				if callType.directCall {
					Skip("this should only be run for smart contract calls")
				}
				s.CreateTestClawbackVestingAccount(contractAddr, toAddr)
				// send some funds to the smart contract
				// authorization to be able to fund from the smart contract is already in the setup
				err = evmosutil.FundAccount(s.ctx, s.app.BankKeeper, contractAddr.Bytes(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(100))))
				Expect(err).ToNot(HaveOccurred(), "error while funding the contract: %v", err)

				txSenderInitialBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
				vestAccInitialBal := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)

				// Build and execute the tx to fund the vesting account from a smart contract
				fundVestingAccArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						contractAddr,
						toAddr,
						uint64(time.Now().Unix()),
						defaultPeriods,
						emptyPeriods,
					).
					WithGasPrice(gasPrice.BigInt())

				fundClawbackVestingCheck := passCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, fundVestingAccArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				fees := gasPrice.MulRaw(res.GasUsed)

				// Check the vesting account
				//
				// NOTE: The vesting account is created with the lockup periods only, since the vesting periods are empty.
				// The vesting periods are defaulted to instant vesting, i.e. period length = 0.
				s.ExpectVestingAccount(toAddr, defaultPeriods, instantPeriods)

				// check that tx signer's balance is reduced by the fees paid
				txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
				Expect(txSenderFinalBal.Amount).To(Equal(txSenderInitialBal.Amount.Sub(fees)))

				// check the contract's balance was deducted to fund the vesting account
				contractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
				Expect(contractFinalBal.Amount).To(Equal(sdk.ZeroInt()))

				vestCoinsAmt := math.NewIntFromBigInt(defaultPeriods[0].Amount[0].Amount)
				vestAccFinalBal := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
				Expect(vestAccFinalBal.Amount).To(Equal(vestAccInitialBal.Amount.Add(vestCoinsAmt)))
			})

			It(fmt.Sprintf("contract that calls funder - should NOT fund the vesting account with a smart contract different than the contract that calls the precompile (%s)", callType.name), func() {
				if callType.directCall {
					Skip("this should only be run for smart contract calls")
				}
				counterContract, err := contracts.LoadCounterContract()
				Expect(err).ToNot(HaveOccurred())

				funderContractAddr, err := s.DeployContract(counterContract)
				Expect(err).ToNot(HaveOccurred(), "error while deploying the smart contract: %v", err)

				s.CreateTestClawbackVestingAccount(funderContractAddr, toAddr)
				// send some funds to the smart contract
				// authorization to be able to fund from the smart contract is already in the setup
				funderContractInitialAmt := math.NewInt(200)
				err = evmosutil.FundAccount(s.ctx, s.app.BankKeeper, funderContractAddr.Bytes(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, funderContractInitialAmt)))
				Expect(err).ToNot(HaveOccurred(), "error while funding the contract: %v", err)

				// create authorization for tx sender to use funder's balance to fund a vesting account
				err = vesting.CreateGenericAuthz(s.ctx, s.app.AuthzKeeper, s.address, funderContractAddr, vesting.FundVestingAccountMsgURL)
				Expect(err).ToNot(HaveOccurred(), "error while creating the generic authorization: %v", err)

				txSenderInitialBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)

				// Build and execute the tx to fund the vesting account from a smart contract
				fundVestingAccArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						funderContractAddr,
						toAddr,
						uint64(time.Now().Unix()),
						defaultPeriods,
						emptyPeriods,
					)
				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, fundVestingAccArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred())

				// check that tx signer's balance is reduced by the fees paid
				txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
				Expect(txSenderFinalBal.Amount.LTE(txSenderInitialBal.Amount)).To(BeTrue())

				// the balance of the contract that calls the precompile should remain 0
				contractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
				Expect(contractFinalBal.Amount).To(Equal(sdk.ZeroInt()))

				// the balance of the funder contract should remain unchanged
				funderContractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, funderContractAddr.Bytes(), s.bondDenom)
				Expect(funderContractFinalBal.Amount).To(Equal(funderContractInitialAmt))
			})

			It(fmt.Sprintf("fund using a third party EOA - should NOT fund the vesting account even if authorized (%s)", callType.name), func() {
				if callType.directCall {
					Skip("this should only be run for smart contract calls")
				}

				// send some funds to the funder, a third party EOA
				funderAccAddr, _ := testutiltx.NewAccAddressAndKey()
				funderHexAddr := common.BytesToAddress(funderAccAddr)
				initialFunderBalance := math.NewInt(200)
				err = evmosutil.FundAccount(s.ctx, s.app.BankKeeper, funderAccAddr, sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, initialFunderBalance)))
				Expect(err).ToNot(HaveOccurred(), "error while funding the third party EOA: %v", err)

				// create authorization for tx sender to use funder's balance to fund a vesting account
				err = vesting.CreateGenericAuthz(s.ctx, s.app.AuthzKeeper, s.address, funderHexAddr, vesting.FundVestingAccountMsgURL)
				Expect(err).ToNot(HaveOccurred(), "error while creating the generic authorization: %v", err)

				// create clawback vesting account with the corresponding funder
				s.CreateTestClawbackVestingAccount(funderHexAddr, toAddr)

				txSenderInitialBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)

				// Build and execute the tx to fund the vesting account from a third party EOA
				fundVestingAccArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						funderHexAddr,
						toAddr,
						uint64(time.Now().Unix()),
						defaultPeriods,
						emptyPeriods,
					).
					WithGasPrice(gasPrice.BigInt())

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, fundVestingAccArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred())

				// check that tx signer's balance is reduced by the fees paid
				txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
				Expect(txSenderFinalBal.Amount.LTE(txSenderInitialBal.Amount)).To(BeTrue())

				// check the funders's balance remains unchanged
				funderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, funderAccAddr, s.bondDenom)
				Expect(funderFinalBal.Amount).To(Equal(initialFunderBalance))
			})

			It(fmt.Sprintf("should NOT fund the vesting with tx origin funds when calling the precompile from a smart contract and WITHOUT authorization (%s)", callType.name), func() {
				// when calling from a smart contract
				// the funder (s.address) needs to authorize
				// for the smart contract to use its funds
				if callType.directCall {
					Skip("this should only be run for smart contract calls")
				}
				s.CreateTestClawbackVestingAccount(s.address, toAddr)
				fundVestingAccArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						s.address,
						toAddr,
						uint64(time.Now().Unix()),
						emptyPeriods,
						defaultPeriods,
					)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, fundVestingAccArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should fund the vesting when defining only vesting (%s)", callType.name), func() {
				// when calling from a smart contract
				// the funder (s.address) needs to authorize
				// for the smart contract to use its funds
				if !callType.directCall {
					// create authorization to allow contract to spend the funder's (s.address) balance
					// when funding a vesting account
					err = vesting.CreateGenericAuthz(s.ctx, s.app.AuthzKeeper, contractAddr, s.address, vesting.FundVestingAccountMsgURL)
					Expect(err).ToNot(HaveOccurred(), "error while creating the generic authorization: %v", err)
				}

				s.CreateTestClawbackVestingAccount(s.address, toAddr)
				fundVestingAccArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						s.address,
						toAddr,
						uint64(time.Now().Unix()),
						emptyPeriods,
						defaultPeriods,
					)

				fundClawbackVestingCheck := passCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, fundVestingAccArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				// Check the vesting account
				//
				// NOTE: The vesting account is created with the vesting periods only, since the lockup periods are empty.
				// The lockup periods are defaulted to instant unlocking, i.e. period length = 0.
				s.ExpectVestingAccount(toAddr, instantPeriods, defaultPeriods)
			})

			Context("Table-driven tests for Withdraw Delegator Rewards", func() {
				// testCase is a struct used for cases of contracts calls that have some operation
				// performed before and/or after the precompile call
				type testCase struct {
					transferTo *common.Address
					before     bool
					after      bool
				}

				var (
					args                   contracts.CallArgs
					funderInitialBal       sdk.Coin
					vestingAccInitialBal   sdk.Coin
					contractInitialBalance = math.NewInt(100)
				)

				BeforeEach(func() {
					args = s.BuildCallArgs(callType, contractAddr).
						WithMethodName("fundVestingAccountAndTransfer").
						WithGasPrice(gasPrice.BigInt())

					// send some funds to the contract
					err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, contractAddr.Bytes(), contractInitialBalance.Int64())
					Expect(err).To(BeNil())

					s.CreateTestClawbackVestingAccount(s.address, toAddr)

					funderInitialBal = s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
					vestingAccInitialBal = s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)

					// create authorization to allow contract to spend the funder's (s.address) balance
					// when funding a vesting account
					err = vesting.CreateGenericAuthz(s.ctx, s.app.AuthzKeeper, contractAddr, s.address, vesting.FundVestingAccountMsgURL)
					Expect(err).ToNot(HaveOccurred(), "error while creating the generic authorization: %v", err)
				})

				DescribeTable(fmt.Sprintf("should fund the vesting account from tx origin when defining only vesting (%s)", callType.name), func(tc testCase) {
					if callType.directCall {
						Skip("this should only be run for smart contract calls")
					}
					// if not specified, default the transferTo to the funder address
					if tc.transferTo == nil {
						tc.transferTo = &s.address
					}
					fundVestingAccArgs := args.
						WithArgs(
							s.address,
							toAddr,
							*tc.transferTo,
							uint64(time.Now().Unix()),
							emptyPeriods,
							defaultPeriods,
							tc.before, tc.after, // transfer funds to the funder according to test case
						).
						WithGasPrice(gasPrice.BigInt())

					fundClawbackVestingCheck := passCheck.
						WithExpEvents(vesting.EventTypeFundVestingAccount)

					res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, fundVestingAccArgs, fundClawbackVestingCheck)
					Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

					fees := gasPrice.MulRaw(res.GasUsed)

					transferredToAmt := math.ZeroInt()
					for _, transferred := range []bool{tc.before, tc.after} {
						if transferred {
							transferredToAmt = transferredToAmt.AddRaw(15)
						}
					}
					// Check the vesting account
					//
					// NOTE: The vesting account is created with the vesting periods only, since the lockup periods are empty.
					// The lockup periods are defaulted to instant unlocking, i.e. period length = 0.
					s.ExpectVestingAccount(toAddr, instantPeriods, defaultPeriods)

					// check the contract's balance was deducted to fund the vesting account
					contractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
					Expect(contractFinalBal.Amount).To(Equal(contractInitialBalance.Sub(transferredToAmt)))

					// check that the vesting account received the funds
					vestCoinsAmt := math.NewIntFromBigInt(defaultPeriods[0].Amount[0].Amount)
					vestAccFinalBal := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					expVestAccFinalBal := vestingAccInitialBal.Amount.Add(vestCoinsAmt)
					if *tc.transferTo == toAddr {
						expVestAccFinalBal = expVestAccFinalBal.Add(transferredToAmt)
					}

					Expect(vestAccFinalBal.Amount).To(Equal(expVestAccFinalBal))

					// check that funder balance is reduced by the fees paid, the amt to fund the vesting account,
					// but also got the funds sent from the contract (when corresponds)
					funderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)

					expFunderFinalBal := funderInitialBal.Amount.Sub(fees).Sub(vestCoinsAmt)
					if *tc.transferTo == s.address {
						expFunderFinalBal = expFunderFinalBal.Add(transferredToAmt)
					}

					Expect(funderFinalBal.Amount).To(Equal(expFunderFinalBal))
				},
					Entry("funder balance change before & after precompile call", testCase{
						before: true,
						after:  true,
					}),
					Entry("funder balance change before precompile call", testCase{
						before: true,
						after:  false,
					}),
					Entry("funder balance change after precompile call", testCase{
						before: false,
						after:  true,
					}),
					Entry("vesting acc balance change before & after precompile call", testCase{
						transferTo: &toAddr,
						before:     true,
						after:      true,
					}),
					Entry("vesting acc balance change before precompile call", testCase{
						transferTo: &toAddr,
						before:     true,
						after:      false,
					}),
					Entry("vesting acc balance change after precompile call", testCase{
						transferTo: &toAddr,
						before:     false,
						after:      true,
					}),
				)
			})

			It(fmt.Sprintf("should fund the vesting account from a smart contract when defining only vesting (%s)", callType.name), func() { //nolint:dupl
				if callType.directCall {
					Skip("this should only be run for smart contract calls")
				}
				s.CreateTestClawbackVestingAccount(contractAddr, toAddr)
				// send some funds to the smart contract
				// authorization is already created in the test setup
				err = evmosutil.FundAccount(s.ctx, s.app.BankKeeper, contractAddr.Bytes(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(100))))
				Expect(err).ToNot(HaveOccurred(), "error while funding the contract: %v", err)

				txSenderInitialBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)

				// Build and execute the tx to fund the vesting account from a smart contract
				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						contractAddr,
						toAddr,
						uint64(time.Now().Unix()),
						emptyPeriods,
						defaultPeriods,
					).
					WithGasPrice(gasPrice.BigInt())

				fundClawbackVestingCheck := passCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				fees := gasPrice.MulRaw(res.GasUsed)

				// Check the vesting account
				s.ExpectVestingAccount(toAddr, instantPeriods, defaultPeriods)

				// check that tx signer's balance is reduced by the fees paid
				txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
				Expect(txSenderFinalBal.Amount).To(Equal(txSenderInitialBal.Amount.Sub(fees)))

				// check the contract's balance was deducted to fund the vesting account
				contractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
				Expect(contractFinalBal.Amount).To(Equal(sdk.ZeroInt()))
			})

			It(fmt.Sprintf("should fund the vesting when defining both lockup and vesting (%s)", callType.name), func() {
				s.CreateTestClawbackVestingAccount(s.address, toAddr)
				// funder is an EOA (s.address)
				// if calling funding via contract call,
				// need auth from funder addr (s.address)
				// to the contract address
				if !callType.directCall {
					err = vesting.CreateGenericAuthz(s.ctx, s.app.AuthzKeeper, contractAddr, s.address, vesting.FundVestingAccountMsgURL)
					Expect(err).ToNot(HaveOccurred(), "error while creating the generic authorization: %v", err)
				}
				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						s.address,
						toAddr,
						uint64(time.Now().Unix()),
						defaultPeriods,
						defaultPeriods,
					)

				fundClawbackVestingCheck := passCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				// Check the vesting account
				s.ExpectVestingAccount(toAddr, defaultPeriods, defaultPeriods)
			})

			It(fmt.Sprintf("should fund the vesting account from a smart contract when defining both lockup and vesting  (%s)", callType.name), func() { //nolint:dupl
				if callType.directCall {
					Skip("this should only be run for smart contract calls")
				}
				s.CreateTestClawbackVestingAccount(contractAddr, toAddr)
				// send some funds to the smart contract
				err = evmosutil.FundAccount(s.ctx, s.app.BankKeeper, contractAddr.Bytes(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(100))))
				Expect(err).ToNot(HaveOccurred(), "error while funding the contract: %v", err)

				txSenderInitialBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)

				// Build and execute the tx to fund the vesting account from a smart contract
				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						contractAddr,
						toAddr,
						uint64(time.Now().Unix()),
						defaultPeriods,
						defaultPeriods,
					).
					WithGasPrice(gasPrice.BigInt())

				fundClawbackVestingCheck := passCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				fees := gasPrice.MulRaw(res.GasUsed)

				// Check the vesting account
				s.ExpectVestingAccount(toAddr, defaultPeriods, defaultPeriods)

				// check that tx signer's balance is reduced by the fees paid
				txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
				Expect(txSenderFinalBal.Amount).To(Equal(txSenderInitialBal.Amount.Sub(fees)))

				// check the contract's balance was deducted to fund the vesting account
				contractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
				Expect(contractFinalBal.Amount).To(Equal(sdk.ZeroInt()))
			})

			It(fmt.Sprintf("should not fund the vesting when defining different total coins for lockup and vesting (%s)", callType.name), func() {
				s.CreateTestClawbackVestingAccount(s.address, toAddr)

				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						s.address,
						toAddr,
						uint64(time.Now().Unix()),
						defaultPeriods,
						doublePeriods,
					)

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.WithErrContains("vesting and lockup schedules must have same total coins")
				}

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("vesting and lockup schedules must have same total coins"))
				}
				// Check the vesting account
				acc := s.app.AccountKeeper.GetAccount(s.ctx, vestingAddr.Bytes())
				Expect(acc).To(BeNil(), "vesting account should not exist")
			})

			It(fmt.Sprintf("should not fund the vesting when defining neither lockup nor vesting (%s)", callType.name), func() {
				s.CreateTestClawbackVestingAccount(s.address, toAddr)

				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						s.address,
						vestingAddr,
						uint64(time.Now().Unix()),
						emptyPeriods,
						emptyPeriods,
					)

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.WithErrContains("vesting and/or lockup schedules must be present")
				}

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("vesting and/or lockup schedules must be present"))
				}

				// Check the vesting account
				acc := s.app.AccountKeeper.GetAccount(s.ctx, vestingAddr.Bytes())
				Expect(acc).To(BeNil(), "vesting account should not exist")
			})

			It(fmt.Sprintf("should not fund the vesting when exceeding the funder balance (%s)", callType.name), func() {
				s.CreateTestClawbackVestingAccount(s.address, toAddr)

				balance := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
				exceededBalance := new(big.Int).Add(big.NewInt(1), balance.Amount.BigInt())

				exceedingVesting := []vesting.Period{{
					Length: 10,
					Amount: []cmn.Coin{{Denom: s.bondDenom, Amount: exceededBalance}},
				}}

				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						s.address,
						toAddr,
						uint64(time.Now().Unix()),
						exceedingVesting,
						emptyPeriods,
					)

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.WithErrContains("insufficient funds")
				}

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("insufficient funds"))
				}

				// Check the vesting account
				acc := s.app.AccountKeeper.GetAccount(s.ctx, toAddr.Bytes())
				va := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(va.LockupPeriods).To(BeNil(), "vesting account should not be funded")
				Expect(va.VestingPeriods).To(BeNil(), "vesting account should not be funded")
			})

			It(fmt.Sprintf("should not fund the vesting when not sending as the funder (%s)", callType.name), func() {
				s.CreateTestClawbackVestingAccount(s.address, toAddr)
				differentFunder := testutiltx.GenerateAddress()

				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						differentFunder,
						vestingAddr,
						uint64(time.Now().Unix()),
						defaultPeriods,
						defaultPeriods,
					)

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.
						WithErrContains(
							fmt.Sprintf("tx origin address %s does not match the from address %s",
								s.address,
								differentFunder,
							),
						)
				}

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("does not match the from address"))
				}

				// Check the vesting account
				acc := s.app.AccountKeeper.GetAccount(s.ctx, vestingAddr.Bytes())
				Expect(acc).To(BeNil(), "vesting account should not exist")
			})

			It(fmt.Sprintf("should not fund the vesting when the address is blocked (%s)", callType.name), func() {
				s.CreateTestClawbackVestingAccount(s.address, toAddr)
				moduleAddr := common.BytesToAddress(authtypes.NewModuleAddress("distribution").Bytes())

				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						s.address,
						moduleAddr,
						uint64(time.Now().Unix()),
						defaultPeriods,
						defaultPeriods,
					)

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.WithErrContains("is not allowed to receive funds")
				}

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
				Expect(err).To(HaveOccurred(), "error while creating a clawback vesting account for a module address", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("is not allowed to receive funds"))
				}

				// check that the module address is not a vesting account
				acc := s.app.AccountKeeper.GetAccount(s.ctx, moduleAddr.Bytes())
				Expect(acc).ToNot(BeNil(), "module account should be found")
				_, ok := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(ok).To(BeFalse(), "module account should not be a vesting account")
			})

			It(fmt.Sprintf("should not fund the vesting when the address is blocked - a precompile address (%s)", callType.name), func() {
				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						s.address,
						s.precompile.Address(),
						uint64(time.Now().Unix()),
						defaultPeriods,
						defaultPeriods,
					)

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.WithErrContains("is not allowed to receive funds")
				}

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
				Expect(err).To(HaveOccurred(), "error while creating a clawback vesting account for a module address", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("is not allowed to receive funds"))
				}
			})

			It(fmt.Sprintf("should not fund the vesting when the address is uninitialized (%s)", callType.name), func() {
				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						s.address,
						differentAddr,
						uint64(time.Now().Unix()),
						defaultPeriods,
						defaultPeriods,
					)

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.WithErrContains("does not exist")
				}

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
				Expect(err).To(HaveOccurred(), "error while creating a clawback vesting account for a module address", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("does not exist"))
				}
			})

			It(fmt.Sprintf("should not fund the vesting when the address is the zero address (%s)", callType.name), func() {
				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.FundVestingAccountMethod).
					WithArgs(
						s.address,
						common.Address{},
						uint64(time.Now().Unix()),
						defaultPeriods,
						defaultPeriods,
					)

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.WithErrContains("is not allowed to receive funds")
				}

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
				Expect(err).To(HaveOccurred(), "error while creating a clawback vesting account for the zero address", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("vesting address cannot be the zero address"))
				}
			})
		}
	})

	Context("to claw back from a vesting account", func() {
		expClawbackAmt := math.NewInt(1000)

		BeforeEach(func() {
			s.CreateTestClawbackVestingAccount(s.address, toAddr)
			s.FundTestClawbackVestingAccount()
		})

		for _, callType := range callTypes {
			callType := callType

			Context("without authorization", func() {
				It(fmt.Sprintf("should NOT claw back from the vesting when sending tx from the funder (%s)", callType.name), func() {
					if callType.directCall {
						Skip("this should only be run for smart contract calls")
					}
					balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")

					clawbackArgs := s.BuildCallArgs(callType, contractAddr).
						WithMethodName(vesting.ClawbackMethod).
						WithArgs(
							s.address,
							toAddr,
							differentAddr,
						)

					_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, execRevertedCheck)
					Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)

					balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePost.Amount).To(Equal(balancePre.Amount))
					balanceReceiver := s.app.BankKeeper.GetBalance(s.ctx, differentAddr.Bytes(), s.bondDenom)
					Expect(balanceReceiver.Amount).To(Equal(math.ZeroInt()))
				})
			})

			Context("with authorization", func() {
				BeforeEach(func() {
					if callType.directCall == false {
						err = vesting.CreateGenericAuthz(s.ctx, s.app.AuthzKeeper, contractAddr, s.address, vesting.ClawbackMsgURL)
						Expect(err).ToNot(HaveOccurred(), "error while creating the generic authorization: %v", err)
					}
				})

				It(fmt.Sprintf("should claw back from the vesting when sending as the funder (%s)", callType.name), func() {
					balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")

					clawbackArgs := s.BuildCallArgs(callType, contractAddr).
						WithMethodName(vesting.ClawbackMethod).
						WithArgs(
							s.address,
							toAddr,
							differentAddr,
						)

					clawbackCheck := passCheck.
						WithExpEvents(vesting.EventTypeClawback)

					_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, clawbackCheck)
					Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

					var co vesting.ClawbackOutput
					err = s.precompile.UnpackIntoInterface(&co, vesting.ClawbackMethod, ethRes.Ret)
					Expect(err).ToNot(HaveOccurred(), "error while unpacking the clawback output: %v", err)
					Expect(co.Coins).To(Equal(balances), "expected different clawback amount")

					balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePost.Amount.Int64()).To(Equal(int64(100)), "expected only initial balance after clawback")
					balanceReceiver := s.app.BankKeeper.GetBalance(s.ctx, differentAddr.Bytes(), s.bondDenom)
					Expect(balanceReceiver.Amount).To(Equal(math.NewInt(1000)), "expected receiver to show different balance after clawback")
				})

				Context("table tests for clawback with state changes", func() {
					type testCase struct {
						dest       common.Address
						transferTo *common.Address
						before     bool
						after      bool
					}
					DescribeTable(fmt.Sprintf("smart contract as funder - contract with state changes on destination address - should claw back from the vesting when sending as the funder (%s)", callType.name), func(tc testCase) {
						if callType.directCall {
							Skip("this should only be run for smart contract calls")
						}
						if tc.transferTo == nil {
							tc.transferTo = &tc.dest
						}
						// change the vesting account funder to be the contract
						_, err := s.app.VestingKeeper.UpdateVestingFunder(s.ctx, &vestingtypes.MsgUpdateVestingFunder{
							FunderAddress:    sdk.AccAddress(s.address.Bytes()).String(),
							NewFunderAddress: sdk.AccAddress(contractAddr.Bytes()).String(),
							VestingAddress:   sdk.AccAddress(toAddr.Bytes()).String(),
						})
						Expect(err).ToNot(HaveOccurred())

						// fund the contract to make internal transfers
						contractInitialBalance := math.NewInt(100)
						err = evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, contractAddr.Bytes(), contractInitialBalance.Int64())
						Expect(err).ToNot(HaveOccurred(), "error while funding the contract: %v", err)

						vestAccInitialBal := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
						Expect(vestAccInitialBal.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")

						clawbackArgs := s.BuildCallArgs(callType, contractAddr).
							WithMethodName("clawbackWithTransfer").
							WithArgs(
								contractAddr,
								toAddr,
								tc.dest,
								*tc.transferTo,
								tc.before,
								tc.after,
							)

						clawbackCheck := passCheck.
							WithExpEvents(vesting.EventTypeClawback)

						_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, clawbackCheck)
						Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

						var co vesting.ClawbackOutput
						err = s.precompile.UnpackIntoInterface(&co, vesting.ClawbackMethod, ethRes.Ret)
						Expect(err).ToNot(HaveOccurred(), "error while unpacking the clawback output: %v", err)
						Expect(co.Coins).To(Equal(balances), "expected different clawback amount")

						contractTransferredAmt := math.ZeroInt()
						for _, transferred := range []bool{tc.before, tc.after} {
							if transferred {
								contractTransferredAmt = contractTransferredAmt.AddRaw(15)
							}
						}

						vestAccFinalBalance := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
						expVestAccFinalBal := vestAccInitialBal.Amount.Sub(expClawbackAmt)
						if *tc.transferTo == toAddr {
							expVestAccFinalBal = expVestAccFinalBal.Add(contractTransferredAmt)
						}
						Expect(vestAccFinalBalance.Amount).To(Equal(expVestAccFinalBal), "expected only initial balance after clawback")

						// contract transfers balances when it is not the destination
						if tc.dest == contractAddr {
							contractFinalBalance := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
							Expect(contractFinalBalance.Amount).To(Equal(contractInitialBalance.Add(expClawbackAmt)))
							return
						}

						balanceDest := s.app.BankKeeper.GetBalance(s.ctx, tc.dest.Bytes(), s.bondDenom)
						expBalDest := expClawbackAmt
						if *tc.transferTo == tc.dest {
							expBalDest = expBalDest.Add(contractTransferredAmt)
						}
						Expect(balanceDest.Amount).To(Equal(expBalDest), "expected receiver to show different balance after clawback")

						contractFinalBalance := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
						Expect(contractFinalBalance.Amount).To(Equal(contractInitialBalance.Sub(contractTransferredAmt)))
					},
						Entry("funder is the destination address - state changes before & after precompile call", testCase{
							dest:   contractAddr,
							before: true,
							after:  true,
						}),
						Entry("funder is the destination address - state changes before precompile call", testCase{
							dest:   contractAddr,
							before: true,
							after:  false,
						}),
						Entry("funder is the destination address - state changes after precompile call", testCase{
							dest:   contractAddr,
							before: false,
							after:  true,
						}),
						Entry("another address is the destination address - state changes before & after precompile", testCase{
							dest:   differentAddr,
							before: true,
							after:  true,
						}),
						Entry("another address is the destination address - state changes before precompile", testCase{
							dest:   differentAddr,
							before: true,
							after:  false,
						}),
						Entry("another address is the destination address - state changes after precompile", testCase{
							dest:   differentAddr,
							before: false,
							after:  true,
						}),
						Entry("another address is the destination address - transfer to vest acc before & after precompile", testCase{
							dest:       differentAddr,
							transferTo: &toAddr,
							before:     true,
							after:      true,
						}),
						Entry("another address is the destination address - transfer to vest acc before precompile", testCase{
							dest:       differentAddr,
							transferTo: &toAddr,
							before:     true,
							after:      false,
						}),
						Entry("another address is the destination address - transfer to vest acc after precompile", testCase{
							dest:       differentAddr,
							transferTo: &toAddr,
							before:     false,
							after:      true,
						}),
					)
				})

				It(fmt.Sprintf("should claw back from the vesting when sending as the funder with the caller smart contract as destination for the clawed back funds (%s)", callType.name), func() {
					balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")

					// check the contract's (destination) initial balance. Should be 0
					contractInitialBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
					Expect(contractInitialBal.Amount).To(Equal(sdk.ZeroInt()))

					// get tx sender initial balance
					txSenderInitialBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)

					clawbackArgs := s.BuildCallArgs(callType, contractAddr).
						WithMethodName(vesting.ClawbackMethod).
						WithArgs(
							s.address,
							toAddr,
							contractAddr,
						).
						WithGasPrice(gasPrice.BigInt())

					clawbackCheck := passCheck.
						WithExpEvents(vesting.EventTypeClawback)

					res, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, clawbackCheck)
					Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

					fees := gasPrice.MulRaw(res.GasUsed)

					var co vesting.ClawbackOutput
					err = s.precompile.UnpackIntoInterface(&co, vesting.ClawbackMethod, ethRes.Ret)
					Expect(err).ToNot(HaveOccurred(), "error while unpacking the clawback output: %v", err)
					Expect(co.Coins).To(Equal(balances), "expected different clawback amount")

					// check clawback account balance
					balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePost.Amount.Int64()).To(Equal(int64(100)), "expected only initial balance after clawback")

					// check that tx signer's balance is reduced by the fees paid
					txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
					Expect(txSenderFinalBal.Amount).To(Equal(txSenderInitialBal.Amount.Sub(fees)))

					// check contract's final balance (clawback destination)
					contractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
					Expect(contractFinalBal.Amount).To(Equal(math.NewInt(1000)), "expected receiver to show different balance after clawback")
				})

				It(fmt.Sprintf("clawback with revert after precompile call but before changing contract state - should NOT claw back and revert all balances to initial values (%s)", callType.name), func() { //nolint:dupl
					if callType.directCall {
						Skip("this should only be run for smart contract calls")
					}
					balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")

					clawbackArgs := s.BuildCallArgs(callType, contractAddr).
						WithMethodName("clawbackWithRevert").
						WithArgs(
							s.address,
							toAddr,
							differentAddr,
							true,
						)

					_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, execRevertedCheck)
					Expect(err).To(HaveOccurred())

					balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePost.Amount).To(Equal(balancePre.Amount), "expected no balance change")
					balanceReceiver := s.app.BankKeeper.GetBalance(s.ctx, differentAddr.Bytes(), s.bondDenom)
					Expect(balanceReceiver.Amount).To(Equal(math.ZeroInt()))
				})

				It(fmt.Sprintf("clawback with revert after precompile after changing contract state - should NOT claw back and revert all balances to initial values (%s)", callType.name), func() { //nolint:dupl
					if callType.directCall {
						Skip("this should only be run for smart contract calls")
					}
					balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")

					clawbackArgs := s.BuildCallArgs(callType, contractAddr).
						WithMethodName("clawbackWithRevert").
						WithArgs(
							s.address,
							toAddr,
							differentAddr,
							false,
						)

					_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, execRevertedCheck)
					Expect(err).To(HaveOccurred())

					balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePost.Amount).To(Equal(balancePre.Amount), "expected no balance change")
					balanceReceiver := s.app.BankKeeper.GetBalance(s.ctx, differentAddr.Bytes(), s.bondDenom)
					Expect(balanceReceiver.Amount).To(Equal(math.ZeroInt()))
				})

				It(fmt.Sprintf("another contract as destination - should clawback from the vesting when sending as the funder with another smart contract as destination for the clawed back funds (%s)", callType.name), func() {
					counterContract, err := contracts.LoadCounterContract()
					Expect(err).ToNot(HaveOccurred())

					destContractAddr, err := s.DeployContract(counterContract)
					Expect(err).ToNot(HaveOccurred(), "error while deploying the smart contract: %v", err)

					balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")

					// check the contract's (destination) initial balance. Should be 0
					destContractInitialBal := s.app.BankKeeper.GetBalance(s.ctx, destContractAddr.Bytes(), s.bondDenom)
					Expect(destContractInitialBal.Amount).To(Equal(sdk.ZeroInt()))

					// get tx sender initial balance
					txSenderInitialBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)

					clawbackCheck := passCheck.
						WithExpEvents(vesting.EventTypeClawback)

					clawbackArgs := s.BuildCallArgs(callType, contractAddr).
						WithMethodName(vesting.ClawbackMethod).
						WithArgs(
							s.address,
							toAddr,
							destContractAddr,
						).
						WithGasPrice(gasPrice.BigInt())

					res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, clawbackCheck)
					Expect(err).NotTo(HaveOccurred())
					fees := gasPrice.MulRaw(res.GasUsed)

					// check clawback account balance
					balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePost.Amount).To(Equal(balancePre.Amount.Sub(expClawbackAmt)))

					// check that tx signer's balance is reduced by the fees paid
					txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
					Expect(txSenderFinalBal.Amount).To(Equal(txSenderInitialBal.Amount.Sub(fees)))

					// check caller contract's final balance should be zero
					callerContractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
					Expect(callerContractFinalBal.Amount).To(Equal(math.ZeroInt()))

					// check destination contract's final balance should
					// have received the clawback amt
					destContractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, destContractAddr.Bytes(), s.bondDenom)
					Expect(destContractFinalBal.Amount).To(Equal(destContractInitialBal.Amount.Add(expClawbackAmt)))
				})

				It(fmt.Sprintf("another contract as destination - should claw back from the vesting when sending as the funder with another smart contract as destination and triggering state change on destination contract (%s)", callType.name), func() {
					if callType.directCall {
						Skip("this should only be run for smart contract calls")
					}
					counterContract, err := contracts.LoadCounterContract()
					Expect(err).ToNot(HaveOccurred())

					destContractAddr, err := s.DeployContract(counterContract)
					Expect(err).ToNot(HaveOccurred(), "error while deploying the smart contract: %v", err)

					balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")

					// check the contract's (destination) initial balance. Should be 0
					destContractInitialBal := s.app.BankKeeper.GetBalance(s.ctx, destContractAddr.Bytes(), s.bondDenom)
					Expect(destContractInitialBal.Amount).To(Equal(sdk.ZeroInt()))

					// get tx sender initial balance
					txSenderInitialBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)

					clawbackArgs := s.BuildCallArgs(callType, contractAddr).
						WithMethodName("clawbackWithCounterCall").
						WithArgs(
							s.address,
							toAddr,
							destContractAddr,
						).
						WithGasPrice(gasPrice.BigInt())

					// expect the vesting precompile events and the Counter
					// contract's events
					clawbackCheck := passCheck.
						WithABIEvents(mergeEventMaps(
							s.precompile.Events,
							counterContract.ABI.Events,
						)).
						WithExpEvents([]string{
							"Added", "Changed",
							vesting.EventTypeClawback,
							"Changed",
						}...)

					res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, clawbackCheck)
					Expect(err).NotTo(HaveOccurred())
					fees := gasPrice.MulRaw(res.GasUsed)

					// check clawback account balance
					balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePost.Amount).To(Equal(balancePre.Amount.Sub(expClawbackAmt)))

					// check that tx signer's balance is reduced by the fees paid
					txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
					Expect(txSenderFinalBal.Amount).To(Equal(txSenderInitialBal.Amount.Sub(fees)))

					// check caller contract's final balance should be zero
					callerContractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
					Expect(callerContractFinalBal.Amount).To(Equal(math.ZeroInt()))

					// check destination contract's final balance should
					// have received the clawback amt
					destContractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, destContractAddr.Bytes(), s.bondDenom)
					Expect(destContractFinalBal.Amount).To(Equal(destContractInitialBal.Amount.Add(expClawbackAmt)))
				})

				It(fmt.Sprintf("should return an error when not sending as the funder (%s)", callType.name), func() {
					// create and fund new account
					differentAddr, differentPriv := testutiltx.NewAddrKey()
					err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, differentAddr.Bytes(), 1e18)
					Expect(err).ToNot(HaveOccurred(), "error while funding the account: %v", err)

					balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")

					clawbackArgs := s.BuildCallArgs(callType, contractAddr).
						WithMethodName(vesting.ClawbackMethod).
						WithPrivKey(differentPriv).
						WithArgs(
							s.address,
							toAddr,
							differentAddr,
						)

					clawbackCheck := execRevertedCheck
					if callType.directCall {
						clawbackCheck = failCheck.
							WithErrContains(fmt.Sprintf(
								"tx origin address %s does not match the funder address %s",
								differentAddr, s.address,
							))
					}

					_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, clawbackCheck)
					Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
					if callType.directCall {
						Expect(err.Error()).To(ContainSubstring("does not match the funder address"))
					}
					balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePost).To(Equal(balancePre), "expected balance not to have changed")
				})

				It(fmt.Sprintf("should return an error when the vesting does not exist (%s)", callType.name), func() {
					// fund the new account
					err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, differentAddr.Bytes(), 1e18)
					Expect(err).ToNot(HaveOccurred(), "error while funding the account: %v", err)

					clawbackArgs := s.BuildCallArgs(callType, contractAddr).
						WithMethodName(vesting.ClawbackMethod).
						WithArgs(
							s.address,
							differentAddr,
							s.address,
						)

					clawbackCheck := execRevertedCheck
					if callType.directCall {
						clawbackCheck = failCheck.
							WithErrContains(vestingtypes.ErrNotSubjectToClawback.Error())
					}

					_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, clawbackCheck)
					Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
					if callType.directCall {
						Expect(err.Error()).To(ContainSubstring("%s: %s", sdk.AccAddress(differentAddr.Bytes()), vestingtypes.ErrNotSubjectToClawback.Error()))
					}
				})

				It(fmt.Sprintf("should succeed and return empty Coins when all tokens are vested (%s)", callType.name), func() {
					// commit block with time so that vesting has ended
					ctx, err := evmosutil.CommitAndCreateNewCtx(s.ctx, s.app, time.Hour*24, nil)
					Expect(err).ToNot(HaveOccurred(), "error while committing block: %v", err)
					s.ctx = ctx

					balancePre := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePre.Amount).To(Equal(math.NewInt(1100)), "expected different balance after setup")

					clawbackArgs := s.BuildCallArgs(callType, contractAddr).
						WithMethodName(vesting.ClawbackMethod).
						WithArgs(
							s.address,
							toAddr,
							s.address,
						)

					_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, clawbackArgs, passCheck)
					Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)

					var co vesting.ClawbackOutput
					err = s.precompile.UnpackIntoInterface(&co, vesting.ClawbackMethod, ethRes.Ret)
					Expect(err).ToNot(HaveOccurred(), "error while unpacking the clawback output: %v", err)
					Expect(co.Coins).To(BeEmpty(), "expected empty clawback amount")

					balancePost := s.app.BankKeeper.GetBalance(s.ctx, toAddr.Bytes(), s.bondDenom)
					Expect(balancePost).To(Equal(balancePre), "expected balance not to have changed")
				})
			})
		}
	})

	Context("to update the vesting funder", func() {
		BeforeEach(func() {
			s.CreateTestClawbackVestingAccount(s.address, toAddr)
		})

		for _, callType := range callTypes {
			callType := callType

			BeforeEach(func() {
				if callType.directCall == false {
					err = vesting.CreateGenericAuthz(s.ctx, s.app.AuthzKeeper, contractAddr, s.address, vesting.UpdateVestingFunderMsgURL)
					Expect(err).ToNot(HaveOccurred(), "error while creating the generic authorization: %v", err)
				}
			})

			It(fmt.Sprintf("should update the vesting funder when sending as the funder (%s)", callType.name), func() {
				updateFunderArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.UpdateVestingFunderMethod).
					WithArgs(
						s.address,
						differentAddr,
						toAddr,
					)

				updateFunderCheck := passCheck.
					WithExpEvents(vesting.EventTypeUpdateVestingFunder)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, updateFunderArgs, updateFunderCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				// Check that the vesting account has the new funder
				s.ExpectVestingFunder(toAddr, differentAddr)
			})

			It(fmt.Sprintf("should return an error when not sending as the funder (%s)", callType.name), func() {
				differentAddr, differentPriv := testutiltx.NewAddrKey()
				err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, differentAddr.Bytes(), 1e18)
				Expect(err).ToNot(HaveOccurred(), "error while funding the account: %v", err)

				updateFunderArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.UpdateVestingFunderMethod).
					WithPrivKey(differentPriv).
					WithArgs(
						s.address,
						differentAddr,
						toAddr,
					)

				updateFunderCheck := execRevertedCheck
				if callType.directCall {
					updateFunderCheck = failCheck.
						WithErrContains(fmt.Sprintf(
							"tx origin address %s does not match the funder address %s",
							differentAddr.String(), s.address.String(),
						))
				}

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, updateFunderArgs, updateFunderCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("does not match the funder address"))
				}
				// Check that the vesting account still has the same funder
				s.ExpectVestingFunder(toAddr, s.address)
			})

			It(fmt.Sprintf("should return an error when the account does not exist (%s)", callType.name), func() {
				// Check that there's no account
				nonExistentAddr := testutiltx.GenerateAddress()
				acc := s.app.AccountKeeper.GetAccount(s.ctx, nonExistentAddr.Bytes())
				Expect(acc).To(BeNil(), "expected no account to be found")

				updateFunderArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.UpdateVestingFunderMethod).
					WithArgs(
						s.address,
						differentAddr,
						nonExistentAddr, // the address of the vesting account
					)

				updateFunderCheck := execRevertedCheck
				if callType.directCall {
					updateFunderCheck = failCheck.
						WithErrContains(fmt.Sprintf(
							"account %s does not exist",
							sdk.AccAddress(nonExistentAddr.Bytes()).String(),
						))
				}

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, updateFunderArgs, updateFunderCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("does not exist"))
				}
			})

			It(fmt.Sprintf("should return an error when the account is no vesting account (%s)", callType.name), func() {
				// Check that there's no vesting account
				nonVestingAddr := testutiltx.GenerateAddress()
				err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, nonVestingAddr.Bytes(), 1e18)
				Expect(err).ToNot(HaveOccurred(), "error while funding the account: %v", err)
				acc := s.app.AccountKeeper.GetAccount(s.ctx, nonVestingAddr.Bytes())
				Expect(acc).ToNot(BeNil(), "expected account to be found")
				_, ok := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(ok).To(BeFalse(), "expected account not to be a vesting account")

				updateFunderArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.UpdateVestingFunderMethod).
					WithArgs(
						s.address,
						differentAddr,
						nonVestingAddr, // the address of the vesting account
					)

				updateFunderCheck := execRevertedCheck
				if callType.directCall {
					updateFunderCheck = failCheck.
						WithErrContains(fmt.Sprintf(
							"%s: %s",
							sdk.AccAddress(nonVestingAddr.Bytes()),
							vestingtypes.ErrNotSubjectToClawback.Error(),
						))
				}

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, updateFunderArgs, updateFunderCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(
						ContainSubstring("%s: %s",
							sdk.AccAddress(nonVestingAddr.Bytes()).String(),
							vestingtypes.ErrNotSubjectToClawback.Error(),
						),
					)
				}
			})

			It(fmt.Sprintf("should return an error when the new funder is the zero address (%s)", callType.name), func() {
				updateFunderArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.UpdateVestingFunderMethod).
					WithArgs(
						s.address,
						common.Address{},
						toAddr,
					)

				updateFunderCheck := execRevertedCheck
				if callType.directCall {
					updateFunderCheck = failCheck.
						WithErrContains("new funder address cannot be empty")
				}

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, updateFunderArgs, updateFunderCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should return an error when the new funder is the same as the current funder (%s)", callType.name), func() {
				updateFunderArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.UpdateVestingFunderMethod).
					WithArgs(
						s.address,
						s.address,
						toAddr,
					)

				updateFunderCheck := execRevertedCheck
				if callType.directCall {
					updateFunderCheck = failCheck.
						WithErrContains("new funder address is equal to current funder address")
				}

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, updateFunderArgs, updateFunderCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("new funder address is equal to current funder address"))
				}
				// Check that the vesting account still has the same funder
				s.ExpectVestingFunder(toAddr, s.address)
			})

			It(fmt.Sprintf("should return an error when the new funder is a blocked address (%s)", callType.name), func() {
				moduleAddr := common.BytesToAddress(authtypes.NewModuleAddress("distribution").Bytes())

				updateFunderArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.UpdateVestingFunderMethod).
					WithArgs(
						s.address,
						moduleAddr,
						toAddr,
					)

				updateFunderCheck := execRevertedCheck
				if callType.directCall {
					updateFunderCheck = failCheck.
						WithErrContains("is not allowed to receive funds")
				}

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, updateFunderArgs, updateFunderCheck)
				Expect(err).To(HaveOccurred(), "error while updating the funder to a module address: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("is a blocked address and not allowed to fund vesting accounts"))
				}
			})
		}
	})

	Context("to convert a vesting account", func() {
		BeforeEach(func() {
			// Create a vesting account
			s.CreateTestClawbackVestingAccount(s.address, toAddr)
		})

		for _, callType := range callTypes {
			callType := callType

			It(fmt.Sprintf("should convert the vesting account into a normal one after vesting has ended (%s)", callType.name), func() {
				// commit block with new time so that the vesting period has ended
				s.ctx, err = evmosutil.CommitAndCreateNewCtx(s.ctx, s.app, time.Hour*24, nil)
				Expect(err).To(BeNil(), "failed to commit block")

				convertClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.ConvertVestingAccountMethod).
					WithArgs(
						toAddr,
					)

				convertClawbackCheck := passCheck.
					WithExpEvents(vesting.EventTypeConvertVestingAccount)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, convertClawbackArgs, convertClawbackCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				// Check that the vesting account has been converted
				acc := s.app.AccountKeeper.GetAccount(s.ctx, toAddr.Bytes())
				Expect(acc).ToNot(BeNil(), "expected account to be found")
				_, ok := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(ok).To(BeFalse(), "expected account not to be a vesting account")
			})

			It(fmt.Sprintf("should return an error when the vesting has not ended yet (%s)", callType.name), func() {
				convertClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.ConvertVestingAccountMethod).
					WithArgs(
						toAddr,
					)

				convertClawbackCheck := execRevertedCheck
				if callType.directCall {
					convertClawbackCheck = failCheck.WithErrContains("vesting coins still left in account")
				}

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, convertClawbackArgs, convertClawbackCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should return an error when not sending as the funder (%s)", callType.name), func() {
				// commit block with new time so that the vesting period has ended
				s.ctx, err = evmosutil.CommitAndCreateNewCtx(s.ctx, s.app, time.Hour*24, nil)
				Expect(err).To(BeNil(), "failed to commit block")

				differentAddr, differentPriv := testutiltx.NewAddrKey()
				err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, differentAddr.Bytes(), 1e18)
				Expect(err).ToNot(HaveOccurred(), "error while funding account: %v", err)

				convertClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.ConvertVestingAccountMethod).
					WithPrivKey(differentPriv).
					WithArgs(
						toAddr,
					)

				convertClawbackCheck := execRevertedCheck
				if callType.directCall {
					convertClawbackCheck = failCheck.WithErrContains("sender is not the funder")
				}

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, convertClawbackArgs, convertClawbackCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should return an error when the vesting does not exist (%s)", callType.name), func() {
				err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, differentAddr.Bytes(), 1e18)
				Expect(err).ToNot(HaveOccurred(), "error while funding account: %v", err)

				convertClawbackArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.ConvertVestingAccountMethod).
					WithArgs(
						differentAddr, // this currently has no vesting
					)

				convertClawbackCheck := execRevertedCheck
				if callType.directCall {
					convertClawbackCheck = failCheck.WithErrContains("account not subject to vesting")
				}

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, convertClawbackArgs, convertClawbackCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)

				// Check that the account is no vesting account
				acc := s.app.AccountKeeper.GetAccount(s.ctx, differentAddr.Bytes())
				Expect(acc).ToNot(BeNil(), "expected account to be found")
				_, ok := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(ok).To(BeFalse(), "expected account not to be a vesting account")
			})
		}
	})

	// ---------------------------------------------
	//                     QUERIES
	//
	Context("to get vesting balances", func() {
		for _, callType := range callTypes {
			callType := callType

			It(fmt.Sprintf("should return the vesting when it exists (%s)", callType.name), func() {
				s.CreateTestClawbackVestingAccount(s.address, toAddr)
				s.FundTestClawbackVestingAccount()

				balancesArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.BalancesMethod).
					WithArgs(
						toAddr,
					)

				_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				var res vesting.BalancesOutput
				err = s.precompile.UnpackIntoInterface(&res, vesting.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "error while unpacking response: %v", err)

				expectedCoins := []cmn.Coin{{Denom: s.bondDenom, Amount: big.NewInt(1000)}}
				Expect(res.Locked).To(Equal(expectedCoins), "expected different locked coins")
				Expect(res.Unvested).To(Equal(expectedCoins), "expected different unvested coins")
				Expect(res.Vested).To(BeEmpty(), "expected different vested coins")

				// Commit new block so that the vesting period is at the half and the lockup period is over
				s.ctx, err = evmosutil.CommitAndCreateNewCtx(s.ctx, s.app, time.Second*5000, nil)
				Expect(err).To(BeNil(), "failed to commit block")

				// Recheck balances
				_, ethRes, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				err = s.precompile.UnpackIntoInterface(&res, vesting.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "error while unpacking response: %v", err)

				halfCoins := []cmn.Coin{{Denom: s.bondDenom, Amount: big.NewInt(500)}}
				Expect(res.Locked).To(BeEmpty(), "expected no coins to be locked anymore")
				Expect(res.Unvested).To(Equal(halfCoins), "expected different unvested coins")
				Expect(res.Vested).To(Equal(halfCoins), "expected different vested coins")

				// Commit new block so that the vesting period is over
				s.ctx, err = evmosutil.CommitAndCreateNewCtx(s.ctx, s.app, time.Second*5000, nil)
				Expect(err).To(BeNil(), "failed to commit block")

				// Recheck balances
				_, ethRes, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				err = s.precompile.UnpackIntoInterface(&res, vesting.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "error while unpacking response: %v", err)

				Expect(res.Locked).To(BeEmpty(), "expected no coins to be locked anymore")
				Expect(res.Unvested).To(BeEmpty(), "expected no coins to be unvested anymore")
				Expect(res.Vested).To(Equal(expectedCoins), "expected different vested coins")
			})

			It(fmt.Sprintf("should return an error when the account does not exist (%s)", callType.name), func() {
				balancesArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.BalancesMethod).
					WithArgs(
						toAddr,
					)

				balancesCheck := execRevertedCheck
				if callType.directCall {
					balancesCheck = failCheck.WithErrContains(fmt.Sprintf(
						"code = NotFound desc = account for address '%s'",
						sdk.AccAddress(toAddr.Bytes()).String(),
					))
				}

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, balancesArgs, balancesCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("account at address '%s' either does not exist or is not a vesting account", sdk.AccAddress(toAddr.Bytes())))
				}
			})

			It(fmt.Sprintf("should return an error when the account is not a vesting account (%s)", callType.name), func() {
				err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, toAddr.Bytes(), 1e18)
				Expect(err).ToNot(HaveOccurred(), "error while funding account: %v", err)

				balancesArgs := s.BuildCallArgs(callType, contractAddr).
					WithMethodName(vesting.BalancesMethod).
					WithArgs(
						toAddr,
					)

				balancesCheck := execRevertedCheck
				if callType.directCall {
					balancesCheck = failCheck.WithErrContains(fmt.Sprintf(
						"account at address '%s' is not a vesting account",
						sdk.AccAddress(toAddr.Bytes()).String(),
					))
				}

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, balancesArgs, balancesCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("InvalidArgument desc = account at address"))
				}
			})
		}
	})
})
