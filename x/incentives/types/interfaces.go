package types

import (
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/ethereum/go-ethereum/common"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/evmos/ethermint/x/evm/statedb"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	inflationtypes "github.com/evmos/evmos/v10/x/inflation/types"
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
	IterateAccountBalances(ctx sdk.Context, addr sdk.AccAddress, cb func(sdk.Coin) bool)
}

// GovKeeper defines the expected governance keeper interface used on incentives
type GovKeeper interface {
	Logger(sdk.Context) log.Logger
	GetVotingParams(ctx sdk.Context) govv1beta1.VotingParams
	GetProposal(ctx sdk.Context, proposalID uint64) (govv1beta1.Proposal, bool)
	InsertActiveProposalQueue(ctx sdk.Context, proposalID uint64, timestamp time.Time)
	RemoveFromActiveProposalQueue(ctx sdk.Context, proposalID uint64, timestamp time.Time)
	SetProposal(ctx sdk.Context, proposal govv1beta1.Proposal)
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

type (
	LegacyParams = paramtypes.ParamSet
	// Subspace defines an interface that implements the legacy Cosmos SDK x/params Subspace type.
	// NOTE: This is used solely for migration of the Cosmos SDK x/params managed parameters.
	Subspace interface {
		GetParamSetIfExists(ctx sdk.Context, ps LegacyParams)
	}
)
