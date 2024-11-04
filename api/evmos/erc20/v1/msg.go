// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package erc20v1

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	protov2 "google.golang.org/protobuf/proto"
)

// GetSigners gets the signer's address from the Ethereum tx signature
func GetSigners(msg protov2.Message) ([][]byte, error) {
	msgConvERC20, ok := msg.(*MsgConvertERC20)
	if !ok {
		return nil, fmt.Errorf("invalid type, expected MsgConvertERC20 and got %T", msg)
	}

	// The sender on the msg is a hex address
	sender := common.HexToAddress(msgConvERC20.Sender)

	return [][]byte{sender.Bytes()}, nil
}
