package erc20_test

import (
	"fmt"
	auth "github.com/evmos/evmos/v15/precompiles/authorization"
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

		execRevertedCheck testutil.LogCheckArgs
		failCheck         testutil.LogCheckArgs
		passCheck         testutil.LogCheckArgs
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
		execRevertedCheck = failCheck.WithErrContains("execution reverted")
		passCheck = failCheck.WithExpPass(true)

		err = s.network.NextBlock()
		Expect(err).ToNot(HaveOccurred(), "failed to advance block")
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
	})

	When("querying allowance", func() {
		DescribeTable("it should return an existing allowance", func(callType int) {
			grantee := utiltx.GenerateAddress()
			granter := sender
			expAllowance := big.NewInt(100)

			s.setupSendAuthz(grantee.Bytes(), granter.Priv, sdk.Coins{{s.tokenDenom, sdk.NewIntFromBigInt(expAllowance)}})

			txArgs, allowanceArgs := s.getTxAndCallArgs(callType, contractAddr)
			allowanceArgs.MethodName = auth.AllowanceMethod
			allowanceArgs.Args = []interface{}{granter.Addr, grantee}

			_, ethRes, err := s.callContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, passCheck)
			Expect(err).ToNot(HaveOccurred(), "failed to call contract")

			var allowance *big.Int
			err = s.precompile.UnpackIntoInterface(&allowance, auth.AllowanceMethod, ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			Expect(allowance).To(Equal(expAllowance), "expected different allowance")
		},
			Entry(" - direct call", directCall),
			Entry(" - through contract", contractCall),
		)

		DescribeTable("it should return an error if no allowance exists", func(callType int) {
			grantee := s.keyring.GetAddr(1)
			granter := sender

			balanceGrantee, err := s.grpcHandler.GetBalance(grantee.Bytes(), s.network.GetDenom())
			Expect(err).ToNot(HaveOccurred(), "failed to get balance")
			Expect(balanceGrantee.Balance.Amount.Int64()).ToNot(BeZero(), "expected zero balance")

			txArgs, allowanceArgs := s.getTxAndCallArgs(callType, contractAddr)
			allowanceArgs.MethodName = auth.AllowanceMethod
			allowanceArgs.Args = []interface{}{granter.Addr, grantee}

			noAuthzCheck := failCheck.WithErrContains(
				fmt.Sprintf(auth.ErrAuthzDoesNotExistOrExpired, erc20.SendMsgURL, grantee.String()),
			)
			if callType == contractCall {
				noAuthzCheck = execRevertedCheck
			}

			_, _, err = s.callContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, noAuthzCheck)
			Expect(err).ToNot(HaveOccurred(), "failed to call contract")
		},
			Entry(" - direct call", directCall),
			Entry(" - through contract", contractCall),
		)

		DescribeTable("it should return zero if an allowance exists for other tokens", func(callType int) {
			grantee := s.keyring.GetAddr(1)
			granter := sender
			amount := big.NewInt(100)

			s.setupSendAuthz(grantee.Bytes(), granter.Priv, sdk.Coins{{s.network.GetDenom(), sdk.NewIntFromBigInt(amount)}})

			txArgs, allowanceArgs := s.getTxAndCallArgs(callType, contractAddr)
			allowanceArgs.MethodName = auth.AllowanceMethod
			allowanceArgs.Args = []interface{}{granter.Addr, grantee}

			_, ethRes, err := s.callContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, passCheck)
			Expect(err).ToNot(HaveOccurred(), "failed to call contract")

			var allowance *big.Int
			err = s.precompile.UnpackIntoInterface(&allowance, auth.AllowanceMethod, ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
			Expect(allowance.Int64()).To(BeZero(), "expected zero allowance")
		},
			Entry(" - direct call", directCall),
			Entry(" - through contract", contractCall),
		)

		DescribeTable("it should return an error if the account does not exist", func(callType int) {
			grantee := utiltx.GenerateAddress()
			granter := sender

			txArgs, allowanceArgs := s.getTxAndCallArgs(callType, contractAddr)
			allowanceArgs.MethodName = auth.AllowanceMethod
			allowanceArgs.Args = []interface{}{granter.Addr, grantee}

			noAuthzCheck := failCheck.WithErrContains(
				fmt.Sprintf(auth.ErrAuthzDoesNotExistOrExpired, erc20.SendMsgURL, grantee.String()),
			)
			if callType == contractCall {
				noAuthzCheck = execRevertedCheck
			}

			_, _, err = s.callContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, noAuthzCheck)
			Expect(err).ToNot(HaveOccurred(), "failed to call contract")
		},
			Entry(" - direct call", directCall),
			Entry(" - through contract", contractCall),
		)
	})
})
