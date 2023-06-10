package interchaintest

import (
	"context"
	"fmt"
	"testing"
	"time"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v6"
	"github.com/strangelove-ventures/interchaintest/v6/ibc"
	"github.com/strangelove-ventures/interchaintest/v6/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestInterchain(t *testing.T) {
	// TODO: add short testing mode
	if testing.Short() {
		t.Skip("skipping interchain tests in short mode.")
	}

	// allow parallel testing
	t.Parallel()

	// Set up context
	ctx := context.Background()

	// Set number of nodes
	numFns := 0
	numVals := 1

	// Set up the chain factory
	chainFactory := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "gaia",
			Version:       "v7.0.1",
			NumFullNodes:  &numFns,
			NumValidators: &numVals,
		},
		{
			ChainConfig: ibc.ChainConfig{
				Type:     "cosmos",
				Name:     "evmos",
				ChainID:  "evmos_9000-1",
				CoinType: "60",
				Images: []ibc.DockerImage{
					{
						Repository: "tharsishq/evmos",
						Version:    "latest",
						UidGid:     "1025:1025",
					},
				},
				Bin:            "evmosd",
				Bech32Prefix:   "evmos",
				Denom:          "aevmos",
				GasPrices:      "0.0aevmos",
				GasAdjustment:  1.3,
				EncodingConfig: nil,
				ExtraCodecs:    []string{"ethermint"},
			},
			NumFullNodes:  &numFns,
			NumValidators: &numVals,
		},
	})

	chains, err := chainFactory.Chains(t.Name())
	require.NoError(t, err)
	//Expect(err).To(BeNil(), "expected no error when creating chains")
	gaia, evmos := chains[0], chains[1]

	// Relayer factory -> we are using the Cosmos relayer here; could also use ibc.Hermes
	client, network := interchaintest.DockerSetup(t)
	r := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t)).Build(
		t, client, network,
	)

	// Prep interchain setup
	const ibcPath = "gaia-evmos"
	ic := interchaintest.NewInterchain().
		AddChain(gaia).
		AddChain(evmos).
		AddRelayer(r, "Cosmos relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  gaia,
			Chain2:  evmos,
			Relayer: r,
			Path:    ibcPath,
		})

	// Log location
	f, err := interchaintest.CreateLogFile(fmt.Sprintf("%d.json", time.Now().Unix()))
	require.NoError(t, err)
	//Expect(err).To(BeNil(), "expected no error when creating log file")

	// Reporter/logs
	rep := testreporter.NewReporter(f)
	eRep := rep.RelayerExecReporter(t)

	// Build interchain
	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
		SkipPathCreation:  false,
	}),
	)

	// Create and Fund User Wallets
	fundAmount := int64(10_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", fundAmount, gaia, evmos)
	gaiaUser := users[0]
	evmosUser := users[1]

	gaiaUserBalInitial, err := gaia.GetBalance(ctx, gaiaUser.FormattedAddress(), gaia.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, fundAmount, gaiaUserBalInitial)

	// Get Channel ID
	gaiaChannelInfo, err := r.GetChannels(ctx, eRep, gaia.Config().ChainID)
	require.NoError(t, err)
	gaiaChannelID := gaiaChannelInfo[0].ChannelID

	evmosChannelInfo, err := r.GetChannels(ctx, eRep, evmos.Config().ChainID)
	require.NoError(t, err)
	evmosChannelID := evmosChannelInfo[0].ChannelID

	// Send Transaction
	amountToSend := int64(1_000_000)
	dstAddress := evmosUser.FormattedAddress()
	transfer := ibc.WalletAmount{
		Address: dstAddress,
		Denom:   gaia.Config().Denom,
		Amount:  amountToSend,
	}
	tx, err := gaia.SendIBCTransfer(ctx, gaiaChannelID, gaiaUser.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)
	require.NoError(t, tx.Validate())

	// relay packets and acknowledgements
	require.NoError(t, r.FlushPackets(ctx, eRep, ibcPath, evmosChannelID))
	require.NoError(t, r.FlushAcknowledgements(ctx, eRep, ibcPath, gaiaChannelID))

	// test source wallet has decreased funds
	expectedBal := gaiaUserBalInitial - amountToSend
	gaiaUserBalNew, err := gaia.GetBalance(ctx, gaiaUser.FormattedAddress(), gaia.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, expectedBal, gaiaUserBalNew)

	// Trace IBC Denom
	srcDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", gaiaChannelID, gaia.Config().Denom))
	dstIbcDenom := srcDenomTrace.IBCDenom()

	// Test destination wallet has increased funds
	osmosUserBalNew, err := evmos.GetBalance(ctx, evmosUser.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.Equal(t, amountToSend, osmosUserBalNew)
}
