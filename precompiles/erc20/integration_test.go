package erc20_test

import (
	"math/big"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/contracts"
	auth "github.com/evmos/evmos/v15/precompiles/authorization"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	"github.com/evmos/evmos/v15/precompiles/erc20/testdata"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/utils"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	erc20types "github.com/evmos/evmos/v15/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

var is *IntegrationTestSuite

type IntegrationTestSuite struct {
	// NOTE: we have to use the Unit testing network because we access a keeper in a setup function.
	// Might adjust this on a follow-up PR.
	network *network.UnitTestNetwork
	handler grpc.Handler
	keyring keyring.Keyring
	factory factory.TxFactory

	bondDenom  string
	tokenDenom string

	precompile *erc20.Precompile
}

func (is *IntegrationTestSuite) SetupTest() {
	keys := keyring.New(2)
	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keys.GetAllAccAddrs()...),
	)
	gh := grpc.NewIntegrationHandler(nw)
	tf := factory.New(nw, gh)

	// Set up min deposit in Evmos
	params, err := gh.GetGovParams("deposit")
	Expect(err).ToNot(HaveOccurred(), "failed to get gov params")
	Expect(params).ToNot(BeNil(), "returned gov params are nil")

	updatedParams := params.Params
	updatedParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(nw.GetDenom(), sdk.NewInt(1e18)))
	err = nw.UpdateGovParams(*updatedParams)
	Expect(err).ToNot(HaveOccurred(), "failed to update the min deposit")

	is.network = nw
	is.factory = tf
	is.handler = gh
	is.keyring = keys

	is.bondDenom = nw.GetDenom()
	is.tokenDenom = "xmpl"

	is.precompile = is.setupERC20Precompile(is.tokenDenom)
}

