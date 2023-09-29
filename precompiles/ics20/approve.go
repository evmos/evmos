// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ics20

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/precompiles/authorization"
)

// Approve implements the ICS20 approve transactions.
func (p Precompile) Approve(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	grantee, transferAuthz, err := NewTransferAuthorization(method, args)
	if err != nil {
		return nil, err
	}

	// Approve from ICS20 common module
	if err := authorization.Approve(
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

// Revoke implements the ICS20 authorization revoke transactions.
func (p Precompile) Revoke(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	grantee, err := checkRevokeArgs(args)
	if err != nil {
		return nil, err
	}

	// Revoke from ICS20 common module
	if err := authorization.Revoke(
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

// IncreaseAllowance implements the ICS20 increase allowance transactions.
func (p Precompile) IncreaseAllowance(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	grantee, sourcePort, sourceChannel, denom, amount, err := checkAllowanceArgs(args)
	if err != nil {
		return nil, err
	}

	// IncreaseAllowance from ICS20 common module
	if err := authorization.IncreaseAllowance(
		ctx,
		p.AuthzKeeper,
		p.Address(),
		grantee,
		origin,
		sourcePort,
		sourceChannel,
		denom,
		amount,
		p.ABI.Events[authorization.EventTypeIBCTransferAuthorization],
		stateDB,
	); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// DecreaseAllowance implements the ICS20 decrease allowance transactions.
func (p Precompile) DecreaseAllowance(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	grantee, sourcePort, sourceChannel, denom, amount, err := checkAllowanceArgs(args)
	if err != nil {
		return nil, err
	}

	// DecreaseAllowance from ICS20 common module
	if err := authorization.DecreaseAllowance(
		ctx,
		p.AuthzKeeper,
		p.Address(),
		grantee,
		origin,
		sourcePort,
		sourceChannel,
		denom,
		amount,
		p.ABI.Events[authorization.EventTypeIBCTransferAuthorization],
		stateDB,
	); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// AcceptGrant implements the ICS20 accept grant.
func AcceptGrant(
	ctx sdk.Context,
	caller, origin common.Address,
	msg *transfertypes.MsgTransfer,
	authzAuthorization authz.Authorization,
) (*authz.AcceptResponse, error) {
	transferAuthz, ok := authzAuthorization.(*transfertypes.TransferAuthorization)
	if !ok {
		return nil, authz.ErrUnknownAuthorizationType
	}

	resp, err := transferAuthz.Accept(ctx, msg)
	if err != nil {
		return nil, err
	}

	if !resp.Accept {
		return nil, fmt.Errorf(authorization.ErrAuthzNotAccepted, caller, origin)
	}

	return &resp, nil
}

// UpdateGrant implements the ICS20 authz update grant.
func UpdateGrant(
	ctx sdk.Context,
	authzKeeper authzkeeper.Keeper,
	grantee, origin common.Address,
	expiration *time.Time,
	resp *authz.AcceptResponse,
) (err error) {
	if resp.Delete {
		err = authzKeeper.DeleteGrant(ctx, grantee.Bytes(), origin.Bytes(), authorization.TransferMsg)
	} else if resp.Updated != nil {
		err = authzKeeper.SaveGrant(ctx, grantee.Bytes(), origin.Bytes(), resp.Updated, expiration)
	}

	if err != nil {
		return err
	}

	return nil
}
