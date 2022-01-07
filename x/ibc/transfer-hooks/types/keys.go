package types

import "fmt"

const (
	// ModuleName defines the transfer hooks module name
	ModuleName = "transfer-hooks"

	// StoreKey is the store key string for IBC transfer hooks module
	StoreKey = ModuleName

	// RouterKey is the message route for IBC transfer hooks module
	RouterKey = ModuleName

	Version = "transfer_hooks-1"
)

const (
	prefixTransferHooksEnabled = iota + 1
)

// PrefixKeyTransferHooksEnabled is the key prefix for storing transfer hooks enabled flag
var PrefixKeyTransferHooksEnabled = []byte{prefixTransferHooksEnabled}

// KeyTransferHooksEnabled returns the key that stores a flag to determine if transfer hooks logic should
// be enabled for the given port and channel identifiers.
func KeyTransferHooksEnabled(portID, channelID string) []byte {
	return []byte(fmt.Sprintf("%s/%s", portID, channelID))
}
