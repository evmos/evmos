package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/evoblockchain/evoblock/v8/x/erc20/types"
)

// UpdateParams updates the module parameters EnableERC20 and EnableEMVHook
// values to true.
func UpdateParams(ctx sdk.Context, paramstore *paramtypes.Subspace) error {
	if !paramstore.HasKeyTable() {
		ps := paramstore.WithKeyTable(types.ParamKeyTable())
		paramstore = &ps
	}

	paramstore.Set(ctx, types.ParamStoreKeyEnableErc20, true)
	paramstore.Set(ctx, types.ParamStoreKeyEnableEVMHook, true)
	return nil
}
