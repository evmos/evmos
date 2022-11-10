package types

// IBC EVM transaction events
const (
	EventTypeTimeout      = "timeout"
	EventTypePacket       = "evm_tx_packet"
	EventTypeTransfer     = "ibc_evm_tx"
	EventTypeChannelClose = "channel_closed"

	AttributeKeyRefundSender = "refund_sender"
	AttributeKeyRefundAmount   = "refund_amount"
	AttributeKeyAckSuccess     = "success"
	AttributeKeyAck            = "acknowledgement"
	AttributeKeyAckError       = "error"
	AttributeKeyTraceHash      = "trace_hash"
	AttributeKeyMetadata       = "metadata"
)