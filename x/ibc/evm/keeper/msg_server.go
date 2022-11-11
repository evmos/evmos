package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/ibc/evm/types"
)

var _ types.MsgServer = &Keeper{}

func (k Keeper) CallEVM(goCtx context.Context, msg *types.MsgCallEVM) (*types.MsgCallEVMResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Check if EVM send param is enabled
	if !k.GetSendEvmTxEnabled(ctx) {
		return nil, types.ErrReceiveDisabled
	}

	amount, ok := sdk.NewIntFromString(msg.Amount)
	if !ok {
		return nil, fmt.Errorf("invalid amount %s", msg.Amount)
	}

	coin := sdk.Coin{Denom: msg.Denom, Amount: amount}
	// Check if the account has the amount to process this transaction

	k.
		k.sendEvmTx(ctx)

	res := &types.MsgCallEVMResponse{}
	return res, nil

}
