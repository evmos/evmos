package inflation_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tharsis/ethermint/tests"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	simapp "github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/x/inflation/types"
	inflationtypes "github.com/tharsis/evmos/x/inflation/types"
)

func TesInflationInitGenesis(t *testing.T) {
	// setup feemarketGenesis params
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 1
	feemarketGenesis.Params.NoBaseFee = false
	feemarketGenesis.BaseFee = sdk.NewInt(feemarketGenesis.Params.InitialBaseFee)

	// setup inflation
	inflationGenesis := inflationtypes.DefaultGenesisState()
	inflationGenesis.Params.TeamAddress = sdk.AccAddress(tests.GenerateAddress().Bytes()).String()

	app := simapp.Setup(false, feemarketGenesis, inflationGenesis)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	validateGenesis := types.DefaultGenesisState().Validate()
	require.NoError(t, validateGenesis)
	epochMintProvision, _ := app.InflationKeeper.GetEpochMintProvision(ctx)
	require.Equal(t, int64(0), epochMintProvision)

	// TODO test minting vesting on genesis
}
