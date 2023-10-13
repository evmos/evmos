// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package p256_test

import (
	// . "github.com/onsi/ginkgo/v2"
	// . "github.com/onsi/gomega"

	"github.com/evmos/evmos/v14/precompiles/testutil"
	"github.com/evmos/evmos/v14/precompiles/testutil/contracts"
)

var (
	// defaultCallArgs  are the default arguments for calling the smart contract.
	//
	// NOTE: this has to be populated in a BeforeEach block because the contractAddr would otherwise be a nil address.
	defaultCallArgs contracts.CallArgs

	// defaultLogCheck instantiates a log check arguments struct with the precompile ABI events populated.
	defaultLogCheck testutil.LogCheckArgs
	// passCheck defines the arguments to check if the precompile returns no error
	passCheck testutil.LogCheckArgs
	// outOfGasCheck defines the arguments to check if the precompile returns out of gas error
	outOfGasCheck testutil.LogCheckArgs
)

// var _ = Describe("Calling staking precompile directly", func() {
// 	var (
// 		// oneE18Coin is a sdk.Coin with an amount of 1e18 in the test suite's bonding denomination
// 		oneE18Coin = sdk.NewCoin(s.bondDenom, sdk.NewInt(1e18))
// 		// twoE18Coin is a sdk.Coin with an amount of 2e18 in the test suite's bonding denomination
// 		twoE18Coin = sdk.NewCoin(s.bondDenom, sdk.NewInt(2e18))
// 	)

// 	BeforeEach(func() {
// 		s.SetupTest()
// 		s.NextBlock()

// 		defaultCallArgs = contracts.CallArgs{
// 			ContractAddr: s.precompile.Address(),
// 			PrivKey:      s.privKey,
// 		}
// 		defaultApproveArgs = defaultCallArgs.WithMethodName(authorization.ApproveMethod)

// 		defaultLogCheck = testutil.LogCheckArgs{ExpPass: true}
// 		outOfGasCheck = defaultLogCheck.WithErrContains(vm.ErrOutOfGas.Error())
// 	})

// 	Describe("when the precompile is not enabled in the EVM params", func() {
// 		It("should return an error", func() {
// 			// disable the precompile
// 			params := s.app.EvmKeeper.GetParams(s.ctx)
// 			var activePrecompiles []string
// 			for _, precompile := range params.ActivePrecompiles {
// 				if precompile != s.precompile.Address().String() {
// 					activePrecompiles = append(activePrecompiles, precompile)
// 				}
// 			}

// 			params.ActivePrecompiles = activePrecompiles
// 			err := s.app.EvmKeeper.SetParams(s.ctx, params)
// 			Expect(err).To(BeNil(), "error while setting params")

// 			input := s.signMsg([]byte("hello world"), s.p256Priv)

// 			// try to call the precompile
// 			verifyArg := defaultCallArgs.
// 				WithArgs(input)

// 			_, _, err = contracts.Call(s.ctx, s.app, verifyArg)
// 			Expect(err).To(HaveOccurred(), "expected error while calling the precompile")
// 			Expect(err.Error()).To(ContainSubstring("precompile not enabled"))
// 		})
// 	})

// 	Describe("Revert transaction", func() {
// 		It("should run out of gas if the gas limit is too low", func() {
// 			input := s.signMsg([]byte("hello world"), s.p256Priv)

// 			outOfGasArgs := defaultCallArgs.
// 				WithGasLimit(p256.VerifyGas - 1).
// 				WithArgs(input)

// 			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, outOfGasArgs, outOfGasCheck)
// 			Expect(err).To(HaveOccurred(), "error while calling precompile")
// 		})
// 	})

// 	It("Should refund leftover gas", func() {
// 		balancePre := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
// 		gasPrice := big.NewInt(1e9)

// 		// Call the precompile with a lot of gas
// 		approveArgs := defaultApproveArgs.
// 			WithGasPrice(gasPrice).
// 			WithArgs(s.precompile.Address(), big.NewInt(1e18), []string{staking.DelegateMsg})

// 		approvalCheck := passCheck.WithExpEvents(authorization.EventTypeApproval)

// 		res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, approveArgs, approvalCheck)
// 		Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

// 		s.NextBlock()

// 		balancePost := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
// 		difference := balancePre.Sub(balancePost)

// 		// NOTE: the expected difference is the gas price multiplied by the gas used, because the rest should be refunded
// 		expDifference := gasPrice.Int64() * res.GasUsed
// 		Expect(difference.Amount.Int64()).To(Equal(expDifference), "expected different total transaction cost")
// 	})
// })

// var _ = Describe("Calling staking precompile via Solidity", func() {
// 	var (
// 		// contractAddr is the address of the smart contract that will be deployed
// 		contractAddr common.Address

