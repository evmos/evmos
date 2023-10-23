// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package erc20_test

import (

	//nolint:revive // dot imports are fine for Ginkgo
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v15/precompiles/authorization"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	"github.com/evmos/evmos/v15/testutil/integration/factory"
	"github.com/evmos/evmos/v15/testutil/integration/grpc"
	testkeyring "github.com/evmos/evmos/v15/testutil/integration/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/network"
	integrationutils "github.com/evmos/evmos/v15/testutil/integration/utils"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

type IntegrationTestSuite struct {
	network           network.Network
	factory           factory.TxFactory
	keyring           testkeyring.Keyring
	precompile        *erc20.Precompile
	precompileAddress common.Address
}

var _ = Describe("Calling ERC20 precompile directly", Label("ERC20 Precompile"), Ordered, func() {
	var s *IntegrationTestSuite

	BeforeAll(func() {
		keyring := testkeyring.New(2)
		integrationNetwork := network.New(
			network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		)
		grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
		txFactory := factory.New(integrationNetwork, grpcHandler)

		s = &IntegrationTestSuite{
			network: integrationNetwork,
			factory: txFactory,
			keyring: keyring,
		}
	})

	AfterEach(func() {
		// Start each test with a fresh block
		err := s.network.NextBlock()
		Expect(err).To(BeNil())
	})

	Describe("name method", func() {
		When("token is registered", func() {
			It("should return the name from the metadata", func() {})
		})
		When("token is unregistered", func() {
			It("should build the name from the denom trace base denomination", func() {})
		})
	})

	Describe("symbol method", func() {
		When("token is registered", func() {
			It("should return the symbol from the metadata", func() {})
		})
		When("token is unregistered", func() {
			It("should build the symbol from the denom trace base denomination", func() {})
		})
	})

	Describe("decimals method", func() {
		When("token is registered", func() {
			It("should return the decimals from the metadata", func() {})
		})
		When("token is unregistered", func() {
			It("should assume the decimals from the denom trace base denomination's prefix", func() {})
		})
	})

	Describe("With a registered Native Coin", func() {
		var denom string
		senderPriv := s.keyring.GetPrivKey(0)
		address := s.keyring.GetAddr(0)
		accAddress := s.keyring.GetAccAddr(0)

		BeforeEach(func() {
			// TODO: Transfer tokens via IBC to generate some supply
		})
		It("should return the name", func() {
			nameTxArgs := evmtypes.EvmTxArgs{
				To: &s.precompileAddress,
			}
			nameArgs := factory.CallArgs{
				ContractABI: s.precompile.ABI,
				MethodName:  erc20.NameMethod,
				Args:        []interface{}{},
			}

			metadataRes, err := s.network.GetBankClient().DenomMetadata(
				sdk.WrapSDKContext(s.network.GetContext()),
				&banktypes.QueryDenomMetadataRequest{Denom: denom},
			)
			Expect(err).To(BeNil(), "failed to get metadata")
			name := metadataRes.Metadata.Name

			nameRes, err := s.factory.ExecuteContractCall(senderPriv, nameTxArgs, nameArgs)
			Expect(err).To(BeNil())
			Expect(nameRes.IsOK()).To(Equal(true), "transaction should have succeeded", nameRes.GetLog())

			var nameResponse string
			err = integrationutils.DecodeContractCallResponse(&nameResponse, nameArgs, nameRes)
			Expect(err).To(BeNil())
			Expect(nameResponse).To(Equal(name))
		})
		It("should return the symbol", func() {
			symbolTxArgs := evmtypes.EvmTxArgs{
				To: &s.precompileAddress,
			}
			symbolArgs := factory.CallArgs{
				ContractABI: s.precompile.ABI,
				MethodName:  erc20.SymbolMethod,
				Args:        []interface{}{},
			}

			metadataRes, err := s.network.GetBankClient().DenomMetadata(
				sdk.WrapSDKContext(s.network.GetContext()),
				&banktypes.QueryDenomMetadataRequest{Denom: denom},
			)
			Expect(err).To(BeNil(), "failed to get metadata")
			symbol := metadataRes.Metadata.Symbol

			symbolRes, err := s.factory.ExecuteContractCall(senderPriv, symbolTxArgs, symbolArgs)
			Expect(err).To(BeNil())
			Expect(symbolRes.IsOK()).To(Equal(true), "transaction should have succeeded", symbolRes.GetLog())

			var symbolResponse string
			err = integrationutils.DecodeContractCallResponse(&symbolResponse, symbolArgs, symbolRes)
			Expect(err).To(BeNil())
			Expect(symbolResponse).To(Equal(symbol))
		})
		It("should return the decimals", func() {
			decimalsTxArgs := evmtypes.EvmTxArgs{
				To: &s.precompileAddress,
			}
			decimalsArgs := factory.CallArgs{
				ContractABI: s.precompile.ABI,
				MethodName:  erc20.DecimalsMethod,
				Args:        []interface{}{},
			}

			metadataRes, err := s.network.GetBankClient().DenomMetadata(
				sdk.WrapSDKContext(s.network.GetContext()),
				&banktypes.QueryDenomMetadataRequest{Denom: denom},
			)
			Expect(err).To(BeNil(), "failed to get metadata")
			var decimals uint32
			for i := len(metadataRes.Metadata.DenomUnits); i >= 0; i-- {
				if metadataRes.Metadata.DenomUnits[i].Denom == metadataRes.Metadata.Display {
					decimals = metadataRes.Metadata.DenomUnits[i].Exponent
					break
				}
			}

			decimalsRes, err := s.factory.ExecuteContractCall(senderPriv, decimalsTxArgs, decimalsArgs)
			Expect(err).To(BeNil())
			Expect(decimalsRes.IsOK()).To(Equal(true), "transaction should have succeeded", decimalsRes.GetLog())

			var decimalsResponse uint8
			err = integrationutils.DecodeContractCallResponse(&decimalsResponse, decimalsArgs, decimalsRes)
			Expect(err).To(BeNil())
			Expect(decimalsResponse).To(Equal(uint8(decimals))) // gosec: nosec
		})
		It("should return the totalSupply", func() {
			totalSupplyTxArgs := evmtypes.EvmTxArgs{
				To: &s.precompileAddress,
			}
			totalSupplyArgs := factory.CallArgs{
				ContractABI: s.precompile.ABI,
				MethodName:  erc20.TotalSupplyMethod,
				Args:        []interface{}{},
			}

			supplyRes, err := s.network.GetBankClient().SupplyOf(
				sdk.WrapSDKContext(s.network.GetContext()),
				&banktypes.QuerySupplyOfRequest{Denom: denom},
			)
			Expect(err).To(BeNil(), "failed to get supply")
			supply := supplyRes.Amount.Amount.BigInt()

			totalSupplyRes, err := s.factory.ExecuteContractCall(senderPriv, totalSupplyTxArgs, totalSupplyArgs)
			Expect(err).To(BeNil())
			Expect(totalSupplyRes.IsOK()).To(Equal(true), "transaction should have succeeded", totalSupplyRes.GetLog())

			var totalSupplyResponse *big.Int
			err = integrationutils.DecodeContractCallResponse(&totalSupplyResponse, totalSupplyArgs, totalSupplyRes)
			Expect(err).To(BeNil())
			Expect(totalSupplyResponse).To(Equal(supply))
		})
		It("should return the balance of an account", func() {
			balanceOfTxArgs := evmtypes.EvmTxArgs{
				To: &s.precompileAddress,
			}
			balanceOfArgs := factory.CallArgs{
				ContractABI: s.precompile.ABI,
				MethodName:  erc20.BalanceOfMethod,
				Args:        []interface{}{address},
			}

			balanceRes, err := s.network.GetBankClient().Balance(
				sdk.WrapSDKContext(s.network.GetContext()),
				&banktypes.QueryBalanceRequest{Address: accAddress.String(), Denom: denom},
			)
			Expect(err).To(BeNil(), "failed to get supply")
			balanceAmount := balanceRes.Balance.Amount.BigInt()

			balanceOfRes, err := s.factory.ExecuteContractCall(senderPriv, balanceOfTxArgs, balanceOfArgs)
			Expect(err).To(BeNil())
			Expect(balanceOfRes.IsOK()).To(Equal(true), "transaction should have succeeded", balanceOfRes.GetLog())

			var balanceOfResponse *big.Int
			err = integrationutils.DecodeContractCallResponse(&balanceOfResponse, balanceOfArgs, balanceOfRes)
			Expect(err).To(BeNil())
			Expect(balanceOfResponse).To(Equal(balanceAmount))
		})
		It("should return the allowance of an account", func() {
			allowanceTxArgs := evmtypes.EvmTxArgs{
				To: &s.precompileAddress,
			}
			allowanceArgs := factory.CallArgs{
				ContractABI: s.precompile.ABI,
				MethodName:  authorization.AllowanceMethod,
				Args:        []interface{}{address}, // TODO: update allowance address
			}

			// granteeGrantsRes, err := s.network.GetAuthzClient().GranteeGrants(
			// 	sdk.WrapSDKContext(s.network.GetContext()),
			// 	&authz.QueryGranteeGrantsRequest{Grantee: accAddress.String()}, // TODO: Pagination?
			// )
			// Expect(err).To(BeNil(), "failed to get grant")

			// balanceAmount := granteeGrantsRes.Grants[0].UnpackInterfaces()
			allowanceRes, err := s.factory.ExecuteContractCall(senderPriv, allowanceTxArgs, allowanceArgs)
			Expect(err).To(BeNil())
			Expect(allowanceRes.IsOK()).To(Equal(true), "transaction should have succeeded", allowanceRes.GetLog())

			var allowanceResponse *big.Int
			err = integrationutils.DecodeContractCallResponse(&allowanceResponse, allowanceArgs, allowanceRes)
			Expect(err).To(BeNil())
			// Expect(allowanceResponse).To(Equal(balanceAmount))
		})
	})
})
