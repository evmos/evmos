package erc20_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	"github.com/evmos/evmos/v15/precompiles/erc20/testdata"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	"github.com/evmos/evmos/v15/precompiles/testutil/contracts"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ERC20 Extension -", func() {
	var (
		contractAddr common.Address
		err          error
		failCheck    testutil.LogCheckArgs
		passCheck    testutil.LogCheckArgs
		sender       keyring.Key
	)

	BeforeEach(func() {
		s.SetupTest()

		sender = s.keyring.GetKey(0)

		contractAddr, err = s.factory.DeployContract(
			sender.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract:        testdata.ERC20CallerContract,
				ConstructorArgs: []interface{}{s.precompile.Address()},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

		// Set up the checks
		failCheck = testutil.LogCheckArgs{
			ABIEvents: s.precompile.Events,
			ExpPass:   false,
		}
		passCheck = failCheck.WithExpPass(true)
	})

	When("querying balance", func() {
		DescribeTable("it should return an existing balance", func(callType int) {
			expBalance := big.NewInt(100)

			// Fund account with some tokens
			err := s.network.FundAccount(sender.AccAddr, sdk.Coins{{s.tokenDenom, sdk.NewIntFromBigInt(expBalance)}})
			Expect(err).ToNot(HaveOccurred(), "failed to fund account")

			// Query the balance
			balancesArgs := s.getTxArgs(sender, callType, contractAddr).
				WithMethodName(erc20.BalanceOfMethod).
				WithArgs(sender.Addr)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, balancesArgs, passCheck)
			Expect(err).ToNot(HaveOccurred(), "failed to call contract")
			Expect(ethRes.Ret).ToNot(BeEmpty(), "expected result")

			var res *big.Int
			err = s.precompile.UnpackIntoInterface(&res, erc20.BalanceOfMethod, ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			Expect(res).To(Equal(expBalance), "expected different balance")
		},
			Entry(" - direct call", directCall),
			Entry(" - through contract", contractCall),
		)

		DescribeTable("it should return zero if balance only exists for other tokens", func(callType int) {
			address := utiltx.GenerateAddress()

			// Fund account with some tokens
			err := s.network.FundAccount(sender.AccAddr, sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 100)})
			Expect(err).ToNot(HaveOccurred(), "failed to fund account")

			// Query the balance
			balancesArgs := s.getTxArgs(sender, callType, contractAddr).
				WithMethodName(erc20.BalanceOfMethod).
				WithArgs(address)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, balancesArgs, passCheck)
			Expect(err).ToNot(HaveOccurred(), "failed to call contract")

			var res *big.Int
			err = s.precompile.UnpackIntoInterface(&res, erc20.BalanceOfMethod, ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			Expect(res.Int64()).To(BeZero(), "expected zero balance")
		},
			Entry(" - direct call", directCall),
			Entry(" - through contract", contractCall),
		)

		DescribeTable("it should return zero if the account does not exist", func(callType int) {
			address := utiltx.GenerateAddress()

			balancesArgs := s.getTxArgs(sender, callType, contractAddr).
				WithMethodName(erc20.BalanceOfMethod).
				WithArgs(address)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, balancesArgs, passCheck)
			Expect(err).ToNot(HaveOccurred(), "failed to call contract")
			Expect(ethRes).ToNot(BeNil(), "expected result")

			var res *big.Int
			err = s.precompile.UnpackIntoInterface(&res, erc20.BalanceOfMethod, ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			Expect(res.Int64()).To(BeZero(), "expected zero balance")
		},
			Entry(" - direct call", directCall),
			Entry(" - through contract", contractCall),
		)
	})
})
