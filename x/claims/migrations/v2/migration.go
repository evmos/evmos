package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/tharsis/evmos/v2/x/claims/types"
)

func MigrateStore(ctx sdk.Context, paramstore paramtypes.Subspace) error {
	paramstore.WithKeyTable(types.ParamKeyTable())
	paramstore.Set(ctx, types.ParamStoreKeyAuthorizedChannels, types.DefaultAuthorizedChannels)
	paramstore.Set(ctx, types.ParamStoreKeyEVMChannels, types.DefaultEVMChannels)
	return nil
}
