// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

import (
	context "context"
	"math/big"

	"cosmossdk.io/core/address"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v20/x/evm/core/vm"

	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"
)

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	IterateAccounts(ctx context.Context, cb func(account sdk.AccountI) bool)
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, account sdk.AccountI)
	RemoveAccount(ctx context.Context, account sdk.AccountI)
	GetParams(ctx context.Context) (params authtypes.Params)
	GetSequence(ctx context.Context, account sdk.AccAddress) (uint64, error)
	AddressCodec() address.Codec
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	authtypes.BankKeeper
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

// StakingKeeper returns the historical headers kept in store.
type StakingKeeper interface {
	GetHistoricalInfo(ctx context.Context, height int64) (stakingtypes.HistoricalInfo, error)
	GetValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (stakingtypes.Validator, error)
	ValidatorAddressCodec() address.Codec
}

// FeeMarketKeeper
type FeeMarketKeeper interface {
	GetBaseFee(ctx sdk.Context) *big.Int
	GetParams(ctx sdk.Context) feemarkettypes.Params
	CalculateBaseFee(ctx sdk.Context) *big.Int
}

// Erc20Keeper defines the expected interface needed to instantiate ERC20 precompiles.
type Erc20Keeper interface {
	GetERC20PrecompileInstance(ctx sdk.Context, address common.Address) (contract vm.PrecompiledContract, found bool, err error)
}

type (
	LegacyParams = paramtypes.ParamSet
	// Subspace defines an interface that implements the legacy Cosmos SDK x/params Subspace type.
	// NOTE: This is used solely for migration of the Cosmos SDK x/params managed parameters.
	Subspace interface {
		GetParamSetIfExists(ctx sdk.Context, ps LegacyParams)
	}
)

// BankWrapper defines the methods required by the wrapper around
// the Cosmos SDK x/bank keeper that is used to manage an EVM coin
// with a configurable value for decimals.
type BankWrapper interface {
	BankKeeper

	MintAmountToAccount(ctx context.Context, recipientAddr sdk.AccAddress, amt *big.Int) error
	BurnAmountFromAccount(ctx context.Context, account sdk.AccAddress, amt *big.Int) error
}
