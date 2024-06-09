// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package accesscontrol

import (
	"cosmossdk.io/math"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
)

const (
	MethodMint = "mint"
	MethodBurn = "burn"
)

// Mint mints Bank coins to the recipient address.
func (p Precompile) Mint(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	args []interface{},
) ([]byte, error) {
	to, amount, err := ParseMintArgs(args)
	if err != nil {
		return nil, err
	}

	if err := p.onlyRole(ctx, RoleMinter, contract.CallerAddress); err != nil {
		return nil, err
	}

	// Mint new Coins and send them to the recipient address
	if err := p.BankKeeper.MintCoins(ctx, erc20types.ModuleName, sdk.Coins{{Denom: p.TokenPair.Denom, Amount: math.NewIntFromBigInt(amount)}}); err != nil {
		return nil, err
	}

	if err := p.BankKeeper.SendCoinsFromModuleToAccount(ctx, erc20types.ModuleName, to.Bytes(), sdk.Coins{{Denom: p.TokenPair.Denom, Amount: math.NewIntFromBigInt(amount)}}); err != nil {
		return nil, err
	}

	if err := p.EmitEventMint(ctx, stateDB, to, amount); err != nil {
		return nil, err
	}

	if err := p.ERC20Precompile.EmitTransferEvent(ctx, stateDB, common.Address{}, to, amount); err != nil {
		return nil, err
	}

	return nil, nil
}

// Burn burns Bank coins from the caller address.
func (p Precompile) Burn(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	args []interface{},
) ([]byte, error) {
	amount, err := ParseBurnArgs(args)
	if err != nil {
		return nil, err
	}

	if err := p.onlyRole(ctx, RoleBurner, contract.CallerAddress); err != nil {
		return nil, err
	}

	balance := p.BankKeeper.GetBalance(ctx, contract.CallerAddress.Bytes(), p.TokenPair.Denom)
	if balance.Amount.BigInt().Cmp(amount) < 0 {
		return nil, fmt.Errorf("burn amount exceeds balance")
	}

	// Send coins to module account and then burn them
	if err := p.BankKeeper.SendCoinsFromAccountToModule(ctx, contract.CallerAddress.Bytes(), erc20types.ModuleName, sdk.Coins{{Denom: p.TokenPair.Denom, Amount: math.NewIntFromBigInt(amount)}}); err != nil {
		return nil, err
	}

	if err := p.BankKeeper.BurnCoins(ctx, erc20types.ModuleName, sdk.Coins{{Denom: p.TokenPair.Denom, Amount: math.NewIntFromBigInt(amount)}}); err != nil {
		return nil, err
	}

	if err := p.EmitEventBurn(ctx, stateDB, contract.CallerAddress, amount); err != nil {
		return nil, err
	}

	if err := p.ERC20Precompile.EmitTransferEvent(ctx, stateDB, contract.CallerAddress, common.Address{}, amount); err != nil {
		return nil, err
	}

	return nil, nil
}
