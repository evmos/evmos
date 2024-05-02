// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package strv2_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/contracts"
	commonfactory "github.com/evmos/evmos/v18/testutil/integration/common/factory"
	testfactory "github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v18/utils"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"

	"github.com/evmos/evmos/v18/x/erc20/keeper/testdata"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"

	. "github.com/onsi/ginkgo/v2" //nolint:revive
	. "github.com/onsi/gomega"    //nolint:revive
)

func TestSTRv2Tracking(t *testing.T) {
	// Run Ginkgo BDD tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "STRv2 Tracking Tests")
}

type STRv2TrackingSuite struct {
	keyring testkeyring.Keyring
	network *testnetwork.UnitTestNetwork
	handler grpc.Handler
	factory testfactory.TxFactory

	nativeCoinERC20Addr   common.Address
	registeredERC20Addr   common.Address
	unregisteredERC20Addr common.Address
	wevmosAddr            common.Address
}

const (
	deployerIdx        = 0
	nativeIBCCoinDenom = "coin"
)

var (
	mintAmount     = big.NewInt(1000000000000000000)
	convertAmount  = testnetwork.PrefundedAccountInitialBalance.QuoRaw(10)
	transferAmount = convertAmount.QuoRaw(10).BigInt()
)

var _ = Describe("STRv2 Tracking -", func() {
	var s *STRv2TrackingSuite

	BeforeEach(func() {
		var err error
		s, err = CreateTestSuite(utils.MainnetChainID + "-1")
		Expect(err).ToNot(HaveOccurred(), "failed to create test suite")

		// NOTE: this is necessary to enable e.g. erc20Keeper.BalanceOf(...) to work
		// correctly internally.
		// Removing it will break a bunch of tests giving errors like: "failed to retrieve balance"
		Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")
	})

	When("sending an EVM transaction", func() {
		Context("which interacts with a registered native token pair ERC-20 contract", func() {
			Context("in a direct call to the token pair contract", func() {
				It("should add the from and to addresses to the store if it is not already stored", func() { //nolint:all
					sender := s.keyring.GetKey(0)
					receiver := s.keyring.GetKey(2)

					senderAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")
					receiverAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), receiver.AccAddr)
					Expect(receiverAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")

					_, err := s.factory.ExecuteContractCall(
						sender.Priv,
						evmtypes.EvmTxArgs{
							To: &s.nativeCoinERC20Addr,
						},
						testfactory.CallArgs{
							ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
							MethodName:  "transfer",
							Args: []interface{}{
								receiver.Addr,
								transferAmount,
							},
						},
					)
					Expect(err).ToNot(HaveOccurred(), "failed to transfer tokens of Cosmos native ERC-20 token pair")

					Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

					senderAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeTrue(), "expected address to be stored")
					receiverAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(receiverAddrTracked).To(BeTrue(), "expected address to be stored")
				})

				It("should not fail if the addresses are already stored", func() {
					sender := s.keyring.GetKey(0)
					receiver := s.keyring.GetKey(2)

					senderAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")
					receiverAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), receiver.AccAddr)
					Expect(receiverAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")

					_, err := s.factory.ExecuteContractCall(
						sender.Priv,
						evmtypes.EvmTxArgs{
							To: &s.nativeCoinERC20Addr,
						},
						testfactory.CallArgs{
							ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
							MethodName:  "transfer",
							Args: []interface{}{
								receiver.Addr,
								transferAmount,
							},
						},
					)
					Expect(err).ToNot(HaveOccurred(), "failed to transfer tokens of Cosmos native ERC-20 token pair")

					Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

					_, err = s.factory.ExecuteContractCall(
						sender.Priv,
						evmtypes.EvmTxArgs{
							To: &s.nativeCoinERC20Addr,
						},
						testfactory.CallArgs{
							ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
							MethodName:  "transfer",
							Args: []interface{}{
								receiver.Addr,
								transferAmount,
							},
						},
					)
					Expect(err).ToNot(HaveOccurred(), "failed to transfer tokens of Cosmos native ERC-20 token pair (2nd call)")

					Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

					senderAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeTrue(), "expected address to be still stored")
					receiverAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), receiver.AccAddr)
					Expect(receiverAddrTracked).To(BeTrue(), "expected address to be still stored")
				})
				It("should not store anything if calling a different method than transfer or transferFrom", func() { //nolint:all
					sender := s.keyring.GetKey(0)
					grantee := s.keyring.GetKey(2)

					senderAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")
					receiverAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), grantee.AccAddr)
					Expect(receiverAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")

					_, err := s.factory.ExecuteContractCall(
						sender.Priv,
						evmtypes.EvmTxArgs{
							To: &s.nativeCoinERC20Addr,
						},
						testfactory.CallArgs{
							ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
							MethodName:  "approve",
							Args: []interface{}{
								grantee.Addr,
								transferAmount,
							},
						},
					)
					Expect(err).ToNot(HaveOccurred(), "failed to approve ERC-20 transfer")

					Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

					senderAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeFalse(), "expected address not to be stored")
					receiverAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), grantee.AccAddr)
					Expect(receiverAddrTracked).To(BeFalse(), "expected address not to be stored")
				})
			})

			Context("in a call to the token pair contract from another contract", func() {
				var (
					senderIdx         = deployerIdx
					tokenTransferAddr common.Address
				)

				BeforeEach(func() {
					deployer := s.keyring.GetKey(deployerIdx)
					sender := s.keyring.GetKey(senderIdx)

					var err error
					tokenTransferAddr, err = s.factory.DeployContract(
						deployer.Priv,
						evmtypes.EvmTxArgs{},
						testfactory.ContractDeploymentData{
							Contract:        testdata.TokenTransferContract,
							ConstructorArgs: []interface{}{s.nativeCoinERC20Addr},
						},
					)
					Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC-20 transfer contract")

					Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

					// approve the contract to spend on behalf of the sender
					_, err = s.factory.ExecuteContractCall(
						sender.Priv,
						evmtypes.EvmTxArgs{
							To: &s.nativeCoinERC20Addr,
						},
						testfactory.CallArgs{
							ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
							MethodName:  "approve",
							Args: []interface{}{
								tokenTransferAddr,
								transferAmount,
							},
						},
					)
					Expect(err).ToNot(HaveOccurred(), "failed to approve ERC-20 transfer contract to spend on behalf of the sender")

					Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")
				})

				It("should add the from AND to address to the store if it is not already stored", func() {
					sender := s.keyring.GetKey(senderIdx)
					receiver := s.keyring.GetKey(senderIdx + 1)

					senderAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")
					receiverAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), receiver.AccAddr)
					Expect(receiverAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")

					_, err := s.factory.ExecuteContractCall(
						sender.Priv,
						evmtypes.EvmTxArgs{
							To: &tokenTransferAddr,
						},
						testfactory.CallArgs{
							ContractABI: testdata.TokenTransferContract.ABI,
							MethodName:  "transferToken",
							Args: []interface{}{
								receiver.Addr,
								transferAmount,
							},
						},
					)
					Expect(err).ToNot(HaveOccurred(), "failed to transfer tokens of Cosmos native ERC-20 token pair")

					Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

					senderAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeTrue(), "expected address to be stored")
					receiverAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), receiver.AccAddr)
					Expect(receiverAddrTracked).To(BeTrue(), "expected address to be stored")
				})

				It("should add the from address if sending to the ERC-20 module address", func() {
					sender := s.keyring.GetKey(senderIdx)

					senderAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")

					_, err := s.factory.ExecuteContractCall(
						sender.Priv,
						evmtypes.EvmTxArgs{
							To: &tokenTransferAddr,
						},
						testfactory.CallArgs{
							ContractABI: testdata.TokenTransferContract.ABI,
							MethodName:  "transferToken",
							Args: []interface{}{
								erc20types.ModuleAddress,
								transferAmount,
							},
						},
					)
					Expect(err).ToNot(HaveOccurred(), "failed to transfer tokens of Cosmos native ERC-20 token pair")

					Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

					senderAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeTrue(), "expected address to be stored")
				})
			})

			// NOTE: this is running the coin conversion too
			Context("sending tokens to the module address", func() {
				It("should add the sender address in a direct call", func() {
					sender := s.keyring.GetKey(0)

					senderAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")

					_, err := s.factory.ExecuteContractCall(
						sender.Priv,
						evmtypes.EvmTxArgs{
							To: &s.nativeCoinERC20Addr,
						},
						testfactory.CallArgs{
							ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
							MethodName:  "transfer",
							Args: []interface{}{
								erc20types.ModuleAddress,
								transferAmount,
							},
						},
					)
					Expect(err).ToNot(HaveOccurred(), "failed to transfer tokens of Cosmos native ERC-20 token pair")

					Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

					senderAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeTrue(), "expected address to be stored")
					erc20AddrTrack := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), erc20types.ModuleAddress.Bytes())
					Expect(erc20AddrTrack).To(BeFalse(), "expected module address not to be stored")
				})
			})
		})

		Context("which interacts with a registered non-native token pair ERC-20 contract", func() {
			It("should not add the address to the store", func() {
				deployer := s.keyring.GetKey(deployerIdx)

				addrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), deployer.AccAddr)
				Expect(addrTracked).To(BeFalse(), "expected address not to be stored before conversion")

				_, err := s.factory.ExecuteContractCall(
					deployer.Priv,
					evmtypes.EvmTxArgs{
						To: &s.registeredERC20Addr,
					},
					testfactory.CallArgs{
						ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
						MethodName:  "mint",
						Args: []interface{}{
							deployer.Addr,
							mintAmount,
						},
					},
				)
				Expect(err).ToNot(HaveOccurred(), "failed to interact with registered ERC-20 contract")

				addrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), deployer.AccAddr)
				Expect(addrTracked).To(BeFalse(), "expected address to not be stored")
			})
		})

		Context("which interacts with an unregistered ERC-20 contract", func() {
			It("should not add the address to the store", func() {
				deployer := s.keyring.GetKey(deployerIdx)

				_, err := s.factory.ExecuteContractCall(
					deployer.Priv,
					evmtypes.EvmTxArgs{
						To: &s.unregisteredERC20Addr,
					},
					testfactory.CallArgs{
						ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
						MethodName:  "mint",
						Args: []interface{}{
							deployer.Addr,
							mintAmount,
						},
					},
				)
				Expect(err).ToNot(HaveOccurred(), "failed to interact with unregistered ERC-20 contract")

				deployerAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), deployer.AccAddr)
				Expect(deployerAddrTracked).To(BeFalse(), "expected address to not be stored")
			})
		})
	})

	When("manually converting", func() {
		Context("a registered coin into its ERC-20 representation", func() {
			It("should add the address to the store if it is not already stored", func() {
				sender := s.keyring.GetKey(1)

				addrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
				Expect(addrTracked).To(BeFalse(), "expected address not to be stored before conversion")

				_, err := s.factory.ExecuteCosmosTx(
					sender.Priv,
					commonfactory.CosmosTxArgs{
						Msgs: []sdk.Msg{
							&erc20types.MsgConvertCoin{
								Sender:   sender.AccAddr.String(),
								Receiver: sender.Addr.String(),
								Coin:     sdk.NewCoin(nativeIBCCoinDenom, convertAmount),
							},
						},
					},
				)
				Expect(err).ToNot(HaveOccurred(), "failed to convert native IBC coin")

				Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

				addrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
				Expect(addrTracked).To(BeTrue(), "expected address to be stored")
			})

			// TODO: is this correct? Yes, because only the addresses with ERC-20 tokens are relevant?
			It("should store only the receiving address if the sender and receiver are not the same account", func() {
				sender := s.keyring.GetKey(1)
				receiver := s.keyring.GetKey(2)

				senderAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
				Expect(senderAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")
				receiverAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), receiver.AccAddr)
				Expect(receiverAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")

				_, err := s.factory.ExecuteCosmosTx(
					sender.Priv,
					commonfactory.CosmosTxArgs{
						Msgs: []sdk.Msg{
							&erc20types.MsgConvertCoin{
								Sender:   sender.AccAddr.String(),
								Receiver: receiver.Addr.String(),
								Coin:     sdk.NewCoin(nativeIBCCoinDenom, convertAmount),
							},
						},
					},
				)
				Expect(err).ToNot(HaveOccurred(), "failed to convert native IBC coin")

				Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

				senderAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
				Expect(senderAddrTracked).To(BeFalse(), "expected address to be stored")
				receiverAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), receiver.AccAddr)
				Expect(receiverAddrTracked).To(BeTrue(), "expected address to be stored")
			})

			It("should not fail if the address is already stored", func() {
				sender := s.keyring.GetKey(1)

				addrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
				Expect(addrTracked).To(BeFalse(), "expected address not to be stored before conversion")

				_, err := s.factory.ExecuteCosmosTx(
					sender.Priv,
					commonfactory.CosmosTxArgs{
						Msgs: []sdk.Msg{
							&erc20types.MsgConvertCoin{
								Sender:   sender.AccAddr.String(),
								Receiver: sender.Addr.String(),
								Coin:     sdk.NewCoin(nativeIBCCoinDenom, testnetwork.PrefundedAccountInitialBalance.QuoRaw(10)),
							},
						},
					},
				)
				Expect(err).ToNot(HaveOccurred(), "failed to convert native IBC coin")

				Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

				_, err = s.factory.ExecuteCosmosTx(
					sender.Priv,
					commonfactory.CosmosTxArgs{
						Msgs: []sdk.Msg{
							&erc20types.MsgConvertCoin{
								Sender:   sender.AccAddr.String(),
								Receiver: sender.Addr.String(),
								Coin:     sdk.NewCoin(nativeIBCCoinDenom, testnetwork.PrefundedAccountInitialBalance.QuoRaw(10)),
							},
						},
					},
				)
				Expect(err).ToNot(HaveOccurred(), "failed to convert native IBC coin")

				Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

				addrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
				Expect(addrTracked).To(BeTrue(), "expected address to be still stored")
			})
		})

		Context("a registered ERC-20 representation into its native coin", func() {
			It("should add the address to the store if it is not already stored", func() {
				sender := s.keyring.GetKey(deployerIdx)

				addrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
				Expect(addrTracked).To(BeFalse(), "expected address not to be stored before conversion")

				_, err := s.factory.ExecuteCosmosTx(
					sender.Priv,
					commonfactory.CosmosTxArgs{
						Msgs: []sdk.Msg{
							&erc20types.MsgConvertERC20{
								ContractAddress: s.nativeCoinERC20Addr.String(),
								Sender:          sender.Addr.String(),
								Receiver:        sender.AccAddr.String(),
								Amount:          convertAmount,
							},
						},
					},
				)
				Expect(err).ToNot(HaveOccurred(), "failed to convert native IBC coin")

				Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

				addrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
				Expect(addrTracked).To(BeTrue(), "expected address to be stored")
			})
		})
	})
})

