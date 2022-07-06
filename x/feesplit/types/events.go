package types

// feesplit events
const (
	EventTypeRegisterFeeSplit      = "register_fee_split"
	EventTypeCancelFeeSplit        = "cancel_fee_split"
	EventTypeUpdateFeeSplit        = "update_fee_split"
	EventTypeDistributeDevFeeSplit = "distribute_dev_fee_split"

	AttributeKeyContract          = "contract"
	AttributeKeyWithdrawerAddress = "withdrawer_address"
)
