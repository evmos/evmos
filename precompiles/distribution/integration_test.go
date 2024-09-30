// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package distribution_test

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v20/precompiles/authorization"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/precompiles/distribution"
	"github.com/evmos/evmos/v20/precompiles/staking"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	"github.com/evmos/evmos/v20/precompiles/testutil/contracts"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	testutils "github.com/evmos/evmos/v20/testutil/integration/evmos/utils"
	testutiltx "github.com/evmos/evmos/v20/testutil/tx"
	evmostypes "github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

// General variables used for integration tests
var (
	// differentAddr is an address generated for testing purposes that e.g. raises the different origin error
	differentAddr, diffKey = testutiltx.NewAddrKey()
	// gasPrice is the gas price used for the transactions
	gasPrice = math.NewInt(1e9)
	// callArgs  are the default arguments for calling the smart contract
	//
	// NOTE: this has to be populated in a BeforeEach block because the contractAddr would otherwise be a nil address.
	callArgs factory.CallArgs

	// defaultLogCheck instantiates a log check arguments struct with the precompile ABI events populated.
	defaultLogCheck testutil.LogCheckArgs
	// passCheck defines the arguments to check if the precompile returns no error
	passCheck testutil.LogCheckArgs
	// outOfGasCheck defines the arguments to check if the precompile returns out of gas error
	outOfGasCheck testutil.LogCheckArgs
	// txArgs are the EVM transaction arguments to use in the transactions
	txArgs evmtypes.EvmTxArgs
	// minExpRewardOrCommission is the minimun coins expected for validator's rewards or commission
	// required for the tests
	minExpRewardOrCommission = sdk.NewDecCoins(sdk.NewDecCoin(evmostypes.BaseDenom, testRewardsAmt))
)

func TestPrecompileIntegrationTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Distribution Precompile Suite")
}

