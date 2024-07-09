// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package staking_test

import (
	"fmt"
	"math/big"
	"time"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	compiledcontracts "github.com/evmos/evmos/v18/contracts"
	"github.com/evmos/evmos/v18/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v18/precompiles/authorization"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/distribution"
	"github.com/evmos/evmos/v18/precompiles/staking"
	"github.com/evmos/evmos/v18/precompiles/staking/testdata"
	"github.com/evmos/evmos/v18/precompiles/testutil"
	"github.com/evmos/evmos/v18/precompiles/testutil/contracts"
	evmosutil "github.com/evmos/evmos/v18/testutil"
	testutiltx "github.com/evmos/evmos/v18/testutil/tx"
	"github.com/evmos/evmos/v18/utils"
	"github.com/evmos/evmos/v18/x/evm/core/vm"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
	vestingtypes "github.com/evmos/evmos/v18/x/vesting/types"
)

// General variables used for integration tests
var (
	// valAddr and valAddr2 are the two validator addresses used for testing
	valAddr, valAddr2 sdk.ValAddress

	// defaultCallArgs and defaultApproveArgs are the default arguments for calling the smart contract and to
	// call the approve method specifically.
	//
	// NOTE: this has to be populated in a BeforeEach block because the contractAddr would otherwise be a nil address.
	defaultCallArgs, defaultApproveArgs contracts.CallArgs

	// defaultLogCheck instantiates a log check arguments struct with the precompile ABI events populated.
	defaultLogCheck testutil.LogCheckArgs
	// passCheck defines the arguments to check if the precompile returns no error
	passCheck testutil.LogCheckArgs
	// outOfGasCheck defines the arguments to check if the precompile returns out of gas error
	outOfGasCheck testutil.LogCheckArgs
)

