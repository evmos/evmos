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

func (k Keeper) CallEVM(ctx sdk.Context, abi abi.ABI, contract common.Address, method string, args ...interface{}) (*evmtypes.MsgEthereumTxResponse, error) {
	k.evmKeeper.WithContext(ctx)

	// pack and call method using the given args
	var payload []byte
	var err error
	payload, err = abi.Pack(method, args)

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

func (k Keeper) QueryERC20(ctx sdk.Context, contract common.Address) (string, string, uint32, error) {
	erc20 := contracts.ERC20BurnableContract.ABI

	res, err := k.CallEVM(ctx, erc20, contract, "name", nil)
	if err != nil {
		return "", "", 0, err
	}
	// Name
	// TODO: use UnpackIntoInterface and pass the struct instead of
	unpacked, err := erc20.Unpack("name", res.Ret)
	if err != nil {
		return "", "", 0, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unpack name: %s", err.Error())
	}
	if len(unpacked) != 1 {
		return "", "", 0, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to get property, response array must be 1 element")
	}
	name := unpacked[0].(string)

	_, err = k.CallEVM(ctx, erc20, contract, "symbol", nil)
	if err != nil {
		return "", "", 0, err
	}

	// Symbol
	// TODO: use UnpackIntoInterface and pass the struct instead of
	unpacked, err = erc20.Unpack("symbol", res.Ret)
	if err != nil {
		return "", "", 0, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unpack symbol: %s", err.Error())
	}
	if len(unpacked) != 1 {
		return "", "", 0, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to get property, response array must be 1 element")
	}
	symbol := unpacked[0].(string)

	// Decimals
	_, err = k.CallEVM(ctx, erc20, contract, "decimals", nil)
	if err != nil {
		return "", "", 0, err
	}

	unpacked, err = erc20.Unpack("decimals", res.Ret)
	if err != nil {
		return "", "", 0, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unpack decimals: %s", err.Error())
	}
	if len(unpacked) != 1 {
		return "", "", 0, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to get property, response array must be 1 element")
	}
	decimals := unpacked[0].(uint8)

	// TODO: return name, symbol, decimals, supply
	return name, symbol, uint32(decimals), nil
}
