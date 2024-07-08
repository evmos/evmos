package testdata

import (
	contractutils "github.com/evmos/evmos/v18/contracts/utils"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

func LoadCounterContract() (evmtypes.CompiledContract, error) {
	return contractutils.LoadContractFromJSONFile("Counter.json")
}

func LoadCounterFactoryContract() (evmtypes.CompiledContract, error) {
	return contractutils.LoadContractFromJSONFile("CounterFactory.json")
}
