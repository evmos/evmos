package keeper

import (
	"context"
	"github.com/evmos/evmos/x/ibc/evm/types"
)

var _ types.MsgServer = &Keeper{}

func (k Keeper) CallEVM(ctx context.Context, msg *types.MsgCallEVM) (*types.MsgCallEVMResponse, error) {

	res := *types.MsgCallEVMResponse{}
	return res, nil

}
