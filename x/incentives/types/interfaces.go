package types

import (
	"time"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	inflationtypes "github.com/tharsis/evmos/x/inflation/types"
)

// AccountKeeper defines the expected interface needed to retrieve account info.
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetSequence(sdk.Context, sdk.AccAddress) (uint64, error)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	IsSendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	HasSupply(ctx sdk.Context, denom string) bool
}

// GovKeeper defines the expected governance keeper interface used on incentives
type GovKeeper interface {
	Logger(sdk.Context) log.Logger
	GetVotingParams(ctx sdk.Context) govtypes.VotingParams
	GetProposal(ctx sdk.Context, proposalID uint64) (govtypes.Proposal, bool)
	InsertActiveProposalQueue(ctx sdk.Context, proposalID uint64, timestamp time.Time)
	RemoveFromActiveProposalQueue(ctx sdk.Context, proposalID uint64, timestamp time.Time)
	SetProposal(ctx sdk.Context, proposal govtypes.Proposal)
}

// MintKeeper defines the expected mint keeper interface used on incentives
type MintKeeper interface {
	GetParams(ctx sdk.Context) (params minttypes.Params)
}

// InflationKeeper defines the expected mint keeper interface used on incentives
type InflationKeeper interface {
	GetParams(ctx sdk.Context) (params inflationtypes.Params)
}

// Stakekeeper defines the expected staking keeper interface used on incentives
type StakeKeeper interface{}
