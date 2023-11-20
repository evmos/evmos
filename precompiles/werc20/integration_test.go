package werc20_test

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	"github.com/evmos/evmos/v15/precompiles/werc20"
	"github.com/evmos/evmos/v15/precompiles/werc20/testdata"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	erc20types "github.com/evmos/evmos/v15/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
	"math/big"

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
		//execRevertedCheck = failCheck.WithErrContains("execution reverted")
		passCheck = failCheck.WithExpPass(true)

		err = s.network.NextBlock()
		Expect(err).ToNot(HaveOccurred(), "failed to advance block")

	})

	Context("WEVMOS specific functions", func() {
		It("calling fallback function - should emit the Deposit event but not modify the balance", func() {
			sender := s.keyring.GetKey(0)

			depositCheck := passCheck.WithExpPass(true).WithExpEvents(werc20.EventTypeDeposit)
			txArgs, callArgs := s.getTxAndCallArgs(erc20Call, contractData, "")
			txArgs.Amount = big.NewInt(1e18)

			fmt.Println(txArgs)

			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, callArgs, depositCheck)
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

			balanceCheck := failCheck.WithExpPass(true)
			txArgs, balancesArgs := s.getTxAndCallArgs(erc20Call, contractData, erc20.BalanceOfMethod, sender.Addr)

			_, ethRes, err = s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, balanceCheck)
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

			// Check the balance in the bank module is the same as calling `balanceOf` on the precompile
			balanceAfter := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), sender.AccAddr, s.bondDenom)

			var erc20Balance *big.Int
			err = s.precompile.UnpackIntoInterface(&erc20Balance, erc20.BalanceOfMethod, ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			Expect(balanceAfter.Amount.BigInt()).To(Equal(erc20Balance), "expected different balance")
		})

		It("calling deposit - should emit the Deposit event but not modify the balance", func() {
			sender := s.keyring.GetKey(0)

			depositCheck := passCheck.WithExpPass(true).WithExpEvents(werc20.EventTypeDeposit)
			txArgs, callArgs := s.getTxAndCallArgs(erc20Call, contractData, werc20.DepositMethod)
			txArgs.Amount = big.NewInt(1e18)

			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, callArgs, depositCheck)
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

			balanceCheck := failCheck.WithExpPass(true)
			txArgs, balancesArgs := s.getTxAndCallArgs(erc20Call, contractData, erc20.BalanceOfMethod, sender.Addr)

			_, ethRes, err = s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, balanceCheck)
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

			// Check the balance in the bank module is the same as calling `balanceOf` on the precompile
			balanceAfter := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), sender.AccAddr, s.bondDenom)

			var erc20Balance *big.Int
			err = s.precompile.UnpackIntoInterface(&erc20Balance, erc20.BalanceOfMethod, ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			Expect(balanceAfter.Amount.BigInt()).To(Equal(erc20Balance), "expected different balance")
		})

		It("calling withdraw - should emit the Withdrawal event but not modify the balance", func() {
			// Calling withdraw method
			sender := s.keyring.GetKey(0)
			amount := big.NewInt(1e18)

			withdrawCheck := passCheck.WithExpPass(true).WithExpEvents(werc20.EventTypeWithdraw)
			txArgs, callArgs := s.getTxAndCallArgs(erc20Call, contractData, werc20.WithdrawMethod, amount)

			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, callArgs, withdrawCheck)
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

			balanceCheck := failCheck.WithExpPass(true)
			txArgs, balancesArgs := s.getTxAndCallArgs(erc20Call, contractData, erc20.BalanceOfMethod, sender.Addr)

			_, ethRes, err = s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, balanceCheck)
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

			// Check the balance in the bank module is the same as calling `balanceOf` on the precompile
			balanceAfter := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), sender.AccAddr, s.bondDenom)

			var erc20Balance *big.Int
			err = s.precompile.UnpackIntoInterface(&erc20Balance, erc20.BalanceOfMethod, ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			Expect(balanceAfter.Amount.BigInt()).To(Equal(erc20Balance), "expected different balance")
		})
	})
})
