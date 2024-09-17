// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package staking_test

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	compiledcontracts "github.com/evmos/evmos/v20/contracts"
	"github.com/evmos/evmos/v20/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v20/precompiles/authorization"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/precompiles/distribution"
	"github.com/evmos/evmos/v20/precompiles/staking"
	"github.com/evmos/evmos/v20/precompiles/staking/testdata"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	"github.com/evmos/evmos/v20/precompiles/testutil/contracts"
	evmosutil "github.com/evmos/evmos/v20/testutil"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	testutils "github.com/evmos/evmos/v20/testutil/integration/evmos/utils"
	testutiltx "github.com/evmos/evmos/v20/testutil/tx"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	vestingtypes "github.com/evmos/evmos/v20/x/vesting/types"
)

func TestPrecompileIntegrationTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Staking Precompile Integration Tests")
}

// General variables used for integration tests
var (
	// valAddr and valAddr2 are the two validator addresses used for testing
	valAddr, valAddr2 sdk.ValAddress

	// callArgs and approveCallArgs are the default arguments for calling the smart contract and to
	// call the approve method specifically.
	//
	// NOTE: this has to be populated in a BeforeEach block because the contractAddr would otherwise be a nil address.
	callArgs, approveCallArgs factory.CallArgs
	// txArgs are the EVM transaction arguments to use in the transactions
	txArgs evmtypes.EvmTxArgs
	// defaultLogCheck instantiates a log check arguments struct with the precompile ABI events populated.
	defaultLogCheck testutil.LogCheckArgs
	// passCheck defines the arguments to check if the precompile returns no error
	passCheck testutil.LogCheckArgs
	// outOfGasCheck defines the arguments to check if the precompile returns out of gas error
	outOfGasCheck testutil.LogCheckArgs
)

