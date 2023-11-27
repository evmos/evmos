package erc20_test

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
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

		// FIXME: remove once tests are added
		_ = contractsData
		_ = failCheck
		_ = execRevertedCheck
		_ = passCheck
	})

	Context("basic functionality -", func() {
		When("approving an allowance", func() {
			Context("in a direct call", func() {
				DescribeTable("it should approve an allowance", func(callType CallType) {
					grantee := is.keyring.GetKey(0)
					granter := is.keyring.GetKey(1)
					transferCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 200)}

					// Approve allowance
					txArgs, approveArgs := is.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, transferCoins[0].Amount.BigInt())

					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Check allowance
					is.ExpectSendAuthzForContract(
						callType, contractsData,
						grantee.Addr, granter.Addr, transferCoins,
					)
				},
					Entry(" - direct call", directCall),
					Entry(" - through erc20 contract", erc20Call),
					Entry(" - through erc20 v5 contract", erc20V5Call),
				)

				DescribeTable("it should add a new spend limit to an existing allowance with a different token", func(callType CallType) {
					grantee := is.keyring.GetKey(1)
					granter := is.keyring.GetKey(0)
					bondCoins := sdk.Coins{sdk.NewInt64Coin(is.network.GetDenom(), 200)}
					tokenCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					// set up a previous authorization
					is.setupSendAuthz(grantee.AccAddr, granter.Priv, bondCoins)

					// Approve allowance
					txArgs, approveArgs := is.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, tokenCoins[0].Amount.BigInt())

					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Check allowance contains both spend limits
					is.expectSendAuthz(grantee.AccAddr, granter.AccAddr, bondCoins.Add(tokenCoins...))
				},
					Entry(" - direct call", directCall),

					// NOTE 2: we are not passing the erc20 contract call here because the ERC20 contract
					// only supports the actual token denomination and doesn't know of other allowances.
				)

				DescribeTable("it should set the new spend limit for an existing allowance with the same token", func(callType CallType) {
					grantee := is.keyring.GetKey(1)
					granter := is.keyring.GetKey(0)
					bondCoins := sdk.Coins{sdk.NewInt64Coin(is.network.GetDenom(), 200)}
					tokenCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}
					doubleTokenCoin := sdk.NewInt64Coin(is.tokenDenom, 200)

					// set up a previous authorization
					is.setupSendAuthz(grantee.AccAddr, granter.Priv, bondCoins.Add(doubleTokenCoin))

					// Approve allowance
					txArgs, approveArgs := is.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, tokenCoins[0].Amount.BigInt())

					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Check allowance contains both spend limits
					is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, bondCoins.Add(tokenCoins...))
				},
					Entry(" - direct call", directCall),
					Entry(" - through erc20 contract", erc20Call),
					Entry(" - through erc20 v5 contract", erc20V5Call),
				)

				DescribeTable("it should remove the token from the spend limit of an existing authorization when approving zero", func(callType CallType) {
					grantee := is.keyring.GetKey(1)
					granter := is.keyring.GetKey(0)
					bondCoins := sdk.Coins{sdk.NewInt64Coin(is.network.GetDenom(), 200)}
					tokenCoin := sdk.NewInt64Coin(is.tokenDenom, 100)

					// set up a previous authorization
					is.setupSendAuthz(grantee.AccAddr, granter.Priv, bondCoins.Add(tokenCoin))

					// Approve allowance
					txArgs, approveArgs := is.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, common.Big0)

					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Check allowance contains only the spend limit in network denomination
					is.expectSendAuthz(grantee.AccAddr, granter.AccAddr, bondCoins)
				},
					Entry(" - direct call", directCall),
					// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
					// only supports the actual token denomination and doesn't know of other allowances.
				)

				DescribeTable("it should delete the authorization when approving zero with no other spend limits", func(callType CallType) {
					grantee := is.keyring.GetKey(1)
					granter := is.keyring.GetKey(0)
					tokenCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					// set up a previous authorization
					is.setupSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Priv, tokenCoins)

					// Approve allowance
					txArgs, approveArgs := is.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, common.Big0)

					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Check allowance was deleted
					is.expectNoSendAuthz(grantee.AccAddr, granter.AccAddr)
				},
					Entry(" - direct call", directCall),
					Entry(" - through erc20 contract", erc20Call),
					Entry(" - through erc20 v5 contract", erc20V5Call),
				)

				DescribeTable("it should no-op if approving 0 and no allowance exists", func(callType CallType) {
					grantee := is.keyring.GetKey(1)
					granter := is.keyring.GetKey(0)

					// Approve allowance
					txArgs, approveArgs := is.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, common.Big0)

					// We are expecting an approval to be made, but no authorization stored since it's 0
					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Check still no authorization exists
					is.ExpectNoSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr)
				},
					Entry(" - direct call", directCall),
					Entry(" - through erc20 contract", erc20Call),
					Entry(" - through erc20 v5 contract", erc20V5Call),
				)

				DescribeTable("it should create an allowance if the grantee is the same as the granter", func(callType CallType) {
					grantee := is.keyring.GetKey(0)
					granter := is.keyring.GetKey(0)
					authzCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					// Approve allowance
					txArgs, approveArgs := is.getTxAndCallArgs(
						callType, contractsData,
						auth.ApproveMethod,
						grantee.Addr, authzCoins[0].Amount.BigInt(),
					)

					approvalCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approvalCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					is.ExpectSendAuthzForContract(
						callType, contractsData,
						grantee.Addr, granter.Addr, authzCoins,
					)
				},
					Entry(" - direct call", directCall),
					Entry(" - through erc20 contract", erc20Call),
					Entry(" - through erc20 v5 contract", erc20V5Call),
				)

				DescribeTable("it should return an error if approving 0 and allowance only exists for other tokens", func(callType CallType) {
					grantee := is.keyring.GetKey(1)
					granter := is.keyring.GetKey(0)
					bondCoins := sdk.Coins{sdk.NewInt64Coin(is.network.GetDenom(), 200)}

					// set up a previous authorization
					is.setupSendAuthz(grantee.AccAddr, granter.Priv, bondCoins)

					// Approve allowance
					txArgs, approveArgs := is.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, common.Big0)

					notFoundCheck := failCheck.WithErrContains(
						fmt.Sprintf(erc20.ErrNoAllowanceForToken, is.tokenDenom),
					)

					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, notFoundCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
				},
					Entry(" - direct call", directCall),
					// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
					// only supports the actual token denomination and doesn't know of other allowances.
				)
			})

			// NOTE: We have to split the tests for contract calls into a separate context because
			// when approving through a smart contract, the approval is created between the contract address and the
			// grantee, instead of the sender address and the grantee.
			Context("in a contract call", func() {
				DescribeTable("it should approve an allowance", func(callType CallType) {
					sender := is.keyring.GetKey(0)
					grantee := is.keyring.GetKey(1)
					granter := contractsData.GetContractData(callType).Address // the granter will be the contract address
					transferCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 200)}

					// Approve allowance
					txArgs, approveArgs := is.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, transferCoins[0].Amount.BigInt())

					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Check allowance
					is.ExpectSendAuthzForContract(
						callType, contractsData,
						grantee.Addr, granter, transferCoins,
					)
				},
					Entry(" - through contract", contractCall),
					Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
				)

				DescribeTable("it should set the new spend limit for an existing allowance with the same token", func(callType CallType) {
					sender := is.keyring.GetKey(0)
					grantee := is.keyring.GetKey(1)
					granter := contractsData.GetContractData(callType).Address // the granter will be the contract address
					initialAmount := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}
					newAmount := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 200)}

					// Set up a first approval
					txArgs, approveArgs := is.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, initialAmount[0].Amount.BigInt())
					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)
					_, _, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Set up a second approval which should overwrite the initial one
					txArgs, approveArgs = is.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, newAmount[0].Amount.BigInt())
					approveCheck = passCheck.WithExpEvents(auth.EventTypeApproval)
					_, _, err = is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Check allowance has been updated
					is.ExpectSendAuthzForContract(
						callType, contractsData,
						grantee.Addr, granter, newAmount,
					)
				},
					Entry(" - through contract", contractCall),
					Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
				)

				DescribeTable("it should delete the authorization when approving zero with no other spend limits", func(callType CallType) {
					sender := is.keyring.GetKey(0)
					grantee := is.keyring.GetKey(1)
					granter := contractsData.GetContractData(callType).Address // the granter will be the contract address
					tokenCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					// set up a previous authorization
					txArgs, approveArgs := is.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, tokenCoins[0].Amount.BigInt())
					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)
					_, _, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Approve allowance
					txArgs, approveArgs = is.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, common.Big0)
					_, _, err = is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Check allowance was deleted from the keeper / is returning 0 for smart contracts
					is.ExpectNoSendAuthzForContract(callType, contractsData, grantee.Addr, granter)
				},
					Entry(" - through contract", contractCall),
					Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
				)

				DescribeTable("it should no-op if approving 0 and no allowance exists", func(callType CallType) {
					sender := is.keyring.GetKey(0)
					grantee := is.keyring.GetKey(1)
					granter := contractsData.GetContractData(callType).Address // the granter will be the contract address

					// Approve allowance
					txArgs, approveArgs := is.getTxAndCallArgs(callType, contractsData, auth.ApproveMethod, grantee.Addr, common.Big0)

					// We are expecting an approval event to be emitted, but no authorization to be stored
					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Check still no authorization exists
					is.ExpectNoSendAuthzForContract(callType, contractsData, grantee.Addr, granter)
				},
					Entry(" - through contract", contractCall),
					Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
				)

				DescribeTable("it should create an allowance if the grantee is the same as the granter", func(callType CallType) {
					sender := is.keyring.GetKey(0)
					granter := contractsData.GetContractData(callType).Address // the granter will be the contract address
					grantee := granter
					authzCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					// Approve allowance
					txArgs, approveArgs := is.getTxAndCallArgs(
						callType, contractsData,
						auth.ApproveMethod,
						grantee, authzCoins[0].Amount.BigInt(),
					)

					approvalCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := is.factory.CallContractAndCheckLogs(sender.Priv, txArgs, approveArgs, approvalCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					is.ExpectSendAuthzForContract(
						callType, contractsData,
						grantee, granter, authzCoins,
					)
				},
					Entry(" - through contract", contractCall),
					Entry(" - through erc20 v5 caller contract", erc20V5CallerCall),
				)
			})
		})
	})

	Context("metadata query -", func() {})

	Context("allowance adjustments -", func() {
		var (
			grantee keyring.Key
			granter keyring.Key
		)

		BeforeEach(func() {
			// Deploying the contract which has the increase / decrease allowance methods
			contractAddr, err := is.factory.DeployContract(
				is.keyring.GetPrivKey(0),
				evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
				factory.ContractDeploymentData{
					Contract:        testdata.ERC20AllowanceCallerContract,
					ConstructorArgs: []interface{}{is.precompile.Address()},
				},
			)
			Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

			contractsData.contractData[erc20CallerCall] = ContractData{
				Address: contractAddr,
				ABI:     testdata.ERC20AllowanceCallerContract.ABI,
			}

			grantee = is.keyring.GetKey(0)
			granter = is.keyring.GetKey(1)
		})

		When("no allowance exists", func() {
			DescribeTable("decreasing the allowance should return an error", func(callType CallType) {
				authzCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

				txArgs, decreaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, authzCoins[0].Amount.BigInt())

				notFoundCheck := execRevertedCheck
				if callType == directCall {
					notFoundCheck = failCheck.WithErrContains(
						fmt.Sprintf(auth.ErrAuthzDoesNotExistOrExpired, erc20.SendMsgURL, grantee.Addr.String()),
					)
				}

				_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, notFoundCheck)
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
					authzCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					txArgs, increaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.IncreaseAllowanceMethod, grantee.Addr, authzCoins[0].Amount.BigInt())

					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, authzCoins)
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
					authzCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					txArgs, increaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.IncreaseAllowanceMethod, grantee.Addr, authzCoins[0].Amount.BigInt())

					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, contractAddr, authzCoins)
				},
					Entry(" - contract call", contractCall),
					Entry(" - through erc20 caller contract", erc20CallerCall),
				)
			})
		})

		When("an allowance exists for other tokens", func() {
			var bondCoins sdk.Coins

			BeforeEach(func() {
				bondCoins = sdk.Coins{sdk.NewInt64Coin(is.network.GetDenom(), 200)}
				is.setupSendAuthz(grantee.AccAddr, granter.Priv, bondCoins)
			})

			DescribeTable("increasing the allowance should add the token to the spend limit", func(callType CallType) {
				increaseCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

				txArgs, increaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.IncreaseAllowanceMethod, grantee.Addr, increaseCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, bondCoins.Add(increaseCoins...))
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
			)

			DescribeTable("decreasing the allowance should return an error", func(callType CallType) {
				decreaseCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

				txArgs, decreaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, decreaseCoins[0].Amount.BigInt())

				notFoundCheck := execRevertedCheck
				if callType == directCall {
					notFoundCheck = failCheck.WithErrContains(
						fmt.Sprintf(erc20.ErrNoAllowanceForToken, is.tokenDenom),
					)
				}

				_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, notFoundCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
			)
		})

		When("an allowance exists for the same token", func() {
			var authzCoins sdk.Coins

			BeforeEach(func() {
				authzCoins = sdk.NewCoins(
					sdk.NewInt64Coin(is.network.GetDenom(), 100),
					sdk.NewInt64Coin(is.tokenDenom, 200),
				)

				is.setupSendAuthz(grantee.AccAddr, granter.Priv, authzCoins)
			})

			DescribeTable("increasing the allowance should increase the spend limit", func(callType CallType) {
				increaseCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

				txArgs, increaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.IncreaseAllowanceMethod, grantee.Addr, increaseCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, authzCoins.Add(increaseCoins...))
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
			)

			DescribeTable("decreasing the allowance should decrease the spend limit", func(callType CallType) {
				decreaseCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

				txArgs, decreaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, decreaseCoins[0].Amount.BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, authzCoins.Sub(decreaseCoins...))
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
			)

			DescribeTable("increasing the allowance beyond the max uint256 value should return an error", func(callType CallType) {
				maxUint256Coins := sdk.Coins{sdk.NewCoin(is.tokenDenom, sdk.NewIntFromBigInt(abi.MaxUint256))}

				txArgs, increaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.IncreaseAllowanceMethod, grantee.Addr, maxUint256Coins[0].Amount.BigInt())
				_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, execRevertedCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
			)

			DescribeTable("decreasing the allowance to zero should remove the token from the spend limit", func(callType CallType) {
				txArgs, decreaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, authzCoins.AmountOf(is.tokenDenom).BigInt())

				approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)

				_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, approveCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check that only the spend limit in the network denomination remains
				bondDenom := is.network.GetDenom()
				expCoins := sdk.Coins{sdk.NewCoin(bondDenom, authzCoins.AmountOf(bondDenom))}
				is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, expCoins)
			},
				Entry(" - direct call", directCall),
				// NOTE: we are not passing the erc20 contract call here because the ERC20 contract
				// only supports the actual token denomination and doesn't know of other allowances.
			)

			DescribeTable("decreasing the allowance below zero should return an error", func(callType CallType) {
				decreaseCoins := sdk.Coins{sdk.NewCoin(is.tokenDenom, authzCoins.AmountOf(is.tokenDenom).AddRaw(100))}

				txArgs, decreaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, decreaseCoins[0].Amount.BigInt())
				belowZeroCheck := failCheck.WithErrContains(erc20.ErrDecreasedAllowanceBelowZero.Error())
				_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, belowZeroCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Check that the allowance was not changed
				is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, authzCoins)
			},
				Entry(" - direct call", directCall),
			)
		})

		When("an allowance exists for only the same token", func() {
			// NOTE: we have to split between direct and contract calls here because the ERC20 contract
			// handles the allowance differently by creating an approval between the contract and the grantee, instead
			// of the message sender and the grantee, so we expect different authorizations.
			Context("in direct calls", func() {
				var authzCoins sdk.Coins

				BeforeEach(func() {
					authzCoins = sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					// NOTE: We set up the standard authorization here for the authz keeper and then also
					// set up the authorization for the ERC20 contract, so that we can test both.
					is.setupSendAuthzForContract(directCall, contractsData, grantee.Addr, granter.Priv, authzCoins)
					is.setupSendAuthzForContract(erc20Call, contractsData, grantee.Addr, granter.Priv, authzCoins)
				})

				DescribeTable("increasing the allowance should increase the spend limit", func(callType CallType) {
					increaseCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					txArgs, increaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.IncreaseAllowanceMethod, grantee.Addr, increaseCoins[0].Amount.BigInt())
					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)
					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, authzCoins.Add(increaseCoins...))
				},
					Entry(" - direct call", directCall),
					Entry(" - through erc20 contract", erc20Call),
					// NOTE: The ERC20 V5 contract does not contain these methods
					// Entry(" - through erc20 v5 contract", erc20V5Call),
				)

				DescribeTable("decreasing the allowance should decrease the spend limit", func(callType CallType) {
					decreaseCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 50)}

					txArgs, decreaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, decreaseCoins[0].Amount.BigInt())
					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)
					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, authzCoins.Sub(decreaseCoins...))
				},
					Entry(" - direct call", directCall),
					Entry(" - through erc20 contract", erc20Call),
					// NOTE: The ERC20 V5 contract does not contain these methods
					// Entry(" - through erc20 v5 contract", erc20V5Call),
				)

				DescribeTable("decreasing the allowance to zero should delete the authorization", func(callType CallType) {
					txArgs, decreaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, authzCoins.AmountOf(is.tokenDenom).BigInt())
					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)
					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					is.ExpectNoSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr)
				},
					Entry(" - direct call", directCall),
					Entry(" - through erc20 contract", erc20Call),
					// NOTE: The ERC20 V5 contract does not contain these methods
					// Entry(" - through erc20 v5 contract", erc20V5Call),
				)

				DescribeTable("decreasing the allowance below zero should return an error", func(callType CallType) {
					decreaseAmount := authzCoins.AmountOf(is.tokenDenom).AddRaw(100)

					txArgs, decreaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, decreaseAmount.BigInt())

					belowZeroCheck := failCheck.WithErrContains(erc20.ErrDecreasedAllowanceBelowZero.Error())
					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, decreaseArgs, belowZeroCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Check that the allowance was not changed
					is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, authzCoins)
				},
					Entry(" - direct call", directCall),
					Entry(" - through erc20 contract", erc20Call),
					// NOTE: The ERC20 V5 contract does not contain these methods
					// Entry(" - through erc20 v5 contract", erc20V5Call),
				)

				DescribeTable("increasing the allowance beyond the max uint256 value should return an error", func(callType CallType) {
					maxUint256Coins := sdk.Coins{sdk.NewCoin(is.tokenDenom, sdk.NewIntFromBigInt(abi.MaxUint256))}

					txArgs, increaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.IncreaseAllowanceMethod, grantee.Addr, maxUint256Coins[0].Amount.BigInt())
					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, increaseArgs, execRevertedCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Check that the allowance was not changed
					is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granter.Addr, authzCoins)
				},
					Entry(" - direct call", directCall),
					Entry(" - through erc20 contract", erc20Call),
					// NOTE: The ERC20 V5 contract does not contain these methods
					// Entry(" - through erc20 v5 contract", erc20V5Call),
				)
			})

			Context("in contract calls", func() {
				var (
					authzCoins sdk.Coins
					grantee    keyring.Key
				)

				BeforeEach(func() {
					authzCoins = sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					grantee = is.keyring.GetKey(1)
					callerContractAddr := contractsData.GetContractData(contractCall).Address
					erc20CallerContractAddr := contractsData.GetContractData(erc20CallerCall).Address

					// NOTE: Here we create an authorization between the contract and the grantee for both contracts.
					// This is different from the direct calls, where the authorization is created between the
					// message sender and the grantee.
					txArgs, approveArgs := is.getTxAndCallArgs(contractCall, contractsData, auth.ApproveMethod, grantee.Addr, authzCoins[0].Amount.BigInt())
					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)
					_, _, err := is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
					is.ExpectSendAuthzForContract(contractCall, contractsData, grantee.Addr, callerContractAddr, authzCoins)

					// Create the authorization for the ERC20 caller contract
					txArgs, approveArgs = is.getTxAndCallArgs(erc20CallerCall, contractsData, auth.ApproveMethod, grantee.Addr, authzCoins[0].Amount.BigInt())
					_, _, err = is.factory.CallContractAndCheckLogs(granter.Priv, txArgs, approveArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
					is.ExpectSendAuthzForContract(erc20CallerCall, contractsData, grantee.Addr, erc20CallerContractAddr, authzCoins)
				})

				DescribeTable("increasing the allowance should increase the spend limit", func(callType CallType) {
					senderPriv := is.keyring.GetPrivKey(0)
					granterAddr := contractsData.GetContractData(callType).Address
					increaseCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 100)}

					txArgs, increaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.IncreaseAllowanceMethod, grantee.Addr, increaseCoins[0].Amount.BigInt())
					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)
					_, _, err := is.factory.CallContractAndCheckLogs(senderPriv, txArgs, increaseArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granterAddr, authzCoins.Add(increaseCoins...))
				},
					Entry(" - contract call", contractCall),
					Entry(" - through erc20 caller contract", erc20CallerCall),
				)

				DescribeTable("increasing the allowance beyond the max uint256 value should return an error", func(callType CallType) {
					senderPriv := is.keyring.GetPrivKey(0)
					granterAddr := contractsData.GetContractData(callType).Address
					maxUint256Coins := sdk.Coins{sdk.NewCoin(is.tokenDenom, sdk.NewIntFromBigInt(abi.MaxUint256))}

					txArgs, increaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.IncreaseAllowanceMethod, grantee.Addr, maxUint256Coins[0].Amount.BigInt())
					_, _, err := is.factory.CallContractAndCheckLogs(senderPriv, txArgs, increaseArgs, execRevertedCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Check that the allowance was not changed
					is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granterAddr, authzCoins)
				},
					Entry(" - contract call", contractCall),
					Entry(" - through erc20 caller contract", erc20CallerCall),
				)

				DescribeTable("decreasing the allowance should decrease the spend limit", func(callType CallType) {
					senderPriv := is.keyring.GetPrivKey(0)
					granterAddr := contractsData.GetContractData(callType).Address
					decreaseCoins := sdk.Coins{sdk.NewInt64Coin(is.tokenDenom, 50)}

					txArgs, decreaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, decreaseCoins[0].Amount.BigInt())
					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)
					_, _, err := is.factory.CallContractAndCheckLogs(senderPriv, txArgs, decreaseArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granterAddr, authzCoins.Sub(decreaseCoins...))
				},
					Entry(" - contract call", contractCall),
					Entry(" - through erc20 caller contract", erc20CallerCall),
				)

				DescribeTable("decreasing the allowance to zero should delete the authorization", func(callType CallType) {
					senderPriv := is.keyring.GetPrivKey(0)
					granterAddr := contractsData.GetContractData(callType).Address

					txArgs, decreaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, authzCoins.AmountOf(is.tokenDenom).BigInt())
					approveCheck := passCheck.WithExpEvents(auth.EventTypeApproval)
					_, _, err := is.factory.CallContractAndCheckLogs(senderPriv, txArgs, decreaseArgs, approveCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					is.ExpectNoSendAuthzForContract(callType, contractsData, grantee.Addr, granterAddr)
				},
					Entry(" - contract call", contractCall),
					Entry(" - through erc20 caller contract", erc20CallerCall),
				)

				DescribeTable("decreasing the allowance below zero should return an error", func(callType CallType) {
					senderPriv := is.keyring.GetPrivKey(0)
					granterAddr := contractsData.GetContractData(callType).Address
					decreaseCoins := sdk.Coins{sdk.NewCoin(is.tokenDenom, authzCoins.AmountOf(is.tokenDenom).AddRaw(100))}

					txArgs, decreaseArgs := is.getTxAndCallArgs(callType, contractsData, auth.DecreaseAllowanceMethod, grantee.Addr, decreaseCoins[0].Amount.BigInt())
					_, _, err := is.factory.CallContractAndCheckLogs(senderPriv, txArgs, decreaseArgs, execRevertedCheck)
					Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

					// Check that the allowance was not changed
					is.ExpectSendAuthzForContract(callType, contractsData, grantee.Addr, granterAddr, authzCoins)
				},
					Entry(" - contract call", contractCall),
					Entry(" - through erc20 caller contract", erc20CallerCall),
				)
			})
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
