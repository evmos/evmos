// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package eips_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/app/eips"
	"github.com/evmos/evmos/v19/app/eips/testdata"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	integrationutils "github.com/evmos/evmos/v19/testutil/integration/evmos/utils"

	evmtypes "github.com/evmos/evmos/v19/x/evm/types"

	"github.com/ethereum/go-ethereum/params"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

// Below tests are divided in 3 steps:
//  1. Deploy and interact with contracts to compute the gas used BEFORE enabling
//     the IP.
//  2. Activate the IP under test.
//  3. Deploy and interact with contracts to compute the gas used AFTER enabling
//     the IP.

func TestIPs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EvmosIPs Suite")
}

var _ = Describe("Improvement proposal evmos_0 - ", Ordered, func() {
	var (
		in network.Network
		tf factory.TxFactory
		gh grpc.Handler
		k  keyring.Keyring

		senderPriv  types.PrivKey
		senderPriv2 types.PrivKey
		senderAddr2 common.Address

		// Gas used before enabling the IP.
		gasUsedPre int64
	)

	// Multiplier used to modify the opcodes associated with evmos_0 IP.
	ipMultiplier := uint64(5)

	// The factory counter is used because it will create a new instance of
	// the counter contract, allowing to test the CREATE opcode.
	counterFactoryContract, err := testdata.LoadCounterFactoryContract()
	Expect(err).ToNot(HaveOccurred(), "failed to load Counter Factory contract")

	deploymentData := factory.ContractDeploymentData{
		Contract:        counterFactoryContract,
		ConstructorArgs: []interface{}{},
	}

	BeforeAll(func() {
		k = keyring.New(2)
		in = network.New(
			network.WithPreFundedAccounts(k.GetAllAccAddrs()...),
		)
		gh = grpc.NewIntegrationHandler(in)
		tf = factory.New(in, gh)

		// Account used to deploy the contract before enabling the IP.
		senderPriv = k.GetPrivKey(0)
		// Account used to deploy the contract after enabling the IP. A second
		// account is used to avoid possible additional gas costs due to the change
		// in the Nonce.
		senderPriv2 = k.GetPrivKey(1)
		senderAddr2 = k.GetAddr(1)

		// Set extra IPs to empty to allow testing a single modifier.
		defaultParams := evmtypes.DefaultParams()
		defaultParams.ExtraEIPs = []string{}

		err := integrationutils.UpdateEvmParams(
			integrationutils.UpdateParamsInput{
				Tf:      tf,
				Network: in,
				Pk:      senderPriv,
				Params:  defaultParams,
			},
		)
		Expect(err).To(BeNil(), "failed during update of evm params")
		Expect(in.NextBlock()).To(BeNil())
	})

	It("should deploy the contract before enabling the IP", func() {
		deploymentTxArgs, err := tf.GenerateDeployContractArgs(senderAddr2, evmtypes.EvmTxArgs{}, deploymentData)
		Expect(err).To(BeNil(), "failed to create deployment tx args")

		res, err := tf.ExecuteEthTx(senderPriv2, deploymentTxArgs)
		Expect(err).To(BeNil(), "failed during contract deployment")
		gasUsedPre = res.GasUsed
	})

	It("should enable the new IP", func() {
		eips.Multiplier = ipMultiplier
		newIP := "evmos_0"

		qRes, err := gh.GetEvmParams()
		Expect(err).To(BeNil(), "failed during query to evm params")
		qRes.Params.ExtraEIPs = append(qRes.Params.ExtraEIPs, newIP)
		err = integrationutils.UpdateEvmParams(
			integrationutils.UpdateParamsInput{
				Tf:      tf,
				Network: in,
				Pk:      senderPriv,
				Params:  qRes.Params,
			},
		)
		Expect(err).To(BeNil(), "failed during update of evm params")

		Expect(in.NextBlock()).To(BeNil())

		qRes, err = gh.GetEvmParams()
		Expect(err).To(BeNil(), "failed during query to evm params")
		Expect(qRes.Params.ExtraEIPs).To(ContainElement(newIP), "expected to have IP evmos_0 in evm params")
	})

	It("should change CREATE opcode constant gas after enabling evmos_0 IP", func() {
		gasCostPre := params.CreateGas

		deploymentTxArgs, err := tf.GenerateDeployContractArgs(senderAddr2, evmtypes.EvmTxArgs{}, deploymentData)
		Expect(err).To(BeNil(), "failed to create deployment tx args")

		res, err := tf.ExecuteEthTx(senderPriv2, deploymentTxArgs)
		Expect(err).To(BeNil(), "failed during contract deployment")
		// commit block to update sender nonce
		Expect(in.NextBlock()).To(BeNil())

		gasUsedPost := res.GasUsed

		// The difference in gas is the new cost of the opcode, minus the cost of the
		// opcode before enabling the new eip.
		gasUsedDiff := ipMultiplier*gasCostPre - gasCostPre
		expectedGas := gasUsedPre + int64(gasUsedDiff)
		Expect(gasUsedPost).To(Equal(expectedGas))
	})
})

