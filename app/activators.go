package app

import (
	"github.com/evmos/evmos/v20/app/eips"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

// EvmosActivators defines a map of opcode modifiers associated
// with a key defining the corresponding EIP.
var evmosActivators = map[string]func(*vm.JumpTable){
	"evmos_0": eips.Enable0000,
	"evmos_1": eips.Enable0001,
	"evmos_2": eips.Enable0002,
}
