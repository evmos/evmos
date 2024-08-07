// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package contracts

import (
<<<<<<< HEAD
	contractutils "github.com/evmos/evmos/v19/contracts/utils"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
=======
	contractutils "github.com/evmos/evmos/v19/contracts/utils"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
>>>>>>> main
)

func LoadCounterContract() (evmtypes.CompiledContract, error) {
	return contractutils.LoadContractFromJSONFile("Counter.json")
}
