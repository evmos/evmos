package interchain_test

import (
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/evmos/evmos/v14/encoding"
	"github.com/evmos/evmos/v14/app"
	"github.com/skip-mev/block-sdk/tests/integration"
	"github.com/stretchr/testify/suite"
	ictestutil "github.com/strangelove-ventures/interchaintest/v7/testutil"
	evmoskeyring "github.com/evmos/evmos/v14/crypto/keyring"
)

var (
	// config params
	numValidators = 4
	numFullNodes = 0
	denom         = "aevmos"

	image = ibc.DockerImage{
		Repository: "tharsishq/evmos",
		Version:    "41d3492",
		UidGid:     "1000:1000",
	}
	noHostMount    = false
	gasAdjustment  = float64(2.0)
	encodingConfig = MakeEncodingConfig()

	genesisKV = []cosmos.GenesisKV{
		{
			Key:   "app_state.auction.params.max_bundle_size",
			Value: 3,
		},
		{
			Key:   "app_state.auction.params.reserve_fee.denom",
			Value: denom,
		},
		{
			Key:   "app_state.auction.params.reserve_fee.amount",
			Value: "1",
		},
		{
			Key:   "app_state.auction.params.min_bid_increment.denom",
			Value: denom,
		},
		{
			Key:  "app_state.staking.params.bond_denom",
			Value: denom,
		},
		{
			Key: "app_state.crisis.constant_fee.denom",
			Value: denom,
		},
		{
			Key: "app_state.feemarket.params.no_base_fee",
			Value: true,
		},
		{
			Key: "consensus_params.max_gas",
			Value: -1,
		},
	}

	initCoins = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	// interchain specification
	spec = &interchaintest.ChainSpec{
		Name:          "evmos",
		NumValidators: &numValidators,
		NumFullNodes:  &numFullNodes,
		Version:       "41d3492",
		NoHostMount:   &noHostMount,
		ChainConfig: ibc.ChainConfig{
			EncodingConfig: encodingConfig,
			Images: []ibc.DockerImage{
				image,
			},
			Type:                   "cosmos",
			Denom:                  denom,
			ChainID:                "evmos_9001-1",
			Bin:                    "evmosd",
			Bech32Prefix:           "evmos",
			CoinType:               "118",
			GasAdjustment:          gasAdjustment,
			GasPrices:              fmt.Sprintf("0%s", denom),
			TrustingPeriod:         "48h",
			NoHostMount:            noHostMount,
			ModifyGenesis:          cosmos.ModifyGenesis(genesisKV),
			ModifyGenesisAmounts: func() (sdk.Coin, sdk.Coin) {
				return sdk.NewCoin(denom, sdk.NewIntFromBigInt(initCoins)), sdk.NewCoin(denom, sdk.NewIntFromBigInt(initCoins))
			},
			ConfigFileOverrides: map[string]any{
				"config/client.toml" : ictestutil.Toml{
					"chain-id": "evmos_9001-1",
				},
				"config/app.toml" : ictestutil.Toml{
					"grpc" : ictestutil.Toml{
						"enable": "true",
					},
				},
			},
			UsingChainIDFlagCLI: true,
		},
	}
)

func MakeEncodingConfig() *testutil.TestEncodingConfig {
	ec := encoding.MakeConfig(app.ModuleBasics)
	return &testutil.TestEncodingConfig{
		InterfaceRegistry: ec.InterfaceRegistry,
		Codec: ec.Codec,
		TxConfig: ec.TxConfig,
		Amino: ec.Amino,
	}
}

func TestBlockSDKSuite(t *testing.T) {
	s := integration.NewIntegrationTestSuiteFromSpec(spec)
	s.WithDenom(denom)
	s.WithKeyringOptions(encodingConfig.Codec, evmoskeyring.Option())
	suite.Run(t, s)
}