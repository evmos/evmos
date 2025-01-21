package werc20

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper interface {
	// Mint mints new tokens to the given address
	MintCoins(ctx sdk.Context, sender, to sdk.AccAddress, amount math.Int, token string) error
	// Burn burns tokens from the given address
	BurnCoins(ctx sdk.Context, sender sdk.AccAddress, amount math.Int, token string) error
	// OwnerAddress returns the owner address of the token
	GetTokenPairOwnerAddress(ctx sdk.Context, token string) (sdk.AccAddress, error)
	// TransferOwnership transfers ownership of the token to the new owner
	TransferOwnership(ctx sdk.Context, sender sdk.AccAddress, newOwner sdk.AccAddress, token string) error
}
