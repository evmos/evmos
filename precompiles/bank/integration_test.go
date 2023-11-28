package bank_test

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v15/precompiles/bank"
	"math/big"

	"github.com/evmos/evmos/v15/precompiles/testutil"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

var _ = Describe("Bank Extension -", func() {
	var (
		//BankCallerContractAddr         common.Address
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

		sender = s.keyring.GetKey(0)

		//WERC20ContractAddr, err = s.factory.DeployContract(
		//	sender.Priv,
		//	evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
		//	factory.ContractDeploymentData{
		//		Contract:        testdata.WEVMOSContract,
		//		ConstructorArgs: []interface{}{},
		//	},
		//)
		//Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

		contractData = ContractData{
			ownerPriv:      sender.Priv,
			precompileAddr: s.precompile.Address(),
			precompileABI:  s.precompile.ABI,
		}

		passCheck = testutil.LogCheckArgs{}.WithExpPass(true)

		err = s.network.NextBlock()
		Expect(err).ToNot(HaveOccurred(), "failed to advance block")

		// Default sender and amount
		sender = s.keyring.GetKey(0)
		amount = big.NewInt(1e18)
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

				balanceAfter, err := s.grpcHandler.GetBalance(sender.AccAddr, "xmpl")
				fmt.Println(balanceAfter)
				fmt.Println(ethRes.GasUsed)

				var balances []bank.Balance
				err = s.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				fmt.Println(balances)

			})
		})

		Context("totalSupply query", func() {})

		Context("supplyOf query", func() {})

	})
})
