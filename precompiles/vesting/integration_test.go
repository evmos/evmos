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

			BeforeEach(func() {
				if callType.directCall == false {
					err = vesting.CreateGenericAuthz(s.ctx, s.app.AuthzKeeper, contractAddr, s.address, vesting.FundVestingAccountMsgURL)
					Expect(err).ToNot(HaveOccurred(), "error while creating the generic authorization: %v", err)
				}
			})

			It(fmt.Sprintf("should fund the vesting when defining only lockup (%s)", callType.name), func() {
				s.CreateTestClawbackVestingAccount(s.address, toAddr)
				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
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

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				// Check the vesting account
				//
				// NOTE: The vesting account is created with the lockup periods only, since the vesting periods are empty.
				// The vesting periods are defaulted to instant vesting, i.e. period length = 0.
				s.ExpectVestingAccount(toAddr, defaultPeriods, instantPeriods)
			})

			It(fmt.Sprintf("should fund the vesting when defining only vesting (%s)", callType.name), func() {
				s.CreateTestClawbackVestingAccount(s.address, toAddr)
				createClawbackArgs := s.BuildCallArgs(callType, contractAddr).
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

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createClawbackArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				// Check the vesting account
				//
				// NOTE: The vesting account is created with the vesting periods only, since the lockup periods are empty.
				// The lockup periods are defaulted to instant unlocking, i.e. period length = 0.
				s.ExpectVestingAccount(toAddr, instantPeriods, defaultPeriods)
			})

			It(fmt.Sprintf("should fund the vesting when defining both lockup and vesting (%s)", callType.name), func() {
				s.CreateTestClawbackVestingAccount(s.address, toAddr)
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
		BeforeEach(func() {
			s.CreateTestClawbackVestingAccount(s.address, toAddr)
			s.FundTestClawbackVestingAccount()
		})

		for _, callType := range callTypes {
			callType := callType

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
				// FIXME: error messages in fail check now work differently!
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
