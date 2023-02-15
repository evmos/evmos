package testutil

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/evmos/evmos/v11/app"
)

// Commit commits a block at a given time. Reminder: At the end of each
// Tendermint Consensus round the following methods are run
//  1. BeginBlock
//  2. DeliverTx
//  3. EndBlock
//  4. Commit
func Commit(ctx sdk.Context, app *app.Evmos, t time.Duration) sdk.Context {
	header := ctx.BlockHeader()
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: header.Height})
	_ = app.Commit()

	header.Height++
	header.Time = header.Time.Add(t)
	app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	return app.BaseApp.NewContext(false, header)
}
