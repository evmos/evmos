package claim_test

// import (
// 	"testing"
// 	"time"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/stretchr/testify/require"
// 	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
// 	simapp "github.com/tharsis/evmos/app"
// 	"github.com/tharsis/evmos/x/claim"
// 	"github.com/tharsis/evmos/x/claim/types"
// )

// var (
// 	now         = time.Now().UTC()
// 	acc1        = sdk.AccAddress([]byte("addr1---------------"))
// 	acc2        = sdk.AccAddress([]byte("addr2---------------"))
// 	testGenesis = types.GenesisState{
// 		ModuleAccountBalance: sdk.NewInt64Coin(types.DefaultClaimDenom, 750000000),
// 		Params: types.Params{
// 			AirdropStartTime:   now,
// 			DurationUntilDecay: types.DefaultDurationUntilDecay,
// 			DurationOfDecay:    types.DefaultDurationOfDecay,
// 			ClaimDenom:         types.DefaultClaimDenom, // uosmo
// 		},
// 		ClaimRecords: []types.ClaimRecord{
// 			{
// 				Address:                acc1.String(),
// 				InitialClaimableAmount: sdk.Coins{sdk.NewInt64Coin(types.DefaultClaimDenom, 1000000000)},
// 				ActionCompleted:        []bool{true, false, true, true},
// 			},
// 			{
// 				Address:                acc2.String(),
// 				InitialClaimableAmount: sdk.Coins{sdk.NewInt64Coin(types.DefaultClaimDenom, 500000000)},
// 				ActionCompleted:        []bool{false, false, false, false},
// 			},
// 		},
// 	}
// )

// func TestClaimInitGenesis(t *testing.T) {
// 	app := simapp.Setup(false)
// 	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
// 	ctx = ctx.WithBlockTime(now.Add(time.Second))
// 	genesis := testGenesis
// 	claim.InitGenesis(ctx, *app.ClaimKeeper, genesis)
// 	app.ClaimKeeper.CreateModuleAccount(ctx, sdk.NewInt64Coin(types.DefaultClaimDenom, 750000000))

// 	coin := app.ClaimKeeper.GetModuleAccountBalances(ctx)
// 	require.Equal(t, coin.String(), genesis.ModuleAccountBalance.String())

// 	params, err := app.ClaimKeeper.GetParams(ctx)
// 	require.NoError(t, err)
// 	require.Equal(t, params, genesis.Params)

// 	claimRecords := app.ClaimKeeper.GetClaimRecords(ctx)
// 	require.Equal(t, claimRecords, genesis.ClaimRecords)
// }

// func TestClaimExportGenesis(t *testing.T) {
// 	app := simapp.Setup(false)
// 	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
// 	ctx = ctx.WithBlockTime(now.Add(time.Second))
// 	genesis := testGenesis
// 	claim.InitGenesis(ctx, *app.ClaimKeeper, genesis)
// 	app.ClaimKeeper.CreateModuleAccount(ctx, sdk.NewInt64Coin(types.DefaultClaimDenom, 750000000))

// 	claimRecord, err := app.ClaimKeeper.GetClaimRecord(ctx, acc2)
// 	require.NoError(t, err)
// 	require.Equal(t, claimRecord, types.ClaimRecord{
// 		Address:                acc2.String(),
// 		InitialClaimableAmount: sdk.Coins{sdk.NewInt64Coin(types.DefaultClaimDenom, 500000000)},
// 		ActionCompleted:        []bool{false, false, false, false},
// 	})

// 	claimableAmount, err := app.ClaimKeeper.GetClaimableAmountForAction(ctx, acc2, types.ActionSwap)
// 	require.NoError(t, err)
// 	require.Equal(t, claimableAmount, sdk.Coins{sdk.NewInt64Coin(types.DefaultClaimDenom, 125000000)})

// 	app.ClaimKeeper.AfterSwap(ctx, acc2)

// 	genesisExported := claim.ExportGenesis(ctx, *app.ClaimKeeper)
// 	require.Equal(t, genesisExported.ModuleAccountBalance, genesis.ModuleAccountBalance.Sub(claimableAmount[0]))
// 	require.Equal(t, genesisExported.Params, genesis.Params)
// 	require.Equal(t, genesisExported.ClaimRecords, []types.ClaimRecord{
// 		{
// 			Address:                acc1.String(),
// 			InitialClaimableAmount: sdk.Coins{sdk.NewInt64Coin(types.DefaultClaimDenom, 1000000000)},
// 			ActionCompleted:        []bool{true, false, true, true},
// 		},
// 		{
// 			Address:                acc2.String(),
// 			InitialClaimableAmount: sdk.Coins{sdk.NewInt64Coin(types.DefaultClaimDenom, 500000000)},
// 			ActionCompleted:        []bool{false, true, false, false},
// 		},
// 	})
// }
