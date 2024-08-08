// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package statedb_test

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/contracts"
	stakingprecompile "github.com/evmos/evmos/v19/precompiles/staking"
	testcontracts "github.com/evmos/evmos/v19/precompiles/testutil/contracts"
	testfactory "github.com/evmos/evmos/v19/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"

	//nolint:revive // okay to use dot imports for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // okay to use dot imports for Ginkgo
	. "github.com/onsi/gomega"
)

func TestNestedEVMExtensionCall(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nested EVM Extension Call Test Suite")
}

type testCase struct {
	method                  string
	expDelegation           bool
	expSenderERC20Balance   *big.Int
	expContractERC20Balance *big.Int
}

// This test is a demonstration of the flash loan exploit that was reported.
// This happens when interacting with EVM extensions in smart contract methods,
// where a resulting state change has the same value as the original state value.
//
// Before the fix, this would result in state changes not being persisted after the EVM extension call,
// therefore leaving the loaned funds in the contract.
var _ = Describe("testing the flash loan exploit", Ordered, func() {
	var (
		keyring testkeyring.Keyring
		// NOTE: we need to use the unit test network here because we need it to instantiate the staking precompile correctly
		network *testnetwork.UnitTestNetwork
		handler grpc.Handler
		factory testfactory.TxFactory

		deployer testkeyring.Key

		erc20Addr         common.Address
		flashLoanAddr     common.Address
		flashLoanContract evmtypes.CompiledContract

		validatorToDelegateTo string

		delegatedAmountPre math.Int

		stakingPrecompile *stakingprecompile.Precompile
	)

	mintAmount := big.NewInt(2e18)
	delegateAmount := big.NewInt(1e18)

	BeforeAll(func() {
		keyring = testkeyring.New(2)
		network = testnetwork.NewUnitTestNetwork(
			testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		)
		handler = grpc.NewIntegrationHandler(network)
		factory = testfactory.New(network, handler)

		deployer = keyring.GetKey(0)

		var err error
		stakingPrecompile, err = stakingprecompile.NewPrecompile(network.App.StakingKeeper, network.App.AuthzKeeper)
		Expect(err).ToNot(HaveOccurred(), "failed to create staking precompile")

		// Load the flash loan contract from the compiled JSON data.
		flashLoanContract, err = testcontracts.LoadFlashLoanContract()
		Expect(err).ToNot(HaveOccurred(), "failed to load flash loan contract")
	})

	BeforeEach(func() {
		valsRes, err := handler.GetBondedValidators()
		Expect(err).ToNot(HaveOccurred(), "failed to get bonded validators")

		validatorToDelegateTo = valsRes.Validators[0].OperatorAddress
		qRes, err := handler.GetDelegation(deployer.AccAddr.String(), validatorToDelegateTo)
		Expect(err).ToNot(HaveOccurred(), "failed to get delegation")
		delegatedAmountPre = qRes.DelegationResponse.Balance.Amount
		Expect(network.NextBlock()).To(BeNil())

		// Deploy an ERC-20 token contract.
		erc20Addr, err = factory.DeployContract(
			deployer.Priv,
			evmtypes.EvmTxArgs{},
			testfactory.ContractDeploymentData{
				Contract:        contracts.ERC20MinterBurnerDecimalsContract,
				ConstructorArgs: []interface{}{"TestToken", "TT", uint8(18)},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC-20 contract")
		Expect(network.NextBlock()).To(BeNil())

		// Mint some tokens to the deployer.
		_, err = factory.ExecuteContractCall(
			deployer.Priv,
			evmtypes.EvmTxArgs{To: &erc20Addr},
			testfactory.CallArgs{
				ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
				MethodName:  "mint",
				Args: []interface{}{
					deployer.Addr, mintAmount,
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to mint tokens")
		Expect(network.NextBlock()).To(BeNil())

		// Check the balance of the deployer on the ERC20 contract.
		res, err := factory.ExecuteContractCall(
			deployer.Priv,
			evmtypes.EvmTxArgs{To: &erc20Addr},
			testfactory.CallArgs{
				ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
				MethodName:  "balanceOf",
				Args: []interface{}{
					deployer.Addr,
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to get balance")
		Expect(network.NextBlock()).To(BeNil())

		ethRes, err := evmtypes.DecodeTxResponse(res.Data)
		Expect(err).ToNot(HaveOccurred(), "failed to decode balance of tx response")

		unpacked, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Unpack(
			"balanceOf",
			ethRes.Ret,
		)
		Expect(err).ToNot(HaveOccurred(), "failed to unpack balance")

		balance, ok := unpacked[0].(*big.Int)
		Expect(ok).To(BeTrue(), "failed to convert balance to big.Int")
		Expect(balance.String()).To(Equal(mintAmount.String()), "balance is not correct")

		// Deploy the flash loan contract.
		flashLoanAddr, err = factory.DeployContract(
			deployer.Priv,
			evmtypes.EvmTxArgs{},
			testfactory.ContractDeploymentData{
				Contract: flashLoanContract,
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy flash loan contract")
		Expect(network.NextBlock()).To(BeNil())

		// Approve the flash loan contract to spend tokens. This is required because
		// the contract will get funds from the caller to perform actions.
		_, err = factory.ExecuteContractCall(
			deployer.Priv,
			evmtypes.EvmTxArgs{To: &erc20Addr},
			testfactory.CallArgs{
				ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
				MethodName:  "approve",
				Args: []interface{}{
					flashLoanAddr, mintAmount,
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to approve flash loan contract")
		Expect(network.NextBlock()).To(BeNil())

		// Check the allowance.
		res, err = factory.ExecuteContractCall(
			deployer.Priv,
			evmtypes.EvmTxArgs{To: &erc20Addr},
			testfactory.CallArgs{
				ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
				MethodName:  "allowance",
				Args: []interface{}{
					deployer.Addr, flashLoanAddr,
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to get allowance")
		Expect(network.NextBlock()).To(BeNil())

		ethRes, err = evmtypes.DecodeTxResponse(res.Data)
		Expect(err).ToNot(HaveOccurred(), "failed to decode allowance tx response")

		unpacked, err = contracts.ERC20MinterBurnerDecimalsContract.ABI.Unpack(
			"allowance",
			ethRes.Ret,
		)
		Expect(err).ToNot(HaveOccurred(), "failed to unpack allowance")

		var allowance *big.Int
		allowance, ok = unpacked[0].(*big.Int)
		Expect(ok).To(BeTrue(), "failed to convert allowance to big.Int")
		Expect(allowance.String()).To(Equal(mintAmount.String()), "allowance is not correct")

		// Approve the flash loan contract to delegate tokens on behalf of user.
		precompileAddr := stakingPrecompile.Address()

		_, err = factory.ExecuteContractCall(
			deployer.Priv,
			evmtypes.EvmTxArgs{To: &precompileAddr},
			testfactory.CallArgs{
				ContractABI: stakingPrecompile.ABI,
				MethodName:  "approve",
				Args: []interface{}{
					flashLoanAddr, delegateAmount, []string{stakingprecompile.DelegateMsg},
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to approve flash loan contract")
		Expect(network.NextBlock()).To(BeNil())

		// Check the allowance.
		res, err = factory.ExecuteContractCall(
			deployer.Priv,
			evmtypes.EvmTxArgs{To: &precompileAddr},
			testfactory.CallArgs{
				ContractABI: stakingPrecompile.ABI,
				MethodName:  "allowance",
				Args: []interface{}{
					deployer.Addr, flashLoanAddr, stakingprecompile.DelegateMsg,
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to get allowance")
		Expect(network.NextBlock()).To(BeNil())

		ethRes, err = evmtypes.DecodeTxResponse(res.Data)
		Expect(err).ToNot(HaveOccurred(), "failed to decode allowance tx response")

		err = stakingPrecompile.UnpackIntoInterface(&allowance, "allowance", ethRes.Ret)
		Expect(err).ToNot(HaveOccurred(), "failed to unpack allowance")
	})

	DescribeTable("call the flashLoan contract", func(tc testCase) {
		_, err := factory.ExecuteContractCall(
			deployer.Priv,
			evmtypes.EvmTxArgs{
				To:       &flashLoanAddr,
				GasPrice: big.NewInt(900_000_000),
				GasLimit: 400_000,
			},
			testfactory.CallArgs{
				ContractABI: flashLoanContract.ABI,
				MethodName:  tc.method,
				Args: []interface{}{
					erc20Addr,
					validatorToDelegateTo,
					delegateAmount,
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to execute flash loan")
		Expect(network.NextBlock()).To(BeNil())

		delRes, err := handler.GetDelegation(deployer.AccAddr.String(), validatorToDelegateTo)
		Expect(err).ToNot(HaveOccurred(), "failed to get delegation")
		delAmtPost := delRes.DelegationResponse.Balance.Amount
		if tc.expDelegation {
			Expect(delAmtPost).To(Equal(
				delegatedAmountPre.Add(math.NewIntFromBigInt(delegateAmount))),
				"delegated amount is not correct",
			)
		} else {
			Expect(delAmtPost).To(Equal(delegatedAmountPre))
		}

		// Check the ERC20 token balance of the deployer.
		res, err := factory.ExecuteContractCall(
			deployer.Priv,
			evmtypes.EvmTxArgs{To: &erc20Addr},
			testfactory.CallArgs{
				ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
				MethodName:  "balanceOf",
				Args: []interface{}{
					deployer.Addr,
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to get balance")
		Expect(network.NextBlock()).To(BeNil())

		ethRes, err := evmtypes.DecodeTxResponse(res.Data)
		Expect(err).ToNot(HaveOccurred(), "failed to decode balance of tx response")

		unpacked, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Unpack(
			"balanceOf",
			ethRes.Ret,
		)
		Expect(err).ToNot(HaveOccurred(), "failed to unpack balance")

		balance, ok := unpacked[0].(*big.Int)
		Expect(ok).To(BeTrue(), "failed to convert balance to big.Int")
		Expect(balance.String()).To(Equal(tc.expSenderERC20Balance.String()), "balance is not correct")

		// Check FlashLoan smart contract ERC20 token balance.
		res, err = factory.ExecuteContractCall(
			deployer.Priv,
			evmtypes.EvmTxArgs{To: &erc20Addr},
			testfactory.CallArgs{
				ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
				MethodName:  "balanceOf",
				Args: []interface{}{
					flashLoanAddr,
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to get balance")

		ethRes, err = evmtypes.DecodeTxResponse(res.Data)
		Expect(err).ToNot(HaveOccurred(), "failed to decode balance of tx response")

		unpacked, err = contracts.ERC20MinterBurnerDecimalsContract.ABI.Unpack(
			"balanceOf",
			ethRes.Ret,
		)
		Expect(err).ToNot(HaveOccurred(), "failed to unpack balance")

		balance, ok = unpacked[0].(*big.Int)
		Expect(ok).To(BeTrue(), "failed to convert balance to big.Int")
		Expect(balance.String()).To(Equal(tc.expContractERC20Balance.String()), "balance is not correct")
	},
		Entry("flasLoan method & expect delegation", testCase{
			method:                  "flashLoan",
			expDelegation:           true,
			expSenderERC20Balance:   mintAmount,
			expContractERC20Balance: big.NewInt(0),
		}),
		Entry("flashLoanWithRevert method - delegation reverted", testCase{
			method:                  "flashLoanWithRevert",
			expDelegation:           false,
			expSenderERC20Balance:   delegateAmount,
			expContractERC20Balance: delegateAmount,
		}),
	)
})
