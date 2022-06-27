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
	prefixDeployer
	prefixWithdrawer
)

// KVStore key prefixes
var (
	KeyPrefixFee        = []byte{prefixFee}
	KeyPrefixDeployer   = []byte{prefixDeployer}
	KeyPrefixWithdrawer = []byte{prefixWithdrawer}
)

// GetKeyPrefixDeployerFees returns the KVStore key prefix for storing
// registered fee contract for a deployer
func GetKeyPrefixDeployerFees(deployerAddress sdk.AccAddress) []byte {
	return append(KeyPrefixDeployer, deployerAddress.Bytes()...)
}

// GetKeyPrefixWithdrawerFees returns the KVStore key prefix for storing
// registered fee contract for a deployer
func GetKeyPrefixWithdrawerFees(withdrawerAddress sdk.AccAddress) []byte {
	return append(KeyPrefixWithdrawer, withdrawerAddress.Bytes()...)
}