var _ = Describe("Calling staking precompile directly", func() {
	var (
		// oneE18Coin is a sdk.Coin with an amount of 1e18 in the test suite's bonding denomination
		oneE18Coin = sdk.NewCoin(s.bondDenom, math.NewInt(1e18))
		// twoE18Coin is a sdk.Coin with an amount of 2e18 in the test suite's bonding denomination
		twoE18Coin = sdk.NewCoin(s.bondDenom, math.NewInt(2e18))
	)

	BeforeEach(func() {
		s.SetupTest()
		s.NextBlock()

		valAddr = s.validators[0].GetOperator()
		valAddr2 = s.validators[1].GetOperator()

		defaultCallArgs = contracts.CallArgs{
			ContractAddr: s.precompile.Address(),
			ContractABI:  s.precompile.ABI,
			PrivKey:      s.privKey,
		}
		defaultApproveArgs = defaultCallArgs.WithMethodName(authorization.ApproveMethod)

		defaultLogCheck = testutil.LogCheckArgs{ABIEvents: s.precompile.ABI.Events}
		passCheck = defaultLogCheck.WithExpPass(true)
		outOfGasCheck = defaultLogCheck.WithErrContains(vm.ErrOutOfGas.Error())
	})

	Describe("when the precompile is not enabled in the EVM params", func() {
		It("should return an error", func() {
			// disable the precompile
			params := s.app.EvmKeeper.GetParams(s.ctx)
			var activePrecompiles []string
			for _, precompile := range params.ActivePrecompiles {
				if precompile != s.precompile.Address().String() {
					activePrecompiles = append(activePrecompiles, precompile)
				}
			}
			params.ActivePrecompiles = activePrecompiles
			err := s.app.EvmKeeper.SetParams(s.ctx, params)
			Expect(err).To(BeNil(), "error while setting params")

			// try to call the precompile
			delegateArgs := defaultCallArgs.
				WithMethodName(staking.DelegateMethod).
				WithArgs(
					s.address, valAddr.String(), big.NewInt(2e18),
				)

			failCheck := defaultLogCheck.
				WithErrContains("precompile not enabled")

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, failCheck)
			Expect(err).To(HaveOccurred(), "expected error while calling the precompile")
			Expect(err.Error()).To(ContainSubstring("precompile not enabled"))
		})
	})

	Describe("Revert transaction", func() {
		It("should run out of gas if the gas limit is too low", func() {
			outOfGasArgs := defaultApproveArgs.
				WithGasLimit(30000).
				WithArgs(
					s.precompile.Address(),
					abi.MaxUint256,
					[]string{staking.DelegateMsg},
				)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, outOfGasArgs, outOfGasCheck)
			Expect(err).To(HaveOccurred(), "error while calling precompile")
		})
	})

	Describe("Execute approve transaction", func() {
		// TODO: enable once we check that the spender is not the origin
		// It("should return error if the origin is the spender", func() {
		//	args := defaultApproveArgs.WithArgs(
		//		s.address,
		//		abi.MaxUint256,
		//		[]string{staking.DelegateMsg},
		//	)
		//
		//	differentOriginCheck := defaultLogCheck.WithErrContains(cmn.ErrDifferentOrigin, s.address, addr)
		//
		//	_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, args, differentOriginCheck)
		//	Expect(err).To(BeNil(), "error while calling precompile")
		// })

		It("should return error if the staking method is not supported on the precompile", func() {
			approveArgs := defaultApproveArgs.WithArgs(
				s.precompile.Address(), abi.MaxUint256, []string{distribution.DelegationRewardsMethod},
			)

			logCheckArgs := defaultLogCheck.WithErrContains(
				cmn.ErrInvalidMsgType, "staking", distribution.DelegationRewardsMethod,
			)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, approveArgs, logCheckArgs)
			Expect(err).To(HaveOccurred(), "error while calling the contract and checking logs")
		})

		It("should approve the delegate method with the max uint256 value", func() {
			s.SetupApproval(
				s.privKey, s.precompile.Address(), abi.MaxUint256, []string{staking.DelegateMsg},
			)

			s.ExpectAuthorization(staking.DelegateAuthz, s.precompile.Address(), s.address, nil)
		})

		It("should approve the undelegate method with 1 evmos", func() {
			s.SetupApproval(
				s.privKey, s.precompile.Address(), big.NewInt(1e18), []string{staking.UndelegateMsg},
			)

			s.ExpectAuthorization(staking.UndelegateAuthz, s.precompile.Address(), s.address, &oneE18Coin)
		})

		It("should approve the redelegate method with 2 evmos", func() {
			s.SetupApproval(
				s.privKey, s.precompile.Address(), big.NewInt(2e18), []string{staking.RedelegateMsg},
			)

			s.ExpectAuthorization(staking.RedelegateAuthz, s.precompile.Address(), s.address, &twoE18Coin)
		})

		It("should approve the cancel unbonding delegation method with 1 evmos", func() {
			s.SetupApproval(
				s.privKey, s.precompile.Address(), big.NewInt(1e18), []string{staking.CancelUnbondingDelegationMsg},
			)

			s.ExpectAuthorization(staking.CancelUnbondingDelegationAuthz, s.precompile.Address(), s.address, &oneE18Coin)
		})
	})

	Describe("Execute increase allowance transaction", func() {
		// defaultIncreaseArgs are the default arguments to call the increase allowance method.
		//
		// NOTE: this has to be populated in BeforeEach, because the private key is not initialized outside of it.
		var defaultIncreaseArgs contracts.CallArgs

		BeforeEach(func() {
			s.SetupApproval(
				s.privKey, s.precompile.Address(), big.NewInt(1e18), []string{staking.DelegateMsg},
			)

			defaultIncreaseArgs = defaultCallArgs.WithMethodName(authorization.IncreaseAllowanceMethod)
		})

		// TODO: enable once we check that the spender is not the origin
		// It("should return error if the origin is the spender", func() {
		//	increaseArgs := defaultCallArgs.
		//		WithMethodName(authorization.IncreaseAllowanceMethod).
		//		WithArgs(
		//			s.address, big.NewInt(1e18), []string{staking.DelegateMsg},
		//		)
		//
		//	_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, increaseArgs, differentOriginCheck)
		//	Expect(err).To(BeNil(), "error while calling the contract and checking logs")
		// })

		It("Should increase the allowance of the delegate method with 1 evmos", func() {
			increaseArgs := defaultCallArgs.
				WithMethodName(authorization.IncreaseAllowanceMethod).
				WithArgs(
					s.precompile.Address(), big.NewInt(1e18), []string{staking.DelegateMsg},
				)

			logCheckArgs := passCheck.WithExpEvents(authorization.EventTypeAllowanceChange)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, increaseArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while calling the contract and checking logs")

			s.ExpectAuthorization(staking.DelegateAuthz, s.precompile.Address(), s.address, &twoE18Coin)
		})

		It("should return error if the allowance to increase does not exist", func() {
			increaseArgs := defaultIncreaseArgs.WithArgs(
				s.precompile.Address(), big.NewInt(1e18), []string{staking.UndelegateMsg},
			)

			logCheckArgs := defaultLogCheck.WithErrContains(
				"does not exist",
			)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, increaseArgs, logCheckArgs)
			Expect(err).To(HaveOccurred(), "error while calling the contract and checking logs")
			Expect(err.Error()).To(ContainSubstring("does not exist"))

			authz, _ := s.CheckAuthorization(staking.UndelegateAuthz, s.precompile.Address(), s.address)
			Expect(authz).To(BeNil(), "expected authorization to not be set")
		})
	})

	Describe("Execute decrease allowance transaction", func() {
		// defaultDecreaseArgs are the default arguments to call the decrease allowance method.
		//
		// NOTE: this has to be populated in BeforeEach, because the private key is not initialized outside of it.
		var defaultDecreaseArgs contracts.CallArgs

		BeforeEach(func() {
			s.SetupApproval(
				s.privKey, s.precompile.Address(), big.NewInt(2e18), []string{staking.DelegateMsg},
			)

			defaultDecreaseArgs = defaultCallArgs.WithMethodName(authorization.DecreaseAllowanceMethod)
		})

		// TODO: enable once we check that the spender is not the origin
		// It("should return error if the origin is the spender", func() {
		//	addr, _ := testutiltx.NewAddrKey()
		//	decreaseArgs := defaultDecreaseArgs.WithArgs(
		//		s.precompile.Address(), big.NewInt(1e18), []string{staking.DelegateMsg},
		//	)
		//
		//	logCheckArgs := defaultLogCheck.WithErrContains(
		//		cmn.ErrDifferentOrigin, s.address, addr,
		//	)
		//
		//	_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, decreaseArgs, logCheckArgs)
		//	Expect(err).To(BeNil(), "error while calling the contract and checking logs")
		// })

		It("Should decrease the allowance of the delegate method with 1 evmos", func() {
			decreaseArgs := defaultDecreaseArgs.WithArgs(
				s.precompile.Address(), big.NewInt(1e18), []string{staking.DelegateMsg},
			)

			logCheckArgs := passCheck.WithExpEvents(authorization.EventTypeAllowanceChange)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, decreaseArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while calling the contract and checking logs")

			s.ExpectAuthorization(staking.DelegateAuthz, s.precompile.Address(), s.address, &oneE18Coin)
		})

		It("should return error if the allowance to decrease does not exist", func() {
			decreaseArgs := defaultDecreaseArgs.WithArgs(
				s.precompile.Address(), big.NewInt(1e18), []string{staking.UndelegateMsg},
			)

			logCheckArgs := defaultLogCheck.WithErrContains(
				"does not exist",
			)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, decreaseArgs, logCheckArgs)
			Expect(err).To(HaveOccurred(), "error while calling the contract and checking logs")
			Expect(err.Error()).To(ContainSubstring("does not exist"))

			authz, _ := s.CheckAuthorization(staking.UndelegateAuthz, s.precompile.Address(), s.address)
			Expect(authz).To(BeNil(), "expected authorization to not be set")
		})
	})

	Describe("to revoke an approval", func() {
		var (
			// defaultRevokeArgs are the default arguments to call the revoke method.
			//
			// NOTE: this has to be populated in BeforeEach, because the default call args are not initialized outside of it.
			defaultRevokeArgs contracts.CallArgs

			// granteeAddr is the address of the grantee used in the revocation tests.
			granteeAddr = testutiltx.GenerateAddress()
		)

		BeforeEach(func() {
			defaultRevokeArgs = defaultCallArgs.WithMethodName(authorization.RevokeMethod)
		})

		It("should revoke the approval when executing as the granter", func() {
			typeURLs := []string{staking.DelegateMsg}

			s.SetupApproval(
				s.privKey, granteeAddr, abi.MaxUint256, typeURLs,
			)
			s.ExpectAuthorization(staking.DelegateAuthz, granteeAddr, s.address, nil)

			revokeArgs := defaultRevokeArgs.WithArgs(
				granteeAddr, typeURLs,
			)

			revocationCheck := passCheck.WithExpEvents(authorization.EventTypeRevocation)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, revokeArgs, revocationCheck)
			Expect(err).To(BeNil(), "error while calling the contract and checking logs")

			// check that the authorization is revoked
			authz, _ := s.CheckAuthorization(staking.DelegateAuthz, granteeAddr, s.address)
			Expect(authz).To(BeNil(), "expected authorization to be revoked")
		})

		It("should not revoke the approval when trying to revoke for a different message type", func() {
			typeURLs := []string{staking.DelegateMsg}

			s.SetupApproval(
				s.privKey, granteeAddr, abi.MaxUint256, typeURLs,
			)
			s.ExpectAuthorization(staking.DelegateAuthz, granteeAddr, s.address, nil)

			revokeArgs := defaultRevokeArgs.WithArgs(
				granteeAddr, []string{staking.UndelegateMsg},
			)

			notFoundCheck := defaultLogCheck.
				WithErrContains("failed to delete grant")

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, revokeArgs, notFoundCheck)
			Expect(err).To(HaveOccurred(), "error while calling the contract and checking logs")

			// the authorization should still be there.
			s.ExpectAuthorization(staking.DelegateAuthz, granteeAddr, s.address, nil)
		})

		It("should return error if the approval does not exist", func() {
			revokeArgs := defaultRevokeArgs.WithArgs(
				s.address, []string{staking.DelegateMsg},
			)

			notFoundCheck := defaultLogCheck.
				WithErrContains("failed to delete grant")

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, revokeArgs, notFoundCheck)
			Expect(err).To(HaveOccurred(), "error while calling the contract and checking logs")
		})

		It("should not revoke the approval if sent by someone else than the granter", func() {
			typeURLs := []string{staking.DelegateMsg}

			// set up an approval with a different key than the one used to sign the transaction.
			differentAddr, differentPriv := testutiltx.NewAddrKey()
			err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, differentAddr.Bytes(), 1e18)
			Expect(err).To(BeNil(), "error while funding account")

			s.NextBlock()
			s.SetupApproval(
				differentPriv, granteeAddr, abi.MaxUint256, typeURLs,
			)
			s.ExpectAuthorization(staking.DelegateAuthz, granteeAddr, differentAddr, nil)

			revokeArgs := defaultRevokeArgs.WithArgs(
				differentAddr, typeURLs,
			)

			notFoundCheck := defaultLogCheck.
				WithErrContains("failed to delete grant")

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, revokeArgs, notFoundCheck)
			Expect(err).To(HaveOccurred(), "error while calling the contract and checking logs")

			// the authorization should still be set
			s.ExpectAuthorization(staking.DelegateAuthz, granteeAddr, differentAddr, nil)
		})
	})

	Describe("to create validator", func() {
		var (
			defaultDescription = staking.Description{
				Moniker:         "new node",
				Identity:        "",
				Website:         "",
				SecurityContact: "",
				Details:         "",
			}
			defaultCommission = staking.Commission{
				Rate:          big.NewInt(100000000000000000),
				MaxRate:       big.NewInt(100000000000000000),
				MaxChangeRate: big.NewInt(100000000000000000),
			}
			defaultMinSelfDelegation = big.NewInt(1)
			defaultPubkeyBase64Str   = GenerateBase64PubKey()
			defaultValue             = big.NewInt(1)

			// defaultCreateValidatorArgs are the default arguments for the createValidator call
			//
			// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
			defaultCreateValidatorArgs contracts.CallArgs
		)

		BeforeEach(func() {
			// populate the default createValidator args
			defaultCreateValidatorArgs = defaultCallArgs.WithMethodName(staking.CreateValidatorMethod)
		})

		Context("when validator address is the origin", func() {
			It("should succeed", func() {
				createValidatorArgs := defaultCreateValidatorArgs.WithArgs(
					defaultDescription, defaultCommission, defaultMinSelfDelegation, s.address, defaultPubkeyBase64Str, defaultValue,
				)

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeCreateValidator)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createValidatorArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				_, found := s.app.StakingKeeper.GetValidator(s.ctx, s.address.Bytes())
				Expect(found).To(BeTrue(), "expected validator to be found")
			})
		})

		Context("when validator address is not the origin", func() {
			It("should fail", func() {
				differentAddr := testutiltx.GenerateAddress()

				createValidatorArgs := defaultCreateValidatorArgs.WithArgs(
					defaultDescription, defaultCommission, defaultMinSelfDelegation, differentAddr, defaultPubkeyBase64Str, defaultValue,
				)

				logCheckArgs := defaultLogCheck.WithErrContains(
					fmt.Sprintf(staking.ErrDifferentOriginFromDelegator, s.address, differentAddr),
				)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, createValidatorArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
			})
		})
	})

	Describe("to edit validator", func() {
		var (
			defaultDescription = staking.Description{
				Moniker:         "edit node",
				Identity:        "[do-not-modify]",
				Website:         "[do-not-modify]",
				SecurityContact: "[do-not-modify]",
				Details:         "[do-not-modify]",
			}
			defaultCommissionRate    = big.NewInt(staking.DoNotModifyCommissionRate)
			defaultMinSelfDelegation = big.NewInt(staking.DoNotModifyMinSelfDelegation)

			// defaultEditValidatorArgs are the default arguments for the editValidator call
			//
			// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
			defaultEditValidatorArgs contracts.CallArgs
		)

		BeforeEach(func() {
			// populate the default editValidator args
			defaultEditValidatorArgs = defaultCallArgs.WithMethodName(staking.EditValidatorMethod)
		})

		Context("when origin is equal to validator address", func() {
			It("should succeed", func() {
				// create a new validator
				newAddr, newPriv := testutiltx.NewAccAddressAndKey()
				hexAddr := common.BytesToAddress(newAddr.Bytes())
				err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, newAddr, 2e18)
				Expect(err).To(BeNil(), "error while funding account: %v", err)

				description := staking.Description{
					Moniker:         "new node",
					Identity:        "",
					Website:         "",
					SecurityContact: "",
					Details:         "",
				}
				commission := staking.Commission{
					Rate:          big.NewInt(100000000000000000),
					MaxRate:       big.NewInt(100000000000000000),
					MaxChangeRate: big.NewInt(100000000000000000),
				}
				minSelfDelegation := big.NewInt(1)
				pubkeyBase64Str := "UuhHQmkUh2cPBA6Rg4ei0M2B04cVYGNn/F8SAUsYIb4="
				value := big.NewInt(1e18)

				createValidatorArgs := defaultCallArgs.WithMethodName(staking.CreateValidatorMethod).
					WithPrivKey(newPriv).
					WithArgs(description, commission, minSelfDelegation, hexAddr, pubkeyBase64Str, value)

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeCreateValidator)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, createValidatorArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				s.NextBlock()

				// edit validator
				editValidatorArgs := defaultEditValidatorArgs.
					WithPrivKey(newPriv).
					WithArgs(defaultDescription, hexAddr, defaultCommissionRate, defaultMinSelfDelegation)

				logCheckArgs = passCheck.WithExpEvents(staking.EventTypeEditValidator)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, editValidatorArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				validator, found := s.app.StakingKeeper.GetValidator(s.ctx, newAddr.Bytes())
				Expect(found).To(BeTrue(), "expected validator to be found")
				Expect(validator.Description.Moniker).To(Equal(defaultDescription.Moniker), "expected validator moniker is updated")
				// Other fields should not be modified due to the value "[do-not-modify]".
				Expect(validator.Description.Identity).To(Equal(description.Identity), "expected validator identity not to be updated")
				Expect(validator.Description.Website).To(Equal(description.Website), "expected validator website not to be updated")
				Expect(validator.Description.SecurityContact).To(Equal(description.SecurityContact), "expected validator security contact not to be updated")
				Expect(validator.Description.Details).To(Equal(description.Details), "expected validator details not to be updated")

				Expect(validator.Commission.Rate.BigInt().String()).To(Equal(commission.Rate.String()), "expected validator commission rate remain unchanged")
				Expect(validator.Commission.MaxRate.BigInt().String()).To(Equal(commission.MaxRate.String()), "expected validator max commission rate remain unchanged")
				Expect(validator.Commission.MaxChangeRate.BigInt().String()).To(Equal(commission.MaxChangeRate.String()), "expected validator max change rate remain unchanged")
				Expect(validator.MinSelfDelegation.String()).To(Equal(minSelfDelegation.String()), "expected validator min self delegation remain unchanged")
			})
		})

		Context("with origin different than validator address", func() {
			It("should fail", func() {
				editValidatorArgs := defaultEditValidatorArgs.WithArgs(
					defaultDescription, common.BytesToAddress(valAddr.Bytes()), defaultCommissionRate, defaultMinSelfDelegation,
				)

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeEditValidator)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, editValidatorArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
			})
		})
	})

	Describe("to delegate", func() {
		var (
			// prevDelegation is the delegation that is available prior to the test (an initial delegation is
			// added in the test suite setup).
			prevDelegation stakingtypes.Delegation
			// defaultDelegateArgs are the default arguments for the delegate call
			//
			// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
			defaultDelegateArgs contracts.CallArgs
		)

		BeforeEach(func() {
			// get the delegation that is available prior to the test
			prevDelegation, _ = s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), valAddr)

			// populate the default delegate args
			defaultDelegateArgs = defaultCallArgs.WithMethodName(staking.DelegateMethod)
		})

		Context("as the token owner", func() {
			It("should delegate without need for authorization", func() {
				delegateArgs := defaultDelegateArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(2e18),
				)

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				delegation, found := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), valAddr)
				Expect(found).To(BeTrue(), "expected delegation to be found")
				expShares := prevDelegation.GetShares().Add(math.LegacyNewDec(2))
				Expect(delegation.GetShares()).To(Equal(expShares), "expected different delegation shares")
			})

			It("should not delegate if the account has no sufficient balance", func() {
				// send funds away from account to only have target balance remaining
				balance := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
				targetBalance := math.NewInt(1e17)
				sentBalance := balance.Amount.Sub(targetBalance)
				newAddr, _ := testutiltx.NewAccAddressAndKey()
				err := s.app.BankKeeper.SendCoins(s.ctx, s.address.Bytes(), newAddr,
					sdk.Coins{sdk.Coin{Denom: s.bondDenom, Amount: sentBalance}})
				Expect(err).To(BeNil(), "error while sending coins")

				// try to delegate more than left in account
				delegateArgs := defaultDelegateArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(1e18),
				)

				logCheckArgs := defaultLogCheck.WithErrContains("insufficient funds")

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
				Expect(err.Error()).To(ContainSubstring("insufficient funds"))
			})

			It("should not delegate if the validator does not exist", func() {
				nonExistingAddr := testutiltx.GenerateAddress()
				nonExistingValAddr := sdk.ValAddress(nonExistingAddr.Bytes())

				delegateArgs := defaultDelegateArgs.WithArgs(
					s.address, nonExistingValAddr.String(), big.NewInt(2e18),
				)

				logCheckArgs := defaultLogCheck.WithErrContains("validator does not exist")

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
				Expect(err.Error()).To(ContainSubstring("validator does not exist"))
			})
		})

		Context("on behalf of another account", func() {
			It("should not delegate if delegator address is not the origin", func() {
				differentAddr := testutiltx.GenerateAddress()

				delegateArgs := defaultDelegateArgs.WithArgs(
					differentAddr, valAddr.String(), big.NewInt(2e18),
				)

				logCheckArgs := defaultLogCheck.WithErrContains(
					fmt.Sprintf(staking.ErrDifferentOriginFromDelegator, s.address, differentAddr),
				)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
			})
		})
	})

	Describe("to undelegate", func() {
		// defaultUndelegateArgs are the default arguments for the undelegate call
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
		var defaultUndelegateArgs contracts.CallArgs

		BeforeEach(func() {
			defaultUndelegateArgs = defaultCallArgs.WithMethodName(staking.UndelegateMethod)
		})

		Context("as the token owner", func() {
			It("should undelegate without need for authorization", func() {
				undelegations := s.app.StakingKeeper.GetUnbondingDelegationsFromValidator(s.ctx, s.validators[0].GetOperator())
				Expect(undelegations).To(HaveLen(0), "expected no unbonding delegations before test")

				undelegateArgs := defaultUndelegateArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(1e18),
				)

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeUnbond)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, undelegateArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				undelegations = s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
				Expect(undelegations).To(HaveLen(1), "expected one undelegation")
				Expect(undelegations[0].ValidatorAddress).To(Equal(valAddr.String()), "expected validator address to be %s", valAddr)
			})

			It("should not undelegate if the amount exceeds the delegation", func() {
				undelegateArgs := defaultUndelegateArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(2e18),
				)

				logCheckArgs := defaultLogCheck.WithErrContains("invalid shares amount")

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, undelegateArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
				Expect(err.Error()).To(ContainSubstring("invalid shares amount"))
			})

			It("should not undelegate if the validator does not exist", func() {
				nonExistingAddr := testutiltx.GenerateAddress()
				nonExistingValAddr := sdk.ValAddress(nonExistingAddr.Bytes())

				undelegateArgs := defaultUndelegateArgs.WithArgs(
					s.address, nonExistingValAddr.String(), big.NewInt(1e18),
				)

				logCheckArgs := defaultLogCheck.WithErrContains("validator does not exist")

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, undelegateArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
				Expect(err.Error()).To(ContainSubstring("validator does not exist"))
			})
		})

		Context("on behalf of another account", func() {
			It("should not undelegate if delegator address is not the origin", func() {
				differentAddr := testutiltx.GenerateAddress()

				undelegateArgs := defaultUndelegateArgs.WithArgs(
					differentAddr, valAddr.String(), big.NewInt(1e18),
				)

				logCheckArgs := defaultLogCheck.WithErrContains(
					fmt.Sprintf(staking.ErrDifferentOriginFromDelegator, s.address, differentAddr),
				)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, undelegateArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
			})
		})
	})

	Describe("to redelegate", func() {
		// defaultRedelegateArgs are the default arguments for the redelegate call
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
		var defaultRedelegateArgs contracts.CallArgs

		BeforeEach(func() {
			defaultRedelegateArgs = defaultCallArgs.WithMethodName(staking.RedelegateMethod)
		})

		Context("as the token owner", func() {
			It("should redelegate without need for authorization", func() {
				redelegateArgs := defaultRedelegateArgs.WithArgs(
					s.address, valAddr.String(), valAddr2.String(), big.NewInt(1e18),
				)

				logCheckArgs := passCheck.
					WithExpEvents(staking.EventTypeRedelegate)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegateArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				redelegations := s.app.StakingKeeper.GetAllRedelegations(s.ctx, s.address.Bytes(), valAddr, valAddr2)
				Expect(redelegations).To(HaveLen(1), "expected one redelegation to be found")
				bech32Addr := sdk.AccAddress(s.address.Bytes())
				Expect(redelegations[0].DelegatorAddress).To(Equal(bech32Addr.String()), "expected delegator address to be %s", s.address)
				Expect(redelegations[0].ValidatorSrcAddress).To(Equal(valAddr.String()), "expected source validator address to be %s", valAddr)
				Expect(redelegations[0].ValidatorDstAddress).To(Equal(valAddr2.String()), "expected destination validator address to be %s", valAddr2)
			})

			It("should not redelegate if the amount exceeds the delegation", func() {
				redelegateArgs := defaultRedelegateArgs.WithArgs(
					s.address, valAddr.String(), valAddr2.String(), big.NewInt(2e18),
				)

				logCheckArgs := defaultLogCheck.WithErrContains("invalid shares amount")

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegateArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
				Expect(err.Error()).To(ContainSubstring("invalid shares amount"))
			})

			It("should not redelegate if the validator does not exist", func() {
				nonExistingAddr := testutiltx.GenerateAddress()
				nonExistingValAddr := sdk.ValAddress(nonExistingAddr.Bytes())

				redelegateArgs := defaultRedelegateArgs.WithArgs(
					s.address, valAddr.String(), nonExistingValAddr.String(), big.NewInt(1e18),
				)

				logCheckArgs := defaultLogCheck.WithErrContains("redelegation destination validator not found")

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegateArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
				Expect(err.Error()).To(ContainSubstring("redelegation destination validator not found"))
			})
		})

		Context("on behalf of another account", func() {
			It("should not redelegate if delegator address is not the origin", func() {
				differentAddr := testutiltx.GenerateAddress()

				redelegateArgs := defaultRedelegateArgs.WithArgs(
					differentAddr, valAddr.String(), valAddr2.String(), big.NewInt(1e18),
				)

				logCheckArgs := defaultLogCheck.WithErrContains(
					fmt.Sprintf(staking.ErrDifferentOriginFromDelegator, s.address, differentAddr),
				)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegateArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
			})
		})
	})

	Describe("to cancel an unbonding delegation", func() {
		var (
			// defaultCancelUnbondingArgs are the default arguments for the cancelUnbondingDelegation call
			//
			// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
			defaultCancelUnbondingArgs contracts.CallArgs

			// expCreationHeight is the expected creation height of the unbonding delegation
			expCreationHeight = int64(3)
		)

		BeforeEach(func() {
			defaultCancelUnbondingArgs = defaultCallArgs.WithMethodName(staking.CancelUnbondingDelegationMethod)

			// Set up an unbonding delegation
			undelegateArgs := defaultCallArgs.
				WithMethodName(staking.UndelegateMethod).
				WithArgs(s.address, valAddr.String(), big.NewInt(1e18))

			logCheckArgs := passCheck.
				WithExpEvents(staking.EventTypeUnbond)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, undelegateArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while setting up an unbonding delegation: %v", err)

			s.NextBlock()

			// Check that the unbonding delegation was created
			unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
			Expect(unbondingDelegations).To(HaveLen(1), "expected one unbonding delegation to be found")
			Expect(unbondingDelegations[0].DelegatorAddress).To(Equal(sdk.AccAddress(s.address.Bytes()).String()), "expected delegator address to be %s", s.address)
			Expect(unbondingDelegations[0].ValidatorAddress).To(Equal(valAddr.String()), "expected validator address to be %s", valAddr)
			Expect(unbondingDelegations[0].Entries).To(HaveLen(1), "expected one unbonding delegation entry to be found")
			Expect(unbondingDelegations[0].Entries[0].CreationHeight).To(Equal(expCreationHeight), "expected different creation height")
			Expect(unbondingDelegations[0].Entries[0].Balance).To(Equal(math.NewInt(1e18)), "expected different balance")
		})

		Context("as the token owner", func() {
			It("should cancel unbonding delegation", func() {
				delegations := s.app.StakingKeeper.GetValidatorDelegations(s.ctx, s.validators[0].GetOperator())
				Expect(delegations).To(HaveLen(0))

				cArgs := defaultCancelUnbondingArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(1e18), big.NewInt(expCreationHeight),
				)

				logCheckArgs := passCheck.
					WithExpEvents(staking.EventTypeCancelUnbondingDelegation)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
				Expect(unbondingDelegations).To(HaveLen(0), "expected unbonding delegation to be canceled")

				delegations = s.app.StakingKeeper.GetValidatorDelegations(s.ctx, s.validators[0].GetOperator())
				Expect(delegations).To(HaveLen(1), "expected one delegation to be found")
			})

			It("should not cancel an unbonding delegation if the amount is not correct", func() {
				cArgs := defaultCancelUnbondingArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(2e18), big.NewInt(expCreationHeight),
				)

				logCheckArgs := defaultLogCheck.WithErrContains("amount is greater than the unbonding delegation entry balance")

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
				Expect(err.Error()).To(ContainSubstring("amount is greater than the unbonding delegation entry balance"))

				unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
				Expect(unbondingDelegations).To(HaveLen(1), "expected unbonding delegation not to have been canceled")
			})

			It("should not cancel an unbonding delegation if the creation height is not correct", func() {
				cArgs := defaultCancelUnbondingArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(1e18), big.NewInt(expCreationHeight+1),
				)

				logCheckArgs := defaultLogCheck.WithErrContains("unbonding delegation entry is not found at block height")

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
				Expect(err.Error()).To(ContainSubstring("unbonding delegation entry is not found at block height"))

				unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
				Expect(unbondingDelegations).To(HaveLen(1), "expected unbonding delegation not to have been canceled")
			})
		})
	})

	Describe("Calling precompile txs from a vesting account", func() {
		var (
			funder          common.Address
			vestAcc         common.Address
			vestAccPriv     *ethsecp256k1.PrivKey
			clawbackAccount *vestingtypes.ClawbackVestingAccount
			unvested        sdk.Coins
			vested          sdk.Coins
			// unlockedVested are unlocked vested coins of the vesting schedule
			unlockedVested      sdk.Coins
			defaultDelegateArgs contracts.CallArgs
		)

		BeforeEach(func() {
			// Setup vesting account
			funder = s.address
			vestAcc, vestAccPriv = testutiltx.NewAddrKey()
			vestingAmtTotal := evmosutil.TestVestingSchedule.TotalVestingCoins

			clawbackAccount = s.setupVestingAccount(funder.Bytes(), vestAcc.Bytes())

			// Check if all tokens are unvested at vestingStart
			unvested = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
			vested = clawbackAccount.GetVestedCoins(s.ctx.BlockTime())
			Expect(vestingAmtTotal).To(Equal(unvested))
			Expect(vested.IsZero()).To(BeTrue())

			// populate the default delegate args
			defaultDelegateArgs = defaultCallArgs.WithMethodName(staking.DelegateMethod)
			defaultDelegateArgs = defaultDelegateArgs.WithPrivKey(vestAccPriv)
		})

		Context("before first vesting period - all tokens locked and unvested", func() {
			BeforeEach(func() {
				s.NextBlock()

				// Ensure no tokens are vested
				vested = clawbackAccount.GetVestedCoins(s.ctx.BlockTime())
				unvested = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
				unlocked := clawbackAccount.GetUnlockedCoins(s.ctx.BlockTime())
				zeroCoins := sdk.NewCoins(sdk.NewCoin(s.bondDenom, math.ZeroInt()))
				Expect(vested).To(Equal(zeroCoins), "expected different vested coins")
				Expect(unvested).To(Equal(evmosutil.TestVestingSchedule.TotalVestingCoins), "expected different unvested coins")
				Expect(unlocked).To(Equal(zeroCoins), "expected different unlocked coins")
			})

			It("Should not be able to delegate unvested tokens", func() {
				delegateArgs := defaultDelegateArgs.WithArgs(
					vestAcc, valAddr.String(), unvested.AmountOf(s.bondDenom).BigInt(),
				)

				failCheck := defaultLogCheck.
					WithErrContains("cannot delegate unvested coins")

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, failCheck)
				Expect(err).NotTo(BeNil(), "error while calling the smart contract: %v", err)
				Expect(err.Error()).To(ContainSubstring("coins available for delegation < delegation amount"))
			})

			It("Should be able to delegate tokens not involved in vesting schedule", func() {
				// send some coins to the vesting account
				coinsToDelegate := sdk.NewCoins(sdk.NewCoin(s.bondDenom, math.NewInt(1e18)))
				err := evmosutil.FundAccount(s.ctx, s.app.BankKeeper, clawbackAccount.GetAddress(), coinsToDelegate)
				Expect(err).To(BeNil())

				// check balance is updated
				balance := s.app.BankKeeper.GetBalance(s.ctx, clawbackAccount.GetAddress(), s.bondDenom)
				Expect(balance).To(Equal(accountGasCoverage[0].Add(evmosutil.TestVestingSchedule.TotalVestingCoins[0]).Add(coinsToDelegate[0])))

				delegateArgs := defaultDelegateArgs.WithArgs(
					vestAcc, valAddr.String(), coinsToDelegate.AmountOf(s.bondDenom).BigInt(),
				)

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				delegation, found := s.app.StakingKeeper.GetDelegation(s.ctx, vestAcc.Bytes(), valAddr)
				Expect(found).To(BeTrue(), "expected delegation to be found")
				Expect(delegation.Shares.BigInt()).To(Equal(coinsToDelegate[0].Amount.BigInt()))

				// check vesting balance is untouched
				balancePost := s.app.BankKeeper.GetBalance(s.ctx, clawbackAccount.GetAddress(), s.bondDenom)
				Expect(balancePost.IsGTE(evmosutil.TestVestingSchedule.TotalVestingCoins[0])).To(BeTrue())
			})
		})

		Context("after first vesting period and before lockup - some vested tokens, but still all locked", func() {
			BeforeEach(func() {
				// Surpass cliff but none of lockup duration
				cliffDuration := time.Duration(evmosutil.TestVestingSchedule.CliffPeriodLength)
				s.NextBlockAfter(cliffDuration * time.Second)

				// Check if some, but not all tokens are vested
				vested = clawbackAccount.GetVestedCoins(s.ctx.BlockTime())
				expVested := sdk.NewCoins(sdk.NewCoin(s.bondDenom, evmosutil.TestVestingSchedule.VestedCoinsPerPeriod[0].Amount.Mul(math.NewInt(evmosutil.TestVestingSchedule.CliffMonths))))
				Expect(vested).NotTo(Equal(evmosutil.TestVestingSchedule.TotalVestingCoins), "expected some tokens to have been vested")
				Expect(vested).To(Equal(expVested), "expected different vested amount")

				// check the vested tokens are still locked
				unlockedVested = clawbackAccount.GetUnlockedVestedCoins(s.ctx.BlockTime())
				Expect(unlockedVested).To(Equal(sdk.Coins{}))

				vestingAmtTotal := evmosutil.TestVestingSchedule.TotalVestingCoins
				res, err := s.app.VestingKeeper.Balances(s.ctx, &vestingtypes.QueryBalancesRequest{Address: clawbackAccount.Address})
				Expect(err).To(BeNil())
				Expect(res.Vested).To(Equal(expVested))
				Expect(res.Unvested).To(Equal(vestingAmtTotal.Sub(expVested...)))
				// All coins from vesting schedule should be locked
				Expect(res.Locked).To(Equal(vestingAmtTotal))
			})

			It("Should be able to delegate locked vested tokens", func() {
				delegateArgs := defaultDelegateArgs.WithArgs(
					vestAcc, valAddr.String(), vested[0].Amount.BigInt(),
				)

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				delegation, found := s.app.StakingKeeper.GetDelegation(s.ctx, vestAcc.Bytes(), valAddr)
				Expect(found).To(BeTrue(), "expected delegation to be found")
				Expect(delegation.Shares.BigInt()).To(Equal(vested[0].Amount.BigInt()))
			})

			It("Should be able to delegate locked vested tokens + free tokens (not in vesting schedule)", func() {
				// send some coins to the vesting account
				amt := sdk.NewCoins(sdk.NewCoin(s.bondDenom, math.NewInt(1e18)))
				err := evmosutil.FundAccount(s.ctx, s.app.BankKeeper, clawbackAccount.GetAddress(), amt)
				Expect(err).To(BeNil())

				// check balance is updated
				balance := s.app.BankKeeper.GetBalance(s.ctx, clawbackAccount.GetAddress(), s.bondDenom)
				Expect(balance).To(Equal(accountGasCoverage[0].Add(evmosutil.TestVestingSchedule.TotalVestingCoins[0]).Add(amt[0])))

				coinsToDelegate := amt.Add(vested...)

				delegateArgs := defaultDelegateArgs.WithArgs(
					vestAcc, valAddr.String(), coinsToDelegate[0].Amount.BigInt(),
				)

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				delegation, found := s.app.StakingKeeper.GetDelegation(s.ctx, vestAcc.Bytes(), valAddr)
				Expect(found).To(BeTrue(), "expected delegation to be found")
				Expect(delegation.Shares.BigInt()).To(Equal(coinsToDelegate[0].Amount.BigInt()))
			})
		})

		Context("Between first and second lockup periods - vested coins are unlocked", func() {
			BeforeEach(func() {
				// Surpass first lockup
				vestDuration := time.Duration(evmosutil.TestVestingSchedule.LockupPeriodLength)
				s.NextBlockAfter(vestDuration * time.Second)

				// Check if some, but not all tokens are vested and unlocked
				vested = clawbackAccount.GetVestedCoins(s.ctx.BlockTime())
				unlocked := clawbackAccount.GetUnlockedCoins(s.ctx.BlockTime())
				unlockedVested = clawbackAccount.GetUnlockedVestedCoins(s.ctx.BlockTime())

				expVested := sdk.NewCoins(sdk.NewCoin(s.bondDenom, evmosutil.TestVestingSchedule.VestedCoinsPerPeriod[0].Amount.Mul(math.NewInt(evmosutil.TestVestingSchedule.LockupMonths))))
				expUnlockedVested := expVested

				Expect(vested).NotTo(Equal(evmosutil.TestVestingSchedule.TotalVestingCoins), "expected not all tokens to be vested")
				Expect(vested).To(Equal(expVested), "expected different amount of vested tokens")
				// all vested coins are unlocked
				Expect(unlockedVested).To(Equal(vested))
				Expect(unlocked).To(Equal(evmosutil.TestVestingSchedule.UnlockedCoinsPerLockup))
				Expect(unlockedVested).To(Equal(expUnlockedVested))
			})
			It("Should be able to delegate unlocked vested tokens", func() {
				delegateArgs := defaultDelegateArgs.WithArgs(
					vestAcc, valAddr.String(), unlockedVested[0].Amount.BigInt(),
				)

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				delegation, found := s.app.StakingKeeper.GetDelegation(s.ctx, vestAcc.Bytes(), valAddr)
				Expect(found).To(BeTrue(), "expected delegation to be found")
				Expect(delegation.Shares.BigInt()).To(Equal(unlockedVested[0].Amount.BigInt()))
			})

			It("Cannot delegate more than vested tokens (and free tokens)", func() {
				// calculate the delegatable amount
				balance := s.app.BankKeeper.GetBalance(s.ctx, vestAcc.Bytes(), s.bondDenom)
				unvestedOnly := clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
				delegatable := balance.Sub(unvestedOnly[0])

				delegateArgs := defaultDelegateArgs.WithArgs(
					vestAcc, valAddr.String(), delegatable.Amount.Add(sdk.OneInt()).BigInt(),
				)

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, logCheckArgs)
				Expect(err).NotTo(BeNil(), "error while calling the smart contract: %v", err)
				Expect(err.Error()).To(ContainSubstring("cannot delegate unvested coins"))

				_, found := s.app.StakingKeeper.GetDelegation(s.ctx, vestAcc.Bytes(), valAddr)
				Expect(found).To(BeFalse(), "expected delegation NOT to be found")
			})
		})
	})

	Describe("to query allowance", func() {
		var (
			defaultAllowanceArgs contracts.CallArgs

			differentAddr = testutiltx.GenerateAddress()
		)

		BeforeEach(func() {
			defaultAllowanceArgs = defaultCallArgs.WithMethodName(authorization.AllowanceMethod)
		})

		It("should return an empty allowance if none is set", func() {
			allowanceArgs := defaultAllowanceArgs.WithArgs(
				s.address, differentAddr, staking.CancelUnbondingDelegationMsg,
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, allowanceArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var allowanceInt *big.Int
			err = s.precompile.UnpackIntoInterface(&allowanceInt, "allowance", ethRes.Ret)
			Expect(err).To(BeNil(), "error while unmarshalling the allowance: %v", err)
			Expect(allowanceInt.Int64()).To(BeZero(), "expected allowance to be zero")
		})

		It("should return the granted allowance if set", func() {
			// setup approval for another address
			s.SetupApproval(
				s.privKey, differentAddr, big.NewInt(1e18), []string{staking.CancelUnbondingDelegationMsg},
			)

			// query allowance
			allowanceArgs := defaultAllowanceArgs.WithArgs(
				differentAddr, s.address, staking.CancelUnbondingDelegationMsg,
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, allowanceArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var allowanceInt *big.Int
			err = s.precompile.UnpackIntoInterface(&allowanceInt, "allowance", ethRes.Ret)
			Expect(err).To(BeNil(), "error while unmarshalling the allowance: %v", err)
			Expect(allowanceInt).To(Equal(big.NewInt(1e18)), "expected allowance to be 1e18")
		})
	})

	Describe("Validator queries", func() {
		// defaultValidatorArgs are the default arguments for the validator call
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
		var defaultValidatorArgs contracts.CallArgs

		BeforeEach(func() {
			defaultValidatorArgs = defaultCallArgs.WithMethodName(staking.ValidatorMethod)
		})

		It("should return validator", func() {
			varHexAddr := common.BytesToAddress(valAddr.Bytes())
			validatorArgs := defaultValidatorArgs.WithArgs(
				varHexAddr,
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, validatorArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)
			Expect(valOut.Validator.OperatorAddress).To(Equal(varHexAddr.String()), "expected validator address to match")
			Expect(valOut.Validator.DelegatorShares).To(Equal(big.NewInt(1e18)), "expected different delegator shares")
		})

		It("should return an empty validator if the validator is not found", func() {
			newValHexAddr := testutiltx.GenerateAddress()
			validatorArgs := defaultValidatorArgs.WithArgs(
				newValHexAddr,
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, validatorArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)
			Expect(valOut.Validator.OperatorAddress).To(Equal(""), "expected validator address to be empty")
			Expect(valOut.Validator.Status).To(BeZero(), "expected unspecified bonding status")
		})
	})

	Describe("Validators queries", func() {
		var defaultValidatorArgs contracts.CallArgs

		BeforeEach(func() {
			defaultValidatorArgs = defaultCallArgs.WithMethodName(staking.ValidatorsMethod)
		})

		It("should return validators (default pagination)", func() {
			validatorArgs := defaultValidatorArgs.WithArgs(
				stakingtypes.Bonded.String(),
				query.PageRequest{},
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, validatorArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorsOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)

			Expect(valOut.PageResponse.NextKey).To(BeEmpty())
			Expect(valOut.PageResponse.Total).To(Equal(uint64(len(s.validators))))

			Expect(valOut.Validators).To(HaveLen(len(s.validators)), "expected two validators to be returned")
			// return order can change, that's why each validator is checked individually
			for _, val := range valOut.Validators {
				s.CheckValidatorOutput(val)
			}
		})

		//nolint:dupl // this is a duplicate of the test for smart contract calls to the precompile
		It("should return validators w/pagination limit = 1", func() {
			const limit uint64 = 1
			validatorArgs := defaultValidatorArgs.WithArgs(
				stakingtypes.Bonded.String(),
				query.PageRequest{
					Limit:      limit,
					CountTotal: true,
				},
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, validatorArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorsOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)

			// no pagination, should return default values
			Expect(valOut.PageResponse.NextKey).NotTo(BeEmpty())
			Expect(valOut.PageResponse.Total).To(Equal(uint64(len(s.validators))))

			Expect(valOut.Validators).To(HaveLen(int(limit)), "expected one validator to be returned")

			// return order can change, that's why each validator is checked individually
			for _, val := range valOut.Validators {
				s.CheckValidatorOutput(val)
			}
		})

		It("should return an error if the bonding type is not known", func() {
			validatorArgs := defaultValidatorArgs.WithArgs(
				"15", // invalid bonding type
				query.PageRequest{},
			)

			invalidStatusCheck := defaultLogCheck.WithErrContains("invalid validator status 15")

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, validatorArgs, invalidStatusCheck)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
			Expect(err.Error()).To(ContainSubstring("invalid validator status 15"))
		})

		It("should return an empty array if there are no validators with the given bonding type", func() {
			validatorArgs := defaultValidatorArgs.WithArgs(
				stakingtypes.Unbonded.String(),
				query.PageRequest{},
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, validatorArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorsOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)

			Expect(valOut.PageResponse.NextKey).To(BeEmpty())
			Expect(valOut.PageResponse.Total).To(Equal(uint64(0)))
			Expect(valOut.Validators).To(HaveLen(0), "expected no validators to be returned")
		})
	})

	Describe("Delegation queries", func() {
		var defaultDelegationArgs contracts.CallArgs

		BeforeEach(func() {
			defaultDelegationArgs = defaultCallArgs.WithMethodName(staking.DelegationMethod)
		})

		It("should return a delegation if it is found", func() {
			delegationArgs := defaultDelegationArgs.WithArgs(
				s.address,
				valAddr.String(),
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegationArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var delOut staking.DelegationOutput
			err = s.precompile.UnpackIntoInterface(&delOut, staking.DelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the delegation output: %v", err)
			Expect(delOut.Shares).To(Equal(big.NewInt(1e18)), "expected different shares")
			Expect(delOut.Balance).To(Equal(cmn.Coin{Denom: s.bondDenom, Amount: big.NewInt(1e18)}), "expected different shares")
		})

		It("should return an empty delegation if it is not found", func() {
			newValAddr := sdk.ValAddress(testutiltx.GenerateAddress().Bytes())
			delegationArgs := defaultDelegationArgs.WithArgs(
				s.address, newValAddr.String(),
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegationArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var delOut staking.DelegationOutput
			err = s.precompile.UnpackIntoInterface(&delOut, staking.DelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the delegation output: %v", err)
			Expect(delOut.Shares.Int64()).To(BeZero(), "expected no shares")
			Expect(delOut.Balance.Denom).To(Equal(s.bondDenom), "expected different denomination")
			Expect(delOut.Balance.Amount.Int64()).To(BeZero(), "expected a zero amount")
		})
	})

	Describe("UnbondingDelegation queries", func() {
		var (
			defaultUnbondingDelegationArgs contracts.CallArgs

			// undelAmount is the amount of tokens to be unbonded
			undelAmount = big.NewInt(1e17)
		)

		BeforeEach(func() {
			defaultUnbondingDelegationArgs = defaultCallArgs.WithMethodName(staking.UnbondingDelegationMethod)

			// unbond a delegation
			s.SetupApproval(s.privKey, s.precompile.Address(), abi.MaxUint256, []string{staking.UndelegateMsg})

			unbondArgs := defaultCallArgs.
				WithMethodName(staking.UndelegateMethod).
				WithArgs(s.address, valAddr.String(), undelAmount)
			unbondCheck := passCheck.WithExpEvents(staking.EventTypeUnbond)
			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, unbondArgs, unbondCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			// check that the unbonding delegation exists
			unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
			Expect(unbondingDelegations).To(HaveLen(1), "expected one unbonding delegation")
		})

		It("should return an unbonding delegation if it is found", func() {
			unbondingDelegationsArgs := defaultUnbondingDelegationArgs.WithArgs(
				s.address,
				valAddr.String(),
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, unbondingDelegationsArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var unbondingDelegationOutput staking.UnbondingDelegationOutput
			err = s.precompile.UnpackIntoInterface(&unbondingDelegationOutput, staking.UnbondingDelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the unbonding delegation output: %v", err)
			Expect(unbondingDelegationOutput.UnbondingDelegation.Entries).To(HaveLen(1), "expected one unbonding delegation entry")
			// TODO: why are initial balance and balance the same always?
			Expect(unbondingDelegationOutput.UnbondingDelegation.Entries[0].InitialBalance).To(Equal(undelAmount), "expected different initial balance")
			Expect(unbondingDelegationOutput.UnbondingDelegation.Entries[0].Balance).To(Equal(undelAmount), "expected different balance")
		})

		It("should return an empty slice if the unbonding delegation is not found", func() {
			unbondingDelegationsArgs := defaultUnbondingDelegationArgs.WithArgs(
				s.address,
				valAddr2.String(),
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, unbondingDelegationsArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var unbondingDelegationOutput staking.UnbondingDelegationOutput
			err = s.precompile.UnpackIntoInterface(&unbondingDelegationOutput, staking.UnbondingDelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the unbonding delegation output: %v", err)
			Expect(unbondingDelegationOutput.UnbondingDelegation.Entries).To(HaveLen(0), "expected one unbonding delegation entry")
		})
	})

	Describe("to query a redelegation", func() {
		var defaultRedelegationArgs contracts.CallArgs

		BeforeEach(func() {
			defaultRedelegationArgs = defaultCallArgs.WithMethodName(staking.RedelegationMethod)
		})

		It("should return the redelegation if it exists", func() {
			// approve the redelegation
			s.SetupApproval(s.privKey, s.precompile.Address(), abi.MaxUint256, []string{staking.RedelegateMsg})

			// create a redelegation
			redelegateArgs := defaultCallArgs.
				WithMethodName(staking.RedelegateMethod).
				WithArgs(s.address, valAddr.String(), valAddr2.String(), big.NewInt(1e17))

			redelegateCheck := passCheck.WithExpEvents(staking.EventTypeRedelegate)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegateArgs, redelegateCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			// query the redelegation
			redelegationArgs := defaultRedelegationArgs.WithArgs(
				s.address,
				valAddr.String(),
				valAddr2.String(),
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegationArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var redelegationOutput staking.RedelegationOutput
			err = s.precompile.UnpackIntoInterface(&redelegationOutput, staking.RedelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the redelegation output: %v", err)
			Expect(redelegationOutput.Redelegation.Entries).To(HaveLen(1), "expected one redelegation entry")
			Expect(redelegationOutput.Redelegation.Entries[0].InitialBalance).To(Equal(big.NewInt(1e17)), "expected different initial balance")
			Expect(redelegationOutput.Redelegation.Entries[0].SharesDst).To(Equal(big.NewInt(1e17)), "expected different balance")
		})

		It("should return an empty output if the redelegation is not found", func() {
			redelegationArgs := defaultRedelegationArgs.WithArgs(
				s.address,
				valAddr.String(),
				valAddr2.String(),
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegationArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var redelegationOutput staking.RedelegationOutput
			err = s.precompile.UnpackIntoInterface(&redelegationOutput, staking.RedelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the redelegation output: %v", err)
			Expect(redelegationOutput.Redelegation.Entries).To(HaveLen(0), "expected no redelegation entries")
		})
	})

	Describe("Redelegations queries", func() {
		var (
			// defaultRedelegationsArgs are the default arguments for the redelegations query
			//
			// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
			defaultRedelegationsArgs contracts.CallArgs

			// delAmt is the amount of tokens to be delegated
			delAmt = big.NewInt(3e17)
			// redelTotalCount is the total number of redelegations
			redelTotalCount uint64 = 1
		)

		BeforeEach(func() {
			defaultRedelegationsArgs = defaultCallArgs.WithMethodName(staking.RedelegationsMethod)
			// create some redelegations
			s.SetupApproval(
				s.privKey, s.precompile.Address(), abi.MaxUint256, []string{staking.RedelegateMsg},
			)

			defaultRedelegateArgs := defaultCallArgs.WithMethodName(staking.RedelegateMethod)
			redelegationsArgs := []contracts.CallArgs{
				defaultRedelegateArgs.WithArgs(
					s.address, valAddr.String(), valAddr2.String(), delAmt,
				),
				defaultRedelegateArgs.WithArgs(
					s.address, valAddr.String(), valAddr2.String(), delAmt,
				),
			}

			logCheckArgs := passCheck.
				WithExpEvents(staking.EventTypeRedelegate)

			for _, args := range redelegationsArgs {
				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, args, logCheckArgs)
				Expect(err).To(BeNil(), "error while creating redelegation: %v", err)
			}
		})

		It("should return all redelegations for delegator (default pagination)", func() {
			redelegationArgs := defaultRedelegationsArgs.WithArgs(
				s.address,
				"",
				"",
				query.PageRequest{},
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegationArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var redelOut staking.RedelegationsOutput
			err = s.precompile.UnpackIntoInterface(&redelOut, staking.RedelegationsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)

			Expect(redelOut.PageResponse.NextKey).To(BeEmpty())
			Expect(redelOut.PageResponse.Total).To(Equal(redelTotalCount))

			Expect(redelOut.Response).To(HaveLen(int(redelTotalCount)), "expected two redelegations to be returned")
			// return order can change
			redOrder := []int{0, 1}
			if len(redelOut.Response[0].Entries) == 2 {
				redOrder = []int{1, 0}
			}

			for i, r := range redelOut.Response {
				Expect(r.Entries).To(HaveLen(redOrder[i] + 1))
			}
		})

		It("should return all redelegations for delegator w/pagination", func() {
			// make 2 queries
			// 1st one with pagination limit = 1
			// 2nd using the next page key
			var nextPageKey []byte
			for i := 0; i < 2; i++ {
				var pagination query.PageRequest
				if nextPageKey == nil {
					pagination.Limit = 1
					pagination.CountTotal = true
				} else {
					pagination.Key = nextPageKey
				}
				redelegationArgs := defaultRedelegationsArgs.WithArgs(
					s.address,
					"",
					"",
					pagination,
				)

				_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegationArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var redelOut staking.RedelegationsOutput
				err = s.precompile.UnpackIntoInterface(&redelOut, staking.RedelegationsMethod, ethRes.Ret)
				Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)

				if nextPageKey == nil {
					nextPageKey = redelOut.PageResponse.NextKey
					Expect(redelOut.PageResponse.Total).To(Equal(redelTotalCount))
				} else {
					Expect(redelOut.PageResponse.NextKey).To(BeEmpty())
					Expect(redelOut.PageResponse.Total).To(Equal(uint64(1)))
				}

				Expect(redelOut.Response).To(HaveLen(1), "expected two redelegations to be returned")
				// return order can change
				redOrder := []int{0, 1}
				if len(redelOut.Response[0].Entries) == 2 {
					redOrder = []int{1, 0}
				}

				for i, r := range redelOut.Response {
					Expect(r.Entries).To(HaveLen(redOrder[i] + 1))
				}
			}
		})

		It("should return an empty array if no redelegation is found for the given source validator", func() {
			// NOTE: the way that the functionality is implemented in the Cosmos SDK, the following combinations are
			// possible (see https://github.com/evmos/cosmos-sdk/blob/e773cf768844c87245d0c737cda1893a2819dd89/x/staking/keeper/querier.go#L361-L373):
			//
			// - delegator is NOT empty, source validator is empty, destination validator is empty
			//   --> filtering for all redelegations of the given delegator
			// - delegator is empty, source validator is NOT empty, destination validator is empty
			//   --> filtering for all redelegations with the given source validator
			// - delegator is NOT empty, source validator is NOT empty, destination validator is NOT empty
			//   --> filtering for all redelegations with the given combination of delegator, source and destination validator
			redelegationsArgs := defaultRedelegationsArgs.WithArgs(
				common.Address{}, // passing in an empty address to filter for all redelegations from valAddr2
				valAddr2.String(),
				"",
				query.PageRequest{},
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegationsArgs, passCheck)
			Expect(err).To(BeNil(), "expected error while calling the smart contract")

			var redelOut staking.RedelegationsOutput
			err = s.precompile.UnpackIntoInterface(&redelOut, staking.RedelegationsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)

			Expect(redelOut.PageResponse.NextKey).To(BeEmpty())
			Expect(redelOut.PageResponse.Total).To(BeZero(), "expected no redelegations to be returned")

			Expect(redelOut.Response).To(HaveLen(0), "expected no redelegations to be returned")
		})
	})

	It("Should refund leftover gas", func() {
		balancePre := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
		gasPrice := big.NewInt(1e9)

		// Call the precompile with a lot of gas
		approveArgs := defaultApproveArgs.
			WithGasPrice(gasPrice).
			WithArgs(s.precompile.Address(), big.NewInt(1e18), []string{staking.DelegateMsg})

		approvalCheck := passCheck.WithExpEvents(authorization.EventTypeApproval)

		res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, approveArgs, approvalCheck)
		Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

		s.NextBlock()

		balancePost := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
		difference := balancePre.Sub(balancePost)

		// NOTE: the expected difference is the gas price multiplied by the gas used, because the rest should be refunded
		expDifference := gasPrice.Int64() * res.GasUsed
		Expect(difference.Amount.Int64()).To(Equal(expDifference), "expected different total transaction cost")
	})
})

var _ = Describe("Calling staking precompile via Solidity", func() {
	var (
		// contractAddr is the address of the smart contract that will be deployed
		contractAddr    common.Address
		contractTwoAddr common.Address
		stkReverterAddr common.Address

		// stakingCallerContract is the contract instance calling into the staking precompile
		stakingCallerContract    evmtypes.CompiledContract
		stakingCallerTwoContract evmtypes.CompiledContract
		stakingReverterContract  evmtypes.CompiledContract

		// approvalCheck is a configuration for the log checker to see if an approval event was emitted.
		approvalCheck testutil.LogCheckArgs
		// execRevertedCheck defines the default log checking arguments which include the
		// standard revert message
		execRevertedCheck testutil.LogCheckArgs
		// err is a basic error type
		err error

		// nonExistingAddr is an address that does not exist in the state of the test suite
		nonExistingAddr = testutiltx.GenerateAddress()
		// nonExistingVal is a validator address that does not exist in the state of the test suite
		nonExistingVal = sdk.ValAddress(nonExistingAddr.Bytes())

		testContractInitialBalance = math.NewInt(100)
	)

	BeforeEach(func() {
		s.SetupTest()

		stakingCallerContract, err = testdata.LoadStakingCallerContract()
		Expect(err).To(BeNil(), "error while loading the staking caller contract: %v", err)

		contractAddr, err = s.DeployContract(stakingCallerContract)
		Expect(err).To(BeNil(), "error while deploying the smart contract: %v", err)
		s.NextBlock()

		// Deploy StakingCallerTwo contract
		stakingCallerTwoContract, err = testdata.LoadStakingCallerTwoContract()
		Expect(err).To(BeNil(), "error while loading the StakingCallerTwo contract")

		contractTwoAddr, err = s.DeployContract(stakingCallerTwoContract)
		Expect(err).To(BeNil(), "error while deploying the StakingCallerTwo contract")
		s.NextBlock()

		// Deploy StakingReverter contract
		stakingReverterContract, err = contracts.LoadStakingReverterContract()
		Expect(err).To(BeNil(), "error while loading the StakingReverter contract")

		stkReverterAddr, err = s.DeployContract(stakingReverterContract)
		Expect(err).To(BeNil(), "error while deploying the StakingReverter contract")
		s.NextBlock()

		// send some funds to the StakingCallerTwo & StakingReverter contracts to transfer to the
		// delegator during the tx
		err = evmosutil.FundAccount(s.ctx, s.app.BankKeeper, contractTwoAddr.Bytes(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, testContractInitialBalance)))
		Expect(err).To(BeNil(), "error while funding the smart contract: %v", err)

		err = evmosutil.FundAccount(s.ctx, s.app.BankKeeper, stkReverterAddr.Bytes(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, testContractInitialBalance)))
		Expect(err).To(BeNil(), "error while funding the smart contract: %v", err)

		valAddr = s.validators[0].GetOperator()
		valAddr2 = s.validators[1].GetOperator()

		// check contract was correctly deployed
		cAcc := s.app.EvmKeeper.GetAccount(s.ctx, contractAddr)
		Expect(cAcc).ToNot(BeNil(), "contract account should exist")
		Expect(cAcc.IsContract()).To(BeTrue(), "account should be a contract")

		// populate default call args
		defaultCallArgs = contracts.CallArgs{
			ContractAddr: contractAddr,
			ContractABI:  stakingCallerContract.ABI,
			PrivKey:      s.privKey,
		}
		// populate default approval args
		defaultApproveArgs = defaultCallArgs.WithMethodName("testApprove")

		// populate default log check args
		defaultLogCheck = testutil.LogCheckArgs{
			ABIEvents: s.precompile.Events,
		}
		execRevertedCheck = defaultLogCheck.WithErrContains(vm.ErrExecutionReverted.Error())
		passCheck = defaultLogCheck.WithExpPass(true)
		approvalCheck = passCheck.WithExpEvents(authorization.EventTypeApproval)
	})

	Describe("when the precompile is not enabled in the EVM params", func() {
		It("should return an error", func() {
			// disable the precompile
			params := s.app.EvmKeeper.GetParams(s.ctx)
			var activePrecompiles []string
			for _, precompile := range params.ActivePrecompiles {
				if precompile != s.precompile.Address().String() {
					activePrecompiles = append(activePrecompiles, precompile)
				}
			}
			params.ActivePrecompiles = activePrecompiles
			err := s.app.EvmKeeper.SetParams(s.ctx, params)
			Expect(err).To(BeNil(), "error while setting params")

			// try to call the precompile
			delegateArgs := defaultCallArgs.
				WithMethodName("testDelegate").
				WithArgs(
					s.address, valAddr.String(), big.NewInt(2e18),
				)

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "expected error while calling the precompile")
			Expect(err.Error()).To(ContainSubstring(vm.ErrExecutionReverted.Error()))
		})
	})

	Context("approving methods", func() {
		Context("with valid input", func() {
			It("should approve one method", func() {
				approvalArgs := defaultApproveArgs.WithArgs(
					contractAddr, []string{staking.DelegateMsg}, big.NewInt(1e18),
				)
				s.SetupApprovalWithContractCalls(approvalArgs)
			})

			It("should approve all methods", func() {
				approvalArgs := defaultApproveArgs.
					WithGasLimit(1e8).
					WithArgs(
						contractAddr,
						[]string{staking.DelegateMsg, staking.RedelegateMsg, staking.UndelegateMsg, staking.CancelUnbondingDelegationMsg},
						big.NewInt(1e18),
					)
				s.SetupApprovalWithContractCalls(approvalArgs)
			})

			It("should update a previous approval", func() {
				approvalArgs := defaultApproveArgs.WithArgs(
					contractAddr, []string{staking.DelegateMsg}, big.NewInt(1e18),
				)
				s.SetupApprovalWithContractCalls(approvalArgs)

				s.NextBlock()

				// update approval
				approvalArgs = defaultApproveArgs.WithArgs(
					contractAddr, []string{staking.DelegateMsg}, big.NewInt(2e18),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, approvalArgs, approvalCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				// check approvals
				authorization, expirationTime := s.CheckAuthorization(staking.DelegateAuthz, contractAddr, s.address)
				Expect(authorization).ToNot(BeNil(), "expected authorization to not be nil")
				Expect(expirationTime).ToNot(BeNil(), "expected expiration time to not be nil")
				Expect(authorization.MsgTypeURL()).To(Equal(staking.DelegateMsg), "expected authorization msg type url to be %s", staking.DelegateMsg)
				Expect(authorization.MaxTokens.Amount).To(Equal(math.NewInt(2e18)), "expected different max tokens after updated approval")
			})

			It("should remove approval when setting amount to zero", func() {
				s.SetupApprovalWithContractCalls(
					defaultApproveArgs.WithArgs(contractAddr, []string{staking.DelegateMsg}, big.NewInt(1e18)),
				)

				s.NextBlock()

				// check approvals pre-removal
				allAuthz, err := s.app.AuthzKeeper.GetAuthorizations(s.ctx, contractAddr.Bytes(), s.address.Bytes())
				Expect(err).To(BeNil(), "error while reading authorizations")
				Expect(allAuthz).To(HaveLen(1), "expected no authorizations")

				approveArgs := defaultApproveArgs.WithArgs(
					contractAddr, []string{staking.DelegateMsg}, big.NewInt(0),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, approveArgs, approvalCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract")

				// check approvals after approving with amount 0
				allAuthz, err = s.app.AuthzKeeper.GetAuthorizations(s.ctx, contractAddr.Bytes(), s.address.Bytes())
				Expect(err).To(BeNil(), "error while reading authorizations")
				Expect(allAuthz).To(HaveLen(0), "expected no authorizations")
			})

			It("should not approve if the gas is not enough", func() {
				approveArgs := defaultApproveArgs.
					WithGasLimit(1e5).
					WithArgs(
						contractAddr,
						[]string{
							staking.DelegateMsg,
							staking.UndelegateMsg,
							staking.RedelegateMsg,
							staking.CancelUnbondingDelegationMsg,
						},
						big.NewInt(1e18),
					)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, approveArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract")
			})
		})

		Context("with invalid input", func() {
			// TODO: enable once we check that origin is not the sender
			// It("shouldn't approve any methods for if the sender is the origin", func() {
			//	approveArgs := defaultApproveArgs.WithArgs(
			//		nonExistingAddr, []string{staking.DelegateMsg}, big.NewInt(1e18),
			//	)
			//
			//	_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, approveArgs, execRevertedCheck)
			//	Expect(err).To(BeNil(), "error while calling the smart contract")
			//
			//	// check approvals
			//	allAuthz, err := s.app.AuthzKeeper.GetAuthorizations(s.ctx, contractAddr.Bytes(), s.address.Bytes())
			//	Expect(err).To(BeNil(), "error while reading authorizations")
			//	Expect(allAuthz).To(HaveLen(0), "expected no authorizations")
			// })

			It("shouldn't approve for invalid methods", func() {
				approveArgs := defaultApproveArgs.WithArgs(
					contractAddr, []string{"invalid method"}, big.NewInt(1e18),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, approveArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract")

				// check approvals
				allAuthz, err := s.app.AuthzKeeper.GetAuthorizations(s.ctx, contractAddr.Bytes(), s.address.Bytes())
				Expect(err).To(BeNil(), "error while reading authorizations")
				Expect(allAuthz).To(HaveLen(0), "expected no authorizations")
			})
		})
	})

	Context("to revoke an approval", func() {
		var defaultRevokeArgs contracts.CallArgs

		BeforeEach(func() {
			defaultRevokeArgs = defaultCallArgs.WithMethodName("testRevoke")
		})

		It("should revoke when sending as the granter", func() {
			// set up an approval to be revoked
			cArgs := defaultApproveArgs.WithArgs(
				contractAddr, []string{staking.DelegateMsg}, big.NewInt(1e18),
			)
			s.SetupApprovalWithContractCalls(cArgs)

			s.NextBlock()

			revokeArgs := defaultRevokeArgs.WithArgs(contractAddr, []string{staking.DelegateMsg})

			revocationCheck := passCheck.WithExpEvents(authorization.EventTypeRevocation)

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, revokeArgs, revocationCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract")

			// check approvals
			authz, _ := s.CheckAuthorization(staking.DelegateAuthz, contractAddr, s.address)
			Expect(authz).To(BeNil(), "expected authorization to be revoked")
		})

		It("should not revoke when approval is issued by a different granter", func() {
			// Create a delegate authorization where the granter is a different account from the default test suite one
			createdAuthz := staking.DelegateAuthz
			granteeAddr := testutiltx.GenerateAddress()
			granterAddr := testutiltx.GenerateAddress()
			validators := s.app.StakingKeeper.GetLastValidators(s.ctx)
			valAddrs := make([]sdk.ValAddress, len(validators))
			for i, val := range validators {
				valAddrs[i] = val.GetOperator()
			}
			delegationAuthz, err := stakingtypes.NewStakeAuthorization(
				valAddrs,
				nil,
				createdAuthz,
				&sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(1e18)},
			)
			Expect(err).To(BeNil(), "failed to create authorization")

			expiration := s.ctx.BlockTime().Add(time.Hour * 24 * 365).UTC()
			err = s.app.AuthzKeeper.SaveGrant(s.ctx, granteeAddr.Bytes(), granterAddr.Bytes(), delegationAuthz, &expiration)
			Expect(err).ToNot(HaveOccurred(), "failed to save authorization")
			authz, _ := s.CheckAuthorization(createdAuthz, granteeAddr, granterAddr)
			Expect(authz).ToNot(BeNil(), "expected authorization to be created")

			revokeArgs := defaultRevokeArgs.WithArgs(granteeAddr, []string{staking.DelegateMsg})

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, revokeArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract")

			// check approvals
			authz, _ = s.CheckAuthorization(createdAuthz, granteeAddr, granterAddr)
			Expect(authz).ToNot(BeNil(), "expected authorization not to be revoked")
		})

		It("should revert the execution when no approval is found", func() {
			revokeArgs := defaultRevokeArgs.WithArgs(contractAddr, []string{staking.DelegateMsg})

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, revokeArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract")

			// check approvals
			authz, _ := s.CheckAuthorization(staking.DelegateAuthz, contractAddr, s.address)
			Expect(authz).To(BeNil(), "expected no authorization to be found")
		})

		It("should not revoke if the approval is for a different message type", func() {
			// set up an approval
			cArgs := defaultApproveArgs.WithArgs(
				contractAddr, []string{staking.DelegateMsg}, big.NewInt(1e18),
			)
			s.SetupApprovalWithContractCalls(cArgs)

			s.NextBlock()

			revokeArgs := defaultRevokeArgs.WithArgs(contractAddr, []string{staking.UndelegateMsg})

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, revokeArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract")

			// check approval is still there
			s.ExpectAuthorization(
				staking.DelegateAuthz,
				contractAddr,
				s.address,
				&sdk.Coin{Denom: s.bondDenom, Amount: math.NewInt(1e18)},
			)
		})
	})

	Context("create a validator", func() {
		var (
			valPriv *ethsecp256k1.PrivKey
			valAddr sdk.AccAddress

			defaultDescription = staking.Description{
				Moniker:         "new node",
				Identity:        "",
				Website:         "",
				SecurityContact: "",
				Details:         "",
			}
			defaultCommission = staking.Commission{
				Rate:          big.NewInt(100000000000000000),
				MaxRate:       big.NewInt(100000000000000000),
				MaxChangeRate: big.NewInt(100000000000000000),
			}
			defaultMinSelfDelegation = big.NewInt(1)
			defaultPubkeyBase64Str   = GenerateBase64PubKey()
			defaultValue             = big.NewInt(1e8)

			// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
			defaultCreateValidatorArgs contracts.CallArgs
		)

		BeforeEach(func() {
			defaultCreateValidatorArgs = defaultCallArgs.WithMethodName("testCreateValidator")
			valAddr, valPriv = testutiltx.NewAccAddressAndKey()
			err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, valAddr, 1e18)
			Expect(err).To(BeNil(), "error while funding account: %v", err)

			s.NextBlock()
		})

		It("tx from validator operator - should NOT create a validator", func() {
			cArgs := defaultCreateValidatorArgs.
				WithPrivKey(s.privKey).
				WithArgs(defaultDescription, defaultCommission, defaultMinSelfDelegation, s.address, defaultPubkeyBase64Str, defaultValue)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, execRevertedCheck)
			Expect(err).NotTo(BeNil(), "error while calling the smart contract: %v", err)

			_, found := s.app.StakingKeeper.GetValidator(s.ctx, s.address.Bytes())
			Expect(found).To(BeFalse(), "expected validator NOT to be found")
		})

		It("tx from another EOA - should create a validator fail", func() {
			cArgs := defaultCreateValidatorArgs.
				WithPrivKey(valPriv).
				WithArgs(defaultDescription, defaultCommission, defaultMinSelfDelegation, s.address, defaultPubkeyBase64Str, defaultValue)

			logCheckArgs := defaultLogCheck.WithErrContains(fmt.Sprintf(staking.ErrDifferentOriginFromDelegator, s.address.String(), common.BytesToAddress(valAddr.Bytes()).String()))

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, logCheckArgs)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

			_, found := s.app.StakingKeeper.GetValidator(s.ctx, s.address.Bytes())
			Expect(found).To(BeFalse(), "expected validator not to be found")
		})
	})

	Context("to edit a validator", func() {
		var (
			// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
			defaultEditValArgs contracts.CallArgs
			valPriv            *ethsecp256k1.PrivKey
			valAddr            sdk.AccAddress
			valHexAddr         common.Address

			defaultDescription = staking.Description{
				Moniker:         "edit node",
				Identity:        "[do-not-modify]",
				Website:         "[do-not-modify]",
				SecurityContact: "[do-not-modify]",
				Details:         "[do-not-modify]",
			}
			defaultCommissionRate    = big.NewInt(staking.DoNotModifyCommissionRate)
			defaultMinSelfDelegation = big.NewInt(staking.DoNotModifyMinSelfDelegation)

			minSelfDelegation = big.NewInt(1)

			description = staking.Description{}
			commission  = staking.Commission{}
		)

		BeforeEach(func() {
			defaultEditValArgs = defaultCallArgs.WithMethodName("testEditValidator")

			// create a new validator
			valAddr, valPriv = testutiltx.NewAccAddressAndKey()
			valHexAddr = common.BytesToAddress(valAddr.Bytes())
			err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, valAddr, 2e18)
			Expect(err).To(BeNil(), "error while funding account: %v", err)

			description = staking.Description{
				Moniker:         "original moniker",
				Identity:        "",
				Website:         "",
				SecurityContact: "",
				Details:         "",
			}
			commission = staking.Commission{
				Rate:          big.NewInt(100000000000000000),
				MaxRate:       big.NewInt(100000000000000000),
				MaxChangeRate: big.NewInt(100000000000000000),
			}
			pubkeyBase64Str := "UuhHQmkUh2cPBA6Rg4ei0M2B04cVYGNn/F8SAUsYIb4="
			value := big.NewInt(1e18)

			createValidatorArgs := contracts.CallArgs{
				ContractAddr: s.precompile.Address(),
				ContractABI:  s.precompile.ABI,
				MethodName:   staking.CreateValidatorMethod,
				PrivKey:      valPriv,
				Args:         []interface{}{description, commission, minSelfDelegation, valHexAddr, pubkeyBase64Str, value},
			}

			logCheckArgs := passCheck.WithExpEvents(staking.EventTypeCreateValidator)

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, createValidatorArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			s.NextBlock()
		})

		It("with tx from validator operator - should NOT edit a validator", func() {
			cArgs := defaultEditValArgs.
				WithPrivKey(valPriv).
				WithArgs(
					defaultDescription, valHexAddr,
					defaultCommissionRate, defaultMinSelfDelegation,
				)

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, execRevertedCheck)
			Expect(err).NotTo(BeNil(), "error while calling the smart contract: %v", err)

			validator, found := s.app.StakingKeeper.GetValidator(s.ctx, valAddr.Bytes())
			Expect(found).To(BeTrue(), "expected validator to be found")
			Expect(validator.Description.Moniker).NotTo(Equal(defaultDescription.Moniker), "expected validator moniker NOT to be updated")
		})

		It("with tx from another EOA - should fail", func() {
			cArgs := defaultEditValArgs.
				WithPrivKey(s.privKey).
				WithArgs(
					defaultDescription, valHexAddr,
					defaultCommissionRate, defaultMinSelfDelegation,
				)

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "expected error while calling the precompile")
			Expect(err.Error()).To(ContainSubstring(vm.ErrExecutionReverted.Error()))

			// validator should remain unchanged
			validator, found := s.app.StakingKeeper.GetValidator(s.ctx, valAddr.Bytes())
			Expect(found).To(BeTrue(), "expected validator to be found")
			Expect(validator.Description.Moniker).To(Equal("original moniker"), "expected validator moniker is updated")
			Expect(validator.Commission.Rate.BigInt().String()).To(Equal("100000000000000000"), "expected validator commission rate remain unchanged")
		})
	})

	Context("delegating", func() {
		var (
			// prevDelegation is the delegation that is available prior to the test (an initial delegation is
			// added in the test suite setup).
			prevDelegation stakingtypes.Delegation
			// defaultDelegateArgs are the default arguments for the delegate call
			//
			// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
			defaultDelegateArgs contracts.CallArgs
		)

		BeforeEach(func() {
			// get the delegation that is available prior to the test
			prevDelegation, _ = s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), valAddr)

			defaultDelegateArgs = defaultCallArgs.WithMethodName("testDelegate")
		})

		Context("without approval set", func() {
			BeforeEach(func() {
				authz, _ := s.CheckAuthorization(staking.DelegateAuthz, contractAddr, s.address)
				Expect(authz).To(BeNil(), "expected authorization to be nil")
			})

			It("should not delegate", func() {
				Expect(s.app.EvmKeeper.GetAccount(s.ctx, contractAddr)).ToNot(BeNil(), "expected contract to exist")

				cArgs := defaultDelegateArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(1e18),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				del, _ := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), valAddr)
				Expect(del).To(Equal(prevDelegation), "no new delegation to be found")
			})
		})

		Context("with approval set", func() {
			BeforeEach(func() {
				cArgs := defaultApproveArgs.WithArgs(
					contractAddr, []string{staking.DelegateMsg}, big.NewInt(1e18),
				)
				s.SetupApprovalWithContractCalls(cArgs)
			})

			It("should delegate when not exceeding the allowance", func() {
				cArgs := defaultDelegateArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(1e18),
				)

				logCheckArgs := passCheck.
					WithExpEvents(staking.EventTypeDelegate)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				delegation, found := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), valAddr)
				Expect(found).To(BeTrue(), "expected delegation to be found")
				expShares := prevDelegation.GetShares().Add(math.LegacyNewDec(1))
				Expect(delegation.GetShares()).To(Equal(expShares), "expected delegation shares to be 2")
			})

			Context("Calling the precompile from the StakingReverter contract", func() {
				var (
					txSenderInitialBal     sdk.Coin
					contractInitialBalance sdk.Coin
					gasPrice               = math.NewInt(1e9)
					delAmt                 = math.NewInt(1e18)
				)

				BeforeEach(func() {
					// set approval for the StakingReverter contract
					s.SetupApproval(s.privKey, stkReverterAddr, delAmt.BigInt(), []string{staking.DelegateMsg})

					txSenderInitialBal = s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
					contractInitialBalance = s.app.BankKeeper.GetBalance(s.ctx, stkReverterAddr.Bytes(), s.bondDenom)
				})

				It("should revert the changes and NOT delegate - successful tx", func() {
					callArgs := contracts.CallArgs{
						ContractAddr: stkReverterAddr,
						ContractABI:  stakingReverterContract.ABI,
						PrivKey:      s.privKey,
						MethodName:   "run",
						Args: []interface{}{
							big.NewInt(5), s.validators[0].GetOperator().String(),
						},
						GasPrice: gasPrice.BigInt(),
					}

					// Tx should be successful, but no state changes happened
					res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, callArgs, passCheck)
					Expect(err).To(BeNil())
					fees := gasPrice.MulRaw(res.GasUsed)

					// contract balance should remain unchanged
					contractFinalBalance := s.app.BankKeeper.GetBalance(s.ctx, stkReverterAddr.Bytes(), s.bondDenom)
					Expect(contractFinalBalance.Amount).To(Equal(contractInitialBalance.Amount))

					// No delegation should be created
					_, found := s.app.StakingKeeper.GetDelegation(s.ctx, stkReverterAddr.Bytes(), s.validators[0].GetOperator())
					Expect(found).To(BeFalse(), "expected NO delegation to be found")

					// Only fees deducted on tx sender
					txSenderFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
					Expect(txSenderFinalBal.Amount).To(Equal(txSenderInitialBal.Amount.Sub(fees)))
				})

				It("should revert the changes and NOT delegate - failed tx - max precompile calls reached", func() {
					callArgs := contracts.CallArgs{
						ContractAddr: stkReverterAddr,
						ContractABI:  stakingReverterContract.ABI,
						PrivKey:      s.privKey,
						MethodName:   "run",
						Args: []interface{}{
							big.NewInt(7), s.validators[0].GetOperator().String(),
						},
						GasPrice: gasPrice.BigInt(),
					}

					// Tx should fail due to MaxPrecompileCalls
					_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, callArgs, execRevertedCheck)
					Expect(err).NotTo(BeNil())

					// contract balance should remain unchanged
					contractFinalBalance := s.app.BankKeeper.GetBalance(s.ctx, stkReverterAddr.Bytes(), s.bondDenom)
					Expect(contractFinalBalance.Amount).To(Equal(contractInitialBalance.Amount))

					// No delegation should be created
					_, found := s.app.StakingKeeper.GetDelegation(s.ctx, stkReverterAddr.Bytes(), s.validators[0].GetOperator())
					Expect(found).To(BeFalse(), "expected NO delegation to be found")
				})
			})

			Context("Table-driven tests for Delegate method", func() {
				// testCase is a struct used for cases of contracts calls that have some operation
				// performed before and/or after the precompile call
				type testCase struct {
					before bool
					after  bool
				}

				var (
					args                           contracts.CallArgs
					delegatorInitialBal            sdk.Coin
					contractInitialBalance         sdk.Coin
					bondedTokensPoolInitialBalance sdk.Coin
					delAmt                         = math.NewInt(1e18)
					gasPrice                       = math.NewInt(1e9)
					bondedTokensPoolAccAddr        = authtypes.NewModuleAddress("bonded_tokens_pool")
				)

				BeforeEach(func() {
					// set authorization for contract
					callCArgs := contracts.CallArgs{
						ContractAddr: contractTwoAddr,
						ContractABI:  stakingCallerTwoContract.ABI,
						PrivKey:      s.privKey,
						MethodName:   "testApprove",
						Args: []interface{}{
							contractTwoAddr, []string{staking.DelegateMsg}, delAmt.BigInt(),
						},
					}

					s.SetupApprovalWithContractCalls(callCArgs)

					args = callCArgs.
						WithMethodName("testDelegateWithCounterAndTransfer").
						WithGasPrice(gasPrice.BigInt())

					delegatorInitialBal = s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
					contractInitialBalance = s.app.BankKeeper.GetBalance(s.ctx, contractTwoAddr.Bytes(), s.bondDenom)
					bondedTokensPoolInitialBalance = s.app.BankKeeper.GetBalance(s.ctx, bondedTokensPoolAccAddr, s.bondDenom)
				})

				DescribeTable("should delegate and update balances accordingly", func(tc testCase) {
					cArgs := args.
						WithArgs(
							s.address, valAddr.String(), delAmt.BigInt(), tc.before, tc.after,
						)

					// This is the amount of tokens transferred from the contract to the delegator
					// during the contract call
					transferToDelAmt := math.ZeroInt()
					for _, transferred := range []bool{tc.before, tc.after} {
						if transferred {
							transferToDelAmt = transferToDelAmt.AddRaw(15)
						}
					}

					logCheckArgs := passCheck.
						WithExpEvents(staking.EventTypeDelegate)

					res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
					fees := gasPrice.MulRaw(res.GasUsed)

					// check the contract's balance was deducted to fund the vesting account
					contractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractTwoAddr.Bytes(), s.bondDenom)
					Expect(contractFinalBal.Amount).To(Equal(contractInitialBalance.Amount.Sub(transferToDelAmt)))

					delegation, found := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), valAddr)
					Expect(found).To(BeTrue(), "expected delegation to be found")
					expShares := prevDelegation.GetShares().Add(math.LegacyNewDec(1))
					Expect(delegation.GetShares()).To(Equal(expShares), "expected delegation shares to be 2")

					delegatorFinalBal := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), s.bondDenom)
					Expect(delegatorFinalBal.Amount).To(Equal(delegatorInitialBal.Amount.Sub(fees).Sub(delAmt).Add(transferToDelAmt)))

					// check the bondedTokenPool is updated with the delegated tokens
					bondedTokensPoolFinalBalance := s.app.BankKeeper.GetBalance(s.ctx, bondedTokensPoolAccAddr, s.bondDenom)
					Expect(bondedTokensPoolFinalBalance.Amount).To(Equal(bondedTokensPoolInitialBalance.Amount.Add(delAmt)))
				},
					Entry("contract tx with transfer to delegator before and after precompile call ", testCase{
						before: true,
						after:  true,
					}),
					Entry("contract tx with transfer to delegator before precompile call ", testCase{
						before: true,
						after:  false,
					}),
					Entry("contract tx with transfer to delegator after precompile call ", testCase{
						before: false,
						after:  true,
					}),
				)

				It("should NOT delegate and update balances accordingly - internal transfer to tokens pool", func() {
					cArgs := args.
						WithMethodName("testDelegateWithTransfer").
						WithArgs(
							common.BytesToAddress(bondedTokensPoolAccAddr),
							s.address, valAddr.String(), delAmt.BigInt(), true, true,
						)

					_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, execRevertedCheck)
					Expect(err).NotTo(BeNil())

					// contract balance should remain unchanged
					contractFinalBal := s.app.BankKeeper.GetBalance(s.ctx, contractTwoAddr.Bytes(), s.bondDenom)
					Expect(contractFinalBal.Amount).To(Equal(contractInitialBalance.Amount))

					// check the bondedTokenPool should remain unchanged
					bondedTokensPoolFinalBalance := s.app.BankKeeper.GetBalance(s.ctx, bondedTokensPoolAccAddr, s.bondDenom)
					Expect(bondedTokensPoolFinalBalance.Amount).To(Equal(bondedTokensPoolInitialBalance.Amount))
				})
			})

			It("should not delegate when exceeding the allowance", func() {
				cArgs := defaultDelegateArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(2e18),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				del, _ := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), valAddr)
				Expect(del).To(Equal(prevDelegation), "no new delegation to be found")
			})

			It("should not delegate when sending from a different address", func() {
				newAddr, newPriv := testutiltx.NewAccAddressAndKey()
				err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, newAddr, 1e18)
				Expect(err).To(BeNil(), "error while funding account: %v", err)

				s.NextBlock()

				delegateArgs := defaultDelegateArgs.
					WithPrivKey(newPriv).
					WithArgs(s.address, valAddr.String(), big.NewInt(1e18))

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				del, _ := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), valAddr)
				Expect(del).To(Equal(prevDelegation), "no new delegation to be found")
			})

			It("should not delegate when validator does not exist", func() {
				delegateArgs := defaultDelegateArgs.WithArgs(
					s.address, nonExistingVal.String(), big.NewInt(1e18),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				del, _ := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), nonExistingVal)
				Expect(del).To(BeZero(), "expected no delegation to be found")
			})

			It("shouldn't delegate to a validator that is not in the allow list of the approval", func() {
				// create a new validator, which is not included in the active set of the last block
				testutil.CreateValidator(s.ctx, s.T(), s.privKey.PubKey(), *s.app.StakingKeeper.Keeper, math.NewInt(100))
				newValAddr := sdk.ValAddress(s.address.Bytes())

				delegateArgs := defaultDelegateArgs.WithArgs(
					s.address, newValAddr.String(), big.NewInt(2e18),
				)

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				delegation, _ := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), newValAddr)
				Expect(delegation.GetShares()).To(Equal(math.LegacyNewDecFromInt(math.NewInt(100))), "expected only the delegation from creating the validator, no more")
			})
		})

		Describe("delegation from a vesting account", func() {
			var (
				funder          common.Address
				vestAcc         common.Address
				vestAccPriv     *ethsecp256k1.PrivKey
				clawbackAccount *vestingtypes.ClawbackVestingAccount
				unvested        sdk.Coins
				vested          sdk.Coins
				// unlockedVested are unlocked vested coins of the vesting schedule
				unlockedVested sdk.Coins
				defaultArgs    contracts.CallArgs
			)

			BeforeEach(func() {
				// Setup vesting account
				funder = s.address
				vestAcc, vestAccPriv = testutiltx.NewAddrKey()

				clawbackAccount = s.setupVestingAccount(funder.Bytes(), vestAcc.Bytes())

				// Check if all tokens are unvested at vestingStart
				totalVestingCoins := evmosutil.TestVestingSchedule.TotalVestingCoins
				unvested = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
				vested = clawbackAccount.GetVestedCoins(s.ctx.BlockTime())
				Expect(unvested).To(Equal(totalVestingCoins))
				Expect(vested.IsZero()).To(BeTrue())

				// create approval to allow spending all vesting coins
				cArgs := defaultApproveArgs.WithArgs(
					contractAddr, []string{staking.DelegateMsg}, totalVestingCoins.AmountOf(s.bondDenom).BigInt(),
				).WithPrivKey(vestAccPriv)
				s.SetupApprovalWithContractCalls(cArgs)

				// add the vesting account priv key to the delegate args
				defaultArgs = defaultDelegateArgs.WithPrivKey(vestAccPriv)
			})

			Context("before first vesting period - all tokens locked and unvested", func() {
				BeforeEach(func() {
					s.NextBlock()

					// Ensure no tokens are vested
					vested = clawbackAccount.GetVestedCoins(s.ctx.BlockTime())
					unvested = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
					unlocked := clawbackAccount.GetUnlockedCoins(s.ctx.BlockTime())
					zeroCoins := sdk.NewCoins(sdk.NewCoin(s.bondDenom, math.ZeroInt()))
					Expect(vested).To(Equal(zeroCoins))
					Expect(unvested).To(Equal(evmosutil.TestVestingSchedule.TotalVestingCoins))
					Expect(unlocked).To(Equal(zeroCoins))
				})

				It("Should not be able to delegate unvested tokens", func() {
					delegateArgs := defaultArgs.WithArgs(
						vestAcc, valAddr.String(), unvested.AmountOf(s.bondDenom).BigInt(),
					)

					_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, execRevertedCheck)
					Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
				})

				It("Should be able to delegate tokens not involved in vesting schedule", func() {
					// send some coins to the vesting account
					coinsToDelegate := sdk.NewCoins(sdk.NewCoin(s.bondDenom, math.NewInt(1e18)))
					err := evmosutil.FundAccount(s.ctx, s.app.BankKeeper, clawbackAccount.GetAddress(), coinsToDelegate)
					Expect(err).To(BeNil())

					delegateArgs := defaultArgs.WithArgs(
						vestAcc, valAddr.String(), coinsToDelegate.AmountOf(s.bondDenom).BigInt(),
					)

					logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

					_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					delegation, found := s.app.StakingKeeper.GetDelegation(s.ctx, vestAcc.Bytes(), valAddr)
					Expect(found).To(BeTrue(), "expected delegation to be found")
					Expect(delegation.Shares.BigInt()).To(Equal(coinsToDelegate[0].Amount.BigInt()))
				})
			})

			Context("after first vesting period and before lockup - some vested tokens, but still all locked", func() {
				BeforeEach(func() {
					// Surpass cliff but none of lockup duration
					cliffDuration := time.Duration(evmosutil.TestVestingSchedule.CliffPeriodLength)
					s.NextBlockAfter(cliffDuration * time.Second)

					// Check if some, but not all tokens are vested
					vested = clawbackAccount.GetVestedCoins(s.ctx.BlockTime())
					expVested := sdk.NewCoins(sdk.NewCoin(s.bondDenom, evmosutil.TestVestingSchedule.VestedCoinsPerPeriod[0].Amount.Mul(math.NewInt(evmosutil.TestVestingSchedule.CliffMonths))))
					Expect(vested).NotTo(Equal(evmosutil.TestVestingSchedule.TotalVestingCoins))
					Expect(vested).To(Equal(expVested))

					// check the vested tokens are still locked
					unlockedVested = clawbackAccount.GetUnlockedVestedCoins(s.ctx.BlockTime())
					Expect(unlockedVested).To(Equal(sdk.Coins{}))

					vestingAmtTotal := evmosutil.TestVestingSchedule.TotalVestingCoins
					res, err := s.app.VestingKeeper.Balances(s.ctx, &vestingtypes.QueryBalancesRequest{Address: clawbackAccount.Address})
					Expect(err).To(BeNil())
					Expect(res.Vested).To(Equal(expVested))
					Expect(res.Unvested).To(Equal(vestingAmtTotal.Sub(expVested...)))
					// All coins from vesting schedule should be locked
					Expect(res.Locked).To(Equal(vestingAmtTotal))
				})

				It("Should be able to delegate locked vested tokens", func() {
					delegateArgs := defaultArgs.WithArgs(
						vestAcc, valAddr.String(), vested[0].Amount.BigInt(),
					)

					logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

					_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					delegation, found := s.app.StakingKeeper.GetDelegation(s.ctx, vestAcc.Bytes(), valAddr)
					Expect(found).To(BeTrue(), "expected delegation to be found")
					Expect(delegation.Shares.BigInt()).To(Equal(vested[0].Amount.BigInt()))
				})

				It("Should be able to delegate locked vested tokens + free tokens (not in vesting schedule)", func() {
					// send some coins to the vesting account
					amt := sdk.NewCoins(sdk.NewCoin(s.bondDenom, math.NewInt(1e18)))
					err := evmosutil.FundAccount(s.ctx, s.app.BankKeeper, clawbackAccount.GetAddress(), amt)
					Expect(err).To(BeNil())

					coinsToDelegate := amt.Add(vested...)

					delegateArgs := defaultArgs.WithArgs(
						vestAcc, valAddr.String(), coinsToDelegate[0].Amount.BigInt(),
					)

					logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

					_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					delegation, found := s.app.StakingKeeper.GetDelegation(s.ctx, vestAcc.Bytes(), valAddr)
					Expect(found).To(BeTrue(), "expected delegation to be found")
					Expect(delegation.Shares.BigInt()).To(Equal(coinsToDelegate[0].Amount.BigInt()))
				})
			})

			Context("Between first and second lockup periods - vested coins are unlocked", func() {
				BeforeEach(func() {
					// Surpass first lockup
					vestDuration := time.Duration(evmosutil.TestVestingSchedule.LockupPeriodLength)
					s.NextBlockAfter(vestDuration * time.Second)

					// Check if some, but not all tokens are vested and unlocked
					vested = clawbackAccount.GetVestedCoins(s.ctx.BlockTime())
					unlocked := clawbackAccount.GetUnlockedCoins(s.ctx.BlockTime())
					unlockedVested = clawbackAccount.GetUnlockedVestedCoins(s.ctx.BlockTime())

					expVested := sdk.NewCoins(sdk.NewCoin(s.bondDenom, evmosutil.TestVestingSchedule.VestedCoinsPerPeriod[0].Amount.Mul(math.NewInt(evmosutil.TestVestingSchedule.LockupMonths))))
					expUnlockedVested := expVested

					Expect(vested).NotTo(Equal(evmosutil.TestVestingSchedule.TotalVestingCoins))
					Expect(vested).To(Equal(expVested))
					// all vested coins are unlocked
					Expect(unlockedVested).To(Equal(vested))
					Expect(unlocked).To(Equal(evmosutil.TestVestingSchedule.UnlockedCoinsPerLockup))
					Expect(unlockedVested).To(Equal(expUnlockedVested))
				})
				It("Should be able to delegate unlocked vested tokens", func() {
					delegateArgs := defaultArgs.WithArgs(
						vestAcc, valAddr.String(), unlockedVested[0].Amount.BigInt(),
					)

					logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

					_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegateArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					delegation, found := s.app.StakingKeeper.GetDelegation(s.ctx, vestAcc.Bytes(), valAddr)
					Expect(found).To(BeTrue(), "expected delegation to be found")
					Expect(delegation.Shares.BigInt()).To(Equal(unlockedVested[0].Amount.BigInt()))
				})
			})
		})
	})

	Context("unbonding", func() {
		// NOTE: there's no additional setup necessary because the test suite is already set up with
		// delegations to the validator

		// defaultUndelegateArgs are the default arguments for the undelegate call
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
		var defaultUndelegateArgs contracts.CallArgs

		BeforeEach(func() {
			defaultUndelegateArgs = defaultCallArgs.WithMethodName("testUndelegate")
		})

		Context("without approval set", func() {
			BeforeEach(func() {
				authz, _ := s.CheckAuthorization(staking.UndelegateAuthz, contractAddr, s.address)
				Expect(authz).To(BeNil(), "expected authorization to be nil before test execution")
			})
			It("should not undelegate", func() {
				undelegateArgs := defaultUndelegateArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(1e18),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, undelegateArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				undelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
				Expect(undelegations).To(HaveLen(0), "expected no undelegations to be found")
			})
		})

		Context("with approval set", func() {
			BeforeEach(func() {
				approveArgs := defaultApproveArgs.WithArgs(
					contractAddr, []string{staking.UndelegateMsg}, big.NewInt(1e18),
				)
				s.SetupApprovalWithContractCalls(approveArgs)
			})

			It("should undelegate when not exceeding the allowance", func() {
				undelegateArgs := defaultUndelegateArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(1e18),
				)

				logCheckArgs := defaultLogCheck.
					WithExpEvents(staking.EventTypeUnbond).
					WithExpPass(true)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, undelegateArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				undelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
				Expect(undelegations).To(HaveLen(1), "expected one undelegation")
				Expect(undelegations[0].ValidatorAddress).To(Equal(valAddr.String()), "expected validator address to be %s", valAddr)
			})

			It("should not undelegate when exceeding the allowance", func() {
				undelegateArgs := defaultUndelegateArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(2e18),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, undelegateArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				undelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
				Expect(undelegations).To(HaveLen(0), "expected no undelegations to be found")
			})

			It("should not undelegate if the delegation does not exist", func() {
				undelegateArgs := defaultUndelegateArgs.WithArgs(
					s.address, nonExistingVal.String(), big.NewInt(1e18),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, undelegateArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				undelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
				Expect(undelegations).To(HaveLen(0), "expected no undelegations to be found")
			})

			It("should not undelegate when called from a different address", func() {
				newAddr, newPriv := testutiltx.NewAccAddressAndKey()
				err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, newAddr, 1e18)
				Expect(err).To(BeNil(), "error while funding account: %v", err)

				s.NextBlock()

				undelegateArgs := defaultUndelegateArgs.
					WithPrivKey(newPriv).
					WithArgs(s.address, valAddr.String(), big.NewInt(1e18))

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, undelegateArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				undelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
				Expect(undelegations).To(HaveLen(0), "expected no undelegations to be found")
			})
		})
	})

	Context("redelegating", func() {
		// NOTE: there's no additional setup necessary because the test suite is already set up with
		// delegations to the validator

		// defaultRedelegateArgs are the default arguments for the redelegate call
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
		var defaultRedelegateArgs contracts.CallArgs

		BeforeEach(func() {
			defaultRedelegateArgs = defaultCallArgs.WithMethodName("testRedelegate")
		})

		Context("without approval set", func() {
			BeforeEach(func() {
				authz, _ := s.CheckAuthorization(staking.UndelegateAuthz, contractAddr, s.address)
				Expect(authz).To(BeNil(), "expected authorization to be nil before test execution")
			})

			It("should not redelegate", func() {
				redelegateArgs := defaultRedelegateArgs.WithArgs(
					s.address, valAddr.String(), valAddr2.String(), big.NewInt(1e18),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegateArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				redelegations := s.app.StakingKeeper.GetAllRedelegations(s.ctx, s.address.Bytes(), valAddr, valAddr2)
				Expect(redelegations).To(HaveLen(0), "expected no redelegations to be found")
			})
		})

		Context("with approval set", func() {
			BeforeEach(func() {
				approveArgs := defaultApproveArgs.WithArgs(
					contractAddr, []string{staking.RedelegateMsg}, big.NewInt(1e18),
				)
				s.SetupApprovalWithContractCalls(approveArgs)
			})

			It("should redelegate when not exceeding the allowance", func() {
				redelegateArgs := defaultRedelegateArgs.WithArgs(
					s.address, valAddr.String(), valAddr2.String(), big.NewInt(1e18),
				)

				logCheckArgs := defaultLogCheck.
					WithExpEvents(staking.EventTypeRedelegate).
					WithExpPass(true)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegateArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				redelegations := s.app.StakingKeeper.GetAllRedelegations(s.ctx, s.address.Bytes(), valAddr, valAddr2)
				Expect(redelegations).To(HaveLen(1), "expected one redelegation to be found")
				bech32Addr := sdk.AccAddress(s.address.Bytes())
				Expect(redelegations[0].DelegatorAddress).To(Equal(bech32Addr.String()), "expected delegator address to be %s", s.address)
				Expect(redelegations[0].ValidatorSrcAddress).To(Equal(valAddr.String()), "expected source validator address to be %s", valAddr)
				Expect(redelegations[0].ValidatorDstAddress).To(Equal(valAddr2.String()), "expected destination validator address to be %s", valAddr2)
			})

			It("should not redelegate when exceeding the allowance", func() {
				redelegateArgs := defaultRedelegateArgs.WithArgs(
					s.address, valAddr.String(), valAddr2.String(), big.NewInt(2e18),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegateArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				redelegations := s.app.StakingKeeper.GetAllRedelegations(s.ctx, s.address.Bytes(), valAddr, valAddr2)
				Expect(redelegations).To(HaveLen(0), "expected no redelegations to be found")
			})

			It("should not redelegate if the delegation does not exist", func() {
				redelegateArgs := defaultRedelegateArgs.WithArgs(
					s.address, nonExistingVal.String(), valAddr2.String(), big.NewInt(1e18),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegateArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				redelegations := s.app.StakingKeeper.GetAllRedelegations(s.ctx, s.address.Bytes(), nonExistingVal, valAddr2)
				Expect(redelegations).To(HaveLen(0), "expected no redelegations to be found")
			})

			It("should not redelegate when calling from a different address", func() {
				newAddr, newPriv := testutiltx.NewAccAddressAndKey()
				err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, newAddr, 1e18)
				Expect(err).To(BeNil(), "error while funding account: %v", err)

				s.NextBlock()

				redelegateArgs := defaultRedelegateArgs.
					WithPrivKey(newPriv).
					WithArgs(s.address, valAddr.String(), valAddr2.String(), big.NewInt(1e18))

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegateArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				redelegations := s.app.StakingKeeper.GetAllRedelegations(s.ctx, s.address.Bytes(), valAddr, valAddr2)
				Expect(redelegations).To(HaveLen(0), "expected no redelegations to be found")
			})

			It("should not redelegate when the validator does not exist", func() {
				redelegateArgs := defaultRedelegateArgs.WithArgs(
					s.address, valAddr.String(), nonExistingVal.String(), big.NewInt(1e18),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegateArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				redelegations := s.app.StakingKeeper.GetAllRedelegations(s.ctx, s.address.Bytes(), valAddr, nonExistingVal)
				Expect(redelegations).To(HaveLen(0), "expected no redelegations to be found")
			})
		})
	})

	Context("canceling unbonding delegations", func() {
		var (
			// defaultCancelUnbondingArgs are the default arguments for the cancelUnbondingDelegation call
			//
			// NOTE: this has to be set up in the BeforeEach block because the private key is only available then
			defaultCancelUnbondingArgs contracts.CallArgs

			// expCreationHeight is the expected creation height of the unbonding delegation
			expCreationHeight = int64(6)
		)

		BeforeEach(func() {
			defaultCancelUnbondingArgs = defaultCallArgs.WithMethodName("testCancelUnbonding")

			// Set up an unbonding delegation
			approvalArgs := defaultApproveArgs.WithArgs(
				contractAddr, []string{staking.UndelegateMsg}, big.NewInt(1e18),
			)
			s.SetupApprovalWithContractCalls(approvalArgs)

			s.NextBlock()

			undelegateArgs := defaultCallArgs.
				WithMethodName("testUndelegate").
				WithArgs(s.address, valAddr.String(), big.NewInt(1e18))

			logCheckArgs := defaultLogCheck.
				WithExpEvents(staking.EventTypeUnbond).
				WithExpPass(true)

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, undelegateArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while setting up an unbonding delegation: %v", err)

			s.NextBlock()

			// Check that the unbonding delegation was created
			unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
			Expect(unbondingDelegations).To(HaveLen(1), "expected one unbonding delegation to be found")
			Expect(unbondingDelegations[0].DelegatorAddress).To(Equal(sdk.AccAddress(s.address.Bytes()).String()), "expected delegator address to be %s", s.address)
			Expect(unbondingDelegations[0].ValidatorAddress).To(Equal(valAddr.String()), "expected validator address to be %s", valAddr)
			Expect(unbondingDelegations[0].Entries).To(HaveLen(1), "expected one unbonding delegation entry to be found")
			Expect(unbondingDelegations[0].Entries[0].CreationHeight).To(Equal(s.ctx.BlockHeight()-1), "expected different creation height")
			Expect(unbondingDelegations[0].Entries[0].Balance).To(Equal(math.NewInt(1e18)), "expected different balance")
		})

		Context("without approval set", func() {
			It("should not cancel unbonding delegations", func() {
				cArgs := defaultCancelUnbondingArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(1e18), big.NewInt(s.ctx.BlockHeight()),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
				Expect(unbondingDelegations).To(HaveLen(1), "expected unbonding delegation not to be canceled")
			})
		})

		Context("with approval set", func() {
			BeforeEach(func() {
				// Set up an unbonding delegation
				approvalArgs := defaultApproveArgs.WithArgs(
					contractAddr, []string{staking.CancelUnbondingDelegationMsg}, big.NewInt(1e18),
				)
				s.SetupApprovalWithContractCalls(approvalArgs)

				s.NextBlock()
			})

			It("should cancel unbonding delegations when not exceeding allowance", func() {
				cArgs := defaultCancelUnbondingArgs.WithGasLimit(1e9).WithArgs(
					s.address, valAddr.String(), big.NewInt(1e18), big.NewInt(expCreationHeight),
				)

				logCheckArgs := passCheck.
					WithExpEvents(staking.EventTypeCancelUnbondingDelegation)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
				Expect(unbondingDelegations).To(HaveLen(0), "expected unbonding delegation to be canceled")
			})

			It("should not cancel unbonding delegations when exceeding allowance", func() {
				approvalArgs := defaultApproveArgs.
					WithArgs(contractAddr, []string{staking.CancelUnbondingDelegationMsg}, big.NewInt(1))
				s.SetupApprovalWithContractCalls(approvalArgs)

				cArgs := defaultCancelUnbondingArgs.WithArgs(
					s.address, valAddr.String(), big.NewInt(1e18), big.NewInt(s.ctx.BlockHeight()),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
				Expect(unbondingDelegations).To(HaveLen(1), "expected unbonding delegation to not be canceled")
			})

			It("should not cancel unbonding any delegations when unbonding delegation does not exist", func() {
				cancelArgs := defaultCancelUnbondingArgs.WithArgs(
					s.address, nonExistingVal.String(), big.NewInt(1e18), big.NewInt(s.ctx.BlockHeight()),
				)

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, cancelArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
				Expect(unbondingDelegations).To(HaveLen(1), "expected unbonding delegation to not be canceled")
			})

			It("should not cancel unbonding delegations when calling from a different address", func() {
				newAddr, newPriv := testutiltx.NewAccAddressAndKey()
				err := evmosutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, newAddr, 1e18)
				Expect(err).To(BeNil(), "error while funding account: %v", err)

				s.NextBlock()

				cancelUnbondArgs := defaultCancelUnbondingArgs.
					WithPrivKey(newPriv).
					WithArgs(s.address, valAddr.String(), big.NewInt(1e18), big.NewInt(s.ctx.BlockHeight()))

				_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, cancelUnbondArgs, execRevertedCheck)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

				unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
				Expect(unbondingDelegations).To(HaveLen(1), "expected unbonding delegation to not be canceled")
			})
		})
	})

	Context("querying allowance", func() {
		// defaultAllowanceArgs are the default arguments for querying the allowance
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
		var defaultAllowanceArgs contracts.CallArgs

		BeforeEach(func() {
			defaultAllowanceArgs = defaultCallArgs.WithMethodName("getAllowance")
		})

		It("without approval set it should show no allowance", func() {
			allowanceArgs := defaultAllowanceArgs.WithArgs(
				contractAddr, staking.CancelUnbondingDelegationMsg,
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, allowanceArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var allowanceInt *big.Int
			err = s.precompile.UnpackIntoInterface(&allowanceInt, "allowance", ethRes.Ret)
			Expect(err).To(BeNil(), "error while unmarshalling the allowance: %v", err)
			Expect(allowanceInt.Int64()).To(Equal(int64(0)), "expected empty allowance")
		})

		It("with approval set it should show the granted allowance", func() {
			// setup approval
			approvalArgs := defaultApproveArgs.WithArgs(
				contractAddr, []string{staking.CancelUnbondingDelegationMsg}, big.NewInt(1e18),
			)

			s.SetupApprovalWithContractCalls(approvalArgs)

			// query allowance
			allowanceArgs := defaultAllowanceArgs.WithArgs(
				contractAddr, staking.CancelUnbondingDelegationMsg,
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, allowanceArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var allowanceInt *big.Int
			err = s.precompile.UnpackIntoInterface(&allowanceInt, "allowance", ethRes.Ret)
			Expect(err).To(BeNil(), "error while unmarshalling the allowance: %v", err)
			Expect(allowanceInt).To(Equal(big.NewInt(1e18)), "expected allowance to be 1e18")
		})
	})

	Context("querying validator", func() {
		// defaultValidatorArgs are the default arguments for querying the validator
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
		var defaultValidatorArgs contracts.CallArgs

		BeforeEach(func() {
			defaultValidatorArgs = defaultCallArgs.WithMethodName("getValidator")
		})

		It("with non-existing address should return an empty validator", func() {
			validatorArgs := defaultValidatorArgs.WithArgs(
				nonExistingAddr,
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, validatorArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)
			Expect(valOut.Validator.OperatorAddress).To(Equal(""), "expected empty validator address")
			Expect(valOut.Validator.Status).To(Equal(uint8(0)), "expected validator status to be 0 (unspecified)")
		})

		It("with existing address should return the validator", func() {
			varHexAddr := common.BytesToAddress(valAddr.Bytes())
			validatorArgs := defaultValidatorArgs.WithArgs(
				varHexAddr,
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, validatorArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)
			Expect(valOut.Validator.OperatorAddress).To(Equal(varHexAddr.String()), "expected validator address to match")
			Expect(valOut.Validator.DelegatorShares).To(Equal(big.NewInt(1e18)), "expected different delegator shares")
		})

		It("with status bonded and pagination", func() {
			validatorArgs := defaultCallArgs.
				WithMethodName("getValidators").
				WithArgs(
					stakingtypes.Bonded.String(),
					query.PageRequest{
						Limit:      1,
						CountTotal: true,
					},
				)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, validatorArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorsOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)
			Expect(valOut.PageResponse.Total).To(Equal(uint64(len(s.validators))))
			Expect(valOut.PageResponse.NextKey).NotTo(BeEmpty())
			Expect(valOut.Validators[0].DelegatorShares).To(Equal(big.NewInt(1e18)), "expected different delegator shares")
		})
	})

	Context("querying validators", func() {
		var defaultValidatorsArgs contracts.CallArgs

		BeforeEach(func() {
			defaultValidatorsArgs = defaultCallArgs.WithMethodName("getValidators")
		})

		It("should return validators (default pagination)", func() {
			validatorsArgs := defaultValidatorsArgs.WithArgs(
				stakingtypes.Bonded.String(),
				query.PageRequest{},
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, validatorsArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorsOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)
			Expect(valOut.PageResponse.Total).To(Equal(uint64(len(s.validators))))
			Expect(valOut.PageResponse.NextKey).To(BeEmpty())
			Expect(valOut.Validators).To(HaveLen(len(s.validators)), "expected all validators to be returned")
			// return order can change, that's why each validator is checked individually
			for _, val := range valOut.Validators {
				s.CheckValidatorOutput(val)
			}
		})

		//nolint:dupl // this is a duplicate of the test for EOA calls to the precompile
		It("should return validators with pagination limit = 1", func() {
			const limit uint64 = 1
			validatorArgs := defaultValidatorsArgs.WithArgs(
				stakingtypes.Bonded.String(),
				query.PageRequest{
					Limit:      limit,
					CountTotal: true,
				},
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, validatorArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorsOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)

			// no pagination, should return default values
			Expect(valOut.PageResponse.NextKey).NotTo(BeEmpty())
			Expect(valOut.PageResponse.Total).To(Equal(uint64(len(s.validators))))

			Expect(valOut.Validators).To(HaveLen(int(limit)), "expected one validator to be returned")

			// return order can change, that's why each validator is checked individually
			for _, val := range valOut.Validators {
				s.CheckValidatorOutput(val)
			}
		})

		It("should revert the execution if the bonding type is not known", func() {
			validatorArgs := defaultValidatorsArgs.WithArgs(
				"15", // invalid bonding type
				query.PageRequest{},
			)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, validatorArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
		})

		It("should return an empty array if there are no validators with the given bonding type", func() {
			validatorArgs := defaultValidatorsArgs.WithArgs(
				stakingtypes.Unbonded.String(),
				query.PageRequest{},
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, validatorArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorsOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)

			Expect(valOut.PageResponse.NextKey).To(BeEmpty())
			Expect(valOut.PageResponse.Total).To(Equal(uint64(0)))
			Expect(valOut.Validators).To(HaveLen(0), "expected no validators to be returned")
		})
	})

	Context("querying delegation", func() {
		// defaultDelegationArgs are the default arguments for querying the delegation
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
		var defaultDelegationArgs contracts.CallArgs

		BeforeEach(func() {
			defaultDelegationArgs = defaultCallArgs.WithMethodName("getDelegation")
		})

		It("which does not exist should return an empty delegation", func() {
			delegationArgs := defaultDelegationArgs.WithArgs(
				nonExistingAddr, valAddr.String(),
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegationArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var delOut staking.DelegationOutput
			err = s.precompile.UnpackIntoInterface(&delOut, staking.DelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the delegation output: %v", err)
			Expect(delOut.Balance.Amount.Int64()).To(Equal(int64(0)), "expected a different delegation balance")
			Expect(delOut.Balance.Denom).To(Equal(utils.BaseDenom), "expected a different delegation balance")
		})

		It("which exists should return the delegation", func() {
			delegationArgs := defaultDelegationArgs.WithArgs(
				s.address, valAddr.String(),
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, delegationArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var delOut staking.DelegationOutput
			err = s.precompile.UnpackIntoInterface(&delOut, staking.DelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the delegation output: %v", err)
			Expect(delOut.Balance).To(Equal(
				cmn.Coin{Denom: utils.BaseDenom, Amount: big.NewInt(1e18)}),
				"expected a different delegation balance",
			)
		})
	})

	Context("querying redelegation", func() {
		// defaultRedelegationArgs are the default arguments for querying the redelegation
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
		var defaultRedelegationArgs contracts.CallArgs

		BeforeEach(func() {
			defaultRedelegationArgs = defaultCallArgs.WithMethodName("getRedelegation")
		})

		It("which does not exist should return an empty redelegation", func() {
			redelegationArgs := defaultRedelegationArgs.WithArgs(
				s.address, valAddr.String(), nonExistingVal.String(),
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegationArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var redOut staking.RedelegationOutput
			err = s.precompile.UnpackIntoInterface(&redOut, staking.RedelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the redelegation output: %v", err)
			Expect(redOut.Redelegation.Entries).To(HaveLen(0), "expected no redelegation entries")
		})

		It("which exists should return the redelegation", func() {
			// set up approval
			approvalArgs := defaultApproveArgs.WithArgs(
				contractAddr, []string{staking.RedelegateMsg}, big.NewInt(1e18),
			)
			s.SetupApprovalWithContractCalls(approvalArgs)

			s.NextBlock()

			// set up redelegation
			redelegateArgs := defaultCallArgs.
				WithMethodName("testRedelegate").
				WithArgs(s.address, valAddr.String(), valAddr2.String(), big.NewInt(1))

			redelegateCheck := passCheck.
				WithExpEvents(staking.EventTypeRedelegate)

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegateArgs, redelegateCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			// check that the redelegation was created
			redelegations := s.app.StakingKeeper.GetAllRedelegations(s.ctx, s.address.Bytes(), valAddr, valAddr2)
			Expect(redelegations).To(HaveLen(1), "expected one redelegation to be found")
			bech32Addr := sdk.AccAddress(s.address.Bytes())
			Expect(redelegations[0].DelegatorAddress).To(Equal(bech32Addr.String()), "expected delegator address to be %s", s.address)
			Expect(redelegations[0].ValidatorSrcAddress).To(Equal(valAddr.String()), "expected source validator address to be %s", valAddr)
			Expect(redelegations[0].ValidatorDstAddress).To(Equal(valAddr2.String()), "expected destination validator address to be %s", valAddr2)

			// query redelegation
			redelegationArgs := defaultRedelegationArgs.WithArgs(
				s.address, valAddr.String(), valAddr2.String(),
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegationArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var redOut staking.RedelegationOutput
			err = s.precompile.UnpackIntoInterface(&redOut, staking.RedelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the redelegation output: %v", err)
			Expect(redOut.Redelegation.Entries).To(HaveLen(1), "expected one redelegation entry to be returned")
		})
	})

	Describe("query redelegations", func() {
		// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
		var defaultRedelegationsArgs contracts.CallArgs

		BeforeEach(func() {
			defaultRedelegationsArgs = defaultCallArgs.WithMethodName("getRedelegations")
		})

		It("which exists should return all the existing redelegations w/pagination", func() {
			// set up approval
			approvalArgs := defaultApproveArgs.WithArgs(
				contractAddr, []string{staking.RedelegateMsg}, big.NewInt(1e18),
			)
			s.SetupApprovalWithContractCalls(approvalArgs)
			s.NextBlock()

			// set up redelegation
			redelegateArgs := defaultCallArgs.
				WithMethodName("testRedelegate").
				WithArgs(s.address, valAddr.String(), valAddr2.String(), big.NewInt(1))

			redelegateCheck := passCheck.
				WithExpEvents(staking.EventTypeRedelegate)
			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegateArgs, redelegateCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			// check that the redelegation was created
			redelegations := s.app.StakingKeeper.GetAllRedelegations(s.ctx, s.address.Bytes(), valAddr, valAddr2)
			Expect(redelegations).To(HaveLen(1), "expected one redelegation to be found")
			bech32Addr := sdk.AccAddress(s.address.Bytes())
			Expect(redelegations[0].DelegatorAddress).To(Equal(bech32Addr.String()), "expected delegator address to be %s", s.address)
			Expect(redelegations[0].ValidatorSrcAddress).To(Equal(valAddr.String()), "expected source validator address to be %s", valAddr)
			Expect(redelegations[0].ValidatorDstAddress).To(Equal(valAddr2.String()), "expected destination validator address to be %s", valAddr2)

			// query redelegations by delegator address
			redelegationArgs := defaultRedelegationsArgs.
				WithArgs(
					s.address, "", "", query.PageRequest{Limit: 1, CountTotal: true},
				)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, redelegationArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var redOut staking.RedelegationsOutput
			err = s.precompile.UnpackIntoInterface(&redOut, staking.RedelegationsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the redelegation output: %v", err)
			Expect(redOut.Response).To(HaveLen(1), "expected one redelegation entry to be returned")
			Expect(redOut.Response[0].Entries).To(HaveLen(1), "expected one redelegation entry to be returned")
			Expect(redOut.PageResponse.Total).To(Equal(uint64(1)))
			Expect(redOut.PageResponse.NextKey).To(BeEmpty())
		})
	})

	Context("querying unbonding delegation", func() {
		// defaultQueryUnbondingArgs are the default arguments for querying the unbonding delegation
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key is not initialized before
		var defaultQueryUnbondingArgs contracts.CallArgs

		BeforeEach(func() {
			defaultQueryUnbondingArgs = defaultCallArgs.WithMethodName("getUnbondingDelegation")

			// Set up an unbonding delegation
			approvalArgs := defaultApproveArgs.WithArgs(
				contractAddr, []string{staking.UndelegateMsg}, big.NewInt(1e18),
			)
			s.SetupApprovalWithContractCalls(approvalArgs)

			s.NextBlock()

			undelegateArgs := defaultCallArgs.
				WithMethodName("testUndelegate").
				WithArgs(s.address, valAddr.String(), big.NewInt(1e18))

			logCheckArgs := passCheck.
				WithExpEvents(staking.EventTypeUnbond)

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, undelegateArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while setting up an unbonding delegation: %v", err)

			// Check that the unbonding delegation was created
			unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
			Expect(unbondingDelegations).To(HaveLen(1), "expected one unbonding delegation to be found")
			Expect(unbondingDelegations[0].DelegatorAddress).To(Equal(sdk.AccAddress(s.address.Bytes()).String()), "expected delegator address to be %s", s.address)
			Expect(unbondingDelegations[0].ValidatorAddress).To(Equal(valAddr.String()), "expected validator address to be %s", valAddr)
			Expect(unbondingDelegations[0].Entries).To(HaveLen(1), "expected one unbonding delegation entry to be found")
			Expect(unbondingDelegations[0].Entries[0].CreationHeight).To(Equal(s.ctx.BlockHeight()), "expected different creation height")
			Expect(unbondingDelegations[0].Entries[0].Balance).To(Equal(math.NewInt(1e18)), "expected different balance")
		})

		It("which does not exist should return an empty unbonding delegation", func() {
			queryUnbondingArgs := defaultQueryUnbondingArgs.WithArgs(
				s.address, valAddr2.String(),
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, queryUnbondingArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var unbondingDelegationOutput staking.UnbondingDelegationOutput
			err = s.precompile.UnpackIntoInterface(&unbondingDelegationOutput, staking.UnbondingDelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the unbonding delegation output: %v", err)
			Expect(unbondingDelegationOutput.UnbondingDelegation.Entries).To(HaveLen(0), "expected one unbonding delegation entry")
		})

		It("which exists should return the unbonding delegation", func() {
			queryUnbondingArgs := defaultQueryUnbondingArgs.WithArgs(
				s.address, valAddr.String(),
			)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, queryUnbondingArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var unbondOut staking.UnbondingDelegationOutput
			err = s.precompile.UnpackIntoInterface(&unbondOut, staking.UnbondingDelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the unbonding delegation output: %v", err)
			Expect(unbondOut.UnbondingDelegation.Entries).To(HaveLen(1), "expected one unbonding delegation entry to be returned")
			Expect(unbondOut.UnbondingDelegation.Entries[0].Balance).To(Equal(big.NewInt(1e18)), "expected different balance")
		})
	})

	Context("testing sequential function calls to the precompile", func() {
		// NOTE: there's no additional setup necessary because the test suite is already set up with
		// delegations to the validator
		It("should revert everything if any operation fails", func() {
			cArgs := defaultCallArgs.
				WithMethodName("testApproveAndThenUndelegate").
				WithGasLimit(1e8).
				WithArgs(contractAddr, big.NewInt(250), big.NewInt(500), valAddr.String())

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

			// There should be no authorizations because everything should have been reverted
			authz, _ := s.CheckAuthorization(staking.UndelegateAuthz, contractAddr, s.address)
			Expect(authz).To(BeNil(), "expected authorization to be nil")

			undelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
			Expect(undelegations).To(HaveLen(0), "expected no unbonding delegations")
		})

		It("should write to state if all operations succeed", func() {
			cArgs := defaultCallArgs.
				WithMethodName("testApproveAndThenUndelegate").
				WithGasLimit(1e8).
				WithArgs(contractAddr, big.NewInt(1000), big.NewInt(500), valAddr.String())

			logCheckArgs := passCheck.
				WithExpEvents(authorization.EventTypeApproval, staking.EventTypeUnbond)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, cArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			authz, _ := s.CheckAuthorization(staking.UndelegateAuthz, contractAddr, s.address)
			Expect(authz).ToNot(BeNil(), "expected authorization not to be nil")

			undelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())
			Expect(undelegations).To(HaveLen(1), "expected one unbonding delegation")
			Expect(undelegations[0].ValidatorAddress).To(Equal(valAddr.String()), "expected different validator address")
		})
	})

	Context("when using special call opcodes", func() {
		testcases := []struct {
			// calltype is the opcode to use
			calltype string
			// expTxPass defines if executing transactions should be possible with the given opcode.
			// Queries should work for all options.
			expTxPass bool
		}{
			{"call", true},
			{"callcode", false},
			{"staticcall", false},
			{"delegatecall", false},
		}

		BeforeEach(func() {
			// approve undelegate message
			approveArgs := defaultApproveArgs.WithArgs(
				contractAddr, []string{staking.UndelegateMsg}, big.NewInt(1e18),
			)
			s.SetupApprovalWithContractCalls(approveArgs)

			s.NextBlock()
		})

		for _, tc := range testcases {
			// NOTE: this is necessary because of Ginkgo behavior -- if not done, the value of tc
			// inside the It block will always be the last entry in the testcases slice
			testcase := tc

			It(fmt.Sprintf("should not execute transactions for calltype %q", testcase.calltype), func() {
				args := defaultCallArgs.
					WithMethodName("testCallUndelegate").
					WithArgs(s.address, valAddr.String(), big.NewInt(1e18), testcase.calltype)

				checkArgs := execRevertedCheck
				if testcase.expTxPass {
					checkArgs = passCheck.WithExpEvents(staking.EventTypeUnbond)
				}

				_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, args, checkArgs)
				if testcase.expTxPass {
					Expect(err).To(BeNil(), "error while calling the smart contract for calltype %s: %v", testcase.calltype, err)
				} else {
					Expect(err).To(HaveOccurred(), "error while calling the smart contract for calltype %s: %v", testcase.calltype, err)
				}
				// check no delegations are unbonding
				undelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())

				if testcase.expTxPass {
					Expect(undelegations).To(HaveLen(1), "expected an unbonding delegation")
					Expect(undelegations[0].ValidatorAddress).To(Equal(valAddr.String()), "expected different validator address")
					Expect(undelegations[0].DelegatorAddress).To(Equal(sdk.AccAddress(s.address.Bytes()).String()), "expected different delegator address")
				} else {
					Expect(undelegations).To(HaveLen(0), "expected no unbonding delegations for calltype %s", testcase.calltype)
				}
			})

			It(fmt.Sprintf("should execute queries for calltype %q", testcase.calltype), func() {
				args := defaultCallArgs.
					WithMethodName("testCallDelegation").
					WithArgs(s.address, valAddr.String(), testcase.calltype)

				_, ethRes, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, args, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var delOut staking.DelegationOutput
				err = s.precompile.UnpackIntoInterface(&delOut, staking.DelegationMethod, ethRes.Ret)
				Expect(err).To(BeNil(), "error while unpacking the delegation output: %v", err)
				Expect(delOut.Shares).To(Equal(math.LegacyNewDec(1).BigInt()), "expected different delegation shares")
				Expect(delOut.Balance.Amount).To(Equal(big.NewInt(1e18)), "expected different delegation balance")
				if testcase.calltype != "callcode" { // having some trouble with returning the denom from inline assembly but that's a very special edge case which might never be used
					Expect(delOut.Balance.Denom).To(Equal(s.bondDenom), "expected different denomination")
				}
			})
		}
	})

	// NOTE: These tests were added to replicate a problematic behavior, that occurred when a contract
	// adjusted the state in multiple subsequent function calls, which adjusted the EVM state as well as
	// things from the Cosmos SDK state (e.g. a bank balance).
	// The result was, that changes made to the Cosmos SDK state have been overwritten during the next function
	// call, because the EVM state was not updated in between.
	//
	// This behavior was fixed by updating the EVM state after each function call.
	Context("when triggering multiple state changes in one function", func() {
		// delegationAmount is the amount to be delegated
		delegationAmount := big.NewInt(1e18)

		BeforeEach(func() {
			// Set up funding for the contract address.
			// NOTE: we are first asserting that no balance exists and then check successful
			// funding afterwards.
			balanceBefore := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
			Expect(balanceBefore.Amount.Int64()).To(BeZero(), "expected contract balance to be 0 before funding")

			err = s.app.BankKeeper.SendCoins(
				s.ctx, s.address.Bytes(), contractAddr.Bytes(),
				sdk.Coins{sdk.Coin{Denom: s.bondDenom, Amount: math.NewIntFromBigInt(delegationAmount)}},
			)
			Expect(err).To(BeNil(), "error while sending coins: %v", err)

			s.NextBlock()

			balanceAfterFunding := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
			Expect(balanceAfterFunding.Amount.BigInt()).To(Equal(delegationAmount), "expected different contract balance after funding")

			// Check no delegation exists from the contract to the validator
			_, found := s.app.StakingKeeper.GetDelegation(s.ctx, contractAddr.Bytes(), valAddr)
			Expect(found).To(BeFalse(), "expected delegation not to be found before testing")
		})

		It("delegating and increasing counter should change the bank balance accordingly", func() {
			delegationArgs := defaultCallArgs.
				WithGasLimit(1e9).
				WithMethodName("testDelegateIncrementCounter").
				WithArgs(valAddr.String(), delegationAmount)

			approvalAndDelegationCheck := passCheck.WithExpEvents(
				authorization.EventTypeApproval, staking.EventTypeDelegate,
			)

			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegationArgs, approvalAndDelegationCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			del, found := s.app.StakingKeeper.GetDelegation(s.ctx, contractAddr.Bytes(), valAddr)

			Expect(found).To(BeTrue(), "expected delegation to be found")
			Expect(del.GetShares().BigInt()).To(Equal(delegationAmount), "expected different delegation shares")

			postBalance := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
			Expect(postBalance.Amount.Int64()).To(BeZero(), "expected balance to be 0 after contract call")
		})
	})

	Context("when updating the stateDB prior to calling the precompile", func() {
		It("should utilize the same contract balance to delegate", func() {
			delegationArgs := defaultCallArgs.
				WithGasLimit(1e9).
				WithMethodName("approveDepositAndDelegate").
				WithArgs(valAddr.String()).
				WithAmount(big.NewInt(2e18))

			approvalAndDelegationCheck := passCheck.WithExpEvents(
				authorization.EventTypeApproval, staking.EventTypeDelegate,
			)
			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegationArgs, approvalAndDelegationCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			balance := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
			Expect(balance.Amount.Int64()).To(BeZero(), "expected different contract balance after funding")
			delegation := s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, contractAddr.Bytes())
			Expect(delegation).To(HaveLen(1), "expected one delegation")
			Expect(delegation[0].GetShares().BigInt()).To(Equal(big.NewInt(2e18)), "expected different delegation shares")
		})
		//nolint:dupl
		It("should revert the contract balance to the original value when the precompile fails", func() {
			delegationArgs := defaultCallArgs.
				WithGasLimit(1e9).
				WithMethodName("approveDepositAndDelegateExceedingAllowance").
				WithArgs(valAddr.String()).
				WithAmount(big.NewInt(2e18))

			approvalAndDelegationCheck := defaultLogCheck.WithErrContains(vm.ErrExecutionReverted.Error())
			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegationArgs, approvalAndDelegationCheck)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

			balance := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
			Expect(balance.Amount.Int64()).To(BeZero(), "expected different contract balance after funding")
			auth, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, contractAddr.Bytes(), s.address.Bytes(), staking.DelegateMsg)
			Expect(auth).To(BeNil(), "expected no authorization")
			delegation := s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, contractAddr.Bytes())
			Expect(delegation).To(HaveLen(0), "expected no delegations")
		})

		//nolint:dupl
		It("should revert the contract balance to the original value when the custom logic after the precompile fails ", func() {
			delegationArgs := defaultCallArgs.
				WithGasLimit(1e9).
				WithMethodName("approveDepositDelegateAndFailCustomLogic").
				WithArgs(valAddr.String()).
				WithAmount(big.NewInt(2e18))

			approvalAndDelegationCheck := defaultLogCheck.WithErrContains(vm.ErrExecutionReverted.Error())
			_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, delegationArgs, approvalAndDelegationCheck)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

			balance := s.app.BankKeeper.GetBalance(s.ctx, contractAddr.Bytes(), s.bondDenom)
			Expect(balance.Amount.Int64()).To(BeZero(), "expected different contract balance after funding")
			auth, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, contractAddr.Bytes(), s.address.Bytes(), staking.DelegateMsg)
			Expect(auth).To(BeNil(), "expected no authorization")
			delegation := s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, contractAddr.Bytes())
			Expect(delegation).To(HaveLen(0), "expected no delegations")
		})
	})
})

// These tests are used to check that when batching multiple state changing transactions
// in one block, both states (Cosmos and EVM) are updated or reverted correctly.
//
// For this purpose, we are deploying an ERC20 contract and updating StakingCaller.sol
// to include a method where an ERC20 balance is sent between accounts as well as
// an interaction with the staking precompile is made.
//
// There are ERC20 tokens minted to the address of the deployed StakingCaller contract,
// which will transfer these to the message sender when successfully executed.
// Using the staking EVM extension, there is an approval made before the ERC20 transfer
// as well as a delegation after the ERC20 transfer.
var _ = Describe("Batching cosmos and eth interactions", func() {
	const (
		erc20Name     = "Test"
		erc20Token    = "TTT"
		erc20Decimals = uint8(18)
	)

	var (
		// contractAddr is the address of the deployed StakingCaller contract
		contractAddr common.Address
		// stakingCallerContract is the contract instance calling into the staking precompile
		stakingCallerContract evmtypes.CompiledContract
		// erc20ContractAddr is the address of the deployed ERC20 contract
		erc20ContractAddr common.Address
		// erc20Contract is the compiled ERC20 contract
		erc20Contract = compiledcontracts.ERC20MinterBurnerDecimalsContract

		// err is a standard error
		err error
		// execRevertedCheck is a standard log check for a reverted transaction
		execRevertedCheck = defaultLogCheck.WithErrContains(vm.ErrExecutionReverted.Error())

		// mintAmount is the amount of ERC20 tokens minted to the StakingCaller contract
		mintAmount = big.NewInt(1e18)
		// transferredAmount is the amount of ERC20 tokens to transfer during the tests
		transferredAmount = big.NewInt(1234e9)
	)

	BeforeEach(func() {
		s.SetupTest()
		s.NextBlock()

		stakingCallerContract, err = testdata.LoadStakingCallerContract()
		Expect(err).To(BeNil(), "error while loading the StakingCaller contract")

		// Deploy StakingCaller contract
		contractAddr, err = evmosutil.DeployContract(s.ctx, s.app, s.privKey, s.queryClientEVM, stakingCallerContract)
		Expect(err).To(BeNil(), "error while deploying the StakingCaller contract")

		// Deploy ERC20 contract
		erc20ContractAddr, err = evmosutil.DeployContract(s.ctx, s.app, s.privKey, s.queryClientEVM, erc20Contract,
			erc20Name, erc20Token, erc20Decimals,
		)
		Expect(err).To(BeNil(), "error while deploying the ERC20 contract")

		// Mint tokens to the StakingCaller contract
		mintArgs := contracts.CallArgs{
			ContractAddr: erc20ContractAddr,
			ContractABI:  erc20Contract.ABI,
			MethodName:   "mint",
			PrivKey:      s.privKey,
			Args:         []interface{}{contractAddr, mintAmount},
		}

		mintCheck := testutil.LogCheckArgs{
			ABIEvents: erc20Contract.ABI.Events,
			ExpEvents: []string{"Transfer"}, // minting produces a Transfer event
			ExpPass:   true,
		}

		_, _, err = contracts.CallContractAndCheckLogs(s.ctx, s.app, mintArgs, mintCheck)
		Expect(err).To(BeNil(), "error while minting tokens to the StakingCaller contract")

		// Check that the StakingCaller contract has the correct balance
		erc20Balance := s.app.Erc20Keeper.BalanceOf(s.ctx, erc20Contract.ABI, erc20ContractAddr, contractAddr)
		Expect(erc20Balance).To(Equal(mintAmount), "expected different ERC20 balance for the StakingCaller contract")

		// populate default call args
		defaultCallArgs = contracts.CallArgs{
			ContractABI:  stakingCallerContract.ABI,
			ContractAddr: contractAddr,
			MethodName:   "callERC20AndDelegate",
			PrivKey:      s.privKey,
		}

		// populate default log check args
		defaultLogCheck = testutil.LogCheckArgs{
			ABIEvents: s.precompile.Events,
		}
		execRevertedCheck = defaultLogCheck.WithErrContains(vm.ErrExecutionReverted.Error())
		passCheck = defaultLogCheck.WithExpPass(true)
	})

	Describe("when batching multiple transactions", func() {
		// validator is the validator address used for testing
		var validator sdk.ValAddress

		BeforeEach(func() {
			delegations := s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, s.address.Bytes())
			Expect(delegations).ToNot(HaveLen(0), "expected address to have delegations")

			validator = delegations[0].GetValidatorAddr()

			_ = erc20ContractAddr
		})

		It("should revert both states if a staking transaction fails", func() {
			delegationPre, found := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), validator)
			Expect(found).To(BeTrue(),
				"expected delegation from %s to validator %s to be found",
				sdk.AccAddress(s.address.Bytes()).String(), validator.String(),
			)

			sharesPre := delegationPre.GetShares()

			// NOTE: passing an invalid validator address here should fail AFTER the erc20 transfer was made in the smart contract.
			// Therefore this can be used to check that both EVM and Cosmos states are reverted correctly.
			failArgs := defaultCallArgs.
				WithArgs(erc20ContractAddr, "invalid validator", transferredAmount)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, failArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "expected error while calling the smart contract")
			Expect(err.Error()).To(ContainSubstring("execution reverted"), "expected different error message")

			delegationPost, found := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), validator)
			Expect(found).To(BeTrue(),
				"expected delegation from %s to validator %s to be found after calling the smart contract",
				sdk.AccAddress(s.address.Bytes()).String(), validator.String(),
			)

			auths, err := s.app.AuthzKeeper.GetAuthorizations(s.ctx, contractAddr.Bytes(), s.address.Bytes())
			Expect(err).To(BeNil(), "error while getting authorizations: %v", err)
			sharesPost := delegationPost.GetShares()
			erc20BalancePost := s.app.Erc20Keeper.BalanceOf(s.ctx, erc20Contract.ABI, erc20ContractAddr, s.address)

			Expect(auths).To(BeEmpty(), "expected no authorizations when reverting state")
			Expect(sharesPost).To(Equal(sharesPre), "expected shares to be equal when reverting state")
			Expect(erc20BalancePost.Int64()).To(BeZero(), "expected erc20 balance of target address to be zero when reverting state")
		})

		It("should revert both states if an ERC20 transaction fails", func() {
			delegationPre, found := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), validator)
			Expect(found).To(BeTrue(),
				"expected delegation from %s to validator %s to be found",
				sdk.AccAddress(s.address.Bytes()).String(), validator.String(),
			)

			sharesPre := delegationPre.GetShares()

			// NOTE: trying to transfer more than the balance of the contract should fail AFTER the approval
			// for delegating was made in the smart contract.
			// Therefore this can be used to check that both EVM and Cosmos states are reverted correctly.
			moreThanMintedAmount := new(big.Int).Add(mintAmount, big.NewInt(1))
			failArgs := defaultCallArgs.
				WithArgs(erc20ContractAddr, s.validators[0].OperatorAddress, moreThanMintedAmount)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, failArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "expected error while calling the smart contract")
			Expect(err.Error()).To(ContainSubstring("execution reverted"), "expected different error message")

			delegationPost, found := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), validator)
			Expect(found).To(BeTrue(),
				"expected delegation from %s to validator %s to be found after calling the smart contract",
				sdk.AccAddress(s.address.Bytes()).String(), validator.String(),
			)

			auths, err := s.app.AuthzKeeper.GetAuthorizations(s.ctx, contractAddr.Bytes(), s.address.Bytes())
			Expect(err).To(BeNil(), "error while getting authorizations: %v", err)
			sharesPost := delegationPost.GetShares()
			erc20BalancePost := s.app.Erc20Keeper.BalanceOf(s.ctx, erc20Contract.ABI, erc20ContractAddr, s.address)

			Expect(auths).To(BeEmpty(), "expected no authorizations when reverting state")
			Expect(sharesPost).To(Equal(sharesPre), "expected shares to be equal when reverting state")
			Expect(erc20BalancePost.Int64()).To(BeZero(), "expected erc20 balance of target address to be zero when reverting state")
		})

		It("should persist changes in both the cosmos and eth states", func() {
			delegationPre, found := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), validator)
			Expect(found).To(BeTrue(),
				"expected delegation from %s to validator %s to be found",
				sdk.AccAddress(s.address.Bytes()).String(), validator.String(),
			)

			sharesPre := delegationPre.GetShares()

			// NOTE: trying to transfer more than the balance of the contract should fail AFTER the approval
			// for delegating was made in the smart contract.
			// Therefore this can be used to check that both EVM and Cosmos states are reverted correctly.
			successArgs := defaultCallArgs.
				WithArgs(erc20ContractAddr, s.validators[0].OperatorAddress, transferredAmount)

			// Build combined map of ABI events to check for both ERC20 events as well as precompile events
			//
			// NOTE: only add the transfer event - when adding all contract events to the combined map,
			// the ERC20 Approval event will overwrite the precompile Approval event, which will cause
			// the check to fail because of unexpected events in the logs.
			combinedABIEvents := s.precompile.Events
			combinedABIEvents["Transfer"] = erc20Contract.ABI.Events["Transfer"]

			successCheck := passCheck.
				WithABIEvents(combinedABIEvents).
				WithExpEvents(
					authorization.EventTypeApproval, "Transfer", staking.EventTypeDelegate,
				)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, successArgs, successCheck)
			Expect(err).ToNot(HaveOccurred(), "error while calling the smart contract")

			delegationPost, found := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), validator)
			Expect(found).To(BeTrue(),
				"expected delegation from %s to validator %s to be found after calling the smart contract",
				sdk.AccAddress(s.address.Bytes()).String(), validator.String(),
			)

			auths, err := s.app.AuthzKeeper.GetAuthorizations(s.ctx, contractAddr.Bytes(), s.address.Bytes())
			Expect(err).To(BeNil(), "error while getting authorizations: %v", err)
			sharesPost := delegationPost.GetShares()
			erc20BalancePost := s.app.Erc20Keeper.BalanceOf(s.ctx, erc20Contract.ABI, erc20ContractAddr, s.address)

			Expect(sharesPost.GT(sharesPre)).To(BeTrue(), "expected shares to be more than before")
			Expect(erc20BalancePost).To(Equal(transferredAmount), "expected different erc20 balance of target address")
			// NOTE: there should be no authorizations because the full approved amount is delegated
			Expect(auths).To(HaveLen(0), "expected no authorization to be found")
		})
	})
})
