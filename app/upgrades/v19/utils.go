// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v19

import (
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/ethereum/go-ethereum/common"
	evmostypes "github.com/evmos/evmos/v18/types"
)

var ignoredAddresses = []string{
	// XEN Crypto
	"0x2AB0e9e4eE70FFf1fB9D67031E44F6410170d00e",
	// XEN torrent
	"0x4c4CF206465AbFE5cECb3b581fa1b508Ec514692",
	// Related to XEN
	"0x9Ec1C3DcF667f2035FB4CD2eB42A1566fd54d2B7",
	// XENStake -> Symbol = COXENS
	"0xAF18644083151cf57F914CCCc23c42A1892C218e",
	// DBXen Token on Evmos
	"0xc418B123885d732ED042b16e12e259741863F723",
	// DBXen NFT
	"0xc741C0EC9d5DaD9e6aD481a3BE75295e7D85719B",
}

func getAccountBytecode(ak authkeeper.AccountKeeper, ctx sdk.Context, hexAddress string) common.Hash {
	acc, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	if err != nil {
		panic(err)
	}
	account := ak.GetAccount(ctx, acc)
	ethAccount, ok := account.(evmostypes.EthAccountI)
	if !ok {
		return common.Hash{}
	}
	return ethAccount.GetCodeHash()
}

func generateFilter(ak authkeeper.AccountKeeper, ctx sdk.Context) []common.Hash {
	byteCodesToFilter := []common.Hash{}
	byteCodesToFilter = append(byteCodesToFilter, getAccountBytecode(ak, ctx, "37282e677e4905e65e1506635f8ce637957da75e"))
	byteCodesToFilter = append(byteCodesToFilter, getAccountBytecode(ak, ctx, "Ff8cBBa9989FD10ba39Cc58aBe499eB9Bca14B4E"))
	// Y NOT PROXY
	byteCodesToFilter = append(byteCodesToFilter, getAccountBytecode(ak, ctx, "fdea18a83420cbbb9a516fb1da1873c0b2fbf521"))
	// COXEN Stake
	byteCodesToFilter = append(byteCodesToFilter, getAccountBytecode(ak, ctx, "000436A6B0097deb2c7Fcee1115b2781Bae86f50"))
	// 8k random accounts
	byteCodesToFilter = append(byteCodesToFilter, getAccountBytecode(ak, ctx, "c441d5e10faA165719520C9838e34A025FE98C21"))

	return byteCodesToFilter
}

func isAccountValid(addr string, code common.Hash, filter []common.Hash) bool {
	// Filter the smart contracts deployed by XEN
	if slices.Index(filter, code) != -1 {
		return false
	}
	// filter addresses
	return !slices.Contains(ignoredAddresses, addr)
}