var _ = Describe("Calling staking precompile directly", func() {
	var (
		// s is the precompile test suite to use for the tests
		s *PrecompileTestSuite
		// oneE18Coin is a sdk.Coin with an amount of 1e18 in the test suite's bonding denomination
		oneE18Coin sdk.Coin
		// twoE18Coin is a sdk.Coin with an amount of 2e18 in the test suite's bonding denomination
		twoE18Coin sdk.Coin
	)

	BeforeEach(func() {
		var err error
		s = new(PrecompileTestSuite)
		s.SetupTest()

		valAddr, err = sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
		Expect(err).To(BeNil())
		valAddr2, err = sdk.ValAddressFromBech32(s.network.GetValidators()[1].GetOperator())
		Expect(err).To(BeNil())

		approveCallArgs = factory.CallArgs{
			ContractABI: s.precompile.ABI,
			MethodName:  authorization.ApproveMethod,
		}

		callArgs = factory.CallArgs{
			ContractABI: s.precompile.ABI,
		}

		precompileAddr := s.precompile.Address()
		txArgs = evmtypes.EvmTxArgs{
			To: &precompileAddr,
		}

		defaultLogCheck = testutil.LogCheckArgs{ABIEvents: s.precompile.ABI.Events}
		passCheck = defaultLogCheck.WithExpPass(true)
		outOfGasCheck = defaultLogCheck.WithErrContains(vm.ErrOutOfGas.Error())

		oneE18Coin = sdk.NewCoin(s.bondDenom, math.NewInt(1e18))
		twoE18Coin = sdk.NewCoin(s.bondDenom, math.NewInt(2e18))
	})

	Describe("when the precompile is not enabled in the EVM params", func() {
		It("should succeed but not perform delegation", func() {
			delegator := s.keyring.GetKey(0)
			// disable the precompile
			res, err := s.grpcHandler.GetEvmParams()
			Expect(err).To(BeNil())

			var activePrecompiles []string
			for _, precompile := range res.Params.ActiveStaticPrecompiles {
				if precompile != s.precompile.Address().String() {
					activePrecompiles = append(activePrecompiles, precompile)
				}
			}
			res.Params.ActiveStaticPrecompiles = activePrecompiles

			err = testutils.UpdateEvmParams(testutils.UpdateParamsInput{
				Tf:      s.factory,
				Network: s.network,
				Pk:      delegator.Priv,
				Params:  res.Params,
			})
			Expect(err).To(BeNil(), "error while setting params")

			// get the delegation that is available prior to the test
			qRes, err := s.grpcHandler.GetDelegation(delegator.AccAddr.String(), valAddr.String())
			Expect(err).To(BeNil())
			prevDelegation := qRes.DelegationResponse.Balance
			// try to call the precompile
			callArgs.MethodName = staking.DelegateMethod
			callArgs.Args = []interface{}{delegator.Addr, valAddr.String(), big.NewInt(2e18)}

			// Contract should not be called but the transaction should be successful
			// This is the expected behavior in Ethereum where there is a contract call
			// to a non existing contract
			expectedCheck := defaultLogCheck.
				WithExpEvents([]string{}...).
				WithExpPass(true)

			_, _, err = s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs,
				callArgs,
				expectedCheck,
			)
			Expect(err).To(BeNil(), "expected error while calling the precompile")
			qRes, err = s.grpcHandler.GetDelegation(delegator.AccAddr.String(), valAddr.String())
			Expect(err).To(BeNil())
			postDelegation := qRes.DelegationResponse.Balance
			Expect(postDelegation).To(Equal(prevDelegation), "expected delegation to not change")
		})
	})

	Describe("Revert transaction", func() {
		It("should run out of gas if the gas limit is too low", func() {
			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			approveCallArgs.Args = []interface{}{
				grantee.Addr,
				abi.MaxUint256,
				[]string{staking.DelegateMsg},
			}
			txArgs.GasLimit = 30000

			_, _, err := s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs,
				approveCallArgs,
				outOfGasCheck,
			)
			Expect(err).To(BeNil(), "error while calling precompile")
		})
	})

	Describe("Execute approve transaction", func() {
		// TODO: enable once we check that the spender is not the origin
		// It("should return error if the origin is the spender", func() {
		//	args := defaultApproveArgs.WithArgs(
		//		granter.Addr,
		//		abi.MaxUint256,
		//		[]string{staking.DelegateMsg},
		//	)
		//
		//	differentOriginCheck := defaultLogCheck.WithErrContains(cmn.ErrDifferentOrigin, granter.Addr, addr)
		//
		//	_, _, err := s.factory.CallContractAndCheckLogs(
		//	Expect(err).To(BeNil(), "error while calling precompile")
		// })

		It("should return error if the staking method is not supported on the precompile", func() {
			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			approveCallArgs.Args = []interface{}{
				grantee.Addr,
				abi.MaxUint256,
				[]string{distribution.DelegationRewardsMethod},
			}

			logCheckArgs := defaultLogCheck.WithErrContains(
				cmn.ErrInvalidMsgType, "staking", distribution.DelegationRewardsMethod,
			)

			_, _, err := s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs,
				approveCallArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the contract and checking logs")
		})

		It("should approve the delegate method with the max uint256 value", func() {
			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			s.SetupApproval(
				granter.Priv, grantee.Addr, abi.MaxUint256, []string{staking.DelegateMsg},
			)

			s.ExpectAuthorization(staking.DelegateAuthz, grantee.Addr, granter.Addr, nil)
		})

		It("should approve the undelegate method with 1 evmos", func() {
			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			s.SetupApproval(
				granter.Priv, grantee.Addr, big.NewInt(1e18), []string{staking.UndelegateMsg},
			)

			s.ExpectAuthorization(staking.UndelegateAuthz, grantee.Addr, granter.Addr, &oneE18Coin)
		})

		It("should approve the redelegate method with 2 evmos", func() {
			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			s.SetupApproval(
				granter.Priv, grantee.Addr, big.NewInt(2e18), []string{staking.RedelegateMsg},
			)

			s.ExpectAuthorization(staking.RedelegateAuthz, grantee.Addr, granter.Addr, &twoE18Coin)
		})

		It("should approve the cancel unbonding delegation method with 1 evmos", func() {
			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			s.SetupApproval(
				granter.Priv, grantee.Addr, big.NewInt(1e18), []string{staking.CancelUnbondingDelegationMsg},
			)

			s.ExpectAuthorization(staking.CancelUnbondingDelegationAuthz, grantee.Addr, granter.Addr, &oneE18Coin)
		})
	})

	Describe("Execute increase allowance transaction", func() {
		BeforeEach(func() {
			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			s.SetupApproval(
				granter.Priv, grantee.Addr, big.NewInt(1e18), []string{staking.DelegateMsg},
			)
			callArgs.MethodName = authorization.IncreaseAllowanceMethod
		})

		// TODO: enable once we check that the spender is not the origin
		// It("should return error if the origin is the spender", func() {
		//	increaseArgs := defaultCallArgs.
		//		WithMethodName(authorization.IncreaseAllowanceMethod).
		//		WithArgs(
		//			granter.Addr, big.NewInt(1e18), []string{staking.DelegateMsg},
		//		)
		//
		//	_, _, err := s.factory.CallContractAndCheckLogs(
		//	Expect(err).To(BeNil(), "error while calling the contract and checking logs")
		// })

		It("Should increase the allowance of the delegate method with 1 evmos", func() {
			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			callArgs.Args = []interface{}{
				grantee.Addr, big.NewInt(1e18), []string{staking.DelegateMsg},
			}

			logCheckArgs := passCheck.WithExpEvents(authorization.EventTypeAllowanceChange)

			_, _, err := s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs,
				callArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the contract and checking logs")
			Expect(s.network.NextBlock()).To(BeNil())

			s.ExpectAuthorization(staking.DelegateAuthz, grantee.Addr, granter.Addr, &twoE18Coin)
		})

		It("should return error if the allowance to increase does not exist", func() {
			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			callArgs.Args = []interface{}{
				grantee.Addr, big.NewInt(1e18), []string{staking.UndelegateMsg},
			}

			logCheckArgs := defaultLogCheck.WithErrContains(
				"does not exist",
			)

			_, _, err := s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs,
				callArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the contract and checking logs")
			Expect(s.network.NextBlock()).To(BeNil())

			authz, _, err := CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, staking.UndelegateAuthz, grantee.Addr, granter.Addr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("invalid authorization type. Expected: %d, got: %d", staking.UndelegateAuthz, staking.DelegateAuthz)))
			Expect(authz).To(BeNil(), "expected authorization to not be set")
		})
	})

	Describe("Execute decrease allowance transaction", func() {
		BeforeEach(func() {
			granteeAddr := s.precompile.Address()
			granter := s.keyring.GetKey(0)

			s.SetupApproval(
				granter.Priv, granteeAddr, big.NewInt(2e18), []string{staking.DelegateMsg},
			)

			callArgs.MethodName = authorization.DecreaseAllowanceMethod
		})

		// TODO: enable once we check that the spender is not the origin
		// It("should return error if the origin is the spender", func() {
		//	addr, _ := testutiltx.NewAddrKey()
		//	decreaseArgs := defaultDecreaseArgs.WithArgs(
		//		grantee.Addr, big.NewInt(1e18), []string{staking.DelegateMsg},
		//	)
		//
		//	logCheckArgs := defaultLogCheck.WithErrContains(
		//		cmn.ErrDifferentOrigin, granter.Addr, addr,
		//	)
		//
		//	_, _, err := s.factory.CallContractAndCheckLogs(
		//	Expect(err).To(BeNil(), "error while calling the contract and checking logs")
		// })

		It("Should decrease the allowance of the delegate method with 1 evmos", func() {
			granteeAddr := s.precompile.Address()
			granter := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				granteeAddr, big.NewInt(1e18), []string{staking.DelegateMsg},
			}

			logCheckArgs := passCheck.WithExpEvents(authorization.EventTypeAllowanceChange)

			_, _, err := s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs, callArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the contract and checking logs")
			Expect(s.network.NextBlock()).To(BeNil())

			s.ExpectAuthorization(staking.DelegateAuthz, granteeAddr, granter.Addr, &oneE18Coin)
		})

		It("should return error if the allowance to decrease does not exist", func() {
			granteeAddr := s.precompile.Address()
			granter := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				granteeAddr, big.NewInt(1e18), []string{staking.UndelegateMsg},
			}

			logCheckArgs := defaultLogCheck.WithErrContains(
				"does not exist",
			)

			_, _, err := s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs, callArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the contract and checking logs")
			Expect(s.network.NextBlock()).To(BeNil())

			authz, _, err := CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, staking.UndelegateAuthz, granteeAddr, granter.Addr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("invalid authorization type. Expected: %d, got: %d", staking.UndelegateAuthz, staking.DelegateAuthz)))
			Expect(authz).To(BeNil(), "expected authorization to not be set")
		})
	})

	Describe("to revoke an approval", func() {
		// granteeAddr is the address of the grantee used in the revocation tests.
		granteeAddr := testutiltx.GenerateAddress()

		BeforeEach(func() {
			callArgs.MethodName = authorization.RevokeMethod
		})

		It("should revoke the approval when executing as the granter", func() {
			granter := s.keyring.GetKey(0)
			typeURLs := []string{staking.DelegateMsg}

			s.SetupApproval(
				granter.Priv, granteeAddr, abi.MaxUint256, typeURLs,
			)
			s.ExpectAuthorization(staking.DelegateAuthz, granteeAddr, granter.Addr, nil)

			callArgs.Args = []interface{}{
				granteeAddr, typeURLs,
			}

			revocationCheck := passCheck.WithExpEvents(authorization.EventTypeRevocation)

			_, _, err := s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs, callArgs,
				revocationCheck)
			Expect(err).To(BeNil(), "error while calling the contract and checking logs")
			Expect(s.network.NextBlock()).To(BeNil())

			// check that the authorization is revoked
			authz, _, err := CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, staking.DelegateAuthz, granteeAddr, granter.Addr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("no authorizations found for grantee %s and granter %s", granteeAddr.Hex(), granter.Addr.Hex())))
			Expect(authz).To(BeNil(), "expected authorization to be revoked")
		})

		It("should not revoke the approval when trying to revoke for a different message type", func() {
			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			typeURLs := []string{staking.DelegateMsg}

			s.SetupApproval(
				granter.Priv, grantee.Addr, abi.MaxUint256, typeURLs,
			)
			s.ExpectAuthorization(staking.DelegateAuthz, grantee.Addr, granter.Addr, nil)

			callArgs.Args = []interface{}{
				grantee.Addr, []string{staking.UndelegateMsg},
			}

			notFoundCheck := defaultLogCheck.
				WithErrContains("failed to delete grant")

			_, _, err := s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs, callArgs,
				notFoundCheck,
			)
			Expect(err).To(BeNil(), "error while calling the contract and checking logs")
			Expect(s.network.NextBlock()).To(BeNil())

			// the authorization should still be there.
			s.ExpectAuthorization(staking.DelegateAuthz, grantee.Addr, granter.Addr, nil)
		})

		It("should return error if the approval does not exist", func() {
			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			callArgs.Args = []interface{}{
				grantee.Addr, []string{staking.DelegateMsg},
			}

			notFoundCheck := defaultLogCheck.
				WithErrContains("failed to delete grant")

			_, _, err := s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs, callArgs,
				notFoundCheck,
			)
			Expect(err).To(BeNil(), "error while calling the contract and checking logs")
		})

		It("should not revoke the approval if sent by someone else than the granter", func() {
			typeURLs := []string{staking.DelegateMsg}

			// set up an approval with a different key than the one used to sign the transaction.
			granter := s.keyring.GetKey(0)
			differentSender := s.keyring.GetKey(1)

			s.SetupApproval(
				granter.Priv, granteeAddr, abi.MaxUint256, typeURLs,
			)
			s.ExpectAuthorization(staking.DelegateAuthz, granteeAddr, granter.Addr, nil)

			callArgs.Args = []interface{}{
				granteeAddr, typeURLs,
			}

			notFoundCheck := defaultLogCheck.
				WithErrContains("failed to delete grant")

			_, _, err := s.factory.CallContractAndCheckLogs(
				differentSender.Priv,
				txArgs, callArgs,
				notFoundCheck,
			)
			Expect(err).To(BeNil(), "error while calling the contract and checking logs")
			Expect(s.network.NextBlock()).To(BeNil())

			// the authorization should still be set
			s.ExpectAuthorization(staking.DelegateAuthz, granteeAddr, granter.Addr, nil)
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
		)

		BeforeEach(func() {
			// populate the default createValidator args
			callArgs.MethodName = staking.CreateValidatorMethod
		})

		Context("when validator address is the origin", func() {
			It("should succeed", func() {
				callArgs.Args = []interface{}{
					defaultDescription, defaultCommission, defaultMinSelfDelegation, s.keyring.GetAddr(0), defaultPubkeyBase64Str, defaultValue,
				}

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeCreateValidator)

				_, _, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the contract and checking logs")
				Expect(s.network.NextBlock()).To(BeNil())

				valOperAddr := sdk.ValAddress(s.keyring.GetAccAddr(0)).String()
				qc := s.network.GetStakingClient()
				res, err := qc.Validator(s.network.GetContext(), &stakingtypes.QueryValidatorRequest{ValidatorAddr: valOperAddr})
				Expect(err).To(BeNil())
				Expect(res).NotTo(BeNil())
				Expect(res.Validator.OperatorAddress).To(Equal(valOperAddr))
			})
		})

		Context("when validator address is not the origin", func() {
			It("should fail", func() {
				differentAddr := testutiltx.GenerateAddress()

				callArgs.Args = []interface{}{
					defaultDescription, defaultCommission, defaultMinSelfDelegation, differentAddr, defaultPubkeyBase64Str, defaultValue,
				}

				logCheckArgs := defaultLogCheck.WithErrContains(
					fmt.Sprintf(staking.ErrDifferentOriginFromDelegator, s.keyring.GetAddr(0), differentAddr),
				)

				_, _, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the contract and checking logs")
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
		)

		BeforeEach(func() {
			// populate the default editValidator args
			callArgs.MethodName = staking.EditValidatorMethod
		})

		Context("when origin is equal to validator address", func() {
			It("should succeed", func() {
				// create a new validator
				newAddr, newPriv := testutiltx.NewAccAddressAndKey()
				hexAddr := common.BytesToAddress(newAddr.Bytes())

				err := testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), newAddr, math.NewInt(2e18))
				Expect(err).To(BeNil(), "error while sending coins")
				Expect(s.network.NextBlock()).To(BeNil())

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

				createValidatorArgs := factory.CallArgs{
					ContractABI: s.precompile.ABI,
					MethodName:  staking.CreateValidatorMethod,
					Args:        []interface{}{description, commission, minSelfDelegation, hexAddr, pubkeyBase64Str, value},
				}

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeCreateValidator)
				_, _, err = s.factory.CallContractAndCheckLogs(
					newPriv,
					txArgs, createValidatorArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the contract and checking logs")
				Expect(s.network.NextBlock()).To(BeNil())

				// edit validator
				callArgs.Args = []interface{}{defaultDescription, hexAddr, defaultCommissionRate, defaultMinSelfDelegation}

				logCheckArgs = passCheck.WithExpEvents(staking.EventTypeEditValidator)
				_, _, err = s.factory.CallContractAndCheckLogs(
					newPriv,
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the contract and checking logs")
				Expect(s.network.NextBlock()).To(BeNil())

				valOperAddr := sdk.ValAddress(newAddr.Bytes()).String()
				qc := s.network.GetStakingClient()
				res, err := qc.Validator(s.network.GetContext(), &stakingtypes.QueryValidatorRequest{ValidatorAddr: valOperAddr})
				Expect(err).To(BeNil())
				Expect(res).NotTo(BeNil())
				validator := res.Validator
				Expect(validator.OperatorAddress).To(Equal(valOperAddr))
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
				valHexAddr := common.BytesToAddress(valAddr.Bytes())
				callArgs.Args = []interface{}{
					defaultDescription, valHexAddr, defaultCommissionRate, defaultMinSelfDelegation,
				}

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeEditValidator)
				_, _, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(1),
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).NotTo(BeNil(), "error while calling the contract and checking logs")
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("origin address %s is not the same as validator operator address %s", s.keyring.GetAddr(1), valHexAddr)))
			})
		})
	})
	Describe("to delegate", func() {
		// prevDelegation is the delegation that is available prior to the test (an initial delegation is
		// added in the test suite setup).
		var prevDelegation stakingtypes.Delegation

		BeforeEach(func() {
			delegator := s.keyring.GetKey(0)

			// get the delegation that is available prior to the test
			res, err := s.grpcHandler.GetDelegation(delegator.AccAddr.String(), valAddr.String())
			Expect(err).To(BeNil())
			Expect(res.DelegationResponse).NotTo(BeNil())

			prevDelegation = res.DelegationResponse.Delegation
			// populate the default delegate args
			callArgs.MethodName = staking.DelegateMethod
		})

		Context("as the token owner", func() {
			It("should delegate without need for authorization", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(2e18),
				}

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

				_, _, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				res, err := s.grpcHandler.GetDelegation(delegator.AccAddr.String(), valAddr.String())
				Expect(err).To(BeNil())
				Expect(res.DelegationResponse).NotTo(BeNil())
				expShares := prevDelegation.GetShares().Add(math.LegacyNewDec(2))
				Expect(res.DelegationResponse.Delegation.GetShares()).To(Equal(expShares), "expected different delegation shares")
			})

			It("should not delegate if the account has no sufficient balance", func() {
				newAddr, newAddrPriv := testutiltx.NewAccAddressAndKey()
				err := testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), newAddr, math.NewInt(1e17))
				Expect(err).To(BeNil(), "error while sending coins")
				Expect(s.network.NextBlock()).To(BeNil())

				// try to delegate more than left in account
				callArgs.Args = []interface{}{
					common.BytesToAddress(newAddr), valAddr.String(), big.NewInt(1e18),
				}

				logCheckArgs := defaultLogCheck.WithErrContains("insufficient funds")

				_, _, err = s.factory.CallContractAndCheckLogs(
					newAddrPriv,
					txArgs,
					callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			})

			It("should not delegate if the validator does not exist", func() {
				nonExistingAddr := testutiltx.GenerateAddress()
				nonExistingValAddr := sdk.ValAddress(nonExistingAddr.Bytes())
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, nonExistingValAddr.String(), big.NewInt(2e18),
				}

				logCheckArgs := defaultLogCheck.WithErrContains("validator does not exist")

				_, _, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs,
					callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			})
		})

		Context("on behalf of another account", func() {
			It("should not delegate if delegator address is not the origin", func() {
				delegator := s.keyring.GetKey(0)
				differentAddr := testutiltx.GenerateAddress()

				callArgs.Args = []interface{}{
					differentAddr, valAddr.String(), big.NewInt(2e18),
				}

				logCheckArgs := defaultLogCheck.WithErrContains(
					fmt.Sprintf(staking.ErrDifferentOriginFromDelegator, delegator.Addr, differentAddr),
				)

				_, _, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs,
					callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			})
		})
	})

	Describe("to undelegate", func() {
		BeforeEach(func() {
			callArgs.MethodName = staking.UndelegateMethod
		})

		Context("as the token owner", func() {
			It("should undelegate without need for authorization", func() {
				delegator := s.keyring.GetKey(0)

				valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
				Expect(err).To(BeNil())

				res, err := s.grpcHandler.GetValidatorUnbondingDelegations(valAddr.String())
				Expect(err).To(BeNil())
				Expect(res.UnbondingResponses).To(HaveLen(0), "expected no unbonding delegations before test")

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(1e18),
				}

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeUnbond)

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				delUbdRes, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())
				Expect(delUbdRes.UnbondingResponses).To(HaveLen(1), "expected one undelegation")
				Expect(delUbdRes.UnbondingResponses[0].ValidatorAddress).To(Equal(valAddr.String()), "expected validator address to be %s", valAddr)
			})

			It("should not undelegate if the amount exceeds the delegation", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(2e18),
				}

				logCheckArgs := defaultLogCheck.WithErrContains("invalid shares amount")

				_, _, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			})

			It("should not undelegate if the validator does not exist", func() {
				delegator := s.keyring.GetKey(0)
				nonExistingAddr := testutiltx.GenerateAddress()
				nonExistingValAddr := sdk.ValAddress(nonExistingAddr.Bytes())

				callArgs.Args = []interface{}{
					delegator.Addr, nonExistingValAddr.String(), big.NewInt(1e18),
				}

				logCheckArgs := defaultLogCheck.WithErrContains("validator does not exist")

				_, _, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			})
		})

		Context("on behalf of another account", func() {
			It("should not undelegate if delegator address is not the origin", func() {
				differentAddr := testutiltx.GenerateAddress()
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					differentAddr, valAddr.String(), big.NewInt(1e18),
				}

				logCheckArgs := defaultLogCheck.WithErrContains(
					fmt.Sprintf(staking.ErrDifferentOriginFromDelegator, delegator.Addr, differentAddr),
				)

				_, _, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			})
		})
	})

	Describe("to redelegate", func() {
		BeforeEach(func() {
			callArgs.MethodName = staking.RedelegateMethod
		})

		Context("as the token owner", func() {
			It("should redelegate without need for authorization", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), valAddr2.String(), big.NewInt(1e18),
				}

				logCheckArgs := passCheck.
					WithExpEvents(staking.EventTypeRedelegate)

				_, _, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				res, err := s.grpcHandler.GetRedelegations(delegator.AccAddr.String(), valAddr.String(), valAddr2.String())
				Expect(err).To(BeNil())
				Expect(res.RedelegationResponses).To(HaveLen(1), "expected one redelegation to be found")
				bech32Addr := delegator.AccAddr
				Expect(res.RedelegationResponses[0].Redelegation.DelegatorAddress).To(Equal(bech32Addr.String()), "expected delegator address to be %s", delegator.Addr)
				Expect(res.RedelegationResponses[0].Redelegation.ValidatorSrcAddress).To(Equal(valAddr.String()), "expected source validator address to be %s", valAddr)
				Expect(res.RedelegationResponses[0].Redelegation.ValidatorDstAddress).To(Equal(valAddr2.String()), "expected destination validator address to be %s", valAddr2)
			})

			It("should not redelegate if the amount exceeds the delegation", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), valAddr2.String(), big.NewInt(2e18),
				}

				logCheckArgs := defaultLogCheck.WithErrContains("invalid shares amount")

				_, _, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			})

			It("should not redelegate if the validator does not exist", func() {
				nonExistingAddr := testutiltx.GenerateAddress()
				nonExistingValAddr := sdk.ValAddress(nonExistingAddr.Bytes())
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), nonExistingValAddr.String(), big.NewInt(1e18),
				}

				logCheckArgs := defaultLogCheck.WithErrContains("redelegation destination validator not found")

				_, _, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			})
		})

		Context("on behalf of another account", func() {
			It("should not redelegate if delegator address is not the origin", func() {
				differentAddr := testutiltx.GenerateAddress()
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					differentAddr, valAddr.String(), valAddr2.String(), big.NewInt(1e18),
				}

				logCheckArgs := defaultLogCheck.WithErrContains(
					fmt.Sprintf(staking.ErrDifferentOriginFromDelegator, delegator.Addr, differentAddr),
				)

				_, _, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			})
		})
	})

	Describe("to cancel an unbonding delegation", func() {
		BeforeEach(func() {
			callArgs.MethodName = staking.CancelUnbondingDelegationMethod
			delegator := s.keyring.GetKey(0)

			// Set up an unbonding delegation
			undelegateArgs := factory.CallArgs{
				ContractABI: s.precompile.ABI,
				MethodName:  staking.UndelegateMethod,
				Args: []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(1e18),
				},
			}

			logCheckArgs := passCheck.
				WithExpEvents(staking.EventTypeUnbond)

			_, _, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs,
				undelegateArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while setting up an unbonding delegation: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			creationHeight := s.network.GetContext().BlockHeight()

			// Check that the unbonding delegation was created
			res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
			Expect(err).To(BeNil())
			Expect(res.UnbondingResponses).To(HaveLen(1), "expected one unbonding delegation to be found")
			Expect(res.UnbondingResponses[0].DelegatorAddress).To(Equal(delegator.AccAddr.String()), "expected delegator address to be %s", delegator.Addr)
			Expect(res.UnbondingResponses[0].ValidatorAddress).To(Equal(valAddr.String()), "expected validator address to be %s", valAddr)
			Expect(res.UnbondingResponses[0].Entries).To(HaveLen(1), "expected one unbonding delegation entry to be found")
			Expect(res.UnbondingResponses[0].Entries[0].CreationHeight).To(Equal(creationHeight), "expected different creation height")
			Expect(res.UnbondingResponses[0].Entries[0].Balance).To(Equal(math.NewInt(1e18)), "expected different balance")
		})

		Context("as the token owner", func() {
			It("should cancel unbonding delegation", func() {
				delegator := s.keyring.GetKey(0)

				valDelRes, err := s.grpcHandler.GetValidatorDelegations(s.network.GetValidators()[0].GetOperator())
				Expect(err).To(BeNil())
				Expect(valDelRes.DelegationResponses).To(HaveLen(0))

				creationHeight := s.network.GetContext().BlockHeight()
				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(1e18), big.NewInt(creationHeight),
				}

				logCheckArgs := passCheck.
					WithExpEvents(staking.EventTypeCancelUnbondingDelegation)

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs,
					callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())
				Expect(res.UnbondingResponses).To(HaveLen(0), "expected unbonding delegation to be canceled")

				valDelRes, err = s.grpcHandler.GetValidatorDelegations(s.network.GetValidators()[0].GetOperator())
				Expect(err).To(BeNil())
				Expect(valDelRes.DelegationResponses).To(HaveLen(1), "expected one delegation to be found")
			})

			It("should not cancel an unbonding delegation if the amount is not correct", func() {
				delegator := s.keyring.GetKey(0)

				creationHeight := s.network.GetContext().BlockHeight()
				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(2e18), big.NewInt(creationHeight),
				}

				logCheckArgs := defaultLogCheck.WithErrContains("amount is greater than the unbonding delegation entry balance")

				_, _, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())
				Expect(res.UnbondingResponses).To(HaveLen(1), "expected unbonding delegation not to have been canceled")
			})

			It("should not cancel an unbonding delegation if the creation height is not correct", func() {
				delegator := s.keyring.GetKey(0)

				creationHeight := s.network.GetContext().BlockHeight()
				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(1e18), big.NewInt(creationHeight + 1),
				}

				logCheckArgs := defaultLogCheck.WithErrContains("unbonding delegation entry is not found at block height")

				_, _, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())
				Expect(res.UnbondingResponses).To(HaveLen(1), "expected unbonding delegation not to have been canceled")
			})
		})
	})

	Describe("Calling precompile txs from a vesting account", func() {
		var (
			vestAcc         common.Address
			vestAccPriv     *ethsecp256k1.PrivKey
			clawbackAccount *vestingtypes.ClawbackVestingAccount
			unvested        sdk.Coins
			vested          sdk.Coins
			// unlockedVested are unlocked vested coins of the vesting schedule
			unlockedVested sdk.Coins
			delegateArgs   factory.CallArgs
		)

		BeforeEach(func() {
			// Setup vesting account
			funderKey := s.keyring.GetKey(0)
			vestAccKey := keyring.NewKey()
			vestAcc, vestAccPriv = vestAccKey.Addr, vestAccKey.Priv.(*ethsecp256k1.PrivKey)
			vestingAmtTotal := evmosutil.TestVestingSchedule.TotalVestingCoins

			clawbackAccount = s.setupVestingAccount(funderKey, vestAccKey)

			// Check if all tokens are unvested at vestingStart
			ctx := s.network.GetContext()
			unvested = clawbackAccount.GetVestingCoins(ctx.BlockTime())
			vested = clawbackAccount.GetVestedCoins(ctx.BlockTime())
			Expect(vestingAmtTotal).To(Equal(unvested))
			Expect(vested.IsZero()).To(BeTrue())

			// populate the default delegate args
			delegateArgs = factory.CallArgs{
				ContractABI: s.precompile.ABI,
				MethodName:  staking.DelegateMethod,
			}
		})

		Context("before first vesting period - all tokens locked and unvested", func() {
			BeforeEach(func() {
				Expect(s.network.NextBlock()).To(BeNil())

				ctx := s.network.GetContext()
				// Ensure no tokens are vested
				vested = clawbackAccount.GetVestedCoins(ctx.BlockTime())
				unvested = clawbackAccount.GetVestingCoins(ctx.BlockTime())
				unlocked := clawbackAccount.GetUnlockedCoins(ctx.BlockTime())
				zeroCoins := sdk.NewCoins(sdk.NewCoin(s.bondDenom, math.ZeroInt()))
				Expect(vested).To(Equal(zeroCoins), "expected different vested coins")
				Expect(unvested).To(Equal(evmosutil.TestVestingSchedule.TotalVestingCoins), "expected different unvested coins")
				Expect(unlocked).To(Equal(zeroCoins), "expected different unlocked coins")
			})

			It("Should not be able to delegate unvested tokens", func() {
				delegateArgs.Args = []interface{}{
					vestAcc, valAddr.String(), unvested.AmountOf(s.bondDenom).BigInt(),
				}

				failCheck := defaultLogCheck.
					WithErrContains("cannot delegate unvested coins")

				_, _, err := s.factory.CallContractAndCheckLogs(vestAccPriv, txArgs, delegateArgs, failCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			})

			It("Should be able to delegate tokens not involved in vesting schedule", func() {
				// send some coins to the vesting account
				coinsToDelegate := sdk.NewCoins(sdk.NewCoin(s.bondDenom, math.NewInt(1e18)))
				err := s.factory.FundAccount(s.keyring.GetKey(0), vestAcc.Bytes(), coinsToDelegate)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				// check balance is updated
				balRes, err := s.grpcHandler.GetBalance(clawbackAccount.GetAddress(), s.bondDenom)
				Expect(err).To(BeNil())
				Expect(*balRes.Balance).To(Equal(accountGasCoverage[0].Add(evmosutil.TestVestingSchedule.TotalVestingCoins[0]).Add(coinsToDelegate[0])))

				delegateArgs.Args = []interface{}{
					vestAcc, valAddr.String(), coinsToDelegate.AmountOf(s.bondDenom).BigInt(),
				}

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

				_, _, err = s.factory.CallContractAndCheckLogs(vestAccPriv, txArgs, delegateArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				delRes, err := s.grpcHandler.GetDelegation(sdk.AccAddress(vestAcc.Bytes()).String(), valAddr.String())
				Expect(err).To(BeNil(), "expected delegation to be found")
				Expect(delRes.DelegationResponse.Balance).To(Equal(coinsToDelegate[0]))

				// check vesting balance is untouched
				balRes, err = s.grpcHandler.GetBalance(vestAcc.Bytes(), s.bondDenom)
				Expect(err).To(BeNil())
				Expect(balRes.Balance.IsGTE(evmosutil.TestVestingSchedule.TotalVestingCoins[0])).To(BeTrue())
			})
		})

		Context("after first vesting period and before lockup - some vested tokens, but still all locked", func() {
			BeforeEach(func() {
				// Surpass cliff but none of lockup duration
				cliffDuration := time.Duration(evmosutil.TestVestingSchedule.CliffPeriodLength)
				Expect(s.network.NextBlockAfter(cliffDuration * time.Second)).To(BeNil())
				ctx := s.network.GetContext()

				// Check if some, but not all tokens are vested
				vested = clawbackAccount.GetVestedCoins(ctx.BlockTime())
				expVested := sdk.NewCoins(sdk.NewCoin(s.bondDenom, evmosutil.TestVestingSchedule.VestedCoinsPerPeriod[0].Amount.Mul(math.NewInt(evmosutil.TestVestingSchedule.CliffMonths))))
				Expect(vested).NotTo(Equal(evmosutil.TestVestingSchedule.TotalVestingCoins), "expected some tokens to have been vested")
				Expect(vested).To(Equal(expVested), "expected different vested amount")

				// check the vested tokens are still locked
				unlockedVested = clawbackAccount.GetUnlockedVestedCoins(ctx.BlockTime())
				Expect(unlockedVested).To(Equal(sdk.Coins{}))

				vestingAmtTotal := evmosutil.TestVestingSchedule.TotalVestingCoins
				res, err := s.network.App.VestingKeeper.Balances(ctx, &vestingtypes.QueryBalancesRequest{Address: clawbackAccount.Address})
				Expect(err).To(BeNil())
				Expect(res.Vested).To(Equal(expVested))
				Expect(res.Unvested).To(Equal(vestingAmtTotal.Sub(expVested...)))
				// All coins from vesting schedule should be locked
				Expect(res.Locked).To(Equal(vestingAmtTotal))
			})

			It("Should be able to delegate locked vested tokens", func() {
				delegateArgs.Args = []interface{}{
					vestAcc, valAddr.String(), vested[0].Amount.BigInt(),
				}

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

				_, _, err := s.factory.CallContractAndCheckLogs(vestAccPriv, txArgs, delegateArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				delRes, err := s.grpcHandler.GetDelegation(sdk.AccAddress(vestAcc.Bytes()).String(), valAddr.String())
				Expect(err).To(BeNil(), "expected delegation to be found")
				Expect(delRes.DelegationResponse.Balance).To(Equal(vested[0]))
			})

			It("Should be able to delegate locked vested tokens + free tokens (not in vesting schedule)", func() {
				// send some coins to the vesting account
				amt := sdk.NewCoins(sdk.NewCoin(s.bondDenom, math.NewInt(1e18)))

				err := s.factory.FundAccount(s.keyring.GetKey(0), vestAcc.Bytes(), amt)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				// check balance is updated
				balRes, err := s.grpcHandler.GetBalance(vestAcc.Bytes(), s.bondDenom)
				Expect(err).To(BeNil())
				Expect(*balRes.Balance).To(Equal(accountGasCoverage[0].Add(evmosutil.TestVestingSchedule.TotalVestingCoins[0]).Add(amt[0])))

				coinsToDelegate := amt.Add(vested...)

				delegateArgs.Args = []interface{}{
					vestAcc, valAddr.String(), coinsToDelegate[0].Amount.BigInt(),
				}

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

				_, _, err = s.factory.CallContractAndCheckLogs(vestAccPriv, txArgs, delegateArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				delRes, err := s.grpcHandler.GetDelegation(sdk.AccAddress(vestAcc.Bytes()).String(), valAddr.String())
				Expect(err).To(BeNil(), "expected delegation to be found")
				Expect(delRes.DelegationResponse.Balance).To(Equal(coinsToDelegate[0]))
			})
		})

		Context("Between first and second lockup periods - vested coins are unlocked", func() {
			BeforeEach(func() {
				// Surpass first lockup
				vestDuration := time.Duration(evmosutil.TestVestingSchedule.LockupPeriodLength)
				Expect(s.network.NextBlockAfter(vestDuration * time.Second)).To(BeNil())
				ctx := s.network.GetContext()

				// Check if some, but not all tokens are vested and unlocked
				vested = clawbackAccount.GetVestedCoins(ctx.BlockTime())
				unlocked := clawbackAccount.GetUnlockedCoins(ctx.BlockTime())
				unlockedVested = clawbackAccount.GetUnlockedVestedCoins(ctx.BlockTime())

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
				delegateArgs.Args = []interface{}{
					vestAcc, valAddr.String(), unlockedVested[0].Amount.BigInt(),
				}

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

				_, _, err := s.factory.CallContractAndCheckLogs(vestAccPriv, txArgs, delegateArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				delRes, err := s.grpcHandler.GetDelegation(sdk.AccAddress(vestAcc.Bytes()).String(), valAddr.String())
				Expect(err).To(BeNil(), "expected delegation to be found")
				Expect(delRes.DelegationResponse.Balance).To(Equal(unlockedVested[0]))
			})

			It("Cannot delegate more than vested tokens (and free tokens)", func() {
				ctx := s.network.GetContext()
				// calculate the delegatable amount
				balRes, err := s.grpcHandler.GetBalance(sdk.AccAddress(vestAcc.Bytes()), s.bondDenom)
				Expect(err).To(BeNil())
				unvestedOnly := clawbackAccount.GetVestingCoins(ctx.BlockTime())
				delegatable := balRes.Balance.Sub(unvestedOnly[0])

				delegateArgs.Args = []interface{}{
					vestAcc, valAddr.String(), delegatable.Amount.Add(math.OneInt()).BigInt(),
				}

				logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

				_, _, err = s.factory.CallContractAndCheckLogs(vestAccPriv, txArgs, delegateArgs, logCheckArgs)
				Expect(err).NotTo(BeNil(), "error while calling the smart contract: %v", err)
				Expect(err.Error()).To(ContainSubstring("cannot delegate unvested coins"))

				_, err = s.grpcHandler.GetDelegation(sdk.AccAddress(vestAcc.Bytes()).String(), valAddr.String())
				Expect(err).NotTo(BeNil(), "expected delegation NOT to be found")
				Expect(err.Error()).To(ContainSubstring("not found"), "expected delegation NOT to be found")
			})
		})
	})

	Describe("to query allowance", func() {
		differentAddr := testutiltx.GenerateAddress()

		BeforeEach(func() {
			callArgs.MethodName = authorization.AllowanceMethod
		})

		It("should return an empty allowance if none is set", func() {
			granter := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				granter.Addr, differentAddr, staking.CancelUnbondingDelegationMsg,
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var allowanceInt *big.Int
			err = s.precompile.UnpackIntoInterface(&allowanceInt, "allowance", ethRes.Ret)
			Expect(err).To(BeNil(), "error while unmarshalling the allowance: %v", err)
			Expect(allowanceInt.Int64()).To(BeZero(), "expected allowance to be zero")
		})

		It("should return the granted allowance if set", func() {
			granter := s.keyring.GetKey(0)

			// setup approval for another address
			s.SetupApproval(
				granter.Priv, differentAddr, big.NewInt(1e18), []string{staking.CancelUnbondingDelegationMsg},
			)

			// query allowance
			callArgs.Args = []interface{}{
				differentAddr, granter.Addr, staking.CancelUnbondingDelegationMsg,
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var allowanceInt *big.Int
			err = s.precompile.UnpackIntoInterface(&allowanceInt, "allowance", ethRes.Ret)
			Expect(err).To(BeNil(), "error while unmarshalling the allowance: %v", err)
			Expect(allowanceInt).To(Equal(big.NewInt(1e18)), "expected allowance to be 1e18")
		})
	})

	Describe("Validator queries", func() {
		BeforeEach(func() {
			callArgs.MethodName = staking.ValidatorMethod
		})

		It("should return validator", func() {
			delegator := s.keyring.GetKey(0)

			varHexAddr := common.BytesToAddress(valAddr.Bytes())
			callArgs.Args = []interface{}{varHexAddr}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)
			Expect(valOut.Validator.OperatorAddress).To(Equal(varHexAddr.String()), "expected validator address to match")
			Expect(valOut.Validator.DelegatorShares).To(Equal(big.NewInt(1e18)), "expected different delegator shares")
		})

		It("should return an empty validator if the validator is not found", func() {
			delegator := s.keyring.GetKey(0)

			newValHexAddr := testutiltx.GenerateAddress()
			callArgs.Args = []interface{}{newValHexAddr}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)
			Expect(valOut.Validator.OperatorAddress).To(Equal(""), "expected validator address to be empty")
			Expect(valOut.Validator.Status).To(BeZero(), "expected unspecified bonding status")
		})
	})

	Describe("Validators queries", func() {
		BeforeEach(func() {
			callArgs.MethodName = staking.ValidatorsMethod
		})

		It("should return validators (default pagination)", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				stakingtypes.Bonded.String(),
				query.PageRequest{},
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorsOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)

			Expect(valOut.PageResponse.NextKey).To(BeEmpty())
			Expect(valOut.PageResponse.Total).To(Equal(uint64(len(s.network.GetValidators()))))

			Expect(valOut.Validators).To(HaveLen(len(s.network.GetValidators())), "expected two validators to be returned")
			// return order can change, that's why each validator is checked individually
			for _, val := range valOut.Validators {
				s.CheckValidatorOutput(val)
			}
		})

		//nolint:dupl // this is a duplicate of the test for smart contract calls to the precompile
		It("should return validators w/pagination limit = 1", func() {
			const limit uint64 = 1
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				stakingtypes.Bonded.String(),
				query.PageRequest{
					Limit:      limit,
					CountTotal: true,
				},
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs,
				callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorsOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)

			// no pagination, should return default values
			Expect(valOut.PageResponse.NextKey).NotTo(BeEmpty())
			Expect(valOut.PageResponse.Total).To(Equal(uint64(len(s.network.GetValidators()))))

			Expect(valOut.Validators).To(HaveLen(int(limit)), "expected one validator to be returned")

			// return order can change, that's why each validator is checked individually
			for _, val := range valOut.Validators {
				s.CheckValidatorOutput(val)
			}
		})

		It("should return an error if the bonding type is not known", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				"15", // invalid bonding type
				query.PageRequest{},
			}

			invalidStatusCheck := defaultLogCheck.WithErrContains("invalid validator status 15")

			_, _, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs,
				callArgs,
				invalidStatusCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
		})

		It("should return an empty array if there are no validators with the given bonding type", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				stakingtypes.Unbonded.String(),
				query.PageRequest{},
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs,
				callArgs,
				passCheck,
			)
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
		BeforeEach(func() {
			callArgs.MethodName = staking.DelegationMethod
		})

		It("should return a delegation if it is found", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				delegator.Addr,
				valAddr.String(),
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs,
				callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var delOut staking.DelegationOutput
			err = s.precompile.UnpackIntoInterface(&delOut, staking.DelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the delegation output: %v", err)
			Expect(delOut.Shares).To(Equal(big.NewInt(1e18)), "expected different shares")
			Expect(delOut.Balance).To(Equal(cmn.Coin{Denom: s.bondDenom, Amount: big.NewInt(1e18)}), "expected different shares")
		})

		It("should return an empty delegation if it is not found", func() {
			delegator := s.keyring.GetKey(0)

			newValAddr := sdk.ValAddress(testutiltx.GenerateAddress().Bytes())
			callArgs.Args = []interface{}{
				delegator.Addr,
				newValAddr.String(),
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs,
				callArgs,
				passCheck,
			)
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
		// undelAmount is the amount of tokens to be unbonded
		undelAmount := big.NewInt(1e17)

		BeforeEach(func() {
			callArgs.MethodName = staking.UnbondingDelegationMethod

			delegator := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			// unbond a delegation
			s.SetupApproval(delegator.Priv, grantee.Addr, abi.MaxUint256, []string{staking.UndelegateMsg})

			undelegateArgs := factory.CallArgs{
				ContractABI: s.precompile.ABI,
				MethodName:  staking.UndelegateMethod,
				Args: []interface{}{
					delegator.Addr, valAddr.String(), undelAmount,
				},
			}

			unbondCheck := passCheck.WithExpEvents(staking.EventTypeUnbond)
			_, _, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, undelegateArgs,
				unbondCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			// check that the unbonding delegation exists
			res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
			Expect(err).To(BeNil())
			Expect(res.UnbondingResponses).To(HaveLen(1), "expected one unbonding delegation")
		})

		It("should return an unbonding delegation if it is found", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				delegator.Addr,
				valAddr.String(),
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs,
				callArgs,
				passCheck,
			)
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
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				delegator.Addr,
				valAddr2.String(),
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs,
				callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var unbondingDelegationOutput staking.UnbondingDelegationOutput
			err = s.precompile.UnpackIntoInterface(&unbondingDelegationOutput, staking.UnbondingDelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the unbonding delegation output: %v", err)
			Expect(unbondingDelegationOutput.UnbondingDelegation.Entries).To(HaveLen(0), "expected one unbonding delegation entry")
		})
	})

	Describe("to query a redelegation", func() {
		BeforeEach(func() {
			callArgs.MethodName = staking.RedelegationMethod
		})

		It("should return the redelegation if it exists", func() {
			delegator := s.keyring.GetKey(0)
			granteeAddr := s.precompile.Address()

			// approve the redelegation
			s.SetupApproval(delegator.Priv, granteeAddr, abi.MaxUint256, []string{staking.RedelegateMsg})

			// create a redelegation
			redelegateArgs := factory.CallArgs{
				ContractABI: s.precompile.ABI,
				MethodName:  staking.RedelegateMethod,
				Args: []interface{}{
					delegator.Addr, valAddr.String(), valAddr2.String(), big.NewInt(1e17),
				},
			}

			redelegateCheck := passCheck.WithExpEvents(staking.EventTypeRedelegate)

			_, _, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, redelegateArgs,
				redelegateCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			// query the redelegation
			callArgs.Args = []interface{}{
				delegator.Addr,
				valAddr.String(),
				valAddr2.String(),
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var redelegationOutput staking.RedelegationOutput
			err = s.precompile.UnpackIntoInterface(&redelegationOutput, staking.RedelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the redelegation output: %v", err)
			Expect(redelegationOutput.Redelegation.Entries).To(HaveLen(1), "expected one redelegation entry")
			Expect(redelegationOutput.Redelegation.Entries[0].InitialBalance).To(Equal(big.NewInt(1e17)), "expected different initial balance")
			Expect(redelegationOutput.Redelegation.Entries[0].SharesDst).To(Equal(big.NewInt(1e17)), "expected different balance")
		})

		It("should return an empty output if the redelegation is not found", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				delegator.Addr,
				valAddr.String(),
				valAddr2.String(),
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var redelegationOutput staking.RedelegationOutput
			err = s.precompile.UnpackIntoInterface(&redelegationOutput, staking.RedelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the redelegation output: %v", err)
			Expect(redelegationOutput.Redelegation.Entries).To(HaveLen(0), "expected no redelegation entries")
		})
	})

	Describe("Redelegations queries", func() {
		var (
			// delAmt is the amount of tokens to be delegated
			delAmt = big.NewInt(3e17)
			// redelTotalCount is the total number of redelegations
			redelTotalCount uint64 = 1
		)

		BeforeEach(func() {
			delegator := s.keyring.GetKey(0)
			granteeAddr := s.precompile.Address()

			callArgs.MethodName = staking.RedelegationsMethod
			// create some redelegations
			s.SetupApproval(
				delegator.Priv, granteeAddr, abi.MaxUint256, []string{staking.RedelegateMsg},
			)

			redelegationsArgs := []factory.CallArgs{
				{
					ContractABI: s.precompile.ABI,
					MethodName:  staking.RedelegateMethod,
					Args: []interface{}{
						delegator.Addr, valAddr.String(), valAddr2.String(), delAmt,
					},
				},
				{
					ContractABI: s.precompile.ABI,
					MethodName:  staking.RedelegateMethod,
					Args: []interface{}{
						delegator.Addr, valAddr.String(), valAddr2.String(), delAmt,
					},
				},
			}

			logCheckArgs := passCheck.
				WithExpEvents(staking.EventTypeRedelegate)

			txArgs.GasLimit = 500_000
			for _, args := range redelegationsArgs {
				_, _, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, args,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while creating redelegation: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())
			}
		})

		It("should return all redelegations for delegator (default pagination)", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				delegator.Addr,
				"",
				"",
				query.PageRequest{},
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
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
			delegator := s.keyring.GetKey(0)

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
				callArgs.Args = []interface{}{
					delegator.Addr,
					"",
					"",
					pagination,
				}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

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
			callArgs.Args = []interface{}{
				common.Address{}, // passing in an empty address to filter for all redelegations from valAddr2
				valAddr2.String(),
				"",
				query.PageRequest{},
			}

			sender := s.keyring.GetKey(0)
			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				sender.Priv,
				txArgs,
				callArgs,
				passCheck,
			)
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
		delegator := s.keyring.GetKey(0)

		resBal, err := s.grpcHandler.GetBalance(delegator.AccAddr, s.bondDenom)
		Expect(err).To(BeNil(), "error while getting balance")
		balancePre := resBal.Balance
		gasPrice := big.NewInt(1e9)

		// Call the precompile with a lot of gas
		approveCallArgs.Args = []interface{}{
			s.precompile.Address(), big.NewInt(1e18), []string{staking.DelegateMsg},
		}
		txArgs.GasPrice = gasPrice

		approvalCheck := passCheck.WithExpEvents(authorization.EventTypeApproval)

		res, _, err := s.factory.CallContractAndCheckLogs(
			delegator.Priv,
			txArgs, approveCallArgs,
			approvalCheck,
		)
		Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
		Expect(s.network.NextBlock()).To(BeNil())

		resBal, err = s.grpcHandler.GetBalance(delegator.AccAddr, s.bondDenom)
		Expect(err).To(BeNil(), "error while getting balance")
		balancePost := resBal.Balance
		difference := balancePre.Sub(*balancePost)

		// NOTE: the expected difference is the gas price multiplied by the gas used, because the rest should be refunded
		expDifference := gasPrice.Int64() * res.GasUsed
		Expect(difference.Amount.Int64()).To(Equal(expDifference), "expected different total transaction cost")
	})
})