var _ = Describe("Calling distribution precompile from EOA", func() {
	s := new(PrecompileTestSuite)

	BeforeEach(func() {
		s.SetupTest()

		// set the default call arguments
		callArgs = factory.CallArgs{
			ContractABI: s.precompile.ABI,
		}

		defaultLogCheck = testutil.LogCheckArgs{
			ABIEvents: s.precompile.Events,
		}
		passCheck = defaultLogCheck.WithExpPass(true)
		outOfGasCheck = defaultLogCheck.WithErrContains(vm.ErrOutOfGas.Error())

		// reset tx args each test to avoid keeping custom
		// values of previous tests (e.g. gasLimit)
		precompileAddr := s.precompile.Address()
		txArgs = evmtypes.EvmTxArgs{
			To: &precompileAddr,
		}
	})

	// =====================================
	// 				TRANSACTIONS
	// =====================================
	Describe("Execute SetWithdrawAddress transaction", func() {
		const method = distribution.SetWithdrawAddressMethod

		BeforeEach(func() {
			// set the default call arguments
			callArgs.MethodName = method
		})

		It("should return error if the provided gasLimit is too low", func() {
			txArgs.GasLimit = 30000

			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0),
				differentAddr.String(),
			}
			_, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				outOfGasCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock")

			// withdraw address should remain unchanged
			delAddr := s.keyring.GetAccAddr(0).String()
			res, err := s.grpcHandler.GetDelegatorWithdrawAddr(delAddr)
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(res.WithdrawAddress).To(Equal(delAddr), "expected withdraw address to remain unchanged")
		})

		It("should return error if the origin is different than the delegator", func() {
			callArgs.Args = []interface{}{
				differentAddr,
				s.keyring.GetAddr(0).String(),
			}

			withdrawAddrSetCheck := defaultLogCheck.WithErrContains(cmn.ErrDelegatorDifferentOrigin, s.keyring.GetAddr(0).String(), differentAddr.String())

			_, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				withdrawAddrSetCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")
		})

		It("should set withdraw address", func() {
			// initially, withdraw address should be same as address
			res, err := s.grpcHandler.GetDelegatorWithdrawAddr(s.keyring.GetAccAddr(0).String())
			Expect(err).To(BeNil(), "error while querying withdraw address")
			Expect(res.WithdrawAddress).To(Equal(s.keyring.GetAccAddr(0).String()))

			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0),
				differentAddr.String(),
			}

			withdrawAddrSetCheck := passCheck.
				WithExpEvents(distribution.EventTypeSetWithdrawAddress)

			_, _, err = s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				withdrawAddrSetCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")

			// persist state changes
			Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock")

			// withdraw should be updated
			res, err = s.grpcHandler.GetDelegatorWithdrawAddr(s.keyring.GetAccAddr(0).String())
			Expect(err).To(BeNil(), "error while querying withdraw address")
			Expect(res.WithdrawAddress).To(Equal(sdk.AccAddress(differentAddr.Bytes()).String()), "expected different withdraw address")
		})
	})

	Describe("Execute WithdrawDelegatorRewards transaction", func() {
		var accruedRewards sdk.DecCoins
		BeforeEach(func() {
			var err error
			// set the default call arguments
			callArgs.MethodName = distribution.WithdrawDelegatorRewardsMethod

			accruedRewards, err = testutils.WaitToAccrueRewards(s.network, s.grpcHandler, s.keyring.GetAccAddr(0).String(), minExpRewardOrCommission)
			Expect(err).To(BeNil())
		})

		It("should return error if the origin is different than the delegator", func() {
			callArgs.Args = []interface{}{
				differentAddr,
				s.network.GetValidators()[0].OperatorAddress,
			}

			withdrawalCheck := defaultLogCheck.WithErrContains(cmn.ErrDelegatorDifferentOrigin, s.keyring.GetAddr(0).String(), differentAddr.String())

			_, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				withdrawalCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")
		})

		It("should withdraw delegation rewards", func() {
			// get initial balance
			queryRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil(), "error while calling GetBalance")
			initialBalance := queryRes.Balance

			txArgs.GasPrice = gasPrice.BigInt()
			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0),
				s.network.GetValidators()[0].OperatorAddress,
			}

			withdrawalCheck := passCheck.
				WithExpEvents(distribution.EventTypeWithdrawDelegatorRewards)

			res, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				withdrawalCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock")

			var rewards []cmn.Coin
			err = s.precompile.UnpackIntoInterface(&rewards, distribution.WithdrawDelegatorRewardsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(rewards)).To(Equal(1))

			// The accrued rewards are based on 3 equal delegations to the existing 3 validators
			// The query is from only 1 validator, thus, the expected reward
			// for this delegation is totalAccruedRewards / validatorsCount (3)
			valCount := len(s.network.GetValidators())
			accruedRewardsAmt := accruedRewards.AmountOf(s.bondDenom)
			expRewardPerValidator := accruedRewardsAmt.Quo(math.LegacyNewDec(int64(valCount)))

			Expect(rewards[0].Denom).To(Equal(s.bondDenom))
			Expect(rewards[0].Amount).To(Equal(expRewardPerValidator.TruncateInt().BigInt()))

			// check that the rewards were added to the balance
			queryRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil(), "error while calling GetBalance")
			fees := gasPrice.Mul(math.NewInt(res.GasUsed))
			expFinal := initialBalance.Amount.Add(expRewardPerValidator.TruncateInt()).Sub(fees)
			Expect(queryRes.Balance.Amount).To(Equal(expFinal), "expected final balance to be equal to initial balance + rewards - fees")
		})

		It("should withdraw rewards successfully to the new withdrawer address", func() {
			balRes, err := s.grpcHandler.GetBalance(differentAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			withdrawerInitialBalance := balRes.Balance
			// Set new withdrawer address
			err = s.factory.SetWithdrawAddress(s.keyring.GetPrivKey(0), differentAddr.Bytes())
			Expect(err).To(BeNil())
			// persist state change
			Expect(s.network.NextBlock()).To(BeNil())

			// get initial balance
			queryRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil(), "error while calling GetBalance")
			initialBalance := queryRes.Balance

			// get rewards
			rwRes, err := s.grpcHandler.GetDelegationRewards(s.keyring.GetAccAddr(0).String(), s.network.GetValidators()[0].OperatorAddress)
			Expect(err).To(BeNil())
			expRewardsAmt := rwRes.Rewards.AmountOf(s.bondDenom).TruncateInt()

			txArgs.GasPrice = gasPrice.BigInt()
			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0),
				s.network.GetValidators()[0].OperatorAddress,
			}

			withdrawalCheck := passCheck.
				WithExpEvents(distribution.EventTypeWithdrawDelegatorRewards)

			txArgs.GasLimit = 300_000
			res, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				withdrawalCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock")

			var rewards []cmn.Coin
			err = s.precompile.UnpackIntoInterface(&rewards, distribution.WithdrawDelegatorRewardsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(rewards)).To(Equal(1))

			Expect(rewards[0].Denom).To(Equal(s.bondDenom))
			Expect(rewards[0].Amount).To(Equal(expRewardsAmt.BigInt()))

			// check that the delegator final balance is initialBalance - fee
			queryRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil(), "error while calling GetBalance")
			fees := gasPrice.Mul(math.NewInt(res.GasUsed))
			expDelgatorFinal := initialBalance.Amount.Sub(fees)
			Expect(queryRes.Balance.Amount).To(Equal(expDelgatorFinal), "expected delegator final balance to be equal to initial balance - fees")

			// check that the rewards were added to the withdrawer balance
			queryRes, err = s.grpcHandler.GetBalance(differentAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil(), "error while calling GetBalance")
			expWithdrawerFinal := withdrawerInitialBalance.Amount.Add(expRewardsAmt)

			Expect(queryRes.Balance.Amount).To(Equal(expWithdrawerFinal), "expected withdrawer final balance to be equal to initial balance + rewards")
		})

		It("should withdraw delegation rewards to a smart contract", func() {
			// deploy a smart contract to use as withdrawer
			distributionCallerContract, err := contracts.LoadDistributionCallerContract()
			Expect(err).To(BeNil(), "error while loading the smart contract: %v", err)

			contractAddr, err := s.factory.DeployContract(
				s.keyring.GetPrivKey(0),
				evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
				factory.ContractDeploymentData{
					Contract: distributionCallerContract,
				},
			)
			Expect(err).To(BeNil(), "error while deploying the smart contract: %v", err)
			// persist state change
			Expect(s.network.NextBlock()).To(BeNil())

			balRes, err := s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			initialWithdrawerBalance := balRes.Balance
			Expect(initialWithdrawerBalance.Amount).To(Equal(math.ZeroInt()))

			// set contract address as withdrawer address
			err = s.factory.SetWithdrawAddress(s.keyring.GetPrivKey(0), contractAddr.Bytes())
			Expect(err).To(BeNil())
			// persist state change
			Expect(s.network.NextBlock()).To(BeNil())

			// get tx sender initial balance
			balRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil(), "error while calling GetBalance")
			initialBalance := balRes.Balance

			// get rewards
			rwRes, err := s.grpcHandler.GetDelegationRewards(s.keyring.GetAccAddr(0).String(), s.network.GetValidators()[0].OperatorAddress)
			Expect(err).To(BeNil())
			expRewardsAmt := rwRes.Rewards.AmountOf(s.bondDenom).TruncateInt()

			txArgs.GasPrice = gasPrice.BigInt()
			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0),
				s.network.GetValidators()[0].OperatorAddress,
			}

			withdrawalCheck := passCheck.
				WithExpEvents(distribution.EventTypeWithdrawDelegatorRewards)

			txArgs.GasLimit = 300_000
			res, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				withdrawalCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock")

			var rewards []cmn.Coin
			err = s.precompile.UnpackIntoInterface(&rewards, distribution.WithdrawDelegatorRewardsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(rewards)).To(Equal(1))
			Expect(rewards[0].Denom).To(Equal(s.bondDenom))
			Expect(rewards[0].Amount).To(Equal(expRewardsAmt.BigInt()))

			// check tx sender balance is reduced by fees paid
			balRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			finalBalance := balRes.Balance
			fees := gasPrice.MulRaw(res.GasUsed)
			expFinal := initialBalance.Amount.Sub(fees)
			Expect(finalBalance.Amount).To(Equal(expFinal), "expected final balance to be equal to initial balance - fees")

			// check that the rewards were added to the withdrawer balance
			balRes, err = s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			finalWithdrawerBalance := balRes.Balance
			Expect(finalWithdrawerBalance.Amount).To(Equal(expRewardsAmt))
		})
	})

	Describe("Validator Commission: Execute WithdrawValidatorCommission tx", func() {
		// expCommAmt is the expected commission amount
		expCommAmt := math.NewInt(1)

		BeforeEach(func() {
			// set the default call arguments
			callArgs.MethodName = distribution.WithdrawValidatorCommissionMethod
			valAddr := sdk.ValAddress(s.validatorsKeys[0].AccAddr)

			_, err := testutils.WaitToAccrueCommission(
				s.network, s.grpcHandler,
				valAddr.String(),
				sdk.NewDecCoins(sdk.NewDecCoin(s.bondDenom, expCommAmt)),
			)
			Expect(err).To(BeNil())

			// Send some funds to the validator to pay for fees
			err = testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), s.validatorsKeys[0].AccAddr, math.NewInt(1e17))
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())
		})

		It("should return error if the provided gasLimit is too low", func() {
			txArgs.GasLimit = 50000
			callArgs.Args = []interface{}{
				s.network.GetValidators()[0].OperatorAddress,
			}

			_, _, err := s.factory.CallContractAndCheckLogs(
				s.validatorsKeys[0].Priv,
				txArgs,
				callArgs,
				outOfGasCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")
		})

		It("should return error if the origin is different than the validator", func() {
			callArgs.Args = []interface{}{
				s.network.GetValidators()[0].OperatorAddress,
			}

			validatorHexAddr := common.BytesToAddress(s.validatorsKeys[0].AccAddr)

			withdrawalCheck := defaultLogCheck.WithErrContains(cmn.ErrDelegatorDifferentOrigin, s.keyring.GetAddr(0).String(), validatorHexAddr.String())

			_, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				withdrawalCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")
		})

		It("should withdraw validator commission", func() {
			// initial balance should be the initial amount minus the staked amount used to create the validator
			queryRes, err := s.grpcHandler.GetBalance(s.validatorsKeys[0].AccAddr, s.bondDenom)
			Expect(err).To(BeNil(), "error while calling GetBalance")

			initialBalance := queryRes.Balance

			// get the accrued commission amount
			commRes, err := s.grpcHandler.GetValidatorCommission(s.network.GetValidators()[0].OperatorAddress)
			Expect(err).To(BeNil())
			expCommAmt := commRes.Commission.Commission.AmountOf(s.bondDenom).TruncateInt()

			callArgs.Args = []interface{}{s.network.GetValidators()[0].OperatorAddress}
			txArgs.GasPrice = gasPrice.BigInt()

			withdrawalCheck := passCheck.
				WithExpEvents(distribution.EventTypeWithdrawValidatorCommission)

			txArgs.GasLimit = 300_000
			res, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.validatorsKeys[0].Priv,
				txArgs,
				callArgs,
				withdrawalCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var comm []cmn.Coin
			err = s.precompile.UnpackIntoInterface(&comm, distribution.WithdrawValidatorCommissionMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(comm)).To(Equal(1))
			Expect(comm[0].Denom).To(Equal(s.bondDenom))
			Expect(comm[0].Amount).To(Equal(expCommAmt.BigInt()))

			Expect(s.network.NextBlock()).To(BeNil())

			queryRes, err = s.grpcHandler.GetBalance(s.validatorsKeys[0].AccAddr, s.bondDenom)
			Expect(err).To(BeNil(), "error while calling GetBalance")
			finalBalance := queryRes.Balance

			fees := gasPrice.Mul(math.NewInt(res.GasUsed))
			expFinal := initialBalance.Amount.Add(expCommAmt).Sub(fees)

			Expect(finalBalance.Amount).To(Equal(expFinal), "expected final balance to be equal to the final balance after withdrawing commission")
		})

		It("should withdraw validator commission to a smart contract", func() {
			// deploy a smart contract to use as withdrawer
			distributionCallerContract, err := contracts.LoadDistributionCallerContract()
			Expect(err).To(BeNil(), "error while loading the smart contract: %v", err)

			contractAddr, err := s.factory.DeployContract(
				s.keyring.GetPrivKey(0),
				evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
				factory.ContractDeploymentData{
					Contract: distributionCallerContract,
				},
			)
			Expect(err).To(BeNil(), "error while deploying the smart contract: %v", err)
			// persist state change
			Expect(s.network.NextBlock()).To(BeNil())

			balRes, err := s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			initialWithdrawerBalance := balRes.Balance
			Expect(initialWithdrawerBalance.Amount).To(Equal(math.ZeroInt()))

			// set contract address as withdrawer address
			err = s.factory.SetWithdrawAddress(s.validatorsKeys[0].Priv, contractAddr.Bytes())
			Expect(err).To(BeNil())
			// persist state change
			Expect(s.network.NextBlock()).To(BeNil())

			// get validator initial balance
			balRes, err = s.grpcHandler.GetBalance(s.validatorsKeys[0].AccAddr, s.bondDenom)
			Expect(err).To(BeNil(), "error while calling GetBalance")
			initialBalance := balRes.Balance

			// get the accrued commission amount
			commRes, err := s.grpcHandler.GetValidatorCommission(s.network.GetValidators()[0].OperatorAddress)
			Expect(err).To(BeNil())
			expCommAmt := commRes.Commission.Commission.AmountOf(s.bondDenom).TruncateInt()

			callArgs.Args = []interface{}{s.network.GetValidators()[0].OperatorAddress}
			txArgs.GasPrice = gasPrice.BigInt()

			withdrawalCheck := passCheck.
				WithExpEvents(distribution.EventTypeWithdrawValidatorCommission)

			txArgs.GasLimit = 300_000
			res, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.validatorsKeys[0].Priv,
				txArgs,
				callArgs,
				withdrawalCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")
			// persist state change
			Expect(s.network.NextBlock()).To(BeNil())

			var comm []cmn.Coin
			err = s.precompile.UnpackIntoInterface(&comm, distribution.WithdrawValidatorCommissionMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(comm)).To(Equal(1))
			Expect(comm[0].Denom).To(Equal(s.bondDenom))
			Expect(comm[0].Amount).To(Equal(expCommAmt.BigInt()))

			balRes, err = s.grpcHandler.GetBalance(s.validatorsKeys[0].AccAddr, s.bondDenom)
			Expect(err).To(BeNil(), "error while calling GetBalance")
			finalBalance := balRes.Balance

			fees := gasPrice.MulRaw(res.GasUsed)
			expFinal := initialBalance.Amount.Sub(fees)
			Expect(finalBalance.Amount).To(Equal(expFinal), "expected final balance to be equal to the final balance after withdrawing commission")

			// check that the commission was added to the withdrawer balance
			balRes, err = s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			finalWithdrawerBalance := balRes.Balance
			Expect(finalWithdrawerBalance.Amount).To(Equal(expCommAmt))
		})
	})

	Describe("Execute ClaimRewards transaction", func() {
		// defaultWithdrawRewardsArgs are the default arguments to withdraw rewards
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key otherwise is not yet initialized.
		var accruedRewards sdk.DecCoins

		BeforeEach(func() {
			var err error
			// set the default call arguments
			callArgs.MethodName = distribution.ClaimRewardsMethod
			accruedRewards, err = testutils.WaitToAccrueRewards(
				s.network,
				s.grpcHandler,
				s.keyring.GetAccAddr(0).String(),
				minExpRewardOrCommission)
			Expect(err).To(BeNil(), "error waiting to accrue rewards")
		})

		It("should return err if the origin is different than the delegator", func() {
			callArgs.Args = []interface{}{
				differentAddr, uint32(1),
			}

			claimRewardsCheck := defaultLogCheck.WithErrContains(cmn.ErrDelegatorDifferentOrigin, s.keyring.GetAddr(0).String(), differentAddr.String())

			_, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				claimRewardsCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")
		})

		It("should claim all rewards from all validators", func() {
			queryRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil(), "error while calling GetBalance")
			initialBalance := queryRes.Balance

			valCount := len(s.network.GetValidators())
			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0), uint32(valCount),
			}

			// get base fee to use in tx to then calculate fee paid
			bfQuery, err := s.grpcHandler.GetBaseFee()
			Expect(err).To(BeNil(), "error while calling BaseFee")
			gasPrice := bfQuery.BaseFee
			txArgs.GasPrice = gasPrice.BigInt()

			claimRewardsCheck := passCheck.WithExpEvents(distribution.EventTypeClaimRewards)

			txRes, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				claimRewardsCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")

			// persist state change
			Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock")

			// check that the rewards were added to the balance
			queryRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil(), "error while calling GetBalance")

			// get the fee paid and calculate the expFinalBalance
			fee := gasPrice.Mul(math.NewInt(txRes.GasUsed))
			accruedRewardsAmt := accruedRewards.AmountOf(s.bondDenom).TruncateInt()
			// expected balance is initial + rewards - fee
			expBalanceAmt := initialBalance.Amount.Add(accruedRewardsAmt).Sub(fee)

			finalBalance := queryRes.Balance
			Expect(finalBalance.Amount).To(Equal(expBalanceAmt), "expected final balance to be equal to initial balance + rewards - fees")
		})
	})
	// =====================================
	// 				QUERIES
	// =====================================
	Describe("Execute queries", func() {
		It("should get validator distribution info - validatorDistributionInfo query", func() {
			// fund validator account to make self-delegation
			err := testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), s.validatorsKeys[0].AccAddr, math.NewInt(1e17))
			Expect(err).To(BeNil())
			// persist changes
			Expect(s.network.NextBlock()).To(BeNil())

			opAddr := s.network.GetValidators()[0].OperatorAddress
			// use the validator priv key
			// make a self delegation
			err = s.factory.Delegate(s.validatorsKeys[0].Priv, opAddr, sdk.NewCoin(s.bondDenom, math.NewInt(1)))
			Expect(err).To(BeNil())
			// persist changes
			Expect(s.network.NextBlock()).To(BeNil())

			callArgs.MethodName = distribution.ValidatorDistributionInfoMethod
			callArgs.Args = []interface{}{opAddr}
			txArgs.GasLimit = 200_000

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.validatorsKeys[0].Priv,
				txArgs,
				callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var out distribution.ValidatorDistributionInfoOutput
			err = s.precompile.UnpackIntoInterface(&out, distribution.ValidatorDistributionInfoMethod, ethRes.Ret)
			Expect(err).To(BeNil())

			expAddr := s.validatorsKeys[0].AccAddr.String()
			Expect(expAddr).To(Equal(out.DistributionInfo.OperatorAddress))
			Expect(0).To(Equal(len(out.DistributionInfo.Commission)))
			Expect(0).To(Equal(len(out.DistributionInfo.SelfBondRewards)))
		})

		It("should get validator outstanding rewards - validatorOutstandingRewards query", func() {
			accruedRewards, err := testutils.WaitToAccrueRewards(
				s.network,
				s.grpcHandler,
				s.keyring.GetAccAddr(0).String(),
				minExpRewardOrCommission)
			Expect(err).To(BeNil(), "error waiting to accrue rewards")

			callArgs.MethodName = distribution.ValidatorOutstandingRewardsMethod
			callArgs.Args = []interface{}{s.network.GetValidators()[0].OperatorAddress}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var rewards []cmn.DecCoin
			err = s.precompile.UnpackIntoInterface(&rewards, distribution.ValidatorOutstandingRewardsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(rewards)).To(Equal(1))

			Expect(uint8(18)).To(Equal(rewards[0].Precision))
			Expect(s.bondDenom).To(Equal(rewards[0].Denom))

			// the expected rewards should be the accruedRewards per validator
			// plus the 5% commission
			expRewardAmt := accruedRewards.AmountOf(s.bondDenom).
				Quo(math.LegacyNewDec(3)).             // divide by validators count
				Quo(math.LegacyNewDecWithPrec(95, 2)). // add 5% commission
				Ceil().                                // round up to get the same value
				TruncateInt()

			Expect(rewards[0].Amount).To(Equal(expRewardAmt.BigInt()))
		})

		It("should get validator commission - validatorCommission query", func() {
			opAddr := s.network.GetValidators()[0].OperatorAddress
			accruedCommission, err := testutils.WaitToAccrueCommission(
				s.network,
				s.grpcHandler,
				opAddr,
				minExpRewardOrCommission)
			Expect(err).To(BeNil(), "error waiting to accrue rewards")

			callArgs.MethodName = distribution.ValidatorCommissionMethod
			callArgs.Args = []interface{}{opAddr}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var commission []cmn.DecCoin
			err = s.precompile.UnpackIntoInterface(&commission, distribution.ValidatorCommissionMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(commission)).To(Equal(1))
			Expect(uint8(18)).To(Equal(commission[0].Precision))
			Expect(s.bondDenom).To(Equal(commission[0].Denom))

			expCommissionAmt := accruedCommission.AmountOf(s.bondDenom).TruncateInt()
			Expect(commission[0].Amount).To(Equal(expCommissionAmt.BigInt()))
		})

		Context("validatorSlashes query query", Ordered, func() {
			BeforeAll(func() {
				s.withValidatorSlashes = true
				s.SetupTest()
			})
			AfterAll(func() {
				s.withValidatorSlashes = false
			})

			It("should get validator slashing events (default pagination)", func() {
				callArgs.MethodName = distribution.ValidatorSlashesMethod
				callArgs.Args = []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
					uint64(1), uint64(5),
					query.PageRequest{},
				}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil())

				var out distribution.ValidatorSlashesOutput
				err = s.precompile.UnpackIntoInterface(&out, distribution.ValidatorSlashesMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(out.Slashes)).To(Equal(2))
				// expected values according to the values used on test setup (custom genesis)
				for _, s := range out.Slashes {
					Expect(s.Fraction.Value).To(Equal(math.LegacyNewDecWithPrec(5, 2).BigInt()))
					Expect(s.ValidatorPeriod).To(Equal(uint64(1)))
				}
				Expect(uint64(2)).To(Equal(out.PageResponse.Total))
				Expect(out.PageResponse.NextKey).To(BeEmpty())
			})

			It("should get validator slashing events - query w/pagination limit = 1)", func() {
				callArgs.MethodName = distribution.ValidatorSlashesMethod
				callArgs.Args = []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
					uint64(1), uint64(5),
					query.PageRequest{
						Limit:      1,
						CountTotal: true,
					},
				}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil())

				var out distribution.ValidatorSlashesOutput
				err = s.precompile.UnpackIntoInterface(&out, distribution.ValidatorSlashesMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(out.Slashes)).To(Equal(1))
				Expect(out.Slashes[0].Fraction.Value).To(Equal(math.LegacyNewDecWithPrec(5, 2).BigInt()))
				Expect(out.Slashes[0].ValidatorPeriod).To(Equal(uint64(1)))
				// total slashes count is 2
				Expect(uint64(2)).To(Equal(out.PageResponse.Total))
				Expect(out.PageResponse.NextKey).NotTo(BeEmpty())
			})
		})

		It("should get empty delegation rewards - delegationRewards query", func() {
			callArgs.MethodName = distribution.DelegationRewardsMethod
			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0),
				s.network.GetValidators()[0].OperatorAddress,
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var rewards []cmn.DecCoin
			err = s.precompile.UnpackIntoInterface(&rewards, distribution.DelegationRewardsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(rewards)).To(Equal(0))
		})

		It("should get delegation rewards - delegationRewards query", func() {
			accruedRewards, err := testutils.WaitToAccrueRewards(s.network, s.grpcHandler, s.keyring.GetAccAddr(0).String(), minExpRewardOrCommission)
			Expect(err).To(BeNil())

			callArgs.MethodName = distribution.DelegationRewardsMethod
			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0),
				s.network.GetValidators()[0].OperatorAddress,
			}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var rewards []cmn.DecCoin
			err = s.precompile.UnpackIntoInterface(&rewards, distribution.DelegationRewardsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(rewards)).To(Equal(1))

			// The accrued rewards are based on 3 equal delegations to the existing 3 validators
			// The query is from only 1 validator, thus, the expected reward
			// for this delegation is totalAccruedRewards / validatorsCount (3)
			expRewardAmt := accruedRewards.AmountOf(s.bondDenom).Quo(math.LegacyNewDec(3))

			Expect(rewards[0].Denom).To(Equal(s.bondDenom))
			Expect(rewards[0].Amount).To(Equal(expRewardAmt.TruncateInt().BigInt()))
		})

		It("should get delegators's total rewards - delegationTotalRewards query", func() {
			// wait for rewards to accrue
			accruedRewards, err := testutils.WaitToAccrueRewards(s.network, s.grpcHandler, s.keyring.GetAccAddr(0).String(), minExpRewardOrCommission)
			Expect(err).To(BeNil())

			callArgs.MethodName = distribution.DelegationTotalRewardsMethod
			callArgs.Args = []interface{}{s.keyring.GetAddr(0)}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var out distribution.DelegationTotalRewardsOutput

			err = s.precompile.UnpackIntoInterface(&out, distribution.DelegationTotalRewardsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(3).To(Equal(len(out.Rewards)))

			// The accrued rewards are based on 3 equal delegations to the existing 3 validators
			// The query is from only 1 validator, thus, the expected reward
			// for this delegation is totalAccruedRewards / validatorsCount (3)
			accruedRewardsAmt := accruedRewards.AmountOf(s.bondDenom)
			expRewardPerValidator := accruedRewardsAmt.Quo(math.LegacyNewDec(3))

			// the response order may change
			for _, or := range out.Rewards {
				Expect(1).To(Equal(len(or.Reward)))
				Expect(or.Reward[0].Denom).To(Equal(s.bondDenom))
				Expect(or.Reward[0].Amount).To(Equal(expRewardPerValidator.TruncateInt().BigInt()))
			}

			Expect(1).To(Equal(len(out.Total)))
			Expect(out.Total[0].Amount).To(Equal(accruedRewardsAmt.TruncateInt().BigInt()))
		})

		It("should get all validators a delegators has delegated to - delegatorValidators query", func() {
			callArgs.MethodName = distribution.DelegatorValidatorsMethod
			callArgs.Args = []interface{}{s.keyring.GetAddr(0)}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var validators []string
			err = s.precompile.UnpackIntoInterface(&validators, distribution.DelegatorValidatorsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(3).To(Equal(len(validators)))
		})

		It("should get withdraw address - delegatorWithdrawAddress query", func() {
			callArgs.MethodName = distribution.DelegatorWithdrawAddressMethod
			callArgs.Args = []interface{}{s.keyring.GetAddr(0)}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")

			withdrawAddr, err := s.precompile.Unpack(distribution.DelegatorWithdrawAddressMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			// get the bech32 encoding
			expAddr := s.keyring.GetAccAddr(0)
			Expect(withdrawAddr[0]).To(Equal(expAddr.String()))
		})
	})
})