// 		// approvalCheck is a configuration for the log checker to see if an approval event was emitted.
// 		approvalCheck testutil.LogCheckArgs
// 		// execRevertedCheck defines the default log checking arguments which include the
// 		// standard revert message
// 		execRevertedCheck testutil.LogCheckArgs
// 		// err is a basic error type
// 		err error
// 	)

// 	BeforeEach(func() {
// 		s.SetupTest()
// 		contractAddr, err = s.DeployContract(testdata.StakingCallerContract)
// 		Expect(err).To(BeNil(), "error while deploying the smart contract: %v", err)
// 		valAddr = s.validators[0].GetOperator()
// 		valAddr2 = s.validators[1].GetOperator()

// 		s.NextBlock()

// 		// check contract was correctly deployed
// 		cAcc := s.app.EvmKeeper.GetAccount(s.ctx, contractAddr)
// 		Expect(cAcc).ToNot(BeNil(), "contract account should exist")
// 		Expect(cAcc.IsContract()).To(BeTrue(), "account should be a contract")

// 		// populate default call args
// 		defaultCallArgs = contracts.CallArgs{
// 			ContractAddr: contractAddr,
// 			ContractABI:  testdata.StakingCallerContract.ABI,
// 			PrivKey:      s.privKey,
// 		}
// 		// populate default approval args
// 		defaultApproveArgs = defaultCallArgs.WithMethodName("testApprove")

// 		// populate default log check args
// 		defaultLogCheck = testutil.LogCheckArgs{
// 			ABIEvents: s.precompile.Events,
// 		}
// 		execRevertedCheck = defaultLogCheck.WithErrContains(vm.ErrExecutionReverted.Error())
// 		passCheck = defaultLogCheck.WithExpPass(true)
// 		approvalCheck = passCheck.WithExpEvents(authorization.EventTypeApproval)
// 	})

// 	Describe("when the precompile is not enabled in the EVM params", func() {
// 		It("should return an error", func() {
// 			// disable the precompile
// 			params := s.app.EvmKeeper.GetParams(s.ctx)
// 			var activePrecompiles []string
// 			for _, precompile := range params.ActivePrecompiles {
// 				if precompile != s.precompile.Address().String() {
// 					activePrecompiles = append(activePrecompiles, precompile)
// 				}
// 			}
// 			params.ActivePrecompiles = activePrecompiles
// 			err := s.app.EvmKeeper.SetParams(s.ctx, params)
// 			Expect(err).To(BeNil(), "error while setting params")

// 			// try to call the precompile
// 			delegateArgs := defaultCallArgs.
// 				WithMethodName("testDelegate").
// 				WithArgs(
// 					s.address, valAddr.String(), big.NewInt(2e18),
// 				)

// 			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, execRevertedCheck)
// 			Expect(err).To(HaveOccurred(), "expected error while calling the precompile")
// 			Expect(err.Error()).To(ContainSubstring(vm.ErrExecutionReverted.Error()))
// 		})
// 	})

// 	Context("testing sequential function calls to the precompile", func() {
// 		// NOTE: there's no additional setup necessary because the test suite is already set up with
// 		// delegations to the validator
// 		It("should revert everything if any operation fails", func() {
// 			cArgs := defaultCallArgs.
// 				WithMethodName("testApproveAndThenUndelegate").
// 				WithGasLimit(1e8).
// 				WithArgs(contractAddr, big.NewInt(250), big.NewInt(500), valAddr.String())

// 			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, execRevertedCheck)
// 			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

// 			// There should be no authorizations because everything should have been reverted
// 			authz, _ := s.CheckAuthorization(staking.UndelegateAuthz, contractAddr, s.address)
// 			Expect(authz).To(BeNil(), "expected authorization to be nil")

// 			undelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
// 			Expect(undelegations).To(HaveLen(0), "expected no unbonding delegations")
// 		})

// 		It("should write to state if all operations succeed", func() {
// 			cArgs := defaultCallArgs.
// 				WithMethodName("testApproveAndThenUndelegate").
// 				WithGasLimit(1e8).
// 				WithArgs(contractAddr, big.NewInt(1000), big.NewInt(500), valAddr.String())

// 			logCheckArgs := passCheck.
// 				WithExpEvents(authorization.EventTypeApproval, staking.EventTypeUnbond)

// 			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, logCheckArgs)
// 			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

// 			authz, _ := s.CheckAuthorization(staking.UndelegateAuthz, contractAddr, s.address)
// 			Expect(authz).ToNot(BeNil(), "expected authorization not to be nil")

// 			undelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
// 			Expect(undelegations).To(HaveLen(1), "expected one unbonding delegation")
// 			Expect(undelegations[0].ValidatorAddress).To(Equal(valAddr.String()), "expected different validator address")
// 		})
// 	})

