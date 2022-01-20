package inflation_test

import (
	"testing"
)

func TestMintInitGenesis(t *testing.T) {
	// setup feemarketGenesis params
	// feemarketGenesis := feemarkettypes.DefaultGenesisState()
	// feemarketGenesis.Params.EnableHeight = 1
	// feemarketGenesis.Params.NoBaseFee = false
	// feemarketGenesis.BaseFee = sdk.NewInt(feemarketGenesis.Params.InitialBaseFee)
	// app := simapp.Setup(false, feemarketGenesis)
	// ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// validateGenesis := types.DefaultGenesisState().Validate()
	// require.NoError(t, validateGenesis)

	// // TODO fix vesting account
	// // developerAccount := app.AccountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	// // initialVestingCoins := app.BankKeeper.GetBalance(ctx, developerAccount, sdk.DefaultBondDenom)

	// // expectedVestingCoins, ok := sdk.NewIntFromString("225000000000000")
	// // require.True(t, ok)
	// // require.Equal(t, expectedVestingCoins, initialVestingCoins.Amount)

	// require.Equal(t, int64(0), app.InflationKeeper.GetLastHalvenEpochNum(ctx))
}
