// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package bank

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

const (
	// BalancesMethod defines the ABI method name for the bank Balances
	// query.
	BalancesMethod = "balances"
	// TotalSupplyMethod defines the ABI method name for the bank TotalSupply
	// query.
	TotalSupplyMethod = "totalSupply"
	// SupplyOfMethod defines the ABI method name for the bank SupplyOf
	// query.
	SupplyOfMethod = "supplyOf"
)

// Balances returns given account's balances of all tokens registered in the x/bank module
// and the corresponding ERC20 address (address, amount). The amount returned for each token
// has the original decimals precision stored in the x/bank.
// This method charges the account the corresponding value of an ERC-20
// balanceOf call for each token returned.
func (p Precompile) Balances(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	account, err := ParseBalancesArgs(args)
	if err != nil {
		return nil, fmt.Errorf("error calling account balances in bank precompile: %s", err)
	}

	i := 0
	balances := make([]Balance, 0)

	p.bankKeeper.IterateAccountBalances(ctx, account, func(coin sdk.Coin) bool {
		defer func() { i++ }()

		// NOTE: we already charged for a single balanceOf request so we don't
		// need to charge on the first iteration
		if i > 0 {
			ctx.GasMeter().ConsumeGas(GasBalances, "ERC-20 extension balances method")
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

// TotalSupply returns the total supply of all tokens registered in the x/bank
// module. The amount returned for each token has the original
// decimals precision stored in the x/bank.
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

// SupplyOf returns the total supply of a given registered erc20 token
// from the x/bank module. If the ERC20 token doesn't have a registered
// TokenPair, the method returns a supply of zero.
// The amount returned with this query has the original decimals precision
// stored in the x/bank.
func (p Precompile) SupplyOf(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	erc20ContractAddress, err := ParseSupplyOfArgs(args)
	if err != nil {
		return nil, fmt.Errorf("error getting the supply in bank precompile: %s", err)
	}

	tokenPairID := p.erc20Keeper.GetERC20Map(ctx, erc20ContractAddress)
	tokenPair, found := p.erc20Keeper.GetTokenPair(ctx, tokenPairID)
	if !found {
		return method.Outputs.Pack(big.NewInt(0))
	}

	supply := p.bankKeeper.GetSupply(ctx, tokenPair.Denom)

	return method.Outputs.Pack(supply.Amount.BigInt())
}
