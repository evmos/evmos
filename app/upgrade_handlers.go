package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
)

func SetupUpgradeHandlers(app *Evmos) {
	app.UpgradeKeeper.SetUpgradeHandler("2.0.0", func(ctx sdk.Context, plan upgrade.Plan) {})
}
