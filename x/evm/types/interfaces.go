// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

import (
	"math/big"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v19/x/evm/core/vm"

	feemarkettypes "github.com/evmos/evmos/v19/x/feemarket/types"
)

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	IterateAccounts(ctx sdk.Context, cb func(account authtypes.AccountI) bool)
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	SetAccount(ctx sdk.Context, account authtypes.AccountI)
	RemoveAccount(ctx sdk.Context, account authtypes.AccountI)
	GetParams(ctx sdk.Context) (params authtypes.Params)
	GetSequence(ctx sdk.Context, account sdk.AccAddress) (uint64, error)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	authtypes.BankKeeper
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}

// BankWrapper is a wrapper around the Cosmos SDK bank keeper
// that is used to manage an evm denom with 6 decimals.
// The wrapper makes the corresponding conversions to achieve:
// - With the EVM, the wrapper works always with 18 decimals.
// - With the Cosmos bank module, the wrapper works always with 6 decimals.
type BankWrapper interface {
	IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error
	SendCoins(ctx sdk.Context, from, to sdk.AccAddress, amt sdk.Coins) error
	// GetBalance returns the balance converted to 18 decimals
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	// SendCoinsFromModuleToAccount scales down
	// from 18 decimals to 6 decimals the coins amount provided
	// and sends the coins from the module to the account
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	// SendCoinsFromAccountToModule scales down
	// from 18 decimals to 6 decimals the coins amount provided
	// and sends the coins
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	// MintCoinsToAccount scales down from 18 decimals to 6 decimals the coins amount provided
	// and mints that to the provided account
	MintCoinsToAccount(ctx sdk.Context, moduleName string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	// BurnAccountCoins scales down from 18 decimals to 6 decimals the coins amount provided
	// and burns that coins of the provided account
	BurnAccountCoins(ctx sdk.Context, account sdk.AccAddress, burningModule string, amt sdk.Coins) error
}

// StakingKeeper returns the historical headers kept in store.
type StakingKeeper interface {
	GetHistoricalInfo(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool)
	GetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) (validator stakingtypes.Validator, found bool)
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
