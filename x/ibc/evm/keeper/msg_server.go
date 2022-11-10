package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v9/x/ibc/evm/types"
)

var _ types.MsgServer = &Keeper{}

func (k Keeper) CallEVM(goCtx context.Context, msg *types.MsgCallEVM) (*types.MsgCallEVMResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Check if EVM send param is enabled
	if !k.GetSendEvmTxEnabled(ctx) {
		return nil, types.ErrReceiveDisabled
	}

	coin := sdk.Coin{Denom: msg.Denom, Amount: msg.Amount}
	// Check if the account has the amount to process this transaction

	k.sendEvmTx(ctx)

	res := &types.MsgCallEVMResponse{}
	return res, nil

}
