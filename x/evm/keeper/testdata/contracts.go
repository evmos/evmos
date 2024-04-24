package testdata

import (
	_ "embed" // embed compiled smart contract
	contractutils "github.com/evmos/evmos/v16/contracts/utils"
	"os"

	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

func LoadERC20Contract() (evmtypes.CompiledContract, error) {
	erc20JSON, err := os.ReadFile("ERC20Contract.json")
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	return contractutils.LoadContract(erc20JSON)
}

func LoadMessageCallContract() (evmtypes.CompiledContract, error) {
	messageCallJSON, err := os.ReadFile("MessageCallContract.json")
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	return contractutils.LoadContract(messageCallJSON)
}
