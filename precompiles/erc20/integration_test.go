package erc20_test

import (
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/contracts"
	auth "github.com/evmos/evmos/v15/precompiles/authorization"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	"github.com/evmos/evmos/v15/precompiles/erc20/testdata"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/utils"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	erc20types "github.com/evmos/evmos/v15/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

var _ = Describe("ERC20 Extension -", func() {
	var (
		// contractsData holds the addresses and ABIs for the different
		// contract instances that are subject to testing here.
		contractsData ContractsData

		execRevertedCheck testutil.LogCheckArgs
		failCheck         testutil.LogCheckArgs
		passCheck         testutil.LogCheckArgs
	)

	BeforeEach(func() {
		s.SetupTest()

		sender := s.keyring.GetKey(0)
		contractAddr, err := s.factory.DeployContract(
			sender.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: testdata.ERC20AllowanceCallerContract,
				// NOTE: we're passing the precompile address to the constructor because that initiates the contract
				// to make calls to the correct ERC20 precompile.
				ConstructorArgs: []interface{}{s.precompile.Address()},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

		erc20MinterBurnerAddr, err := s.factory.DeployContract(
			sender.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: contracts.ERC20MinterBurnerDecimalsContract,
				ConstructorArgs: []interface{}{
					"Xmpl", "Xmpl", uint8(6),
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC20 minter burner contract")

		ERC20MinterV5Addr, err := s.factory.DeployContract(
			sender.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: contracts.ERC20MinterV5Contract,
				ConstructorArgs: []interface{}{
					"Xmpl", "Xmpl",
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC20 minter contract")

		erc20MinterV5CallerAddr, err := s.factory.DeployContract(
			sender.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: testdata.ERC20AllowanceCallerContract,
				ConstructorArgs: []interface{}{
					ERC20MinterV5Addr,
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC20 minter caller contract")

		// Store the data of the deployed contracts
		contractsData = ContractsData{
			ownerPriv: sender.Priv,
			contractData: map[CallType]ContractData{
				directCall: {
					Address: s.precompile.Address(),
					ABI:     s.precompile.ABI,
				},
				contractCall: {
					Address: contractAddr,
					ABI:     testdata.ERC20AllowanceCallerContract.ABI,
				},
				erc20Call: {
					Address: erc20MinterBurnerAddr,
					ABI:     contracts.ERC20MinterBurnerDecimalsContract.ABI,
				},
				erc20V5Call: {
					Address: ERC20MinterV5Addr,
					ABI:     contracts.ERC20MinterV5Contract.ABI,
				},
				erc20V5CallerCall: {
					Address: erc20MinterV5CallerAddr,
					ABI:     testdata.ERC20AllowanceCallerContract.ABI,
				},
			},
		}

		failCheck = testutil.LogCheckArgs{ABIEvents: s.precompile.Events}
		execRevertedCheck = failCheck.WithErrContains("execution reverted")
		passCheck = failCheck.WithExpPass(true)

		err = s.network.NextBlock()
		Expect(err).ToNot(HaveOccurred(), "failed to advance block")
	})

	Context("basic functionality -", func() {
		When("querying balance", func() {
			DescribeTable("it should return an existing balance", func(callType CallType) {
				sender := s.keyring.GetKey(0)
				expBalance := big.NewInt(100)
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, expBalance.Int64())}

				// Fund account with some tokens
				s.fundWithTokens(callType, contractsData, sender.Addr, fundCoins)

				// Query the balance
				txArgs, balancesArgs := s.getTxAndCallArgs(callType, contractsData, erc20.BalanceOfMethod, sender.Addr)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balance *big.Int
				err = s.precompile.UnpackIntoInterface(&balance, erc20.BalanceOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(balance).To(Equal(expBalance), "expected different balance")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 contract", erc20Call),
				Entry(" - through erc20 v5 contract", erc20V5Call),
				Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
			)

			DescribeTable("it should return zero if balance only exists for other tokens", func(callType CallType) {
				sender := s.keyring.GetKey(0)
				address := utiltx.GenerateAddress()
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 100)}

				// Fund account with some tokens
				err := s.network.FundAccount(sender.AccAddr, fundCoins)
				Expect(err).ToNot(HaveOccurred(), "failed to fund account")

				// Query the balance
				txArgs, balancesArgs := s.getTxAndCallArgs(callType, contractsData, erc20.BalanceOfMethod, address)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balance *big.Int
				err = s.precompile.UnpackIntoInterface(&balance, erc20.BalanceOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(balance.Int64()).To(BeZero(), "expected zero balance")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contracts
				// only support the actual token denomination and don't know of other balances.
			)

			DescribeTable("it should return zero if the account does not exist", func(callType CallType) {
				sender := s.keyring.GetKey(0)
				address := utiltx.GenerateAddress()

				// Query the balance
				txArgs, balancesArgs := s.getTxAndCallArgs(callType, contractsData, erc20.BalanceOfMethod, address)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balance *big.Int
				err = s.precompile.UnpackIntoInterface(&balance, erc20.BalanceOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(balance.Int64()).To(BeZero(), "expected zero balance")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 contract", erc20Call),
				Entry(" - through erc20 v5 contract", erc20V5Call),
				Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
			)
		})

		When("querying allowance", func() {
			DescribeTable("it should return an existing allowance", func(callType CallType) {
				grantee := utiltx.GenerateAddress()
				granter := s.keyring.GetKey(0)
				authzCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				s.setupSendAuthzForContract(callType, contractsData, grantee, granter.Priv, authzCoins)

				txArgs, allowanceArgs := s.getTxAndCallArgs(callType, contractsData, auth.AllowanceMethod, granter.Addr, grantee)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var allowance *big.Int
				err = s.precompile.UnpackIntoInterface(&allowance, auth.AllowanceMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(allowance).To(Equal(authzCoins[0].Amount.BigInt()), "expected different allowance")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 contract", erc20Call),
				Entry(" - through erc20 v5 contract", erc20V5Call),
				Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
			)

			DescribeTable("it should return zero if no allowance exists", func(callType CallType) {
				grantee := s.keyring.GetAddr(1)
				granter := s.keyring.GetKey(0)

				txArgs, allowanceArgs := s.getTxAndCallArgs(callType, contractsData, auth.AllowanceMethod, granter.Addr, grantee)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var allowance *big.Int
				err = s.precompile.UnpackIntoInterface(&allowance, auth.AllowanceMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(allowance.Int64()).To(BeZero(), "expected zero allowance")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 contract", erc20Call),
				Entry(" - through erc20 v5 contract", erc20V5Call),
				Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
			)

			DescribeTable("it should return zero if an allowance exists for other tokens", func(callType CallType) {
				grantee := s.keyring.GetKey(1)
				granter := s.keyring.GetKey(0)
				authzCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 100)}

				s.setupSendAuthz(grantee.AccAddr, granter.Priv, authzCoins)

				txArgs, allowanceArgs := s.getTxAndCallArgs(callType, contractsData, auth.AllowanceMethod, granter.Addr, grantee.Addr)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var allowance *big.Int
				err = s.precompile.UnpackIntoInterface(&allowance, auth.AllowanceMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(allowance.Int64()).To(BeZero(), "expected zero allowance")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
			)

			DescribeTable("it should return zero if the account does not exist", func(callType CallType) {
				grantee := utiltx.GenerateAddress()
				granter := s.keyring.GetKey(0)

				txArgs, allowanceArgs := s.getTxAndCallArgs(callType, contractsData, auth.AllowanceMethod, granter.Addr, grantee)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var allowance *big.Int
				err = s.precompile.UnpackIntoInterface(&allowance, auth.AllowanceMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(allowance.Int64()).To(BeZero(), "expected zero allowance")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 contract", erc20Call),
				Entry(" - through erc20 v5 contract", erc20V5Call),
				Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
			)
		})

		When("querying total supply", func() {
			DescribeTable("it should return the total supply", func(callType CallType) {
				sender := s.keyring.GetKey(0)
				expSupply := big.NewInt(100)
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, expSupply.Int64())}

				// Fund account with some tokens
				s.fundWithTokens(callType, contractsData, sender.Addr, fundCoins)

				// Query the balance
				txArgs, supplyArgs := s.getTxAndCallArgs(callType, contractsData, erc20.TotalSupplyMethod)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var supply *big.Int
				err = s.precompile.UnpackIntoInterface(&supply, erc20.TotalSupplyMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(supply).To(Equal(expSupply), "expected different supply")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 contract", erc20Call),
				Entry(" - through erc20 v5 contract", erc20V5Call),
				Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
			)

			DescribeTable("it should return zero if no tokens exist", func(callType CallType) {
				sender := s.keyring.GetKey(0)
				txArgs, supplyArgs := s.getTxAndCallArgs(callType, contractsData, erc20.TotalSupplyMethod)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var supply *big.Int
				err = s.precompile.UnpackIntoInterface(&supply, erc20.TotalSupplyMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(supply.Int64()).To(BeZero(), "expected zero supply")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 contract", erc20Call),
				Entry(" - through erc20 v5 contract", erc20V5Call),
				Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
			)
		})

		When("transferring tokens", func() {
			DescribeTable("it should transfer tokens to a non-existing address", func(callType CallType, expGasUsed int64) {
				sender := s.keyring.GetKey(0)
				receiver := utiltx.GenerateAddress()
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				transferCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				// Fund account with some tokens
				s.fundWithTokens(callType, contractsData, sender.Addr, fundCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(callType, contractsData, erc20.TransferMethod, receiver, transferCoins[0].Amount.BigInt())

				transferCheck := passCheck.WithExpEvents(erc20.EventTypeTransfer)

				res, _, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectBalancesForContract(
					callType, contractsData,
					[]ExpectedBalance{
						{address: sender.AccAddr, expCoins: fundCoins.Sub(transferCoins...)},
						{address: receiver.Bytes(), expCoins: transferCoins},
					},
				)

				Expect(res.GasUsed).To(Equal(expGasUsed), "expected different gas used")
			},
				// FIXME: The gas used on the precompile is much higher than on the EVM
				Entry(" - direct call", directCall, int64(3_021_572)),
				Entry(" - through erc20 contract", erc20Call, int64(54_381)),
				Entry(" - through erc20 v5 contract", erc20V5Call, int64(52_122)),
			)

			DescribeTable("it should transfer tokens to an existing address", func(callType CallType) {
				sender := s.keyring.GetKey(0)
				receiver := s.keyring.GetKey(1)
				fundCoinsSender := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				fundCoinsReceiver := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 500)}
				transferCoin := sdk.NewInt64Coin(s.tokenDenom, 100)

				// Fund accounts with some tokens
				s.fundWithTokens(callType, contractsData, sender.Addr, fundCoinsSender)
				s.fundWithTokens(callType, contractsData, receiver.Addr, fundCoinsReceiver)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(callType, contractsData, erc20.TransferMethod, receiver.Addr, transferCoin.Amount.BigInt())

				transferCheck := passCheck.WithExpEvents(erc20.EventTypeTransfer)

				_, _, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectBalancesForContract(
					callType, contractsData,
					[]ExpectedBalance{
						{address: sender.AccAddr, expCoins: fundCoinsSender.Sub(transferCoin)},
						{address: receiver.AccAddr, expCoins: fundCoinsReceiver.Add(transferCoin)},
					},
				)
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because transferring using a caller contract
				// is only supported through transferFrom method.
				Entry(" - through erc20 contract", erc20Call),
				Entry(" - through erc20 v5 contract", erc20V5Call),
			)

			DescribeTable("it should return an error trying to call from a smart contract", func(callType CallType) {
				sender := s.keyring.GetKey(0)
				receiver := s.keyring.GetAddr(1)
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				transferCoin := sdk.NewInt64Coin(s.tokenDenom, 100)

				// Fund account with some tokens
				s.fundWithTokens(callType, contractsData, sender.Addr, fundCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(callType, contractsData, erc20.TransferMethod, receiver, transferCoin.Amount.BigInt())

				_, _, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, execRevertedCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				// NOTE: we are not passing the direct call here because this test is specific to the contract calls
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
			)

			DescribeTable("it should return an error if the sender does not have enough tokens", func(callType CallType) {
				sender := s.keyring.GetKey(0)
				receiver := s.keyring.GetAddr(1)
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 200)}
				transferCoin := sdk.NewInt64Coin(s.tokenDenom, 300)

				// Fund account with some tokens
				s.fundWithTokens(callType, contractsData, sender.Addr, fundCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(callType, contractsData, erc20.TransferMethod, receiver, transferCoin.Amount.BigInt())

				insufficientBalanceCheck := failCheck.WithErrContains(
					erc20.ErrTransferAmountExceedsBalance.Error(),
				)

				_, _, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, insufficientBalanceCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test is for direct calls only

				Entry(" - through erc20 contract", erc20Call),
				// // TODO: The ERC20 V5 contract is raising the ERC-6093 standardized error which we are not as of yet
				// Entry(" - through erc20 v5 contract", erc20V5Call),
			)
		})

		When("transferring tokens from another account", func() {
			DescribeTable("it should transfer tokens from another account with a sufficient approval set", func(callType CallType) {
				owner := s.keyring.GetKey(0)
				spender := s.keyring.GetKey(1)
				receiver := utiltx.GenerateAddress()

				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				transferCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				// Fund account with some tokens
				s.fundWithTokens(callType, contractsData, owner.Addr, fundCoins)

				// Set allowance
				s.setupSendAuthzForContract(callType, contractsData, spender.Addr, owner.Priv, transferCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(
					callType, contractsData,
					erc20.TransferFromMethod,
					owner.Addr, receiver, transferCoins[0].Amount.BigInt(),
				)

				transferCheck := passCheck.WithExpEvents(erc20.EventTypeTransfer)

				_, _, err := s.factory.CallContractAndCheckLogs(spender.Priv, txArgs, transferArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectBalancesForContract(
					callType, contractsData,
					[]ExpectedBalance{
						{address: owner.AccAddr, expCoins: fundCoins.Sub(transferCoins...)},
						{address: receiver.Bytes(), expCoins: transferCoins},
					},
				)

				// Check that the allowance was removed since we authorized only the transferred amount
				s.ExpectNoSendAuthzForContract(
					callType, contractsData,
					spender.Addr, owner.Addr,
				)
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test is for direct calls only

				// FIXME: other than the EVM extension, the ERC20 contract emits an additional Approval event (we only emit 1x Transfer)
				// NOTE: Interestingly, the new ERC20 v5 contract does not emit the additional Approval event
				// Entry("- through erc20 contract", erc20Call),
				Entry(" - through erc20 v5 contract", erc20V5Call),
			)

			DescribeTable("it should transfer tokens using a smart contract with a sufficient approval set", func(callType CallType) {
				owner := s.keyring.GetKey(0)
				receiver := utiltx.GenerateAddress()
				fundCoin := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				transferCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				// NOTE: the spender will be the contract address
				spender := contractsData.GetContractData(callType).Address

				// Fund account with some tokens
				s.fundWithTokens(callType, contractsData, owner.Addr, fundCoin)

				// Set allowance
				s.setupSendAuthzForContract(
					callType, contractsData,
					spender, owner.Priv, transferCoins,
				)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(
					callType, contractsData,
					erc20.TransferFromMethod,
					owner.Addr, receiver, transferCoins[0].Amount.BigInt(),
				)

				transferCheck := passCheck.WithExpEvents(erc20.EventTypeTransfer)

				_, _, err := s.factory.CallContractAndCheckLogs(owner.Priv, txArgs, transferArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectBalancesForContract(
					callType, contractsData,
					[]ExpectedBalance{
						{address: owner.AccAddr, expCoins: fundCoin.Sub(transferCoins...)},
						{address: receiver.Bytes(), expCoins: transferCoins},
					},
				)

				// Check that the allowance was removed since we authorized only the transferred amount
				s.ExpectNoSendAuthzForContract(
					callType, contractsData,
					spender, owner.Addr,
				)
			},
				// Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				// NOTE: we are not passing the erc20 contract call here because this is supposed to
				// test external contract calls
				Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
			)

			DescribeTable("it should transfer funds from a smart contract with a sufficient allowance and triggered from another account", func(callType CallType) {
				msgSender := s.keyring.GetKey(0)
				owner := s.keyring.GetKey(1)
				receiver := utiltx.GenerateAddress()

				// NOTE: the spender will be the contract address
				spender := contractsData.GetContractData(callType).Address

				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				transferCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				// Fund account with some tokens
				s.fundWithTokens(callType, contractsData, owner.Addr, fundCoins)

				// Set allowance
				s.setupSendAuthzForContract(
					callType, contractsData,
					spender, owner.Priv, transferCoins,
				)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(
					callType, contractsData,
					erc20.TransferFromMethod,
					owner.Addr, receiver, transferCoins[0].Amount.BigInt(),
				)

				transferCheck := passCheck.WithExpEvents(erc20.EventTypeTransfer)

				_, _, err := s.factory.CallContractAndCheckLogs(msgSender.Priv, txArgs, transferArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				// NOTE: we are not passing the direct call here because this test is specific to the contract calls

				Entry(" - through contract", contractCall),
				Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
			)

			DescribeTable("it should return an error when the spender does not have enough allowance", func(callType CallType) {
				owner := s.keyring.GetKey(0)
				spender := s.keyring.GetKey(1)
				receiver := utiltx.GenerateAddress()
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				authzCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}
				transferCoin := sdk.NewInt64Coin(s.tokenDenom, 200)

				// Fund account with some tokens
				s.fundWithTokens(callType, contractsData, owner.Addr, fundCoins)
				// Set allowance
				s.setupSendAuthzForContract(
					callType, contractsData,
					spender.Addr, owner.Priv, authzCoins,
				)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(
					callType, contractsData,
					erc20.TransferFromMethod,
					owner.Addr, receiver, transferCoin.Amount.BigInt(),
				)

				insufficientAllowanceCheck := failCheck.WithErrContains(erc20.ErrInsufficientAllowance.Error())

				_, _, err := s.factory.CallContractAndCheckLogs(spender.Priv, txArgs, transferArgs, insufficientAllowanceCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test case only covers direct calls

				Entry(" - through erc20 contract", erc20Call),

				// TODO: the ERC20 V5 contract is raising the ERC-6093 standardized error which we are not using as of yet
				// Entry(" - through erc20 v5 contract", erc20V5Call),
			)

			DescribeTable("it should return an error when using smart contract and the spender does not have enough allowance", func(callType CallType) {
				from := s.keyring.GetKey(0)
				receiver := utiltx.GenerateAddress()
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 400)}
				authzCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}
				transferCoin := sdk.NewInt64Coin(s.tokenDenom, 300)

				// NOTE: the spender will be the contract address
				spender := contractsData.GetContractData(callType).Address

				// Fund account with some tokens
				s.fundWithTokens(callType, contractsData, from.Addr, fundCoins)

				// Set allowance
				s.setupSendAuthzForContract(callType, contractsData, spender, from.Priv, authzCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(
					callType, contractsData,
					erc20.TransferFromMethod,
					from.Addr, receiver, transferCoin.Amount.BigInt(),
				)

				_, _, err := s.factory.CallContractAndCheckLogs(from.Priv, txArgs, transferArgs, execRevertedCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				// NOTE: we are not passing the direct call here because this test is for contract calls only
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
			)

			DescribeTable("it should return an error if there is no allowance set", func(callType CallType) {
				sender := s.keyring.GetKey(0)
				from := s.keyring.GetKey(1)
				receiver := utiltx.GenerateAddress()
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				transferCoin := sdk.NewInt64Coin(s.tokenDenom, 100)

				// Fund account with some tokens
				s.fundWithTokens(callType, contractsData, from.Addr, fundCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(
					callType, contractsData,
					erc20.TransferFromMethod,
					from.Addr, receiver, transferCoin.Amount.BigInt(),
				)

				insufficientAllowanceCheck := failCheck.WithErrContains(
					erc20.ErrInsufficientAllowance.Error(),
				)

				_, _, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, insufficientAllowanceCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test case only covers direct calls

				// FIXME: we have a different error here than the EVM extension
				// -- says "ERC20: transfer amount exceeds allowance" instead of "authorization not found"
				Entry(" - through erc20 contract", erc20Call),

				// TODO: the ERC20 V5 contract is raising the ERC-6093 standardized error which we are not using as of yet
				// Entry(" - through erc20 v5 contract", erc20V5Call),
			)

			DescribeTable("it should return an error if the sender does not have enough tokens", func(callType CallType) {
				sender := s.keyring.GetKey(0)
				from := s.keyring.GetKey(1)
				receiver := utiltx.GenerateAddress()
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 200)}
				transferCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}

				// Fund account with some tokens
				s.fundWithTokens(callType, contractsData, from.Addr, fundCoins)

				// Set allowance
				s.setupSendAuthzForContract(
					callType, contractsData,
					sender.Addr, from.Priv, transferCoins,
				)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(callType, contractsData, erc20.TransferFromMethod, from.Addr, receiver, transferCoins[0].Amount.BigInt())

				insufficientBalanceCheck := failCheck.WithErrContains(
					erc20.ErrTransferAmountExceedsBalance.Error(),
				)

				_, _, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, insufficientBalanceCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test case only covers direct calls

				Entry(" - through erc20 contract", erc20Call),

				// TODO: the ERC20 V5 contract is raising the ERC-6093 standardized error which we are not using as of yet
				// Entry(" - through erc20 v5 contract", erc20V5Call),
			)
		})

		When("approving an allowance", func() {
			DescribeTable("it should approve an allowance", func(callType CallType) {
				grantee := s.keyring.GetKey(0)
				granter := s.keyring.GetKey(1)
				transferCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 200)}

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, transferCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check allowance
				s.ExpectSendAuthzForContract(
					callType, contractsData,
					grantee.Addr, granter.Addr, transferCoins,
				)
			},
				Entry(" - direct call", directCall),
				Entry(" - through erc20 contract", erc20Call),
				Entry(" - through erc20 v5 contract", erc20V5Call),

				// TODO: add contract tests
			)

			DescribeTable("it should add a new spend limit to an existing allowance with a different token", func(callType CallType) {
				grantee := s.keyring.GetKey(1)
				granter := s.keyring.GetKey(0)
				bondCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 200)}
				tokenCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				// set up a previous authorization
				s.setupSendAuthz(grantee.AccAddr, granter.Priv, bondCoins)

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, tokenCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check allowance contains both spend limits
				s.expectSendAuthz(grantee.AccAddr, granter.AccAddr, bondCoins.Add(tokenCoins...))
			},
				Entry(" - direct call", directCall),
				// NOTE 2: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.

				// TODO: add contract tests
			)

			DescribeTable("it should set the new spend limit for an existing allowance with the same token", func(callType CallType) {
				grantee := s.keyring.GetKey(1)
				granter := s.keyring.GetKey(0)
				bondCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 200)}
				tokenCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}
				doubleTokenCoin := sdk.NewInt64Coin(s.tokenDenom, 200)

				// set up a previous authorization
				s.setupSendAuthz(grantee.AccAddr, granter.Priv, bondCoins.Add(doubleTokenCoin))

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, tokenCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check allowance contains both spend limits
				s.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, bondCoins.Add(tokenCoins...))
			},
				Entry(" - direct call", directCall),
				Entry(" - through erc20 contract", erc20Call),
				Entry(" - through erc20 v5 contract", erc20V5Call),
				// TODO: add contract tests
			)

			DescribeTable("it should remove the token from the spend limit of an existing authorization when approving zero", func(callType CallType) {
				grantee := s.keyring.GetKey(1)
				granter := s.keyring.GetKey(0)
				bondCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 200)}
				tokenCoin := sdk.NewInt64Coin(s.tokenDenom, 100)

				// set up a previous authorization
				s.setupSendAuthz(grantee.AccAddr, granter.Priv, bondCoins.Add(tokenCoin))

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, common.Big0)

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check allowance contains only the spend limit in network denomination
				s.expectSendAuthz(grantee.AccAddr, granter.AccAddr, bondCoins)
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.

				// TODO: add contract tests
			)

			DescribeTable("it should delete the authorization when approving zero with no other spend limits", func(callType CallType) {
				grantee := s.keyring.GetKey(1)
				granter := s.keyring.GetKey(0)
				tokenCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				// set up a previous authorization
				s.setupSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Priv, tokenCoins)

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, common.Big0)

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check allowance was deleted
				s.expectNoSendAuthz(grantee.AccAddr, granter.AccAddr)
			},
				Entry(" - direct call", directCall),
				Entry(" - through erc20 contract", erc20Call),
				Entry(" - through erc20 v5 contract", erc20V5Call),

				// TODO: add contract tests
			)

			DescribeTable("it should no-op if approving 0 and no allowance exists", func(callType CallType) {
				grantee := s.keyring.GetKey(1)
				granter := s.keyring.GetKey(0)

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, common.Big0)

				// We are expecting an approval to be made, but no authorization stored since it's 0
				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check still no authorization exists
				s.ExpectNoSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr)
			},
				Entry(" - direct call", directCall),
				Entry(" - through erc20 contract", erc20Call),
				Entry(" - through erc20 v5 contract", erc20V5Call),

				// TODO: add contract tests
			)

			DescribeTable("it should create an allowance if the grantee is the same as the granter", func(callType CallType) {
				grantee := s.keyring.GetKey(0)
				granter := s.keyring.GetKey(0)
				authzCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(
					callType, contractsData,
					auth.ApproveMethod,
					grantee.Addr, authzCoins[0].Amount.BigInt(),
				)

				approvalCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approvalCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectSendAuthzForContract(
					callType, contractsData,
					grantee.Addr, granter.Addr, authzCoins,
				)
			},
				Entry(" - direct call", directCall),
				Entry(" - through erc20 contract", erc20Call),
				Entry(" - through erc20 v5 contract", erc20V5Call),

				// TODO: add contract tests
			)

			DescribeTable("it should return an error if approving 0 and allowance only exists for other tokens", func(callType CallType) {
				grantee := s.keyring.GetKey(1)
				granter := s.keyring.GetKey(0)
				bondCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 200)}

				// set up a previous authorization
				s.setupSendAuthz(grantee.AccAddr, granter.Priv, bondCoins)

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, common.Big0)

				notFoundCheck := failCheck.WithErrContains(
					fmt.Sprintf(erc20.ErrNoAllowanceForToken, s.tokenDenom),
				)

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, notFoundCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.

				// TODO: add contract tests
			)
		})
	})

	Context("metadata query -", func() {
		Context("for a token without registered metadata", func() {
			BeforeEach(func() {
				// Deploy ERC20NoMetadata contract for this test
				erc20NoMetadataAddr, err := s.factory.DeployContract(
					s.keyring.GetPrivKey(0),
					evmtypes.EvmTxArgs{},
					factory.ContractDeploymentData{
						Contract: contracts.ERC20NoMetadataContract,
					},
				)
				Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

				// NOTE: update the address but leave the ABI as it is, so that the ABI includes
				// the metadata methods but the contract doesn't have them.
				contractsData.contractData[erc20Call] = ContractData{
					Address: erc20NoMetadataAddr,
					ABI:     contracts.ERC20MinterBurnerDecimalsContract.ABI,
				}
			})

			DescribeTable("querying the name should return an error", func(callType CallType) {
				txArgs, nameArgs := s.getTxAndCallArgs(callType, contractsData, erc20.NameMethod)

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, nameArgs, execRevertedCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				// FIXME: Instead of "not supported" or similar this just returns the general "execution reverted" without any other info
				// -- do we really want the same behavior for the EVM extension?
				Entry(" - through erc20 contract", erc20Call), // NOTE: we're passing the ERC20 contract call here which was adjusted to point to a contract without metadata to expect the same errors
			)

			DescribeTable("querying the symbol should return an error", func(callType CallType) {
				txArgs, symbolArgs := s.getTxAndCallArgs(callType, contractsData, erc20.SymbolMethod)

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, symbolArgs, execRevertedCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				// FIXME: Instead of "not supported" or similar this just returns the general "execution reverted" without any other info
				// -- do we really want the same behavior for the EVM extension?
				Entry(" - through erc20 contract", erc20Call), // NOTE: we're passing the ERC20 contract call here which was adjusted to point to a contract without metadata to expect the same errors
			)

			DescribeTable("querying the decimals should return an error", func(callType CallType) {
				txArgs, decimalsArgs := s.getTxAndCallArgs(callType, contractsData, erc20.DecimalsMethod)

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, decimalsArgs, execRevertedCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				// FIXME: Instead of "not supported" or similar this just returns the general "execution reverted" without any other info
				// -- do we really want the same behavior for the EVM extension?
				Entry(" - through erc20 contract", erc20Call), // NOTE: we're passing the ERC20 contract call here which was adjusted to point to a contract without metadata to expect the same errors
			)
		})

		It("should register the ERC20", func() {
			erc20V5Addr := contractsData.GetContractData(erc20V5Call).Address

			// Register the deployed erc20 contract as a token pair
			_, err := utils.RegisterERC20(s.factory, s.network, utils.ERC20RegistrationData{
				Address:      erc20V5Addr,
				Denom:        s.tokenDenom,
				ProposerPriv: s.keyring.GetPrivKey(0),
			})
			Expect(err).ToNot(HaveOccurred(), "failed to register ERC20 token")
		})

		Context("for a token with available metadata", func() {
			const (
				denom       = "axmpl"
				expSymbol   = "Xmpl"
				expDecimals = uint8(18)
			)

			var (
				erc20Addr common.Address
				expName   string
			)

			BeforeEach(func() {
				erc20Addr = contractsData.GetContractData(erc20V5Call).Address
				expName = erc20types.CreateDenom(erc20Addr.String())

				// Register ERC20 token pair for this test
				tokenPair, err := utils.RegisterERC20(s.factory, s.network, utils.ERC20RegistrationData{
					Address:      erc20Addr,
					Denom:        denom,
					ProposerPriv: s.keyring.GetPrivKey(0),
				})
				Expect(err).ToNot(HaveOccurred(), "failed to register ERC20 token")

				// overwrite the other precompile with this one, so that the test utils like s.getTxAndCallArgs still work.
				s.precompile = s.setupERC20PrecompileForTokenPair(tokenPair)

				// update this in the global contractsData
				contractsData.contractData[directCall] = ContractData{
					Address: s.precompile.Address(),
					ABI:     s.precompile.ABI,
				}

				// Deploy contract calling the ERC20 precompile
				callerAddr, err := s.factory.DeployContract(
					s.keyring.GetPrivKey(0),
					evmtypes.EvmTxArgs{},
					factory.ContractDeploymentData{
						Contract: testdata.ERC20AllowanceCallerContract,
						ConstructorArgs: []interface{}{
							s.precompile.Address(),
						},
					},
				)
				Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

				contractsData.contractData[contractCall] = ContractData{
					Address: callerAddr,
					ABI:     testdata.ERC20AllowanceCallerContract.ABI,
				}
			})

			DescribeTable("querying the name should return the name", func(callType CallType) {
				txArgs, nameArgs := s.getTxAndCallArgs(callType, contractsData, erc20.NameMethod)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, nameArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var name string
				err = s.precompile.UnpackIntoInterface(&name, erc20.NameMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(name).To(Equal(expName), "expected different name")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 v5 contract", erc20V5Call),
			)

			DescribeTable("querying the symbol should return the symbol", func(callType CallType) {
				txArgs, symbolArgs := s.getTxAndCallArgs(callType, contractsData, erc20.SymbolMethod)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, symbolArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var symbol string
				err = s.precompile.UnpackIntoInterface(&symbol, erc20.SymbolMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(symbol).To(Equal(expSymbol), "expected different symbol")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 v5 contract", erc20V5Call),
			)

			DescribeTable("querying the decimals should return the decimals", func(callType CallType) {
				txArgs, decimalsArgs := s.getTxAndCallArgs(callType, contractsData, erc20.DecimalsMethod)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, decimalsArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var decimals uint8
				err = s.precompile.UnpackIntoInterface(&decimals, erc20.DecimalsMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(decimals).To(Equal(expDecimals), "expected different decimals")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 v5 contract", erc20V5Call),
			)
		})
	})

	Context("allowance adjustments -", func() {
		var (
			grantee keyring.Key
			granter keyring.Key
		)

		BeforeEach(func() {
			// Deploying the contract which has the increase / decrease allowance methods
			contractAddr, err := s.factory.DeployContract(
				s.keyring.GetPrivKey(0),
				evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
				factory.ContractDeploymentData{
					Contract:        testdata.ERC20AllowanceCallerContract,
					ConstructorArgs: []interface{}{s.precompile.Address()},
				},
			)
			Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

			contractsData.contractData[erc20CallerCall] = ContractData{
				Address: contractAddr,
				ABI:     testdata.ERC20AllowanceCallerContract.ABI,
			}

			grantee = s.keyring.GetKey(0)
			granter = s.keyring.GetKey(1)
		})

		When("no allowance exists", func() {
			DescribeTable("decreasing the allowance should return an error", func(callType CallType) {
				authzCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				txArgs, decreaseArgs := s.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, authzCoins[0].Amount.BigInt())

				notFoundCheck := execRevertedCheck
				if callType == directCall {
					notFoundCheck = failCheck.WithErrContains(
						fmt.Sprintf(auth.ErrAuthzDoesNotExistOrExpired, erc20.SendMsgURL, grantee.Addr.String()),
					)
				}

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, notFoundCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				Entry(" - through erc20 contract", erc20Call),
				// NOTE: The ERC20 V5 contract does not contain these methods
				// Entry(" - through erc20 v5 contract", erc20V5Call),
				Entry(" - contract call", contractCall),
				Entry(" - through erc20 caller contract", erc20CallerCall),
			)

			// NOTE: We have to split between direct and contract calls here because the ERC20 behavior
			// for approvals is different, so we expect different authorizations here
			Context("in direct calls", func() {
				DescribeTable("increasing the allowance should create a new authorization", func(callType CallType) {
					authzCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

					txArgs, increaseArgs := s.getTxAndCallArgs(callType, contractsData, auth.IncreaseAllowanceMethod, grantee.Addr, authzCoins[0].Amount.BigInt())

					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					s.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, authzCoins)
				},
					Entry(" - direct call", directCall),
					Entry(" - through erc20 contract", erc20Call),
					// NOTE: The ERC20 V5 contract does not contain these methods
					// Entry(" - through erc20 v5 contract", erc20V5Call),
				)
			})

			Context("in contract calls", func() {
				DescribeTable("increasing the allowance should create a new authorization", func(callType CallType) {
					contractAddr := contractsData.GetContractData(callType).Address
					authzCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

					txArgs, increaseArgs := s.getTxAndCallArgs(callType, contractsData, auth.IncreaseAllowanceMethod, grantee.Addr, authzCoins[0].Amount.BigInt())

					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					s.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, contractAddr, authzCoins)
				},
					Entry(" - contract call", contractCall),
					Entry(" - through erc20 caller contract", erc20CallerCall),
				)
			})
		})

		When("an allowance exists for other tokens", func() {
			var bondCoins sdk.Coins

			BeforeEach(func() {
				bondCoins = sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 200)}
				s.setupSendAuthz(grantee.AccAddr, granter.Priv, bondCoins)
			})

			DescribeTable("increasing the allowance should add the token to the spend limit", func(callType CallType) {
				increaseCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				txArgs, increaseArgs := s.getTxAndCallArgs(callType, contractsData, auth.IncreaseAllowanceMethod, grantee.Addr, increaseCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, bondCoins.Add(increaseCoins...))
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.

				// TODO: add contract tests
			)

			DescribeTable("decreasing the allowance should return an error", func(callType CallType) {
				decreaseCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				txArgs, decreaseArgs := s.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, decreaseCoins[0].Amount.BigInt())

				notFoundCheck := execRevertedCheck
				if callType == directCall {
					notFoundCheck = failCheck.WithErrContains(
						fmt.Sprintf(erc20.ErrNoAllowanceForToken, s.tokenDenom),
					)
				}

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, notFoundCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.

				// TODO: add contract tests
			)
		})

		When("an allowance exists for the same token", func() {
			var authzCoins sdk.Coins

			BeforeEach(func() {
				authzCoins = sdk.NewCoins(
					sdk.NewInt64Coin(s.network.GetDenom(), 100),
					sdk.NewInt64Coin(s.tokenDenom, 200),
				)

				s.setupSendAuthz(grantee.AccAddr, granter.Priv, authzCoins)
			})

			DescribeTable("increasing the allowance should increase the spend limit", func(callType CallType) {
				increaseCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				txArgs, increaseArgs := s.getTxAndCallArgs(callType, contractsData, auth.IncreaseAllowanceMethod, grantee.Addr, increaseCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, authzCoins.Add(increaseCoins...))
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.

				// TODO: add contract tests
			)

			DescribeTable("decreasing the allowance should decrease the spend limit", func(callType CallType) {
				decreaseCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				txArgs, decreaseArgs := s.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, decreaseCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, authzCoins.Sub(decreaseCoins...))
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.

				// TODO: add contract tests
			)

			DescribeTable("increasing the allowance beyond the max uint256 value should return an error", func(callType CallType) {
				maxUint256Coins := sdk.Coins{sdk.NewCoin(s.tokenDenom, sdk.NewIntFromBigInt(abi.MaxUint256))}

				txArgs, increaseArgs := s.getTxAndCallArgs(callType, contractsData, auth.IncreaseAllowanceMethod, grantee.Addr, maxUint256Coins[0].Amount.BigInt())

				overflowCheck := execRevertedCheck
				if callType == directCall {
					overflowCheck = failCheck.WithErrContains("integer overflow when increasing")
				}

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, overflowCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.

				// TODO: add contract tests
			)

			DescribeTable("decreasing the allowance to zero should remove the token from the spend limit", func(callType CallType) {
				txArgs, decreaseArgs := s.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, authzCoins.AmountOf(s.tokenDenom).BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check that only the spend limit in the network denomination remains
				bondDenom := s.network.GetDenom()
				expCoins := sdk.Coins{sdk.NewCoin(bondDenom, authzCoins.AmountOf(bondDenom))}
				s.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, expCoins)
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.

				// TODO: switch this around, have most test cases for only the token denom and then one special case for
				// the network denom

				// TODO: add contract tests
			)

			DescribeTable("decreasing the allowance below zero should return an error", func(callType CallType) {
				decreaseCoins := sdk.Coins{sdk.NewCoin(s.tokenDenom, authzCoins.AmountOf(s.tokenDenom).AddRaw(100))}

				txArgs, decreaseArgs := s.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, decreaseCoins[0].Amount.BigInt())

				overflowCheck := execRevertedCheck
				if callType == directCall {
					overflowCheck = failCheck.WithErrContains("subtracted value cannot be greater than existing allowance")
				}

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, overflowCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check that the allowance was not changed
				s.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, authzCoins)
			},
				Entry(" - direct call", directCall),

				// TODO: add contract tests
			)
		})

		When("an allowance exists for only the same token", func() {
			DescribeTable("decreasing the allowance to zero should delete the authorization", func(callType CallType) {
				authzCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				s.setupSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Priv, authzCoins)

				txArgs, decreaseArgs := s.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, authzCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.expectNoSendAuthz(grantee.AccAddr, granter.AccAddr)
			},
				Entry(" - direct call", directCall),
				Entry(" - through erc20 contract", erc20Call),
				// NOTE: The ERC20 V5 contract does not contain these methods
				// Entry(" - through erc20 v5 contract", erc20V5Call),

				// TODO: add contract tests
			)
		})
	})
})

var _ = Describe("ERC20 Extension migration Flows -", func() {
	When("migrating an existing ERC20 token", func() {
		var (
			contractData ContractsData

			tokenDenom  = "xmpl"
			tokenName   = "Xmpl"
			tokenSymbol = strings.ToUpper(tokenDenom)

			supply = sdk.NewInt64Coin(tokenDenom, 1000000000000000000)
		)

		BeforeEach(func() {
			contractOwner := s.keyring.GetKey(0)

			// Deploy an ERC20 contract
			erc20Addr, err := s.factory.DeployContract(
				contractOwner.Priv,
				evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
				factory.ContractDeploymentData{
					Contract: contracts.ERC20MinterV5Contract,
					ConstructorArgs: []interface{}{
						tokenName, tokenSymbol,
					},
				},
			)
			Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

			// NOTE: We need to overwrite the information in the contractData here for this specific
			// deployed contract.
			contractData = ContractsData{
				ownerPriv: contractOwner.Priv,
				contractData: map[CallType]ContractData{
					erc20V5Call: {
						Address: erc20Addr,
						ABI:     contracts.ERC20MinterV5Contract.ABI,
					},
				},
			}

			err = s.network.NextBlock()
			Expect(err).ToNot(HaveOccurred(), "failed to commit block")

			// Register the deployed erc20 contract as a token pair
			_, err = utils.RegisterERC20(s.factory, s.network, utils.ERC20RegistrationData{
				Address:      erc20Addr,
				Denom:        tokenDenom,
				ProposerPriv: contractOwner.Priv,
			})
			Expect(err).ToNot(HaveOccurred(), "failed to register ERC20 token")

			err = s.network.NextBlock()
			Expect(err).ToNot(HaveOccurred(), "failed to commit block")

			// Mint the supply of tokens
			err = s.MintERC20(erc20V5Call, contractData, contractOwner.Addr, supply.Amount.BigInt())
			Expect(err).ToNot(HaveOccurred(), "failed to mint tokens")

			err = s.network.NextBlock()
			Expect(err).ToNot(HaveOccurred(), "failed to commit block")

			// Check that the supply was minted
			s.ExpectBalancesForERC20(erc20V5Call, contractData, []ExpectedBalance{{
				address:  contractOwner.AccAddr,
				expCoins: sdk.Coins{supply},
			}})
		})

		It("should migrate the full token balance to the bank module", func() {
			// TODO: implement test on follow-up PR
			Skip("will be addressed on follow-up PR")

			Expect(true).To(BeFalse(), "not implemented")
		})
	})

	When("migrating an extended ERC20 token (e.g. ERC20Votes)", func() {
		It("should migrate the full token balance to the bank module", func() {
			// TODO: make sure that extended tokens are compatible with the ERC20 extensions
			Skip("not included in first tranche")

			Expect(true).To(BeFalse(), "not implemented")
		})
	})

	When("running the migration logic for a set of existing ERC20 tokens", func() {
		BeforeEach(func() {
			// TODO: Add some ERC20 tokens and then run migration logic
			// TODO: check here that the balance cannot be queried from the bank keeper before migrating the token
		})

		It("should add and enable the corresponding EVM extensions", func() {
			Skip("will be addressed in follow-up PR")

			Expect(true).To(BeFalse(), "not implemented")
		})

		It("should be possible to query the balances through the bank module", func() {
			Skip("will be addressed in follow-up PR")

			Expect(true).To(BeFalse(), "not implemented")
		})

		It("should return all tokens when querying all balances for an account", func() {
			Skip("will be addressed in follow-up PR")

			Expect(true).To(BeFalse(), "not implemented")
		})
	})

	When("registering a native IBC coin", func() {
		BeforeEach(func() {
			// TODO: Add some IBC coins, register the token pair and then run migration logic
		})

		It("should add the corresponding EVM extensions", func() {
			Skip("will be addressed in follow-up PR")

			Expect(true).To(BeFalse(), "not implemented")
		})

		It("should be possible to query the balances using an EVM transaction", func() {
			Skip("will be addressed in follow-up PR")

			Expect(true).To(BeFalse(), "not implemented")
		})
	})

	When("using Evmos (not wEvmos) in smart contracts", func() {
		It("should be using straight Evmos for sending funds in smart contracts", func() {
			Skip("will be addressed in follow-up PR")

			Expect(true).To(BeFalse(), "not implemented")
		})
	})
})
