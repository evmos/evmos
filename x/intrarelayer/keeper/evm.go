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
	nonce := k.evmKeeper.GetNonce(types.ModuleAddress)

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

	if res.Failed() {
		return nil, fmt.Errorf("%s", res.VmError)
	}

	return res, nil
}

func (k Keeper) DeployToEVMWithPayload(ctx sdk.Context, transferData []byte) (*evmtypes.MsgEthereumTxResponse, error) {
	k.evmKeeper.WithContext(ctx)
	nonce := k.evmKeeper.GetNonce(types.ModuleAddress)

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

	params := k.evmKeeper.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(k.evmKeeper.ChainID())
	rules := ethCfg.Rules(big.NewInt(ctx.BlockHeight()))

	k.evmKeeper.PrepareAccessList(msg.From(), msg.To(), vm.ActivePrecompiles(rules), msg.AccessList())

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
	erc20 := contracts.ERC20BurnableContract.ABI

	// Name
	res, err := k.CallEVM(ctx, erc20, contract, "name")
	if err != nil {
		return types.ERC20Data{}, err
	}

	nameResp := types.NewERC20StringResponse()
	err = erc20.UnpackIntoInterface(&nameResp, "name", res.Ret)
	if err != nil {
		return types.ERC20Data{}, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unpack name: %s", err.Error())
	}

	// Symbol
	res, err = k.CallEVM(ctx, erc20, contract, "symbol")
	if err != nil {
		return types.ERC20Data{}, err
	}

	symbolResp := types.NewERC20StringResponse()
	err = erc20.UnpackIntoInterface(&symbolResp, "symbol", res.Ret)
	if err != nil {
		return types.ERC20Data{}, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unpack symbol: %s", err.Error())
	}

	// Decimals
	res, err = k.CallEVM(ctx, erc20, contract, "decimals")
	if err != nil {
		return types.ERC20Data{}, err
	}

	decimalResp := types.NewERC20Uint8Response()
	err = erc20.UnpackIntoInterface(&decimalResp, "decimals", res.Ret)
	if err != nil {
		return types.ERC20Data{}, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unpack decimals: %s", err.Error())
	}

	return types.NewERC20Data(nameResp.Name, symbolResp.Name, decimalResp.Value), nil
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

	// Initialise a new contract and set the code that is to be used by the EVM.
	// The contract is a scoped environment for this execution context only.
	code := evm.StateDB.GetCode(contractAddr)
	if len(code) == 0 {
		// Invalid contract address
		return nil, fmt.Errorf("Invalid contract address")
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
