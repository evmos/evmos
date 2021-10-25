package contracts

import (
	"encoding/json"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	evmkeeper "github.com/tharsis/ethermint/x/evm/keeper"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

func GetERC20Property(k *evmkeeper.Keeper, ctx sdk.Context, to common.Address, property string) (interface{}, error) {
	erc20 := ERC20BurnableContract
	from := types.ModuleAddress
	return abiGetProperty(k, ctx, from, to, erc20.ABI, property)

}

func abiGetProperty(k *evmkeeper.Keeper, ctx sdk.Context, from common.Address, to common.Address, contract abi.ABI, property string) (interface{}, error) {
	ctorArgs, err := contract.Pack(property)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to create ABI for erc20: %s", err.Error())
	}

	encoded_msg := (*hexutil.Bytes)(&ctorArgs)
	// TODO: get gas price
	gas := hexutil.Uint64(40000)
	gasPrice := (*hexutil.Big)(big.NewInt(0))

	args := &evmtypes.TransactionArgs{
		From:     &from,
		To:       &to,
		Gas:      &gas,
		GasPrice: gasPrice,
		Data:     encoded_msg,
	}

	bz, err := json.Marshal(&args)
	if err != nil {
		return nil, err
	}

	// TODO: set correct gas price
	req := evmtypes.EthCallRequest{
		Args:   bz,
		GasCap: 100000,
	}

	resp, err := k.EthCall(sdk.WrapSDKContext(ctx), &req)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to send eth_call: %s", err.Error())
	}

	// this calls should be enough for getting values
	// contract := NewContract(caller, AccountRef(addrCopy), value, gas)
	// contract.SetCallCode(&addrCopy, evm.StateDB.GetCodeHash(addrCopy), code)
	// ret, err = evm.interpreter.Run(contract, input, false)

	res, err := contract.Unpack(property, resp.Ret)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unpack property response: %s", err.Error())
	}
	if len(res) != 1 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to get property, response array must be 1 element")
	}
	return res[0], nil
}