// 	Context("when using special call opcodes", func() {
// 		testcases := []struct {
// 			// calltype is the opcode to use
// 			calltype string
// 			// expTxPass defines if executing transactions should be possible with the given opcode.
// 			// Queries should work for all options.
// 			expTxPass bool
// 		}{
// 			{"call", true},
// 			{"callcode", false},
// 			{"staticcall", false},
// 			{"delegatecall", false},
// 		}

// 		BeforeEach(func() {
// 			// approve undelegate message
// 			approveArgs := defaultApproveArgs.WithArgs(
// 				contractAddr, []string{staking.UndelegateMsg}, big.NewInt(1e18),
// 			)
// 			s.SetupApprovalWithContractCalls(approveArgs)

// 			s.NextBlock()
// 		})

// 		for _, tc := range testcases {
// 			// NOTE: this is necessary because of Ginkgo behavior -- if not done, the value of tc
// 			// inside the It block will always be the last entry in the testcases slice
// 			testcase := tc

// 			It(fmt.Sprintf("should not execute transactions for calltype %q", testcase.calltype), func() {
// 				args := defaultCallArgs.
// 					WithMethodName("testCallUndelegate").
// 					WithArgs(s.address, valAddr.String(), big.NewInt(1e18), testcase.calltype)

// 				checkArgs := execRevertedCheck
// 				if testcase.expTxPass {
// 					checkArgs = passCheck.WithExpEvents(staking.EventTypeUnbond)
// 				}

// 				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, args, checkArgs)
// 				if testcase.expTxPass {
// 					Expect(err).To(BeNil(), "error while calling the smart contract for calltype %s: %v", testcase.calltype, err)
// 				} else {
// 					Expect(err).To(HaveOccurred(), "error while calling the smart contract for calltype %s: %v", testcase.calltype, err)
// 				}
// 				// check no delegations are unbonding
// 				undelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())

// 				if testcase.expTxPass {
// 					Expect(undelegations).To(HaveLen(1), "expected an unbonding delegation")
// 					Expect(undelegations[0].ValidatorAddress).To(Equal(valAddr.String()), "expected different validator address")
// 					Expect(undelegations[0].DelegatorAddress).To(Equal(sdk.AccAddress(s.address.Bytes()).String()), "expected different delegator address")
// 				} else {
// 					Expect(undelegations).To(HaveLen(0), "expected no unbonding delegations for calltype %s", testcase.calltype)
// 				}
// 			})

// 			It(fmt.Sprintf("should execute queries for calltype %q", testcase.calltype), func() {
// 				args := defaultCallArgs.
// 					WithMethodName("testCallDelegation").
// 					WithArgs(s.address, valAddr.String(), testcase.calltype)

// 				_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, args, passCheck)
// 				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

// 				var delOut staking.DelegationOutput
// 				err = s.precompile.UnpackIntoInterface(&delOut, staking.DelegationMethod, ethRes.Ret)
// 				Expect(err).To(BeNil(), "error while unpacking the delegation output: %v", err)
// 				Expect(delOut.Shares).To(Equal(sdk.NewDec(1).BigInt()), "expected different delegation shares")
// 				Expect(delOut.Balance.Amount).To(Equal(big.NewInt(1e18)), "expected different delegation balance")
// 				if testcase.calltype != "callcode" { // having some trouble with returning the denom from inline assembly but that's a very special edge case which might never be used
// 					Expect(delOut.Balance.Denom).To(Equal(s.bondDenom), "expected different denomination")
// 				}
// 			})
// 		}
// 	})

// 	// NOTE: These tests were added to replicate a problematic behavior, that occurred when a contract
// 	// adjusted the state in multiple subsequent function calls, which adjusted the EVM state as well as
// 	// things from the Cosmos SDK state (e.g. a bank balance).
// 	// The result was, that changes made to the Cosmos SDK state have been overwritten during the next function
// 	// call, because the EVM state was not updated in between.
// 	//
// 	// This behavior was fixed by updating the EVM state after each function call.
// 	Context("when triggering multiple state changes in one function", func() {
// 		// delegationAmount is the amount to be delegated
// 		delegationAmount := big.NewInt(1e18)

// 		BeforeEach(func() {
// 			// Set up funding for the contract address.
// 			// NOTE: we are first asserting that no balance exists and then check successful
// 			// funding afterwards.
// 			balanceBefore := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
// 			Expect(balanceBefore.Amount.Int64()).To(BeZero(), "expected contract balance to be 0 before funding")

