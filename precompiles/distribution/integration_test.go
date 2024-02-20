// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package distribution_test

import (
	"fmt"
	"math/big"

	"github.com/evmos/evmos/v16/utils"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v16/precompiles/common"
	"github.com/evmos/evmos/v16/precompiles/distribution"
	"github.com/evmos/evmos/v16/precompiles/testutil"
	"github.com/evmos/evmos/v16/precompiles/testutil/contracts"
	evmosutil "github.com/evmos/evmos/v16/testutil"
	testutiltx "github.com/evmos/evmos/v16/testutil/tx"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

// General variables used for integration tests
var (
	// differentAddr is an address generated for testing purposes that e.g. raises the different origin error
	differentAddr = testutiltx.GenerateAddress()
	// expRewardAmt is the expected amount of rewards
	expRewardAmt = big.NewInt(2000000000000000000)
	// gasPrice is the gas price used for the transactions
	gasPrice = big.NewInt(1e9)
	// defaultCallArgs  are the default arguments for calling the smart contract
	//
	// NOTE: this has to be populated in a BeforeEach block because the contractAddr would otherwise be a nil address.
	defaultCallArgs contracts.CallArgs

	// defaultLogCheck instantiates a log check arguments struct with the precompile ABI events populated.
	defaultLogCheck testutil.LogCheckArgs
	// differentOriginCheck defines the arguments to check if the precompile returns different origin error
	differentOriginCheck testutil.LogCheckArgs
	// passCheck defines the arguments to check if the precompile returns no error
	passCheck testutil.LogCheckArgs
	// outOfGasCheck defines the arguments to check if the precompile returns out of gas error
	outOfGasCheck testutil.LogCheckArgs
)

var _ = Describe("Calling distribution precompile from EOA", func() {
	BeforeEach(func() {
		s.SetupTest()

		// set the default call arguments
		defaultCallArgs = contracts.CallArgs{
			ContractAddr: s.precompile.Address(),
			ContractABI:  s.precompile.ABI,
			PrivKey:      s.keyring.GetPrivKey(0),
		}

		defaultLogCheck = testutil.LogCheckArgs{
			ABIEvents: s.precompile.ABI.Events,
		}
		differentOriginCheck = defaultLogCheck.WithErrContains(cmn.ErrDifferentOrigin, s.keyring.GetAddr(0), differentAddr)
		passCheck = defaultLogCheck.WithExpPass(true)
		outOfGasCheck = defaultLogCheck.WithErrContains(vm.ErrOutOfGas.Error())
	})

	// =====================================
	// 				TRANSACTIONS
	// =====================================
	Describe("Execute SetWithdrawAddress transaction", func() {
		const method = distribution.SetWithdrawAddressMethod
		// defaultSetWithdrawArgs are the default arguments to set the withdraw address
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key otherwise is not yet initialized.
		var defaultSetWithdrawArgs contracts.CallArgs

		BeforeEach(func() {
			// set the default call arguments
			defaultSetWithdrawArgs = defaultCallArgs.WithMethodName(method)
		})

		It("should return error if the provided gasLimit is too low", func() {
			setWithdrawArgs := defaultSetWithdrawArgs.
				WithGasLimit(30000).
				WithArgs(s.keyring.GetAddr(0), differentAddr.String())

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, setWithdrawArgs, outOfGasCheck)
			Expect(err).To(HaveOccurred(), "error while calling the precompile")
			Expect(err.Error()).To(ContainSubstring("out of gas"), "expected out of gas error")

			// withdraw address should remain unchanged
			withdrawAddr, err := s.network.App.DistrKeeper.GetDelegatorWithdrawAddr(s.network.GetContext(), s.keyring.GetAddr(0).Bytes())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(withdrawAddr.Bytes()).To(Equal(s.keyring.GetAddr(0).Bytes()), "expected withdraw address to remain unchanged")
		})

		It("should return error if the origin is different than the delegator", func() {
			setWithdrawArgs := defaultSetWithdrawArgs.WithArgs(differentAddr, s.keyring.GetAddr(0).String())

			withdrawAddrSetCheck := defaultLogCheck.WithErrContains(cmn.ErrDifferentOrigin, s.keyring.GetAddr(0).String(), differentAddr.String())

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, setWithdrawArgs, withdrawAddrSetCheck)
			Expect(err).To(HaveOccurred(), "error while calling the precompile")
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(cmn.ErrDifferentOrigin, s.keyring.GetAddr(0), differentAddr)), "expected different origin error")
		})

		It("should set withdraw address", func() {
			// initially, withdraw address should be same as address
			withdrawAddr, err := s.network.App.DistrKeeper.GetDelegatorWithdrawAddr(s.network.GetContext(), s.keyring.GetAddr(0).Bytes())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(withdrawAddr.Bytes()).To(Equal(s.keyring.GetAddr(0).Bytes()))

			setWithdrawArgs := defaultSetWithdrawArgs.WithArgs(s.keyring.GetAddr(0), differentAddr.String())

			withdrawAddrSetCheck := passCheck.
				WithExpEvents(distribution.EventTypeSetWithdrawAddress)

			_, _, err = contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, setWithdrawArgs, withdrawAddrSetCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			// withdraw should be updated
			withdrawAddr, err = s.network.App.DistrKeeper.GetDelegatorWithdrawAddr(s.network.GetContext(), s.keyring.GetAddr(0).Bytes())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(withdrawAddr.Bytes()).To(Equal(differentAddr.Bytes()), "expected different withdraw address")
		})
	})

	Describe("Execute WithdrawDelegatorRewards transaction", func() {
		// defaultWithdrawRewardsArgs are the default arguments to withdraw rewards
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key otherwise is not yet initialized.
		var defaultWithdrawRewardsArgs contracts.CallArgs

		BeforeEach(func() {
			// set the default call arguments
			defaultWithdrawRewardsArgs = defaultCallArgs.WithMethodName(distribution.WithdrawDelegatorRewardsMethod)
			// FIXME will need to use the WaitToAccrueRewards func
			// // s.prepareStakingRewards(stakingRewards{s.keyring.GetAddr(0).Bytes(), s.network.GetValidators()[0], rewards})
		})

		It("should return error if the origin is different than the delegator", func() {
			withdrawRewardsArgs := defaultWithdrawRewardsArgs.WithArgs(differentAddr, s.network.GetValidators()[0].OperatorAddress)

			withdrawalCheck := defaultLogCheck.WithErrContains(cmn.ErrDifferentOrigin, s.keyring.GetAddr(0).String(), differentAddr.String())

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, withdrawRewardsArgs, withdrawalCheck)
			Expect(err).To(HaveOccurred(), "error while calling the precompile")
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(cmn.ErrDifferentOrigin, s.keyring.GetAddr(0), differentAddr)), "expected different origin error")
		})

		It("should withdraw delegation rewards", func() {
			// get initial balance
			initialBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)
			Expect(initialBalance.Amount).To(Equal(initialBalance.Amount))

			withdrawRewardsArgs := defaultWithdrawRewardsArgs.
				WithArgs(s.keyring.GetAddr(0), s.network.GetValidators()[0].OperatorAddress).
				WithGasPrice(gasPrice)

			withdrawalCheck := passCheck.
				WithExpEvents(distribution.EventTypeWithdrawDelegatorRewards)

			res, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, withdrawRewardsArgs, withdrawalCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var rewards []cmn.Coin
			err = s.precompile.UnpackIntoInterface(&rewards, distribution.WithdrawDelegatorRewardsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(rewards)).To(Equal(1))
			Expect(rewards[0].Denom).To(Equal(s.bondDenom))
			Expect(rewards[0].Amount).To(Equal(expRewardAmt))

			// check that the rewards were added to the balance
			finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)
			fees := gasPrice.Int64() * res.GasUsed
			expFinal := initialBalance.Amount.Int64() + expRewardAmt.Int64() - fees
			Expect(finalBalance.Amount.Equal(math.NewInt(expFinal))).To(BeTrue(), "expected final balance to be equal to initial balance + rewards - fees")
		})
	})

	Describe("Validator Commission: Execute WithdrawValidatorCommission tx", func() {
		var (
			// defaultWithdrawCommissionArgs are the default arguments to withdraw commission
			//
			// NOTE: this has to be populated in the BeforeEach block because the private key otherwise is not yet initialized.
			defaultWithdrawCommissionArgs contracts.CallArgs

			// expCommAmt is the expected commission amount
			expCommAmt = big.NewInt(1)
			// commDec is the commission rate
			commDec  = math.LegacyNewDec(1)
			valAddr  sdk.ValAddress
			stakeAmt math.Int
		)

		BeforeEach(func() {
			// set the default call arguments
			defaultWithdrawCommissionArgs = defaultCallArgs.WithMethodName(
				distribution.WithdrawValidatorCommissionMethod,
			)

			// create a validator with s.keyring.GetAddr(0) and s.keyring.GetPrivKey(0) because this account is
			// used for signing txs
			stakeAmt = math.NewInt(100)
			testutil.CreateValidator(s.network.GetContext(), s.T(), s.keyring.GetPrivKey(0).PubKey(), s.network.App.StakingKeeper, stakeAmt)

			// set some validator commission
			valAddr = s.keyring.GetAddr(0).Bytes()
			val, err := s.network.App.StakingKeeper.Validator(s.network.GetContext(), valAddr)
			Expect(err).To(BeNil(), "error while calling the precompile")
			valCommission := sdk.DecCoins{sdk.NewDecCoinFromDec(s.bondDenom, commDec)}

			s.network.App.DistrKeeper.SetValidatorAccumulatedCommission(s.network.GetContext(), valAddr, distrtypes.ValidatorAccumulatedCommission{Commission: valCommission})
			s.network.App.DistrKeeper.AllocateTokensToValidator(s.network.GetContext(), val, sdk.DecCoins{sdk.NewDecCoin(s.bondDenom, stakeAmt)})
		})

		It("should return error if the provided gasLimit is too low", func() {
			withdrawCommissionArgs := defaultWithdrawCommissionArgs.
				WithGasLimit(50000).
				WithArgs(valAddr.String())

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, withdrawCommissionArgs, outOfGasCheck)
			Expect(err).To(HaveOccurred(), "error while calling the precompile")
			Expect(err.Error()).To(ContainSubstring("out of gas"), "expected out of gas error")
		})

		It("should return error if the origin is different than the validator", func() {
			withdrawCommissionArgs := defaultWithdrawCommissionArgs.WithArgs(s.network.GetValidators()[0].OperatorAddress)
			validatorHexAddr := common.BytesToAddress([]byte(s.network.GetValidators()[0].GetOperator()))

			withdrawalCheck := defaultLogCheck.WithErrContains(cmn.ErrDifferentOrigin, s.keyring.GetAddr(0).String(), validatorHexAddr.String())

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, withdrawCommissionArgs, withdrawalCheck)
			Expect(err).To(HaveOccurred(), "error while calling the precompile")
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(cmn.ErrDifferentOrigin, s.keyring.GetAddr(0), validatorHexAddr)), "expected different origin error")
		})

		It("should withdraw validator commission", func() {
			// initial balance should be the initial amount minus the staked amount used to create the validator
			initialBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)
			Expect(initialBalance.Amount).To(Equal(math.NewInt(4999999999999999900)))

			withdrawCommissionArgs := defaultWithdrawCommissionArgs.
				WithArgs(valAddr.String()).
				WithGasPrice(gasPrice)

			withdrawalCheck := passCheck.
				WithExpEvents(distribution.EventTypeWithdrawValidatorCommission)

			res, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, withdrawCommissionArgs, withdrawalCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var comm []cmn.Coin
			err = s.precompile.UnpackIntoInterface(&comm, distribution.WithdrawValidatorCommissionMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(comm)).To(Equal(1))
			Expect(comm[0].Denom).To(Equal(s.bondDenom))
			Expect(comm[0].Amount).To(Equal(expCommAmt))

			finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)
			fees := gasPrice.Int64() * res.GasUsed
			expFinal := initialBalance.Amount.Int64() + expCommAmt.Int64() - fees
			Expect(finalBalance.Amount.Equal(math.NewInt(expFinal))).To(BeTrue(), "expected final balance to be equal to the final balance after withdrawing commission")
		})
	})

	Describe("Execute ClaimRewards transaction", func() {
		// defaultWithdrawRewardsArgs are the default arguments to withdraw rewards
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key otherwise is not yet initialized.
		var defaultClaimRewardsArgs contracts.CallArgs
		startingBalance := math.NewInt(5e18)
		expectedBalance := math.NewInt(8999665039062500000)

		BeforeEach(func() {
			// set the default call arguments
			defaultClaimRewardsArgs = defaultCallArgs.WithMethodName(distribution.ClaimRewardsMethod)
			// s.prepareStakingRewards(stakingRewards{s.keyring.GetAddr(0).Bytes(), s.network.GetValidators()[0], rewards})
			// s.prepareStakingRewards(stakingRewards{s.keyring.GetAddr(0).Bytes(), s.network.GetValidators()[1], rewards})
		})

		It("should return err if the origin is different than the delegator", func() {
			claimRewardsArgs := defaultClaimRewardsArgs.WithArgs(differentAddr, uint32(1))

			claimRewardsCheck := defaultLogCheck.WithErrContains(cmn.ErrDifferentOrigin, s.keyring.GetAddr(0).String(), differentAddr.String())

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, claimRewardsArgs, claimRewardsCheck)
			Expect(err).To(HaveOccurred(), "error while calling the precompile")
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(cmn.ErrDifferentOrigin, s.keyring.GetAddr(0), differentAddr)), "expected different origin error")
		})

		It("should claim all rewards from all validators", func() {
			initialBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)
			Expect(initialBalance.Amount).To(Equal(startingBalance))

			claimRewardsArgs := defaultClaimRewardsArgs.WithArgs(s.keyring.GetAddr(0), uint32(2))
			claimRewardsCheck := passCheck.WithExpEvents(distribution.EventTypeClaimRewards)

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, claimRewardsArgs, claimRewardsCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			// check that the rewards were added to the balance
			finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)
			Expect(finalBalance.Amount.Equal(expectedBalance)).To(BeTrue(), "expected final balance to be equal to initial balance + rewards - fees")
		})
	})

	// =====================================
	// 				QUERIES
	// =====================================
	Describe("Execute queries", func() {
		It("should get validator distribution info - validatorDistributionInfo query", func() {
			// FIXME this could be broken
			// One way is to use the accKeeper.AddressCodec().StringToBytes(string)
			addr := sdk.AccAddress(s.network.GetValidators()[0].GetOperator())
			// fund validator account to make self-delegation
			err := evmosutil.FundAccountWithBaseDenom(s.network.GetContext(), s.network.App.BankKeeper, addr, 10)
			Expect(err).To(BeNil())
			// make a self delegation
			_, err = s.network.App.StakingKeeper.Delegate(s.network.GetContext(), addr, math.NewInt(1), stakingtypes.Unspecified, s.network.GetValidators()[0], true)
			Expect(err).To(BeNil())

			valDistArgs := defaultCallArgs.
				WithMethodName(distribution.ValidatorDistributionInfoMethod).
				WithArgs(s.network.GetValidators()[0].OperatorAddress)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, valDistArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var out distribution.ValidatorDistributionInfoOutput
			err = s.precompile.UnpackIntoInterface(&out, distribution.ValidatorDistributionInfoMethod, ethRes.Ret)
			Expect(err).To(BeNil())

			expAddr := s.network.GetValidators()[0].GetOperator()
			Expect(expAddr).To(Equal(out.DistributionInfo.OperatorAddress))
			Expect(0).To(Equal(len(out.DistributionInfo.Commission)))
			Expect(0).To(Equal(len(out.DistributionInfo.SelfBondRewards)))
		})

		It("should get validator outstanding rewards - validatorOutstandingRewards query", func() { //nolint:dupl
			valRewards := sdk.DecCoins{sdk.NewDecCoinFromDec(s.bondDenom, math.LegacyNewDec(1))}
			// set outstanding rewards
			err := s.network.App.DistrKeeper.SetValidatorOutstandingRewards(s.network.GetContext(), sdk.ValAddress(s.network.GetValidators()[0].GetOperator()), distrtypes.ValidatorOutstandingRewards{Rewards: valRewards})
			Expect(err).To(BeNil(), "error while calling the precompile")

			valOutRewardsArgs := defaultCallArgs.
				WithMethodName(distribution.ValidatorOutstandingRewardsMethod).
				WithArgs(s.network.GetValidators()[0].OperatorAddress)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, valOutRewardsArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var rewards []cmn.DecCoin
			err = s.precompile.UnpackIntoInterface(&rewards, distribution.ValidatorOutstandingRewardsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(rewards)).To(Equal(1))
			Expect(uint8(18)).To(Equal(rewards[0].Precision))
			Expect(s.bondDenom).To(Equal(rewards[0].Denom))
			Expect(expValAmount).To(Equal(rewards[0].Amount.Int64()))
		})

		It("should get validator commission - validatorCommission query", func() { //nolint:dupl
			// set commission
			valCommission := sdk.DecCoins{sdk.NewDecCoinFromDec(s.bondDenom, math.LegacyNewDec(1))}
			err := s.network.App.DistrKeeper.SetValidatorAccumulatedCommission(s.network.GetContext(), sdk.ValAddress(s.network.GetValidators()[0].GetOperator()), distrtypes.ValidatorAccumulatedCommission{Commission: valCommission})
			Expect(err).To(BeNil(), "error while calling the precompile")

			valCommArgs := defaultCallArgs.
				WithMethodName(distribution.ValidatorCommissionMethod).
				WithArgs(s.network.GetValidators()[0].OperatorAddress)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, valCommArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var commission []cmn.DecCoin
			err = s.precompile.UnpackIntoInterface(&commission, distribution.ValidatorCommissionMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(commission)).To(Equal(1))
			Expect(uint8(18)).To(Equal(commission[0].Precision))
			Expect(s.bondDenom).To(Equal(commission[0].Denom))
			Expect(expValAmount).To(Equal(commission[0].Amount.Int64()))
		})

		Context("validatorSlashes query query", func() {
			It("should get validator slashing events (default pagination)", func() {
				// set slash event
				slashEvent := distrtypes.ValidatorSlashEvent{ValidatorPeriod: 1, Fraction: math.LegacyNewDec(5)}
				err := s.network.App.DistrKeeper.SetValidatorSlashEvent(s.network.GetContext(), sdk.ValAddress(s.network.GetValidators()[0].GetOperator()), 2, 1, slashEvent)
				Expect(err).To(BeNil(), "error while calling the precompile")

				valSlashArgs := defaultCallArgs.
					WithMethodName(distribution.ValidatorSlashesMethod).
					WithArgs(
						s.network.GetValidators()[0].OperatorAddress,
						uint64(1), uint64(5),
						query.PageRequest{},
					)

				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, valSlashArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the precompile")

				var out distribution.ValidatorSlashesOutput
				err = s.precompile.UnpackIntoInterface(&out, distribution.ValidatorSlashesMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(out.Slashes)).To(Equal(1))
				Expect(slashEvent.Fraction.BigInt()).To(Equal(out.Slashes[0].Fraction.Value))
				Expect(slashEvent.ValidatorPeriod).To(Equal(out.Slashes[0].ValidatorPeriod))
				Expect(uint64(1)).To(Equal(out.PageResponse.Total))
				Expect(out.PageResponse.NextKey).To(BeEmpty())
			})

			It("should get validator slashing events - query w/pagination limit = 1)", func() {
				// TODO fixme, this should be done with a custom genesis
				// set 2 slashing events for validator[0]
				// slashEvent := s.setupValidatorSlashes(sdk.ValAddress(s.network.GetValidators()[0].GetOperator()), 2)

				valSlashArgs := defaultCallArgs.
					WithMethodName(distribution.ValidatorSlashesMethod).
					WithArgs(
						s.network.GetValidators()[0].OperatorAddress,
						uint64(1), uint64(5),
						query.PageRequest{
							Limit:      1,
							CountTotal: true,
						},
					)

				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, valSlashArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the precompile")

				var out distribution.ValidatorSlashesOutput
				err = s.precompile.UnpackIntoInterface(&out, distribution.ValidatorSlashesMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				// TODO FIXME
				// Expect(len(out.Slashes)).To(Equal(1))
				// Expect(slashEvent.Fraction.BigInt()).To(Equal(out.Slashes[0].Fraction.Value))
				// Expect(slashEvent.ValidatorPeriod).To(Equal(out.Slashes[0].ValidatorPeriod))
				// total slashes count is 2
				Expect(uint64(2)).To(Equal(out.PageResponse.Total))
				Expect(out.PageResponse.NextKey).NotTo(BeEmpty())
			})
		})

		It("should get delegation rewards - delegationRewards query", func() {
			// TODO FIXME
			// // s.prepareStakingRewards(stakingRewards{s.keyring.GetAddr(0).Bytes(), s.network.GetValidators()[0], rewards})

			delRewardsArgs := defaultCallArgs.
				WithMethodName(distribution.DelegationRewardsMethod).
				WithArgs(s.keyring.GetAddr(0), s.network.GetValidators()[0].OperatorAddress)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, delRewardsArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var rewards []cmn.DecCoin
			err = s.precompile.UnpackIntoInterface(&rewards, distribution.DelegationRewardsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(len(rewards)).To(Equal(1))
			Expect(rewards[0].Denom).To(Equal(s.bondDenom))
			Expect(rewards[0].Amount.Int64()).To(Equal(testRewards))
		})

		It("should get delegators's total rewards - delegationTotalRewards query", func() {
			// set rewards
			// s.prepareStakingRewards(stakingRewards{s.keyring.GetAddr(0).Bytes(), s.network.GetValidators()[0], rewards})

			delTotalRewardsArgs := defaultCallArgs.
				WithMethodName(distribution.DelegationTotalRewardsMethod).
				WithArgs(s.keyring.GetAddr(0))

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, delTotalRewardsArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var (
				out distribution.DelegationTotalRewardsOutput
				i   int
			)
			err = s.precompile.UnpackIntoInterface(&out, distribution.DelegationTotalRewardsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(2).To(Equal(len(out.Rewards)))

			// the response order may change
			if out.Rewards[0].ValidatorAddress == s.network.GetValidators()[0].OperatorAddress {
				Expect(s.network.GetValidators()[0].OperatorAddress).To(Equal(out.Rewards[0].ValidatorAddress))
				Expect(s.network.GetValidators()[1].OperatorAddress).To(Equal(out.Rewards[1].ValidatorAddress))
				Expect(0).To(Equal(len(out.Rewards[1].Reward)))
			} else {
				i = 1
				Expect(s.network.GetValidators()[0].OperatorAddress).To(Equal(out.Rewards[1].ValidatorAddress))
				Expect(s.network.GetValidators()[1].OperatorAddress).To(Equal(out.Rewards[0].ValidatorAddress))
				Expect(0).To(Equal(len(out.Rewards[0].Reward)))
			}

			// only validator[i] has rewards
			Expect(1).To(Equal(len(out.Rewards[i].Reward)))
			Expect(s.bondDenom).To(Equal(out.Rewards[i].Reward[0].Denom))
			Expect(uint8(math.LegacyPrecision)).To(Equal(out.Rewards[i].Reward[0].Precision))
			Expect(testRewards).To(Equal(out.Rewards[i].Reward[0].Amount.Int64()))

			Expect(1).To(Equal(len(out.Total)))
			Expect(testRewards).To(Equal(out.Total[0].Amount.Int64()))
		})

		It("should get all validators a delegators has delegated to - delegatorValidators query", func() {
			delValArgs := defaultCallArgs.
				WithMethodName(distribution.DelegatorValidatorsMethod).
				WithArgs(s.keyring.GetAddr(0))

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, delValArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var validators []string
			err = s.precompile.UnpackIntoInterface(&validators, distribution.DelegatorValidatorsMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			Expect(2).To(Equal(len(validators)))

			// the response order may change
			if validators[0] == s.network.GetValidators()[0].OperatorAddress {
				Expect(s.network.GetValidators()[0].OperatorAddress).To(Equal(validators[0]))
				Expect(s.network.GetValidators()[1].OperatorAddress).To(Equal(validators[1]))
			} else {
				Expect(s.network.GetValidators()[1].OperatorAddress).To(Equal(validators[0]))
				Expect(s.network.GetValidators()[0].OperatorAddress).To(Equal(validators[1]))
			}
		})

		It("should get withdraw address - delegatorWithdrawAddress query", func() {
			// set the withdraw address
			err := s.network.App.DistrKeeper.SetWithdrawAddr(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), differentAddr.Bytes())
			Expect(err).To(BeNil())

			delWithdrawAddrArgs := defaultCallArgs.
				WithMethodName(distribution.DelegatorWithdrawAddressMethod).
				WithArgs(s.keyring.GetAddr(0))

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, delWithdrawAddrArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			withdrawAddr, err := s.precompile.Unpack(distribution.DelegatorWithdrawAddressMethod, ethRes.Ret)
			Expect(err).To(BeNil())
			// get the bech32 encoding
			expAddr := sdk.AccAddress(differentAddr.Bytes())
			Expect(withdrawAddr[0]).To(Equal(expAddr.String()))
		})
	})
})

var _ = Describe("Calling distribution precompile from another contract", func() {
	var (
		// initBalanceAmt is the initial balance for testing
		initBalanceAmt = math.NewInt(5000000000000000000)

		// contractAddr is the address of the smart contract that will be deployed
		contractAddr common.Address
		// err is a basic error type
		err error

		// execRevertedCheck defines the default log checking arguments which includes the
		// standard revert message.
		execRevertedCheck testutil.LogCheckArgs
	)

	BeforeEach(func() {
		s.SetupTest()
		// TODO FIXME
		// contractAddr, err = s.factory.DeployContract(contracts.DistributionCallerContract)
		Expect(err).To(BeNil(), "error while deploying the smart contract: %v", err)

		// NextBlock the smart contract
		s.network.NextBlock()

		// check contract was correctly deployed
		cAcc := s.network.App.EvmKeeper.GetAccount(s.network.GetContext(), contractAddr)
		Expect(cAcc).ToNot(BeNil(), "contract account should exist")
		Expect(cAcc.IsContract()).To(BeTrue(), "account should be a contract")

		// populate default call args
		defaultCallArgs = contracts.CallArgs{
			ContractAddr: contractAddr,
			ContractABI:  contracts.DistributionCallerContract.ABI,
			PrivKey:      s.keyring.GetPrivKey(0),
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
		var (
			// defaultSetWithdrawAddrArgs are the default arguments for the set withdraw address call
			//
			// NOTE: this has to be populated in a BeforeEach block because the contractAddr would otherwise be a nil address.
			defaultSetWithdrawAddrArgs contracts.CallArgs
			// newWithdrawer is the address to set the withdraw address to
			newWithdrawer = differentAddr
		)

		BeforeEach(func() {
			// withdraw address should be same as address
			withdrawer, err := s.network.App.DistrKeeper.GetDelegatorWithdrawAddr(s.network.GetContext(), s.keyring.GetAddr(0).Bytes())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(withdrawer.Bytes()).To(Equal(s.keyring.GetAddr(0).Bytes()))

			// populate default arguments
			defaultSetWithdrawAddrArgs = defaultCallArgs.WithMethodName(
				"testSetWithdrawAddress",
			)
		})

		It("should set withdraw address successfully", func() {
			setWithdrawAddrArgs := defaultSetWithdrawAddrArgs.WithArgs(
				s.keyring.GetAddr(0), newWithdrawer.String(),
			)

			setWithdrawCheck := passCheck.WithExpEvents(distribution.EventTypeSetWithdrawAddress)

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, setWithdrawAddrArgs, setWithdrawCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			withdrawer, err := s.network.App.DistrKeeper.GetDelegatorWithdrawAddr(s.network.GetContext(), s.keyring.GetAddr(0).Bytes())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(withdrawer.Bytes()).To(Equal(newWithdrawer.Bytes()))
		})
	})

	Context("setWithdrawerAddress with contract as delegator", func() {
		var (
			// defaultSetWithdrawAddrArgs are the default arguments for the set withdraw address call
			//
			// NOTE: this has to be populated in a BeforeEach block because the contractAddr would otherwise be a nil address.
			defaultSetWithdrawAddrArgs contracts.CallArgs
			// newWithdrawer is the address to set the withdraw address to
			newWithdrawer = differentAddr
		)

		BeforeEach(func() {
			// withdraw address should be same as address
			withdrawer, err := s.network.App.DistrKeeper.GetDelegatorWithdrawAddr(s.network.GetContext(), s.keyring.GetAddr(0).Bytes())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(withdrawer.Bytes()).To(Equal(s.keyring.GetAddr(0).Bytes()))

			// populate default arguments
			defaultSetWithdrawAddrArgs = defaultCallArgs.WithMethodName(
				"testSetWithdrawAddressFromContract",
			)
		})

		It("should set withdraw address successfully without origin check", func() {
			setWithdrawAddrArgs := defaultSetWithdrawAddrArgs.WithArgs(newWithdrawer.String())

			setWithdrawCheck := passCheck.WithExpEvents(distribution.EventTypeSetWithdrawAddress)

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, setWithdrawAddrArgs, setWithdrawCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			withdrawer, err := s.network.App.DistrKeeper.GetDelegatorWithdrawAddr(s.network.GetContext(), contractAddr.Bytes())
			Expect(err).To(BeNil(), "error while calling GetDelegatorWithdrawAddr: %v", err)
			Expect(withdrawer.Bytes()).To(Equal(newWithdrawer.Bytes()))
		})
	})

	Context("withdrawDelegatorRewards", func() {
		var (
			// defaultWithdrawDelRewardsArgs are the default arguments for the withdraw delegator rewards call
			//
			// NOTE: this has to be populated in a BeforeEach block because the contractAddr would otherwise be a nil address.
			defaultWithdrawDelRewardsArgs contracts.CallArgs
			// initialBalance is the initial balance of the delegator
			initialBalance sdk.Coin
		)

		BeforeEach(func() {
			// set some rewards for s.keyring.GetAddr(0) & another address
			// s.prepareStakingRewards([]stakingRewards{
			// 	{s.keyring.GetAddr(0).Bytes(), s.network.GetValidators()[0], rewards},
			// 	{differentAddr.Bytes(), s.network.GetValidators()[0], rewards},
			// }...)

			initialBalance = s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)

			// populate default arguments
			defaultWithdrawDelRewardsArgs = defaultCallArgs.WithMethodName(
				"testWithdrawDelegatorRewards",
			)
		})

		It("should not withdraw rewards when sending from a different address", func() {
			withdrawDelRewardsArgs := defaultWithdrawDelRewardsArgs.WithArgs(
				differentAddr, s.network.GetValidators()[0].OperatorAddress,
			)

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, withdrawDelRewardsArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

			// balance should be equal as initial balance or less (because of fees)
			finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)
			Expect(finalBalance.Amount.Uint64() <= initialBalance.Amount.Uint64()).To(BeTrue())

			// differentAddr balance should remain unchanged
			differentAddrFinalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), differentAddr.Bytes(), s.bondDenom)
			Expect(differentAddrFinalBalance.Amount).To(Equal(math.ZeroInt()))
		})

		It("should withdraw rewards successfully", func() {
			withdrawDelRewardsArgs := defaultWithdrawDelRewardsArgs.WithArgs(
				s.keyring.GetAddr(0), s.network.GetValidators()[0].OperatorAddress,
			)

			logCheckArgs := passCheck.
				WithExpEvents(distribution.EventTypeWithdrawDelegatorRewards)

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, withdrawDelRewardsArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			// balance should remain unchanged
			finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)
			Expect(finalBalance.Amount.GT(initialBalance.Amount)).To(BeTrue(), "expected final balance to be greater than initial balance after withdrawing rewards")
		})

		It("should withdraw rewards successfully to the new withdrawer address", func() {
			initialBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), differentAddr.Bytes(), s.bondDenom)
			// Set new withdrawer address
			err := s.network.App.DistrKeeper.SetWithdrawAddr(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), differentAddr.Bytes())
			Expect(err).To(BeNil())

			withdrawDelRewardsArgs := defaultWithdrawDelRewardsArgs.WithArgs(
				s.keyring.GetAddr(0), s.network.GetValidators()[0].OperatorAddress,
			)

			logCheckArgs := passCheck.
				WithExpEvents(distribution.EventTypeWithdrawDelegatorRewards)

			_, _, err = contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, withdrawDelRewardsArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			// should increase balance by rewards
			finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), differentAddr.Bytes(), s.bondDenom)
			Expect(finalBalance.Amount.GT(initialBalance.Amount)).To(BeTrue(), "expected final balance to be greater than initial balance after withdrawing rewards")
		})
	})

	Context("withdrawDelegatorRewards with contract as delegator", func() {
		var (
			// defaultWithdrawDelRewardsArgs are the default arguments for the withdraw delegator rewards call
			//
			// NOTE: this has to be populated in a BeforeEach block because the contractAddr would otherwise be a nil address.
			defaultWithdrawDelRewardsArgs contracts.CallArgs
			// initialBalance is the initial balance of the delegator
			initialBalance sdk.Coin
		)

		BeforeEach(func() {
			// set some rewards for s.keyring.GetAddr(0) & another address
			// s.prepareStakingRewards([]stakingRewards{
			// 	{
			// 		Delegator: contractAddr.Bytes(),
			// 		Validator: s.network.GetValidators()[0],
			// 		RewardAmt: rewards,
			// 	},
			// }...)

			initialBalance = s.network.App.BankKeeper.GetBalance(s.network.GetContext(), contractAddr.Bytes(), s.bondDenom)

			// populate default arguments
			defaultWithdrawDelRewardsArgs = defaultCallArgs.WithMethodName(
				"testWithdrawDelegatorRewardsFromContract",
			)
		})

		It("should withdraw rewards successfully without origin check", func() {
			withdrawDelRewardsArgs := defaultWithdrawDelRewardsArgs.WithArgs(s.network.GetValidators()[0].OperatorAddress)

			logCheckArgs := passCheck.WithExpEvents(distribution.EventTypeWithdrawDelegatorRewards)

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, withdrawDelRewardsArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			// balance should increase
			finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), contractAddr.Bytes(), s.bondDenom)
			Expect(finalBalance.Amount.GT(initialBalance.Amount)).To(BeTrue(), "expected final balance to be greater than initial balance after withdrawing rewards")
		})
	})

	Context("withdrawValidatorCommission", func() {
		var (
			// defaultWithdrawValCommArgs are the default arguments for the withdraw validator commission call
			//
			// NOTE: this has to be populated in a BeforeEach block because the contractAddr would otherwise be a nil address.
			defaultWithdrawValCommArgs contracts.CallArgs
			// commDec is the commission rate of the validator
			commDec = math.LegacyNewDec(1)
			// valAddr is the address of the validator
			valAddr sdk.ValAddress
			// initialBalance is the initial balance of the delegator
			initialBalance sdk.Coin
		)

		BeforeEach(func() {
			// create a validator with s.keyring.GetAddr(0) because is the address
			// used for signing txs
			valAddr = s.keyring.GetAddr(0).Bytes()
			stakeAmt := math.NewInt(100)
			testutil.CreateValidator(s.network.GetContext(), s.T(), s.keyring.GetPrivKey(0).PubKey(), s.network.App.StakingKeeper, stakeAmt)

			// set some commissions to validators
			var valAddresses []sdk.ValAddress
			valAddresses = append(
				valAddresses,
				valAddr,
				sdk.ValAddress(s.network.GetValidators()[0].GetOperator()),
				sdk.ValAddress(s.network.GetValidators()[1].GetOperator()),
			)

			for _, addr := range valAddresses {
				val, err := s.network.App.StakingKeeper.Validator(s.network.GetContext(), addr)
				Expect(err).To(BeNil(), "error while calling the precompile")
				valCommission := sdk.DecCoins{sdk.NewDecCoinFromDec(s.bondDenom, commDec)}

				s.network.App.DistrKeeper.SetValidatorAccumulatedCommission(
					s.network.GetContext(), addr,
					distrtypes.ValidatorAccumulatedCommission{Commission: valCommission},
				)
				s.network.App.DistrKeeper.AllocateTokensToValidator(s.network.GetContext(), val, sdk.DecCoins{sdk.NewDecCoin(s.bondDenom, stakeAmt)})
			}

			initialBalance = s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)

			// populate default arguments
			defaultWithdrawValCommArgs = defaultCallArgs.WithMethodName(
				"testWithdrawValidatorCommission",
			)
		})

		It("should not withdraw commission from validator when sending from a different address", func() {
			withdrawValCommArgs := defaultWithdrawValCommArgs.WithArgs(
				s.network.GetValidators()[0].OperatorAddress,
			)

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, withdrawValCommArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

			// balance should be equal as initial balance or less (because of fees)
			finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)
			Expect(finalBalance.Amount.Uint64() <= initialBalance.Amount.Uint64()).To(BeTrue())

			// validator's balance should remain unchanged
			valFinalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), sdk.AccAddress(s.network.GetValidators()[0].GetOperator()), s.bondDenom)
			Expect(valFinalBalance.Amount).To(Equal(math.ZeroInt()))
		})

		It("should withdraw commission successfully", func() {
			withdrawValCommArgs := defaultWithdrawValCommArgs.
				WithArgs(valAddr.String()).
				WithGasPrice(gasPrice)
			logCheckArgs := passCheck.
				WithExpEvents(distribution.EventTypeWithdrawValidatorCommission)

			res, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, withdrawValCommArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)
			fees := gasPrice.Int64() * res.GasUsed
			expFinal := initialBalance.Amount.Int64() + expValAmount - fees
			Expect(finalBalance.Amount).To(Equal(math.NewInt(expFinal)), "expected final balance to be equal to initial balance + validator commission - fees")
		})
	})

	Context("claimRewards", func() {
		var (
			// defaultClaimRewardsArgs are the default arguments for the claim rewards call
			//
			// NOTE: this has to be populated in a BeforeEach block because the contractAddr would otherwise be a nil address.
			defaultClaimRewardsArgs contracts.CallArgs
			// initialBalance is the initial balance of the delegator
			initialBalance sdk.Coin
		)

		BeforeEach(func() {
			// set some rewards for s.keyring.GetAddr(0) & another address
			// s.prepareStakingRewards([]stakingRewards{
			// 	{s.keyring.GetAddr(0).Bytes(), s.network.GetValidators()[0], rewards},
			// 	{differentAddr.Bytes(), s.network.GetValidators()[0], rewards},
			// }...)

			initialBalance = s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)

			// populate default arguments
			defaultClaimRewardsArgs = defaultCallArgs.WithMethodName(
				"testClaimRewards",
			)
		})

		It("should not claim rewards when sending from a different address", func() {
			claimRewardsArgs := defaultClaimRewardsArgs.WithArgs(
				differentAddr, uint32(1),
			)

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, claimRewardsArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

			// balance should be equal as initial balance or less (because of fees)
			finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)
			Expect(finalBalance.Amount.Uint64() <= initialBalance.Amount.Uint64()).To(BeTrue())

			// differentAddr balance should remain unchanged
			differentAddrFinalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), differentAddr.Bytes(), s.bondDenom)
			Expect(differentAddrFinalBalance.Amount).To(Equal(math.ZeroInt()))
		})

		It("should claim rewards successfully", func() {
			claimRewardsArgs := defaultClaimRewardsArgs.WithArgs(
				s.keyring.GetAddr(0), uint32(2),
			)

			logCheckArgs := passCheck.
				WithExpEvents(distribution.EventTypeClaimRewards)

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, claimRewardsArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			// balance should remain unchanged
			finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)
			Expect(finalBalance.Amount.GT(initialBalance.Amount)).To(BeTrue(), "expected final balance to be greater than initial balance after claiming rewards")
		})
	})

	Context("claimRewards with contract as delegator", func() {
		var (
			// defaultClaimRewardsArgs are the default arguments for the  claim rewards call
			//
			// NOTE: this has to be populated in a BeforeEach block because the contractAddr would otherwise be a nil address.
			defaultClaimRewardsArgs contracts.CallArgs
			// expectedBalance is the total after claiming from both validators
			expectedBalance sdk.Coin
		)

		BeforeEach(func() {
			// set some rewards for s.keyring.GetAddr(0) & another address
			// s.prepareStakingRewards([]stakingRewards{
			// 	{
			// 		Delegator: contractAddr.Bytes(),
			// 		Validator: s.network.GetValidators()[0],
			// 		RewardAmt: rewards,
			// 	}, {
			// 		Delegator: contractAddr.Bytes(),
			// 		Validator: s.network.GetValidators()[1],
			// 		RewardAmt: rewards,
			// 	},
			// }...)

			expectedBalance = sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(2e18)}

			// populate default arguments
			defaultClaimRewardsArgs = defaultCallArgs.WithMethodName(
				"testClaimRewards",
			)
		})

		It("should withdraw rewards successfully without origin check", func() {
			claimRewardsArgs := defaultClaimRewardsArgs.WithArgs(contractAddr, uint32(2))

			logCheckArgs := passCheck.WithExpEvents(distribution.EventTypeClaimRewards)

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, claimRewardsArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			// balance should increase
			finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), contractAddr.Bytes(), s.bondDenom)
			Expect(finalBalance.Amount.Equal(expectedBalance.Amount)).To(BeTrue(), "expected final balance to be greater than initial balance after withdrawing rewards")
		})

		It("should withdraw rewards successfully to a different address without origin check", func() {
			initialBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)

			err := s.network.App.DistrKeeper.SetWithdrawAddr(s.network.GetContext(), contractAddr.Bytes(), s.keyring.GetAddr(0).Bytes())
			Expect(err).To(BeNil())

			claimRewardsArgs := defaultClaimRewardsArgs.WithArgs(contractAddr, uint32(2))

			logCheckArgs := passCheck.WithExpEvents(distribution.EventTypeClaimRewards)

			_, _, err = contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, claimRewardsArgs, logCheckArgs)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			// balance should increase
			finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)
			Expect(finalBalance.Amount.GT(initialBalance.Amount)).To(BeTrue(), "expected final balance to be greater than initial balance after withdrawing rewards")
		})
	})

	Context("Forbidden operations", func() {
		It("should revert state: modify withdraw address & then try to withdraw rewards corresponding to another user", func() {
			// set rewards to another user
			// s.prepareStakingRewards(stakingRewards{differentAddr.Bytes(), s.network.GetValidators()[0], rewards})

			revertArgs := defaultCallArgs.
				WithMethodName("testRevertState").
				WithArgs(
					differentAddr.String(), differentAddr, s.network.GetValidators()[0].OperatorAddress,
				)

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, revertArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

			// check withdraw address didn't change
			withdrawer, err := s.network.App.DistrKeeper.GetDelegatorWithdrawAddr(s.network.GetContext(), s.keyring.GetAddr(0).Bytes())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(withdrawer.Bytes()).To(Equal(s.keyring.GetAddr(0).Bytes()))

			// check signer address balance should've decreased (fees paid)
			finalBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.bondDenom)
			Expect(finalBalance.Amount.Uint64() <= initBalanceAmt.Uint64()).To(BeTrue())

			// check other address' balance remained unchanged
			finalBalance = s.network.App.BankKeeper.GetBalance(s.network.GetContext(), differentAddr.Bytes(), s.bondDenom)
			Expect(finalBalance.Amount).To(Equal(math.ZeroInt()))
		})

		It("should not allow to call SetWithdrawAddress using delegatecall", func() {
			setWithdrawAddrArgs := defaultCallArgs.
				WithMethodName("delegateCallSetWithdrawAddress").
				WithArgs(s.keyring.GetAddr(0), differentAddr.String())

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, setWithdrawAddrArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

			// check withdraw address didn't change
			withdrawer, err := s.network.App.DistrKeeper.GetDelegatorWithdrawAddr(s.network.GetContext(), s.keyring.GetAddr(0).Bytes())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(withdrawer.Bytes()).To(Equal(s.keyring.GetAddr(0).Bytes()))
		})

		It("should not allow to call txs (SetWithdrawAddress) using staticcall", func() {
			setWithdrawAddrArgs := defaultCallArgs.
				WithMethodName("staticCallSetWithdrawAddress").
				WithArgs(s.keyring.GetAddr(0), differentAddr.String())

			_, _, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, setWithdrawAddrArgs, execRevertedCheck)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
			// check withdraw address didn't change
			withdrawer, err := s.network.App.DistrKeeper.GetDelegatorWithdrawAddr(s.network.GetContext(), s.keyring.GetAddr(0).Bytes())
			Expect(err).To(BeNil(), "error while calling the precompile")
			Expect(withdrawer.Bytes()).To(Equal(s.keyring.GetAddr(0).Bytes()))
		})
	})

	// ===================================
	//				QUERIES
	// ===================================
	Context("Distribution precompile queries", func() {
		Context("get validator distribution info", func() {
			// defaultValDistArgs are the default arguments for the getValidatorDistributionInfo query
			//
			// NOTE: this has to be populated in BeforeEach because the test suite setup is not available prior to that.
			var defaultValDistArgs contracts.CallArgs

			BeforeEach(func() {
				addr := sdk.AccAddress(s.network.GetValidators()[0].GetOperator())
				// fund validator account to make self-delegation
				err := evmosutil.FundAccountWithBaseDenom(s.network.GetContext(), s.network.App.BankKeeper, addr, 10)
				Expect(err).To(BeNil())
				// make a self delegation
				_, err = s.network.App.StakingKeeper.Delegate(s.network.GetContext(), addr, math.NewInt(1), stakingtypes.Unspecified, s.network.GetValidators()[0], true)
				Expect(err).To(BeNil())

				defaultValDistArgs = defaultCallArgs.
					WithMethodName("getValidatorDistributionInfo").
					WithArgs(s.network.GetValidators()[0].OperatorAddress)
			})

			It("should get validator distribution info", func() {
				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, defaultValDistArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out distribution.ValidatorDistributionInfoOutput
				err = s.precompile.UnpackIntoInterface(&out, distribution.ValidatorDistributionInfoMethod, ethRes.Ret)
				Expect(err).To(BeNil())

				expAddr := sdk.AccAddress(s.network.GetValidators()[0].GetOperator())
				Expect(expAddr.String()).To(Equal(out.DistributionInfo.OperatorAddress))
				Expect(0).To(Equal(len(out.DistributionInfo.Commission)))
				Expect(0).To(Equal(len(out.DistributionInfo.SelfBondRewards)))
			})
		})

		Context("get validator outstanding rewards", func() { //nolint:dupl
			// defaultValOutRewardsArgs are the default arguments for the getValidatorOutstandingRewards query
			//
			// NOTE: this has to be populated in BeforeEach because the test suite setup is not available prior to that.
			var defaultValOutRewardsArgs contracts.CallArgs

			BeforeEach(func() {
				defaultValOutRewardsArgs = defaultCallArgs.
					WithMethodName("getValidatorOutstandingRewards").
					WithArgs(s.network.GetValidators()[0].OperatorAddress)
			})

			It("should not get rewards - validator without outstanding rewards", func() {
				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, defaultValOutRewardsArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var rewards []cmn.DecCoin
				err = s.precompile.UnpackIntoInterface(&rewards, distribution.ValidatorOutstandingRewardsMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(rewards)).To(Equal(0))
			})

			It("should get rewards - validator with outstanding rewards", func() {
				valRewards := sdk.DecCoins{sdk.NewDecCoinFromDec(s.bondDenom, math.LegacyNewDec(1))}
				// set outstanding rewards
				err := s.network.App.DistrKeeper.SetValidatorOutstandingRewards(s.network.GetContext(), sdk.ValAddress(s.network.GetValidators()[0].GetOperator()), distrtypes.ValidatorOutstandingRewards{Rewards: valRewards})
				Expect(err).To(BeNil(), "error while calling the precompile")

				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, defaultValOutRewardsArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var rewards []cmn.DecCoin
				err = s.precompile.UnpackIntoInterface(&rewards, distribution.ValidatorOutstandingRewardsMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(rewards)).To(Equal(1))
				Expect(uint8(18)).To(Equal(rewards[0].Precision))
				Expect(s.bondDenom).To(Equal(rewards[0].Denom))
				Expect(expValAmount).To(Equal(rewards[0].Amount.Int64()))
			})
		})

		Context("get validator commission", func() { //nolint:dupl
			// defaultValCommArgs are the default arguments for the getValidatorCommission query
			//
			// NOTE: this has to be populated in BeforeEach because the test suite setup is not available prior to that.
			var defaultValCommArgs contracts.CallArgs

			BeforeEach(func() {
				defaultValCommArgs = defaultCallArgs.
					WithMethodName("getValidatorCommission").
					WithArgs(s.network.GetValidators()[0].OperatorAddress)
			})

			It("should not get commission - validator without commission", func() {
				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, defaultValCommArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var commission []cmn.DecCoin
				err = s.precompile.UnpackIntoInterface(&commission, distribution.ValidatorCommissionMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(commission)).To(Equal(0))
			})

			It("should get commission - validator with commission", func() {
				// set commission
				valCommission := sdk.DecCoins{sdk.NewDecCoinFromDec(s.bondDenom, math.LegacyNewDec(1))}
				err := s.network.App.DistrKeeper.SetValidatorAccumulatedCommission(s.network.GetContext(), sdk.ValAddress(s.network.GetValidators()[0].GetOperator()), distrtypes.ValidatorAccumulatedCommission{Commission: valCommission})
				Expect(err).To(BeNil(), "error while calling contract")

				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, defaultValCommArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var commission []cmn.DecCoin
				err = s.precompile.UnpackIntoInterface(&commission, distribution.ValidatorCommissionMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(commission)).To(Equal(1))
				Expect(uint8(18)).To(Equal(commission[0].Precision))
				Expect(s.bondDenom).To(Equal(commission[0].Denom))
				Expect(expValAmount).To(Equal(commission[0].Amount.Int64()))
			})
		})

		Context("get validator slashing events", func() {
			// defaultValSlashArgs are the default arguments for the getValidatorSlashes query
			//
			// NOTE: this has to be populated in BeforeEach because the test suite setup is not available prior to that.
			var defaultValSlashArgs contracts.CallArgs

			BeforeEach(func() {
				defaultValSlashArgs = defaultCallArgs.
					WithMethodName("getValidatorSlashes").
					WithArgs(
						s.network.GetValidators()[0].OperatorAddress,
						uint64(1), uint64(5),
						query.PageRequest{},
					)
			})

			It("should not get slashing events - validator without slashes", func() {
				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, defaultValSlashArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out distribution.ValidatorSlashesOutput
				err = s.precompile.UnpackIntoInterface(&out, distribution.ValidatorSlashesMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(out.Slashes)).To(Equal(0))
			})

			It("should get slashing events - validator with slashes (default pagination)", func() {
				// TODO fixme using custom genesis
				// set slash event
				// slashEvent := s.setupValidatorSlashes(sdk.ValAddress(s.network.GetValidators()[0].GetOperator()), 1)

				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, defaultValSlashArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out distribution.ValidatorSlashesOutput
				err = s.precompile.UnpackIntoInterface(&out, distribution.ValidatorSlashesMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(out.Slashes)).To(Equal(1))
				// Expect(slashEvent.Fraction.BigInt()).To(Equal(out.Slashes[0].Fraction.Value))
				// Expect(slashEvent.ValidatorPeriod).To(Equal(out.Slashes[0].ValidatorPeriod))
				Expect(uint64(1)).To(Equal(out.PageResponse.Total))
				Expect(out.PageResponse.NextKey).To(BeEmpty())
			})

			It("should get slashing events - validator with slashes w/pagination", func() {
				// TODO fixme using custom genesis
				// set 2 slashing events
				// slashEvent := s.setupValidatorSlashes(sdk.ValAddress(s.network.GetValidators()[0].GetOperator()), 2)

				// set pagination
				defaultValSlashArgs.Args = []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
					uint64(1), uint64(5),
					query.PageRequest{
						Limit:      1,
						CountTotal: true,
					},
				}

				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, defaultValSlashArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out distribution.ValidatorSlashesOutput
				err = s.precompile.UnpackIntoInterface(&out, distribution.ValidatorSlashesMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(out.Slashes)).To(Equal(1))
				// Expect(slashEvent.Fraction.BigInt()).To(Equal(out.Slashes[0].Fraction.Value))
				// Expect(slashEvent.ValidatorPeriod).To(Equal(out.Slashes[0].ValidatorPeriod))
				Expect(uint64(2)).To(Equal(out.PageResponse.Total))
				Expect(out.PageResponse.NextKey).NotTo(BeEmpty())
			})
		})

		Context("get delegation rewards", func() {
			// defaultDelRewardsArgs are the default arguments for the getDelegationRewards query
			//
			// NOTE: this has to be populated in BeforeEach because the test suite setup is not available prior to that.
			var defaultDelRewardsArgs contracts.CallArgs

			BeforeEach(func() {
				defaultDelRewardsArgs = defaultCallArgs.
					WithMethodName("getDelegationRewards").
					WithArgs(s.keyring.GetAddr(0), s.network.GetValidators()[0].OperatorAddress)
			})

			It("should not get rewards - no rewards available", func() {
				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, defaultDelRewardsArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var rewards []cmn.DecCoin
				err = s.precompile.UnpackIntoInterface(&rewards, distribution.DelegationRewardsMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(rewards)).To(Equal(0))
			})
			It("should get rewards", func() {
				// s.prepareStakingRewards(stakingRewards{s.keyring.GetAddr(0).Bytes(), s.network.GetValidators()[0], rewards})

				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, defaultDelRewardsArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var rewards []cmn.DecCoin
				err = s.precompile.UnpackIntoInterface(&rewards, distribution.DelegationRewardsMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(rewards)).To(Equal(1))
				Expect(len(rewards)).To(Equal(1))
				Expect(rewards[0].Denom).To(Equal(s.bondDenom))
				Expect(rewards[0].Amount.Int64()).To(Equal(testRewards))
			})
		})

		Context("get delegator's total rewards", func() {
			// defaultDelTotalRewardsArgs are the default arguments for the getDelegationTotalRewards query
			//
			// NOTE: this has to be populated in BeforeEach because the test suite setup is not available prior to that.
			var defaultDelTotalRewardsArgs contracts.CallArgs

			BeforeEach(func() {
				defaultDelTotalRewardsArgs = defaultCallArgs.
					WithMethodName("getDelegationTotalRewards").
					WithArgs(s.keyring.GetAddr(0))
			})

			It("should not get rewards - no rewards available", func() {
				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, defaultDelTotalRewardsArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out distribution.DelegationTotalRewardsOutput
				err = s.precompile.UnpackIntoInterface(&out, distribution.DelegationTotalRewardsMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(len(out.Rewards)).To(Equal(2))
				Expect(len(out.Rewards[0].Reward)).To(Equal(0))
				Expect(len(out.Rewards[1].Reward)).To(Equal(0))
			})
			It("should get total rewards", func() {
				// set rewards
				// s.prepareStakingRewards(stakingRewards{s.keyring.GetAddr(0).Bytes(), s.network.GetValidators()[0], rewards})

				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, defaultDelTotalRewardsArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var (
					out distribution.DelegationTotalRewardsOutput
					i   int
				)
				err = s.precompile.UnpackIntoInterface(&out, distribution.DelegationTotalRewardsMethod, ethRes.Ret)
				Expect(err).To(BeNil())

				// the response order may change
				if out.Rewards[0].ValidatorAddress == s.network.GetValidators()[0].OperatorAddress {
					Expect(s.network.GetValidators()[0].OperatorAddress).To(Equal(out.Rewards[0].ValidatorAddress))
					Expect(s.network.GetValidators()[1].OperatorAddress).To(Equal(out.Rewards[1].ValidatorAddress))
					Expect(0).To(Equal(len(out.Rewards[1].Reward)))
				} else {
					i = 1
					Expect(s.network.GetValidators()[0].OperatorAddress).To(Equal(out.Rewards[1].ValidatorAddress))
					Expect(s.network.GetValidators()[1].OperatorAddress).To(Equal(out.Rewards[0].ValidatorAddress))
					Expect(0).To(Equal(len(out.Rewards[0].Reward)))
				}

				// only validator[i] has rewards
				Expect(1).To(Equal(len(out.Rewards[i].Reward)))
				Expect(s.bondDenom).To(Equal(out.Rewards[i].Reward[0].Denom))
				Expect(uint8(math.LegacyPrecision)).To(Equal(out.Rewards[i].Reward[0].Precision))
				Expect(testRewards).To(Equal(out.Rewards[i].Reward[0].Amount.Int64()))

				Expect(1).To(Equal(len(out.Total)))
				Expect(testRewards).To(Equal(out.Total[0].Amount.Int64()))
			})
		})

		Context("get all delegator validators", func() {
			// defaultDelValArgs are the default arguments for the getDelegatorValidators query
			//
			// NOTE: this has to be populated in BeforeEach because the test suite setup is not available prior to that.
			var defaultDelValArgs contracts.CallArgs

			BeforeEach(func() {
				defaultDelValArgs = defaultCallArgs.
					WithMethodName("getDelegatorValidators").
					WithArgs(s.keyring.GetAddr(0))
			})

			It("should get all validators a delegator has delegated to", func() {
				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, defaultDelValArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var validators []string
				err = s.precompile.UnpackIntoInterface(&validators, distribution.DelegatorValidatorsMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				Expect(2).To(Equal(len(validators)))

				// the response order may change
				if validators[0] == s.network.GetValidators()[0].OperatorAddress {
					Expect(s.network.GetValidators()[0].OperatorAddress).To(Equal(validators[0]))
					Expect(s.network.GetValidators()[1].OperatorAddress).To(Equal(validators[1]))
				} else {
					Expect(s.network.GetValidators()[1].OperatorAddress).To(Equal(validators[0]))
					Expect(s.network.GetValidators()[0].OperatorAddress).To(Equal(validators[1]))
				}
			})
		})

		Context("get withdraw address", func() {
			// defaultWithdrawAddrArgs are the default arguments for the getDelegatorWithdrawAddress query
			//
			// NOTE: this has to be populated in BeforeEach because the test suite setup is not available prior to that.
			var defaultWithdrawAddrArgs contracts.CallArgs

			BeforeEach(func() {
				defaultWithdrawAddrArgs = defaultCallArgs.
					WithMethodName("getDelegatorWithdrawAddress").
					WithArgs(s.keyring.GetAddr(0))
			})

			It("should get withdraw address", func() {
				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, defaultWithdrawAddrArgs, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				withdrawAddr, err := s.precompile.Unpack(distribution.DelegatorWithdrawAddressMethod, ethRes.Ret)
				Expect(err).To(BeNil())
				// get the bech32 encoding
				expAddr := sdk.AccAddress(s.keyring.GetAddr(0).Bytes())
				Expect(withdrawAddr[0]).To(Equal(expAddr.String()))
			})

			It("should call GetWithdrawAddress using staticcall", func() {
				staticCallArgs := defaultCallArgs.
					WithMethodName("staticCallGetWithdrawAddress").
					WithArgs(s.keyring.GetAddr(0))

				_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, staticCallArgs, passCheck)
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
