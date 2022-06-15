package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// constants
const (
	// module name
	ModuleName = "fees"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName
)

// prefix bytes for the fees persistent store
const (
	prefixFee = iota + 1
	prefixFeeWithdrawal
	prefixDeployerFees
)

// KVStore key prefixes
var (
	KeyPrefixFee           = []byte{prefixFee}
	KeyPrefixFeeWithdrawal = []byte{prefixFeeWithdrawal}
	KeyPrefixDeployerFees  = []byte{prefixDeployerFees}
)

// GetKeyPrefixDeployerFees returns the KVStore key prefix for storing
// registered fee infos for a deployer
func GetKeyPrefixDeployerFees(deployerAddress sdk.AccAddress) []byte {
	return append(KeyPrefixDeployerFees, deployerAddress.Bytes()...)
}
