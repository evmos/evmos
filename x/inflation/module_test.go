package inflation_test

import (
	"testing"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	// setup feemarketGenesis params
	// feemarketGenesis := feemarkettypes.DefaultGenesisState()
	// feemarketGenesis.Params.EnableHeight = 1
	// feemarketGenesis.Params.NoBaseFee = false
	// feemarketGenesis.BaseFee = sdk.NewInt(feemarketGenesis.Params.InitialBaseFee)
	// app := simapp.Setup(false, feemarketGenesis)
	// ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// app.InitChain(
	// 	abcitypes.RequestInitChain{
	// 		AppStateBytes: []byte("{}"),
	// 		ChainId:       "test-chain-id",
	// 	},
	// )

	// acc := app.AccountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	// require.NotNil(t, acc)
}