var _ = Describe("Improvement proposal evmos_1 - ", Ordered, func() {
	var (
		in network.Network
		tf factory.TxFactory
		gh grpc.Handler
		k  keyring.Keyring

		senderPriv types.PrivKey

		// Gas used before enabling the IP.
		gasUsedPre int64

		// The address of the factory counter.
		counterFactoryAddr common.Address
	)

	// Multiplier used to modify the opcodes associated with evmos_1.
	eipMultiplier := uint64(5)
	initialCounterValue := 1

	// The counter factory contract is used to deploy a counter contract and
	// perform state transition using the CALL opcode.
	counterFactoryContract, err := testdata.LoadCounterFactoryContract()
	Expect(err).ToNot(HaveOccurred(), "failed to load Counter Factory contract")

	BeforeAll(func() {
		k = keyring.New(1)
		in = network.New(
			network.WithPreFundedAccounts(k.GetAllAccAddrs()...),
		)
		gh = grpc.NewIntegrationHandler(in)
		tf = factory.New(in, gh)

		senderPriv = k.GetPrivKey(0)

		// Set extra IPs to empty to allow testing a single modifier.
		defaultParams := evmtypes.DefaultParams()
		defaultParams.ExtraEIPs = []string{}
		err = integrationutils.UpdateEvmParams(
			integrationutils.UpdateParamsInput{
				Tf:      tf,
				Network: in,
				Pk:      senderPriv,
				Params:  defaultParams,
			},
		)
		Expect(err).To(BeNil(), "failed during update of evm params")

		Expect(in.NextBlock()).To(BeNil())
	})

	It("should deploy the contract before enabling the IP", func() {
		counterFactoryAddr, err = tf.DeployContract(
			senderPriv,
			evmtypes.EvmTxArgs{},
			factory.ContractDeploymentData{
				Contract:        counterFactoryContract,
				ConstructorArgs: []interface{}{},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy counter factory contract")
		Expect(in.NextBlock()).To(BeNil())

		res, err := tf.ExecuteContractCall(
			senderPriv,
			evmtypes.EvmTxArgs{To: &counterFactoryAddr},
			factory.CallArgs{
				ContractABI: counterFactoryContract.ABI,
				MethodName:  "incrementCounter",
				Args:        []interface{}{},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to increment counter value")
		gasUsedPre = res.GasUsed

		Expect(in.NextBlock()).To(BeNil())

		// Query the counter value to check proper state transition later.
		res, err = tf.ExecuteContractCall(
			senderPriv,
			evmtypes.EvmTxArgs{To: &counterFactoryAddr},
			factory.CallArgs{
				ContractABI: counterFactoryContract.ABI,
				MethodName:  "getCounterValue",
				Args:        []interface{}{},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to get counter value")
		Expect(in.NextBlock()).To(BeNil())

		ethRes, err := evmtypes.DecodeTxResponse(res.Data)
		Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

		unpacked, err := counterFactoryContract.ABI.Unpack(
			"getCounterValue",
			ethRes.Ret,
		)
		Expect(err).ToNot(HaveOccurred(), "failed to unpack counter value")

		counter, ok := unpacked[0].(*big.Int)
		Expect(ok).To(BeTrue(), "failed to convert counter to big.Int")
		Expect(counter.String()).To(Equal(fmt.Sprintf("%d", initialCounterValue+1)), "counter is not correct")
	})
	It("should enable the new IP", func() {
		eips.Multiplier = eipMultiplier
		newIP := "evmos_1"

		qRes, err := gh.GetEvmParams()
		Expect(err).To(BeNil(), "failed during query to evm params")
		qRes.Params.ExtraEIPs = append(qRes.Params.ExtraEIPs, newIP)

		err = integrationutils.UpdateEvmParams(
			integrationutils.UpdateParamsInput{
				Tf:      tf,
				Network: in,
				Pk:      senderPriv,
				Params:  qRes.Params,
			},
		)
		Expect(err).To(BeNil(), "failed during update of evm params")

		Expect(in.NextBlock()).To(BeNil())

		qRes, err = gh.GetEvmParams()
		Expect(err).To(BeNil(), "failed during query to evm params")
		Expect(qRes.Params.ExtraEIPs).To(ContainElement(newIP), "expected to have ip evmos_1 in evm params")
	})
	It("should change CALL opcode constant gas after enabling IP", func() {
		// Constant gas cost used before enabling the new IP.
		gasCostPre := params.WarmStorageReadCostEIP2929

		res, err := tf.ExecuteContractCall(
			senderPriv,
			evmtypes.EvmTxArgs{To: &counterFactoryAddr},
			factory.CallArgs{
				ContractABI: counterFactoryContract.ABI,
				MethodName:  "incrementCounter",
				Args:        []interface{}{},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to increment counter value")
		gasUsedPost := res.GasUsed
		Expect(in.NextBlock()).To(BeNil())

		res, err = tf.ExecuteContractCall(
			senderPriv,
			evmtypes.EvmTxArgs{To: &counterFactoryAddr},
			factory.CallArgs{
				ContractABI: counterFactoryContract.ABI,
				MethodName:  "getCounterValue",
				Args:        []interface{}{},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to get counter value")
		Expect(in.NextBlock()).To(BeNil())

		ethRes, err := evmtypes.DecodeTxResponse(res.Data)
		Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

		unpacked, err := counterFactoryContract.ABI.Unpack(
			"getCounterValue",
			ethRes.Ret,
		)
		Expect(err).ToNot(HaveOccurred(), "failed to unpack counter value")

		counter, ok := unpacked[0].(*big.Int)
		Expect(ok).To(BeTrue(), "failed to convert counter to big.Int")
		Expect(counter.String()).To(Equal(fmt.Sprintf("%d", initialCounterValue+2)), "counter is not updated correctly")

		// The difference in gas is the new cost of the opcode, minus the cost of the
		// opcode before enabling the new eip.
		gasUsedDiff := eipMultiplier*gasCostPre - gasCostPre
		expectedGas := gasUsedPre + int64(gasUsedDiff)
		Expect(gasUsedPost).To(Equal(expectedGas))
	})
})

var _ = Describe("Improvement proposal evmos_2 - ", Ordered, func() {
	var (
		in network.Network
		tf factory.TxFactory
		gh grpc.Handler
		k  keyring.Keyring

		senderPriv  types.PrivKey
		senderAddr  common.Address
		senderPriv2 types.PrivKey
		senderAddr2 common.Address
		gasUsedPre  int64
	)
	// Constant gas used to modify the opcodes associated with evmos_2.
	constantGas := uint64(500)

	counterContract, err := testdata.LoadCounterContract()
	Expect(err).ToNot(HaveOccurred(), "failed to load Counter contract")

	deploymentData := factory.ContractDeploymentData{
		Contract:        counterContract,
		ConstructorArgs: []interface{}{},
	}
	BeforeAll(func() {
		k = keyring.New(2)
		in = network.New(
			network.WithPreFundedAccounts(k.GetAllAccAddrs()...),
		)
		gh = grpc.NewIntegrationHandler(in)
		tf = factory.New(in, gh)

		// Account used to deploy the contract before enabling the IP.
		senderPriv = k.GetPrivKey(0)
		senderAddr = k.GetAddr(0)
		// Account used to deploy the contract after enabling the IP. A second
		// account is used to avoid possible additional gas costs due to the change
		// in the Nonce.
		senderPriv2 = k.GetPrivKey(0)
		senderAddr2 = k.GetAddr(0)

		// Set extra IPs to empty to allow testing a single modifier.
		defaultParams := evmtypes.DefaultParams()
		defaultParams.ExtraEIPs = []string{}

		err = integrationutils.UpdateEvmParams(
			integrationutils.UpdateParamsInput{
				Tf:      tf,
				Network: in,
				Pk:      senderPriv,
				Params:  defaultParams,
			},
		)
		Expect(err).To(BeNil(), "failed during update of evm params")

		Expect(in.NextBlock()).To(BeNil())
	})

	It("should deploy the contract before enabling the IP", func() {
		deploymentTxArgs, err := tf.GenerateDeployContractArgs(senderAddr, evmtypes.EvmTxArgs{}, deploymentData)
		Expect(err).To(BeNil(), "failed to create deployment tx args")

		res, err := tf.ExecuteEthTx(senderPriv, deploymentTxArgs)
		Expect(err).To(BeNil(), "failed during contract deployment")
		Expect(in.NextBlock()).To(BeNil())

		gasUsedPre = res.GasUsed
	})

	It("should enable the new IP", func() {
		eips.SstoreConstantGas = constantGas
		newIP := "evmos_2"

		qRes, err := gh.GetEvmParams()
		Expect(err).To(BeNil(), "failed during query to evm params")
		qRes.Params.ExtraEIPs = append(qRes.Params.ExtraEIPs, newIP)
		err = integrationutils.UpdateEvmParams(
			integrationutils.UpdateParamsInput{
				Tf:      tf,
				Network: in,
				Pk:      senderPriv,
				Params:  qRes.Params,
			},
		)
		Expect(err).To(BeNil(), "failed during update of evm params")

		Expect(in.NextBlock()).To(BeNil())

		qRes, err = gh.GetEvmParams()
		Expect(err).To(BeNil(), "failed during query to evm params")
		Expect(qRes.Params.ExtraEIPs).To(ContainElement(newIP), "expected to have ip evmos_2 in evm params")
	})

	It("should change SSTORE opcode constant gas after enabling IP", func() {
		deploymentTxArgs, err := tf.GenerateDeployContractArgs(senderAddr2, evmtypes.EvmTxArgs{}, deploymentData)
		Expect(err).To(BeNil(), "failed to create deployment tx args")

		res, err := tf.ExecuteEthTx(senderPriv2, deploymentTxArgs)
		Expect(err).To(BeNil(), "failed during contract deployment")
		Expect(in.NextBlock()).To(BeNil())

		gasUsedPost := res.GasUsed

		// The expected gas is previous gas plus the constant gas because
		// previous this eip, SSTORE was using only the dynamic gas.
		expectedGas := gasUsedPre + int64(constantGas)
		Expect(gasUsedPost).To(Equal(expectedGas))
	})
})
