package vm

import (
	"github.com/ethereum/go-ethereum/params"
)

// enable1884 applies EIP-1884 to the given jump table:
// Increate cost of contract creation.
func enable0000(jt *JumpTable) {
	// Gas cost changes
	jt[CREATE].constantGas = params.CreateGas * 2
	jt[CREATE2].constantGas = params.CreateGas * 2
}

// Update call gas costs
func enable0001(jt *JumpTable) {
	// Gas cost changes
	jt[CALL].constantGas = params.CallGasEIP150 * 10
}

// Update store gas costs
func enable0002(jt *JumpTable) {
	// Gas cost changes
	jt[SSTORE].constantGas = params.SstoreSetGas * 10
}
