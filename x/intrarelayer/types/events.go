package types

// intrarelayer events
const (
	EventTypeTokenLock            = "token_lock"
	EventTypeTokenUnlock          = "token_unlock"
	EventTypeMint                 = "mint"
	EventTypeConvertCoin          = "convert_coin"
	EventTypeBurn                 = "burn"
	EventTypeRegisterTokenPair    = "register_token_pair"
	EventTypeEnableTokenRelay     = "enable_token_relay"
	EventTypeUpdateTokenPairERC20 = "update_token_pair_erc20"

	AttributeKeyCosmosCoin = "cosmos_coin"
	AttributeKeyERC20Token = "erc20_token"
	AttributeKeyReceiver   = "receiver"
)