func TestIntegrationSuite(t *testing.T) {
	is = new(IntegrationTestSuite)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "ERC20 Extension Suite")
}

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
		is.SetupTest()

		sender := is.keyring.GetKey(0)
		contractAddr, err := is.factory.DeployContract(
			sender.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: testdata.ERC20AllowanceCallerContract,
				// NOTE: we're passing the precompile address to the constructor because that initiates the contract
				// to make calls to the correct ERC20 precompile.
				ConstructorArgs: []interface{}{is.precompile.Address()},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

		erc20MinterBurnerAddr, err := is.factory.DeployContract(
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

		ERC20MinterV5Addr, err := is.factory.DeployContract(
			sender.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: testdata.ERC20MinterV5Contract,
				ConstructorArgs: []interface{}{
					"Xmpl", "Xmpl",
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC20 minter contract")

		erc20MinterV5CallerAddr, err := is.factory.DeployContract(
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
					Address: is.precompile.Address(),
					ABI:     is.precompile.ABI,
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
					ABI:     testdata.ERC20MinterV5Contract.ABI,
				},
				erc20V5CallerCall: {
					Address: erc20MinterV5CallerAddr,
					ABI:     testdata.ERC20AllowanceCallerContract.ABI,
				},
			},
		}

		failCheck = testutil.LogCheckArgs{ABIEvents: is.precompile.Events}
		execRevertedCheck = failCheck.WithErrContains("execution reverted")
		passCheck = failCheck.WithExpPass(true)

		err = is.network.NextBlock()
		Expect(err).ToNot(HaveOccurred(), "failed to advance block")
	})

	Context("basic functionality -", func() {
		When("transferring tokens", func() {
			DescribeTable("it should transfer tokens to a non-existing address", func(callType CallType, expGasUsed int64) {
				sender := is.keyring.GetKey(0)
				receiver := utiltx.GenerateAddress()
				fundCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 300)}
				transferCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

				// Fund account with some tokens
				is.fundWithTokens(callType, contractsData, sender.Addr, fundCoins)

				// Transfer tokens
				txArgs, transferArgs := is.getTxAndCallArgs(callType, contractsData, erc20.TransferMethod, receiver, transferCoins[0].Amount.BigInt())

				transferCheck := passCheck.WithExpEvents(erc20.EventTypeTransfer)

				res, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var success bool
				err = is.precompile.UnpackIntoInterface(&success, erc20.TransferMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(success).To(BeTrue(), "expected transfer to succeed")

				is.ExpectBalancesForContract(
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
				Entry(" - through erc20 v5 contract", erc20V5Call, int64(52_113)),
			)

			DescribeTable("it should transfer tokens to an existing address", func(callType CallType) {
				sender := is.keyring.GetKey(0)
				receiver := is.keyring.GetKey(1)
				fundCoinsSender := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 300)}
				fundCoinsReceiver := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 500)}
				transferCoin := sdk.NewInt64Coin(is.tokenDenom, 100)

				// Fund accounts with some tokens
				is.fundWithTokens(callType, contractsData, sender.Addr, fundCoinsSender)
				is.fundWithTokens(callType, contractsData, receiver.Addr, fundCoinsReceiver)

				// Transfer tokens
				txArgs, transferArgs := is.getTxAndCallArgs(callType, contractsData, erc20.TransferMethod, receiver.Addr, transferCoin.Amount.BigInt())

				transferCheck := passCheck.WithExpEvents(erc20.EventTypeTransfer)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, transferCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var success bool
				err = is.precompile.UnpackIntoInterface(&success, erc20.TransferMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(success).To(BeTrue(), "expected transfer to succeed")

				is.ExpectBalancesForContract(
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
				sender := is.keyring.GetKey(0)
				receiver := is.keyring.GetAddr(1)
				fundCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 300)}
				transferCoin := sdk.NewInt64Coin(is.tokenDenom, 100)

				// Fund account with some tokens
				is.fundWithTokens(callType, contractsData, sender.Addr, fundCoins)

				// Transfer tokens
				txArgs, transferArgs := is.getTxAndCallArgs(callType, contractsData, erc20.TransferMethod, receiver, transferCoin.Amount.BigInt())

				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, execRevertedCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
				Expect(ethRes).To(BeNil(), "expected empty result")
			},
				// NOTE: we are not passing the direct call here because this test is specific to the contract calls
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
			)

			DescribeTable("it should return an error if the sender does not have enough tokens", func(callType CallType) {
				sender := is.keyring.GetKey(0)
				receiver := is.keyring.GetAddr(1)
				fundCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 200)}
				transferCoin := sdk.NewInt64Coin(is.tokenDenom, 300)

				// Fund account with some tokens
				is.fundWithTokens(callType, contractsData, sender.Addr, fundCoins)

				// Transfer tokens
				txArgs, transferArgs := is.getTxAndCallArgs(callType, contractsData, erc20.TransferMethod, receiver, transferCoin.Amount.BigInt())

				insufficientBalanceCheck := failCheck.WithErrContains(
					erc20.ErrTransferAmountExceedsBalance.Error(),
				)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, insufficientBalanceCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
				Expect(ethRes).To(BeNil(), "expected empty result")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the contract call here because this test is for direct calls only

				Entry(" - through erc20 contract", erc20Call),
				// // TODO: The ERC20 V5 contract is raising the ERC-6093 standardized error which we are not as of yet
				// Entry(" - through erc20 v5 contract", erc20V5Call),
			)
		})

		When("transferring tokens from another account", func() {
			Context("in a direct call to the token contract", func() {
				DescribeTable("it should transfer tokens from another account with a sufficient approval set", func(callType CallType) {
					owner := is.keyring.GetKey(0)
					spender := is.keyring.GetKey(1)
					receiver := utiltx.GenerateAddress()

					fundCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 300)}
					transferCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					// Fund account with some tokens
					is.fundWithTokens(callType, contractsData, owner.Addr, fundCoins)

					// Set allowance
					is.setupSendAuthzForContract(callType, contractsData, spender.Addr, owner.Priv, transferCoins)

					// Transfer tokens
					txArgs, transferArgs := is.getTxAndCallArgs(
						callType, contractsData,
						erc20.TransferFromMethod,
						owner.Addr, receiver, transferCoins[0].Amount.BigInt(),
					)

					transferCheck := passCheck.WithExpEvents(
						erc20.EventTypeTransfer,
						auth.EventTypeApproval,
					)

					_, ethRes, err := is.factory.CallContractAndCheckLogs(spender.Priv, txArgs, transferArgs, transferCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					var success bool
					err = is.precompile.UnpackIntoInterface(&success, erc20.TransferFromMethod, ethRes.Ret)
					Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
					Expect(success).To(BeTrue(), "expected transferFrom to succeed")

					is.ExpectBalancesForContract(
						callType, contractsData,
						[]ExpectedBalance{
							{address: owner.AccAddr, expCoins: fundCoins.Sub(transferCoins...)},
							{address: receiver.Bytes(), expCoins: transferCoins},
						},
					)

					// Check that the allowance was removed since we authorized only the transferred amount
					is.ExpectNoSendAuthzForContract(
						callType, contractsData,
						spender.Addr, owner.Addr,
					)
				},
					Entry(" - direct call", directCall),
					// NOTE: we are not passing the contract call here because this test is for direct calls only

					Entry(" - through erc20 contract", erc20Call),
					Entry(" - through erc20 v5 contract", erc20V5Call),
				)

				DescribeTable("it should transfer funds from the own account in case sufficient approval is set", func(callType CallType) {
					owner := is.keyring.GetKey(0)
					receiver := utiltx.GenerateAddress()

					fundCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 300)}
					transferCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					// Fund account with some tokens
					is.fundWithTokens(callType, contractsData, owner.Addr, fundCoins)

					// NOTE: Here we set up the allowance using the contract calls instead of the helper utils,
					// because the `MsgGrant` used there doesn't allow the sender to be the same as the spender,
					// but the ERC20 contracts do.
					txArgs, approveArgs := is.getTxAndCallArgs(
						callType, contractsData,
						auth.ApproveMethod,
						owner.Addr, transferCoins[0].Amount.BigInt(),
					)

					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := is.factory.CallContractAndCheckLogs(owner.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					is.ExpectSendAuthzForContract(
						callType, contractsData,
						owner.Addr, owner.Addr, transferCoins,
					)

					// Transfer tokens
					txArgs, transferArgs := is.getTxAndCallArgs(
						callType, contractsData,
						erc20.TransferFromMethod,
						owner.Addr, receiver, transferCoins[0].Amount.BigInt(),
					)

					transferCheck := passCheck.WithExpEvents(
						erc20.EventTypeTransfer,
						auth.EventTypeApproval,
					)

					_, ethRes, err := is.factory.CallContractAndCheckLogs(owner.Priv, txArgs, transferArgs, transferCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					var success bool
					err = is.precompile.UnpackIntoInterface(&success, erc20.TransferFromMethod, ethRes.Ret)
					Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
					Expect(success).To(BeTrue(), "expected transferFrom to succeed")

					is.ExpectBalancesForContract(
						callType, contractsData,
						[]ExpectedBalance{
							{address: owner.AccAddr, expCoins: fundCoins.Sub(transferCoins...)},
							{address: receiver.Bytes(), expCoins: transferCoins},
						},
					)

					// Check that the allowance was removed since we authorized only the transferred amount
					// FIXME: This is not working for the case where we transfer from the own account
					// because the allowance is not removed on the SDK side.
					is.ExpectNoSendAuthzForContract(
						callType, contractsData,
						owner.Addr, owner.Addr,
					)
				},
					Entry(" - direct call", directCall),
					// NOTE: we are not passing the contract call here because this test case only covers direct calls

					Entry(" - through erc20 contract", erc20Call),
					Entry(" - through erc20 v5 contract", erc20V5Call),
				)

				DescribeTable("it should return an error when the spender does not have enough allowance", func(callType CallType) {
					owner := is.keyring.GetKey(0)
					spender := is.keyring.GetKey(1)
					receiver := utiltx.GenerateAddress()
					fundCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 300)}
					authzCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}
					transferCoin := sdk.NewInt64Coin(is.tokenDenom, 200)

					// Fund account with some tokens
					is.fundWithTokens(callType, contractsData, owner.Addr, fundCoins)
					// Set allowance
					is.setupSendAuthzForContract(
						callType, contractsData,
						spender.Addr, owner.Priv, authzCoins,
					)

					// Transfer tokens
					txArgs, transferArgs := is.getTxAndCallArgs(
						callType, contractsData,
						erc20.TransferFromMethod,
						owner.Addr, receiver, transferCoin.Amount.BigInt(),
					)

					insufficientAllowanceCheck := failCheck.WithErrContains(erc20.ErrInsufficientAllowance.Error())

					_, ethRes, err := is.factory.CallContractAndCheckLogs(spender.Priv, txArgs, transferArgs, insufficientAllowanceCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
					Expect(ethRes).To(BeNil(), "expected empty result")
				},
					Entry(" - direct call", directCall),
					// NOTE: we are not passing the contract call here because this test case only covers direct calls

					Entry(" - through erc20 contract", erc20Call),

					// TODO: the ERC20 V5 contract is raising the ERC-6093 standardized error which we are not using as of yet
					// Entry(" - through erc20 v5 contract", erc20V5Call),
				)

				DescribeTable("it should return an error if there is no allowance set", func(callType CallType) {
					sender := is.keyring.GetKey(0)
					from := is.keyring.GetKey(1)
					receiver := utiltx.GenerateAddress()
					fundCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 300)}
					transferCoin := sdk.NewInt64Coin(is.tokenDenom, 100)

					// Fund account with some tokens
					is.fundWithTokens(callType, contractsData, from.Addr, fundCoins)

					// Transfer tokens
					txArgs, transferArgs := is.getTxAndCallArgs(
						callType, contractsData,
						erc20.TransferFromMethod,
						from.Addr, receiver, transferCoin.Amount.BigInt(),
					)

					insufficientAllowanceCheck := failCheck.WithErrContains(
						erc20.ErrInsufficientAllowance.Error(),
					)

					_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, insufficientAllowanceCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
					Expect(ethRes).To(BeNil(), "expected empty result")
				},
					Entry(" - direct call", directCall),
					// NOTE: we are not passing the contract call here because this test case only covers direct calls

					Entry(" - through erc20 contract", erc20Call),

					// TODO: the ERC20 V5 contract is raising the ERC-6093 standardized error which we are not using as of yet
					// Entry(" - through erc20 v5 contract", erc20V5Call),
				)

				DescribeTable("it should return an error if the sender does not have enough tokens", func(callType CallType) {
					sender := is.keyring.GetKey(0)
					from := is.keyring.GetKey(1)
					receiver := utiltx.GenerateAddress()
					fundCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 200)}
					transferCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 300)}

					// Fund account with some tokens
					is.fundWithTokens(callType, contractsData, from.Addr, fundCoins)

					// Set allowance
					is.setupSendAuthzForContract(
						callType, contractsData,
						sender.Addr, from.Priv, transferCoins,
					)

					// Transfer tokens
					txArgs, transferArgs := is.getTxAndCallArgs(callType, contractsData, erc20.TransferFromMethod, from.Addr, receiver, transferCoins[0].Amount.BigInt())

					insufficientBalanceCheck := failCheck.WithErrContains(
						erc20.ErrTransferAmountExceedsBalance.Error(),
					)

					_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, transferArgs, insufficientBalanceCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
					Expect(ethRes).To(BeNil(), "expected empty result")
				},
					Entry(" - direct call", directCall),
					// NOTE: we are not passing the contract call here because this test case only covers direct calls

					Entry(" - through erc20 contract", erc20Call),

					// TODO: the ERC20 V5 contract is raising the ERC-6093 standardized error which we are not using as of yet
					// Entry(" - through erc20 v5 contract", erc20V5Call),
				)
			})

			Context("in a call from another smart contract to the token contract", func() {
				DescribeTable("it should transfer tokens with a sufficient approval set", func(callType CallType) {
					owner := is.keyring.GetKey(0)
					receiver := utiltx.GenerateAddress()
					fundCoin := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 300)}
					transferCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					// NOTE: the spender will be the contract address
					spender := contractsData.GetContractData(callType).Address

					// Fund account with some tokens
					is.fundWithTokens(callType, contractsData, owner.Addr, fundCoin)

					// Set allowance
					is.setupSendAuthzForContract(
						callType, contractsData,
						spender, owner.Priv, transferCoins,
					)

					// Transfer tokens
					txArgs, transferArgs := is.getTxAndCallArgs(
						callType, contractsData,
						erc20.TransferFromMethod,
						owner.Addr, receiver, transferCoins[0].Amount.BigInt(),
					)

					transferCheck := passCheck.WithExpEvents(
						erc20.EventTypeTransfer,
						auth.EventTypeApproval,
					)

					_, ethRes, err := is.factory.CallContractAndCheckLogs(owner.Priv, txArgs, transferArgs, transferCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					var success bool
					err = is.precompile.UnpackIntoInterface(&success, erc20.TransferFromMethod, ethRes.Ret)
					Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
					Expect(success).To(BeTrue(), "expected transferFrom to succeed")

					is.ExpectBalancesForContract(
						callType, contractsData,
						[]ExpectedBalance{
							{address: owner.AccAddr, expCoins: fundCoin.Sub(transferCoins...)},
							{address: receiver.Bytes(), expCoins: transferCoins},
						},
					)

					// Check that the allowance was removed since we authorized only the transferred amount
					is.ExpectNoSendAuthzForContract(
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

				DescribeTable("it should transfer funds with a sufficient allowance and triggered from another account", func(callType CallType) {
					msgSender := is.keyring.GetKey(0)
					owner := is.keyring.GetKey(1)
					receiver := utiltx.GenerateAddress()

					// NOTE: the spender will be the contract address
					spender := contractsData.GetContractData(callType).Address

					fundCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 300)}
					transferCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					// Fund account with some tokens
					is.fundWithTokens(callType, contractsData, owner.Addr, fundCoins)

					// Set allowance
					is.setupSendAuthzForContract(
						callType, contractsData,
						spender, owner.Priv, transferCoins,
					)

					// Transfer tokens
					txArgs, transferArgs := is.getTxAndCallArgs(
						callType, contractsData,
						erc20.TransferFromMethod,
						owner.Addr, receiver, transferCoins[0].Amount.BigInt(),
					)

					transferCheck := passCheck.WithExpEvents(
						erc20.EventTypeTransfer,
						auth.EventTypeApproval,
					)

					_, ethRes, err := is.factory.CallContractAndCheckLogs(msgSender.Priv, txArgs, transferArgs, transferCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					var success bool
					err = is.precompile.UnpackIntoInterface(&success, erc20.TransferFromMethod, ethRes.Ret)
					Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
					Expect(success).To(BeTrue(), "expected transferFrom to succeed")

					is.ExpectBalancesForContract(
						callType, contractsData,
						[]ExpectedBalance{
							{address: owner.AccAddr, expCoins: fundCoins.Sub(transferCoins...)},
							{address: receiver.Bytes(), expCoins: transferCoins},
						},
					)

					// Check that the allowance was removed since we authorized only the transferred amount
					is.ExpectNoSendAuthzForContract(
						callType, contractsData,
						spender, owner.Addr,
					)
				},
					// NOTE: we are not passing the direct call here because this test is specific to the contract calls

					Entry(" - through contract", contractCall),
					Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
				)

				DescribeTable("it should return an error when the spender does not have enough allowance", func(callType CallType) {
					from := is.keyring.GetKey(0)
					receiver := utiltx.GenerateAddress()
					fundCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 400)}
					authzCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}
					transferCoin := sdk.NewInt64Coin(is.tokenDenom, 300)

					// NOTE: the spender will be the contract address
					spender := contractsData.GetContractData(callType).Address

					// Fund account with some tokens
					is.fundWithTokens(callType, contractsData, from.Addr, fundCoins)

					// Set allowance
					is.setupSendAuthzForContract(callType, contractsData, spender, from.Priv, authzCoins)

					// Transfer tokens
					txArgs, transferArgs := is.getTxAndCallArgs(
						callType, contractsData,
						erc20.TransferFromMethod,
						from.Addr, receiver, transferCoin.Amount.BigInt(),
					)

					_, ethRes, err := is.factory.CallContractAndCheckLogs(from.Priv, txArgs, transferArgs, execRevertedCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
					Expect(ethRes).To(BeNil(), "expected empty result")
				},
					// NOTE: we are not passing the direct call here because this test is for contract calls only
					Entry(" - through contract", contractCall),
					Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
				)
			})
		})

		When("querying balance", func() {
			DescribeTable("it should return an existing balance", func(callType CallType) {
				sender := is.keyring.GetKey(0)
				expBalance := big.NewInt(100)
				fundCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, expBalance.Int64())}

				// Fund account with some tokens
				is.fundWithTokens(callType, contractsData, sender.Addr, fundCoins)

				// Query the balance
				txArgs, balancesArgs := is.getTxAndCallArgs(callType, contractsData, erc20.BalanceOfMethod, sender.Addr)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balance *big.Int
				err = is.precompile.UnpackIntoInterface(&balance, erc20.BalanceOfMethod, ethRes.Ret)
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
				sender := is.keyring.GetKey(0)
				address := utiltx.GenerateAddress()
				fundCoins := sdk.Coins{sdk.NewInt64Coin(is.network.GetDenom(), 100)}

				// Fund account with some tokens
				err := is.network.FundAccount(sender.AccAddr, fundCoins)
				Expect(err).ToNot(HaveOccurred(), "failed to fund account")

				// Query the balance
				txArgs, balancesArgs := is.getTxAndCallArgs(callType, contractsData, erc20.BalanceOfMethod, address)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balance *big.Int
				err = is.precompile.UnpackIntoInterface(&balance, erc20.BalanceOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(balance.Int64()).To(BeZero(), "expected zero balance")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contracts
				// only support the actual token denomination and don't know of other balances.
			)

			DescribeTable("it should return zero if the account does not exist", func(callType CallType) {
				sender := is.keyring.GetKey(0)
				address := utiltx.GenerateAddress()

				// Query the balance
				txArgs, balancesArgs := is.getTxAndCallArgs(callType, contractsData, erc20.BalanceOfMethod, address)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balance *big.Int
				err = is.precompile.UnpackIntoInterface(&balance, erc20.BalanceOfMethod, ethRes.Ret)
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
				granter := is.keyring.GetKey(0)
				authzCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

				is.setupSendAuthzForContract(callType, contractsData, grantee, granter.Priv, authzCoins)

				txArgs, allowanceArgs := is.getTxAndCallArgs(callType, contractsData, auth.AllowanceMethod, granter.Addr, grantee)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var allowance *big.Int
				err = is.precompile.UnpackIntoInterface(&allowance, auth.AllowanceMethod, ethRes.Ret)
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
				grantee := is.keyring.GetAddr(1)
				granter := is.keyring.GetKey(0)

				txArgs, allowanceArgs := is.getTxAndCallArgs(callType, contractsData, auth.AllowanceMethod, granter.Addr, grantee)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var allowance *big.Int
				err = is.precompile.UnpackIntoInterface(&allowance, auth.AllowanceMethod, ethRes.Ret)
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
				grantee := is.keyring.GetKey(1)
				granter := is.keyring.GetKey(0)
				authzCoins := sdk.Coins{sdk.NewInt64Coin(is.network.GetDenom(), 100)}

				is.setupSendAuthz(grantee.AccAddr, granter.Priv, authzCoins)

				txArgs, allowanceArgs := is.getTxAndCallArgs(callType, contractsData, auth.AllowanceMethod, granter.Addr, grantee.Addr)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var allowance *big.Int
				err = is.precompile.UnpackIntoInterface(&allowance, auth.AllowanceMethod, ethRes.Ret)
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
				granter := is.keyring.GetKey(0)

				txArgs, allowanceArgs := is.getTxAndCallArgs(callType, contractsData, auth.AllowanceMethod, granter.Addr, grantee)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, allowanceArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var allowance *big.Int
				err = is.precompile.UnpackIntoInterface(&allowance, auth.AllowanceMethod, ethRes.Ret)
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
				sender := is.keyring.GetKey(0)
				expSupply := big.NewInt(100)
				fundCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, expSupply.Int64())}

				// Fund account with some tokens
				is.fundWithTokens(callType, contractsData, sender.Addr, fundCoins)

				// Query the balance
				txArgs, supplyArgs := is.getTxAndCallArgs(callType, contractsData, erc20.TotalSupplyMethod)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var supply *big.Int
				err = is.precompile.UnpackIntoInterface(&supply, erc20.TotalSupplyMethod, ethRes.Ret)
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
				sender := is.keyring.GetKey(0)
				txArgs, supplyArgs := is.getTxAndCallArgs(callType, contractsData, erc20.TotalSupplyMethod)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var supply *big.Int
				err = is.precompile.UnpackIntoInterface(&supply, erc20.TotalSupplyMethod, ethRes.Ret)
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
	})

	Context("metadata query -", func() {
		Context("for a token without registered metadata", func() {
			BeforeEach(func() {
				// Deploy ERC20NoMetadata contract for this test
				erc20NoMetadataAddr, err := is.factory.DeployContract(
					is.keyring.GetPrivKey(0),
					evmtypes.EvmTxArgs{},
					factory.ContractDeploymentData{
						Contract: testdata.ERC20NoMetadataContract,
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
				txArgs, nameArgs := is.getTxAndCallArgs(callType, contractsData, erc20.NameMethod)

				_, _, err := is.factory.CallContractAndCheckLogs(is.keyring.GetPrivKey(0), txArgs, nameArgs, execRevertedCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 contract", erc20Call), // NOTE: we're passing the ERC20 contract call here which was adjusted to point to a contract without metadata to expect the same errors
			)

			DescribeTable("querying the symbol should return an error", func(callType CallType) {
				txArgs, symbolArgs := is.getTxAndCallArgs(callType, contractsData, erc20.SymbolMethod)

				_, _, err := is.factory.CallContractAndCheckLogs(is.keyring.GetPrivKey(0), txArgs, symbolArgs, execRevertedCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 contract", erc20Call), // NOTE: we're passing the ERC20 contract call here which was adjusted to point to a contract without metadata to expect the same errors
			)

			DescribeTable("querying the decimals should return an error", func(callType CallType) {
				txArgs, decimalsArgs := is.getTxAndCallArgs(callType, contractsData, erc20.DecimalsMethod)

				_, _, err := is.factory.CallContractAndCheckLogs(is.keyring.GetPrivKey(0), txArgs, decimalsArgs, execRevertedCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 contract", erc20Call), // NOTE: we're passing the ERC20 contract call here which was adjusted to point to a contract without metadata to expect the same errors
			)
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
				tokenPair, err := utils.RegisterERC20(is.factory, is.network, utils.ERC20RegistrationData{
					Address:      erc20Addr,
					Denom:        denom,
					ProposerPriv: is.keyring.GetPrivKey(0),
				})
				Expect(err).ToNot(HaveOccurred(), "failed to register ERC20 token")

				// overwrite the other precompile with this one, so that the test utils like is.getTxAndCallArgs still work.
				is.precompile, err = setupERC20PrecompileForTokenPair(*is.network, tokenPair)
				Expect(err).ToNot(HaveOccurred(), "failed to set up erc20 precompile")

				// update this in the global contractsData
				contractsData.contractData[directCall] = ContractData{
					Address: is.precompile.Address(),
					ABI:     is.precompile.ABI,
				}

				// Deploy contract calling the ERC20 precompile
				callerAddr, err := is.factory.DeployContract(
					is.keyring.GetPrivKey(0),
					evmtypes.EvmTxArgs{},
					factory.ContractDeploymentData{
						Contract: testdata.ERC20AllowanceCallerContract,
						ConstructorArgs: []interface{}{
							is.precompile.Address(),
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
				txArgs, nameArgs := is.getTxAndCallArgs(callType, contractsData, erc20.NameMethod)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(is.keyring.GetPrivKey(0), txArgs, nameArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var name string
				err = is.precompile.UnpackIntoInterface(&name, erc20.NameMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(name).To(Equal(expName), "expected different name")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 v5 contract", erc20V5Call),
			)

			DescribeTable("querying the symbol should return the symbol", func(callType CallType) {
				txArgs, symbolArgs := is.getTxAndCallArgs(callType, contractsData, erc20.SymbolMethod)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(is.keyring.GetPrivKey(0), txArgs, symbolArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var symbol string
				err = is.precompile.UnpackIntoInterface(&symbol, erc20.SymbolMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(symbol).To(Equal(expSymbol), "expected different symbol")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 v5 contract", erc20V5Call),
			)

			DescribeTable("querying the decimals should return the decimals", func(callType CallType) {
				txArgs, decimalsArgs := is.getTxAndCallArgs(callType, contractsData, erc20.DecimalsMethod)

				_, ethRes, err := is.factory.CallContractAndCheckLogs(is.keyring.GetPrivKey(0), txArgs, decimalsArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var decimals uint8
				err = is.precompile.UnpackIntoInterface(&decimals, erc20.DecimalsMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
				Expect(decimals).To(Equal(expDecimals), "expected different decimals")
			},
				Entry(" - direct call", directCall),
				Entry(" - through contract", contractCall),
				Entry(" - through erc20 v5 contract", erc20V5Call),
			)
		})
	})

	Context("allowance adjustments -", func() {})
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
			is.SetupTest()

			contractOwner := is.keyring.GetKey(0)

			// Deploy an ERC20 contract
			erc20Addr, err := is.factory.DeployContract(
				contractOwner.Priv,
				evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
				factory.ContractDeploymentData{
					Contract: testdata.ERC20MinterV5Contract,
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
						ABI:     testdata.ERC20MinterV5Contract.ABI,
					},
				},
			}

			err = is.network.NextBlock()
			Expect(err).ToNot(HaveOccurred(), "failed to commit block")

			// Register the deployed erc20 contract as a token pair
			_, err = utils.RegisterERC20(is.factory, is.network, utils.ERC20RegistrationData{
				Address:      erc20Addr,
				Denom:        tokenDenom,
				ProposerPriv: contractOwner.Priv,
			})
			Expect(err).ToNot(HaveOccurred(), "failed to register ERC20 token")

			err = is.network.NextBlock()
			Expect(err).ToNot(HaveOccurred(), "failed to commit block")

			// Mint the supply of tokens
			err = is.MintERC20(erc20V5Call, contractData, contractOwner.Addr, supply.Amount.BigInt())
			Expect(err).ToNot(HaveOccurred(), "failed to mint tokens")

			err = is.network.NextBlock()
			Expect(err).ToNot(HaveOccurred(), "failed to commit block")

			// Check that the supply was minted
			is.ExpectBalancesForERC20(erc20V5Call, contractData, []ExpectedBalance{{
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
