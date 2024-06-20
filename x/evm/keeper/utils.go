package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/x/evm/types"
)

// IsContract determines if the given address is a smart contract.
func (k *Keeper) IsContract(ctx sdk.Context, addr common.Address) bool {
	return !types.BytesAreEmptyCodeHash(k.GetCodeHash(ctx, addr).Bytes())
}
