package eips

import (
	"github.com/evmos/evmos/v18/x/evm/core/vm"
	"github.com/evmos/evmos/v18/x/evm/types"
)

func ExtendEips(eips map[int]func(*vm.JumpTable)) {
	vm.ExtendActivators(eips)
}

func UpdateDefaultExtraEIPs(eips []int64) {
	types.DefaultExtraEIPs = append(types.DefaultExtraEIPs, eips...)
}
