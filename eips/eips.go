package eips

import (
	"github.com/ethereum/go-ethereum/params"
	"github.com/evmos/evmos/v18/x/evm/core/vm"
)

var ExtendedActivators = map[int]func(*vm.JumpTable){
	// EXTERNAL EIPs
	0000: enable0000,
	0001: enable0001,
	0002: enable0002,
}

var DefaultEnabledEIPs = []int64{
	0000,
	0001,
	0002,
}

// enable1884 applies EIP-1884 to the given jump table:
// Increate cost of contract creation.
func enable0000(jt *vm.JumpTable) {
	// Gas cost changes
	jt[vm.CREATE].SetConstantGas(params.CreateGas * 2)
	jt[vm.CREATE2].SetConstantGas(params.CreateGas * 2)
}

// Update call gas costs
func enable0001(jt *vm.JumpTable) {
	// Gas cost changes
	jt[vm.CALL].SetConstantGas(params.CallGasEIP150 * 2)
}

// Update store gas costs
func enable0002(jt *vm.JumpTable) {
	// Gas cost changes
	jt[vm.SSTORE].SetConstantGas(params.SstoreSetGas * 2)
}
