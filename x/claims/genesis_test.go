package claims_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"
	simapp "github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/x/claims"
	"github.com/tharsis/evmos/x/claims/types"
)

var (
	now         = time.Now().UTC()
	acc1        = sdk.AccAddress([]byte("addr1---------------"))
	acc2        = sdk.AccAddress([]byte("addr2---------------"))
	testGenesis = types.GenesisState{
		Params: types.Params{
			AirdropStartTime:   now,
			DurationUntilDecay: types.DefaultDurationUntilDecay,
			DurationOfDecay:    types.DefaultDurationOfDecay,
			ClaimDenom:         types.DefaultClaimDenom, // aevmos
		},
		ClaimRecords: []types.ClaimRecordAddress{
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
	feemarketGenesis.BaseFee = sdk.NewInt(feemarketGenesis.Params.InitialBaseFee)

	app := simapp.Setup(false, feemarketGenesis)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis

	coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(10400)))
	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, coins)
	require.NoError(t, err)

	claims.InitGenesis(ctx, app.ClaimsKeeper, genesis)

	params := app.ClaimsKeeper.GetParams(ctx)
	require.Equal(t, params, genesis.Params)

	claimRecords := app.ClaimsKeeper.GetClaimRecords(ctx)
	require.Equal(t, claimRecords, genesis.ClaimRecords)
}

func TestClaimExportGenesis(t *testing.T) {
	// setup feemarketGenesis params
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 1
	feemarketGenesis.Params.NoBaseFee = false
	feemarketGenesis.BaseFee = sdk.NewInt(feemarketGenesis.Params.InitialBaseFee)

	app := simapp.Setup(false, feemarketGenesis)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis

	coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(10400)))
	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, coins)
	require.NoError(t, err)

	claims.InitGenesis(ctx, app.ClaimsKeeper, genesis)

	claimRecord, found := app.ClaimsKeeper.GetClaimRecord(ctx, acc2)
	require.True(t, found)
	require.Equal(t, claimRecord, types.ClaimRecord{
		InitialClaimableAmount: sdk.NewInt(400),
		ActionsCompleted:       []bool{false, false, false, false},
	})

	claimableAmount := app.ClaimsKeeper.GetClaimableAmountForAction(ctx, acc2, claimRecord, types.ActionIBCTransfer, genesis.Params)
	require.Equal(t, claimableAmount, sdk.NewInt(100))

	genesisExported := claims.ExportGenesis(ctx, app.ClaimsKeeper)
	require.Equal(t, genesisExported.Params, genesis.Params)
	require.Equal(t, genesisExported.ClaimRecords, []types.ClaimRecordAddress{
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
