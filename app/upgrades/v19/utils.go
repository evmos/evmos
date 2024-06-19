package v19

import (
	"fmt"
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/ethereum/go-ethereum/common"
	evmostypes "github.com/evmos/evmos/v18/types"
)

func GetAccountBytecode(ak authkeeper.AccountKeeper, ctx sdk.Context, hexAddress string) common.Hash {
	acc, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	if err != nil {
		panic(err)
	}
	account := ak.GetAccount(ctx, acc)
	ethAccount, ok := account.(evmostypes.EthAccountI)
	if !ok {
		// panic("Cant get bytecode")
		return common.Hash{}
	}
	return ethAccount.GetCodeHash()
}

func GenerateFilter(ak authkeeper.AccountKeeper, ctx sdk.Context) []common.Hash {
	byteCodesToFilter := []common.Hash{}
	byteCodesToFilter = append(byteCodesToFilter, GetAccountBytecode(ak, ctx, "37282e677e4905e65e1506635f8ce637957da75e"))
	byteCodesToFilter = append(byteCodesToFilter, GetAccountBytecode(ak, ctx, "Ff8cBBa9989FD10ba39Cc58aBe499eB9Bca14B4E"))
	// Y NOT PROXY
	byteCodesToFilter = append(byteCodesToFilter, GetAccountBytecode(ak, ctx, "fdea18a83420cbbb9a516fb1da1873c0b2fbf521"))
	// COXEN Stake
	byteCodesToFilter = append(byteCodesToFilter, GetAccountBytecode(ak, ctx, "000436A6B0097deb2c7Fcee1115b2781Bae86f50"))
	// 8k random accounts
	byteCodesToFilter = append(byteCodesToFilter, GetAccountBytecode(ak, ctx, "c441d5e10faA165719520C9838e34A025FE98C21"))

	fmt.Println(byteCodesToFilter)
	return byteCodesToFilter
}

func IsAccountValid(addr string, code common.Hash, filter []common.Hash) bool {
	// Filter the smart contracts deployed by XEN
	if slices.Index(filter, code) != -1 {
		return false
	}
	// XEN Crypto
	if addr == "0x2AB0e9e4eE70FFf1fB9D67031E44F6410170d00e" {
		return false
	}

	// XEN torrent
	if addr == "0x4c4CF206465AbFE5cECb3b581fa1b508Ec514692" {
		return false
	}

	// Related to XEN
	if addr == "0x9Ec1C3DcF667f2035FB4CD2eB42A1566fd54d2B7" {
		return false
	}

	// XENStake -> Symbol = COXENS
	if addr == "0xAF18644083151cf57F914CCCc23c42A1892C218e" {
		return false
	}

	// DBXen Token on Evmos
	if addr == "0xc418B123885d732ED042b16e12e259741863F723" {
		return false
	}

	// DBXen NFT
	if addr == "0xc741C0EC9d5DaD9e6aD481a3BE75295e7D85719B" {
		return false
	}

	return true
}
