package inflation_test

// TODO
// import (
// 	"testing"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/stretchr/testify/require"
// 	abcitypes "github.com/tendermint/tendermint/abci/types"
// 	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
// 	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

// 	simapp "github.com/tharsis/evmos/app"
// 	"github.com/tharsis/evmos/x/inflation/types"
// )

// func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
// 	// setup feemarketGenesis params
// 	feemarketGenesis := feemarkettypes.DefaultGenesisState()
// 	feemarketGenesis.Params.EnableHeight = 1
// 	feemarketGenesis.Params.NoBaseFee = false
// 	feemarketGenesis.BaseFee = sdk.NewInt(feemarketGenesis.Params.InitialBaseFee)
// 	app := simapp.Setup(false, feemarketGenesis)
// 	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

// 	app.InitChain(
// 		abcitypes.RequestInitChain{
// 			AppStateBytes: []byte("{}"),
// 			ChainId:       "test-chain-id",
// 		},
// 	)
// 	address := app.AccountKeeper.GetModuleAddress(types.ModuleName)
// 	acc := app.AccountKeeper.GetAccount(ctx, address)
// 	require.NotNil(t, acc)
// }