var _ = Describe("STRv2 Tracking Wevmos-", func() {
	var s *STRv2TrackingSuite

	BeforeEach(func() {
		var err error
		s, err = CreateTestSuite(utils.TestingChainID + "-1")
		Expect(err).ToNot(HaveOccurred(), "failed to create test suite")

		// NOTE: this is necessary to enable e.g. erc20Keeper.BalanceOf(...) to work
		// correctly internally.
		// Removing it will break a bunch of tests giving errors like: "failed to retrieve balance"
		Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

		// Deploy WEVMOS contract
		s.wevmosAddr, err = s.factory.DeployContract(
			s.keyring.GetPrivKey(erc20Deployer),
			evmtypes.EvmTxArgs{},
			testfactory.ContractDeploymentData{Contract: contracts.WEVMOSContract},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy wevmos contract")
		// Send WEVMOS to account
		_, err = s.factory.ExecuteEthTx(
			s.keyring.GetPrivKey(0),
			evmtypes.EvmTxArgs{
				To:     &s.wevmosAddr,
				Amount: sentWEVMOS.BigInt(),
				// FIXME: the gas simulation is not working correctly - otherwise results in out of gas
				GasLimit: 100_000,
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deposit to wevmos")

		s.network.App.Erc20Keeper.DeleteSTRv2Address(s.network.GetContext(), s.keyring.GetKey(0).AccAddr)
		s.network.App.Erc20Keeper.DeleteSTRv2Address(s.network.GetContext(), s.keyring.GetKey(2).AccAddr)
	})

	When("sending an EVM transaction", func() {
		Context("which interacts with a registered native token pair ERC-20 contract", func() {
			Context("in a direct call to the token pair contract", func() {
				It("should add the from and to addresses to the store if it is not already stored", func() {
					sender := s.keyring.GetKey(0)
					receiver := s.keyring.GetKey(2)

					senderAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")
					receiverAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), receiver.AccAddr)
					Expect(receiverAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")

					_, err := s.factory.ExecuteContractCall(
						sender.Priv,
						evmtypes.EvmTxArgs{
							To:       &s.wevmosAddr,
							GasLimit: 100_000,
						},
						testfactory.CallArgs{
							ContractABI: contracts.WEVMOSContract.ABI,
							MethodName:  "transfer",
							Args: []interface{}{
								receiver.Addr,
								transferAmount,
							},
						},
					)
					Expect(err).ToNot(HaveOccurred(), "failed to transfer tokens of Cosmos native ERC-20 token pair")

					Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

					senderAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeTrue(), "expected address to be stored")
					receiverAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(receiverAddrTracked).To(BeTrue(), "expected address to be stored")
				})
				It("should not fail if the addresses are already stored", func() {
					sender := s.keyring.GetKey(0)
					receiver := s.keyring.GetKey(2)

					senderAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")
					receiverAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), receiver.AccAddr)
					Expect(receiverAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")

					s.network.App.Erc20Keeper.SetSTRv2Address(s.network.GetContext(), s.keyring.GetKey(0).AccAddr)
					s.network.App.Erc20Keeper.SetSTRv2Address(s.network.GetContext(), s.keyring.GetKey(2).AccAddr)

					_, err := s.factory.ExecuteContractCall(
						sender.Priv,
						evmtypes.EvmTxArgs{
							To: &s.wevmosAddr,
						},
						testfactory.CallArgs{
							ContractABI: contracts.WEVMOSContract.ABI,
							MethodName:  "transfer",
							Args: []interface{}{
								receiver.Addr,
								transferAmount,
							},
						},
					)
					Expect(err).ToNot(HaveOccurred(), "failed to transfer tokens of Cosmos native ERC-20 token pair")

					Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

					senderAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(senderAddrTracked).To(BeTrue(), "expected address to be stored")
					receiverAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.AccAddr)
					Expect(receiverAddrTracked).To(BeTrue(), "expected address to be stored")
				})
			})
		})
	})
})
