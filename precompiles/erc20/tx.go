// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package erc20

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// TransferMethod defines the ABI method name for the IERC20 transfer
	// transaction.
	TransferMethod = "transfer"
	// TransferFromMethod defines the ABI method name for the IERC20 transferFrom
	// transaction.
	TransferFromMethod = "transferFrom"
)

func (p Precompile) Transfer(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	from := contract.CallerAddress
	to, amount, err := ParseTransferArgs(args)
	if err != nil {
		return nil, err
	}

	return p.transfer(ctx, contract, stateDB, method, from, to, amount)
}

func (p Precompile) TransferFrom(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	from, to, amount, err := ParseTransferFromArgs(args)
	if err != nil {
		return nil, err
	}

	return p.transfer(ctx, contract, stateDB, method, from, to, amount)
}

func (p Precompile) transfer(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	from, to common.Address,
	amount *big.Int,
) (data []byte, err error) {
	coins := sdk.Coins{{Denom: p.tokenPair.Denom, Amount: sdk.NewIntFromBigInt(amount)}}

	msg := banktypes.NewMsgSend(from.Bytes(), to.Bytes(), coins)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	spender := sdk.AccAddress(contract.CallerAddress.Bytes()) // aka. grantee

	if sdk.AccAddress(from.Bytes()).Equals(spender) {
		msgSrv := bankkeeper.NewMsgServerImpl(p.bankKeeper)
		_, err = msgSrv.Send(sdk.WrapSDKContext(ctx), msg)
	} else {
		_, err = p.authzKeeper.DispatchActions(ctx, spender, []sdk.Msg{msg})
	}

	if err != nil {
		// TODO: pack failure bool?
		bz, _ := method.Outputs.Pack(false)
		return bz, err
	}

	if err := p.EmitTransferEvent(ctx, stateDB, from, to, amount); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}
