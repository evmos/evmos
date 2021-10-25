package keeper

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tharsis/ethermint/server/config"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/tharsis/evmos/x/intrarelayer/types"
	"github.com/tharsis/evmos/x/intrarelayer/types/contracts"
)

type ERC20Data struct {
	Name     string
	Symbol   string
	Decimals uint8
}

func (k Keeper) CallEVM(ctx sdk.Context, abi abi.ABI, contract common.Address, method string, args ...interface{}) (*evmtypes.MsgEthereumTxResponse, error) {
	k.evmKeeper.WithContext(ctx)

	// pack and call method using the given args
	payload, err := abi.Pack(method, args...)

	if err != nil {
		return nil, sdkerrors.Wrap(
			types.ErrWritingEthTxPayload,
			sdkerrors.Wrap(err, "failed to create transaction payload").Error(),
		)
	}

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
		payload,
		ethtypes.AccessList{}, // AccessList
		true,                  // checkNonce
	)

	res, err := k.evmKeeper.ApplyMessage(msg, evmtypes.NewNoOpTracer(), true)
	if err != nil {
		return nil, err
	}

	if res.Failed() {
		return nil, fmt.Errorf("contract call failed: method '%s' %s, %s", method, contract, res.VmError)
	}

	return res, nil
}

func (k Keeper) QueryERC20(ctx sdk.Context, contract common.Address) (ERC20Data, error) {
	erc20 := contracts.ERC20BurnableContract.ABI
	ret := ERC20Data{}

	// Name
	res, err := k.CallEVM(ctx, erc20, contract, "name")
	if err != nil {
		return ret, err
	}

	nameResp := struct {
		Name string
	}{}
	err = erc20.UnpackIntoInterface(&nameResp, "name", res.Ret)
	if err != nil {
		return ret, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unpack name: %s", err.Error())
	}

	// Symbol
	res, err = k.CallEVM(ctx, erc20, contract, "symbol")
	if err != nil {
		return ret, err
	}

	symbolResp := struct {
		Name string
	}{}
	err = erc20.UnpackIntoInterface(&symbolResp, "symbol", res.Ret)
	if err != nil {
		return ret, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unpack symbol: %s", err.Error())
	}

	// Decimals
	res, err = k.CallEVM(ctx, erc20, contract, "decimals")
	if err != nil {
		return ret, err
	}

	decimalResp := struct {
		Value uint8
	}{}
	err = erc20.UnpackIntoInterface(&decimalResp, "decimals", res.Ret)
	if err != nil {
		return ret, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unpack decimals: %s", err.Error())
	}

	ret.Name = nameResp.Name
	ret.Symbol = symbolResp.Name
	ret.Decimals = decimalResp.Value
	return ret, nil
}
