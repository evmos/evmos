package evm

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/evmos/evmos/v9/x/ibc/evm/types"
)

// NewHandler defines the IBC EVM transaction handler instance
func NewHandler(server types.MsgServer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgCallEVM:
			res, err := server.CallEVM(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			err := sdkerrors.Wrapf(errortypes.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, err
		}
	}
}
