package v17_test

import (
	"fmt"
	"github.com/evmos/evmos/v16/x/erc20"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/evmos/evmos/v16/x/evm"

	cmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/contracts"
	testfactory "github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v16/utils"
	"github.com/evmos/evmos/v16/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCreateDummyGenesis(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Creating Dummy Genesis")
}

const (
	// nKeys is the number of keys to generate
	nKeys = 100_000
	// nTokenPairs is the number of token pairs to generate
	nTokenPairs = 15
)

var _ = Describe("creating a dummy genesis state", Ordered, func() {
	var (
		keyring testkeyring.Keyring
		network *testnetwork.UnitTestNetwork
		handler grpc.Handler
		factory testfactory.TxFactory
	)

	Context("creating the network with a custom genesis state", Ordered, func() {
		It("should run without errors", func() {
			keyring = testkeyring.New(nKeys)
			genesisBalances := createGenesisBalances(keyring)
			network = testnetwork.NewUnitTestNetwork(
				testnetwork.WithBalances(genesisBalances...),
			)
			handler = grpc.NewIntegrationHandler(network)
			factory = testfactory.New(network, handler)
		})

		It("should initialize the chain with a selection of tokens", func() {
			expEvmosSupply := sdk.Coin{
				Denom:  utils.BaseDenom,
				Amount: testnetwork.PrefundedAccountInitialBalance.MulRaw(nKeys),
			}
			supply := network.App.BankKeeper.GetSupply(network.GetContext(), utils.BaseDenom)
			// TODO: Check why there's more supply than expected? Because of the vals?
			Expect(supply.IsGTE(expEvmosSupply)).To(BeTrue(), "expected different evmos supply")

			for _, denom := range CoinDenoms {
				supply = network.App.BankKeeper.GetSupply(network.GetContext(), denom)
				Expect(supply.IsZero()).To(BeFalse(), "supply for %s is zero", denom)
			}
		})
	})

	Context("deploy the WEVMOS contract", Ordered, func() {
		It("should run without errors", func() {
			wevmosAddr, err := factory.DeployContract(
				keyring.GetPrivKey(0),
				evmtypes.EvmTxArgs{},
				testfactory.ContractDeploymentData{
					Contract: contracts.ERC20MinterBurnerDecimalsContract,
					ConstructorArgs: []interface{}{
						"WEVMOS",
						"WEVMOS",
						uint8(18),
					},
				},
			)
			Expect(err).To(BeNil(), "failed to deploy WEVMOS contract")
			fmt.Println("deployed WEVMOS contract at ", wevmosAddr)
		})
	})

	Context("registering the token pairs", Ordered, func() {
		It("should run without errors", func() {
			for _, denom := range CoinDenoms {
				coinMetadata := CreateFullMetadata(denom, strings.ToUpper(denom), denom)

				_, err := network.App.Erc20Keeper.RegisterCoin(network.GetContext(), coinMetadata)
				Expect(err).To(BeNil(), "failed to register token pair")
			}
		})

		It("should have the token pairs registered", func() {
			tokenPairs := network.App.Erc20Keeper.GetTokenPairs(network.GetContext())
			Expect(len(tokenPairs)).To(Equal(nTokenPairs), "unexpected number of token pairs")
		})

		It("should not have any ERC-20 supply", func() {
			abi := contracts.ERC20MinterBurnerDecimalsContract.ABI

			for _, denom := range CoinDenoms {
				tokenPairID := network.App.Erc20Keeper.GetTokenPairID(network.GetContext(), denom)
				tokenPair, found := network.App.Erc20Keeper.GetTokenPair(network.GetContext(), tokenPairID)
				Expect(found).To(BeTrue(), "failed to get token pair")

				addr := common.HexToAddress(tokenPair.Erc20Address)

				res, err := factory.ExecuteContractCall(
					keyring.GetPrivKey(0),
					evmtypes.EvmTxArgs{
						To: &addr,
					},
					testfactory.CallArgs{
						ContractABI: abi,
						MethodName:  "totalSupply",
					},
				)
				Expect(err).To(BeNil(), "failed to execute contract call")
				Expect(res).ToNot(BeNil(), "contract call result is nil")

				ethRes, err := evmtypes.DecodeTxResponse(res.Data)
				Expect(err).To(BeNil(), "failed to decode contract call result")

				var supply *big.Int
				err = abi.UnpackIntoInterface(&supply, "totalSupply", ethRes.Ret)
				Expect(err).To(BeNil(), "failed to unpack total supply")
				Expect(supply.Int64()).To(BeZero(), "supply for %s is not zero", denom)
			}
		})
	})

	Context("converting the token pair balances to the ERC-20 representation", Ordered, func() {
		It("should run without errors", func() {
			for _, key := range keyring.GetAllKeys() {
				for _, denom := range CoinDenoms {
					denomBalance := network.App.BankKeeper.GetBalance(network.GetContext(), key.AccAddr, denom)
					if denomBalance.IsZero() {
						continue
					}

					res, err := network.App.Erc20Keeper.ConvertCoin(network.GetContext(), &types.MsgConvertCoin{
						Coin:     sdk.Coin{Denom: denom, Amount: denomBalance.Amount},
						Receiver: key.Addr.String(),
						Sender:   key.AccAddr.String(),
					})
					if err != nil {
						fmt.Println("err not nil for ", key.AccAddr.String(), denom, denomBalance.Amount)
					}
					Expect(err).To(BeNil(), "failed to convert coin")
					Expect(res).ToNot(BeNil(), "failed to convert coin")
				}
			}
		})
	})

	Context("exporting the genesis state", Ordered, func() {
		It("should export the individual keeper states", func() {
			// - bank
			// - auth
			// - erc20
			// - evm
			// - (staking)
			// - (distribution)

			// TODO: refactor this with generics

			// Export bank state
			bankGenState := network.App.BankKeeper.ExportGenesis(network.GetContext())
			Expect(bankGenState).ToNot(BeNil(), "failed to export bank genesis state")
			out, err := bankGenState.Marshal()
			Expect(err).To(BeNil(), "failed to marshal bank genesis state")
			err = os.WriteFile("bank_gen_state.json", out, 0o600)
			Expect(err).To(BeNil(), "failed to write bank gen state to file")

			// Export auth state
			authGenState := network.App.AccountKeeper.ExportGenesis(network.GetContext())
			Expect(authGenState).ToNot(BeNil(), "failed to export auth genesis state")
			out, err = authGenState.Marshal()
			Expect(err).To(BeNil(), "failed to marshal auth genesis state")
			err = os.WriteFile("auth_gen_state.json", out, 0o600)
			Expect(err).To(BeNil(), "failed to write auth gen state to file")

			// Export erc20 state
			erc20GenState := erc20.ExportGenesis(network.GetContext(), network.App.Erc20Keeper)
			Expect(erc20GenState).ToNot(BeNil(), "failed to export erc20 genesis state")
			out, err = erc20GenState.Marshal()
			Expect(err).To(BeNil(), "failed to marshal erc20 genesis state")
			err = os.WriteFile("erc20_gen_state.json", out, 0o600)
			Expect(err).To(BeNil(), "failed to write erc20 gen state to file")

			// Export EVM state
			evmGenState := evm.ExportGenesis(network.GetContext(), network.App.EvmKeeper, network.App.AccountKeeper)
			Expect(evmGenState).ToNot(BeNil(), "failed to export evm genesis state")
			out, err = evmGenState.Marshal()
			Expect(err).To(BeNil(), "failed to marshal evm genesis state")
			err = os.WriteFile("evm_gen_state.json", out, 0o600)
			Expect(err).To(BeNil(), "failed to write evm gen state to file")
		})

		It("should import the exported EVM genesis state", func() {
			// Read gen state from file
			out, err := os.ReadFile("evm_gen_state.json")
			Expect(err).To(BeNil(), "failed to read evm gen state from file")

			// Unmarshal gen state
			var evmGenState evmtypes.GenesisState
			err = evmGenState.Unmarshal(out)
			Expect(err).To(BeNil(), "failed to unmarshal evm gen state")
			Expect(evmGenState).ToNot(BeNil(), "evm gen state is nil")
		})

		It("should export the whole state to a JSON file", func() {
			Skip("export not necessary right now")

			var (
				jailedVals      []string
				modulesToExport []string // passing empty slice will default to all modules
			)

			exportedState, err := network.App.ExportAppStateAndValidators(
				false, // no need to export for zero height
				jailedVals,
				modulesToExport,
			)
			Expect(err).To(BeNil(), "failed to export genesis state")
			Expect(exportedState).ToNot(BeNil(), "exported state is nil")
			Expect(exportedState.AppState).ToNot(BeNil(), "exported state is nil")
			Expect(exportedState.Height).ToNot(BeZero(), "exported height is zero")
			Expect(exportedState.ConsensusParams).ToNot(BeNil(), "exported consensus params are nil")

			exportedStateFile := "exported_genesis.json"

			// NOTE: The following logic to export the genesis state is copied from
			// the `ExportCmd`: https://github.com/evmos/cosmos-sdk/blob/v0.47.5-evmos.2/server/export.go#L83-L99
			genDoc := &cmtypes.GenesisDoc{
				AppState:      exportedState.AppState,
				ChainID:       network.GetChainID(),
				Validators:    exportedState.Validators,
				InitialHeight: exportedState.Height,
				ConsensusParams: &cmtypes.ConsensusParams{
					Block: cmtypes.BlockParams{
						MaxBytes: exportedState.ConsensusParams.Block.MaxBytes,
						MaxGas:   exportedState.ConsensusParams.Block.MaxGas,
					},
					Evidence: cmtypes.EvidenceParams{
						MaxAgeNumBlocks: exportedState.ConsensusParams.Evidence.MaxAgeNumBlocks,
						MaxAgeDuration:  exportedState.ConsensusParams.Evidence.MaxAgeDuration,
						MaxBytes:        exportedState.ConsensusParams.Evidence.MaxBytes,
					},
					Validator: cmtypes.ValidatorParams{
						PubKeyTypes: exportedState.ConsensusParams.Validator.PubKeyTypes,
					},
				},
			}

			if _, err := os.Stat(exportedStateFile); err != nil && !os.IsNotExist(err) {
				panic("genesis file already exists")
			}

			// NOTE: This approach to exporting the genesis file is copied from the
			// `InitCmd`: https://github.com/evmos/evmos/blob/c159c98e191f95f334e606d59ad384e47c325258/cmd/evmosd/init.go#L157-L159
			err = genutil.ExportGenesisFile(genDoc, exportedStateFile)
			Expect(err).To(BeNil(), "failed to export genesis file")

			_, err = os.Stat(exportedStateFile)
			Expect(err).To(BeNil(), "exported genesis file does not exist after exporting")
		})
	})
})
