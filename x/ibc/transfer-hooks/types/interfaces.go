package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TransferKeeper interface {
	DenomPathFromHash(ctx sdk.Context, denom string) (string, error)
}
