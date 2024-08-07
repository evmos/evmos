// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/evmos/evmos/v19/x/evm/keeper"
	"github.com/evmos/evmos/v19/x/evm/types"
)

// InitGenesis initializes genesis state based on exported genesis
func InitGenesis(
	ctx sdk.Context,
	k *keeper.Keeper,
	accountKeeper types.AccountKeeper,
	data types.GenesisState,
) []abci.ValidatorUpdate {
	k.WithChainID(ctx)

	err := k.SetParams(ctx, data.Params)
	if err != nil {
		panic(fmt.Errorf("error setting params %s", err))
	}

	// ensure evm module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the EVM module account has not been set")
	}

	for _, account := range data.Accounts {
		address := common.HexToAddress(account.Address)
		accAddress := sdk.AccAddress(address.Bytes())

		// check that the account is actually found in the account keeper
		acc := accountKeeper.GetAccount(ctx, accAddress)
		if acc == nil {
			panic(fmt.Errorf("account not found for address %s", account.Address))
		}

		code := common.Hex2Bytes(account.Code)
		codeHash := crypto.Keccak256Hash(code).Bytes()

		if !types.IsEmptyCodeHash(codeHash) {
			k.SetCodeHash(ctx, address.Bytes(), codeHash)
		}

		if len(code) != 0 {
			k.SetCode(ctx, codeHash, code)
		}

		for _, storage := range account.Storage {
			k.SetState(ctx, address, common.HexToHash(storage.Key), common.HexToHash(storage.Value).Bytes())
		}
	}

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports genesis state of the EVM module
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	var ethGenAccounts []types.GenesisAccount
	k.IterateContracts(ctx, func(address common.Address, codeHash common.Hash) (stop bool) {
		storage := k.GetAccountStorage(ctx, address)

		genAccount := types.GenesisAccount{
			Address: address.String(),
			Code:    common.Bytes2Hex(k.GetCode(ctx, codeHash)),
			Storage: storage,
		}

		ethGenAccounts = append(ethGenAccounts, genAccount)
		return false
	})

	return &types.GenesisState{
		Accounts: ethGenAccounts,
		Params:   k.GetParams(ctx),
	}
}
