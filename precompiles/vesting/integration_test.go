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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/precompiles/authorization"
	cmn "github.com/evmos/evmos/v16/precompiles/common"
	"github.com/evmos/evmos/v16/precompiles/testutil"
	"github.com/evmos/evmos/v16/precompiles/vesting"
	"github.com/evmos/evmos/v16/precompiles/vesting/testdata"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	testutils "github.com/evmos/evmos/v16/testutil/integration/evmos/utils"
	testutiltx "github.com/evmos/evmos/v16/testutil/tx"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	vestingtypes "github.com/evmos/evmos/v16/x/vesting/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

const vestingAccIdx = 1

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
)

func TestPrecompileIntegrationTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Precompile Test Suite")
}

var _ = Describe("Interacting with the vesting extension", func() {
	var s *PrecompileTestSuite

	BeforeEach(func() {
		// Setup the test suite
		s = new(PrecompileTestSuite)
		s.SetupTest()

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
		contractAddr, err = s.factory.DeployContract(
			s.keyring.GetPrivKey(0),
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: testdata.VestingCallerContract,
			},
		)
		Expect(err).ToNot(HaveOccurred(), "error while deploying the smart contract: %v", err)
		Expect(s.network.NextBlock()).To(BeNil())

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
				err := testutils.FundAccountWithBaseDenom(
					s.factory, s.network,
					s.keyring.GetKey(0),
					toAddr.Bytes(), math.NewInt(10000),
				)
				Expect(err).ToNot(HaveOccurred(), "error while funding the account: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.CreateClawbackVestingAccountMethod
				callArgs.Args = []interface{}{
					funderAddr,
					s.keyring.GetAddr(0),
					false,
				}

				createClawbackCheck := passCheck.WithExpEvents(vesting.EventTypeCreateClawbackVestingAccount)

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, createClawbackCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				// Check the vesting account
				s.ExpectSimpleVestingAccount(s.keyring.GetAddr(0), funderAddr)
			})

			It(fmt.Sprintf("should not create a clawback vesting account if the account is not initialized (%s)", callType.name), func() {
				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.CreateClawbackVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					toAddr,
					false,
				}

				createClawbackCheck := failCheck.WithErrContains("account is not initialized")

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, createClawbackCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), toAddr.Bytes())
				Expect(acc).To(BeNil(), "account should not be created")
			})

			It(fmt.Sprintf("should not create a clawback vesting account if the vesting account is the zero address (%s)", callType.name), func() {
				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.CreateClawbackVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					common.Address{},
					false,
				}

				createClawbackCheck := failCheck.WithErrContains("invalid address")

				if !callType.directCall {
					createClawbackCheck = failCheck.WithErrContains("execution reverted")
				}

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, createClawbackCheck)
				Expect(err).To(BeNil(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should not create a clawback vesting account if the funder account is the zero address (%s)", callType.name), func() {
				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.CreateClawbackVestingAccountMethod
				callArgs.Args = []interface{}{
					common.Address{},
					s.keyring.GetAddr(0),
					false,
				}

				createClawbackCheck := failCheck.WithErrContains("invalid address")

				if !callType.directCall {
					createClawbackCheck = failCheck.WithErrContains("execution reverted")
				}

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, createClawbackCheck)
				Expect(err).To(BeNil(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should not create a clawback vesting account if the origin is different than the vesting address (%s)", callType.name), func() {
				err := testutils.FundAccountWithBaseDenom(
					s.factory, s.network,
					s.keyring.GetKey(0),
					differentAddr.Bytes(),
					math.NewInt(10000),
				)
				Expect(err).ToNot(HaveOccurred(), "error while funding the account: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.CreateClawbackVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					differentAddr,
					false,
				}

				createClawbackCheck := failCheck.WithErrContains("origin is different than the vesting address")

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, createClawbackCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("does not match the from address"))
				}
			})

			It(fmt.Sprintf("should not create a clawback vesting account for a smart contract (%s)", callType.name), func() {
				if callType.directCall {
					Skip("this should only be run for smart contract calls")
				}

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = "createClawbackVestingAccountForContract"

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, failCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(err.Error()).To(ContainSubstring("execution reverted"))
				Expect(s.network.NextBlock()).To(BeNil())

				// Check that the smart contract was not converted
				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), contractAddr.Bytes())
				Expect(acc).ToNot(BeNil(), "smart contract should be found")
				_, ok := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(ok).To(BeFalse(), "smart contract should not be converted to a vesting account")
			})

			It(fmt.Sprintf("should not create a clawback vesting account if the account already is subject to vesting (%s)", callType.name), func() {
				addr, priv := testutiltx.NewAddrKey()
				err := testutils.FundAccountWithBaseDenom(
					s.factory, s.network,
					s.keyring.GetKey(0),
					addr.Bytes(),
					math.NewInt(1e18),
				)
				Expect(err).ToNot(HaveOccurred(), "error while funding the account: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				err = s.factory.CreateClawbackVestingAccount(priv, s.keyring.GetAccAddr(0), false)
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.CreateClawbackVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					addr,
					false,
				}

				createClawbackCheck := failCheck.WithErrContains("account is already subject to vesting")

				_, _, err = s.factory.CallContractAndCheckLogs(priv, txArgs, callArgs, createClawbackCheck)
				Expect(s.network.NextBlock()).To(BeNil())
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
					approvalCallArgs := factory.CallArgs{
						ContractABI: s.precompile.ABI,
						MethodName:  "approve",
						Args: []interface{}{
							contractAddr,
							vesting.FundVestingAccountMsgURL,
						},
					}

					precompileAddr := s.precompile.Address()
					logCheck := passCheck.WithExpEvents(authorization.EventTypeApproval)

					_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), evmtypes.EvmTxArgs{To: &precompileAddr}, approvalCallArgs, logCheck)
					Expect(err).To(BeNil())
					Expect(s.network.NextBlock()).To(BeNil())
				}
			})

			It(fmt.Sprintf("should fund the vesting when defining only lockup (%s)", callType.name), func() {
				err = s.factory.CreateClawbackVestingAccount(s.keyring.GetPrivKey(vestingAccIdx), s.keyring.GetAccAddr(0), false)
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					s.keyring.GetAddr(vestingAccIdx),
					uint64(time.Now().Unix()),
					defaultPeriods,
					emptyPeriods,
				}

				fundClawbackVestingCheck := passCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				// Check the vesting account
				//
				// NOTE: The vesting account is created with the lockup periods only, since the vesting periods are empty.
				// The vesting periods are defaulted to instant vesting, i.e. period length = 0.
				s.ExpectVestingAccount(s.keyring.GetAddr(vestingAccIdx), defaultPeriods, instantPeriods)
			})

			It(fmt.Sprintf("should fund the vesting when defining only vesting (%s)", callType.name), func() {
				err = s.factory.CreateClawbackVestingAccount(s.keyring.GetPrivKey(vestingAccIdx), s.keyring.GetAccAddr(0), false)
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					s.keyring.GetAddr(vestingAccIdx),
					uint64(time.Now().Unix()),
					emptyPeriods,
					defaultPeriods,
				}

				fundClawbackVestingCheck := passCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				// Check the vesting account
				//
				// NOTE: The vesting account is created with the vesting periods only, since the lockup periods are empty.
				// The lockup periods are defaulted to instant unlocking, i.e. period length = 0.
				s.ExpectVestingAccount(s.keyring.GetAddr(vestingAccIdx), instantPeriods, defaultPeriods)
			})

			It(fmt.Sprintf("should fund the vesting when defining both lockup and vesting (%s)", callType.name), func() {
				err = s.factory.CreateClawbackVestingAccount(s.keyring.GetPrivKey(vestingAccIdx), s.keyring.GetAccAddr(0), false)
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					s.keyring.GetAddr(vestingAccIdx),
					uint64(time.Now().Unix()),
					defaultPeriods,
					defaultPeriods,
				}

				fundClawbackVestingCheck := passCheck.
					WithExpEvents(vesting.EventTypeFundVestingAccount)

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				// Check the vesting account
				s.ExpectVestingAccount(s.keyring.GetAddr(vestingAccIdx), defaultPeriods, defaultPeriods)
			})

			It(fmt.Sprintf("should not fund the vesting when defining different total coins for lockup and vesting (%s)", callType.name), func() {
				err = s.factory.CreateClawbackVestingAccount(s.keyring.GetPrivKey(vestingAccIdx), s.keyring.GetAccAddr(0), false)
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					s.keyring.GetAddr(vestingAccIdx),
					uint64(time.Now().Unix()),
					defaultPeriods,
					doublePeriods,
				}

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.WithErrContains("vesting and lockup schedules must have same total coins")
				}

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				// Check the vesting account
				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), s.keyring.GetAccAddr(vestingAccIdx))
				Expect(acc).ToNot(BeNil(), "account should exist")
				vestAcc, ok := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(ok).To(BeTrue())
				Expect(vestAcc.LockupPeriods).To(BeEmpty())
				Expect(vestAcc.VestingPeriods).To(BeEmpty())
			})

			It(fmt.Sprintf("should not fund the vesting when defining neither lockup nor vesting (%s)", callType.name), func() {
				err = s.factory.CreateClawbackVestingAccount(s.keyring.GetPrivKey(vestingAccIdx), s.keyring.GetAccAddr(0), false)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					s.keyring.GetAddr(vestingAccIdx),
					uint64(time.Now().Unix()),
					emptyPeriods,
					emptyPeriods,
				}

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.WithErrContains("vesting and/or lockup schedules must be present")
				}

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				// Check the vesting account
				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), s.keyring.GetAccAddr(vestingAccIdx))
				Expect(acc).ToNot(BeNil(), "account should exist")
				vestAcc, ok := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(ok).To(BeTrue())
				Expect(vestAcc.LockupPeriods).To(BeEmpty())
				Expect(vestAcc.VestingPeriods).To(BeEmpty())
			})

			It(fmt.Sprintf("should not fund the vesting when exceeding the funder balance (%s)", callType.name), func() {
				err = s.factory.CreateClawbackVestingAccount(s.keyring.GetPrivKey(vestingAccIdx), s.keyring.GetAccAddr(0), false)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				res, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
				Expect(err).To(BeNil())
				balance := res.Balance
				exceededBalance := new(big.Int).Add(big.NewInt(1), balance.Amount.BigInt())

				exceedingVesting := []vesting.Period{{
					Length: 10,
					Amount: []cmn.Coin{{Denom: s.bondDenom, Amount: exceededBalance}},
				}}

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					s.keyring.GetAddr(vestingAccIdx),
					uint64(time.Now().Unix()),
					exceedingVesting,
					emptyPeriods,
				}

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.WithErrContains("insufficient funds")
				}

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				// Check the vesting account
				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), s.keyring.GetAccAddr(vestingAccIdx))
				va, ok := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(ok).To(BeTrue())
				Expect(va.LockupPeriods).To(BeNil(), "vesting account should not be funded")
				Expect(va.VestingPeriods).To(BeNil(), "vesting account should not be funded")
			})

			It(fmt.Sprintf("should not fund the vesting when not sending as the funder (%s)", callType.name), func() {
				err = s.factory.CreateClawbackVestingAccount(s.keyring.GetPrivKey(vestingAccIdx), s.keyring.GetAccAddr(0), false)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				differentFunder := testutiltx.GenerateAddress()

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					differentFunder,
					s.keyring.GetAddr(vestingAccIdx),
					uint64(time.Now().Unix()),
					defaultPeriods,
					defaultPeriods,
				}

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.
						WithErrContains(
							fmt.Sprintf("tx origin address %s does not match the from address %s",
								s.keyring.GetAddr(0),
								differentFunder,
							),
						)
				}

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				// Check the vesting account
				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), s.keyring.GetAccAddr(vestingAccIdx))
				Expect(acc).ToNot(BeNil(), "account should exist")
				va, ok := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(ok).To(BeTrue())
				Expect(va.LockupPeriods).To(BeNil(), "vesting account should not be funded")
				Expect(va.VestingPeriods).To(BeNil(), "vesting account should not be funded")
			})

			It(fmt.Sprintf("should not fund the vesting when the address is blocked (%s)", callType.name), func() {
				moduleAddr := common.BytesToAddress(authtypes.NewModuleAddress("distribution").Bytes())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					moduleAddr,
					uint64(time.Now().Unix()),
					defaultPeriods,
					defaultPeriods,
				}

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.WithErrContains("is not allowed to receive funds")
				}

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).NotTo(HaveOccurred(), "error while creating a clawback vesting account for a module address", err)

				// check that the module address is not a vesting account
				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), moduleAddr.Bytes())
				Expect(acc).ToNot(BeNil(), "module account should be found")
				_, ok := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(ok).To(BeFalse(), "module account should not be a vesting account")
			})

			It(fmt.Sprintf("should not fund the vesting when the address is blocked - a precompile address (%s)", callType.name), func() {
				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					s.precompile.Address(),
					uint64(time.Now().Unix()),
					defaultPeriods,
					defaultPeriods,
				}

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.WithErrContains("is not allowed to receive funds")
				}

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).NotTo(HaveOccurred(), "error while creating a clawback vesting account for a module address", err)
			})

			It(fmt.Sprintf("should not fund the vesting when the address is uninitialized (%s)", callType.name), func() {
				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					differentAddr,
					uint64(time.Now().Unix()),
					defaultPeriods,
					defaultPeriods,
				}

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.WithErrContains("does not exist")
				}

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).NotTo(HaveOccurred(), "error while creating a clawback vesting account for a module address", err)
			})

			It(fmt.Sprintf("should not fund the vesting when the address is the zero address (%s)", callType.name), func() {
				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.FundVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					common.Address{},
					uint64(time.Now().Unix()),
					defaultPeriods,
					defaultPeriods,
				}

				fundClawbackVestingCheck := execRevertedCheck
				if callType.directCall {
					fundClawbackVestingCheck = failCheck.WithErrContains("invalid address")
				}

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, fundClawbackVestingCheck)
				Expect(err).NotTo(HaveOccurred(), "error while creating a clawback vesting account for the zero address", err)
			})
		}
	})

	Context("to claw back from a vesting account", func() {
		BeforeEach(func() {
			err = s.factory.CreateClawbackVestingAccount(s.keyring.GetPrivKey(vestingAccIdx), s.keyring.GetAccAddr(0), false)
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			err = s.factory.FundVestingAccount(s.keyring.GetPrivKey(0), s.keyring.GetAccAddr(vestingAccIdx), time.Now(), sdkLockupPeriods, sdkVestingPeriods)
			Expect(s.network.NextBlock()).To(BeNil())
		})

		for _, callType := range callTypes {
			callType := callType

			BeforeEach(func() {
				if callType.directCall == false {
					approvalCallArgs := factory.CallArgs{
						ContractABI: s.precompile.ABI,
						MethodName:  "approve",
						Args: []interface{}{
							contractAddr,
							vesting.ClawbackMsgURL,
						},
					}

					precompileAddr := s.precompile.Address()
					logCheck := passCheck.WithExpEvents(authorization.EventTypeApproval)

					_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), evmtypes.EvmTxArgs{To: &precompileAddr}, approvalCallArgs, logCheck)
					Expect(err).To(BeNil())
					Expect(s.network.NextBlock()).To(BeNil())
				}
			})

			It(fmt.Sprintf("should claw back from the vesting when sending as the funder (%s)", callType.name), func() {
				res, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(vestingAccIdx), s.bondDenom)
				Expect(err).To(BeNil())
				balancePre := res.Balance

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.ClawbackMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					s.keyring.GetAddr(vestingAccIdx),
					differentAddr,
				}

				clawbackCheck := passCheck.
					WithExpEvents(vesting.EventTypeClawback)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, clawbackCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				var (
					co             vesting.ClawbackOutput
					expClawbackAmt = math.NewInt(1000)
				)

				err = s.precompile.UnpackIntoInterface(&co, vesting.ClawbackMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "error while unpacking the clawback output: %v", err)
				Expect(co.Coins).To(Equal(balances), "expected different clawback amount")

				res, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(vestingAccIdx), s.bondDenom)
				Expect(err).To(BeNil())
				balancePost := res.Balance
				Expect(balancePost.Amount).To(Equal(balancePre.Amount.Sub(expClawbackAmt)), "expected only initial balance after clawback")
				res, err = s.grpcHandler.GetBalance(differentAddr.Bytes(), s.bondDenom)
				Expect(err).To(BeNil())
				balanceReceiver := res.Balance
				Expect(balanceReceiver.Amount).To(Equal(expClawbackAmt), "expected receiver to show different balance after clawback")
			})

			It(fmt.Sprintf("should return an error when not sending as the funder (%s)", callType.name), func() {
				// create and fund new account
				differentAddr, differentPriv := testutiltx.NewAddrKey()
				err := testutils.FundAccountWithBaseDenom(
					s.factory, s.network,
					s.keyring.GetKey(0),
					differentAddr.Bytes(),
					math.NewInt(1e18),
				)
				Expect(err).ToNot(HaveOccurred(), "error while funding the account: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				res, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(vestingAccIdx), s.bondDenom)
				Expect(err).To(BeNil())
				balancePre := res.Balance

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.ClawbackMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					s.keyring.GetAddr(vestingAccIdx),
					differentAddr,
				}

				clawbackCheck := execRevertedCheck
				if callType.directCall {
					clawbackCheck = failCheck.
						WithErrContains(fmt.Sprintf(
							"tx origin address %s does not match the funder address %s",
							differentAddr, s.keyring.GetAddr(0),
						))
				}

				_, _, err = s.factory.CallContractAndCheckLogs(differentPriv, txArgs, callArgs, clawbackCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)

				res, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(vestingAccIdx), s.bondDenom)
				Expect(err).To(BeNil())
				balancePost := res.Balance
				Expect(balancePost).To(Equal(balancePre), "expected balance not to have changed")
			})

			It(fmt.Sprintf("should return an error when the vesting does not exist (%s)", callType.name), func() {
				// fund the new account
				err := testutils.FundAccountWithBaseDenom(
					s.factory, s.network,
					s.keyring.GetKey(0),
					differentAddr.Bytes(),
					math.NewInt(1e18),
				)
				Expect(err).ToNot(HaveOccurred(), "error while funding the account: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.ClawbackMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					differentAddr,
					s.keyring.GetAddr(0),
				}

				clawbackCheck := execRevertedCheck
				// FIXME: error messages in fail check now work differently!
				if callType.directCall {
					clawbackCheck = failCheck.
						WithErrContains(vestingtypes.ErrNotSubjectToClawback.Error())
				}

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, clawbackCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should succeed and return empty Coins when all tokens are vested (%s)", callType.name), func() {
				// commit block with time so that vesting has ended
				err := s.network.NextBlockAfter(time.Hour * 24)
				Expect(err).ToNot(HaveOccurred(), "error while committing block: %v", err)

				res, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(vestingAccIdx), s.bondDenom)
				Expect(err).To(BeNil())
				balancePre := res.Balance

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.ClawbackMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					s.keyring.GetAddr(vestingAccIdx),
					s.keyring.GetAddr(0),
				}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, passCheck)
				Expect(err).To(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				var co vesting.ClawbackOutput
				err = s.precompile.UnpackIntoInterface(&co, vesting.ClawbackMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "error while unpacking the clawback output: %v", err)
				Expect(co.Coins).To(BeEmpty(), "expected empty clawback amount")

				res, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(vestingAccIdx), s.bondDenom)
				Expect(err).To(BeNil())
				balancePost := res.Balance
				Expect(balancePost).To(Equal(balancePre), "expected balance not to have changed")
			})
		}
	})

	Context("to update the vesting funder", func() {
		BeforeEach(func() {
			err = s.factory.CreateClawbackVestingAccount(s.keyring.GetPrivKey(vestingAccIdx), s.keyring.GetAccAddr(0), false)
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())
		})

		for _, callType := range callTypes {
			callType := callType

			BeforeEach(func() {
				if callType.directCall == false {
					approvalCallArgs := factory.CallArgs{
						ContractABI: s.precompile.ABI,
						MethodName:  "approve",
						Args: []interface{}{
							contractAddr,
							vesting.UpdateVestingFunderMsgURL,
						},
					}

					precompileAddr := s.precompile.Address()
					logCheck := passCheck.WithExpEvents(authorization.EventTypeApproval)

					_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), evmtypes.EvmTxArgs{To: &precompileAddr}, approvalCallArgs, logCheck)
					Expect(err).To(BeNil())
					Expect(s.network.NextBlock()).To(BeNil())
				}
			})

			It(fmt.Sprintf("should update the vesting funder when sending as the funder (%s)", callType.name), func() {
				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.UpdateVestingFunderMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					differentAddr,
					s.keyring.GetAddr(vestingAccIdx),
				}

				updateFunderCheck := passCheck.
					WithExpEvents(vesting.EventTypeUpdateVestingFunder)

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, updateFunderCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				// Check that the vesting account has the new funder
				s.ExpectVestingFunder(s.keyring.GetAddr(vestingAccIdx), differentAddr)
			})

			It(fmt.Sprintf("should return an error when not sending as the funder (%s)", callType.name), func() {
				differentAddr, differentPriv := testutiltx.NewAddrKey()
				err := testutils.FundAccountWithBaseDenom(
					s.factory, s.network,
					s.keyring.GetKey(0),
					differentAddr.Bytes(),
					math.NewInt(1e18),
				)
				Expect(err).ToNot(HaveOccurred(), "error while funding the account: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.UpdateVestingFunderMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					differentAddr,
					s.keyring.GetAddr(vestingAccIdx),
				}

				updateFunderCheck := execRevertedCheck
				if callType.directCall {
					updateFunderCheck = failCheck.
						WithErrContains(fmt.Sprintf(
							"tx origin address %s does not match the funder address %s",
							differentAddr.String(), s.keyring.GetAddr(0).String(),
						))
				}

				_, _, err = s.factory.CallContractAndCheckLogs(differentPriv, txArgs, callArgs, updateFunderCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				// Check that the vesting account still has the same funder
				s.ExpectVestingFunder(s.keyring.GetAddr(vestingAccIdx), s.keyring.GetAddr(0))
			})

			It(fmt.Sprintf("should return an error when the account does not exist (%s)", callType.name), func() {
				// Check that there's no account
				nonExistentAddr := testutiltx.GenerateAddress()
				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), nonExistentAddr.Bytes())
				Expect(acc).To(BeNil(), "expected no account to be found")

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.UpdateVestingFunderMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					differentAddr,
					nonExistentAddr, // the address of the vesting account
				}

				updateFunderCheck := execRevertedCheck
				if callType.directCall {
					updateFunderCheck = failCheck.
						WithErrContains(fmt.Sprintf(
							"account at address %s does not exist",
							sdk.AccAddress(nonExistentAddr.Bytes()).String(),
						))
				}

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, updateFunderCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
				if callType.directCall {
					Expect(err.Error()).To(ContainSubstring("does not exist"))
				}
			})

			It(fmt.Sprintf("should return an error when the account is no vesting account (%s)", callType.name), func() {
				// Check that there's no vesting account
				nonVestingAddr := testutiltx.GenerateAddress()
				err := testutils.FundAccountWithBaseDenom(
					s.factory, s.network,
					s.keyring.GetKey(0),
					nonVestingAddr.Bytes(),
					math.NewInt(1e18),
				)
				Expect(err).ToNot(HaveOccurred(), "error while funding the account: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), nonVestingAddr.Bytes())
				Expect(acc).ToNot(BeNil(), "expected account to be found")
				_, ok := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(ok).To(BeFalse(), "expected account not to be a vesting account")

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.UpdateVestingFunderMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					differentAddr,
					nonVestingAddr, // the address of the vesting account
				}

				updateFunderCheck := execRevertedCheck
				if callType.directCall {
					updateFunderCheck = failCheck.
						WithErrContains(fmt.Sprintf(
							"%s: %s",
							sdk.AccAddress(nonVestingAddr.Bytes()),
							vestingtypes.ErrNotSubjectToClawback.Error(),
						))
				}

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, updateFunderCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should return an error when the new funder is the zero address (%s)", callType.name), func() {
				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.UpdateVestingFunderMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					common.Address{},
					s.keyring.GetAddr(vestingAccIdx),
				}

				updateFunderCheck := execRevertedCheck
				if callType.directCall {
					updateFunderCheck = failCheck.
						WithErrContains("new funder address cannot be the zero address")
				}

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, updateFunderCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should return an error when the new funder is the same as the current funder (%s)", callType.name), func() {
				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.UpdateVestingFunderMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					s.keyring.GetAddr(0),
					s.keyring.GetAddr(vestingAccIdx),
				}

				updateFunderCheck := execRevertedCheck
				if callType.directCall {
					updateFunderCheck = failCheck.
						WithErrContains("new funder address is equal to current funder address")
				}

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, updateFunderCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)

				// Check that the vesting account still has the same funder
				s.ExpectVestingFunder(s.keyring.GetAddr(vestingAccIdx), s.keyring.GetAddr(0))
			})

			It(fmt.Sprintf("should return an error when the new funder is a blocked address (%s)", callType.name), func() {
				moduleAddr := common.BytesToAddress(authtypes.NewModuleAddress("distribution").Bytes())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.UpdateVestingFunderMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0),
					moduleAddr,
					s.keyring.GetAddr(vestingAccIdx),
				}

				updateFunderCheck := execRevertedCheck
				if callType.directCall {
					updateFunderCheck = failCheck.
						WithErrContains("not allowed to fund vesting accounts")
				}

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, updateFunderCheck)
				Expect(err).NotTo(HaveOccurred(), "error while updating the funder to a module address: %v", err)
			})
		}
	})

	Context("to convert a vesting account", func() {
		BeforeEach(func() {
			// Create a vesting account
			err = s.factory.CreateClawbackVestingAccount(s.keyring.GetPrivKey(vestingAccIdx), s.keyring.GetAccAddr(0), false)
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())
		})

		for _, callType := range callTypes {
			callType := callType

			It(fmt.Sprintf("should convert the vesting account into a normal one after vesting has ended (%s)", callType.name), func() {
				// commit block with new time so that the vesting period has ended
				err = s.network.NextBlockAfter(time.Duration(time.Now().Add(time.Hour * 24).Hour()))
				Expect(err).To(BeNil(), "failed to commit block")

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.ConvertVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(vestingAccIdx),
				}

				convertClawbackCheck := passCheck.
					WithExpEvents(vesting.EventTypeConvertVestingAccount)

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, convertClawbackCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				// Check that the vesting account has been converted
				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), s.keyring.GetAccAddr(vestingAccIdx))
				Expect(acc).ToNot(BeNil(), "expected account to be found")
				_, ok := acc.(*vestingtypes.ClawbackVestingAccount)
				Expect(ok).To(BeFalse(), "expected account not to be a vesting account")
			})

			It(fmt.Sprintf("should return an error when the vesting has not ended yet (%s)", callType.name), func() {
				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.ConvertVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(vestingAccIdx),
				}

				convertClawbackCheck := execRevertedCheck
				if callType.directCall {
					convertClawbackCheck = failCheck.WithErrContains("vesting coins still left in account")
				}

				txArgs.GasLimit = 300_000
				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, convertClawbackCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should return an error when not sending as the funder (%s)", callType.name), func() {
				// commit block with new time so that the vesting period has ended
				err := s.network.NextBlockAfter(time.Hour * 24)
				Expect(err).To(BeNil(), "failed to commit block")

				differentAddr, differentPriv := testutiltx.NewAddrKey()
				err = testutils.FundAccountWithBaseDenom(
					s.factory, s.network,
					s.keyring.GetKey(0),
					differentAddr.Bytes(),
					math.NewInt(1e18),
				)
				Expect(err).ToNot(HaveOccurred(), "error while funding account: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.ConvertVestingAccountMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(vestingAccIdx),
				}

				convertClawbackCheck := execRevertedCheck
				if callType.directCall {
					convertClawbackCheck = failCheck.WithErrContains("sender is not the funder")
				}

				txArgs.GasLimit = 300_000
				_, _, err = s.factory.CallContractAndCheckLogs(differentPriv, txArgs, callArgs, convertClawbackCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should return an error when the vesting does not exist (%s)", callType.name), func() {
				err = testutils.FundAccountWithBaseDenom(
					s.factory, s.network,
					s.keyring.GetKey(0),
					differentAddr.Bytes(),
					math.NewInt(1e18),
				)
				Expect(err).ToNot(HaveOccurred(), "error while funding account: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.ConvertVestingAccountMethod
				callArgs.Args = []interface{}{
					differentAddr, // this currently has no vesting
				}

				convertClawbackCheck := execRevertedCheck
				if callType.directCall {
					convertClawbackCheck = failCheck.WithErrContains("account is not subject to clawback vesting")
				}

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, convertClawbackCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)

				// Check that the account is no vesting account
				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), differentAddr.Bytes())
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
				err = s.factory.CreateClawbackVestingAccount(s.keyring.GetPrivKey(vestingAccIdx), s.keyring.GetAccAddr(0), false)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				err = s.factory.FundVestingAccount(s.keyring.GetPrivKey(0), s.keyring.GetAccAddr(vestingAccIdx), time.Now(), sdkLockupPeriods, sdkVestingPeriods)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.BalancesMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(vestingAccIdx),
				}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				var res vesting.BalancesOutput
				err = s.precompile.UnpackIntoInterface(&res, vesting.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "error while unpacking response: %v", err)

				expectedCoins := []cmn.Coin{{Denom: s.bondDenom, Amount: big.NewInt(1000)}}
				Expect(res.Locked).To(Equal(expectedCoins), "expected different locked coins")
				Expect(res.Unvested).To(Equal(expectedCoins), "expected different unvested coins")
				Expect(res.Vested).To(BeEmpty(), "expected different vested coins")

				// Commit new block so that the vesting period is at the half and the lockup period is over
				err = s.network.NextBlockAfter(time.Duration(time.Now().Add(time.Second * 5000).Second()))
				Expect(err).To(BeNil(), "failed to commit block")

				// Recheck balances
				_, ethRes, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, passCheck)
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
				_, ethRes, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "error while calling the contract: %v", err)

				err = s.precompile.UnpackIntoInterface(&res, vesting.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "error while unpacking response: %v", err)

				Expect(res.Locked).To(BeEmpty(), "expected no coins to be locked anymore")
				Expect(res.Unvested).To(BeEmpty(), "expected no coins to be unvested anymore")
				Expect(res.Vested).To(Equal(expectedCoins), "expected different vested coins")
			})

			It(fmt.Sprintf("should return an error when the account does not exist (%s)", callType.name), func() {
				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.BalancesMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(vestingAccIdx),
				}

				balancesCheck := execRevertedCheck
				if callType.directCall {
					balancesCheck = failCheck.WithErrContains(fmt.Sprintf(
						"account at address '%s' either does not exist or is not a vesting account", s.keyring.GetAccAddr(vestingAccIdx)))
				}

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, balancesCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
			})

			It(fmt.Sprintf("should return an error when the account is not a vesting account (%s)", callType.name), func() {
				err = testutils.FundAccountWithBaseDenom(
					s.factory, s.network,
					s.keyring.GetKey(0),
					s.keyring.GetAccAddr(vestingAccIdx),
					math.NewInt(1e18),
				)
				Expect(err).ToNot(HaveOccurred(), "error while funding account: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs, txArgs := s.BuildCallArgs(callType, contractAddr)
				callArgs.MethodName = vesting.BalancesMethod
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(vestingAccIdx),
				}

				balancesCheck := execRevertedCheck
				if callType.directCall {
					balancesCheck = failCheck.WithErrContains(fmt.Sprintf(
						"account at address '%s' is not a vesting account",
						s.keyring.GetAccAddr(vestingAccIdx),
					))
				}

				_, _, err = s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, balancesCheck)
				Expect(err).NotTo(HaveOccurred(), "error while calling the contract: %v", err)
			})
		}
	})
})
