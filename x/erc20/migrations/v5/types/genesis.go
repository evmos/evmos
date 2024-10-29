// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v5types

import (
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/erc20/types"
)

var DefaultTokenPairs = []V5TokenPair{
	{
		Erc20Address:  types.WEVMOSContractMainnet,
		Denom:         utils.BaseDenom,
		Enabled:       true,
		ContractOwner: OWNER_MODULE,
	},
}
