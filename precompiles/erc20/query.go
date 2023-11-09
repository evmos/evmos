// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/evmos/evmos/v15/precompiles/authorization"
	transferkeeper "github.com/evmos/evmos/v15/x/ibc/transfer/keeper"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// NameMethod defines the ABI method name for the ERC-20 Name
	// query.
	NameMethod = "name"
	// SymbolMethod defines the ABI method name for the ERC-20 Symbol
	// query.
	SymbolMethod = "symbol"
	// DecimalsMethod defines the ABI method name for the ERC-20 Decimals
	// query.
	DecimalsMethod = "decimals"
	// TotalSupplyMethod defines the ABI method name for the ERC-20 TotalSupply
	// query.
	TotalSupplyMethod = "totalSupply"
	// BalanceOfMethod defines the ABI method name for the ERC-20 BalanceOf
	// query.
	BalanceOfMethod = "balanceOf"
)

// Name returns the name of the token. If the token metadata is registered in the
// bank module, it returns its name. Otherwise it returns the base denomination of
// the token capitalized (eg. uatom -> Atom).
func (p Precompile) Name(
	ctx sdk.Context,
	_ *vm.Contract,
	_ vm.StateDB,
	method *abi.Method,
	_ []interface{},
) ([]byte, error) {
	metadata, found := p.bankKeeper.GetDenomMetaData(ctx, p.tokenPair.Denom)
	if found {
		return method.Outputs.Pack(metadata.Name)
	}

	// Infer the denomination name from the coin denomination base denom
	denomTrace, err := GetDenomTrace(p.transferKeeper, ctx, p.tokenPair.Denom)
	if err != nil {
		// FIXME: return 'not supported' (same error as when you call the method on an ERC20.sol)
		return nil, err
	}

	// safety check
	if len(denomTrace.BaseDenom) < 3 {
		// FIXME: return not supported (same error as when you call the method on an ERC20.sol)
		return nil, nil
	}

	name := strings.ToUpper(string(denomTrace.BaseDenom[1])) + denomTrace.BaseDenom[2:]
	return method.Outputs.Pack(name)
}

// Symbol returns the symbol of the token. If the token metadata is registered in the
// bank module, it returns its symbol. Otherwise it returns the base denomination of
// the token in uppercase (eg. uatom -> ATOM).
func (p Precompile) Symbol(
	ctx sdk.Context,
	_ *vm.Contract,
	_ vm.StateDB,
	method *abi.Method,
	_ []interface{},
) ([]byte, error) {
	metadata, found := p.bankKeeper.GetDenomMetaData(ctx, p.tokenPair.Denom)
	if found {
		return method.Outputs.Pack(metadata.Symbol)
	}

	denomTrace, err := GetDenomTrace(p.transferKeeper, ctx, p.tokenPair.Denom)
	if err != nil {
		// FIXME: return not supported (same error as when you call the method on an ERC20.sol)
		return nil, err
	}

	// safety check
	if len(denomTrace.BaseDenom) < 3 {
		// FIXME: return not supported (same error as when you call the method on an ERC20.sol)
		return nil, nil
	}

	symbol := strings.ToUpper(denomTrace.BaseDenom[1:])
	return method.Outputs.Pack(symbol)
}

// Decimals returns the decimals places of the token. If the token metadata is registered in the
// bank module, it returns its the display denomination exponent. Otherwise it infers the decimal
// value from the first character of the base denomination (eg. uatom -> 6).
func (p Precompile) Decimals(
	ctx sdk.Context,
	_ *vm.Contract,
	_ vm.StateDB,
	method *abi.Method,
	_ []interface{},
) ([]byte, error) {
	metadata, found := p.bankKeeper.GetDenomMetaData(ctx, p.tokenPair.Denom)
	if !found {
		denomTrace, err := GetDenomTrace(p.transferKeeper, ctx, p.tokenPair.Denom)
		if err != nil {
			// FIXME: return not supported (same error as when you call the method on an ERC20.sol)
			return nil, err
		}

		// we assume the decimal from the first character of the denomination
		switch string(denomTrace.BaseDenom[0]) {
		case "u": // micro (u) -> 6 decimals
			return method.Outputs.Pack(uint8(6))
		case "a": // atto (a) -> 18 decimals
			return method.Outputs.Pack(uint8(18))
		}
		// FIXME: return not supported (same error as when you call the method on an ERC20.sol)
		return nil, nil
	}

	var decimals uint32
	for i := len(metadata.DenomUnits); i >= 0; i-- {
		if metadata.DenomUnits[i].Denom == metadata.Display {
			decimals = metadata.DenomUnits[i].Exponent
			break
		}
	}

	if decimals > math.MaxUint8 {
		return nil, errors.New("uint8 overflow: invalid decimals")
	}

	return method.Outputs.Pack(uint8(decimals)) //#nosec G701
}

