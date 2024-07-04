// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package app_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/contracts"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"

	evmtypes "github.com/evmos/evmos/v18/x/evm/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

func TestEIPs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EIPs Suite")
}

var _ = Describe("Custom EIPs - ", func() {
	// Variables used and modified by the below closures.
	var (
		in network.Network
		tf factory.TxFactory
		gh grpc.Handler
		k  keyring.Keyring

		senderPriv types.PrivKey
		senderAddr common.Address
	)

	BeforeEach(func() {
		// configure the network before each ordered execution

		k = keyring.New(1)
		in = network.New(
			network.WithPreFundedAccounts(k.GetAllAccAddrs()...),
		)
		gh = grpc.NewIntegrationHandler(in)
		tf = factory.New(in, gh)

		senderPriv = k.GetPrivKey(0)
		senderAddr = k.GetAddr(0)
	})

	Describe("EIP0000", Ordered, func() {
		// Used to store the gas used by the first instantiation of the contract
		var initialGasUsed int64

		constructorArgs := []interface{}{"coin", "token", uint8(100)}
		compiledContract := contracts.ERC20MinterBurnerDecimalsContract

		// Deploy the contract used to check the effect of the EIP changes.
		txArgs := evmtypes.EvmTxArgs{}
		deploymentData := factory.ContractDeploymentData{
			Contract:        compiledContract,
			ConstructorArgs: constructorArgs,
		}

		It("should deploy the contract before enabling the EIP", func() {
			deploymentTxArgs, err := tf.GenerateDeployContractArgs(senderAddr, txArgs, deploymentData)
			Expect(err).To(BeNil(), "Creation of deployment tx args should not fail")

			res, err := tf.ExecuteEthTx(senderPriv, deploymentTxArgs)
			Expect(err).To(BeNil(), "Contract deployment should not fail")

			fmt.Println("First deployment gas used: ", res.GasUsed)

			err = in.NextBlock()
			Expect(err).To(BeNil())

			res, err = tf.ExecuteEthTx(senderPriv, deploymentTxArgs)
			Expect(err).To(BeNil(), "Contract deployment should not fail")

			fmt.Println("Second deployment gas used: ", res.GasUsed)
			initialGasUsed = res.GasUsed
		})
		It("should change the gas used after enabling the EIP", func() {
			// change the eip and redeploy/call the contract

			newEIP := 0002
			// prevGasCost := 20_000

			deploymentTxArgs, err := tf.GenerateDeployContractArgs(senderAddr, txArgs, deploymentData)
			Expect(err).To(BeNil(), "Creation of deployment tx args should not fail")

			defaultParams := evmtypes.DefaultParams()
			fmt.Println(defaultParams.ExtraEIPs)
			defaultParams.ExtraEIPs = append(defaultParams.ExtraEIPs, int64(newEIP))

			err = in.UpdateEvmParams(defaultParams)
			Expect(err).To(BeNil(), "EVM paramas should be correctly modified")

			qRes, err := gh.GetEvmParams()
			Expect(err).To(BeNil(), "Query to EVM Params should not fail")
			Expect(qRes.Params.ExtraEIPs).To(ContainElement(int64(newEIP)), "Expected to have EIP 0000 in EVM Params")

			res, err := tf.ExecuteEthTx(senderPriv, deploymentTxArgs)
			Expect(err).To(BeNil(), "Contract deployment should not fail")

			fmt.Println("Gas used: ", res.GasUsed)
			fmt.Println("Old gas used: ", initialGasUsed)
			fmt.Println("Difference: ", res.GasUsed-initialGasUsed)
			Expect(res.GasUsed).To(Equal(initialGasUsed))
		})
	})
})
