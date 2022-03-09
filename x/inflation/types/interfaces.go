package types // noalias

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the contract required for account APIs.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) types.ModuleAccountI
	GetAccount(sdk.Context, sdk.AccAddress) types.AccountI
	SetAccount(sdk.Context, types.AccountI)
}

// BankKeeper defines the contract needed to be fulfilled for banking and supply
// dependencies.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	HasSupply(ctx sdk.Context, denom string) bool
}

// DistrKeeper defines the contract needed to be fulfilled for distribution keeper
type DistrKeeper interface {
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// StakingKeeper expected staking keeper
type StakingKeeper interface {
	// BondedRatio the fraction of the staking tokens which are currently bonded
	BondedRatio(ctx sdk.Context) sdk.Dec
	StakingTokenSupply(ctx sdk.Context) sdk.Int
	TotalBondedTokens(ctx sdk.Context) sdk.Int
}
