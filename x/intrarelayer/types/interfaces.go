package types

import (
	"time"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
	BlockedAddr(addr sdk.AccAddress) bool
	GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)
	HasSupply(ctx sdk.Context, denom string) bool
}

// EVMKeeper defines the expected EVM keeper interface used on intrarelayer
// TODO: define
type EVMKeeper interface{}

// GovKeeper defines the expected governance keeper interface used on intrarelayer
type GovKeeper interface {
	Logger(sdk.Context) log.Logger
	GetVotingParams(ctx sdk.Context) govtypes.VotingParams
	GetProposal(ctx sdk.Context, proposalID uint64) (govtypes.Proposal, bool)
	InsertActiveProposalQueue(ctx sdk.Context, proposalID uint64, timestamp time.Time)
	RemoveFromActiveProposalQueue(ctx sdk.Context, proposalID uint64, timestamp time.Time)
	SetProposal(ctx sdk.Context, proposal govtypes.Proposal)
}
