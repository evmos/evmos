package werc20_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	"github.com/evmos/evmos/v15/precompiles/werc20"
	"github.com/evmos/evmos/v15/precompiles/werc20/testdata"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	erc20types "github.com/evmos/evmos/v15/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

var _ = Describe("WEVMOS Extension -", func() {
	var (
		WEVMOSContractAddr common.Address
		err                error
		sender             keyring.Key

		// contractData is a helper struct to hold the addresses and ABIs for the
		// different contract instances that are subject to testing here.
		contractData ContractData

		_         testutil.LogCheckArgs
		failCheck testutil.LogCheckArgs
		passCheck testutil.LogCheckArgs
	)

	BeforeEach(func() {
		s.SetupTest()

		sender = s.keyring.GetKey(0)

		WEVMOSContractAddr, err = s.factory.DeployContract(
			sender.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract:        testdata.WEVMOSContract,
				ConstructorArgs: []interface{}{},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

		// Create the token pair for WEVMOS <-> EVMOS.
		tokenPair := erc20types.NewTokenPair(WEVMOSContractAddr, s.bondDenom, erc20types.OWNER_MODULE)
		s.network.App.Erc20Keeper.SetTokenPair(s.network.GetContext(), tokenPair)

		precompile, err := werc20.NewPrecompile(
			tokenPair,
			s.network.App.BankKeeper,
			s.network.App.AuthzKeeper,
			s.network.App.TransferKeeper,
		)

		Expect(err).ToNot(HaveOccurred(), "failed to create wevmos extension")
		s.precompile = precompile

		err = s.network.App.EvmKeeper.AddEVMExtensions(s.network.GetContext(), precompile)
		Expect(err).ToNot(HaveOccurred(), "failed to add wevmos extension")

		contractData = ContractData{
			ownerPriv:      sender.Priv,
			erc20Addr:      WEVMOSContractAddr,
			erc20ABI:       testdata.WEVMOSContract.ABI,
			precompileAddr: s.precompile.Address(),
			precompileABI:  s.precompile.ABI,
		}

		failCheck = testutil.LogCheckArgs{ABIEvents: s.precompile.Events}
		// execRevertedCheck = failCheck.WithErrContains("execution reverted")
		passCheck = failCheck.WithExpPass(true)

		err = s.network.NextBlock()
		Expect(err).ToNot(HaveOccurred(), "failed to advance block")
	})

	Context("WEVMOS specific functions", func() {
		When("calling deposit and withdraw correctly", func() {
			It("calling deposit - should emit the Deposit event but not modify the balance", func() {
				sender := s.keyring.GetKey(0)

				depositCheck := passCheck.WithExpPass(true).WithExpEvents(werc20.EventTypeDeposit)
				txArgs, callArgs := s.getTxAndCallArgs(erc20Call, contractData, werc20.DepositMethod)
				txArgs.Amount = big.NewInt(1e18)

				_, _, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, callArgs, depositCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.checkBalances(failCheck, sender, contractData)
			})

			It("calling withdraw - should emit the Withdrawal event but not modify the balance", func() {
				// Calling withdraw method
				sender := s.keyring.GetKey(0)
				amount := big.NewInt(1e18)

				withdrawCheck := passCheck.WithExpPass(true).WithExpEvents(werc20.EventTypeWithdrawal)
				txArgs, callArgs := s.getTxAndCallArgs(erc20Call, contractData, werc20.WithdrawMethod, amount)

				_, _, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, callArgs, withdrawCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.checkBalances(failCheck, sender, contractData)
			})
		})

		// TODO: How do we actually check the method types here? We can see the correct ones being populated in the
		// by printing the line in the cmn.Precompile
		When("calling with incomplete data or amount", func() {
			It("calls with no function, no call data, with amount - should call `receive` ", func() {
				sender := s.keyring.GetKey(0)

				txArgs, callArgs := s.getTxAndCallArgs(erc20Call, contractData, "")
				txArgs.Amount = big.NewInt(1e18)

				res, err := s.factory.ExecuteContractCall(sender.Priv, txArgs, callArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				depositCheck := passCheck.WithExpPass(true).WithExpEvents(werc20.EventTypeDeposit)
				depositCheck.Res = res
				err = testutil.CheckLogs(depositCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.checkBalances(failCheck, sender, contractData)
			})

			It("calls with no function, short call data, with amount -  should call `fallback` ", func() {
				sender := s.keyring.GetKey(0)

				txArgs, _ := s.getTxAndCallArgs(erc20Call, contractData, "")
				txArgs.Amount = big.NewInt(1e18)
				txArgs.Input = []byte{1, 2, 3} // 3 dummy bytes

				res, err := s.factory.ExecuteEthTx(sender.Priv, txArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				depositCheck := passCheck.WithExpPass(true).WithExpEvents(werc20.EventTypeDeposit)
				depositCheck.Res = res
				err = testutil.CheckLogs(depositCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.checkBalances(failCheck, sender, contractData)
			})

			It("calls with no function, standard length call data, with amount - should call `fallback` ", func() {
				sender := s.keyring.GetKey(0)

				txArgs, _ := s.getTxAndCallArgs(erc20Call, contractData, "")
				txArgs.Amount = big.NewInt(1e18)
				txArgs.Input = []byte{1, 2, 3, 4} // 4 dummy bytes needed for minimum length

				res, err := s.factory.ExecuteEthTx(sender.Priv, txArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				depositCheck := passCheck.WithExpPass(true).WithExpEvents(werc20.EventTypeDeposit)
				depositCheck.Res = res
				err = testutil.CheckLogs(depositCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.checkBalances(failCheck, sender, contractData)
			})

			It("calls with no function, standard length call data, no amount -  should call `fallback` ", func() {
				sender := s.keyring.GetKey(0)

				txArgs, _ := s.getTxAndCallArgs(erc20Call, contractData, "")
				txArgs.Input = []byte{1, 2, 3, 4} // 4 dummy bytes needed for minimum length

				res, err := s.factory.ExecuteEthTx(sender.Priv, txArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				depositCheck := passCheck.WithExpPass(true).WithExpEvents(werc20.EventTypeDeposit)
				depositCheck.Res = res
				err = testutil.CheckLogs(depositCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.checkBalances(failCheck, sender, contractData)
			})
		})
	})

	// TODO: Add more granular cases but don't want to just confirm the same functionality as ERC20 tests.
	// Context("ERC20 specific functions", func() {
	//	When("querying balance", func() {
	//		DescribeTable("it should return an existing balance", func(callType int) {
	//			Entry("direct WERC20 contract call", func() {})
	//			Entry("contract call", func() {})
	//		})
	//	})
	//
	//	When("querying allowance", func() {
	//		DescribeTable("it should return an existing allowance", func(callType int) {
	//			Entry("direct WERC20 contract call", func() {})
	//			Entry("contract call", func() {})
	//		})
	//	})
	//
	//	When("querying total supply", func() {
	//		DescribeTable("it should return the total supply", func(callType int) {
	//			Entry("direct WERC20 contract call", func() {})
	//			Entry("contract call", func() {})
	//		})
	//	})
	//
	//	When("approving for a spender", func() {
	//		DescribeTable("it should approve the spender", func(callType int) {
	//			Entry("direct WERC20 contract call", func() {})
	//			Entry("contract call", func() {})
	//		})
	//	})
	//
	//	When("transferring tokens from contract caller", func() {
	//		DescribeTable("it should transfer the tokens", func(callType int) {
	//			Entry("direct WERC20 contract call", func() {})
	//			Entry("contract call", func() {})
	//		})
	//	})
	//
	//	When("transferring tokens on behalf of a custom `from` ", func() {
	//		DescribeTable("it should transfer the tokens", func(callType int) {
	//			Entry("direct WERC20 contract call", func() {})
	//			Entry("contract call", func() {})
	//		})
	//	})
	//})
})
