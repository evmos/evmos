package testdata

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"
	"errors"

	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

var (
	//go:embed ERC20Contract.json
	erc20JSON []byte

	//go:embed TestMessageCall.json
	testMessageCallJSON []byte
)

func LoadERC20Contract() (evmtypes.CompiledContract, error) {
	var ERC20Contract evmtypes.CompiledContract
	err := json.Unmarshal(erc20JSON, &ERC20Contract)
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	if len(ERC20Contract.Bin) == 0 {
		return evmtypes.CompiledContract{}, errors.New("load contract failed")
	}

	return ERC20Contract, nil
}

func LoadMessageCallContract() (evmtypes.CompiledContract, error) {
	var messageCallContract evmtypes.CompiledContract
	err := json.Unmarshal(testMessageCallJSON, &messageCallContract)
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	if len(messageCallContract.Bin) == 0 {
		return evmtypes.CompiledContract{}, errors.New("load contract failed")
	}

	return messageCallContract, nil
}
