// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package vm

import (
	"github.com/holiman/uint256"
)

// Gas costs
const (
	GasQuickStep   uint64 = 2
	GasFastestStep uint64 = 3
	GasFastStep    uint64 = 5
	GasMidStep     uint64 = 8
	GasSlowStep    uint64 = 10
	GasExtStep     uint64 = 20
)

// callGas returns the actual gas cost of the call.
//
// The cost of gas was changed during the homestead price change HF.
// As part of EIP 150 (TangerineWhistle), the returned gas is gas - base * 63 / 64.
func callGas(availableGas, base uint64, callCost *uint256.Int) (uint64, error) {
	availableGas -= base
	gas := availableGas - availableGas/64
	// If the bit length exceeds 64 bit we know that the newly calculated "gas" for EIP150
	// is smaller than the requested amount. Therefore we return the new gas instead
	// of returning an error.
	if !callCost.IsUint64() || gas < callCost.Uint64() {
		return gas, nil
	}

	if !callCost.IsUint64() {
		return 0, ErrGasUintOverflow
	}

	return callCost.Uint64(), nil
}
