package v17_test

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v16/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
			// TODO: Check why there's more supply than expected? Because of vals?
			Expect(supply.IsGTE(expEvmosSupply)).To(BeTrue(), "expected different evmos supply")

			for _, denom := range CoinDenoms {
				supply = network.App.BankKeeper.GetSupply(network.GetContext(), denom)
				fmt.Printf("supply for %q: %s\n", denom, supply.String())
				Expect(supply.IsZero()).To(BeFalse(), "supply for %s is zero", denom)
			}
		})
	})

	Context("exporting the genesis state", Ordered, func() {
		It("should run without errors", func() {
			var (
				jailedVals      []string
				modulesToExport []string // passing empty slice will default to all modules
			)

			exportedState, err := network.App.ExportAppStateAndValidators(
				true,
				jailedVals,
				modulesToExport,
			)
			Expect(err).To(BeNil(), "failed to export genesis state")
			Expect(exportedState).ToNot(BeNil(), "exported state is nil")

			fmt.Println("exported state", exportedState)
		})

	})
})
