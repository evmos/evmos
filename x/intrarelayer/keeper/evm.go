package keeper

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/tharsis/ethermint/server/config"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/tharsis/evmos/x/intrarelayer/types"
	"github.com/tharsis/evmos/x/intrarelayer/types/contracts"
)

func (k Keeper) CallEVMWithPayload(ctx sdk.Context, contract common.Address, transferData []byte) (*evmtypes.MsgEthereumTxResponse, error) {
	k.evmKeeper.WithContext(ctx)

	nonce, err := k.accountKeeper.GetSequence(ctx, types.ModuleAddress.Bytes())
	if err != nil {
		return nil, err
	}

	msg := ethtypes.NewMessage(
		types.ModuleAddress,
		&contract,
		nonce,
		big.NewInt(0),        // amount
		config.DefaultGasCap, // gasLimit
		big.NewInt(0),        // gasFeeCap
		big.NewInt(0),        // gasTipCap
		big.NewInt(0),        // gasPrice
		transferData,
		ethtypes.AccessList{}, // AccessList
		true,                  // checkNonce
	)

	res, err := k.evmKeeper.ApplyMessage(msg, evmtypes.NewNoOpTracer(), true)
	if err != nil {
		return nil, err
	}

	k.evmKeeper.SetNonce(types.ModuleAddress, nonce+1)

	if res.Failed() {
		return nil, fmt.Errorf("%s", res.VmError)
	}

	return res, nil
}

func (k Keeper) DeployToEVMWithPayload(ctx sdk.Context, transferData []byte) (*evmtypes.MsgEthereumTxResponse, error) {
	k.evmKeeper.WithContext(ctx)

	nonce, err := k.accountKeeper.GetSequence(ctx, types.ModuleAddress.Bytes())
	if err != nil {
		return nil, err
	}

	access := ethtypes.AccessList{}
	k.evmKeeper.AddAddressToAccessList(types.ModuleAddress)
	msg := ethtypes.NewMessage(
		types.ModuleAddress,
		nil,
		nonce,
		big.NewInt(0),        // amount
		config.DefaultGasCap, // gasLimit
		big.NewInt(0),        // gasFeeCap
		big.NewInt(0),        // gasTipCap
		big.NewInt(0),        // gasPrice
		transferData,
		access, // AccessList
		true,   // checkNonce
	)

	res, err := k.evmKeeper.ApplyMessage(msg, evmtypes.NewNoOpTracer(), true)
	if err != nil {
		return nil, err
	}

	if res.Failed() {
		return nil, fmt.Errorf("%s", res.VmError)
	}

	return res, nil
}

func (k Keeper) CallEVM(ctx sdk.Context, abi abi.ABI, contract common.Address, method string, args ...interface{}) (*evmtypes.MsgEthereumTxResponse, error) {
	// pack and call method using the given args
	payload, err := abi.Pack(method, args...)
	if err != nil {
		return nil, sdkerrors.Wrap(
			types.ErrWritingEthTxPayload,
			sdkerrors.Wrap(err, "failed to create transaction payload").Error(),
		)
	}

	resp, err := k.CallEVMWithPayload(ctx, contract, payload)
	if err != nil {
		return nil, fmt.Errorf("contract call failed: method '%s' %s, %s", method, contract, err)
	}
	return resp, nil
}

func (k Keeper) QueryERC20(ctx sdk.Context, contract common.Address) (types.ERC20Data, error) {
	var (
		nameRes    types.ERC20StringResponse
		symbolRes  types.ERC20StringResponse
		decimalRes types.ERC20Uint8Response
	)

	erc20 := contracts.ERC20BurnableContract.ABI

	// Name
	res, err := k.CallEVM(ctx, erc20, contract, "name")
	if err != nil {
		return types.ERC20Data{}, err
	}

	if err := erc20.UnpackIntoInterface(&nameRes, "name", res.Ret); err != nil {
		return types.ERC20Data{}, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unpack name: %s", err.Error())
	}

	// Symbol
	res, err = k.CallEVM(ctx, erc20, contract, "symbol")
	if err != nil {
		return types.ERC20Data{}, err
	}

	if err = erc20.UnpackIntoInterface(&symbolRes, "symbol", res.Ret); err != nil {
		return types.ERC20Data{}, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unpack symbol: %s", err.Error())
	}

	// Decimals
	res, err = k.CallEVM(ctx, erc20, contract, "decimals")
	if err != nil {
		return types.ERC20Data{}, err
	}

	if err := erc20.UnpackIntoInterface(&decimalRes, "decimals", res.Ret); err != nil {
		return types.ERC20Data{}, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unpack decimals: %s", err.Error())
	}

	return types.NewERC20Data(nameRes.Value, symbolRes.Value, decimalRes.Value), nil
}

