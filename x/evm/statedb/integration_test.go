// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package statedb_test

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/contracts"
	stakingprecompile "github.com/evmos/evmos/v18/precompiles/staking"
	testfactory "github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v18/x/evm/statedb/testdata"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"

	//nolint:revive // okay to use dot imports for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // okay to use dot imports for Ginkgo
	. "github.com/onsi/gomega"
)

func TestNestedEVMExtensionCall(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nested EVM Extension Call Test Suite")
}

// NOTE: This must ONLY be added to the main repository once the vulnerability has been patched!
//
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

		valsRes, err := handler.GetBondedValidators()
		Expect(err).ToNot(HaveOccurred(), "failed to get bonded validators")

		validatorToDelegateTo = valsRes.Validators[0].OperatorAddress
		res, err := handler.GetDelegation(deployer.AccAddr.String(), validatorToDelegateTo)
		Expect(err).ToNot(HaveOccurred(), "failed to get delegation")
		delegatedAmountPre = res.DelegationResponse.Balance.Amount

		// Load the flash loan contract from the compiled JSON data
		flashLoanContract, err = testdata.LoadFlashLoanContract()
		Expect(err).ToNot(HaveOccurred(), "failed to load flash loan contract")
	})

	It("should deploy an ERC-20 token contract", func() {
		var err error
		erc20Addr, err = factory.DeployContract(
			deployer.Priv,
			evmtypes.EvmTxArgs{},
			testfactory.ContractDeploymentData{
				Contract:        contracts.ERC20MinterBurnerDecimalsContract,
				ConstructorArgs: []interface{}{"TestToken", "TT", uint8(18)},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC-20 contract")
	})

	It("should mint some tokens", func() {
		// Mint some tokens to the deployer
		_, err := factory.ExecuteContractCall(
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

		// Check the balance of the deployer
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

		ethRes, err := evmtypes.DecodeTxResponse(res.Data)
		Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

		unpacked, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Unpack(
			"balanceOf",
			ethRes.Ret,
		)
		Expect(err).ToNot(HaveOccurred(), "failed to unpack balance")

		balance, ok := unpacked[0].(*big.Int)
		Expect(ok).To(BeTrue(), "failed to convert balance to big.Int")
		Expect(balance.String()).To(Equal(mintAmount.String()), "balance is not correct")
	})

	It("should deploy the flash loan contract", func() {
		var err error
		flashLoanAddr, err = factory.DeployContract(
			deployer.Priv,
			evmtypes.EvmTxArgs{},
			testfactory.ContractDeploymentData{
				Contract: flashLoanContract,
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy flash loan contract")
	})

	It("should approve the flash loan contract to spend tokens", func() {
		_, err := factory.ExecuteContractCall(
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

		// Check the allowance
		res, err := factory.ExecuteContractCall(
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

		ethRes, err := evmtypes.DecodeTxResponse(res.Data)
		Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

		unpacked, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Unpack(
			"allowance",
			ethRes.Ret,
		)
		Expect(err).ToNot(HaveOccurred(), "failed to unpack allowance")

		allowance, ok := unpacked[0].(*big.Int)
		Expect(ok).To(BeTrue(), "failed to convert allowance to big.Int")
		Expect(allowance.String()).To(Equal(mintAmount.String()), "allowance is not correct")
	})

	It("should approve the flash loan contract to delegate tokens on behalf of user", func() {
		precompileAddr := stakingPrecompile.Address()

		_, err := factory.ExecuteContractCall(
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

		// Check the allowance
		res, err := factory.ExecuteContractCall(
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

		ethRes, err := evmtypes.DecodeTxResponse(res.Data)
		Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

		var allowance *big.Int
		err = stakingPrecompile.ABI.UnpackIntoInterface(&allowance, "allowance", ethRes.Ret)
		Expect(err).ToNot(HaveOccurred(), "failed to unpack allowance")
	})

	It("should execute the flash loan contract", func() {
		// Execute the flash loan contract
		_, err := factory.ExecuteContractCall(
			deployer.Priv,
			evmtypes.EvmTxArgs{To: &flashLoanAddr},
			testfactory.CallArgs{
				ContractABI: flashLoanContract.ABI,
				MethodName:  "flashLoan",
				Args: []interface{}{
					erc20Addr,
					validatorToDelegateTo,
					delegateAmount,
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to execute flash loan")
	})

	It("should show the delegation", func() {
		delRes, err := handler.GetDelegation(deployer.AccAddr.String(), validatorToDelegateTo)
		Expect(err).ToNot(HaveOccurred(), "failed to get delegation")
		Expect(delRes.DelegationResponse.Balance.Amount.String()).To(Equal(
			delegatedAmountPre.Add(math.NewIntFromBigInt(delegateAmount)).String()),
			"delegated amount is not correct",
		)
	})

	It("should have returned the funds from the flash loan", func() {
		// Check the balance of the deployer
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

		ethRes, err := evmtypes.DecodeTxResponse(res.Data)
		Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

		unpacked, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Unpack(
			"balanceOf",
			ethRes.Ret,
		)
		Expect(err).ToNot(HaveOccurred(), "failed to unpack balance")

		balance, ok := unpacked[0].(*big.Int)
		Expect(ok).To(BeTrue(), "failed to convert balance to big.Int")
		Expect(balance.String()).To(Equal(mintAmount.String()), "balance is not correct")
	})
})