var _ = Describe("Calling distribution precompile from another contract", Ordered, func() {
	s := new(PrecompileTestSuite)
	// testCase is a struct used for cases of contracts calls that have some operation
	// performed before and/or after the precompile call
	type testCase struct {
		withdrawer *common.Address
		before     bool
		after      bool
	}

	var (
		distrCallerContract evmtypes.CompiledContract
		// contractAddr is the address of the smart contract that will be deployed
		contractAddr common.Address
		err          error

		// execRevertedCheck defines the default log checking arguments which includes the
		// standard revert message.
		execRevertedCheck testutil.LogCheckArgs
	)

	BeforeAll(func() {
		distrCallerContract, err = contracts.LoadDistributionCallerContract()
		Expect(err).To(BeNil(), "error while loading the smart contract: %v", err)
	})

	BeforeEach(func() {
		s.SetupTest()

		contractAddr, err = s.factory.DeployContract(
			s.keyring.GetPrivKey(0),
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: distrCallerContract,
			},
		)
		Expect(err).To(BeNil(), "error while deploying the smart contract: %v", err)

		// NextBlock the smart contract
		Expect(s.network.NextBlock()).To(BeNil(), "error calling NextBlock: %v", err)

		// check contract was correctly deployed
		cAcc := s.network.App.EvmKeeper.GetAccount(s.network.GetContext(), contractAddr)
		Expect(cAcc).ToNot(BeNil(), "contract account should exist")
		Expect(cAcc.IsContract()).To(BeTrue(), "account should be a contract")

		// populate default call args
		callArgs = factory.CallArgs{
			ContractABI: distrCallerContract.ABI,
		}

		// reset tx args each test to avoid keeping custom
		// values of previous tests (e.g. gasLimit)
		txArgs = evmtypes.EvmTxArgs{
			To: &contractAddr,
		}

		// default log check arguments
		defaultLogCheck = testutil.LogCheckArgs{ABIEvents: s.precompile.Events}
		execRevertedCheck = defaultLogCheck.WithErrContains("execution reverted")
		passCheck = defaultLogCheck.WithExpPass(true)
	})

	// =====================================
	// 				TRANSACTIONS
	// =====================================
	Context("setWithdrawAddress", func() {
		// newWithdrawer is the address to set the withdraw address to
		newWithdrawer := differentAddr

		BeforeEach(func() {
			// withdraw address should be same as address
			res, err := s.grpcHandler.GetDelegatorWithdrawAddr(s.keyring.GetAccAddr(0).String())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(res.WithdrawAddress).To(Equal(s.keyring.GetAccAddr(0).String()))

			// populate default arguments
			callArgs.MethodName = "testSetWithdrawAddress"
		})

		It("should set withdraw address successfully", func() {
			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0), newWithdrawer.String(),
			}

			setWithdrawCheck := passCheck.WithExpEvents(distribution.EventTypeSetWithdrawAddress)

			_, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				setWithdrawCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock: %v", err)

			queryRes, err := s.grpcHandler.GetDelegatorWithdrawAddr(s.keyring.GetAccAddr(0).String())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(queryRes.WithdrawAddress).To(Equal(sdk.AccAddress(newWithdrawer.Bytes()).String()))
		})
	})

	Context("setWithdrawerAddress with contract as delegator", func() {
		// newWithdrawer is the address to set the withdraw address to
		newWithdrawer := differentAddr

		BeforeEach(func() {
			// withdraw address should be same as address
			res, err := s.grpcHandler.GetDelegatorWithdrawAddr(s.keyring.GetAccAddr(0).String())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(res.WithdrawAddress).To(Equal(s.keyring.GetAccAddr(0).String()))

			// populate default arguments
			callArgs.MethodName = "testSetWithdrawAddressFromContract"
		})

		It("should set withdraw address successfully without origin check", func() {
			callArgs.Args = []interface{}{newWithdrawer.String()}
			setWithdrawCheck := passCheck.WithExpEvents(distribution.EventTypeSetWithdrawAddress)

			_, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				setWithdrawCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock: %v", err)

			res, err := s.grpcHandler.GetDelegatorWithdrawAddr(sdk.AccAddress(contractAddr.Bytes()).String())
			Expect(err).To(BeNil(), "error while calling GetDelegatorWithdrawAddr: %v", err)
			Expect(res.WithdrawAddress).To(Equal(sdk.AccAddress(newWithdrawer.Bytes()).String()))
		})
	})

	Context("withdrawDelegatorRewards", func() {
		// initialBalance is the initial balance of the delegator
		var initialBalance *sdk.Coin

		BeforeEach(func() {
			// fund the diffAddr
			err := testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), differentAddr.Bytes(), math.NewInt(2e18))
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			// make a delegation
			err = s.factory.Delegate(diffKey, s.network.GetValidators()[0].OperatorAddress, sdk.NewCoin(s.bondDenom, math.NewInt(1e18)))
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			// wait to accrue some rewards for s.keyring.GetAddr(0) & another address
			_, err = testutils.WaitToAccrueRewards(s.network, s.grpcHandler, sdk.AccAddress(differentAddr.Bytes()).String(), minExpRewardOrCommission)
			Expect(err).To(BeNil())

			// check if s.keyring.GetAddr(0) accrued rewards too
			_, err = testutils.WaitToAccrueRewards(s.network, s.grpcHandler, s.keyring.GetAccAddr(0).String(), minExpRewardOrCommission)
			Expect(err).To(BeNil())

			balRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			initialBalance = balRes.Balance

			callArgs.MethodName = "testWithdrawDelegatorRewards"

			// set gas price to calculate fees paid
			txArgs.GasPrice = gasPrice.BigInt()
		})

		It("should not withdraw rewards when sending from a different address", func() {
			balRes, err := s.grpcHandler.GetBalance(differentAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			differentAddrInitialBalance := balRes.Balance

			callArgs.Args = []interface{}{
				differentAddr, s.network.GetValidators()[0].OperatorAddress,
			}

			res, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock: %v", err)

			// balance should be equal as initial balance or less (because of fees)
			balRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			finalBalance := balRes.Balance
			fees := gasPrice.Mul(math.NewInt(res.GasUsed))
			Expect(finalBalance.Amount).To(Equal(initialBalance.Amount.Sub(fees)))

			// differentAddr balance should remain unchanged
			balRes, err = s.grpcHandler.GetBalance(differentAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			differentAddrFinalBalance := balRes.Balance
			Expect(differentAddrFinalBalance.Amount).To(Equal(differentAddrInitialBalance.Amount))
		})

		It("should withdraw rewards successfully", func() {
			balRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			initBalanceAmt := balRes.Balance.Amount

			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0), s.network.GetValidators()[0].OperatorAddress,
			}

			rwRes, err := s.grpcHandler.GetDelegationRewards(s.keyring.GetAccAddr(0).String(), s.network.GetValidators()[0].OperatorAddress)
			Expect(err).To(BeNil())
			expRewardsAmt := rwRes.Rewards.AmountOf(s.bondDenom).TruncateInt()

			logCheckArgs := passCheck.
				WithExpEvents(distribution.EventTypeWithdrawDelegatorRewards)

			res, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock: %v", err)

			// balance should increase
			balRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())

			fees := gasPrice.Mul(math.NewInt(res.GasUsed))

			Expect(balRes.Balance.Amount).To(Equal(initBalanceAmt.Add(expRewardsAmt).Sub(fees)), "expected final balance to be greater than initial balance after withdrawing rewards")
		})

		DescribeTable("should withdraw rewards successfully to the new withdrawer address", func(tc testCase) {
			balRes, err := s.grpcHandler.GetBalance(tc.withdrawer.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			withdrawerInitialBalance := balRes.Balance

			// Set new withdrawer address
			err = s.factory.SetWithdrawAddress(s.keyring.GetPrivKey(0), tc.withdrawer.Bytes())
			Expect(err).To(BeNil())
			// persist state change
			Expect(s.network.NextBlock()).To(BeNil())

			// get delegator initial balance
			balRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			delegatorInitialBalance := balRes.Balance

			// get the expected rewards for the delegation
			rwRes, err := s.grpcHandler.GetDelegationRewards(s.keyring.GetAccAddr(0).String(), s.network.GetValidators()[0].OperatorAddress)
			Expect(err).To(BeNil())
			expRewardsAmt := rwRes.Rewards.AmountOf(s.bondDenom).TruncateInt()

			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0), s.network.GetValidators()[0].OperatorAddress,
			}

			logCheckArgs := passCheck.
				WithExpEvents(distribution.EventTypeWithdrawDelegatorRewards)

			res, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock: %v", err)

			var rewards []cmn.Coin
			err = s.precompile.UnpackIntoInterface(&rewards, distribution.WithdrawDelegatorRewardsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(rewards)).To(Equal(1))

			Expect(rewards[0].Denom).To(Equal(s.bondDenom))
			Expect(rewards[0].Amount).To(Equal(expRewardsAmt.BigInt()))

			// should increase withdrawer balance by rewards
			balRes, err = s.grpcHandler.GetBalance(tc.withdrawer.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())

			Expect(balRes.Balance.Amount).To(Equal(withdrawerInitialBalance.Amount.Add(expRewardsAmt)), "expected final balance to be greater than initial balance after withdrawing rewards")

			// check that the delegator final balance is initialBalance - fee
			balRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil(), "error while calling GetBalance")
			fees := gasPrice.Mul(math.NewInt(res.GasUsed))

			expDelgatorFinal := delegatorInitialBalance.Amount.Sub(fees)
			Expect(balRes.Balance.Amount).To(Equal(expDelgatorFinal), "expected delegator final balance to be equal to initial balance - fees")
		},
			Entry("withdrawer addr is existing acc", testCase{
				withdrawer: &differentAddr,
			}),
			Entry("withdrawer addr is non-existing acc", testCase{
				withdrawer: func() *common.Address {
					addr := testutiltx.GenerateAddress()
					return &addr
				}(),
			}),
		)

		// Specific BeforeEach for table-driven tests
		Context("Table-driven tests for Withdraw Delegator Rewards", func() {
			contractInitialBalance := math.NewInt(100)

			BeforeEach(func() {
				callArgs.MethodName = "testWithdrawDelegatorRewardsWithTransfer"

				// send some funds to the contract
				err := testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), contractAddr.Bytes(), contractInitialBalance)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())
			})

			DescribeTable("withdraw delegation rewards with internal transfers to delegator - should withdraw rewards successfully to the withdrawer address",
				func(tc testCase) {
					balRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
					Expect(err).To(BeNil())
					if tc.withdrawer != nil {
						// Set new withdrawer address
						err = s.factory.SetWithdrawAddress(s.keyring.GetPrivKey(0), tc.withdrawer.Bytes())
						Expect(err).To(BeNil())
						// persist state change
						Expect(s.network.NextBlock()).To(BeNil())
						balRes, err = s.grpcHandler.GetBalance(tc.withdrawer.Bytes(), s.bondDenom)
						Expect(err).To(BeNil())
					}
					withdrawerInitialBalance := balRes.Balance

					balRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
					Expect(err).To(BeNil())
					delInitialBalance := balRes.Balance

					// get the pending rewards to claim
					qRes, err := s.grpcHandler.GetDelegationRewards(s.keyring.GetAccAddr(0).String(), s.network.GetValidators()[0].OperatorAddress)
					Expect(err).To(BeNil())
					expRewards := qRes.Rewards.AmountOf(s.bondDenom).TruncateInt()

					callArgs.Args = []interface{}{
						s.keyring.GetAddr(0), s.network.GetValidators()[0].OperatorAddress, tc.before, tc.after,
					}

					logCheckArgs := passCheck.
						WithExpEvents(distribution.EventTypeWithdrawDelegatorRewards)

					res, _, err := s.factory.CallContractAndCheckLogs(
						s.keyring.GetPrivKey(0),
						txArgs,
						callArgs,
						logCheckArgs,
					)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
					Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock: %v", err)

					fees := gasPrice.MulRaw(res.GasUsed)

					// check balances
					contractTransferredAmt := math.ZeroInt()
					for _, transferred := range []bool{tc.before, tc.after} {
						if transferred {
							contractTransferredAmt = contractTransferredAmt.AddRaw(15)
						}
					}
					// contract balance be updated according to the transferred amount
					balRes, err = s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
					Expect(err).To(BeNil())
					contractFinalBalance := balRes.Balance
					Expect(contractFinalBalance.Amount).To(Equal(contractInitialBalance.Sub(contractTransferredAmt)))

					expDelFinalBalance := delInitialBalance.Amount.Sub(fees).Add(contractTransferredAmt).Add(expRewards)
					if tc.withdrawer != nil {
						expDelFinalBalance = delInitialBalance.Amount.Sub(fees).Add(contractTransferredAmt)
						expWithdrawerFinalBalance := withdrawerInitialBalance.Amount.Add(expRewards)
						// withdrawer balance should have the rewards
						balRes, err = s.grpcHandler.GetBalance(tc.withdrawer.Bytes(), s.bondDenom)
						Expect(err).To(BeNil())
						withdrawerFinalBalance := balRes.Balance
						Expect(withdrawerFinalBalance.Amount).To(Equal(expWithdrawerFinalBalance), "expected final balance to be greater than initial balance after withdrawing rewards")
					}

					// delegator balance should have the transferred amt - fees + rewards (when is the withdrawer)
					balRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
					Expect(err).To(BeNil())
					delFinalBalance := balRes.Balance
					Expect(delFinalBalance.Amount).To(Equal(expDelFinalBalance), "expected final balance to be greater than initial balance after withdrawing rewards")
				},

				Entry("delegator == withdrawer - with internal transfers before and after precompile call", testCase{
					before: true,
					after:  true,
				}),

				Entry("delegator == withdrawer - with internal transfers before precompile call", testCase{
					before: true,
					after:  false,
				}),

				Entry("delegator == withdrawer - with internal transfers after precompile call", testCase{
					before: false,
					after:  true,
				}),
				Entry("delegator != withdrawer - with internal transfers before and after precompile call", testCase{
					withdrawer: &differentAddr,
					before:     true,
					after:      true,
				}),

				Entry("delegator != withdrawer - with internal transfers before precompile call", testCase{
					withdrawer: &differentAddr,
					before:     true,
					after:      false,
				}),

				Entry("delegator != withdrawer - with internal transfers after precompile call", testCase{
					withdrawer: &differentAddr,
					before:     false,
					after:      true,
				}),
			)

			DescribeTable("should revert withdraw rewards successfully and update correspondingly the withdrawer and contract's balances", func(tc testCase) {
				// Set new withdrawer address
				err = s.factory.SetWithdrawAddress(s.keyring.GetPrivKey(0), tc.withdrawer.Bytes())
				Expect(err).To(BeNil())
				// persist state change
				Expect(s.network.NextBlock()).To(BeNil())

				// get the pending rewards to claim
				qRes, err := s.grpcHandler.GetDelegationRewards(s.keyring.GetAccAddr(0).String(), s.network.GetValidators()[0].OperatorAddress)
				Expect(err).To(BeNil())
				initRewards := qRes.Rewards.AmountOf(s.bondDenom).TruncateInt()

				balRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
				Expect(err).To(BeNil())
				delInitBalance := balRes.Balance
				balRes, err = s.grpcHandler.GetBalance(tc.withdrawer.Bytes(), s.bondDenom)
				Expect(err).To(BeNil())
				withdrawerInitBalance := balRes.Balance

				// update args to call the corresponding contract method
				callArgs.MethodName = "revertWithdrawRewardsAndTransfer"
				callArgs.Args = []interface{}{
					s.keyring.GetAddr(0), *tc.withdrawer, s.network.GetValidators()[0].OperatorAddress, true,
				}

				res, _, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock: %v", err)
				fees := gasPrice.MulRaw(res.GasUsed)

				// check balances
				contractTransferredAmt := math.NewInt(15)
				// contract balance be updated according to the transferred amount
				balRes, err = s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
				Expect(err).To(BeNil())
				contractFinalBalance := balRes.Balance
				Expect(contractFinalBalance.Amount).To(Equal(contractInitialBalance.Sub(contractTransferredAmt)))

				// delegator balance should be initial_balance - fees
				balRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
				Expect(err).To(BeNil())
				delFinalBalance := balRes.Balance
				Expect(delFinalBalance.Amount).To(Equal(delInitBalance.Amount.Sub(fees)))

				// withdrawer balance should increase by the transferred amount only
				// the rewards withdrawal should revert
				balRes, err = s.grpcHandler.GetBalance(tc.withdrawer.Bytes(), s.bondDenom)
				Expect(err).To(BeNil())
				withdrawerFinalBalance := balRes.Balance
				Expect(withdrawerFinalBalance.Amount).To(Equal(withdrawerInitBalance.Amount.Add(contractTransferredAmt)), "expected final balance to be greater than initial balance after withdrawing rewards")

				// rewards to claim should be the same or more than before
				qRes, err = s.grpcHandler.GetDelegationRewards(s.keyring.GetAccAddr(0).String(), s.network.GetValidators()[0].OperatorAddress)
				Expect(err).To(BeNil())
				finalRewards := qRes.Rewards.AmountOf(s.bondDenom).TruncateInt()
				Expect(finalRewards.GTE(initRewards)).To(BeTrue())
			},
				Entry("withdrawer addr is existing acc", testCase{
					withdrawer: &differentAddr,
				}),
				Entry("withdrawer addr is non-existing acc", testCase{
					withdrawer: func() *common.Address {
						addr := testutiltx.GenerateAddress()
						return &addr
					}(),
				}),
			)
		})
	})

	Context("withdrawDelegatorRewards with contract as delegator", func() {
		var (
			// initialBalance is the initial balance of the delegator
			initialBalance    *sdk.Coin
			accruedRewardsAmt math.Int
		)

		BeforeEach(func() { //nolint:dupl
			// send funds to the contract
			err := testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), contractAddr.Bytes(), math.NewInt(2e18))
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			stkPrecompile, err := s.getStakingPrecompile()
			Expect(err).To(BeNil())
			// make a delegation with contract as delegator
			logCheck := testutil.LogCheckArgs{
				ExpPass:   true,
				ABIEvents: stkPrecompile.ABI.Events,
				ExpEvents: []string{authorization.EventTypeApproval, staking.EventTypeDelegate},
			}
			_, _, err = s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				factory.CallArgs{
					ContractABI: distrCallerContract.ABI,
					MethodName:  "testDelegateFromContract",
					Args: []interface{}{
						s.network.GetValidators()[0].OperatorAddress,
						big.NewInt(1e18),
					},
				},
				logCheck,
			)
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			// wait to accrue some rewards for contract address
			rwRes, err := testutils.WaitToAccrueRewards(s.network, s.grpcHandler, sdk.AccAddress(contractAddr.Bytes()).String(), minExpRewardOrCommission)
			Expect(err).To(BeNil())

			// contract's accrued rewards amt
			accruedRewardsAmt = rwRes.AmountOf(s.bondDenom).TruncateInt()

			balRes, err := s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			initialBalance = balRes.Balance

			// populate default arguments
			callArgs.MethodName = "testWithdrawDelegatorRewardsFromContract"
		})

		It("should withdraw rewards successfully without origin check", func() {
			callArgs.Args = []interface{}{s.network.GetValidators()[0].OperatorAddress}

			logCheckArgs := passCheck.WithExpEvents(distribution.EventTypeWithdrawDelegatorRewards)

			_, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock: %v", err)

			// balance should increase
			balRes, err := s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			finalBalance := balRes.Balance
			Expect(finalBalance.Amount).To(Equal(initialBalance.Amount.Add(accruedRewardsAmt)), "expected final balance to be greater than initial balance after withdrawing rewards")
		})

		It("should withdraw rewards successfully without origin check to a withdrawer address", func() {
			withdrawerAddr, _ := testutiltx.NewAccAddressAndKey()

			balRes, err := s.grpcHandler.GetBalance(withdrawerAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			initialWithdrawerBalance := balRes.Balance
			Expect(initialWithdrawerBalance.Amount).To(Equal(math.ZeroInt()))

			// call the smart contract to update the withdrawer
			// Set new withdrawer address for the contract
			setWithdrawCheck := passCheck.WithExpEvents(distribution.EventTypeSetWithdrawAddress)
			res1, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				factory.CallArgs{
					ContractABI: distrCallerContract.ABI,
					MethodName:  "testSetWithdrawAddressFromContract",
					Args:        []interface{}{withdrawerAddr.String()},
				},
				setWithdrawCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(res1.IsOK()).To(BeTrue(), "error while calling the smart contract")
			Expect(s.network.NextBlock()).To(BeNil())

			// get accrued rewards prev to tx
			rwRes, err := s.grpcHandler.GetDelegationRewards(sdk.AccAddress(contractAddr.Bytes()).String(), s.network.GetValidators()[0].OperatorAddress)
			Expect(err).To(BeNil())
			accruedRewardsAmt = rwRes.Rewards.AmountOf(s.bondDenom).TruncateInt()

			callArgs.Args = []interface{}{s.network.GetValidators()[0].OperatorAddress}
			logCheckArgs := passCheck.WithExpEvents(distribution.EventTypeWithdrawDelegatorRewards)

			txArgs.GasLimit = 300_000
			_, _, err = s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock: %v", err)

			// withdrawer balance should increase with the rewards amt
			balRes, err = s.grpcHandler.GetBalance(withdrawerAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			finalWithdrawerBalance := balRes.Balance
			Expect(finalWithdrawerBalance.Amount).To(Equal(accruedRewardsAmt), "expected final balance to be greater than initial balance after withdrawing rewards")

			// delegator balance (contract) should remain unchanged
			balRes, err = s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			finalDelegatorBalance := balRes.Balance
			Expect(finalDelegatorBalance.Amount.Equal(initialBalance.Amount)).To(BeTrue(), "expected delegator final balance remain unchanged after withdrawing rewards to withdrawer")
		})

		Context("Withdraw Delegator Rewards with another smart contract (different than the contract calling the precompile) as delegator", func() {
			var (
				delContractAddr        common.Address
				contractInitialBalance = math.NewInt(100)
			)
			BeforeEach(func() {
				callArgs.MethodName = "testWithdrawDelegatorRewardsWithTransfer"

				// deploy another delegator contract
				delContractAddr, err = s.factory.DeployContract(
					s.keyring.GetPrivKey(0),
					evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
					factory.ContractDeploymentData{
						Contract: distrCallerContract,
					},
				)
				Expect(err).To(BeNil(), "error while deploying the smart contract: %v", err)
				// NextBlock the smart contract
				Expect(s.network.NextBlock()).To(BeNil(), "error calling NextBlock: %v", err)

				// send funds to the contract
				err := testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), delContractAddr.Bytes(), math.NewInt(2e18))
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				stkPrecompile, err := s.getStakingPrecompile()
				Expect(err).To(BeNil())
				// make a delegation with contract as delegator
				logCheck := testutil.LogCheckArgs{
					ExpPass:   true,
					ABIEvents: stkPrecompile.ABI.Events,
					ExpEvents: []string{authorization.EventTypeApproval, staking.EventTypeDelegate},
				}
				_, _, err = s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					evmtypes.EvmTxArgs{
						To: &delContractAddr,
					},
					factory.CallArgs{
						ContractABI: distrCallerContract.ABI,
						MethodName:  "testDelegateFromContract",
						Args: []interface{}{
							s.network.GetValidators()[0].OperatorAddress,
							big.NewInt(1e18),
						},
					},
					logCheck,
				)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				// wait to accrue some rewards for contract address
				rwRes, err := testutils.WaitToAccrueRewards(s.network, s.grpcHandler, sdk.AccAddress(delContractAddr.Bytes()).String(), minExpRewardOrCommission)
				Expect(err).To(BeNil())

				// contract's accrued rewards amt
				accruedRewardsAmt = rwRes.AmountOf(s.bondDenom).TruncateInt()

				balRes, err := s.grpcHandler.GetBalance(delContractAddr.Bytes(), s.bondDenom)
				Expect(err).To(BeNil())
				initialBalance = balRes.Balance

				// send some funds to the contract
				err = testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), contractAddr.Bytes(), contractInitialBalance)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())
			})

			It("should NOT allow to withdraw rewards", func() {
				balRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
				Expect(err).To(BeNil())
				txSenderInitialBalance := balRes.Balance
				balRes, err = s.grpcHandler.GetBalance(delContractAddr.Bytes(), s.bondDenom)
				Expect(err).To(BeNil())
				delInitialBalance := balRes.Balance
				balRes, err = s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
				Expect(err).To(BeNil())
				callerContractInitialBal := balRes.Balance

				// get the pending rewards to claim
				rwRes, err := s.grpcHandler.GetDelegationRewards(sdk.AccAddress(delContractAddr.Bytes()).String(), s.network.GetValidators()[0].OperatorAddress)
				Expect(err).To(BeNil())
				expRewards := rwRes.Rewards.AmountOf(s.bondDenom).TruncateInt()

				callArgs.Args = []interface{}{delContractAddr, s.network.GetValidators()[0].OperatorAddress, true, true}

				res, _, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					execRevertedCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock: %v", err)
				fees := gasPrice.MulRaw(res.GasUsed)
				// check balances
				// tx signer final balance should be the initial balance - fees
				balRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
				Expect(err).To(BeNil())
				txSignerFinalBalance := balRes.Balance
				Expect(txSignerFinalBalance.Amount).To(Equal(txSenderInitialBalance.Amount.Sub(fees)))

				// caller contract balance should remain unchanged
				balRes, err = s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
				Expect(err).To(BeNil())
				contractFinalBalance := balRes.Balance
				Expect(contractFinalBalance).To(Equal(callerContractInitialBal))

				// delegator balance should remain unchanged
				balRes, err = s.grpcHandler.GetBalance(delContractAddr.Bytes(), s.bondDenom)
				Expect(err).To(BeNil())
				delFinalBalance := balRes.Balance
				Expect(delFinalBalance).To(Equal(delInitialBalance))

				// delegation rewards should remain be the same or higher
				rwRes, err = s.grpcHandler.GetDelegationRewards(sdk.AccAddress(delContractAddr.Bytes()).String(), s.network.GetValidators()[0].OperatorAddress)
				Expect(err).To(BeNil())
				finalRewards := rwRes.Rewards.AmountOf(s.bondDenom).TruncateInt()
				Expect(finalRewards.GTE(expRewards)).To(BeTrue())
			})
		})

		It("should withdraw rewards successfully without origin check to a withdrawer address", func() {
			withdrawerAddr, _ := testutiltx.NewAccAddressAndKey()

			balRes, err := s.grpcHandler.GetBalance(withdrawerAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			initialWithdrawerBalance := balRes.Balance
			Expect(initialWithdrawerBalance.Amount).To(Equal(math.ZeroInt()))

			// Set new withdrawer address for the contract
			setWithdrawCheck := passCheck.WithExpEvents(distribution.EventTypeSetWithdrawAddress)
			res1, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				factory.CallArgs{
					ContractABI: distrCallerContract.ABI,
					MethodName:  "testSetWithdrawAddressFromContract",
					Args:        []interface{}{withdrawerAddr.String()},
				},
				setWithdrawCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(res1.IsOK()).To(BeTrue(), "error while calling the smart contract")
			Expect(s.network.NextBlock()).To(BeNil())

			// get the pending rewards to claim
			rwRes, err := s.grpcHandler.GetDelegationRewards(sdk.AccAddress(contractAddr.Bytes()).String(), s.network.GetValidators()[0].OperatorAddress)
			Expect(err).To(BeNil())
			expRewards := rwRes.Rewards.AmountOf(s.bondDenom).TruncateInt()

			logCheckArgs := passCheck.WithExpEvents(distribution.EventTypeWithdrawDelegatorRewards)

			callArgs.Args = []interface{}{s.network.GetValidators()[0].OperatorAddress}

			txArgs.GasLimit = 500_000
			_, _, err = s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil(), "error on NextBlock: %v", err)

			// withdrawer balance should increase with the rewards amt
			balRes, err = s.grpcHandler.GetBalance(withdrawerAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			finalWithdrawerBalance := balRes.Balance
			Expect(finalWithdrawerBalance.Amount.Equal(expRewards)).To(BeTrue(), "expected final balance to be greater than initial balance after withdrawing rewards")

			// delegator balance (contract) should remain unchanged
			balRes, err = s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			finalDelegatorBalance := balRes.Balance
			Expect(finalDelegatorBalance.Amount.Equal(initialBalance.Amount)).To(BeTrue(), "expected delegator final balance remain unchanged after withdrawing rewards to withdrawer")
		})
	})

	Context("withdrawValidatorCommission", func() {
		var (
			// initialBalance is the initial balance of the delegator
			initialBalance *sdk.Coin
			// valInitialBalance is the initial balance of the validator
			valInitialBalance    *sdk.Coin
			accruedCommissionAmt math.Int
		)

		BeforeEach(func() {
			// fund validator's account to pay for fees
			err := testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), s.validatorsKeys[0].AccAddr, math.NewInt(1e18))
			Expect(err).To(BeNil())

			res, err := testutils.WaitToAccrueCommission(s.network, s.grpcHandler, s.network.GetValidators()[0].OperatorAddress, minExpRewardOrCommission)
			Expect(err).To(BeNil())
			accruedCommissionAmt = res.AmountOf(s.bondDenom).TruncateInt()

			balRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			initialBalance = balRes.Balance

			// get validators initial balance
			balRes, err = s.grpcHandler.GetBalance(s.validatorsKeys[0].AccAddr, s.bondDenom)
			Expect(err).To(BeNil())
			valInitialBalance = balRes.Balance

			// populate default arguments
			callArgs.MethodName = "testWithdrawValidatorCommission"
		})

		It("should not withdraw commission from validator when sending from a different address", func() {
			callArgs.Args = []interface{}{
				s.network.GetValidators()[0].OperatorAddress,
			}

			res, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			// balance should be equal as initial balance or less (because of fees)
			balRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			finalBalance := balRes.Balance

			fees := gasPrice.Mul(math.NewInt(res.GasUsed))
			Expect(finalBalance.Amount).To(Equal(initialBalance.Amount.Sub(fees)))

			// validator's balance should remain unchanged
			balRes, err = s.grpcHandler.GetBalance(s.validatorsKeys[0].AccAddr, s.bondDenom)
			Expect(err).To(BeNil())
			valFinalBalance := balRes.Balance
			Expect(valFinalBalance.Amount).To(Equal(valInitialBalance.Amount))
		})

		It("should withdraw commission successfully", func() {
			callArgs.Args = []interface{}{s.network.GetValidators()[0].OperatorAddress}

			logCheckArgs := passCheck.
				WithExpEvents(distribution.EventTypeWithdrawValidatorCommission)

			txArgs.GasPrice = gasPrice.BigInt()
			res, _, err := s.factory.CallContractAndCheckLogs(
				s.validatorsKeys[0].Priv,
				txArgs,
				callArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			balRes, err := s.grpcHandler.GetBalance(s.validatorsKeys[0].AccAddr, s.bondDenom)
			Expect(err).To(BeNil())
			valFinalBalance := balRes.Balance
			fees := gasPrice.Mul(math.NewInt(res.GasUsed))
			expFinal := valInitialBalance.Amount.Add(accruedCommissionAmt).Sub(fees)
			Expect(valFinalBalance.Amount).To(Equal(expFinal), "expected final balance to be equal to initial balance + validator commission - fees")
		})

		It("should withdraw commission successfully to withdrawer address (contract)", func() {
			balRes, err := s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			initialWithdrawerBalance := balRes.Balance
			Expect(initialWithdrawerBalance.Amount).To(Equal(math.ZeroInt()))

			// Set new withdrawer address
			err = s.factory.SetWithdrawAddress(s.validatorsKeys[0].Priv, contractAddr.Bytes())
			Expect(err).To(BeNil())
			// persist state change
			Expect(s.network.NextBlock()).To(BeNil())

			qRes, err := s.grpcHandler.GetValidatorCommission(s.network.GetValidators()[0].OperatorAddress)
			Expect(err).To(BeNil())
			accruedCommissionAmt = qRes.Commission.Commission.AmountOf(s.bondDenom).TruncateInt()

			// validator acc balance before the tx
			balRes, err = s.grpcHandler.GetBalance(s.validatorsKeys[0].AccAddr, s.bondDenom)
			Expect(err).To(BeNil())
			initialBalance := balRes.Balance

			callArgs.Args = []interface{}{s.network.GetValidators()[0].OperatorAddress}

			logCheckArgs := passCheck.
				WithExpEvents(distribution.EventTypeWithdrawValidatorCommission)

			txArgs.GasLimit = 500_000
			txArgs.GasPrice = gasPrice.BigInt()
			res, _, err := s.factory.CallContractAndCheckLogs(
				s.validatorsKeys[0].Priv,
				txArgs,
				callArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			balRes, err = s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			finalWithdrawerBalance := balRes.Balance
			Expect(finalWithdrawerBalance.Amount).To(Equal(initialWithdrawerBalance.Amount.Add(accruedCommissionAmt)), "expected final balance to be equal to initial balance + validator commission")

			balRes, err = s.grpcHandler.GetBalance(s.validatorsKeys[0].AccAddr, s.bondDenom)
			Expect(err).To(BeNil())
			finalBalance := balRes.Balance
			fees := gasPrice.MulRaw(res.GasUsed)
			expFinal := initialBalance.Amount.Sub(fees)
			Expect(finalBalance.Amount).To(Equal(expFinal), "expected final balance to be equal to initial balance  - fees")
		})

		// Specific BeforeEach for table-driven tests
		Context("Table-driven tests for Withdraw Validator Commission", func() {
			contractInitialBalance := math.NewInt(100)
			BeforeEach(func() {
				callArgs.MethodName = "testWithdrawValidatorCommissionWithTransfer"

				// send some funds to the contract
				err := testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), contractAddr.Bytes(), contractInitialBalance)
				Expect(err).To(BeNil())
				// persist state change
				Expect(s.network.NextBlock()).To(BeNil())
			})

			DescribeTable("withdraw validator commission with state changes in withdrawer - should withdraw commission successfully to the withdrawer address",
				func(tc testCase) {
					withdrawerAddr := s.validatorsKeys[0].Addr
					balRes, err := s.grpcHandler.GetBalance(s.validatorsKeys[0].AccAddr, s.bondDenom)
					Expect(err).To(BeNil())
					if tc.withdrawer != nil {
						withdrawerAddr = *tc.withdrawer
						// Set new withdrawer address
						err = s.factory.SetWithdrawAddress(s.validatorsKeys[0].Priv, tc.withdrawer.Bytes())
						Expect(err).To(BeNil())
						// persist state change
						Expect(s.network.NextBlock()).To(BeNil())
						balRes, err = s.grpcHandler.GetBalance(tc.withdrawer.Bytes(), s.bondDenom)
						Expect(err).To(BeNil())
					}
					withdrawerInitialBalance := balRes.Balance

					// validator acc balance before the tx
					balRes, err = s.grpcHandler.GetBalance(s.validatorsKeys[0].AccAddr, s.bondDenom)
					Expect(err).To(BeNil())
					valInitialBalance := balRes.Balance

					// get the pending commission to claim
					qRes, err := s.grpcHandler.GetValidatorCommission(s.network.GetValidators()[0].OperatorAddress)
					Expect(err).To(BeNil())
					expCommission := qRes.Commission.Commission.AmountOf(s.bondDenom).TruncateInt()

					callArgs.Args = []interface{}{s.network.GetValidators()[0].OperatorAddress, withdrawerAddr, tc.before, tc.after}

					logCheckArgs := passCheck.
						WithExpEvents(distribution.EventTypeWithdrawValidatorCommission)

					txArgs.GasPrice = gasPrice.BigInt()
					txArgs.GasLimit = 600_000
					res, _, err := s.factory.CallContractAndCheckLogs(
						s.validatorsKeys[0].Priv,
						txArgs,
						callArgs,
						logCheckArgs,
					)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
					Expect(s.network.NextBlock()).To(BeNil())

					fees := gasPrice.MulRaw(res.GasUsed)

					// calculate the transferred amt during the call
					contractTransferredAmt := math.ZeroInt()
					for _, transferred := range []bool{tc.before, tc.after} {
						if transferred {
							contractTransferredAmt = contractTransferredAmt.AddRaw(15)
						}
					}

					// check balances
					expContractFinalBalance := contractInitialBalance.Sub(contractTransferredAmt)
					expValFinalBalance := valInitialBalance.Amount.Sub(fees).Add(contractTransferredAmt).Add(expCommission)
					if tc.withdrawer != nil {
						expValFinalBalance = valInitialBalance.Amount.Sub(fees)
						if *tc.withdrawer == contractAddr {
							// no internal transfers if the contract itself is the withdrawer
							expContractFinalBalance = contractInitialBalance.Add(expCommission)
						} else {
							expWithdrawerFinalBalance := withdrawerInitialBalance.Amount.Add(expCommission).Add(contractTransferredAmt)
							// withdrawer balance should have the rewards
							balRes, err = s.grpcHandler.GetBalance(tc.withdrawer.Bytes(), s.bondDenom)
							Expect(err).To(BeNil())
							withdrawerFinalBalance := balRes.Balance
							Expect(withdrawerFinalBalance.Amount).To(Equal(expWithdrawerFinalBalance), "expected final balance to be greater than initial balance after withdrawing rewards")
						}
					}

					// contract balance be updated according to the transferred amount
					balRes, err = s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
					Expect(err).To(BeNil())
					contractFinalBalance := balRes.Balance
					Expect(contractFinalBalance.Amount).To(Equal(expContractFinalBalance))

					// validator balance should have the transferred amt - fees + rewards (when is the withdrawer)
					balRes, err = s.grpcHandler.GetBalance(s.validatorsKeys[0].AccAddr, s.bondDenom)
					Expect(err).To(BeNil())
					valFinalBalance := balRes.Balance
					Expect(valFinalBalance.Amount).To(Equal(expValFinalBalance), "expected final balance to be greater than initial balance after withdrawing rewards")
				},

				Entry("validator == withdrawer - with internal transfers before and after precompile call", testCase{
					before: true,
					after:  true,
				}),

				Entry("validator == withdrawer - with internal transfers before precompile call", testCase{
					before: true,
					after:  false,
				}),

				Entry("validator == withdrawer - with internal transfers after precompile call", testCase{
					before: false,
					after:  true,
				}),
				Entry("validator != withdrawer - with internal transfers before and after precompile call", testCase{
					withdrawer: &differentAddr,
					before:     true,
					after:      true,
				}),

				Entry("validator != withdrawer - with internal transfers before precompile call", testCase{
					withdrawer: &differentAddr,
					before:     true,
					after:      false,
				}),

				Entry("validator != withdrawer - with internal transfers after precompile call", testCase{
					withdrawer: &differentAddr,
					before:     false,
					after:      true,
				}),
				Entry("contract as withdrawer - with contract state change before and after precompile call", testCase{
					withdrawer: &contractAddr,
					before:     true,
					after:      true,
				}),

				Entry("contract as withdrawer - with contract state change before precompile call", testCase{
					withdrawer: &contractAddr,
					before:     true,
					after:      false,
				}),

				Entry("contract as withdrawer - with contract state change after precompile call", testCase{
					withdrawer: &contractAddr,
					before:     false,
					after:      true,
				}),
			)
		})
	})

	Context("claimRewards", func() {
		var (
			// initialBalance is the initial balance of the delegator
			initialBalance *sdk.Coin
			// diffAddrInitialBalance is the initial balance of the different address
			diffAddrInitialBalance *sdk.Coin
			accruedRewardsAmt      math.Int
		)

		BeforeEach(func() {
			// fund the diffAddr
			err := testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), differentAddr.Bytes(), math.NewInt(2e18))
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			// make a delegation
			err = s.factory.Delegate(diffKey, s.network.GetValidators()[0].OperatorAddress, sdk.NewCoin(s.bondDenom, math.NewInt(1e18)))
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			// wait to accrue some rewards for s.keyring.GetAddr(0) & another address
			_, err = testutils.WaitToAccrueRewards(s.network, s.grpcHandler, sdk.AccAddress(differentAddr.Bytes()).String(), minExpRewardOrCommission)
			Expect(err).To(BeNil())

			// check if s.keyring.GetAddr(0) accrued rewards too
			res, err := s.grpcHandler.GetDelegationTotalRewards(s.keyring.GetAccAddr(0).String())
			Expect(err).To(BeNil())

			accruedRewardsAmt = res.Total.AmountOf(s.bondDenom).TruncateInt()
			Expect(accruedRewardsAmt.IsPositive()).To(BeTrue())

			balRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			initialBalance = balRes.Balance

			balRes, err = s.grpcHandler.GetBalance(differentAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			diffAddrInitialBalance = balRes.Balance

			// populate default arguments
			callArgs.MethodName = "testClaimRewards"
			txArgs.GasPrice = gasPrice.BigInt()
		})

		It("should not claim rewards when sending from a different address", func() {
			callArgs.Args = []interface{}{differentAddr, uint32(1)}

			res, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			// balance should be equal as initial balance or less (because of fees)
			balRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			finalBalance := balRes.Balance
			fees := gasPrice.Mul(math.NewInt(res.GasUsed))
			Expect(finalBalance.Amount).To(Equal(initialBalance.Amount.Sub(fees)))

			// differentAddr balance should remain unchanged
			balRes, err = s.grpcHandler.GetBalance(differentAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			differentAddrFinalBalance := balRes.Balance
			Expect(differentAddrFinalBalance.Amount).To(Equal(diffAddrInitialBalance.Amount))
		})

		It("should claim rewards successfully", func() {
			callArgs.Args = []interface{}{s.keyring.GetAddr(0), uint32(2)}

			logCheckArgs := passCheck.
				WithExpEvents(distribution.EventTypeClaimRewards)

			_, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			// balance should remain unchanged
			balRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			finalBalance := balRes.Balance
			Expect(finalBalance.Amount.GT(initialBalance.Amount)).To(BeTrue(), "expected final balance to be greater than initial balance after claiming rewards")
		})

		Context("Table driven tests", func() {
			contractInitialBalance := math.NewInt(100)
			BeforeEach(func() {
				callArgs.MethodName = "testClaimRewardsWithTransfer"

				// send some funds to the contract
				err = testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), contractAddr.Bytes(), contractInitialBalance)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				// make a delegation with key 1
				err = s.factory.Delegate(s.keyring.GetKey(1).Priv, s.network.GetValidators()[0].OperatorAddress, sdk.NewCoin(s.bondDenom, math.NewInt(1e18)))
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				// wait to accrue some rewards for key 1
				_, err := testutils.WaitToAccrueRewards(s.network, s.grpcHandler, s.keyring.GetAccAddr(1).String(), minExpRewardOrCommission)
				Expect(err).To(BeNil())
			})

			DescribeTable("claimRewards with transfer to withdrawer", func(tc testCase) {
				initialBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAccAddr(1), s.bondDenom)

				// get the pending rewards to claim
				res, err := s.grpcHandler.GetDelegationTotalRewards(s.keyring.GetAccAddr(1).String())
				Expect(err).To(BeNil())
				expRewards := res.Total.AmountOf(s.bondDenom).TruncateInt()

				callArgs.Args = []interface{}{s.keyring.GetAddr(1), uint32(2), tc.before, tc.after}

				logCheckArgs := passCheck.
					WithExpEvents(distribution.EventTypeClaimRewards)
				txArgs.GasLimit = 400_000 // set gas limit to avoid out of gas error
				_, ethres, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(1),
					txArgs,
					callArgs,
					logCheckArgs,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
				err = s.network.NextBlock()
				Expect(err).To(BeNil())

				fees := math.NewIntFromBigInt(txArgs.GasPrice).MulRaw(int64(ethres.GasUsed))

				// calculate the transferred amt during the call
				contractTransferredAmt := math.ZeroInt()
				for _, transferred := range []bool{tc.before, tc.after} {
					if transferred {
						contractTransferredAmt = contractTransferredAmt.AddRaw(15)
					}
				}

				// check balances
				expContractFinalBalance := contractInitialBalance.Sub(contractTransferredAmt)
				expDelFinalBalance := initialBalance.Amount.Sub(fees).Add(contractTransferredAmt).Add(expRewards)

				// contract balance be updated according to the transferred amount
				contractFinalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), contractAddr.Bytes(), s.bondDenom)
				Expect(contractFinalBalance.Amount).To(Equal(expContractFinalBalance))

				// delegator (and withdrawer) balance should be updated
				finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAccAddr(1), s.bondDenom)
				Expect(finalBalance.Amount).To(Equal(expDelFinalBalance), "expected final balance to be greater than initial balance after claiming rewards")
			},
				Entry("claim rewards with transfer to withdrawer before and after precompile call", testCase{
					before: true,
					after:  true,
				}),
				Entry("claim rewards with transfer to withdrawer before precompile call", testCase{
					before: true,
					after:  false,
				}),
				Entry("claim rewards with transfer to withdrawer after precompile call", testCase{
					before: false,
					after:  true,
				}),
			)
		})
	})

	Context("claimRewards with contract as delegator", func() {
		var (
			initialBalance    *sdk.Coin
			accruedRewardsAmt math.Int
		)

		BeforeEach(func() { //nolint:dupl
			// send funds to the contract
			err := testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), contractAddr.Bytes(), math.NewInt(2e18))
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			stkPrecompile, err := s.getStakingPrecompile()
			Expect(err).To(BeNil())
			// make a delegation with contract as delegator
			logCheck := testutil.LogCheckArgs{
				ExpPass:   true,
				ABIEvents: stkPrecompile.ABI.Events,
				ExpEvents: []string{authorization.EventTypeApproval, staking.EventTypeDelegate},
			}
			_, _, err = s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				factory.CallArgs{
					ContractABI: distrCallerContract.ABI,
					MethodName:  "testDelegateFromContract",
					Args: []interface{}{
						s.network.GetValidators()[0].OperatorAddress,
						big.NewInt(1e18),
					},
				},
				logCheck,
			)
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			// wait to accrue some rewards for contract address
			rwRes, err := testutils.WaitToAccrueRewards(s.network, s.grpcHandler, sdk.AccAddress(contractAddr.Bytes()).String(), minExpRewardOrCommission)
			Expect(err).To(BeNil())

			// contract's accrued rewards amt
			accruedRewardsAmt = rwRes.AmountOf(s.bondDenom).TruncateInt()

			balRes, err := s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			initialBalance = balRes.Balance

			// populate default arguments
			callArgs.MethodName = "testClaimRewards"
		})

		It("should withdraw rewards successfully without origin check", func() {
			balRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			signerInitialBalance := balRes.Balance

			callArgs.Args = []interface{}{contractAddr, uint32(2)}
			txArgs.GasPrice = gasPrice.BigInt()

			logCheckArgs := passCheck.WithExpEvents(distribution.EventTypeClaimRewards)

			res, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			// tx signer should have paid the fees
			fees := gasPrice.Mul(math.NewInt(res.GasUsed))
			balRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			signerFinalBalance := balRes.Balance
			Expect(signerFinalBalance.Amount).To(Equal(signerInitialBalance.Amount.Sub(fees)))

			// contract's balance should increase
			balRes, err = s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			finalBalance := balRes.Balance
			Expect(finalBalance.Amount).To(Equal(initialBalance.Amount.Add(accruedRewardsAmt)), "expected final balance to be greater than initial balance after withdrawing rewards")
		})

		It("should withdraw rewards successfully to a different address without origin check", func() {
			balanceRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			signerInitialBalance := balanceRes.Balance

			balRes, err := s.grpcHandler.GetBalance(differentAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			withdrawerInitialBalance := balRes.Balance

			balRes, err = s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			contractInitialBalance := balRes.Balance

			txArgs.GasPrice = gasPrice.BigInt()

			// Set new withdrawer address for the contract
			setWithdrawCheck := passCheck.WithExpEvents(distribution.EventTypeSetWithdrawAddress)
			res1, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				factory.CallArgs{
					ContractABI: distrCallerContract.ABI,
					MethodName:  "testSetWithdrawAddressFromContract",
					Args:        []interface{}{differentAddr.String()},
				},
				setWithdrawCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			callArgs.Args = []interface{}{contractAddr, uint32(2)}

			logCheckArgs := passCheck.WithExpEvents(distribution.EventTypeClaimRewards)

			rwRes, err := s.grpcHandler.GetDelegationRewards(sdk.AccAddress(contractAddr.Bytes()).String(), s.network.GetValidators()[0].OperatorAddress)
			Expect(err).To(BeNil())
			accruedRewardsAmt = rwRes.Rewards.AmountOf(s.bondDenom).TruncateInt()

			txArgs.GasLimit = 200_000
			res2, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				logCheckArgs,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			// signer balance should decrease - paid for fees
			fees := gasPrice.Mul(math.NewInt(res1.GasUsed)).Add(gasPrice.Mul(math.NewInt(res2.GasUsed)))

			balRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			signerFinalBalance := balRes.Balance
			Expect(signerFinalBalance.Amount).To(Equal(signerInitialBalance.Amount.Sub(fees)), "expected signer's final balance to be less than initial balance after withdrawing rewards")

			// withdrawer balance should increase
			balRes, err = s.grpcHandler.GetBalance(differentAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			withdrawerFinalBalance := balRes.Balance
			Expect(withdrawerFinalBalance.Amount).To(Equal(withdrawerInitialBalance.Amount.Add(accruedRewardsAmt)))

			// contract balance should remain unchanged
			balRes, err = s.grpcHandler.GetBalance(contractAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			contractFinalBalance := balRes.Balance
			Expect(contractFinalBalance.Amount).To(Equal(contractInitialBalance.Amount))
		})
	})

	Context("Forbidden operations", func() {
		It("should revert state: modify withdraw address & then try to withdraw rewards corresponding to another user", func() {
			// check signer address balance should've decreased (fees paid)
			balanceRes, err := s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			initBalanceAmt := balanceRes.Balance.Amount

			_, err = testutils.WaitToAccrueRewards(s.network, s.grpcHandler, s.keyring.GetAccAddr(0).String(), minExpRewardOrCommission)
			Expect(err).To(BeNil())

			callArgs.MethodName = "testRevertState"
			callArgs.Args = []interface{}{
				differentAddr.String(), differentAddr, s.network.GetValidators()[0].OperatorAddress,
			}

			_, _, err = s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			// check withdraw address didn't change
			res, err := s.grpcHandler.GetDelegatorWithdrawAddr(s.keyring.GetAccAddr(0).String())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(res.WithdrawAddress).To(Equal(s.keyring.GetAccAddr(0).String()))

			// check signer address balance should've decreased (fees paid)
			balanceRes, err = s.grpcHandler.GetBalance(s.keyring.GetAccAddr(0), s.bondDenom)
			Expect(err).To(BeNil())
			Expect(balanceRes.Balance.Amount.LTE(initBalanceAmt)).To(BeTrue())

			// check other address' balance remained unchanged
			balanceRes, err = s.grpcHandler.GetBalance(differentAddr.Bytes(), s.bondDenom)
			Expect(err).To(BeNil())
			Expect(balanceRes.Balance.Amount).To(Equal(math.ZeroInt()))
		})

		It("should not allow to call SetWithdrawAddress using delegatecall", func() {
			callArgs.MethodName = "delegateCallSetWithdrawAddress"
			callArgs.Args = []interface{}{s.keyring.GetAddr(0), differentAddr.String()}

			_, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())

			// check withdraw address didn't change
			res, err := s.grpcHandler.GetDelegatorWithdrawAddr(s.keyring.GetAccAddr(0).String())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(res.WithdrawAddress).To(Equal(s.keyring.GetAccAddr(0).String()))
		})

		It("should not allow to call txs (SetWithdrawAddress) using staticcall", func() {
			callArgs.MethodName = "staticCallSetWithdrawAddress"
			callArgs.Args = []interface{}{s.keyring.GetAddr(0), differentAddr.String()}

			_, _, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				execRevertedCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
			Expect(s.network.NextBlock()).To(BeNil())
			// check withdraw address didn't change
			res, err := s.grpcHandler.GetDelegatorWithdrawAddr(s.keyring.GetAccAddr(0).String())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(res.WithdrawAddress).To(Equal(s.keyring.GetAccAddr(0).String()))
		})
	})

	// ===================================
	//				QUERIES
	// ===================================
	Context("Distribution precompile queries", Ordered, func() {
		It("should get validator distribution info", func() {
			// fund validator account to make self-delegation
			err := testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), s.validatorsKeys[0].AccAddr, math.NewInt(1e17))
			Expect(err).To(BeNil())
			// persist changes
			Expect(s.network.NextBlock()).To(BeNil())

			opAddr := s.network.GetValidators()[0].OperatorAddress
			// use the validator priv key
			// make a self delegation
			err = s.factory.Delegate(s.validatorsKeys[0].Priv, opAddr, sdk.NewCoin(s.bondDenom, math.NewInt(1)))
			Expect(err).To(BeNil())
			// persist changes
			Expect(s.network.NextBlock()).To(BeNil())

			callArgs.MethodName = "getValidatorDistributionInfo"
			callArgs.Args = []interface{}{opAddr}
			txArgs.GasLimit = 200_000

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.validatorsKeys[0].Priv,
				txArgs,
				callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var out distribution.ValidatorDistributionInfoOutput
			err = s.precompile.UnpackIntoInterface(&out, distribution.ValidatorDistributionInfoMethod, ethRes.Ret)
			Expect(err).To(BeNil())

			expAddr := s.validatorsKeys[0].AccAddr.String()

			Expect(expAddr).To(Equal(out.DistributionInfo.OperatorAddress))
			Expect(1).To(Equal(len(out.DistributionInfo.Commission)))
			Expect(0).To(Equal(len(out.DistributionInfo.SelfBondRewards)))
		})

		It("should get validator outstanding rewards", func() {
			opAddr := s.network.GetValidators()[0].OperatorAddress
			callArgs.MethodName = "getValidatorOutstandingRewards"
			callArgs.Args = []interface{}{opAddr}

			_, err := testutils.WaitToAccrueRewards(s.network, s.grpcHandler, s.keyring.GetAccAddr(0).String(), minExpRewardOrCommission)
			Expect(err).To(BeNil(), "error while calling the precompile")

			_, ethRes, err := s.factory.CallContractAndCheckLogs(
				s.keyring.GetPrivKey(0),
				txArgs,
				callArgs,
				passCheck,
			)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var rewards []cmn.DecCoin
			err = s.precompile.UnpackIntoInterface(&rewards, distribution.ValidatorOutstandingRewardsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(rewards)).To(Equal(1))
			Expect(uint8(18)).To(Equal(rewards[0].Precision))
			Expect(s.bondDenom).To(Equal(rewards[0].Denom))

			res, err := s.grpcHandler.GetValidatorOutstandingRewards(opAddr)
			Expect(err).To(BeNil())

			expRewardsAmt := res.Rewards.Rewards.AmountOf(s.bondDenom).TruncateInt()
			Expect(expRewardsAmt.IsPositive()).To(BeTrue())
			Expect(rewards[0].Amount).To(Equal(expRewardsAmt.BigInt()))
		})

		Context("get validator commission", func() {
			BeforeEach(func() {
				callArgs.MethodName = "getValidatorCommission"
				callArgs.Args = []interface{}{s.network.GetValidators()[0].OperatorAddress}
			})

			It("should not get commission - validator without commission", func() {
				// fund validator account to claim commission (if any)
				err = testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), s.validatorsKeys[0].AccAddr, math.NewInt(1e18))
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				// withdraw validator commission
				err = s.factory.WithdrawValidatorCommission(s.validatorsKeys[0].Priv)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var commission []cmn.DecCoin
				err = s.precompile.UnpackIntoInterface(&commission, distribution.ValidatorCommissionMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(commission)).To(Equal(1))
				Expect(commission[0].Amount.Int64()).To(Equal(int64(0)))
			})

			It("should get commission - validator with commission", func() {
				_, err = testutils.WaitToAccrueCommission(s.network, s.grpcHandler, s.network.GetValidators()[0].OperatorAddress, minExpRewardOrCommission)
				Expect(err).To(BeNil())

				commRes, err := s.grpcHandler.GetValidatorCommission(s.network.GetValidators()[0].OperatorAddress)
				Expect(err).To(BeNil())

				accruedCommission := commRes.Commission.Commission

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var commission []cmn.DecCoin
				err = s.precompile.UnpackIntoInterface(&commission, distribution.ValidatorCommissionMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(commission)).To(Equal(1))
				Expect(uint8(18)).To(Equal(commission[0].Precision))
				Expect(s.bondDenom).To(Equal(commission[0].Denom))

				accruedCommissionAmt := accruedCommission.AmountOf(s.bondDenom).TruncateInt()

				Expect(commission[0].Amount).To(Equal(accruedCommissionAmt.BigInt()))
			})
		})

		Context("get validator slashing events", Ordered, func() {
			BeforeEach(func() {
				callArgs.MethodName = "getValidatorSlashes"
				callArgs.Args = []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
					uint64(1), uint64(5),
					query.PageRequest{},
				}
			})

			AfterEach(func() {
				// NOTE: The first test case will not have the slashes
				// so keep this in mind when adding/removing new testcases
				s.withValidatorSlashes = true
			})

			AfterAll(func() {
				s.withValidatorSlashes = false
			})

			It("should not get slashing events - validator without slashes", func() {
				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out distribution.ValidatorSlashesOutput
				err = s.precompile.UnpackIntoInterface(&out, distribution.ValidatorSlashesMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(out.Slashes)).To(Equal(0))
			})

			It("should get slashing events - validator with slashes (default pagination)", func() {
				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out distribution.ValidatorSlashesOutput
				err = s.precompile.UnpackIntoInterface(&out, distribution.ValidatorSlashesMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(out.Slashes)).To(Equal(2))
				// expected values according to the values used on test setup (custom genesis)
				for _, s := range out.Slashes {
					Expect(s.Fraction.Value).To(Equal(math.LegacyNewDecWithPrec(5, 2).BigInt()))
					Expect(s.ValidatorPeriod).To(Equal(uint64(1)))
				}
				Expect(uint64(2)).To(Equal(out.PageResponse.Total))
				Expect(out.PageResponse.NextKey).To(BeEmpty())
			})

			It("should get slashing events - validator with slashes w/pagination", func() {
				// set pagination
				callArgs.Args = []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
					uint64(1), uint64(5),
					query.PageRequest{
						Limit:      1,
						CountTotal: true,
					},
				}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out distribution.ValidatorSlashesOutput
				err = s.precompile.UnpackIntoInterface(&out, distribution.ValidatorSlashesMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(out.Slashes)).To(Equal(1))
				Expect(out.Slashes[0].Fraction.Value).To(Equal(math.LegacyNewDecWithPrec(5, 2).BigInt()))
				Expect(out.Slashes[0].ValidatorPeriod).To(Equal(uint64(1)))
				Expect(uint64(2)).To(Equal(out.PageResponse.Total))
				Expect(out.PageResponse.NextKey).NotTo(BeEmpty())
			})
		})

		Context("get delegation rewards", func() {
			BeforeEach(func() {
				callArgs.MethodName = "getDelegationRewards"
				callArgs.Args = []interface{}{s.keyring.GetAddr(0), s.network.GetValidators()[0].OperatorAddress}
			})

			It("should not get rewards - no rewards available", func() {
				// withdraw rewards if available
				err := s.factory.WithdrawDelegationRewards(s.keyring.GetPrivKey(0), s.network.GetValidators()[0].OperatorAddress)
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				// add gas limit to avoid out of gas error
				txArgs.GasLimit = 200_000
				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var rewards []cmn.DecCoin
				err = s.precompile.UnpackIntoInterface(&rewards, distribution.DelegationRewardsMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(rewards)).To(Equal(0))
			})
			It("should get rewards", func() {
				accruedRewards, err := testutils.WaitToAccrueRewards(s.network, s.grpcHandler, s.keyring.GetAccAddr(0).String(), minExpRewardOrCommission)
				Expect(err).To(BeNil())

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var rewards []cmn.DecCoin
				err = s.precompile.UnpackIntoInterface(&rewards, distribution.DelegationRewardsMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(rewards)).To(Equal(1))
				Expect(len(rewards)).To(Equal(1))
				Expect(rewards[0].Denom).To(Equal(s.bondDenom))

				// The accrued rewards are based on 3 equal delegations to the existing 3 validators
				// The query is from only 1 validator, thus, the expected reward
				// for this delegation is totalAccruedRewards / validatorsCount (3)
				accruedRewardsAmt := accruedRewards.AmountOf(s.bondDenom)
				expRewardPerValidator := accruedRewardsAmt.Quo(math.LegacyNewDec(3)).TruncateInt()

				Expect(rewards[0].Amount).To(Equal(expRewardPerValidator.BigInt()))
			})
		})

		Context("get delegator's total rewards", func() {
			BeforeEach(func() {
				callArgs.MethodName = "getDelegationTotalRewards"
				callArgs.Args = []interface{}{s.keyring.GetAddr(0)}
			})

			It("should not get rewards - no rewards available", func() {
				// Create a delegation
				err := s.factory.Delegate(s.keyring.GetPrivKey(1), s.network.GetValidators()[0].OperatorAddress, sdk.NewCoin(s.bondDenom, math.NewInt(1)))
				Expect(err).To(BeNil())
				Expect(s.network.NextBlock()).To(BeNil())

				callArgs.Args = []interface{}{s.keyring.GetAddr(1)}
				txArgs.GasLimit = 200_000 // set gas limit to avoid out of gas error
				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(1),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out distribution.DelegationTotalRewardsOutput
				err = s.precompile.UnpackIntoInterface(&out, distribution.DelegationTotalRewardsMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(out.Rewards)).To(Equal(1))
				Expect(len(out.Rewards[0].Reward)).To(Equal(0))
			})

			It("should get total rewards", func() {
				// wait to get rewards
				accruedRewards, err := testutils.WaitToAccrueRewards(s.network, s.grpcHandler, s.keyring.GetAccAddr(0).String(), minExpRewardOrCommission)
				Expect(err).To(BeNil())

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out distribution.DelegationTotalRewardsOutput

				err = s.precompile.UnpackIntoInterface(&out, distribution.DelegationTotalRewardsMethod, ethRes.Ret)
				Expect(err).To(BeNil())

				// The accrued rewards are based on 3 equal delegations to the existing 3 validators
				accruedRewardsAmt := accruedRewards.AmountOf(s.bondDenom)
				expRewardPerValidator := accruedRewardsAmt.Quo(math.LegacyNewDec(3))

				// the response order may change
				for _, or := range out.Rewards {
					Expect(1).To(Equal(len(or.Reward)))
					Expect(or.Reward[0].Denom).To(Equal(s.bondDenom))
					Expect(or.Reward[0].Amount).To(Equal(expRewardPerValidator.TruncateInt().BigInt()))
				}

				Expect(1).To(Equal(len(out.Total)))
				Expect(out.Total[0].Amount).To(Equal(accruedRewardsAmt.TruncateInt().BigInt()))
			})

			Context("query call with revert - all changes should revert to corresponding stateDB snapshot", func() {
				var (
					reverterContract           evmtypes.CompiledContract
					reverterAddr               common.Address
					testContractInitialBalance = math.NewInt(1000)
				)
				BeforeEach(func() {
					var err error
					// Deploy Reverter contract
					reverterContract, err = contracts.LoadReverterContract()
					Expect(err).To(BeNil(), "error while loading the Reverter contract")

					reverterAddr, err = s.factory.DeployContract(
						s.keyring.GetPrivKey(0),
						evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
						factory.ContractDeploymentData{
							Contract: reverterContract,
						},
					)
					Expect(err).To(BeNil(), "error while deploying the smart contract: %v", err)
					// persist state change
					Expect(s.network.NextBlock()).To(BeNil())

					// send some funds to the Reverter contracts to transfer to the
					// delegator during the tx
					err = testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), reverterAddr.Bytes(), testContractInitialBalance)
					Expect(err).To(BeNil(), "error while funding the smart contract: %v", err)
					Expect(s.network.NextBlock()).To(BeNil())
				})

				It("should revert the execution - Reverter contract", func() {
					args := factory.CallArgs{
						ContractABI: reverterContract.ABI,
						MethodName:  "run",
					}
					_, _, err = s.factory.CallContractAndCheckLogs(
						s.keyring.GetPrivKey(0),
						evmtypes.EvmTxArgs{
							To:       &reverterAddr,
							GasPrice: gasPrice.BigInt(),
						},
						args,
						execRevertedCheck,
					)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					balRes, err := s.grpcHandler.GetBalance(reverterAddr.Bytes(), s.bondDenom)
					Expect(err).To(BeNil())

					contractFinalBalance := balRes.Balance
					Expect(contractFinalBalance.Amount).To(Equal(testContractInitialBalance))
				})
			})
		})

		Context("get all delegator validators", func() {
			BeforeEach(func() {
				callArgs.MethodName = "getDelegatorValidators"
				callArgs.Args = []interface{}{s.keyring.GetAddr(0)}
			})

			It("should get all validators a delegator has delegated to", func() {
				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var validators []string
				err = s.precompile.UnpackIntoInterface(&validators, distribution.DelegatorValidatorsMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(3).To(Equal(len(validators)))
			})
		})

		Context("get withdraw address", func() {
			BeforeEach(func() {
				callArgs.MethodName = "getDelegatorWithdrawAddress"
				callArgs.Args = []interface{}{s.keyring.GetAddr(0)}
			})

			It("should get withdraw address", func() {
				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				withdrawAddr, err := s.precompile.Unpack(distribution.DelegatorWithdrawAddressMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				// get the bech32 encoding
				expAddr := sdk.AccAddress(s.keyring.GetAddr(0).Bytes())
				Expect(withdrawAddr[0]).To(Equal(expAddr.String()))
			})

			It("should call GetWithdrawAddress using staticcall", func() {
				callArgs.MethodName = "staticCallGetWithdrawAddress"
				callArgs.Args = []interface{}{s.keyring.GetAddr(0)}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				withdrawAddr, err := s.precompile.Unpack(distribution.DelegatorWithdrawAddressMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				// get the bech32 encoding
				expAddr := sdk.AccAddress(s.keyring.GetAddr(0).Bytes())
				Expect(withdrawAddr[0]).To(ContainSubstring(expAddr.String()))
			})
		})
	})
})
