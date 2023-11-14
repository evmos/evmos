package erc20_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	"github.com/evmos/evmos/v15/precompiles/erc20/testdata"
	"github.com/evmos/evmos/v15/precompiles/testutil"
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
		sender       keyring.Key

		failCheck testutil.LogCheckArgs
		passCheck testutil.LogCheckArgs
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

		failCheck = testutil.LogCheckArgs{ABIEvents: s.precompile.Events}
		passCheck = failCheck.WithExpPass(true)
	})

	When("querying balance", func() {
		DescribeTable("it should return an existing balance", func(callType int) {
			expBalance := big.NewInt(100)

			// Fund account with some tokens
			err := s.network.FundAccount(sender.AccAddr, sdk.Coins{{s.tokenDenom, sdk.NewIntFromBigInt(expBalance)}})
			Expect(err).ToNot(HaveOccurred(), "failed to fund account")

			// Query the balance
			txArgs, balancesArgs := s.getTxAndCallArgs(callType, contractAddr)
			balancesArgs.MethodName = erc20.BalanceOfMethod
			balancesArgs.Args = []interface{}{sender.Addr}

			_, ethRes, err := s.callContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, passCheck)
			Expect(err).ToNot(HaveOccurred(), "failed to call contract")

			var balance *big.Int
			err = s.precompile.UnpackIntoInterface(&balance, erc20.BalanceOfMethod, ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			Expect(balance).To(Equal(expBalance), "expected different balance")
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
			txArgs, balancesArgs := s.getTxAndCallArgs(callType, contractAddr)
			balancesArgs.MethodName = erc20.BalanceOfMethod
			balancesArgs.Args = []interface{}{address}

			_, ethRes, err := s.callContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, passCheck)
			Expect(err).ToNot(HaveOccurred(), "failed to call contract")

			var balance *big.Int
			err = s.precompile.UnpackIntoInterface(&balance, erc20.BalanceOfMethod, ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			Expect(balance.Int64()).To(BeZero(), "expected zero balance")
		},
			Entry(" - direct call", directCall),
			Entry(" - through contract", contractCall),
		)

		DescribeTable("it should return zero if the account does not exist", func(callType int) {
			address := utiltx.GenerateAddress()

			// Query the balance
			txArgs, balancesArgs := s.getTxAndCallArgs(callType, contractAddr)
			balancesArgs.MethodName = erc20.BalanceOfMethod
			balancesArgs.Args = []interface{}{address}

			res, err := s.factory.ExecuteContractCall(sender.Priv, txArgs, balancesArgs)
			Expect(err).ToNot(HaveOccurred(), "failed to call contract")
			Expect(res.IsOK()).To(BeTrue(), "expected tx to be ok")

			ethRes, err := evmtypes.DecodeTxResponse(res.Data)
			Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")
			Expect(ethRes.Ret).ToNot(BeEmpty(), "expected result")

			var balance *big.Int
			err = s.precompile.UnpackIntoInterface(&balance, erc20.BalanceOfMethod, ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			Expect(balance.Int64()).To(BeZero(), "expected zero balance")
		},
			Entry(" - direct call", directCall),
			Entry(" - through contract", contractCall),
		)
	})
})
