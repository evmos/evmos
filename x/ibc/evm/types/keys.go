package types

const (
	// ModuleName defines the IBC EVM transaction name
	ModuleName = "ibc-evm-tx"

	// Version defines the current version the IBC tranfer
	// module supports
	Version = "ibc-evm-tx-1"

	// PortID is the default port id that transfer module binds to
	PortID = "evm-tx"

	// StoreKey is the store key string for IBC transfer
	StoreKey = ModuleName

	// RouterKey is the message route for IBC transfer
	RouterKey = ModuleName

	// QuerierRoute is the querier route for IBC transfer
	QuerierRoute = ModuleName
)

var (
	// PortKey defines the key to store the port ID in store
	PortKey = []byte{0x01}
)
