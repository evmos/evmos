package types

import (
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/tharsis/ethermint/x/evm/statedb"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	inflationtypes "github.com/tharsis/evmos/v4/x/inflation/types"
)

// AccountKeeper defines the expected interface needed to retrieve account info.
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetSequence(sdk.Context, sdk.AccAddress) (uint64, error)
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
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

// InflationKeeper defines the expected mint keeper interface used on incentives
type InflationKeeper interface {
	GetParams(ctx sdk.Context) (params inflationtypes.Params)
}

// EVMKeeper defines the expected EVM keeper interface used on erc20
type EVMKeeper interface {
	GetParams(ctx sdk.Context) evmtypes.Params
	GetAccountWithoutBalance(ctx sdk.Context, addr common.Address) *statedb.Account
}

// Stakekeeper defines the expected staking keeper interface used on incentives
type StakeKeeper interface{}
