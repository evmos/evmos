// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/precompiles/authorization"
	"github.com/evmos/evmos/v14/precompiles/ics20"
)

// Approve implements the ICS20 approve transactions.
func (p Precompile) Approve(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	grantee, transferAuthz, err := NewTransferAuthorization(args)
	if err != nil {
		return nil, err
	}

	if err := ics20.Approve(
		ctx,
		p.AuthzKeeper,
		p.channelKeeper,
		p.Address(),
		grantee,
		origin,
		p.ApprovalExpiration,
		transferAuthz,
		p.ABI.Events[authorization.EventTypeIBCTransferAuthorization],
		stateDB,
	); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// IncreaseAllowance implements the ICS20 increase allowance transactions specifically for
// the Osmosis channel.
func (p Precompile) IncreaseAllowance(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	grantee, denom, amount, err := checkAllowanceArgs(args)
	if err != nil {
		return nil, err
	}

	if err := ics20.IncreaseAllowance(
		ctx,
		p.AuthzKeeper,
		p.Address(),
		grantee,
		origin,
		transfertypes.PortID,
		OsmosisChannelID,
		denom,
		amount,
		p.ABI.Events[authorization.EventTypeIBCTransferAuthorization],
		stateDB,
	); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// DecreaseAllowance implements the ICS20 decrease allowance transactions specifically for
// the Osmosis channel.
func (p Precompile) DecreaseAllowance(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	grantee, denom, amount, err := checkAllowanceArgs(args)
	if err != nil {
		return nil, err
	}

	if err := ics20.DecreaseAllowance(
		ctx,
		p.AuthzKeeper,
		p.Address(),
		grantee,
		origin,
		transfertypes.PortID,
		OsmosisChannelID,
		denom,
		amount,
		p.ABI.Events[authorization.EventTypeIBCTransferAuthorization],
		stateDB,
	); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// Revoke implements the ICS20 revoke transactions.
func (p Precompile) Revoke(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	grantee, ok := args[0].(common.Address)
	if !ok || grantee == (common.Address{}) {
		return nil, fmt.Errorf(authorization.ErrInvalidGrantee, args[0])
	}

	if err := ics20.Revoke(
		ctx,
		p.AuthzKeeper,
		p.Address(),
		grantee,
		origin,
		p.ABI.Events[authorization.EventTypeRevokeIBCTransferAuthorization],
		stateDB,
	); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}
