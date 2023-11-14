// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// BalancesMethod defines the ABI method name for the bank Balances
	// query.
	BalancesMethod = "balances"
	// TotalSupplyMethod defines the ABI method name for the bank TotalSupply
	// query.
	TotalSupplyMethod = "totalSupply"
)

// Balances returns all the native token balances (address, amount) for a given
// account. This method charges the account the corresponding value of an ERC-20
// balanceOf call for each token returned.
func (p Precompile) Balances(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	account, err := ParseBalancesArgs(args)
	if err != nil {
		return nil, err
	}

	i := 0
	balances := make([]Balance, 0)

	p.bankKeeper.IterateAccountBalances(ctx, account, func(coin sdk.Coin) bool {
		defer func() { i++ }()

		// NOTE: we already charged for a single balanceOf request so we don't
		// need to charge on the first iteration
		if i > 0 {
			ctx.GasMeter().ConsumeGas(GasBalanceOf, "ERC-20 extension balances method")
		}

		contractAddress, err := p.erc20Keeper.GetCoinAddress(ctx, coin.Denom)
		if err != nil {
			return false
		}

		balances = append(balances, Balance{
			ContractAddress: contractAddress,
			Amount:          coin.Amount.BigInt(),
		})

		return false
	})

	return method.Outputs.Pack(balances)
}

// TotalSupply returns the total supply of all the native tokens.
// This method charges the account the corresponding value of a ERC-20 totalSupply
// call for each token returned.
func (p Precompile) TotalSupply(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	_ []interface{},
) ([]byte, error) {
	i := 0
	totalSupply := make([]Balance, 0)

	p.bankKeeper.IterateTotalSupply(ctx, func(coin sdk.Coin) bool {
		defer func() { i++ }()

		// NOTE: we already charged for a single totalSupply request so we don't
		// need to charge on the first iteration
		if i > 0 {
			ctx.GasMeter().ConsumeGas(GasTotalSupply, "ERC-20 extension totalSupply method")
		}

		contractAddress, err := p.erc20Keeper.GetCoinAddress(ctx, coin.Denom)
		if err != nil {
			return false
		}

		totalSupply = append(totalSupply, Balance{
			ContractAddress: contractAddress,
			Amount:          coin.Amount.BigInt(),
		})

		return false
	})

	return method.Outputs.Pack(totalSupply)
}
