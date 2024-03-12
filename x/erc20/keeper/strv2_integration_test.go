// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/contracts"
	commonfactory "github.com/evmos/evmos/v16/testutil/integration/common/factory"
	testfactory "github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"math/big"
	"testing"
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

var _ = Describe("STRv2 Tracking", func() {
	var (
		s *STRv2TrackingSuite

		nativeCoinERC20Addr   common.Address
		registeredERC20Addr   common.Address
		unregisteredERC20Addr common.Address
	)

	BeforeEach(func() {
		s = SetupTestWithIBCCoinsInGenesis()
		Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

		deployer := s.keyring.GetKey(deployerIdx)

		// ------------------------------------------------------------------
		// Register the native IBC coin
		ibcCoinMetaData := banktypes.Metadata{
			Description: "The native IBC coin",
			Base:        nativeIBCCoinDenom,
			Display:     nativeIBCCoinDenom,
			DenomUnits: []*banktypes.DenomUnit{
				{Denom: nativeIBCCoinDenom, Exponent: 0},
				{Denom: "u" + nativeIBCCoinDenom, Exponent: 6},
			},
		}

		ibcNativeTokenPair, err := s.network.App.Erc20Keeper.RegisterCoin(s.network.GetContext(), ibcCoinMetaData)
		Expect(err).ToNot(HaveOccurred(), "failed to register native IBC coin")
		nativeCoinERC20Addr = common.HexToAddress(ibcNativeTokenPair.Erc20Address)

		Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

		// Convert a balance for the deployer
		msgConvertCoin := erc20types.MsgConvertCoin{
			Sender:   deployer.AccAddr.String(),
			Receiver: deployer.Addr.String(),
			Coin:     sdk.NewCoin(nativeIBCCoinDenom, convertAmount),
		}
		_, err = s.factory.ExecuteCosmosTx(deployer.Priv, commonfactory.CosmosTxArgs{
			Msgs: []sdk.Msg{&msgConvertCoin},
		})
		Expect(err).ToNot(HaveOccurred(), "failed to convert native IBC coin")

		// ------------------------------------------------------------------
		// Register an ERC-20 token pair
		registeredERC20Addr, err = s.DeployERC20Contract(deployer, ERC20ConstructorArgs{
			Name:     "TestToken",
			Symbol:   "TTK",
			Decimals: 18,
		})
		Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC-20 contract")

		_, err = s.network.App.Erc20Keeper.RegisterERC20(s.network.GetContext(), registeredERC20Addr)
		Expect(err).ToNot(HaveOccurred(), "failed to register token pair")

		Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

		// ------------------------------------------------------------------
		// Deploy an unregistered ERC-20 contract
		unregisteredERC20Addr, err = s.DeployERC20Contract(deployer, ERC20ConstructorArgs{
			Name:     "UnregisteredToken",
			Symbol:   "URT",
			Decimals: 18,
		})
		Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC-20 contract")

		Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")
	})

	When("sending an EVM transaction", func() {
		Context("which interacts with a registered native token pair ERC-20 contract", func() {
			Context("in a direct call to the token pair contract", func() {
				It("should add the address to the store if it is not already stored", func() {
					sender := s.keyring.GetKey(0)
					receiver := s.keyring.GetKey(2)

					senderAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.Addr)
					Expect(senderAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")
					receiverAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), receiver.Addr)
					Expect(receiverAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")

					_, err := s.factory.ExecuteContractCall(
						sender.Priv,
						evmtypes.EvmTxArgs{
							To: &nativeCoinERC20Addr,
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

					senderAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.Addr)
					Expect(senderAddrTracked).To(BeTrue(), "expected address to be stored")
					receiverAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.Addr)
					Expect(receiverAddrTracked).To(BeTrue(), "expected address to be stored")
				})

				It("should not fail if the address is already stored", func() {
					sender := s.keyring.GetKey(0)
					receiver := s.keyring.GetKey(2)

					senderAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.Addr)
					Expect(senderAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")
					receiverAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), receiver.Addr)
					Expect(receiverAddrTracked).To(BeFalse(), "expected address not to be stored before conversion")

					_, err := s.factory.ExecuteContractCall(
						sender.Priv,
						evmtypes.EvmTxArgs{
							To: &nativeCoinERC20Addr,
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
							To: &nativeCoinERC20Addr,
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

					senderAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.Addr)
					Expect(senderAddrTracked).To(BeTrue(), "expected address to be still stored")
					receiverAddrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), receiver.Addr)
					Expect(receiverAddrTracked).To(BeTrue(), "expected address to be still stored")
				})
			})

			Context("in a call to the token pair contract from another contract", func() {
				It("should add the address to the store if it is not already stored", func() {
					Expect(true).To(BeFalse(), "not implemented")
				})

				It("should not fail if the address is already stored", func() {
					Expect(true).To(BeFalse(), "not implemented")
				})
			})
		})

		Context("which interacts with a registered non-native token pair ERC-20 contract", func() {
			It("should not add the address to the store", func() {
				deployer := s.keyring.GetKey(deployerIdx)

				addrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), deployer.Addr)
				Expect(addrTracked).To(BeFalse(), "expected address not to be stored before conversion")

				_, err := s.factory.ExecuteContractCall(
					deployer.Priv,
					evmtypes.EvmTxArgs{
						To: &registeredERC20Addr,
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

				addrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), deployer.Addr)
				Expect(addrTracked).To(BeFalse(), "expected address to not be stored")
			})
		})

		Context("which interacts with an unregistered ERC-20 contract", func() {
			It("should not add the address to the store", func() {
				deployer := s.keyring.GetKey(0)

				_, err := s.factory.ExecuteContractCall(
					deployer.Priv,
					evmtypes.EvmTxArgs{
						To: &unregisteredERC20Addr,
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
				Expect(err).ToNot(HaveOccurred(), "failed to mint tokens for non-registered ERC-20 contract")

				deployerAddrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), deployer.Addr)
				Expect(deployerAddrTracked).To(BeFalse(), "expected address to not be stored")
			})
		})
	})

	When("when receiving an incoming IBC transfer", func() {
		Context("for a registered IBC asset", func() {
			It("should add the address to the store if it is not already stored", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})

			It("should not fail if the address is already stored", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})
		})

		Context("for an unregistered IBC asset", func() {
			It("should not add the address to the store", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})
		})
	})

	When("sending an IBC transfer", func() {
		Context("for a registered IBC asset", func() {
			It("should add the address to the store if it is not already stored", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})

			It("should not fail if the address is already stored", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})
		})

		Context("for an unregistered IBC asset", func() {
			It("should not add the address to the store", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})
		})
	})

	When("manually converting", func() {
		Context("a registered coin into its ERC-20 representation", func() {
			It("should add the address to the store if it is not already stored", func() {
				sender := s.keyring.GetKey(1)

				addrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.Addr)
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

				addrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.Addr)
				Expect(addrTracked).To(BeTrue(), "expected address to be stored")
			})

			It("should not fail if the address is already stored", func() {
				sender := s.keyring.GetKey(1)

				addrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.Addr)
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

				addrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.Addr)
				Expect(addrTracked).To(BeTrue(), "expected address to be still stored")
			})
		})

		Context("a registered ERC-20 representation into its native coin", func() {
			It("should add the address to the store if it is not already stored", func() {
				sender := s.keyring.GetKey(deployerIdx)

				// TODO: this will probably be wrong, because the address is stored after the conversion
				// To handle this better we might need to manually create the genesis
				addrTracked := s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.Addr)
				Expect(addrTracked).To(BeFalse(), "expected address not to be stored before conversion")

				_, err := s.factory.ExecuteCosmosTx(
					sender.Priv,
					commonfactory.CosmosTxArgs{
						Msgs: []sdk.Msg{
							&erc20types.MsgConvertERC20{
								Sender:   sender.Addr.String(),
								Receiver: sender.AccAddr.String(),
								Amount:   convertAmount,
							},
						},
					},
				)
				Expect(err).ToNot(HaveOccurred(), "failed to convert native IBC coin")

				Expect(s.network.NextBlock()).To(BeNil(), "failed to advance block")

				addrTracked = s.network.App.Erc20Keeper.HasSTRv2Address(s.network.GetContext(), sender.Addr)
				Expect(addrTracked).To(BeTrue(), "expected address to be stored")
			})
		})
	})
})
