package types

// fees events
const (
	EventTypeRegisterContract      = "register_fee_contract"
	EventTypeCancelContract        = "cancel_fee_contract"
	EventTypeDistributeFeeContract = "distribute_fee_contract"

	AttributeKeyContract = "contract"
	AttributeKeyEpochs   = "epochs"
)
