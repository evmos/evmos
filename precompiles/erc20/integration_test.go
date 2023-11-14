package erc20_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	"github.com/evmos/evmos/v15/precompiles/testutil/contracts"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"math/big"
)

var _ = Describe("ERC20 Extension - ", func() {
	var (
		defaultCallArgs contracts.CallArgs

		// contractCall returns the call arguments in order to call the ERC20 extension through
		// a smart contract.
		contractCall func() contracts.CallArgs
		// directCall returns the call arguments in order to call the ERC20 extension directly.
		directCall func() contracts.CallArgs

		sender            keyring.Key
		failCheck         testutil.LogCheckArgs
		execRevertedCheck testutil.LogCheckArgs
		passCheck         testutil.LogCheckArgs
	)

	BeforeEach(func() {
		s.SetupTest()

		sender = s.keyring.GetKey(0)
		defaultCallArgs = contracts.CallArgs{
			PrivKey: sender.Priv,
		}

		contractCall = func() contracts.CallArgs {
			return defaultCallArgs
			// FIXME: add contract call support
			// WithABI(s.precompile.ABI).
			// WithAddress(s.precompile.Address())
		}
		_ = contractCall

		directCall = func() contracts.CallArgs {
			return defaultCallArgs.
				WithABI(s.precompile.ABI).
				WithAddress(s.precompile.Address())
		}
		_ = directCall

		// Set up the checks
		failCheck = testutil.LogCheckArgs{
			ABIEvents: s.precompile.Events,
			ExpPass:   false,
		}
		execRevertedCheck = failCheck.WithErrContains("execution reverted")
		passCheck = failCheck.WithExpPass(true)

		// TODO: remove these once used
		_ = execRevertedCheck
		_ = passCheck
	})

	When("querying balance", func() {
		It("should return an existing balance", func() {
			expBalance := big.NewInt(100)

			// Fund account with some tokens
			err := s.network.FundAccount(sender.AccAddr, sdk.Coins{{s.tokenDenom, sdk.NewIntFromBigInt(expBalance)}})
			Expect(err).ToNot(HaveOccurred(), "failed to fund account")

			// Query the balance
			balancesArgs := directCall().
				WithMethodName(erc20.BalanceOfMethod).
				WithArgs(sender.Addr)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, balancesArgs, passCheck)
			Expect(err).ToNot(HaveOccurred(), "failed to call contract")
			Expect(ethRes.Ret).ToNot(BeEmpty(), "expected result")

			var res *big.Int
			err = s.precompile.UnpackIntoInterface(&res, erc20.BalanceOfMethod, ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			Expect(res).To(Equal(expBalance), "expected different balance")
		})

		It("should return zero if balance only exists for other tokens", func() {
			address := utiltx.GenerateAddress()

			// Fund account with some tokens
			err := s.network.FundAccount(sender.AccAddr, sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 100)})
			Expect(err).ToNot(HaveOccurred(), "failed to fund account")

			// Query the balance
			balancesArgs := directCall().
				WithMethodName(erc20.BalanceOfMethod).
				WithArgs(address)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, balancesArgs, passCheck)
			Expect(err).ToNot(HaveOccurred(), "failed to call contract")

			var res *big.Int
			err = s.precompile.UnpackIntoInterface(&res, erc20.BalanceOfMethod, ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			Expect(res.Int64()).To(BeZero(), "expected zero balance")
		})

		It("should return zero if the account does not exist", func() {
			address := utiltx.GenerateAddress()

			balancesArgs := directCall().
				WithMethodName(erc20.BalanceOfMethod).
				WithArgs(address)

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.network.GetContext(), s.network.App, balancesArgs, passCheck)
			Expect(err).ToNot(HaveOccurred(), "failed to call contract")
			Expect(ethRes).ToNot(BeNil(), "expected result")
			println("ethRes.Ret", ethRes.Ret)

			var res *big.Int
			err = s.precompile.UnpackIntoInterface(&res, erc20.BalanceOfMethod, ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			Expect(res.Int64()).To(BeZero(), "expected zero balance")
		})
	})
})
