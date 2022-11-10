package types

const (
	// DefaultSendEnabled enabled
	DefaultSendEvmTxEnabled = true
	// DefaultReceiveEnabled enabled
	DefaultReceiveEvmTxEnabled = true
)

var (
	// KeySendEvmTxEnabled is store's key for SendEvmTxEnabled Params
	KeySendEvmTxEnabled = []byte("SendEvmTxEnabled")
	// KeyReceiveEnabled is store's key for ReceiveEvmTxEnabled Params
	KeyReceiveEvmTxEnabled = []byte("ReceiveEvmTxEnabled")
)

// TODO: this should be in proto file
type Params struct {
	// send_enabled enables or disables all cross-chain token transfers from this
	// chain.
	SendEvmTxEnabled bool `protobuf:"varint,1,opt,name=send_evm_tx_enabled,json=sendEvmTxEnabled,proto3" json:"send_evm_tx_enabled,omitempty" yaml:"send_evm_tx_enabled"`
	// receive_enabled enables or disables all cross-chain token transfers to this
	// chain.
	ReceiveEvmTxEnabled bool `protobuf:"varint,2,opt,name=receive_evm_tx_enabled,json=receiveEvmTxEnabled,proto3" json:"receive_evm_tx_enabled,omitempty" yaml:"receive_enabled"`
}