var _ = Describe("Calling staking precompile via Solidity", Ordered, func() {
	var (
		// s is the precompile test suite to use for the tests
		s *PrecompileTestSuite
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
		nonExistingVal             = sdk.ValAddress(nonExistingAddr.Bytes())
		testContractInitialBalance = math.NewInt(1e18)
	)

	BeforeAll(func() {
		stakingCallerContract, err = testdata.LoadStakingCallerContract()
		Expect(err).To(BeNil())
		stakingCallerTwoContract, err = testdata.LoadStakingCallerTwoContract()
		Expect(err).To(BeNil(), "error while loading the StakingCallerTwo contract")
		stakingReverterContract, err = contracts.LoadStakingReverterContract()
		Expect(err).To(BeNil(), "error while loading the StakingReverter contract")
	})

	BeforeEach(func() {
		s = new(PrecompileTestSuite)
		s.SetupTest()
		delegator := s.keyring.GetKey(0)

		contractAddr, err = s.factory.DeployContract(
			delegator.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: stakingCallerContract,
			},
		)
		Expect(err).To(BeNil(), "error while deploying the smart contract: %v", err)
		valAddr, err = sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
		Expect(err).To(BeNil())
		valAddr2, err = sdk.ValAddressFromBech32(s.network.GetValidators()[1].GetOperator())
		Expect(err).To(BeNil())

		Expect(s.network.NextBlock()).To(BeNil())

		// Deploy StakingCallerTwo contract
		contractTwoAddr, err = s.factory.DeployContract(
			delegator.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: stakingCallerTwoContract,
			},
		)
		Expect(err).To(BeNil(), "error while deploying the StakingCallerTwo contract")
		Expect(s.network.NextBlock()).To(BeNil())

		// Deploy StakingReverter contract
		stkReverterAddr, err = s.factory.DeployContract(
			delegator.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: stakingReverterContract,
			},
		)
		Expect(err).To(BeNil(), "error while deploying the StakingReverter contract")
		Expect(s.network.NextBlock()).To(BeNil())

		// send some funds to the StakingCallerTwo & StakingReverter contracts to transfer to the
		// delegator during the tx
		err := testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), contractTwoAddr.Bytes(), testContractInitialBalance)
		Expect(err).To(BeNil(), "error while funding the smart contract: %v", err)
		Expect(s.network.NextBlock()).To(BeNil())
		err = testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), stkReverterAddr.Bytes(), testContractInitialBalance)
		Expect(err).To(BeNil(), "error while funding the smart contract: %v", err)
		Expect(s.network.NextBlock()).To(BeNil())

		// check contract was correctly deployed
		cAcc := s.network.App.EvmKeeper.GetAccount(s.network.GetContext(), contractAddr)
		Expect(cAcc).ToNot(BeNil(), "contract account should exist")
		Expect(cAcc.IsContract()).To(BeTrue(), "account should be a contract")

		// populate default TxArgs
		txArgs.To = &contractAddr
		// populate default call args
		callArgs = factory.CallArgs{
			ContractABI: stakingCallerContract.ABI,
		}
		// populate default approval args
		approveCallArgs = factory.CallArgs{
			ContractABI: stakingCallerContract.ABI,
			MethodName:  "testApprove",
		}
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
			delegator := s.keyring.GetKey(0)

			// disable the precompile
			res, err := s.grpcHandler.GetEvmParams()
			Expect(err).To(BeNil(), "error while setting params")
			params := res.Params
			var activePrecompiles []string
			for _, precompile := range params.ActiveStaticPrecompiles {
				if precompile != s.precompile.Address().String() {
					activePrecompiles = append(activePrecompiles, precompile)
				}
			}
			params.ActiveStaticPrecompiles = activePrecompiles

			err = testutils.UpdateEvmParams(testutils.UpdateParamsInput{
				Tf:      s.factory,
				Network: s.network,
				Pk:      delegator.Priv,
				Params:  params,
			})
			Expect(err).To(BeNil(), "error while setting params")

			// try to call the precompile
			callArgs.MethodName = "testDelegate"
			callArgs.Args = []interface{}{
				delegator.Addr, valAddr.String(), big.NewInt(2e18),
			}

			_, _, err = s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "expected error while calling the precompile")
		})
	})

	Context("approving methods", func() {
		Context("with valid input", func() {
			It("should approve one method", func() {
				granter := s.keyring.GetKey(0)

				approveCallArgs.Args = []interface{}{
					contractAddr, []string{staking.DelegateMsg}, big.NewInt(1e18),
				}

				s.SetupApprovalWithContractCalls(granter, txArgs, approveCallArgs)
			})

			It("should approve all methods", func() {
				granter := s.keyring.GetKey(0)

				approveCallArgs.Args = []interface{}{
					contractAddr,
					[]string{staking.DelegateMsg, staking.RedelegateMsg, staking.UndelegateMsg, staking.CancelUnbondingDelegationMsg},
					big.NewInt(1e18),
				}
				txArgs.GasLimit = 1e8
				s.SetupApprovalWithContractCalls(granter, txArgs, approveCallArgs)
			})

			It("should update a previous approval", func() {
				granter := s.keyring.GetKey(0)

				approveCallArgs.Args = []interface{}{
					contractAddr, []string{staking.DelegateMsg}, big.NewInt(1e18),
				}

				s.SetupApprovalWithContractCalls(granter, txArgs, approveCallArgs)

				// update approval
				approveCallArgs.Args = []interface{}{
					contractAddr, []string{staking.DelegateMsg}, big.NewInt(2e18),
				}
				_, _, err = s.factory.CallContractAndCheckLogs(
					granter.Priv,
					txArgs, approveCallArgs,
					approvalCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				// check approvals
				authorization, expirationTime, err := CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, staking.DelegateAuthz, contractAddr, granter.Addr)
				Expect(err).To(BeNil())
				Expect(authorization).ToNot(BeNil(), "expected authorization to not be nil")
				Expect(expirationTime).ToNot(BeNil(), "expected expiration time to not be nil")
				Expect(authorization.MsgTypeURL()).To(Equal(staking.DelegateMsg), "expected authorization msg type url to be %s", staking.DelegateMsg)
				Expect(authorization.MaxTokens.Amount).To(Equal(math.NewInt(2e18)), "expected different max tokens after updated approval")
			})

			It("should remove approval when setting amount to zero", func() {
				granter := s.keyring.GetKey(0)

				approveCallArgs.Args = []interface{}{
					contractAddr, []string{staking.DelegateMsg}, big.NewInt(1e18),
				}
				s.SetupApprovalWithContractCalls(granter, txArgs, approveCallArgs)
				Expect(s.network.NextBlock()).To(BeNil())

				// check approvals pre-removal
				allAuthz, err := s.grpcHandler.GetAuthorizations(sdk.AccAddress(contractAddr.Bytes()).String(), granter.AccAddr.String())
				Expect(err).To(BeNil(), "error while reading authorizations")
				Expect(allAuthz).To(HaveLen(1), "expected no authorizations")

				approveCallArgs.Args = []interface{}{
					contractAddr, []string{staking.DelegateMsg}, big.NewInt(0),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					granter.Priv,
					txArgs, approveCallArgs,
					approvalCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract")
				Expect(s.network.NextBlock()).To(BeNil())

				// check approvals after approving with amount 0
				allAuthz, err = s.grpcHandler.GetAuthorizations(sdk.AccAddress(contractAddr.Bytes()).String(), granter.AccAddr.String())
				Expect(err).To(BeNil(), "error while reading authorizations")
				Expect(allAuthz).To(HaveLen(0), "expected no authorizations")
			})

			It("should not approve if the gas is not enough", func() {
				granter := s.keyring.GetKey(0)

				txArgs.GasLimit = 1e5
				approveCallArgs.Args = []interface{}{
					contractAddr,
					[]string{
						staking.DelegateMsg,
						staking.UndelegateMsg,
						staking.RedelegateMsg,
						staking.CancelUnbondingDelegationMsg,
					},
					big.NewInt(1e18),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					granter.Priv,
					txArgs, approveCallArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract")
			})
		})

		Context("with invalid input", func() {
			// TODO: enable once we check that origin is not the sender
			// It("shouldn't approve any methods for if the sender is the origin", func() {
			//	approveArgs := defaultApproveArgs.WithArgs(
			//		nonExistingAddr, []string{staking.DelegateMsg}, big.NewInt(1e18),
			//	)
			//
			//	_, _, err = s.factory.CallContractAndCheckLogs(
			//	Expect(err).To(BeNil(), "error while calling the smart contract")
			//
			//	// check approvals
			//	allAuthz, err := s.network.App.AuthzKeeper.GetAuthorizations(s.network.GetContext(), contractAddr.Bytes(), delegator.AccAddr)
			//	Expect(err).To(BeNil(), "error while reading authorizations")
			//	Expect(allAuthz).To(HaveLen(0), "expected no authorizations")
			// })

			It("shouldn't approve for invalid methods", func() {
				granter := s.keyring.GetKey(0)

				approveCallArgs.Args = []interface{}{
					contractAddr, []string{"invalid method"}, big.NewInt(1e18),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					granter.Priv,
					txArgs, approveCallArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract")

				// check approvals
				allAuthz, err := s.grpcHandler.GetAuthorizations(sdk.AccAddress(contractAddr.Bytes()).String(), granter.AccAddr.String())
				Expect(err).To(BeNil(), "error while reading authorizations")
				Expect(allAuthz).To(HaveLen(0), "expected no authorizations")
			})
		})
	})

	Context("to revoke an approval", func() {
		BeforeEach(func() {
			callArgs.MethodName = "testRevoke"
		})

		It("should revoke when sending as the granter", func() {
			granter := s.keyring.GetKey(0)

			// set up an approval to be revoked
			approveCallArgs.Args = []interface{}{
				contractAddr, []string{staking.DelegateMsg}, big.NewInt(1e18),
			}

			s.SetupApprovalWithContractCalls(granter, txArgs, approveCallArgs)

			callArgs.Args = []interface{}{contractAddr, []string{staking.DelegateMsg}}

			revocationCheck := passCheck.WithExpEvents(authorization.EventTypeRevocation)

			_, _, err = s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs, callArgs,
				revocationCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract")
			Expect(s.network.NextBlock()).To(BeNil())

			// check approvals
			authz, _, err := CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, staking.DelegateAuthz, contractAddr, granter.Addr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("no authorizations found for grantee %s and granter %s", contractAddr.Hex(), granter.Addr.Hex())))
			Expect(authz).To(BeNil(), "expected authorization to be revoked")
		})

		It("should not revoke when approval is issued by a different granter", func() {
			// Create a delegate authorization where the granter is a different account from the default test suite one
			createdAuthz := staking.DelegateAuthz
			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)
			differentGranterIdx := s.keyring.AddKey()
			differentGranter := s.keyring.GetKey(differentGranterIdx)
			validators, err := s.network.App.StakingKeeper.GetLastValidators(s.network.GetContext())
			Expect(err).To(BeNil())

			valAddrs := make([]sdk.ValAddress, len(validators))
			for i, val := range validators {
				parsedAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
				Expect(err).To(BeNil())
				valAddrs[i] = parsedAddr
			}
			delegationAuthz, err := stakingtypes.NewStakeAuthorization(
				valAddrs,
				nil,
				createdAuthz,
				&sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(1e18)},
			)
			Expect(err).To(BeNil(), "failed to create authorization")

			expiration := s.network.GetContext().BlockTime().Add(time.Hour * 24 * 365).UTC()
			err = s.network.App.AuthzKeeper.SaveGrant(s.network.GetContext(), grantee.AccAddr, differentGranter.AccAddr, delegationAuthz, &expiration)
			Expect(err).ToNot(HaveOccurred(), "failed to save authorization")
			authz, _, err := CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, createdAuthz, grantee.Addr, differentGranter.Addr)
			Expect(err).To(BeNil())
			Expect(authz).ToNot(BeNil(), "expected authorization to be created")

			callArgs.Args = []interface{}{grantee.Addr, []string{staking.DelegateMsg}}

			_, _, err = s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs, callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract")

			// check approvals
			authz, _, err = CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, createdAuthz, grantee.Addr, differentGranter.Addr)
			Expect(err).To(BeNil())
			Expect(authz).ToNot(BeNil(), "expected authorization not to be revoked")
		})

		It("should revert the execution when no approval is found", func() {
			granter := s.keyring.GetKey(0)
			callArgs.Args = []interface{}{contractAddr, []string{staking.DelegateMsg}}

			_, _, err = s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs, callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract")

			// check approvals
			authz, _, err := CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, staking.DelegateAuthz, contractAddr, granter.Addr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("no authorizations found for grantee %s and granter %s", contractAddr.Hex(), granter.Addr.Hex())))
			Expect(authz).To(BeNil(), "expected no authorization to be found")
		})

		It("should not revoke if the approval is for a different message type", func() {
			granter := s.keyring.GetKey(0)

			// set up an approval
			approveCallArgs.Args = []interface{}{
				contractAddr, []string{staking.DelegateMsg}, big.NewInt(1e18),
			}

			s.SetupApprovalWithContractCalls(granter, txArgs, approveCallArgs)

			Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

			callArgs.Args = []interface{}{contractAddr, []string{staking.UndelegateMsg}}

			_, _, err = s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs, callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract")

			// check approval is still there
			s.ExpectAuthorization(
				staking.DelegateAuthz,
				contractAddr,
				granter.Addr,
				&sdk.Coin{Denom: s.bondDenom, Amount: math.NewInt(1e18)},
			)
		})
	})

	Context("create a validator", func() {
		var (
			valPriv    *ethsecp256k1.PrivKey
			valAddr    sdk.AccAddress
			valHexAddr common.Address

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
		)

		BeforeEach(func() {
			callArgs.MethodName = "testCreateValidator"
			valAddr, valPriv = testutiltx.NewAccAddressAndKey()
			valHexAddr = common.BytesToAddress(valAddr.Bytes())
			err = testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), valAddr.Bytes(), math.NewInt(1e18))
			Expect(err).To(BeNil(), "error while funding account: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())
		})

		It("tx from validator operator - should NOT create a validator", func() {
			callArgs.Args = []interface{}{
				defaultDescription, defaultCommission, defaultMinSelfDelegation, valHexAddr, defaultPubkeyBase64Str, defaultValue,
			}

			_, _, err = s.factory.CallContractAndCheckLogs(
				valPriv,
				txArgs, callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract")
			Expect(s.network.NextBlock()).To(BeNil())

			qc := s.network.GetStakingClient()
			_, err := qc.Validator(s.network.GetContext(), &stakingtypes.QueryValidatorRequest{ValidatorAddr: sdk.ValAddress(valAddr).String()})
			Expect(err).NotTo(BeNil(), "expected validator NOT to be found")
			Expect(err.Error()).To(ContainSubstring("not found"), "expected validator NOT to be found")
		})

		It("tx from another EOA - should create a validator fail", func() {
			callArgs.Args = []interface{}{
				defaultDescription, defaultCommission, defaultMinSelfDelegation, valHexAddr, defaultPubkeyBase64Str, defaultValue,
			}

			_, _, err = s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs, callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract")
			Expect(s.network.NextBlock()).To(BeNil())

			qc := s.network.GetStakingClient()
			_, err := qc.Validator(s.network.GetContext(), &stakingtypes.QueryValidatorRequest{ValidatorAddr: sdk.ValAddress(valAddr).String()})
			Expect(err).NotTo(BeNil(), "expected validator NOT to be found")
			Expect(err.Error()).To(ContainSubstring("not found"), "expected validator NOT to be found")
		})
	})

	Context("to edit a validator", func() {
		var (
			valPriv    *ethsecp256k1.PrivKey
			valAddr    sdk.AccAddress
			valHexAddr common.Address

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
			callArgs.MethodName = "testEditValidator"

			// create a new validator
			valAddr, valPriv = testutiltx.NewAccAddressAndKey()
			valHexAddr = common.BytesToAddress(valAddr.Bytes())
			err = testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), valAddr.Bytes(), math.NewInt(2e18))
			Expect(err).To(BeNil(), "error while funding account: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

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

			createValidatorArgs := factory.CallArgs{
				ContractABI: s.precompile.ABI,
				MethodName:  staking.CreateValidatorMethod,
				Args:        []interface{}{description, commission, minSelfDelegation, valHexAddr, pubkeyBase64Str, value},
			}

			logCheckArgs := passCheck.WithExpEvents(staking.EventTypeCreateValidator)

			toAddr := s.precompile.Address()
			_, _, err = s.factory.CallContractAndCheckLogs(
				valPriv,
				evmtypes.EvmTxArgs{
					To: &toAddr,
				},
				createValidatorArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract")
			Expect(s.network.NextBlock()).To(BeNil())
		})

		It("with tx from validator operator - should NOT edit a validator", func() {
			callArgs.Args = []interface{}{
				defaultDescription, valHexAddr,
				defaultCommissionRate, defaultMinSelfDelegation,
			}

			_, _, err = s.factory.CallContractAndCheckLogs(
				valPriv,
				txArgs,
				callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract")
			Expect(s.network.NextBlock()).To(BeNil())

			qc := s.network.GetStakingClient()
			qRes, err := qc.Validator(s.network.GetContext(), &stakingtypes.QueryValidatorRequest{ValidatorAddr: sdk.ValAddress(valAddr).String()})
			Expect(err).To(BeNil())
			Expect(qRes).NotTo(BeNil())
			validator := qRes.Validator
			Expect(validator.Description.Moniker).NotTo(Equal(defaultDescription.Moniker), "expected validator moniker NOT to be updated")
		})

		It("with tx from another EOA - should fail", func() {
			callArgs.Args = []interface{}{
				defaultDescription, valHexAddr,
				defaultCommissionRate, defaultMinSelfDelegation,
			}

			_, _, err = s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract")
			Expect(s.network.NextBlock()).To(BeNil())

			// validator should remain unchanged
			qc := s.network.GetStakingClient()
			qRes, err := qc.Validator(s.network.GetContext(), &stakingtypes.QueryValidatorRequest{ValidatorAddr: sdk.ValAddress(valAddr).String()})
			Expect(err).To(BeNil())
			Expect(qRes).NotTo(BeNil())

			validator := qRes.Validator
			Expect(validator.Description.Moniker).To(Equal("original moniker"), "expected validator moniker is updated")
			Expect(validator.Commission.Rate.BigInt().String()).To(Equal("100000000000000000"), "expected validator commission rate remain unchanged")
		})
	})

	Context("delegating", func() {
		// prevDelegation is the delegation that is available prior to the test (an initial delegation is
		// added in the test suite setup).
		var prevDelegation stakingtypes.Delegation

		BeforeEach(func() {
			delegator := s.keyring.GetKey(0)

			// get the delegation that is available prior to the test
			res, err := s.grpcHandler.GetDelegation(delegator.AccAddr.String(), valAddr.String())
			Expect(err).To(BeNil())
			Expect(res.DelegationResponse).NotTo(BeNil())

			prevDelegation = res.DelegationResponse.Delegation
			callArgs.MethodName = "testDelegate"
		})
		Context("without approval set", func() {
			BeforeEach(func() {
				granter := s.keyring.GetKey(0)

				authz, _, err := CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, staking.DelegateAuthz, contractAddr, granter.Addr)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("no authorizations found for grantee %s and granter %s", contractAddr.Hex(), granter.Addr.Hex())))
				Expect(authz).To(BeNil(), "expected authorization to be nil")
			})

			It("should not delegate", func() {
				Expect(s.network.App.EvmKeeper.GetAccount(s.network.GetContext(), contractAddr)).ToNot(BeNil(), "expected contract to exist")
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(1e18),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				res, err := s.grpcHandler.GetDelegation(delegator.AccAddr.String(), valAddr.String())
				Expect(err).To(BeNil())
				Expect(res.DelegationResponse).NotTo(BeNil())
				Expect(res.DelegationResponse.Delegation).To(Equal(prevDelegation), "no new delegation to be found")
			})
		})

		Context("with approval set", func() {
			BeforeEach(func() {
				granter := s.keyring.GetKey(0)

				approveCallArgs.Args = []interface{}{
					contractAddr, []string{staking.DelegateMsg}, big.NewInt(1e18),
				}

				s.SetupApprovalWithContractCalls(granter, txArgs, approveCallArgs)
				// add gas limit to avoid out of gas error
				txArgs.GasLimit = 500_000
			})

			It("should delegate when not exceeding the allowance", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(1e18),
				}

				logCheckArgs := passCheck.
					WithExpEvents(staking.EventTypeDelegate)

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				res, err := s.grpcHandler.GetDelegation(delegator.AccAddr.String(), valAddr.String())
				Expect(err).To(BeNil())
				Expect(res.DelegationResponse).NotTo(BeNil())
				delegation := res.DelegationResponse.Delegation

				expShares := prevDelegation.GetShares().Add(math.LegacyNewDec(1))
				Expect(delegation.GetShares()).To(Equal(expShares), "expected delegation shares to be 2")
			})

			Context("Calling the precompile from the StakingReverter contract", func() {
				var (
					txSenderInitialBal     *sdk.Coin
					contractInitialBalance *sdk.Coin
					gasPrice               = math.NewInt(1e9)
					delAmt                 = math.NewInt(1e18)
				)

				BeforeEach(func() {
					// set approval for the StakingReverter contract
					s.SetupApproval(s.keyring.GetPrivKey(0), stkReverterAddr, delAmt.BigInt(), []string{staking.DelegateMsg})

					balRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
					Expect(err).To(BeNil())
					txSenderInitialBal = balRes.Balance
					balRes, err = s.grpcHandler.GetBalance(stkReverterAddr.Bytes(), s.bondDenom)
					Expect(err).To(BeNil())
					contractInitialBalance = balRes.Balance
				})

				It("should revert the changes and NOT delegate - successful tx", func() {
					callArgs := factory.CallArgs{
						ContractABI: stakingReverterContract.ABI,
						MethodName:  "run",
						Args: []interface{}{
							big.NewInt(5), s.network.GetValidators()[0].OperatorAddress,
						},
					}

					// Tx should be successful, but no state changes happened
					res, _, err := s.factory.CallContractAndCheckLogs(
						s.keyring.GetPrivKey(0),
						evmtypes.EvmTxArgs{
							To:       &stkReverterAddr,
							GasPrice: gasPrice.BigInt(),
						},
						callArgs,
						passCheck,
					)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
					Expect(s.network.NextBlock()).To(BeNil())

					fees := gasPrice.MulRaw(res.GasUsed)

					// contract balance should remain unchanged
					balRes, err := s.grpcHandler.GetBalance(stkReverterAddr.Bytes(), s.bondDenom)
					Expect(err).To(BeNil())
					contractFinalBalance := balRes.Balance
					Expect(contractFinalBalance.Amount).To(Equal(contractInitialBalance.Amount))

					// No delegation should be created
					_, err = s.grpcHandler.GetDelegation(sdk.AccAddress(stkReverterAddr.Bytes()).String(), s.network.GetValidators()[0].OperatorAddress)
					Expect(err).NotTo(BeNil())
					Expect(err.Error()).To(ContainSubstring("not found"), "expected NO delegation created")

					// Only fees deducted on tx sender
					balRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
					Expect(err).To(BeNil())
					txSenderFinalBal := balRes.Balance
					Expect(txSenderFinalBal.Amount).To(Equal(txSenderInitialBal.Amount.Sub(fees)))
				})

				It("should revert the changes and NOT delegate - failed tx - max precompile calls reached", func() {
					callArgs := factory.CallArgs{
						ContractABI: stakingReverterContract.ABI,
						MethodName:  "multipleDelegations",
						Args: []interface{}{
							big.NewInt(int64(evmtypes.MaxPrecompileCalls + 2)), s.network.GetValidators()[0].OperatorAddress,
						},
					}

					// Tx should fail due to MaxPrecompileCalls
					_, _, err := s.factory.CallContractAndCheckLogs(
						s.keyring.GetPrivKey(0),
						evmtypes.EvmTxArgs{
							To:       &stkReverterAddr,
							GasPrice: gasPrice.BigInt(),
						},
						callArgs,
						execRevertedCheck,
					)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					// contract balance should remain unchanged
					balRes, err := s.grpcHandler.GetBalance(stkReverterAddr.Bytes(), s.bondDenom)
					Expect(err).To(BeNil())
					contractFinalBalance := balRes.Balance
					Expect(contractFinalBalance.Amount).To(Equal(contractInitialBalance.Amount))

					// No delegation should be created
					_, err = s.grpcHandler.GetDelegation(sdk.AccAddress(stkReverterAddr.Bytes()).String(), s.network.GetValidators()[0].OperatorAddress)
					Expect(err).NotTo(BeNil())
					Expect(err.Error()).To(ContainSubstring("not found"), "expected NO delegation created")
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
					args                           factory.CallArgs
					delegatorInitialBal            *sdk.Coin
					contractInitialBalance         *sdk.Coin
					bondedTokensPoolInitialBalance *sdk.Coin
					delAmt                         = math.NewInt(1e18)
					gasPrice                       = math.NewInt(1e9)
					bondedTokensPoolAccAddr        = authtypes.NewModuleAddress("bonded_tokens_pool")
				)

				BeforeEach(func() {
					// set authorization for contract
					args = factory.CallArgs{
						ContractABI: stakingCallerTwoContract.ABI,
						MethodName:  "testApprove",
						Args: []interface{}{
							contractTwoAddr, []string{staking.DelegateMsg}, delAmt.BigInt(),
						},
					}

					s.SetupApprovalWithContractCalls(
						s.keyring.GetKey(0),
						evmtypes.EvmTxArgs{
							To: &contractTwoAddr,
						}, args)

					balRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
					Expect(err).To(BeNil())
					delegatorInitialBal = balRes.Balance
					balRes, err = s.grpcHandler.GetBalance(contractTwoAddr.Bytes(), s.bondDenom)
					Expect(err).To(BeNil())
					contractInitialBalance = balRes.Balance
					balRes, err = s.grpcHandler.GetBalance(bondedTokensPoolAccAddr, s.bondDenom)
					Expect(err).To(BeNil())
					bondedTokensPoolInitialBalance = balRes.Balance

					args.MethodName = "testDelegateWithCounterAndTransfer"
				})

				DescribeTable("should delegate and update balances accordingly", func(tc testCase) {
					args.Args = []interface{}{
						s.keyring.GetAddr(0), valAddr.String(), delAmt.BigInt(), tc.before, tc.after,
					}

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

					res, _, err := s.factory.CallContractAndCheckLogs(
						s.keyring.GetPrivKey(0),
						evmtypes.EvmTxArgs{
							To:       &contractTwoAddr,
							GasPrice: gasPrice.BigInt(),
						},
						args,
						logCheckArgs,
					)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
					Expect(s.network.NextBlock()).To(BeNil())

					fees := gasPrice.MulRaw(res.GasUsed)

					// check the contract's balance was deducted to fund the vesting account
					balRes, err := s.grpcHandler.GetBalance(contractTwoAddr.Bytes(), s.bondDenom)
					contractFinalBal := balRes.Balance
					Expect(err).To(BeNil())
					Expect(contractFinalBal.Amount).To(Equal(contractInitialBalance.Amount.Sub(transferToDelAmt)))

					qRes, err := s.grpcHandler.GetDelegation(s.keyring.GetAccAddr(0).String(), valAddr.String())
					Expect(err).To(BeNil())
					Expect(qRes).NotTo(BeNil(), "expected delegation to be found")
					delegation := qRes.DelegationResponse.Delegation
					expShares := prevDelegation.GetShares().Add(math.LegacyNewDec(1))
					Expect(delegation.GetShares()).To(Equal(expShares), "expected delegation shares to be 2")

					balRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
					Expect(err).To(BeNil())
					delegatorFinalBal := balRes.Balance
					Expect(delegatorFinalBal.Amount).To(Equal(delegatorInitialBal.Amount.Sub(fees).Sub(delAmt).Add(transferToDelAmt)))

					// check the bondedTokenPool is updated with the delegated tokens
					balRes, err = s.grpcHandler.GetBalance(bondedTokensPoolAccAddr, s.bondDenom)
					bondedTokensPoolFinalBalance := balRes.Balance
					Expect(err).To(BeNil())
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
					args.MethodName = "testDelegateWithTransfer"
					args.Args = []interface{}{
						common.BytesToAddress(bondedTokensPoolAccAddr),
						s.keyring.GetAddr(0), valAddr.String(), delAmt.BigInt(), true, true,
					}

					_, _, err := s.factory.CallContractAndCheckLogs(
						s.keyring.GetPrivKey(0),
						txArgs,
						args,
						execRevertedCheck,
					)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
					Expect(s.network.NextBlock()).To(BeNil())

					// contract balance should remain unchanged
					balRes, err := s.grpcHandler.GetBalance(contractTwoAddr.Bytes(), s.bondDenom)
					Expect(err).To(BeNil())
					contractFinalBal := balRes.Balance
					Expect(contractFinalBal.Amount).To(Equal(contractInitialBalance.Amount))

					// check the bondedTokenPool should remain unchanged
					balRes, err = s.grpcHandler.GetBalance(bondedTokensPoolAccAddr, s.bondDenom)
					Expect(err).To(BeNil())
					bondedTokensPoolFinalBalance := balRes.Balance
					Expect(bondedTokensPoolFinalBalance.Amount).To(Equal(bondedTokensPoolInitialBalance.Amount))
				})
			})

			It("should not delegate when exceeding the allowance", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(2e18),
				}
				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				res, err := s.grpcHandler.GetDelegation(delegator.AccAddr.String(), valAddr.String())
				Expect(err).To(BeNil())
				Expect(res.DelegationResponse).NotTo(BeNil())
				Expect(res.DelegationResponse.Delegation).To(Equal(prevDelegation), "no new delegation to be found")
			})

			It("should not delegate when sending from a different address", func() {
				delegator := s.keyring.GetKey(0)
				differentSender := s.keyring.GetKey(1)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(1e18),
				}
				_, _, err = s.factory.CallContractAndCheckLogs(
					differentSender.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				res, err := s.grpcHandler.GetDelegation(delegator.AccAddr.String(), valAddr.String())
				Expect(err).To(BeNil())
				Expect(res.DelegationResponse).NotTo(BeNil())
				Expect(res.DelegationResponse.Delegation).To(Equal(prevDelegation), "no new delegation to be found")
			})

			It("should not delegate when validator does not exist", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, nonExistingVal.String(), big.NewInt(1e18),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				res, err := s.grpcHandler.GetDelegation(delegator.AccAddr.String(), nonExistingVal.String())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("delegation with delegator %s not found for validator %s", delegator.AccAddr.String(), nonExistingVal.String())))
				Expect(res).To(BeNil())
			})

			It("shouldn't delegate to a validator that is not in the allow list of the approval", func() {
				// create a new validator, which is not included in the active set of the last block
				commValue := math.LegacyNewDecWithPrec(5, 2)
				commission := stakingtypes.NewCommissionRates(commValue, commValue, commValue)
				validatorKey := ed25519.GenPrivKey()
				delegator := s.keyring.GetKey(0)
				err := s.factory.CreateValidator(delegator.Priv, validatorKey.PubKey(), sdk.NewCoin(s.bondDenom, math.NewInt(1)), stakingtypes.Description{Moniker: "NewValidator"}, commission, math.NewInt(1))
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				newValAddr := sdk.ValAddress(delegator.AccAddr.Bytes())

				res, err := s.grpcHandler.GetDelegation(delegator.AccAddr.String(), newValAddr.String())
				Expect(err).To(BeNil())
				Expect(res.DelegationResponse).NotTo(BeNil())
				prevDelegation = res.DelegationResponse.Delegation

				callArgs.Args = []interface{}{
					delegator.Addr, newValAddr.String(), big.NewInt(2e18),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				res, err = s.grpcHandler.GetDelegation(delegator.AccAddr.String(), newValAddr.String())
				Expect(err).To(BeNil())
				Expect(res.DelegationResponse).NotTo(BeNil())

				delegation := res.DelegationResponse.Delegation
				Expect(delegation.GetShares()).To(Equal(prevDelegation.GetShares()), "expected only the delegation from creating the validator, no more")
			})
		})

		Describe("delegation from a vesting account", func() {
			var (
				vestAcc         common.Address
				vestAccPriv     *ethsecp256k1.PrivKey
				clawbackAccount *vestingtypes.ClawbackVestingAccount
				unvested        sdk.Coins
				vested          sdk.Coins
				// unlockedVested are unlocked vested coins of the vesting schedule
				unlockedVested sdk.Coins
			)

			BeforeEach(func() {
				// Setup vesting account
				funderKey := s.keyring.GetKey(0)
				vestAccKey := keyring.NewKey()
				vestAcc, vestAccPriv = vestAccKey.Addr, vestAccKey.Priv.(*ethsecp256k1.PrivKey)

				clawbackAccount = s.setupVestingAccount(funderKey, vestAccKey)
				ctx := s.network.GetContext()

				// Check if all tokens are unvested at vestingStart
				totalVestingCoins := evmosutil.TestVestingSchedule.TotalVestingCoins
				unvested = clawbackAccount.GetVestingCoins(ctx.BlockTime())
				vested = clawbackAccount.GetVestedCoins(ctx.BlockTime())
				Expect(unvested).To(Equal(totalVestingCoins))
				Expect(vested.IsZero()).To(BeTrue())

				approveCallArgs.Args = []interface{}{
					contractAddr, []string{staking.DelegateMsg}, totalVestingCoins.AmountOf(s.bondDenom).BigInt(),
				}

				// create approval to allow spending all vesting coins
				s.SetupApprovalWithContractCalls(vestAccKey, txArgs, approveCallArgs)
			})

			Context("before first vesting period - all tokens locked and unvested", func() {
				BeforeEach(func() {
					Expect(s.network.NextBlock()).To(BeNil())
					ctx := s.network.GetContext()

					// Ensure no tokens are vested
					vested = clawbackAccount.GetVestedCoins(ctx.BlockTime())
					unvested = clawbackAccount.GetVestingCoins(ctx.BlockTime())
					unlocked := clawbackAccount.GetUnlockedCoins(ctx.BlockTime())
					zeroCoins := sdk.NewCoins(sdk.NewCoin(s.bondDenom, math.ZeroInt()))
					Expect(vested).To(Equal(zeroCoins))
					Expect(unvested).To(Equal(evmosutil.TestVestingSchedule.TotalVestingCoins))
					Expect(unlocked).To(Equal(zeroCoins))
				})

				It("Should not be able to delegate unvested tokens", func() {
					callArgs.Args = []interface{}{
						vestAcc, valAddr.String(), unvested.AmountOf(s.bondDenom).BigInt(),
					}

					_, _, err := s.factory.CallContractAndCheckLogs(vestAccPriv, txArgs, callArgs, execRevertedCheck)
					Expect(err).NotTo(HaveOccurred(), "error while calling the smart contract: %v", err)
				})

				It("Should be able to delegate tokens not involved in vesting schedule", func() {
					// send some coins to the vesting account
					coinsToDelegate := sdk.NewCoins(sdk.NewCoin(s.bondDenom, math.NewInt(1e18)))
					err := s.factory.FundAccount(s.keyring.GetKey(0), vestAcc.Bytes(), coinsToDelegate)
					Expect(err).To(BeNil())
					Expect(s.network.NextBlock()).To(BeNil())

					callArgs.Args = []interface{}{
						vestAcc, valAddr.String(), coinsToDelegate.AmountOf(s.bondDenom).BigInt(),
					}

					logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

					_, _, err = s.factory.CallContractAndCheckLogs(vestAccPriv, txArgs, callArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					delRes, err := s.grpcHandler.GetDelegation(sdk.AccAddress(vestAcc.Bytes()).String(), valAddr.String())
					Expect(err).To(BeNil(), "expected delegation to be found")
					Expect(delRes.DelegationResponse.Balance).To(Equal(coinsToDelegate[0]))
				})
			})

			Context("after first vesting period and before lockup - some vested tokens, but still all locked", func() {
				BeforeEach(func() {
					// Surpass cliff but none of lockup duration
					cliffDuration := time.Duration(evmosutil.TestVestingSchedule.CliffPeriodLength)
					Expect(s.network.NextBlockAfter(cliffDuration * time.Second)).Error().To(BeNil())

					// Check if some, but not all tokens are vested
					vested = clawbackAccount.GetVestedCoins(s.network.GetContext().BlockTime())
					expVested := sdk.NewCoins(sdk.NewCoin(s.bondDenom, evmosutil.TestVestingSchedule.VestedCoinsPerPeriod[0].Amount.Mul(math.NewInt(evmosutil.TestVestingSchedule.CliffMonths))))
					Expect(vested).NotTo(Equal(evmosutil.TestVestingSchedule.TotalVestingCoins))
					Expect(vested).To(Equal(expVested))

					// check the vested tokens are still locked
					unlockedVested = clawbackAccount.GetUnlockedVestedCoins(s.network.GetContext().BlockTime())
					Expect(unlockedVested).To(Equal(sdk.Coins{}))

					vestingAmtTotal := evmosutil.TestVestingSchedule.TotalVestingCoins
					res, err := s.network.App.VestingKeeper.Balances(s.network.GetContext(), &vestingtypes.QueryBalancesRequest{Address: clawbackAccount.Address})
					Expect(err).To(BeNil())
					Expect(res.Vested).To(Equal(expVested))
					Expect(res.Unvested).To(Equal(vestingAmtTotal.Sub(expVested...)))
					// All coins from vesting schedule should be locked
					Expect(res.Locked).To(Equal(vestingAmtTotal))
				})

				It("Should be able to delegate locked vested tokens", func() {
					callArgs.Args = []interface{}{
						vestAcc, valAddr.String(), vested[0].Amount.BigInt(),
					}

					logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

					_, _, err := s.factory.CallContractAndCheckLogs(vestAccPriv, txArgs, callArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
					Expect(s.network.NextBlock()).To(BeNil())

					delRes, err := s.grpcHandler.GetDelegation(sdk.AccAddress(vestAcc.Bytes()).String(), valAddr.String())
					Expect(err).To(BeNil(), "expected delegation to be found")
					Expect(delRes.DelegationResponse.Balance).To(Equal(vested[0]))
				})

				It("Should be able to delegate locked vested tokens + free tokens (not in vesting schedule)", func() {
					// send some coins to the vesting account
					amt := sdk.NewCoins(sdk.NewCoin(s.bondDenom, math.NewInt(1e18)))

					err := s.factory.FundAccount(s.keyring.GetKey(0), vestAcc.Bytes(), amt)
					Expect(err).To(BeNil())
					Expect(s.network.NextBlock()).To(BeNil())

					coinsToDelegate := amt.Add(vested...)

					callArgs.Args = []interface{}{
						vestAcc, valAddr.String(), coinsToDelegate[0].Amount.BigInt(),
					}

					logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

					_, _, err = s.factory.CallContractAndCheckLogs(vestAccPriv, txArgs, callArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
					Expect(s.network.NextBlock()).To(BeNil())

					delRes, err := s.grpcHandler.GetDelegation(sdk.AccAddress(vestAcc.Bytes()).String(), valAddr.String())
					Expect(err).To(BeNil(), "expected delegation to be found")
					Expect(delRes.DelegationResponse.Balance).To(Equal(coinsToDelegate[0]))
				})
			})

			Context("Between first and second lockup periods - vested coins are unlocked", func() {
				BeforeEach(func() {
					// Surpass first lockup
					vestDuration := time.Duration(evmosutil.TestVestingSchedule.LockupPeriodLength)
					Expect(s.network.NextBlockAfter(vestDuration * time.Second)).To(BeNil())

					// Check if some, but not all tokens are vested and unlocked
					vested = clawbackAccount.GetVestedCoins(s.network.GetContext().BlockTime())
					unlocked := clawbackAccount.GetUnlockedCoins(s.network.GetContext().BlockTime())
					unlockedVested = clawbackAccount.GetUnlockedVestedCoins(s.network.GetContext().BlockTime())

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
					callArgs.Args = []interface{}{
						vestAcc, valAddr.String(), unlockedVested[0].Amount.BigInt(),
					}

					logCheckArgs := passCheck.WithExpEvents(staking.EventTypeDelegate)

					_, _, err := s.factory.CallContractAndCheckLogs(vestAccPriv, txArgs, callArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					delRes, err := s.grpcHandler.GetDelegation(sdk.AccAddress(vestAcc.Bytes()).String(), valAddr.String())
					Expect(err).To(BeNil(), "expected delegation to be found")
					Expect(delRes.DelegationResponse.Balance).To(Equal(unlockedVested[0]))
				})
			})
		})
	})

	Context("unbonding", func() {
		// NOTE: there's no additional setup necessary because the test suite is already set up with
		// delegations to the validator
		BeforeEach(func() {
			callArgs.MethodName = "testUndelegate"
		})
		Context("without approval set", func() {
			BeforeEach(func() {
				delegator := s.keyring.GetKey(0)

				authz, _, err := CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, staking.UndelegateAuthz, contractAddr, delegator.Addr)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("no authorizations found for grantee %s and granter %s", contractAddr.Hex(), delegator.Addr.Hex())))
				Expect(authz).To(BeNil(), "expected authorization to be nil before test execution")
			})
			It("should not undelegate", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(1e18),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())
				Expect(res.UnbondingResponses).To(BeEmpty())
			})
		})

		Context("with approval set", func() {
			BeforeEach(func() {
				granter := s.keyring.GetKey(0)

				approveCallArgs.Args = []interface{}{
					contractAddr, []string{staking.UndelegateMsg}, big.NewInt(1e18),
				}

				s.SetupApprovalWithContractCalls(granter, txArgs, approveCallArgs)
				// set gas limit to avoid out of gas error
				txArgs.GasLimit = 500_000
			})

			It("should undelegate when not exceeding the allowance", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(1e18),
				}

				logCheckArgs := defaultLogCheck.
					WithExpEvents(staking.EventTypeUnbond).
					WithExpPass(true)

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())
				Expect(res.UnbondingResponses).To(HaveLen(1), "expected one undelegation")
				Expect(res.UnbondingResponses[0].ValidatorAddress).To(Equal(valAddr.String()), "expected validator address to be %s", valAddr)
			})

			It("should not undelegate when exceeding the allowance", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(2e18),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())
				Expect(res.UnbondingResponses).To(BeEmpty())
			})

			It("should not undelegate if the delegation does not exist", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, nonExistingVal.String(), big.NewInt(1e18),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())
				Expect(res.UnbondingResponses).To(BeEmpty())
			})

			It("should not undelegate when called from a different address", func() {
				delegator := s.keyring.GetKey(0)
				differentSender := s.keyring.GetKey(1)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(1e18),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					differentSender.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())
				Expect(res.UnbondingResponses).To(BeEmpty())
			})
		})
	})

	Context("redelegating", func() {
		// NOTE: there's no additional setup necessary because the test suite is already set up with
		// delegations to the validator

		BeforeEach(func() {
			callArgs.MethodName = "testRedelegate"
		})
		Context("without approval set", func() {
			BeforeEach(func() {
				granter := s.keyring.GetKey(0)

				authz, _, err := CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, staking.UndelegateAuthz, contractAddr, granter.Addr)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("no authorizations found for grantee %s and granter %s", contractAddr.Hex(), granter.Addr.Hex())))
				Expect(authz).To(BeNil(), "expected authorization to be nil before test execution")
			})

			It("should not redelegate", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), valAddr2.String(), big.NewInt(1e18),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				res, err := s.grpcHandler.GetRedelegations(delegator.AccAddr.String(), valAddr.String(), valAddr2.String())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("redelegation not found for delegator address %s from validator address %s", delegator.AccAddr, valAddr)))
				Expect(res).To(BeNil(), "expected no redelegations to be found")
			})
		})

		Context("with approval set", func() {
			BeforeEach(func() {
				granter := s.keyring.GetKey(0)

				approveCallArgs.Args = []interface{}{
					contractAddr, []string{staking.RedelegateMsg}, big.NewInt(1e18),
				}

				s.SetupApprovalWithContractCalls(granter, txArgs, approveCallArgs)
			})

			It("should redelegate when not exceeding the allowance", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), valAddr2.String(), big.NewInt(1e18),
				}

				logCheckArgs := defaultLogCheck.
					WithExpEvents(staking.EventTypeRedelegate).
					WithExpPass(true)

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

				res, err := s.grpcHandler.GetRedelegations(delegator.AccAddr.String(), valAddr.String(), valAddr2.String())
				Expect(err).To(BeNil())
				Expect(res.RedelegationResponses).To(HaveLen(1), "expected one redelegation to be found")
				Expect(res.RedelegationResponses[0].Redelegation.DelegatorAddress).To(Equal(delegator.AccAddr.String()), "expected delegator address to be %s", delegator.AccAddr)
				Expect(res.RedelegationResponses[0].Redelegation.ValidatorSrcAddress).To(Equal(valAddr.String()), "expected source validator address to be %s", valAddr)
				Expect(res.RedelegationResponses[0].Redelegation.ValidatorDstAddress).To(Equal(valAddr2.String()), "expected destination validator address to be %s", valAddr2)
			})

			It("should not redelegate when exceeding the allowance", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), valAddr2.String(), big.NewInt(2e18),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				res, err := s.grpcHandler.GetRedelegations(delegator.AccAddr.String(), valAddr.String(), valAddr2.String())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("redelegation not found for delegator address %s from validator address %s", delegator.AccAddr, valAddr)))
				Expect(res).To(BeNil(), "expected no redelegations to be found")
			})

			It("should not redelegate if the delegation does not exist", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, nonExistingVal.String(), valAddr2.String(), big.NewInt(1e18),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				res, err := s.grpcHandler.GetRedelegations(delegator.AccAddr.String(), nonExistingVal.String(), valAddr2.String())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("redelegation not found for delegator address %s from validator address %s", delegator.AccAddr, nonExistingVal)))
				Expect(res).To(BeNil(), "expected no redelegations to be found")
			})

			It("should not redelegate when calling from a different address", func() {
				delegator := s.keyring.GetKey(0)
				differentSender := s.keyring.GetKey(1)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), valAddr2.String(), big.NewInt(1e18),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					differentSender.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				res, err := s.grpcHandler.GetRedelegations(delegator.AccAddr.String(), valAddr.String(), valAddr2.String())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("redelegation not found for delegator address %s from validator address %s", delegator.AccAddr, valAddr)))
				Expect(res).To(BeNil(), "expected no redelegations to be found")
			})

			It("should not redelegate when the validator does not exist", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), nonExistingVal.String(), big.NewInt(1e18),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				res, err := s.grpcHandler.GetRedelegations(delegator.AccAddr.String(), valAddr.String(), nonExistingVal.String())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("redelegation not found for delegator address %s from validator address %s", delegator.AccAddr, valAddr)))
				Expect(res).To(BeNil())
			})
		})
	})

	Context("canceling unbonding delegations", func() {
		// expCreationHeight is the expected creation height of the unbonding delegation
		var expCreationHeight int64

		BeforeEach(func() {
			granter := s.keyring.GetKey(0)

			callArgs.MethodName = "testCancelUnbonding"
			// Set up an unbonding delegation
			approveCallArgs.Args = []interface{}{
				contractAddr, []string{staking.UndelegateMsg}, big.NewInt(1e18),
			}

			s.SetupApprovalWithContractCalls(granter, txArgs, approveCallArgs)

			Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

			delegator := s.keyring.GetKey(0)
			undelegateArgs := factory.CallArgs{
				ContractABI: stakingCallerContract.ABI,
				MethodName:  "testUndelegate",
				Args:        []interface{}{delegator.Addr, valAddr.String(), big.NewInt(1e18)},
			}

			logCheckArgs := defaultLogCheck.
				WithExpEvents(staking.EventTypeUnbond).
				WithExpPass(true)

			_, _, err = s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, undelegateArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while setting up an unbonding delegation: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			expCreationHeight = s.network.GetContext().BlockHeight()
			// Check that the unbonding delegation was created
			res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
			Expect(err).To(BeNil())
			Expect(res.UnbondingResponses).To(HaveLen(1), "expected one unbonding delegation to be found")
			Expect(res.UnbondingResponses[0].DelegatorAddress).To(Equal(delegator.AccAddr.String()), "expected delegator address to be %s", delegator.Addr)
			Expect(res.UnbondingResponses[0].ValidatorAddress).To(Equal(valAddr.String()), "expected validator address to be %s", valAddr)
			Expect(res.UnbondingResponses[0].Entries).To(HaveLen(1), "expected one unbonding delegation entry to be found")
			Expect(res.UnbondingResponses[0].Entries[0].CreationHeight).To(Equal(expCreationHeight), "expected different creation height")
			Expect(res.UnbondingResponses[0].Entries[0].Balance).To(Equal(math.NewInt(1e18)), "expected different balance")
		})

		Context("without approval set", func() {
			It("should not cancel unbonding delegations", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(1e18), big.NewInt(expCreationHeight),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())
				Expect(res.UnbondingResponses).To(HaveLen(1), "expected unbonding delegation not to be canceled")
			})
		})

		Context("with approval set", func() {
			BeforeEach(func() {
				granter := s.keyring.GetKey(0)

				// Set up an unbonding delegation
				approveCallArgs.Args = []interface{}{
					contractAddr, []string{staking.CancelUnbondingDelegationMsg}, big.NewInt(1e18),
				}

				s.SetupApprovalWithContractCalls(granter, txArgs, approveCallArgs)

				Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")
			})

			It("should cancel unbonding delegations when not exceeding allowance", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(1e18), big.NewInt(expCreationHeight),
				}

				txArgs.GasLimit = 1e9

				logCheckArgs := passCheck.
					WithExpEvents(staking.EventTypeCancelUnbondingDelegation)

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())
				Expect(res.UnbondingResponses).To(BeEmpty(), "expected unbonding delegation to be canceled")
			})

			It("should not cancel unbonding delegations when exceeding allowance", func() {
				delegator := s.keyring.GetKey(0)

				approveCallArgs.Args = []interface{}{contractAddr, []string{staking.CancelUnbondingDelegationMsg}, big.NewInt(1)}
				s.SetupApprovalWithContractCalls(delegator, txArgs, approveCallArgs)

				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(1e18), big.NewInt(expCreationHeight),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())
				Expect(res.UnbondingResponses).To(HaveLen(1), "expected unbonding delegation to not be canceled")
			})

			It("should not cancel unbonding any delegations when unbonding delegation does not exist", func() {
				delegator := s.keyring.GetKey(0)

				callArgs.Args = []interface{}{
					delegator.Addr,
					nonExistingVal.String(),
					big.NewInt(1e18),
					big.NewInt(expCreationHeight),
				}

				_, _, err = s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs,
					callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())
				Expect(res.UnbondingResponses).To(HaveLen(1), "expected unbonding delegation to not be canceled")
			})

			It("should not cancel unbonding delegations when calling from a different address", func() {
				delegator := s.keyring.GetKey(0)
				differentSender := s.keyring.GetKey(1)

				callArgs.Args = []interface{}{delegator.Addr, valAddr.String(), big.NewInt(1e18), big.NewInt(expCreationHeight)}

				_, _, err = s.factory.CallContractAndCheckLogs(
					differentSender.Priv,
					txArgs, callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())
				Expect(res.UnbondingResponses).To(HaveLen(1), "expected unbonding delegation to not be canceled")
			})
		})
	})

	Context("querying allowance", func() {
		BeforeEach(func() {
			callArgs.MethodName = "getAllowance"
		})
		It("without approval set it should show no allowance", func() {
			granter := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				contractAddr, staking.CancelUnbondingDelegationMsg,
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var allowanceInt *big.Int
			err = s.precompile.UnpackIntoInterface(&allowanceInt, "allowance", ethRes.Ret)
			Expect(err).To(BeNil(), "error while unmarshalling the allowance: %v", err)
			Expect(allowanceInt.Int64()).To(Equal(int64(0)), "expected empty allowance")
		})

		It("with approval set it should show the granted allowance", func() {
			granter := s.keyring.GetKey(0)

			// setup approval
			approveCallArgs.Args = []interface{}{
				contractAddr, []string{staking.CancelUnbondingDelegationMsg}, big.NewInt(1e18),
			}

			s.SetupApprovalWithContractCalls(granter, txArgs, approveCallArgs)

			// query allowance
			callArgs.Args = []interface{}{
				contractAddr, staking.CancelUnbondingDelegationMsg,
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				granter.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var allowanceInt *big.Int
			err = s.precompile.UnpackIntoInterface(&allowanceInt, "allowance", ethRes.Ret)
			Expect(err).To(BeNil(), "error while unmarshalling the allowance: %v", err)
			Expect(allowanceInt).To(Equal(big.NewInt(1e18)), "expected allowance to be 1e18")
		})
	})

	Context("querying validator", func() {
		BeforeEach(func() {
			callArgs.MethodName = "getValidator"
		})
		It("with non-existing address should return an empty validator", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				nonExistingAddr,
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)
			Expect(valOut.Validator.OperatorAddress).To(Equal(""), "expected empty validator address")
			Expect(valOut.Validator.Status).To(Equal(uint8(0)), "expected validator status to be 0 (unspecified)")
		})

		It("with existing address should return the validator", func() {
			delegator := s.keyring.GetKey(0)

			valHexAddr := common.BytesToAddress(valAddr.Bytes())
			callArgs.Args = []interface{}{valHexAddr}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)
			Expect(valOut.Validator.OperatorAddress).To(Equal(valHexAddr.String()), "expected validator address to match")
			Expect(valOut.Validator.DelegatorShares).To(Equal(big.NewInt(1e18)), "expected different delegator shares")
		})

		It("with status bonded and pagination", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.MethodName = "getValidators"
			callArgs.Args = []interface{}{
				stakingtypes.Bonded.String(),
				query.PageRequest{
					Limit:      1,
					CountTotal: true,
				},
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorsOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)
			Expect(valOut.PageResponse.Total).To(Equal(uint64(len(s.network.GetValidators()))))
			Expect(valOut.PageResponse.NextKey).NotTo(BeEmpty())
			Expect(valOut.Validators[0].DelegatorShares).To(Equal(big.NewInt(1e18)), "expected different delegator shares")
		})
	})

	Context("querying validators", func() {
		BeforeEach(func() {
			callArgs.MethodName = "getValidators"
		})
		It("should return validators (default pagination)", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				stakingtypes.Bonded.String(),
				query.PageRequest{},
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorsOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)
			Expect(valOut.PageResponse.Total).To(Equal(uint64(len(s.network.GetValidators()))))
			Expect(valOut.PageResponse.NextKey).To(BeEmpty())
			Expect(valOut.Validators).To(HaveLen(len(s.network.GetValidators())), "expected all validators to be returned")
			// return order can change, that's why each validator is checked individually
			for _, val := range valOut.Validators {
				s.CheckValidatorOutput(val)
			}
		})

		//nolint:dupl // this is a duplicate of the test for EOA calls to the precompile
		It("should return validators with pagination limit = 1", func() {
			const limit uint64 = 1
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				stakingtypes.Bonded.String(),
				query.PageRequest{
					Limit:      limit,
					CountTotal: true,
				},
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var valOut staking.ValidatorsOutput
			err = s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorsMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the validator output: %v", err)

			// no pagination, should return default values
			Expect(valOut.PageResponse.NextKey).NotTo(BeEmpty())
			Expect(valOut.PageResponse.Total).To(Equal(uint64(len(s.network.GetValidators()))))

			Expect(valOut.Validators).To(HaveLen(int(limit)), "expected one validator to be returned")

			// return order can change, that's why each validator is checked individually
			for _, val := range valOut.Validators {
				s.CheckValidatorOutput(val)
			}
		})

		It("should revert the execution if the bonding type is not known", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				"15", // invalid bonding type
				query.PageRequest{},
			}

			_, _, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
		})

		It("should return an empty array if there are no validators with the given bonding type", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				stakingtypes.Unbonded.String(),
				query.PageRequest{},
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
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
		BeforeEach(func() {
			callArgs.MethodName = "getDelegation"
		})
		It("which does not exist should return an empty delegation", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				nonExistingAddr, valAddr.String(),
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var delOut staking.DelegationOutput
			err = s.precompile.UnpackIntoInterface(&delOut, staking.DelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the delegation output: %v", err)
			Expect(delOut.Balance.Amount.Int64()).To(Equal(int64(0)), "expected a different delegation balance")
			Expect(delOut.Balance.Denom).To(Equal(s.bondDenom), "expected a different delegation balance")
		})

		It("which exists should return the delegation", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				delegator.Addr, valAddr.String(),
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var delOut staking.DelegationOutput
			err = s.precompile.UnpackIntoInterface(&delOut, staking.DelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the delegation output: %v", err)
			Expect(delOut.Balance).To(Equal(
				cmn.Coin{Denom: s.bondDenom, Amount: big.NewInt(1e18)}),
				"expected a different delegation balance",
			)
		})
	})

	Context("querying redelegation", func() {
		BeforeEach(func() {
			callArgs.MethodName = "getRedelegation"
		})

		It("which does not exist should return an empty redelegation", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				delegator.Addr, valAddr.String(), nonExistingVal.String(),
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var redOut staking.RedelegationOutput
			err = s.precompile.UnpackIntoInterface(&redOut, staking.RedelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the redelegation output: %v", err)
			Expect(redOut.Redelegation.Entries).To(HaveLen(0), "expected no redelegation entries")
		})

		It("which exists should return the redelegation", func() {
			delegator := s.keyring.GetKey(0)

			// set up approval
			approveCallArgs.Args = []interface{}{
				contractAddr, []string{staking.RedelegateMsg}, big.NewInt(1e18),
			}
			s.SetupApprovalWithContractCalls(delegator, txArgs, approveCallArgs)

			Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

			// set up redelegation
			redelegateArgs := factory.CallArgs{
				ContractABI: stakingCallerContract.ABI,
				MethodName:  "testRedelegate",
				Args:        []interface{}{delegator.Addr, valAddr.String(), valAddr2.String(), big.NewInt(1)},
			}

			redelegateCheck := passCheck.
				WithExpEvents(staking.EventTypeRedelegate)

			_, _, err = s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, redelegateArgs,
				redelegateCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			// check that the redelegation was created
			res, err := s.grpcHandler.GetRedelegations(delegator.AccAddr.String(), valAddr.String(), valAddr2.String())
			Expect(err).To(BeNil())
			Expect(res.RedelegationResponses).To(HaveLen(1), "expected one redelegation to be found")
			bech32Addr := delegator.AccAddr
			Expect(res.RedelegationResponses[0].Redelegation.DelegatorAddress).To(Equal(bech32Addr.String()), "expected delegator address to be %s", delegator.Addr)
			Expect(res.RedelegationResponses[0].Redelegation.ValidatorSrcAddress).To(Equal(valAddr.String()), "expected source validator address to be %s", valAddr)
			Expect(res.RedelegationResponses[0].Redelegation.ValidatorDstAddress).To(Equal(valAddr2.String()), "expected destination validator address to be %s", valAddr2)

			// query redelegation
			callArgs.Args = []interface{}{
				delegator.Addr, valAddr.String(), valAddr2.String(),
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var redOut staking.RedelegationOutput
			err = s.precompile.UnpackIntoInterface(&redOut, staking.RedelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the redelegation output: %v", err)
			Expect(redOut.Redelegation.Entries).To(HaveLen(1), "expected one redelegation entry to be returned")
		})
	})

	Describe("query redelegations", func() {
		BeforeEach(func() {
			callArgs.MethodName = "getRedelegations"
		})
		It("which exists should return all the existing redelegations w/pagination", func() {
			delegator := s.keyring.GetKey(0)

			// set up approval
			approveCallArgs.Args = []interface{}{
				contractAddr, []string{staking.RedelegateMsg}, big.NewInt(1e18),
			}
			s.SetupApprovalWithContractCalls(delegator, txArgs, approveCallArgs)
			Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

			// set up redelegation
			redelegateArgs := factory.CallArgs{
				ContractABI: stakingCallerContract.ABI,
				MethodName:  "testRedelegate",
				Args:        []interface{}{delegator.Addr, valAddr.String(), valAddr2.String(), big.NewInt(1)},
			}

			redelegateCheck := passCheck.
				WithExpEvents(staking.EventTypeRedelegate)
			_, _, err = s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, redelegateArgs,
				redelegateCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			// check that the redelegation was created
			res, err := s.grpcHandler.GetRedelegations(delegator.AccAddr.String(), valAddr.String(), valAddr2.String())
			Expect(err).To(BeNil())
			Expect(res.RedelegationResponses).To(HaveLen(1), "expected one redelegation to be found")
			bech32Addr := delegator.AccAddr
			Expect(res.RedelegationResponses[0].Redelegation.DelegatorAddress).To(Equal(bech32Addr.String()), "expected delegator address to be %s", delegator.Addr)
			Expect(res.RedelegationResponses[0].Redelegation.ValidatorSrcAddress).To(Equal(valAddr.String()), "expected source validator address to be %s", valAddr)
			Expect(res.RedelegationResponses[0].Redelegation.ValidatorDstAddress).To(Equal(valAddr2.String()), "expected destination validator address to be %s", valAddr2)

			// query redelegations by delegator address
			callArgs.Args = []interface{}{
				delegator.Addr, "", "", query.PageRequest{Limit: 1, CountTotal: true},
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

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
		BeforeEach(func() {
			delegator := s.keyring.GetKey(0)

			callArgs.MethodName = "getUnbondingDelegation"
			// Set up an unbonding delegation
			approveCallArgs.Args = []interface{}{
				contractAddr, []string{staking.UndelegateMsg}, big.NewInt(1e18),
			}

			s.SetupApprovalWithContractCalls(delegator, txArgs, approveCallArgs)

			Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

			undelegateArgs := factory.CallArgs{
				ContractABI: stakingCallerContract.ABI,
				MethodName:  "testUndelegate",
				Args:        []interface{}{delegator.Addr, valAddr.String(), big.NewInt(1e18)},
			}

			logCheckArgs := passCheck.
				WithExpEvents(staking.EventTypeUnbond)

			_, _, err = s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, undelegateArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while setting up an unbonding delegation: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			// Check that the unbonding delegation was created
			res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
			Expect(err).To(BeNil())
			Expect(res.UnbondingResponses).To(HaveLen(1), "expected one unbonding delegation to be found")
			Expect(res.UnbondingResponses[0].DelegatorAddress).To(Equal(delegator.AccAddr.String()), "expected delegator address to be %s", delegator.Addr)
			Expect(res.UnbondingResponses[0].ValidatorAddress).To(Equal(valAddr.String()), "expected validator address to be %s", valAddr)
			Expect(res.UnbondingResponses[0].Entries).To(HaveLen(1), "expected one unbonding delegation entry to be found")
			Expect(res.UnbondingResponses[0].Entries[0].CreationHeight).To(Equal(s.network.GetContext().BlockHeight()), "expected different creation height")
			Expect(res.UnbondingResponses[0].Entries[0].Balance).To(Equal(math.NewInt(1e18)), "expected different balance")
		})

		It("which does not exist should return an empty unbonding delegation", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				delegator.Addr, valAddr2.String(),
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var unbondingDelegationOutput staking.UnbondingDelegationOutput
			err = s.precompile.UnpackIntoInterface(&unbondingDelegationOutput, staking.UnbondingDelegationMethod, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the unbonding delegation output: %v", err)
			Expect(unbondingDelegationOutput.UnbondingDelegation.Entries).To(HaveLen(0), "expected one unbonding delegation entry")
		})

		It("which exists should return the unbonding delegation", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.Args = []interface{}{
				delegator.Addr, valAddr.String(),
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs, passCheck)
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
			delegator := s.keyring.GetKey(0)

			cArgs := factory.CallArgs{
				ContractABI: stakingCallerContract.ABI,
				MethodName:  "testApproveAndThenUndelegate",
				Args:        []interface{}{contractAddr, big.NewInt(250), big.NewInt(500), valAddr.String()},
			}
			txArgs.GasLimit = 1e8

			_, _, err = s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, cArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			// There should be no authorizations because everything should have been reverted
			authz, _, err := CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, staking.UndelegateAuthz, contractAddr, delegator.Addr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("no authorizations found for grantee %s and granter %s", contractAddr.Hex(), delegator.Addr.Hex())))
			Expect(authz).To(BeNil(), "expected authorization to be nil")

			res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
			Expect(err).To(BeNil())
			Expect(res.UnbondingResponses).To(HaveLen(0), "expected no unbonding delegations")
		})

		It("should write to state if all operations succeed", func() {
			delegator := s.keyring.GetKey(0)

			cArgs := factory.CallArgs{
				ContractABI: stakingCallerContract.ABI,
				MethodName:  "testApproveAndThenUndelegate",
				Args:        []interface{}{contractAddr, big.NewInt(1000), big.NewInt(500), valAddr.String()},
			}
			txArgs.GasLimit = 1e8

			logCheckArgs := passCheck.
				WithExpEvents(authorization.EventTypeApproval, staking.EventTypeUnbond)

			_, _, err := s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, cArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			authz, _, err := CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, staking.UndelegateAuthz, contractAddr, delegator.Addr)
			Expect(err).To(BeNil())
			Expect(authz).ToNot(BeNil(), "expected authorization not to be nil")

			res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
			Expect(err).To(BeNil())
			Expect(res.UnbondingResponses).To(HaveLen(1), "expected one unbonding delegation")
			Expect(res.UnbondingResponses[0].ValidatorAddress).To(Equal(valAddr.String()), "expected different validator address")
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
			granter := s.keyring.GetKey(0)

			// approve undelegate message
			approveCallArgs.Args = []interface{}{
				contractAddr, []string{staking.UndelegateMsg}, big.NewInt(1e18),
			}

			s.SetupApprovalWithContractCalls(granter, txArgs, approveCallArgs)

			Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")
		})

		for _, tc := range testcases {
			// NOTE: this is necessary because of Ginkgo behavior -- if not done, the value of tc
			// inside the It block will always be the last entry in the testcases slice
			testcase := tc

			It(fmt.Sprintf("should not execute transactions for calltype %q", testcase.calltype), func() {
				delegator := s.keyring.GetKey(0)

				callArgs.MethodName = "testCallUndelegate"
				callArgs.Args = []interface{}{
					delegator.Addr, valAddr.String(), big.NewInt(1e18), testcase.calltype,
				}

				checkArgs := execRevertedCheck
				if testcase.expTxPass {
					checkArgs = passCheck.WithExpEvents(staking.EventTypeUnbond)
				}

				_, _, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs,
					checkArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract for calltype %s: %v", testcase.calltype, err)
				Expect(s.network.NextBlock()).To(BeNil())

				// check no delegations are unbonding
				res, err := s.grpcHandler.GetDelegatorUnbondingDelegations(delegator.AccAddr.String())
				Expect(err).To(BeNil())

				if testcase.expTxPass {
					Expect(res.UnbondingResponses).To(HaveLen(1), "expected an unbonding delegation")
					Expect(res.UnbondingResponses[0].ValidatorAddress).To(Equal(valAddr.String()), "expected different validator address")
					Expect(res.UnbondingResponses[0].DelegatorAddress).To(Equal(delegator.AccAddr.String()), "expected different delegator address")
				} else {
					Expect(res.UnbondingResponses).To(HaveLen(0), "expected no unbonding delegations for calltype %s", testcase.calltype)
				}
			})

			It(fmt.Sprintf("should execute queries for calltype %q", testcase.calltype), func() {
				delegator := s.keyring.GetKey(0)

				callArgs.MethodName = "testCallDelegation"
				callArgs.Args = []interface{}{delegator.Addr, valAddr.String(), testcase.calltype}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					delegator.Priv,
					txArgs, callArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil())

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
			resBal, err := s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil(), "error while getting balance")

			balanceBefore := resBal.Balance
			Expect(balanceBefore.Amount.Int64()).To(BeZero(), "expected contract balance to be 0 before funding")

			err = testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), contractAddr.Bytes(), math.NewIntFromBigInt(delegationAmount))
			Expect(err).To(BeNil(), "error while funding account")
			Expect(s.network.NextBlock()).To(BeNil())

			resBal, err = s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil(), "error while getting balance")

			balanceAfterFunding := resBal.Balance
			Expect(balanceAfterFunding.Amount.BigInt()).To(Equal(delegationAmount), "expected different contract balance after funding")

			// Check no delegation exists from the contract to the validator
			res, err := s.grpcHandler.GetDelegation(sdk.AccAddress(contractAddr.Bytes()).String(), valAddr.String())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("delegation with delegator %s not found for validator %s", sdk.AccAddress(contractAddr.Bytes()), valAddr)))
			Expect(res).To(BeNil())
		})

		It("delegating and increasing counter should change the bank balance accordingly", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.MethodName = "testDelegateIncrementCounter"
			callArgs.Args = []interface{}{valAddr.String(), delegationAmount}
			txArgs.GasLimit = 1e9

			approvalAndDelegationCheck := passCheck.WithExpEvents(
				authorization.EventTypeApproval, staking.EventTypeDelegate,
			)

			_, _, err = s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				approvalAndDelegationCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			res, err := s.grpcHandler.GetDelegation(sdk.AccAddress(contractAddr.Bytes()).String(), valAddr.String())
			Expect(err).To(BeNil())
			Expect(res.DelegationResponse).NotTo(BeNil())
			Expect(res.DelegationResponse.Delegation.GetShares().BigInt()).To(Equal(delegationAmount), "expected different delegation shares")

			resBal, err := s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil(), "error while getting balance")

			postBalance := resBal.Balance
			Expect(postBalance.Amount.Int64()).To(BeZero(), "expected balance to be 0 after contract call")
		})
	})

	Context("when updating the stateDB prior to calling the precompile", func() {
		It("should utilize the same contract balance to delegate", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.MethodName = "approveDepositAndDelegate"
			callArgs.Args = []interface{}{valAddr.String()}

			txArgs.Amount = big.NewInt(2e18)
			txArgs.GasLimit = 1e9

			approvalAndDelegationCheck := passCheck.WithExpEvents(
				authorization.EventTypeApproval, staking.EventTypeDelegate,
			)
			_, _, err = s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				approvalAndDelegationCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			resBal, err := s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil(), "error while getting balance")
			balance := resBal.Balance

			Expect(balance.Amount.Int64()).To(BeZero(), "expected different contract balance after funding")
			res, err := s.grpcHandler.GetDelegatorDelegations(sdk.AccAddress(contractAddr.Bytes()).String())
			Expect(err).To(BeNil())
			Expect(res.DelegationResponses).To(HaveLen(1), "expected one delegation")
			Expect(res.DelegationResponses[0].Delegation.GetShares().BigInt()).To(Equal(big.NewInt(2e18)), "expected different delegation shares")
		})
		//nolint:dupl
		It("should revert the contract balance to the original value when the precompile fails", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.MethodName = "approveDepositAndDelegateExceedingAllowance"
			callArgs.Args = []interface{}{valAddr.String()}

			txArgs.Amount = big.NewInt(2e18)
			txArgs.GasLimit = 1e9

			approvalAndDelegationCheck := defaultLogCheck.WithErrContains(vm.ErrExecutionReverted.Error())
			_, _, err = s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				approvalAndDelegationCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			resBal, err := s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil(), "error while getting balance")
			balance := resBal.Balance

			Expect(balance.Amount.Int64()).To(BeZero(), "expected different contract balance after funding")
			auth, _ := s.network.App.AuthzKeeper.GetAuthorization(s.network.GetContext(), contractAddr.Bytes(), delegator.AccAddr, staking.DelegateMsg)
			Expect(auth).To(BeNil(), "expected no authorization")
			res, err := s.grpcHandler.GetDelegatorDelegations(sdk.AccAddress(contractAddr.Bytes()).String())
			Expect(err).To(BeNil())
			Expect(res.DelegationResponses).To(HaveLen(0), "expected no delegations")
		})

		//nolint:dupl
		It("should revert the contract balance to the original value when the custom logic after the precompile fails ", func() {
			delegator := s.keyring.GetKey(0)

			callArgs.MethodName = "approveDepositDelegateAndFailCustomLogic"
			callArgs.Args = []interface{}{valAddr.String()}

			txArgs.Amount = big.NewInt(2e18)
			txArgs.GasLimit = 1e9

			approvalAndDelegationCheck := defaultLogCheck.WithErrContains(vm.ErrExecutionReverted.Error())
			_, _, err = s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				approvalAndDelegationCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			resBal, err := s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil(), "error while getting balance")

			balance := resBal.Balance
			Expect(balance.Amount.Int64()).To(BeZero(), "expected different contract balance after funding")
			auth, _ := s.network.App.AuthzKeeper.GetAuthorization(s.network.GetContext(), contractAddr.Bytes(), delegator.AccAddr, staking.DelegateMsg)
			Expect(auth).To(BeNil(), "expected no authorization")
			res, err := s.grpcHandler.GetDelegatorDelegations(sdk.AccAddress(contractAddr.Bytes()).String())
			Expect(err).To(BeNil())
			Expect(res.DelegationResponses).To(HaveLen(0), "expected no delegations")
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
		s *PrecompileTestSuite
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
		// s is the precompile test suite to use for the tests
	)

	BeforeEach(func() {
		s = new(PrecompileTestSuite)
		s.SetupTest()
		delegator := s.keyring.GetKey(0)

		stakingCallerContract, err = testdata.LoadStakingCallerContract()
		Expect(err).To(BeNil(), "error while loading the StakingCaller contract")

		// Deploy StakingCaller contract
		contractAddr, err = s.factory.DeployContract(
			delegator.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: stakingCallerContract,
			},
		)
		Expect(err).To(BeNil(), "error while deploying the StakingCaller contract")
		Expect(s.network.NextBlock()).To(BeNil())

		// Deploy ERC20 contract
		erc20ContractAddr, err = s.factory.DeployContract(
			delegator.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract:        erc20Contract,
				ConstructorArgs: []interface{}{erc20Name, erc20Token, erc20Decimals},
			},
		)
		Expect(err).To(BeNil(), "error while deploying the ERC20 contract")
		Expect(s.network.NextBlock()).To(BeNil())

		// Mint tokens to the StakingCaller contract
		mintArgs := factory.CallArgs{
			ContractABI: erc20Contract.ABI,
			MethodName:  "mint",
			Args:        []interface{}{contractAddr, mintAmount},
		}

		txArgs = evmtypes.EvmTxArgs{
			To: &erc20ContractAddr,
		}

		mintCheck := testutil.LogCheckArgs{
			ABIEvents: erc20Contract.ABI.Events,
			ExpEvents: []string{"Transfer"}, // minting produces a Transfer event
			ExpPass:   true,
		}

		_, _, err = s.factory.CallContractAndCheckLogs(
			delegator.Priv,
			txArgs, mintArgs, mintCheck)
		Expect(err).To(BeNil(), "error while minting tokens to the StakingCaller contract")
		Expect(s.network.NextBlock()).To(BeNil())

		// Check that the StakingCaller contract has the correct balance
		erc20Balance := s.network.App.Erc20Keeper.BalanceOf(s.network.GetContext(), erc20Contract.ABI, erc20ContractAddr, contractAddr)
		Expect(erc20Balance).To(Equal(mintAmount), "expected different ERC20 balance for the StakingCaller contract")

		// populate default call args
		callArgs = factory.CallArgs{
			ContractABI: stakingCallerContract.ABI,
			MethodName:  "callERC20AndDelegate",
		}

		txArgs.To = &contractAddr

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
			delegator := s.keyring.GetKey(0)

			res, err := s.grpcHandler.GetDelegatorDelegations(delegator.AccAddr.String())
			Expect(err).To(BeNil())
			Expect(res.DelegationResponses).ToNot(HaveLen(0), "expected address to have delegations")

			validator, err = sdk.ValAddressFromBech32(res.DelegationResponses[0].Delegation.ValidatorAddress)
			Expect(err).To(BeNil())

			_ = erc20ContractAddr
		})

		It("should revert both states if a staking transaction fails", func() {
			delegator := s.keyring.GetKey(0)

			res, err := s.grpcHandler.GetDelegation(delegator.AccAddr.String(), validator.String())
			Expect(err).To(BeNil())
			Expect(res.DelegationResponse).NotTo(BeNil())

			delegationPre := res.DelegationResponse.Delegation
			sharesPre := delegationPre.GetShares()

			// NOTE: passing an invalid validator address here should fail AFTER the erc20 transfer was made in the smart contract.
			// Therefore this can be used to check that both EVM and Cosmos states are reverted correctly.
			callArgs.Args = []interface{}{erc20ContractAddr, "invalid validator", transferredAmount}

			_, _, err = s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				execRevertedCheck)
			Expect(err).To(BeNil(), "expected error while calling the smart contract")
			Expect(s.network.NextBlock()).To(BeNil())

			res, err = s.grpcHandler.GetDelegation(delegator.AccAddr.String(), validator.String())
			Expect(err).To(BeNil())
			Expect(res.DelegationResponse).NotTo(BeNil())
			delegationPost := res.DelegationResponse.Delegation

			auths, err := s.grpcHandler.GetAuthorizations(sdk.AccAddress(contractAddr.Bytes()).String(), delegator.AccAddr.String())
			Expect(err).To(BeNil(), "error while getting authorizations: %v", err)
			sharesPost := delegationPost.GetShares()
			erc20BalancePost := s.network.App.Erc20Keeper.BalanceOf(s.network.GetContext(), erc20Contract.ABI, erc20ContractAddr, delegator.Addr)

			Expect(auths).To(BeEmpty(), "expected no authorizations when reverting state")
			Expect(sharesPost).To(Equal(sharesPre), "expected shares to be equal when reverting state")
			Expect(erc20BalancePost.Int64()).To(BeZero(), "expected erc20 balance of target address to be zero when reverting state")
		})

		It("should revert both states if an ERC20 transaction fails", func() {
			delegator := s.keyring.GetKey(0)

			res, err := s.grpcHandler.GetDelegation(delegator.AccAddr.String(), validator.String())
			Expect(err).To(BeNil())
			Expect(res.DelegationResponse).NotTo(BeNil())

			delegationPre := res.DelegationResponse.Delegation
			sharesPre := delegationPre.GetShares()

			// NOTE: trying to transfer more than the balance of the contract should fail AFTER the approval
			// for delegating was made in the smart contract.
			// Therefore this can be used to check that both EVM and Cosmos states are reverted correctly.
			moreThanMintedAmount := new(big.Int).Add(mintAmount, big.NewInt(1))
			callArgs.Args = []interface{}{erc20ContractAddr, s.network.GetValidators()[0].OperatorAddress, moreThanMintedAmount}

			_, _, err = s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				execRevertedCheck)
			Expect(err).To(BeNil(), "expected error while calling the smart contract")
			Expect(s.network.NextBlock()).To(BeNil())

			res, err = s.grpcHandler.GetDelegation(delegator.AccAddr.String(), validator.String())
			Expect(err).To(BeNil())
			Expect(res.DelegationResponse).NotTo(BeNil())
			delegationPost := res.DelegationResponse.Delegation

			auths, err := s.grpcHandler.GetAuthorizations(sdk.AccAddress(contractAddr.Bytes()).String(), delegator.AccAddr.String())
			Expect(err).To(BeNil(), "error while getting authorizations: %v", err)
			sharesPost := delegationPost.GetShares()
			erc20BalancePost := s.network.App.Erc20Keeper.BalanceOf(s.network.GetContext(), erc20Contract.ABI, erc20ContractAddr, delegator.Addr)

			Expect(auths).To(BeEmpty(), "expected no authorizations when reverting state")
			Expect(sharesPost).To(Equal(sharesPre), "expected shares to be equal when reverting state")
			Expect(erc20BalancePost.Int64()).To(BeZero(), "expected erc20 balance of target address to be zero when reverting state")
		})

		It("should persist changes in both the cosmos and eth states", func() {
			delegator := s.keyring.GetKey(0)

			res, err := s.grpcHandler.GetDelegation(delegator.AccAddr.String(), validator.String())
			Expect(err).To(BeNil())
			Expect(res.DelegationResponse).NotTo(BeNil())

			delegationPre := res.DelegationResponse.Delegation
			sharesPre := delegationPre.GetShares()

			// NOTE: trying to transfer more than the balance of the contract should fail AFTER the approval
			// for delegating was made in the smart contract.
			// Therefore this can be used to check that both EVM and Cosmos states are reverted correctly.
			callArgs.Args = []interface{}{erc20ContractAddr, s.network.GetValidators()[0].OperatorAddress, transferredAmount}

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

			_, _, err = s.factory.CallContractAndCheckLogs(
				delegator.Priv,
				txArgs, callArgs,
				successCheck)
			Expect(err).ToNot(HaveOccurred(), "error while calling the smart contract")
			Expect(s.network.NextBlock()).To(BeNil())

			res, err = s.grpcHandler.GetDelegation(delegator.AccAddr.String(), validator.String())
			Expect(err).To(BeNil())
			Expect(res.DelegationResponse).NotTo(BeNil(),
				"expected delegation from %s to validator %s to be found after calling the smart contract",
				delegator.AccAddr.String(), validator.String(),
			)
			delegationPost := res.DelegationResponse.Delegation

			auths, err := s.grpcHandler.GetAuthorizations(sdk.AccAddress(contractAddr.Bytes()).String(), delegator.AccAddr.String())
			Expect(err).To(BeNil(), "error while getting authorizations: %v", err)
			sharesPost := delegationPost.GetShares()
			erc20BalancePost := s.network.App.Erc20Keeper.BalanceOf(s.network.GetContext(), erc20Contract.ABI, erc20ContractAddr, delegator.Addr)

			Expect(sharesPost.GT(sharesPre)).To(BeTrue(), "expected shares to be more than before")
			Expect(erc20BalancePost).To(Equal(transferredAmount), "expected different erc20 balance of target address")
			// NOTE: there should be no authorizations because the full approved amount is delegated
			Expect(auths).To(HaveLen(0), "expected no authorization to be found")
		})
	})
})
