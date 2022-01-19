package mint_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/osmosis-labs/osmosis/app"
	"github.com/osmosis-labs/osmosis/x/mint/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestMintInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	validateGenesis := types.ValidateGenesis(*types.DefaultGenesisState())
	require.NoError(t, validateGenesis)

	developerAccount := app.AccountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	initialVestingCoins := app.BankKeeper.GetBalance(ctx, developerAccount, sdk.DefaultBondDenom)

	expectedVestingCoins, ok := sdk.NewIntFromString("225000000000000")
	require.True(t, ok)
	require.Equal(t, expectedVestingCoins, initialVestingCoins.Amount)
	require.Equal(t, int64(0), app.MintKeeper.GetLastHalvenEpochNum(ctx))
}
