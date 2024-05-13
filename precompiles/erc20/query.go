// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v18/ibc"
	auth "github.com/evmos/evmos/v18/precompiles/authorization"

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
// bank module, it returns its name. Otherwise, it returns the base denomination of
// the token capitalized (e.g. uatom -> Atom).
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

	baseDenom, err := p.getBaseDenomFromIBCVoucher(ctx, p.tokenPair.Denom)
	if err != nil {
		return nil, ConvertErrToERC20Error(err)
	}

	name := strings.ToUpper(string(baseDenom[1])) + baseDenom[2:]
	return method.Outputs.Pack(name)
}

// Symbol returns the symbol of the token. If the token metadata is registered in the
// bank module, it returns its symbol. Otherwise, it returns the base denomination of
// the token in uppercase (e.g. uatom -> ATOM).
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

	baseDenom, err := p.getBaseDenomFromIBCVoucher(ctx, p.tokenPair.Denom)
	if err != nil {
		return nil, ConvertErrToERC20Error(err)
	}

	symbol := strings.ToUpper(baseDenom[1:])
	return method.Outputs.Pack(symbol)
}

// Decimals returns the decimals places of the token. If the token metadata is registered in the
// bank module, it returns the display denomination exponent. Otherwise, it infers the decimal
// value from the first character of the base denomination (e.g. uatom -> 6).
func (p Precompile) Decimals(
	ctx sdk.Context,
	_ *vm.Contract,
	_ vm.StateDB,
	method *abi.Method,
	_ []interface{},
) ([]byte, error) {
	metadata, found := p.bankKeeper.GetDenomMetaData(ctx, p.tokenPair.Denom)
	if !found {
		denomTrace, err := ibc.GetDenomTrace(p.transferKeeper, ctx, p.tokenPair.Denom)
		if err != nil {
			return nil, ConvertErrToERC20Error(err)
		}

		// we assume the decimal from the first character of the denomination
		decimals, err := ibc.DeriveDecimalsFromDenom(denomTrace.BaseDenom)
		if err != nil {
			return nil, ConvertErrToERC20Error(err)
		}
		return method.Outputs.Pack(decimals)
	}

	var (
		decimals     uint32
		displayFound bool
	)
	for i := len(metadata.DenomUnits) - 1; i >= 0; i-- {
		if metadata.DenomUnits[i].Denom == metadata.Display {
			decimals = metadata.DenomUnits[i].Exponent
			displayFound = true
			break
		}
	}

	if !displayFound {
		return nil, ConvertErrToERC20Error(fmt.Errorf(
			"display denomination not found for denom: %q",
			p.tokenPair.Denom,
		))
	}

	if decimals > math.MaxUint8 {
		return nil, ConvertErrToERC20Error(fmt.Errorf(
			"uint8 overflow: invalid decimals: %d",
			decimals,
		))
	}

	return method.Outputs.Pack(uint8(decimals)) //#nosec G701 // we are checking for overflow above
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

	// NOTE: In case the allowance is queried by the owner, we return the max uint256 value, which
	// resembles an infinite allowance.
	if bytes.Equal(owner.Bytes(), spender.Bytes()) {
		return method.Outputs.Pack(abi.MaxUint256)
	}

	_, _, allowance, err := GetAuthzExpirationAndAllowance(p.AuthzKeeper, ctx, spender, owner, p.tokenPair.Denom)
	if err != nil {
		// NOTE: We are not returning the error here, because we want to align the behavior with
		// standard ERC20 smart contracts, which return zero if an allowance is not found.
		allowance = common.Big0
	}

	return method.Outputs.Pack(allowance)
}

// GetAuthzExpirationAndAllowance returns the authorization, its expiration as well as the amount of denom
// that the grantee is allowed to spend on behalf of the granter.
func GetAuthzExpirationAndAllowance(
	authzKeeper authzkeeper.Keeper,
	ctx sdk.Context,
	grantee, granter common.Address,
	denom string,
) (authz.Authorization, *time.Time, *big.Int, error) {
	authorization, expiration, err := auth.CheckAuthzExists(ctx, authzKeeper, grantee, granter, SendMsgURL)
	if err != nil {
		return nil, nil, common.Big0, err
	}

	sendAuth, ok := authorization.(*banktypes.SendAuthorization)
	if !ok {
		return nil, nil, common.Big0, fmt.Errorf(
			"expected authorization to be a %T", banktypes.SendAuthorization{},
		)
	}

	allowance := sendAuth.SpendLimit.AmountOfNoDenomValidation(denom)
	return authorization, expiration, allowance.BigInt(), nil
}

// getBaseDenomFromIBCVoucher returns the base denomination from the given IBC voucher denomination.
func (p Precompile) getBaseDenomFromIBCVoucher(ctx sdk.Context, denom string) (string, error) {
	// Infer the denomination name from the coin denomination base denom
	denomTrace, err := ibc.GetDenomTrace(p.transferKeeper, ctx, denom)
	if err != nil {
		// FIXME: return 'not supported' (same error as when you call the method on an ERC20.sol)
		return "", err
	}

	// safety check
	if len(denomTrace.BaseDenom) < 3 {
		// FIXME: return not supported (same error as when you call the method on an ERC20.sol)
		return "", fmt.Errorf("invalid base denomination; should be at least length 3; got: %q", denomTrace.BaseDenom)
	}

	return denomTrace.BaseDenom, nil
}
