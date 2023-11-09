// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	context "context"
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/evmos/evmos/v15/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"

	inflationtypes "github.com/evmos/evmos/v15/x/inflation/v1/types"
)

// AccountKeeper defines the expected interface needed to retrieve account info.
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetSequence(context.Context, sdk.AccAddress) (uint64, error)
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	IsSendEnabledCoin(ctx context.Context, coin sdk.Coin) bool
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	HasSupply(ctx context.Context, denom string) bool
	IterateAccountBalances(ctx context.Context, addr sdk.AccAddress, cb func(sdk.Coin) bool)
}

// GovKeeper defines the expected governance keeper interface used on incentives
type GovKeeper interface {
	Logger(context.Context) log.Logger
	GetVotingParams(ctx context.Context) govv1beta1.VotingParams
	GetProposal(ctx context.Context, proposalID uint64) (govv1beta1.Proposal, bool)
	InsertActiveProposalQueue(ctx context.Context, proposalID uint64, timestamp time.Time)
	RemoveFromActiveProposalQueue(ctx context.Context, proposalID uint64, timestamp time.Time)
	SetProposal(ctx context.Context, proposal govv1beta1.Proposal)
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
		WithKeyTable(table paramtypes.KeyTable) paramtypes.Subspace
	}
)
