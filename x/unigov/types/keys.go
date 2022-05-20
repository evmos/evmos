package types

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// ModuleName defines the module name
	ModuleName = "unigov"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_unigov"
)

var ModuleAddress commmon.Address

func init() {
	ModuleAddress = common.Bytes2Address(authtypes.NewModuleAddres(ModuleName).Bytes())
}

func KeyPrefix(p string) []byte {
	return []byte(p)
}
