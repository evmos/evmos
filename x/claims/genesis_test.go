package claims_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"
	simapp "github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/x/claims"
	"github.com/tharsis/evmos/x/claims/types"
	inflationtypes "github.com/tharsis/evmos/x/inflation/types"
)

var (
	now         = time.Now().UTC()
	acc1, _     = sdk.AccAddressFromBech32("evmos1qxx0fdsmruzuar2fay88lfw6sce6emamyu2s8h4d")
	acc2, _     = sdk.AccAddressFromBech32("evmos1nsrs4t7dngkdltehkm3p6n8dp22sz3mct9uhc8")
	testGenesis = types.GenesisState{
		Params: types.Params{
			AirdropStartTime:   now,
			DurationUntilDecay: types.DefaultDurationUntilDecay,
			DurationOfDecay:    types.DefaultDurationOfDecay,
			ClaimsDenom:        types.DefaultClaimsDenom, // aevmos
		},
		ClaimsRecords: []types.ClaimsRecordAddress{
			{
				Address:                acc1.String(),
				InitialClaimableAmount: sdk.NewInt(10000),
				ActionsCompleted:       []bool{true, false, true, true},
			},
			{
				Address:                acc2.String(),
				InitialClaimableAmount: sdk.NewInt(400),
				ActionsCompleted:       []bool{false, false, false, false},
			},
		},
	}
)

func TestClaimInitGenesis(t *testing.T) {
	// setup feemarketGenesis params
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 1
	feemarketGenesis.Params.NoBaseFee = false

	app := simapp.Setup(false, feemarketGenesis)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis

	coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(10400)))
	err := app.BankKeeper.MintCoins(ctx, inflationtypes.ModuleName, coins)
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, inflationtypes.ModuleName, types.ModuleName, coins)
	require.NoError(t, err)

	claims.InitGenesis(ctx, app.ClaimsKeeper, genesis)

	params := app.ClaimsKeeper.GetParams(ctx)
	require.Equal(t, params, genesis.Params)

	claimsRecords := app.ClaimsKeeper.GetClaimsRecords(ctx)
	require.Equal(t, claimsRecords, genesis.ClaimsRecords)
}

func TestClaimExportGenesis(t *testing.T) {
	// setup feemarketGenesis params
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 1
	feemarketGenesis.Params.NoBaseFee = false

	app := simapp.Setup(false, feemarketGenesis)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis

	coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(10400)))
	err := app.BankKeeper.MintCoins(ctx, inflationtypes.ModuleName, coins)
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, inflationtypes.ModuleName, types.ModuleName, coins)
	require.NoError(t, err)

	claims.InitGenesis(ctx, app.ClaimsKeeper, genesis)

	claimsRecord, found := app.ClaimsKeeper.GetClaimsRecord(ctx, acc2)
	require.True(t, found)
	require.Equal(t, claimsRecord, types.ClaimsRecord{
		InitialClaimableAmount: sdk.NewInt(400),
		ActionsCompleted:       []bool{false, false, false, false},
	})

	claimableAmount := app.ClaimsKeeper.GetClaimableAmountForAction(ctx, claimsRecord, types.ActionIBCTransfer, genesis.Params)
	require.Equal(t, claimableAmount, sdk.NewInt(100))

	genesisExported := claims.ExportGenesis(ctx, app.ClaimsKeeper)
	require.Equal(t, genesisExported.Params, genesis.Params)
	require.Equal(t, genesisExported.ClaimsRecords, []types.ClaimsRecordAddress{
		{
			Address:                acc1.String(),
			InitialClaimableAmount: sdk.NewInt(10000),
			ActionsCompleted:       []bool{true, false, true, true},
		},
		{
			Address:                acc2.String(),
			InitialClaimableAmount: sdk.NewInt(400),
			ActionsCompleted:       []bool{false, false, false, false},
		},
	})
}