// 			err = s.app.BankKeeper.SendCoins(
// 				s.ctx, s.address.Bytes(), contractAddr.Bytes(),
// 				sdk.Coins{sdk.Coin{Denom: s.bondDenom, Amount: math.NewIntFromBigInt(delegationAmount)}},
// 			)
// 			Expect(err).To(BeNil(), "error while sending coins: %v", err)

// 			s.NextBlock()

// 			balanceAfterFunding := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
// 			Expect(balanceAfterFunding.Amount.BigInt()).To(Equal(delegationAmount), "expected different contract balance after funding")

// 			// Check no delegation exists from the contract to the validator
// 			_, found := s.app.StakingKeeper.GetDelegation(s.ctx, contractAddr.Bytes(), valAddr)
// 			Expect(found).To(BeFalse(), "expected delegation not to be found before testing")
// 		})

// 		It("delegating and increasing counter should change the bank balance accordingly", func() {
// 			delegationArgs := defaultCallArgs.
// 				WithGasLimit(1e9).
// 				WithMethodName("testDelegateIncrementCounter").
// 				WithArgs(valAddr.String(), delegationAmount)

// 			approvalAndDelegationCheck := passCheck.WithExpEvents(
// 				authorization.EventTypeApproval, staking.EventTypeDelegate,
// 			)

// 			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegationArgs, approvalAndDelegationCheck)
// 			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

// 			del, found := s.app.StakingKeeper.GetDelegation(s.ctx, contractAddr.Bytes(), valAddr)

// 			Expect(found).To(BeTrue(), "expected delegation to be found")
// 			Expect(del.GetShares().BigInt()).To(Equal(delegationAmount), "expected different delegation shares")

// 			postBalance := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
// 			Expect(postBalance.Amount.Int64()).To(BeZero(), "expected balance to be 0 after contract call")
// 		})
// 	})

// 	Context("when updating the stateDB prior to calling the precompile", func() {
// 		It("should utilize the same contract balance to delegate", func() {
// 			delegationArgs := defaultCallArgs.
// 				WithGasLimit(1e9).
// 				WithMethodName("approveDepositAndDelegate").
// 				WithArgs(valAddr.String()).
// 				WithAmount(big.NewInt(2e18))

// 			approvalAndDelegationCheck := passCheck.WithExpEvents(
// 				authorization.EventTypeApproval, staking.EventTypeDelegate,
// 			)
// 			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegationArgs, approvalAndDelegationCheck)
// 			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
// 			balance := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
// 			Expect(balance.Amount.Int64()).To(BeZero(), "expected different contract balance after funding")
// 			delegation := s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, contractAddr.Bytes())
// 			Expect(delegation).To(HaveLen(1), "expected one delegation")
// 			Expect(delegation[0].GetShares().BigInt()).To(Equal(big.NewInt(2e18)), "expected different delegation shares")
// 		})
// 		//nolint:dupl
// 		It("should revert the contract balance to the original value when the precompile fails", func() {
// 			delegationArgs := defaultCallArgs.
// 				WithGasLimit(1e9).
// 				WithMethodName("approveDepositAndDelegateExceedingAllowance").
// 				WithArgs(valAddr.String()).
// 				WithAmount(big.NewInt(2e18))

// 			approvalAndDelegationCheck := defaultLogCheck.WithErrContains(vm.ErrExecutionReverted.Error())
// 			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegationArgs, approvalAndDelegationCheck)
// 			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

// 			balance := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
// 			Expect(balance.Amount.Int64()).To(BeZero(), "expected different contract balance after funding")
// 			auth, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, contractAddr.Bytes(), s.address.Bytes(), staking.DelegateMsg)
// 			Expect(auth).To(BeNil(), "expected no authorization")
// 			delegation := s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, contractAddr.Bytes())
// 			Expect(delegation).To(HaveLen(0), "expected no delegations")
// 		})

// 		//nolint:dupl
// 		It("should revert the contract balance to the original value when the custom logic after the precompile fails ", func() {
// 			delegationArgs := defaultCallArgs.
// 				WithGasLimit(1e9).
// 				WithMethodName("approveDepositDelegateAndFailCustomLogic").
// 				WithArgs(valAddr.String()).
// 				WithAmount(big.NewInt(2e18))

// 			approvalAndDelegationCheck := defaultLogCheck.WithErrContains(vm.ErrExecutionReverted.Error())
// 			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegationArgs, approvalAndDelegationCheck)
// 			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

// 			balance := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
// 			Expect(balance.Amount.Int64()).To(BeZero(), "expected different contract balance after funding")
// 			auth, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, contractAddr.Bytes(), s.address.Bytes(), staking.DelegateMsg)
// 			Expect(auth).To(BeNil(), "expected no authorization")
// 			delegation := s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, contractAddr.Bytes())
// 			Expect(delegation).To(HaveLen(0), "expected no delegations")
// 		})
// 	})
// })
