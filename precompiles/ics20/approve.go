// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ics20

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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

<<<<<<< HEAD
	// If one of the allocations contains a non-existing channel, throw and error
	for _, allocation := range transferAuthz.Allocations {
		found := p.channelKeeper.HasChannel(ctx, allocation.SourcePort, allocation.SourceChannel)
		if !found {
			return nil, errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", allocation.SourcePort, allocation.SourceChannel)
		}
	}

	// Only the origin can approve a transfer to the grantee address
	expiration := ctx.BlockTime().Add(p.ApprovalExpiration).UTC()
	if err = p.AuthzKeeper.SaveGrant(ctx, grantee.Bytes(), origin.Bytes(), transferAuthz, &expiration); err != nil {
		return nil, err
	}

	// Emit the IBC transfer authorization event
	allocation := transferAuthz.Allocations[0]
	if err = p.EmitIBCTransferAuthorizationEvent(
=======
	// Approve from ICS20 common module
	if err := Approve(
>>>>>>> ee3e7daf (impv(ics20): Common approval methods and tests refactor (#1849))
		ctx,
		p.AuthzKeeper,
		p.channelKeeper,
		p.Address(),
		grantee,
		origin,
<<<<<<< HEAD
		allocation.SourcePort,
		allocation.SourceChannel,
		allocation.SpendLimit,
=======
		p.ApprovalExpiration,
		transferAuthz,
		p.ABI.Events[authorization.EventTypeIBCTransferAuthorization],
		stateDB,
>>>>>>> ee3e7daf (impv(ics20): Common approval methods and tests refactor (#1849))
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
	if err := Revoke(
		ctx,
		p.AuthzKeeper,
		p.Address(),
		grantee,
		origin,
		p.ABI.Events[authorization.EventTypeIBCTransferAuthorization],
		stateDB,
	); err != nil {
		return nil, err
	}
<<<<<<< HEAD

	if err = p.EmitIBCRevokeAuthorizationEvent(ctx, stateDB, grantee, origin); err != nil {
		return nil, err
	}

=======
>>>>>>> ee3e7daf (impv(ics20): Common approval methods and tests refactor (#1849))
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
	if err := IncreaseAllowance(
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
	if err := DecreaseAllowance(
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
