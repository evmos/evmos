package erc20_test

import (
	"fmt"
	"math/big"

	auth "github.com/evmos/evmos/v15/precompiles/authorization"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/contracts"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	"github.com/evmos/evmos/v15/precompiles/erc20/testdata"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

var _ = Describe("ERC20 Extension -", func() {
	var (
		contractAddr common.Address
		// contractData is a helper struct to hold the addresses and ABIs for the
		// different contract instances that are subject to testing here.
		contractData          ContractData
		erc20MinterBurnerAddr common.Address
		err                   error
		sender                keyring.Key

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
				Contract:        testdata.ERC20AllowanceCallerContract,
				ConstructorArgs: []interface{}{s.precompile.Address()},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

		erc20MinterBurnerAddr, err = s.factory.DeployContract(
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

		contractData = ContractData{
			ownerPriv:      sender.Priv,
			erc20Addr:      erc20MinterBurnerAddr,
			erc20ABI:       contracts.ERC20MinterBurnerDecimalsContract.ABI,
			contractAddr:   contractAddr,
			contractABI:    testdata.ERC20AllowanceCallerContract.ABI,
			precompileAddr: s.precompile.Address(),
			precompileABI:  s.precompile.ABI,
		}

		failCheck = testutil.LogCheckArgs{ABIEvents: s.precompile.Events}
		execRevertedCheck = failCheck.WithErrContains("execution reverted")
		passCheck = failCheck.WithExpPass(true)

		err = s.network.NextBlock()
		Expect(err).ToNot(HaveOccurred(), "failed to advance block")
	})

	Context("basic functionality -", func() {
		When("querying balance", func() {
			DescribeTable("it should return an existing balance", func(callType int) {
				expBalance := big.NewInt(100)
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, expBalance.Int64())}

				// Fund account with some tokens
				s.fundWithTokens(callType, contractData, sender.Addr, fundCoins)

				// Query the balance
				txArgs, balancesArgs := s.getTxAndCallArgs(callType, contractData, erc20.BalanceOfMethod, sender.Addr)

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
			)

			DescribeTable("it should return zero if balance only exists for other tokens", func(callType int) {
				address := utiltx.GenerateAddress()
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 100)}

				// Fund account with some tokens
				err := s.network.FundAccount(sender.AccAddr, fundCoins)
				Expect(err).ToNot(HaveOccurred(), "failed to fund account")

				// Query the balance
				txArgs, balancesArgs := s.getTxAndCallArgs(callType, contractData, erc20.BalanceOfMethod, address)

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

			DescribeTable("it should return zero if the account does not exist", func(callType int) {
				address := utiltx.GenerateAddress()

				// Query the balance
				txArgs, balancesArgs := s.getTxAndCallArgs(callType, contractData, erc20.BalanceOfMethod, address)

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
			)
		})

		When("querying allowance", func() {
			DescribeTable("it should return an existing allowance", func(callType int) {
				grantee := utiltx.GenerateAddress()
				granter := sender
				authzCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				// Make topics for the grantee and granter common.Address types
				granteeTopic, err := cmn.MakeTopic(grantee)
				Expect(err).ToNot(HaveOccurred(), "failed to make topic")
				println("Grantee topic: ", granteeTopic.String())
				granterTopic, err := cmn.MakeTopic(granter.Addr)
				Expect(err).ToNot(HaveOccurred(), "failed to make topic")
				println("Granter topic: ", granterTopic.String())
				erc20Topic, err := cmn.MakeTopic(contractData.erc20Addr)
				Expect(err).ToNot(HaveOccurred(), "failed to make topic")
				println("ERC20 contract topic: ", erc20Topic.String())
				ownerTopic, err := cmn.MakeTopic(sender.Addr)
				Expect(err).ToNot(HaveOccurred(), "failed to make topic")
				println("Owner topic: ", ownerTopic.String())

				s.setupSendAuthzForContract(callType, contractData, grantee, granter.Priv, authzCoins)

				txArgs, allowanceArgs := s.getTxAndCallArgs(callType, contractData, auth.AllowanceMethod, granter.Addr, grantee)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var allowance *big.Int
				err = s.precompile.UnpackIntoInterface(&allowance, auth.AllowanceMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(allowance).To(Equal(authzCoins[0].Amount.BigInt()), "expected different allowance")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				// FIXME: This returns an empty allowance but the logs show that an approval was made - however the approval was made for a different owner??
				Entry(" - through erc20 contract", erc20Call),
			)

			DescribeTable("it should return an error if no allowance exists", func(callType int) {
				grantee := s.keyring.GetAddr(1)
				granter := sender

				txArgs, allowanceArgs := s.getTxAndCallArgs(callType, contractData, auth.AllowanceMethod, granter.Addr, grantee)

				noAuthzCheck := failCheck.WithErrContains(
					fmt.Sprintf(auth.ErrAuthzDoesNotExistOrExpired, erc20.SendMsgURL, grantee.String()),
				)
				if callType == contractCall {
					noAuthzCheck = execRevertedCheck
				}

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, noAuthzCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				// FIXME: other than the extension, the ERC20 contract doesn't return an error but returns a zero allowance
				Entry(" - through erc20 contract", erc20Call),
			)

			DescribeTable("it should return zero if an allowance exists for other tokens", func(callType int) {
				grantee := s.keyring.GetKey(1)
				granter := sender
				authzCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 100)}

				s.setupSendAuthzForContract(callType, contractData, grantee.Addr, granter.Priv, authzCoins)

				txArgs, allowanceArgs := s.getTxAndCallArgs(callType, contractData, auth.AllowanceMethod, granter.Addr, grantee)

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

			DescribeTable("it should return an error if the account does not exist", func(callType int) {
				grantee := utiltx.GenerateAddress()
				granter := sender

				txArgs, allowanceArgs := s.getTxAndCallArgs(callType, contractData, auth.AllowanceMethod, granter.Addr, grantee)

				noAuthzCheck := failCheck.WithErrContains(
					fmt.Sprintf(auth.ErrAuthzDoesNotExistOrExpired, erc20.SendMsgURL, grantee.String()),
				)
				if callType == contractCall {
					noAuthzCheck = execRevertedCheck
				}

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, noAuthzCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				// FIXME: Other than the extension, the ERC20 contract doesn't return an error but returns a zero allowance
				Entry(" - through erc20 contract", erc20Call),
			)
		})

		When("querying total supply", func() {
			DescribeTable("it should return the total supply", func(callType int) {
				expSupply := big.NewInt(100)
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, expSupply.Int64())}

				// Fund account with some tokens
				s.fundWithTokens(callType, contractData, sender.Addr, fundCoins)

				// Query the balance
				txArgs, supplyArgs := s.getTxAndCallArgs(callType, contractData, erc20.TotalSupplyMethod)

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
			)

			DescribeTable("it should return zero if no tokens exist", func(callType int) {
				txArgs, supplyArgs := s.getTxAndCallArgs(callType, contractData, erc20.TotalSupplyMethod)

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
			)
		})

		When("transferring tokens", func() {
			DescribeTable("it should transfer tokens to a non-existing address", func(callType int) {
				receiver := utiltx.GenerateAddress()
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				transferCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				// Fund account with some tokens
				s.fundWithTokens(callType, contractData, sender.Addr, fundCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(callType, contractData, erc20.TransferMethod, receiver, transferCoins[0].Amount.BigInt())

				transferCheck := passCheck.WithExpEvents(erc20.EventTypeTransfer)

				res, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectBalancesForContract(
					callType, contractData,
					[]ExpectedBalance{
						{address: sender.AccAddr, expCoins: fundCoins.Sub(transferCoins...)},
						{address: receiver.Bytes(), expCoins: transferCoins},
					},
				)

				// TODO: Check gas
				fmt.Println("Gas used (res): ", res.GasUsed)
				fmt.Println("Gas used (ethRes): ", ethRes.GasUsed)
				// Expect(res.GasUsed).To(Equal(uint64(0)), "expected different gas used")
				// Expect(ethRes.GasUsed).To(Equal(1), "expected different gas used")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because that requires an authorization which is
				// a separate test case.
				Entry(" - through erc20 contract", erc20Call),
			)

			DescribeTable("it should transfer tokens to an existing address", func(callType int) {
				receiver := s.keyring.GetKey(1)
				fundCoinsSender := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				fundCoinsReceiver := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 500)}
				transferCoin := sdk.NewInt64Coin(s.tokenDenom, 100)

				// Fund accounts with some tokens
				s.fundWithTokens(callType, contractData, sender.Addr, fundCoinsSender)
				s.fundWithTokens(callType, contractData, receiver.Addr, fundCoinsReceiver)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(callType, contractData, erc20.TransferMethod, receiver.Addr, transferCoin.Amount.BigInt())

				transferCheck := passCheck.WithExpEvents(erc20.EventTypeTransfer)

				_, _, err = s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectBalances([]ExpectedBalance{
					{address: sender.AccAddr, expCoins: fundCoinsSender.Sub(transferCoin)},
					{address: receiver.AccAddr, expCoins: fundCoinsReceiver.Add(transferCoin)},
				})
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because that requires an authorization which is
				// a separate test case.
				Entry(" - through erc20 contract", erc20Call),
			)

			DescribeTable("it should return an error trying to call from a smart contract", func(callType int) {
				receiver := s.keyring.GetAddr(1)
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				transferCoin := sdk.NewInt64Coin(s.tokenDenom, 100)

				// Fund account with some tokens
				s.fundWithTokens(callType, contractData, sender.Addr, fundCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(callType, contractData, erc20.TransferMethod, receiver, transferCoin.Amount.BigInt())

				_, _, err = s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, execRevertedCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				// NOTE: we are not passing the direct call here because this test is specific to the contract calls
				Entry(" - through contract", contractCall),
			)

			DescribeTable("it should return an error if the sender does not have enough tokens", func(callType int) {
				receiver := s.keyring.GetAddr(1)
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 200)}
				transferCoin := sdk.NewInt64Coin(s.tokenDenom, 300)

				// Fund account with some tokens
				s.fundWithTokens(callType, contractData, sender.Addr, fundCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(callType, contractData, erc20.TransferMethod, receiver, transferCoin.Amount.BigInt())

				insufficientBalanceCheck := failCheck.WithErrContains(
					"spendable balance 200xmpl is smaller than 300xmpl: insufficient funds",
				)

				_, _, err = s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, insufficientBalanceCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test is for direct calls only

				// FIXME: This error message currently is different from the EVM extension message
				Entry(" - through erc20 contract", erc20Call),
			)
		})

		When("transferring tokens from another account", func() {
			DescribeTable("it should transfer tokens from another account with a sufficient approval set", func(callType int) {
				owner := sender
				spender := s.keyring.GetKey(1)
				receiver := utiltx.GenerateAddress()

				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				transferCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				// Fund account with some tokens
				s.fundWithTokens(callType, contractData, owner.Addr, fundCoins)

				// Set allowance
				s.setupSendAuthzForContract(callType, contractData, spender.Addr, owner.Priv, transferCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(callType, contractData, erc20.TransferFromMethod, owner.Addr, receiver, transferCoins[0].Amount.BigInt())

				transferCheck := passCheck.WithExpEvents(erc20.EventTypeTransfer)

				_, _, err = s.factory.CallContractAndCheckLogs(spender.Priv, txArgs, transferArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectBalancesForContract(
					callType, contractData,
					[]ExpectedBalance{
						{address: owner.AccAddr, expCoins: fundCoins.Sub(transferCoins...)},
						{address: receiver.Bytes(), expCoins: transferCoins},
					},
				)

				// Check that the allowance was removed since we authorized only the transferred amount
				//
				// TODO: adjust for original ERC20 calls
				s.expectNoSendAuthz(spender.AccAddr, owner.AccAddr)
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test is for direct calls only

				// FIXME: this fails currently because of the error setting up the allowance (see above)
				Entry("- through erc20 contract", erc20Call),
			)

			DescribeTable("it should transfer tokens using a smart contract with a sufficient approval set", func(callType int) {
				owner := sender
				spender := contractAddr // NOTE: in case of a contract call the spender is the contract itself
				receiver := utiltx.GenerateAddress()
				fundCoin := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				transferCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				// Fund account with some tokens
				s.fundWithTokens(callType, contractData, owner.Addr, fundCoin)

				// Set allowance
				s.setupSendAuthz(spender.Bytes(), owner.Priv, transferCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(
					callType, contractData,
					erc20.TransferFromMethod,
					owner.Addr, receiver, transferCoins[0].Amount.BigInt(),
				)

				transferCheck := passCheck.WithExpEvents(erc20.EventTypeTransfer)

				_, _, err = s.factory.CallContractAndCheckLogs(owner.Priv, txArgs, transferArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectBalances([]ExpectedBalance{
					{address: owner.AccAddr, expCoins: fundCoin.Sub(transferCoins...)},
					{address: receiver.Bytes(), expCoins: transferCoins},
				})

				// Check that the allowance was removed since we authorized only the transferred amount
				s.expectNoSendAuthz(spender.Bytes(), owner.AccAddr)
			},
				// Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
			)

			DescribeTable("it should return an error trying to send using a smart contract but triggered from another account", func(callType int) {
				msgSender := s.keyring.GetKey(0)
				owner := s.keyring.GetKey(1)
				receiver := utiltx.GenerateAddress()
				spender := contractAddr

				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				transferCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				// Fund account with some tokens
				s.fundWithTokens(callType, contractData, owner.Addr, fundCoins)

				// Set allowance
				s.setupSendAuthz(spender.Bytes(), owner.Priv, transferCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(
					callType, contractData,
					erc20.TransferFromMethod,
					owner.Addr, receiver, transferCoins[0].Amount.BigInt(),
				)

				_, _, err = s.factory.CallContractAndCheckLogs(msgSender.Priv, txArgs, transferArgs, execRevertedCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				// NOTE: we are not passing the direct call here because this test is specific to the contract calls

				// FIXME: This is working right now! We should probably block this.
				Entry(" - through contract", contractCall),
				// NOTE: this fails currently because of the error setting up the approval (see above)
				Entry(" - through erc20 contract", erc20Call),
			)

			DescribeTable("it should return an error when the spender does not have enough allowance", func(callType int) {
				owner := sender
				spender := s.keyring.GetKey(1)
				receiver := utiltx.GenerateAddress()
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				authzCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}
				transferCoin := sdk.NewInt64Coin(s.tokenDenom, 200)

				// Fund account with some tokens
				s.fundWithTokens(callType, contractData, owner.Addr, fundCoins)
				// Set allowance
				s.setupSendAuthzForContract(
					callType, contractData,
					spender.Addr, owner.Priv, authzCoins,
				)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(callType, contractData, erc20.TransferFromMethod, owner.Addr, receiver, transferCoin.Amount.BigInt())

				insufficientAllowanceCheck := failCheck.WithErrContains("requested amount is more than spend limit")

				_, _, err = s.factory.CallContractAndCheckLogs(spender.Priv, txArgs, transferArgs, insufficientAllowanceCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test case only covers direct calls

				// FIXME: we have a different error here than the EVM extension
				Entry(" - through erc20 contract", erc20Call),
			)

			DescribeTable("it should return an error when using smart contract and the spender does not have enough allowance", func(callType int) {
				from := sender
				spender := contractAddr // NOTE: in case of a contract call the spender is the contract itself
				receiver := utiltx.GenerateAddress()
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 400)}
				authzCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}
				transferCoin := sdk.NewInt64Coin(s.tokenDenom, 300)

				// Fund account with some tokens
				s.fundWithTokens(callType, contractData, from.Addr, fundCoins)

				// Set allowance
				s.setupSendAuthz(spender.Bytes(), from.Priv, authzCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(callType, contractData, erc20.TransferFromMethod, from.Addr, receiver, transferCoin.Amount.BigInt())

				_, _, err = s.factory.CallContractAndCheckLogs(from.Priv, txArgs, transferArgs, execRevertedCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				// NOTE: we are not passing the direct call here because this test is for contract calls only
				Entry(" - through contract", contractCall),
			)

			DescribeTable("it should return an error if there is no allowance set", func(callType int) {
				from := s.keyring.GetKey(1)
				receiver := utiltx.GenerateAddress()
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}
				transferCoin := sdk.NewInt64Coin(s.tokenDenom, 100)

				// Fund account with some tokens
				s.fundWithTokens(callType, contractData, from.Addr, fundCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(callType, contractData, erc20.TransferFromMethod, from.Addr, receiver, transferCoin.Amount.BigInt())

				insufficientAllowanceCheck := failCheck.WithErrContains(
					"authorization not found",
				)

				_, _, err = s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, insufficientAllowanceCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test case only covers direct calls

				// FIXME: we have a different error here than the EVM extension
				Entry(" - through erc20 contract", erc20Call),
			)

			DescribeTable("it should return an error if the sender does not have enough tokens", func(callType int) {
				from := s.keyring.GetKey(1)
				receiver := utiltx.GenerateAddress()
				fundCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 200)}
				transferCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 300)}

				// Fund account with some tokens
				s.fundWithTokens(callType, contractData, from.Addr, fundCoins)

				// Set allowance
				s.setupSendAuthz(sender.AccAddr, from.Priv, transferCoins)

				// Transfer tokens
				txArgs, transferArgs := s.getTxAndCallArgs(callType, contractData, erc20.TransferFromMethod, from.Addr, receiver, transferCoins[0].Amount.BigInt())

				insufficientBalanceCheck := failCheck.WithErrContains(
					"spendable balance 200xmpl is smaller than 300xmpl: insufficient funds",
				)

				_, _, err = s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, insufficientBalanceCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test case only covers direct calls

				// FIXME: we have a different error here than the EVM extension
				Entry(" - through erc20 contract", erc20Call),
			)
		})

		When("approving an allowance", func() {
			DescribeTable("it should approve an allowance", func(callType int) {
				grantee := s.keyring.GetKey(0)
				granter := s.keyring.GetKey(1)
				transferCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 200)}

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractData, auth.ApproveMethod, grantee.Addr, transferCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check allowance
				s.ExpectSendAuthzForContract(
					callType, contractData,
					grantee.Addr, granter.Addr, transferCoins,
				)
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test case only covers direct calls

				// FIXME: erc20 approval bug (see above)
				Entry(" - through erc20 contract", erc20Call),
			)

			DescribeTable("it should add a new spend limit to an existing allowance with a different token", func(callType int) {
				grantee := s.keyring.GetKey(1)
				granter := sender
				bondCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 200)}
				tokenCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				// Setup a previous authorization
				s.setupSendAuthz(grantee.AccAddr, granter.Priv, bondCoins)

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractData, auth.ApproveMethod, grantee.Addr, tokenCoins[0].Amount.BigInt())
				approveArgs.MethodName = auth.ApproveMethod
				approveArgs.Args = []interface{}{grantee.Addr, tokenCoins[0].Amount.BigInt()}

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check allowance contains both spend limits
				s.ExpectSendAuthzForContract(callType, contractData, grantee.Addr, granter.Addr, bondCoins.Add(tokenCoins...))
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test case only covers direct calls
				//
				// NOTE 2: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
			)

			DescribeTable("it should set the new spend limit for an existing allowance with the same token", func(callType int) {
				grantee := s.keyring.GetKey(1)
				granter := sender
				bondCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 200)}
				tokenCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}
				doubleTokenCoin := sdk.NewInt64Coin(s.tokenDenom, 200)

				// Setup a previous authorization
				s.setupSendAuthz(grantee.AccAddr, granter.Priv, bondCoins.Add(doubleTokenCoin))

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractData, auth.ApproveMethod, grantee.Addr, tokenCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check allowance contains both spend limits
				s.ExpectSendAuthzForContract(callType, contractData, grantee.Addr, granter.Addr, bondCoins.Add(tokenCoins...))
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test case only covers direct calls
				Entry(" - through erc20 contract", erc20Call),
			)

			DescribeTable("it should remove the token from the spend limit of an existing authorization when approving zero", func(callType int) {
				grantee := s.keyring.GetKey(1)
				granter := sender
				bondCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 200)}
				tokenCoin := sdk.NewInt64Coin(s.tokenDenom, 100)

				// Setup a previous authorization
				s.setupSendAuthzForContract(callType, contractData, grantee.Addr, granter.Priv, bondCoins.Add(tokenCoin))

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractData, auth.ApproveMethod, grantee.Addr, common.Big0)

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check allowance contains only the spend limit in network denomination
				s.ExpectSendAuthzForContract(callType, contractData, grantee.Addr, granter.Addr, bondCoins)
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test case only covers direct calls
				// NOTE 2: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
			)

			DescribeTable("it should delete the authorization when approving zero with no other spend limits", func(callType int) {
				grantee := s.keyring.GetKey(1)
				granter := sender
				tokenCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				// Setup a previous authorization
				s.setupSendAuthzForContract(callType, contractData, grantee.Addr, granter.Priv, tokenCoins)

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractData, auth.ApproveMethod, grantee.Addr, common.Big0)

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check allowance was deleted
				s.expectNoSendAuthz(grantee.AccAddr, granter.AccAddr)
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test case only covers direct calls
				Entry(" - through erc20 contract", erc20Call),
			)

			DescribeTable("it should return an error if approving 0 and no allowance exists", func(callType int) {
				grantee := s.keyring.GetKey(1)
				granter := sender

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractData, auth.ApproveMethod, grantee.Addr, common.Big0)

				nonPosCheck := failCheck.WithErrContains("cannot approve non-positive values")

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, nonPosCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check allowance contains both spend limits
				authzs, err := s.grpcHandler.GetAuthorizations(grantee.AccAddr.String(), granter.AccAddr.String())
				Expect(err).ToNot(HaveOccurred(), "failed to get authorizations")
				Expect(authzs).To(HaveLen(0), "expected different number of authorizations")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test case only covers direct calls
				Entry(" - through erc20 contract", erc20Call),
			)

			// TODO: this is passing right now?? Should we allow someone to create an authorization for themselves?
			DescribeTable("it should return an error if the grantee is the same as the granter", func(callType int) {
				grantee := sender
				granter := sender
				amount := big.NewInt(100)

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractData, auth.ApproveMethod, grantee.Addr, amount)

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, execRevertedCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test case only covers direct calls
				Entry(" - through erc20 contract", erc20Call),
			)

			DescribeTable("it should return an error if approving 0 and allowance only exists for other tokens", func(callType int) {
				grantee := s.keyring.GetKey(1)
				granter := sender
				bondCoins := sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 200)}

				// Setup a previous authorization
				s.setupSendAuthzForContract(callType, contractData, grantee.Addr, granter.Priv, bondCoins)

				// Approve allowance
				txArgs, approveArgs := s.getTxAndCallArgs(callType, contractData, auth.ApproveMethod, grantee.Addr, common.Big0)

				notFoundCheck := failCheck.WithErrContains(
					fmt.Sprintf("allowance for token %s does not exist", s.tokenDenom),
				)

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, notFoundCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test case only covers direct calls
				//
				// NOTE 2: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
			)
		})
	})

	Context("metadata query -", func() {
		Context("for a non-IBC token without registered metadata", func() {
			BeforeEach(func() {
				// Deploy ERC20NoMetadata contract for this test
				erc20Addr, err := s.factory.DeployContract(
					sender.Priv,
					evmtypes.EvmTxArgs{},
					factory.ContractDeploymentData{
						Contract: contracts.ERC20NoMetadataContract,
					},
				)
				Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

				// NOTE: update the address but leave the ABI as it is, so that the ABI includes
				// the metadata methods but the contract doesn't have them.
				contractData.erc20Addr = erc20Addr
			})

			DescribeTable("querying the name should return an error", func(callType int) {
				txArgs, nameArgs := s.getTxAndCallArgs(callType, contractData, erc20.NameMethod)

				noIBCVoucherCheck := failCheck.WithErrContains(
					fmt.Sprintf("denom is not an IBC voucher: %s", s.tokenDenom),
				)
				if callType == contractCall {
					noIBCVoucherCheck = execRevertedCheck
				}

				_, _, err = s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, nameArgs, noIBCVoucherCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				// FIXME: Instead of "not supported" or similar this just returns the general "execution reverted" without any other info
				// -- do we really want the same behavior for the EVM extension?
				Entry(" - through erc20 contract", erc20Call), // NOTE: we're passing the ERC20 contract call here which was adjusted to point to a contract without metadata to expect the same errors
			)

			DescribeTable("querying the symbol should return an error", func(callType int) {
				txArgs, symbolArgs := s.getTxAndCallArgs(callType, contractData, erc20.SymbolMethod)

				noIBCVoucherCheck := failCheck.WithErrContains(
					fmt.Sprintf("denom is not an IBC voucher: %s", s.tokenDenom),
				)
				if callType == contractCall {
					noIBCVoucherCheck = execRevertedCheck
				}

				_, _, err = s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, symbolArgs, noIBCVoucherCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				// FIXME: Instead of "not supported" or similar this just returns the general "execution reverted" without any other info
				// -- do we really want the same behavior for the EVM extension?
				Entry(" - through erc20 contract", erc20Call), // NOTE: we're passing the ERC20 contract call here which was adjusted to point to a contract without metadata to expect the same errors
			)

			DescribeTable("querying the decimals should return an error", func(callType int) {
				txArgs, decimalsArgs := s.getTxAndCallArgs(callType, contractData, erc20.DecimalsMethod)

				noIBCVoucherCheck := failCheck.WithErrContains(
					fmt.Sprintf("denom is not an IBC voucher: %s", s.tokenDenom),
				)
				if callType == contractCall {
					noIBCVoucherCheck = execRevertedCheck
				}

				_, _, err = s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, decimalsArgs, noIBCVoucherCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				// FIXME: Instead of "not supported" or similar this just returns the general "execution reverted" without any other info
				// -- do we really want the same behavior for the EVM extension?
				Entry(" - through erc20 contract", erc20Call), // NOTE: we're passing the ERC20 contract call here which was adjusted to point to a contract without metadata to expect the same errors
			)
		})

		Context("for an IBC coin without registered metadata", func() {
			const (
				denom       = "uxmpl"
				expName     = "Xmpl"
				expSymbol   = "XMPL"
				expDecimals = uint8(6)
			)

			BeforeEach(func() {
				// Deploy new precompile for this test using the xmpl ibc voucher
				//
				// NOTE: this is not the same as the one used in the other tests
				// because we need to set the metadata for this one.
				ibcDenomTrace := types.DenomTrace{Path: "channel-0", BaseDenom: denom}

				// overwrite the other precompile with this one, so that the test utils like s.getTxAndCallArgs still work.
				s.precompile = s.setupERC20Precompile(ibcDenomTrace.IBCDenom())

				// update this in the global contractData
				contractData.precompileABI = s.precompile.ABI
				contractData.precompileAddr = s.precompile.Address()

				// TODO: can I handle this differently, just using the integration utils and not using the keeper directly?
				s.network.App.TransferKeeper.SetDenomTrace(s.network.GetContext(), ibcDenomTrace)
			})

			DescribeTable("querying the name should return the name", func(callType int) {
				txArgs, nameArgs := s.getTxAndCallArgs(directCall, contractData, erc20.NameMethod)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, nameArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var name string
				err = s.precompile.UnpackIntoInterface(&name, erc20.NameMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(name).To(Equal(expName), "expected different name")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 contract", erc20Call), // NOTE: we're passing the ERC20 contract call here because it also has Metadata that can be queried
			)

			DescribeTable("querying the symbol should return the symbol", func(callType int) {
				txArgs, symbolArgs := s.getTxAndCallArgs(directCall, contractData, erc20.SymbolMethod)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, symbolArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var symbol string
				err = s.precompile.UnpackIntoInterface(&symbol, erc20.SymbolMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(symbol).To(Equal(expSymbol), "expected different symbol")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 contract", erc20Call), // NOTE: we're passing the ERC20 contract call here because it also has Metadata that can be queried
			)

			DescribeTable("querying the decimals should return the decimals", func(callType int) {
				txArgs, decimalsArgs := s.getTxAndCallArgs(callType, contractData, erc20.DecimalsMethod)

				_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, decimalsArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var decimals uint8
				err = s.precompile.UnpackIntoInterface(&decimals, erc20.DecimalsMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(decimals).To(Equal(expDecimals), "expected different decimals")
			},
				Entry(" - direct call", directCall),
				// FIXME: this is failing??
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 contract", erc20Call), // NOTE: we're passing the ERC20 contract call here because it also has Metadata that can be queried
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
			contractAddr, err = s.factory.DeployContract(
				sender.Priv,
				evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
				factory.ContractDeploymentData{
					Contract:        testdata.ERC20AllowanceCallerContract,
					ConstructorArgs: []interface{}{s.precompile.Address()},
				},
			)
			Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

			grantee = s.keyring.GetKey(0)
			granter = s.keyring.GetKey(1)
		})

		When("no allowance exists", func() {
			DescribeTable("increasing the allowance should create a new authorization", func(callType int) {
				authzCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				txArgs, increaseArgs := s.getTxAndCallArgs(callType, contractData, auth.IncreaseAllowanceMethod, grantee.Addr, authzCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectSendAuthzForContract(callType, contractData, grantee.Addr, granter.Addr, authzCoins)
			},
				Entry(" - direct call", directCall),
				// FIXME: This is also not creating the authorization from the granter to the grantee but from the contract to the grantee.
				Entry(" - through contract", contractCall),
			)

			DescribeTable("decreasing the allowance should return an error", func(callType int) {
				authzCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				txArgs, decreaseArgs := s.getTxAndCallArgs(callType, contractData, auth.DecreaseAllowanceMethod, grantee.Addr, authzCoins[0].Amount.BigInt())

				notFoundCheck := execRevertedCheck
				if callType == directCall {
					notFoundCheck = failCheck.WithErrContains(
						fmt.Sprintf(auth.ErrAuthzDoesNotExistOrExpired, erc20.SendMsgURL, grantee.Addr.String()),
					)
				}

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, notFoundCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
			)
		})

		When("an allowance exists for other tokens", func() {
			var bondCoins sdk.Coins

			BeforeEach(func() {
				bondCoins = sdk.Coins{sdk.NewInt64Coin(s.network.GetDenom(), 200)}
				s.setupSendAuthz(grantee.AccAddr, granter.Priv, bondCoins)
			})

			DescribeTable("increasing the allowance should add the token to the spend limit", func(callType int) {
				increaseCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				txArgs, increaseArgs := s.getTxAndCallArgs(callType, contractData, auth.IncreaseAllowanceMethod, grantee.Addr, increaseCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectSendAuthzForContract(callType, contractData, grantee.Addr, granter.Addr, bondCoins.Add(increaseCoins...))
			},
				Entry(" - direct call", directCall),
				// FIXME: failing for the same reason
				Entry(" - through contract", contractCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
			)

			DescribeTable("decreasing the allowance should return an error", func(callType int) {
				decreaseCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				txArgs, decreaseArgs := s.getTxAndCallArgs(callType, contractData, auth.DecreaseAllowanceMethod, grantee.Addr, decreaseCoins[0].Amount.BigInt())

				notFoundCheck := execRevertedCheck
				if callType == directCall {
					notFoundCheck = failCheck.WithErrContains(
						fmt.Sprintf("allowance for token %s does not exist", s.tokenDenom),
					)
				}

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, notFoundCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
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

			DescribeTable("increasing the allowance should increase the spend limit", func(callType int) {
				increaseCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				txArgs, increaseArgs := s.getTxAndCallArgs(callType, contractData, auth.IncreaseAllowanceMethod, grantee.Addr, increaseCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectSendAuthzForContract(callType, contractData, grantee.Addr, granter.Addr, authzCoins.Add(increaseCoins...))
			},
				Entry(" - direct call", directCall),
				// FIXME: this also shows interesting behavior because when calling this there is an authorization from the contract to the grantee
				// instead of the increasing the one from granter to the grantee, because the granter is always taken as the contract caller (in this case the smart contract),
				// even though we sign with the granter key. I think this is different to how we have it implemented for other precompiles, e.g. staking?
				//
				// See IncreaseAllowance method in the approve.go file:
				//
				// ```
				// granter := contract.CallerAddress
				// ```
				Entry(" - through contract", contractCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
			)

			DescribeTable("decreasing the allowance should decrease the spend limit", func(callType int) {
				decreaseCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				txArgs, decreaseArgs := s.getTxAndCallArgs(callType, contractData, auth.DecreaseAllowanceMethod, grantee.Addr, decreaseCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.ExpectSendAuthzForContract(callType, contractData, grantee.Addr, granter.Addr, authzCoins.Sub(decreaseCoins...))
			},
				Entry(" - direct call", directCall),
				// FIXME: This is failing for the same reason as the increase allowance test above.
				// It tries to decrease from the contract to the grantee (which doesn't exist) instead of the granter to the grantee.
				Entry(" - through contract", contractCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
			)

			DescribeTable("increasing the allowance beyond the max uint256 value should return an error", func(callType int) {
				maxUint256Coins := sdk.Coins{sdk.NewCoin(s.tokenDenom, sdk.NewIntFromBigInt(abi.MaxUint256))}

				txArgs, increaseArgs := s.getTxAndCallArgs(callType, contractData, auth.IncreaseAllowanceMethod, grantee.Addr, maxUint256Coins[0].Amount.BigInt())

				overflowCheck := execRevertedCheck
				if callType == directCall {
					overflowCheck = failCheck.WithErrContains("integer overflow when increasing")
				}

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, overflowCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// FIXME: Failing
				Entry(" - through contract", contractCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
			)

			DescribeTable("decreasing the allowance to zero should remove the token from the spend limit", func(callType int) {
				txArgs, decreaseArgs := s.getTxAndCallArgs(callType, contractData, auth.DecreaseAllowanceMethod, grantee.Addr, authzCoins.AmountOf(s.tokenDenom).BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check that only the spend limit in the network denomination remains
				bondDenom := s.network.GetDenom()
				expCoins := sdk.Coins{sdk.NewCoin(bondDenom, authzCoins.AmountOf(bondDenom))}
				s.ExpectSendAuthzForContract(callType, contractData, grantee.Addr, granter.Addr, expCoins)
			},
				Entry(" - direct call", directCall),
				// FIXME: Failing for same reason
				Entry(" - through contract", contractCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
				//
				// TODO: switch this around, have most test cases for only the token denom and then one special case for
				// the network denom
			)

			DescribeTable("decreasing the allowance below zero should return an error", func(callType int) {
				decreaseCoins := sdk.Coins{sdk.NewCoin(s.tokenDenom, authzCoins.AmountOf(s.tokenDenom).AddRaw(100))}

				txArgs, decreaseArgs := s.getTxAndCallArgs(callType, contractData, auth.DecreaseAllowanceMethod, grantee.Addr, decreaseCoins[0].Amount.BigInt())

				overflowCheck := execRevertedCheck
				if callType == directCall {
					overflowCheck = failCheck.WithErrContains("subtracted value cannot be greater than existing allowance")
				}

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, overflowCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check that the allowance was not changed
				s.ExpectSendAuthzForContract(callType, contractData, grantee.Addr, granter.Addr, authzCoins)
			},
				Entry(" - direct call", directCall),
				// FIXME: It's expected to fail with "execution reverted" but fails for the wrong reason (see above)
				Entry(" - through contract", contractCall),
			)
		})

		When("an allowance exists for only the same token", func() {
			DescribeTable("decreasing the allowance to zero should delete the authorization", func(callType int) {
				authzCoins := sdk.Coins{sdk.NewInt64Coin(s.tokenDenom, 100)}

				s.setupSendAuthzForContract(callType, contractData, grantee.Addr, granter.Priv, authzCoins)

				txArgs, decreaseArgs := s.getTxAndCallArgs(callType, contractData, auth.DecreaseAllowanceMethod, grantee.Addr, authzCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				s.expectNoSendAuthz(grantee.AccAddr, granter.AccAddr)
			},
				Entry(" - direct call", directCall),
				// FIXME: failing for same reason
				Entry(" - through contract", contractCall),
				// FIXME: failing because of erc20 approval bug (see above)
				Entry(" - through erc20 contract", erc20Call),
			)
		})
	})
})

var _ = Describe("ERC20 Extension - migration Flows -", func() {
	When("migrating an existing ERC20 token", func() {
		It("should migrate the full token balance to the bank module", func() {
			Expect(true).To(BeFalse(), "not implemented")
		})
	})

	When("migrating an extended ERC20 token (e.g. ERC20Votes)", func() {
		// TODO: make sure that extended tokens are compatible with the ERC20 extensions
		It("should migrate the full token balance to the bank module", func() {
			Expect(true).To(BeFalse(), "not implemented")
		})
	})

	When("running the migration logic for a set of existing ERC20 tokens", func() {
		BeforeEach(func() {
			// TODO: Add some ERC20 tokens and then run migration logic
			// TODO: check here that the balance cannot be queried from the bank keeper before migrating the token
		})

		It("should add and enable the corresponding EVM extensions", func() {
			Expect(true).To(BeFalse(), "not implemented")
		})

		It("should be possible to query the balances through the bank module", func() {
			Expect(true).To(BeFalse(), "not implemented")
		})

		It("should return all tokens when querying all balances for an account", func() {
			Expect(true).To(BeFalse(), "not implemented")
		})
	})

	When("registering a native IBC coin", func() {
		BeforeEach(func() {
			// TODO: Add some IBC coins, register the token pair and then run migration logic
		})

		It("should add the corresponding EVM extensions", func() {
			Expect(true).To(BeFalse(), "not implemented")
		})

		It("should be possible to query the balances using an EVM transaction", func() {
			Expect(true).To(BeFalse(), "not implemented")
		})
	})

	When("using Evmos (not wEvmos) in smart contracts", func() {
		It("should be using straight Evmos for sending funds in smart contracts", func() {
			Expect(true).To(BeFalse(), "not implemented")
		})
	})
})
