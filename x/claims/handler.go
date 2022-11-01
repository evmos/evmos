package claims

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/evmos/evmos/v10/x/claims/keeper"
	"github.com/evmos/evmos/v10/x/claims/types"
)

// NewHandler returns claim module messages
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(_ sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
		return nil, sdkerrors.Wrap(errortypes.ErrUnknownRequest, errMsg)
	}
}