func (k Keeper) ExecuteEVM(ctx sdk.Context, contractAddr, from common.Address, transferData []byte) ([]byte, error) {
	params := k.evmKeeper.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(k.evmKeeper.ChainID())
	// NOTE: pass in an empty coinbase address and nil tracer as we don't need them for the check below
	cfg := &evmtypes.EVMConfig{
		ChainConfig: ethCfg,
		Params:      params,
		CoinBase:    common.Address{},
		BaseFee:     big.NewInt(0),
	}
	msg := k.createModuleTx(&contractAddr, from, transferData)

	vmConfig := k.evmKeeper.VMConfig(msg, cfg.Params, evmtypes.NewNoOpTracer())
	evm := k.evmKeeper.NewEVM(msg, cfg, evmtypes.NewNoOpTracer())
	interpreter := vm.NewEVMInterpreter(evm, vmConfig)

	// Initialize a new contract and set the code that is to be used by the EVM.
	// The contract is a scoped environment for this execution context only.
	code := evm.StateDB.GetCode(contractAddr)
	if len(code) == 0 {
		// Invalid contract address
		return nil, fmt.Errorf("invalid contract address")
	}

	// TODO: define gas value
	gas := uint64(2000000)
	addrCopy := contractAddr
	contract := vm.NewContract(vm.AccountRef(from), vm.AccountRef(contractAddr), new(big.Int), gas)
	contract.SetCallCode(&addrCopy, evm.StateDB.GetCodeHash(addrCopy), code)
	ret, err := interpreter.Run(contract, transferData, false)
	if err != nil {
		return nil, err
	}

	// TODO: validate ret?

	return ret, err
}

func (k Keeper) createModuleTx(contractAddr *common.Address, from common.Address, transferData []byte) ethtypes.Message {
	msg := ethtypes.NewMessage(
		from, contractAddr,
		k.evmKeeper.GetNonce(from),
		big.NewInt(0),
		uint64(2000000),
		big.NewInt(0),
		big.NewInt(20000000),
		big.NewInt(20000000),
		transferData,
		ethtypes.AccessList{},
		false,
	)
	return msg
}

func (k Keeper) DeployEVM(ctx sdk.Context, contractAddr, from common.Address, transferData []byte) ([]byte, error) {
	params := k.evmKeeper.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(k.evmKeeper.ChainID())
	// NOTE: pass in an empty coinbase address and nil tracer as we don't need them for the check below
	cfg := &evmtypes.EVMConfig{
		ChainConfig: ethCfg,
		Params:      params,
		CoinBase:    common.Address{},
		BaseFee:     big.NewInt(0),
	}

	nonce, err := k.accountKeeper.GetSequence(ctx, types.ModuleAddress.Bytes())
	if err != nil {
		return nil, err
	}

	msg := ethtypes.NewMessage(
		from, nil,
		nonce,
		big.NewInt(0),
		uint64(2000000),
		big.NewInt(0),
		big.NewInt(20000000),
		big.NewInt(20000000),
		transferData,
		ethtypes.AccessList{},
		false,
	)

	rules := ethCfg.Rules(big.NewInt(ctx.BlockHeight()))
	k.evmKeeper.PrepareAccessList(msg.From(), msg.To(), vm.ActivePrecompiles(rules), msg.AccessList())

	vmConfig := k.evmKeeper.VMConfig(msg, cfg.Params, evmtypes.NewNoOpTracer())
	evm := k.evmKeeper.NewEVM(msg, cfg, evmtypes.NewNoOpTracer())
	interpreter := vm.NewEVMInterpreter(evm, vmConfig)

	evm.StateDB.CreateAccount(contractAddr)
	evm.Context.Transfer(evm.StateDB, from, contractAddr, big.NewInt(0))

	// TODO: define gas value
	gas := uint64(3000000)
	addrCopy := contractAddr
	contract := vm.NewContract(vm.AccountRef(from), vm.AccountRef(contractAddr), new(big.Int), gas)
	contract.SetCallCode(&addrCopy, evm.StateDB.GetCodeHash(addrCopy), transferData)

	ret, err := interpreter.Run(contract, nil, false)
	if err != nil {
		return nil, err
	}

	// if the contract creation ran successfully and no errors were returned
	// calculate the gas required to store the code. If the code could not
	// be stored due to not enough gas set an error and let it be handled
	// by the error checking condition below.
	createDataGas := uint64(len(ret)) * 200
	if contract.UseGas(createDataGas) {
		evm.StateDB.SetCode(contractAddr, ret)
	}

	k.evmKeeper.SetNonce(types.ModuleAddress, nonce+1)
	// TODO: validate ret?

	return ret, err
}
