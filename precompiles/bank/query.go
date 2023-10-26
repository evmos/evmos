// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package bank

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
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
// account. This method charges the account the corresponding value of a ERC-20
// balanceOf call for each token returned.
func (p Precompile) Balances(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	account, err := ParseBalances(args)
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

		contractAddress, ok := p.GetCoinAddress(ctx, coin.Denom)
		if !ok {
			return false
		}

		balances = append(balances, Balance{
			ContractAddress: contractAddress,
			Amount:          coin.Amount.BigInt(),
		})

		return false
	})

	if len(balances) > 1 {
		cost := uint64((len(balances) - 1)) * GasBalanceOf
		ctx.GasMeter().ConsumeGas(cost, "erc-20 extension balances method")
	}

	return method.Outputs.Pack(balances)
}

// Balances returns the total supply of all the native tokens.
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

		contractAddress, ok := p.GetCoinAddress(ctx, coin.Denom)
		if !ok {
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

func (p Precompile) GetCoinAddress(ctx sdk.Context, denom string) (contractAddress common.Address, ok bool) {
	id := p.erc20Keeper.GetDenomMap(ctx, denom)
	if len(id) != 0 {
		tokenPair, found := p.erc20Keeper.GetTokenPair(ctx, id)
		if !found {
			return common.Address{}, false
		}

		return tokenPair.GetERC20Contract(), true
	}

	if !strings.HasPrefix(denom, "ibc/") {
		return common.Address{}, false
	}

	if len(denom) < 5 || strings.TrimSpace(denom[4:]) == "" {
		return common.Address{}, false
	}

	bz, err := transfertypes.ParseHexHash(denom[4:])
	if err != nil {
		return common.Address{}, false
	}

	return common.BytesToAddress(bz), true
}