// TotalSupply returns the amount of tokens in existence. It fetches the supply
// of the coin from the bank keeper and returns zero if not found.
func (p Precompile) TotalSupply(
	ctx sdk.Context,
	_ *vm.Contract,
	_ vm.StateDB,
	method *abi.Method,
	_ []interface{},
) ([]byte, error) {
	supply := p.bankKeeper.GetSupply(ctx, p.tokenPair.Denom)

	return method.Outputs.Pack(supply.Amount.BigInt())
}

// BalanceOf returns the amount of tokens owned by account. It fetches the balance
// of the coin from the bank keeper and returns zero if not found.
func (p Precompile) BalanceOf(
	ctx sdk.Context,
	_ *vm.Contract,
	_ vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	account, err := ParseBalanceOfArgs(args)
	if err != nil {
		return nil, err
	}

	balance := p.bankKeeper.GetBalance(ctx, account.Bytes(), p.tokenPair.Denom)

	return method.Outputs.Pack(balance.Amount.BigInt())
}

// Allowance returns the remaining allowance of a spender to the contract by
// checking the existence of a bank SendAuthorization.
func (p Precompile) Allowance(
	ctx sdk.Context,
	_ *vm.Contract,
	_ vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	owner, spender, err := ParseAllowanceArgs(args)
	if err != nil {
		return nil, err
	}

	granter := owner
	grantee := spender

	_, _, allowance, err := GetAuthzExpirationAndAllowance(p.AuthzKeeper, ctx, grantee, granter, p.tokenPair.Denom)
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(allowance)
}

// GetDenomTrace returns the denomination trace from the corresponding IBC denomination. If the
// denomination is is not an IBC voucher or the trace is not found, it returns an error.
func GetDenomTrace(
	transferKeeper transferkeeper.Keeper,
	ctx sdk.Context,
	denom string,
) (transfertypes.DenomTrace, error) {
	if !strings.HasPrefix(denom, "ibc/") {
		return transfertypes.DenomTrace{}, fmt.Errorf("denom is not an IBC voucher: %s", denom)
	}

	hash, err := transfertypes.ParseHexHash(denom[4:])
	if err != nil {
		return transfertypes.DenomTrace{}, err
	}

	denomTrace, found := transferKeeper.GetDenomTrace(ctx, hash)
	if !found {
		return transfertypes.DenomTrace{}, errors.New("denom trace not found")
	}

	return denomTrace, nil
}

// GetAuthzExpirationAndAllowance returns the authorization, its expiration as well as the amount of denom
// that the grantee is allowed to spend on behalf of the granter.
func GetAuthzExpirationAndAllowance(
	authzKeeper authzkeeper.Keeper,
	ctx sdk.Context,
	grantee, granter common.Address,
	denom string,
) (authz.Authorization, *time.Time, *big.Int, error) {
	authorization, expiration, err := authorization.CheckAuthzExists(ctx, authzKeeper, grantee, granter, SendMsgURL)
	// TODO: return error if doesn't exist?
	if err != nil {
		return nil, nil, common.Big0, err
	}

	sendAuth, ok := authorization.(*banktypes.SendAuthorization)
	if !ok {
		// TODO: return error if invalid authorization?
		return nil, nil, common.Big0, nil
	}

	allowance := sendAuth.SpendLimit.AmountOfNoDenomValidation(denom)
	return authorization, expiration, allowance.BigInt(), nil
}
