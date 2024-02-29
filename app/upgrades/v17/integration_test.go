package v17_test

import (
	"fmt"
	cmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v16/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"testing"
)

func TestCreateDummyGenesis(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Creating Dummy Genesis")
}

const nKeys = 10

var _ = Describe("creating a dummy genesis state", Ordered, func() {
	var (
		keyring testkeyring.Keyring
		network *testnetwork.UnitTestNetwork
	)

	Context("creating the network with a custom genesis state", Ordered, func() {
		It("should run without errors", func() {
			keyring = testkeyring.New(nKeys)
			genesisBalances := createGenesisBalances(keyring)
			fmt.Println("genesis balances", genesisBalances)
			network = testnetwork.NewUnitTestNetwork(
				testnetwork.WithBalances(genesisBalances...),
			)
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

		//It("should have a balance for validators", func() {
		//	for i, val := range network.App.StakingKeeper.GetAllValidators(network.GetContext()) {
		//		fmt.Printf("validator %d: %s\n", i, val.OperatorAddress)
		//		balance := network.App.BankKeeper.GetBalance(network.GetContext(), val.GetOperator().Bytes(), utils.BaseDenom)
		//		Expect(balance.IsPositive()).To(BeTrue(), "validator %s has no balance", val.GetOperator())
		//		fmt.Println("validator balance", balance)
		//	}
		//})
	})

	Context("exporting the genesis state", Ordered, func() {
		It("should run without errors", func() {
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
