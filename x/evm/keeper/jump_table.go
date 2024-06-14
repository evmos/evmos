package keeper

import (
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/evmos/evmos/v18/x/evm/types"
)

type JumpTableModifier struct {
	jt *vm.JumpTable
}

func NewJumptTableModifier(rules params.Rules) *JumpTableModifier {
	jt := vm.DefaultJumpTable(rules)
	return &JumpTableModifier{
		jt: jt,
	}
}

func (j *JumpTableModifier) GetVMJumpTable() *vm.JumpTable {
	return j.jt
}

func (j *JumpTableModifier) UpdateCustomOpcodes(updates []types.CustomOpCode) {
	for _, update := range updates {
		j.updateOpcodeConstantGas(update.OpCode, update.ConstantGas)
		// TODO we will have to have handle other opcode updates here
	}
}

func (j *JumpTableModifier) updateOpcodeConstantGas(opcode types.OpCode, gas uint64) {
	// We know this can't fail because opcodes are constraint by enum
	// and params validation
	op := vm.StringToOp(opcode.String())
	j.jt[op].UpdateConstantGas(gas)
}
