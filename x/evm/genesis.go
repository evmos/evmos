// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evm

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/evmos/evmos/v18/utils"
	"github.com/evmos/evmos/v18/x/evm/keeper"
	"github.com/evmos/evmos/v18/x/evm/types"
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

		// check that the EVM balance the matches the account balance
		acc := accountKeeper.GetAccount(ctx, accAddress)
		if acc == nil {
			panic(fmt.Errorf("account not found for address %s", account.Address))
		}

		code := common.Hex2Bytes(account.Code)
		codeHash := crypto.Keccak256Hash(code)

		// TODO: I think this can be removed now that the EVM state is detached from the account keeper state
		// This basically was cross-checking that both are in sync.
		//
		// // we ignore the empty Code hash checking, see ethermint PR#1234
		// if len(account.Code) != 0 && !bytes.Equal(ethAcct.GetCodeHash().Bytes(), codeHash.Bytes()) {
		// 	s := "the evm state code doesn't match with the codehash\n"
		// 	panic(fmt.Sprintf("%s account: %s , evm state codehash: %v, ethAccount codehash: %v, evm state code: %s\n",
		// 		s, account.Address, codeHash, ethAcct.GetCodeHash(), account.Code))
		// }

		// TODO: Do we need to add the code hash to the genesis accounts too?
		//
		// TODO: what is the significance of the code hash? Why do both need to be stored and not just
		// the code related to the account?
		k.SetCodeHash(ctx, address, codeHash)
		k.SetCode(ctx, codeHash.Bytes(), code)

		for _, storage := range account.Storage {
			k.SetState(ctx, address, common.HexToHash(storage.Key), common.HexToHash(storage.Value).Bytes())
		}
	}

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports genesis state of the EVM module
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper, ak types.AccountKeeper) *types.GenesisState {
	var ethGenAccounts []types.GenesisAccount
	ak.IterateAccounts(ctx, func(account sdk.AccountI) bool {
		acc, ok := account.(*authtypes.BaseAccount)
		if !ok {
			return false
		}

		address, err := utils.Bech32ToHexAddr(acc.Address)
		if err != nil {
			return false
		}

		codeHash := k.GetCodeHash(ctx, address)
		// TODO: this is never true (I think)
		if !types.BytesAreEmptyCodeHash(codeHash.Bytes()) {
			// only store smart contracts in the EVM genesis state
			return false
		}

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
