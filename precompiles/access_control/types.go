// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package accesscontrol

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

func ParseRoleArgs(args []interface{}) (common.Hash, common.Address, error) {
	if len(args) != 2 {
		return common.Hash{}, common.Address{}, fmt.Errorf("invalid number of arguments")
	}
	role, ok := args[0].(common.Hash)
	if !ok {
		return common.Hash{}, common.Address{}, fmt.Errorf("invalid role argument")
	}

	account, ok := args[1].(common.Address)
	if !ok {
		return common.Hash{}, common.Address{}, fmt.Errorf("invalid account argument")
	}

	return role, account, nil
}
