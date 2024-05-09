package keeper

import "github.com/ethereum/go-ethereum/core"

func IsTransferCall(msg core.Message) bool {
	return msg.To() != nil && msg.Data() == nil
}
