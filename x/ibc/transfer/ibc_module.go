package transfer

import (
	ibctransfer "github.com/cosmos/ibc-go/v5/modules/apps/transfer"
	"github.com/evmos/evmos/v10/x/ibc/transfer/keeper"
)

// IBCModule implements the ICS26 interface for transfer given the transfer keeper.
type IBCModule struct {
	*ibctransfer.IBCModule
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k keeper.Keeper) IBCModule {
	transferModule := ibctransfer.NewIBCModule(*k.Keeper)
	return IBCModule{
		IBCModule: &transferModule,
	}
}
