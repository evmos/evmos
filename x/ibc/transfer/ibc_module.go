package transfer

import (
	ibctransfer "github.com/cosmos/ibc-go/v5/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v5/modules/apps/transfer/keeper"
)

// IBCModule implements the ICS26 interface for transfer given the transfer keeper.
type IBCModule struct {
	*ibctransfer.IBCModule
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k ibctransferkeeper.Keeper) IBCModule {
	transferModule := ibctransfer.NewIBCModule(k)
	return IBCModule{
		IBCModule: &transferModule,
	}
}
