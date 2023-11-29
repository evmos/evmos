package bank_test

import (
	"math/big"

	evmosutiltx "github.com/evmos/evmos/v15/testutil/tx"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v15/precompiles/bank"

	"github.com/evmos/evmos/v15/precompiles/testutil"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

var _ = Describe("Bank Extension -", func() {
	var (
		// BankCallerContractAddr         common.Address
		err    error
		sender keyring.Key
		amount *big.Int

		// contractData is a helper struct to hold the addresses and ABIs for the
		// different contract instances that are subject to testing here.
		contractData ContractData
		passCheck    testutil.LogCheckArgs
	)

	BeforeEach(func() {
		s.SetupTest()

		// Default sender and amount
		sender = s.keyring.GetKey(0)
		amount = big.NewInt(1e18)

		contractData = ContractData{
			ownerPriv:      sender.Priv,
			precompileAddr: s.precompile.Address(),
			precompileABI:  s.precompile.ABI,
		}

		passCheck = testutil.LogCheckArgs{}.WithExpPass(true)

		err = s.network.NextBlock()
		Expect(err).ToNot(HaveOccurred(), "failed to advance block")
	})

	Context("Direct precompile queries", func() {
		Context("balances query", func() {
			It("should return the correct balance", func() {
				balanceBefore, err := s.grpcHandler.GetBalance(sender.AccAddr, "xmpl")
				Expect(err).ToNot(HaveOccurred(), "failed to get balance")
				Expect(balanceBefore.Balance.Amount).To(Equal(sdk.NewInt(0)))
				Expect(balanceBefore.Balance.Denom).To(Equal("xmpl"))

				s.mintAndSendCoin("xmpl", s.keyring.GetAccAddr(0), sdk.NewInt(amount.Int64()))

				queryArgs, balancesArgs := s.getTxAndCallArgs(directCall, contractData, bank.BalancesMethod, sender.Addr)
				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = s.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				balanceAfter, err := s.grpcHandler.GetBalance(sender.AccAddr, "xmpl")
				Expect(err).ToNot(HaveOccurred(), "failed to get balance")

				Expect(sdk.NewInt(balances[1].Amount.Int64())).To(Equal(balanceAfter.Balance.Amount))
			})

			It("should return a single token balance", func() {
				queryArgs, balancesArgs := s.getTxAndCallArgs(directCall, contractData, bank.BalancesMethod, sender.Addr)
				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = s.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				balanceAfter, err := s.grpcHandler.GetBalance(sender.AccAddr, "aevmos")
				Expect(err).ToNot(HaveOccurred(), "failed to get balance")

				Expect(sdk.NewInt(balances[0].Amount.Int64())).To(Equal(balanceAfter.Balance.Amount))
			})

			It("should return no balance for new account", func() {
				s.mintAndSendCoin("xmpl", s.keyring.GetAccAddr(0), sdk.NewInt(amount.Int64()))

				queryArgs, balancesArgs := s.getTxAndCallArgs(directCall, contractData, bank.BalancesMethod, sender.Addr)
				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = s.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				Expect(balances).To(BeEmpty())
			})

			It("should consume the correct amount of gas", func() {
				s.mintAndSendCoin("xmpl", s.keyring.GetAccAddr(0), sdk.NewInt(amount.Int64()))

				queryArgs, balancesArgs := s.getTxAndCallArgs(directCall, contractData, bank.BalancesMethod, sender.Addr)
				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = s.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				gasUsed := Max(bank.GasBalanceOf, len(balances)*bank.GasBalanceOf)
				// Here increasing the GasBalanceOf will increase the use of gas so they will never be equal
				Expect(gasUsed).To(BeNumerically("<=", ethRes.GasUsed))
			})
		})

		Context("totalSupply query", func() {
			It("should return the correct total supply", func() {
				queryArgs, supplyArgs := s.getTxAndCallArgs(directCall, contractData, bank.TotalSupplyMethod)
				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = s.precompile.UnpackIntoInterface(&balances, bank.TotalSupplyMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				evmosTotalSupply, ok := new(big.Int).SetString("11000000000000000000", 10)
				Expect(ok).To(BeTrue(), "failed to parse evmos total supply")
				xmplTotalSupply := amount

				Expect(balances[0].Amount).To(Equal(evmosTotalSupply))
				Expect(balances[1].Amount).To(Equal(xmplTotalSupply))
			})
		})

		Context("supplyOf query", func() {
			It("should return the supply of Evmos", func() {
				queryArgs, supplyArgs := s.getTxAndCallArgs(directCall, contractData, bank.SupplyOfMethod, s.evmosAddr)
				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				out, err := s.precompile.Unpack(bank.SupplyOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				evmosTotalSupply, ok := new(big.Int).SetString("11000000000000000000", 10)
				Expect(ok).To(BeTrue(), "failed to parse evmos total supply")

				Expect(out[0].(*big.Int)).To(Equal(evmosTotalSupply))
			})

			It("should return the supply of XMPL", func() {
				queryArgs, supplyArgs := s.getTxAndCallArgs(directCall, contractData, bank.SupplyOfMethod, s.xmplAddr)
				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				out, err := s.precompile.Unpack(bank.SupplyOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				Expect(out[0].(*big.Int)).To(Equal(amount))
			})

			It("should return an error for a non existing token", func() {
				queryArgs, supplyArgs := s.getTxAndCallArgs(directCall, contractData, bank.SupplyOfMethod, evmosutiltx.GenerateAddress())
				_, _, err := s.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).To(HaveOccurred(), "unexpected result calling contract")
			})

			It("should consume the correct amount of gas", func() {
				queryArgs, supplyArgs := s.getTxAndCallArgs(directCall, contractData, bank.SupplyOfMethod, s.xmplAddr)
				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Here increasing the GasSupplyOf will increase the use of gas so they will never be equal
				Expect(bank.GasSupplyOf).To(BeNumerically("<=", ethRes.GasUsed))
			})
		})
	})
})
