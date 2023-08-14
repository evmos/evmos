// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ics20

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/precompiles/authorization"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
)

// TransferMsg is the ICS20 transfer message type.
var TransferMsg = sdk.MsgTypeURL(&transfertypes.MsgTransfer{})

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
		ctx,
		stateDB,
		grantee,
		origin,
		allocation.SourcePort,
		allocation.SourceChannel,
		allocation.SpendLimit,
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

	// NOTE: we do not need to check the expiration as it will return nil if both found or expired
	msgAuthz, _, err := authorization.CheckAuthzExists(ctx, p.AuthzKeeper, grantee, origin, TransferMsg)
	if err != nil {
		return nil, fmt.Errorf(authorization.ErrAuthzDoesNotExistOrExpired, grantee, origin)
	}

	// check that the stored authorization matches the transfer authorization
	if _, ok := msgAuthz.(*transfertypes.TransferAuthorization); !ok {
		return nil, authz.ErrUnknownAuthorizationType
	}

	if err = p.AuthzKeeper.DeleteGrant(ctx, grantee.Bytes(), origin.Bytes(), TransferMsg); err != nil {
		return nil, err
	}

	if err = p.EmitIBCRevokeAuthorizationEvent(ctx, stateDB, grantee, origin); err != nil {
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

	// NOTE: we do not need to check the expiration as it will return nil if both found or expired
	msgAuthz, expiration, err := authorization.CheckAuthzExists(ctx, p.AuthzKeeper, grantee, origin, TransferMsg)
	if err != nil {
		return nil, fmt.Errorf(authorization.ErrAuthzDoesNotExistOrExpired, grantee, origin)
	}

	// NOTE: we do not need to check the expiration as it will return nil if both found or expired
	transferAuthz, ok := msgAuthz.(*transfertypes.TransferAuthorization)
	if !ok {
		return nil, authz.ErrUnknownAuthorizationType
	}

	// Check if the allocations matches the arguments provided and returns the index of the allocation and coin found
	spendLimit, allocationIdx, err := checkAllocationExists(transferAuthz.Allocations, sourcePort, sourceChannel, denom)
	if err != nil {
		return nil, err
	}

	allowance := sdk.NewIntFromBigInt(amount)
	if _, overflow := cmn.SafeAdd(spendLimit.Amount, allowance); overflow {
		return nil, errors.New(cmn.ErrIntegerOverflow)
	}

	allowanceCoin := sdk.Coin{Denom: denom, Amount: allowance}

	transferAuthz.Allocations[allocationIdx].SpendLimit = transferAuthz.Allocations[allocationIdx].SpendLimit.Add(allowanceCoin)

	if err = p.AuthzKeeper.SaveGrant(ctx, grantee.Bytes(), origin.Bytes(), transferAuthz, expiration); err != nil {
		return nil, err
	}

	if err = authorization.EmitAllowanceChangeEvent(cmn.EmitEventArgs{
		Ctx:            ctx,
		StateDB:        stateDB,
		ContractAddr:   p.Address(),
		ContractEvents: p.ABI.Events,
		EventData: authorization.EventAllowanceChange{
			Granter: origin,
			Grantee: grantee,
			Values:  []*big.Int{amount},
			Methods: []string{TransferMsg},
		},
	}); err != nil {
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

	// NOTE: we do not need to check the expiration as it will return nil if both found or expired
	msgAuthz, expiration, err := authorization.CheckAuthzExists(ctx, p.AuthzKeeper, grantee, origin, TransferMsg)
	if err != nil {
		return nil, fmt.Errorf(authorization.ErrAuthzDoesNotExistOrExpired, grantee, origin)
	}

	transferAuthz, ok := msgAuthz.(*transfertypes.TransferAuthorization)
	if !ok {
		return nil, authz.ErrUnknownAuthorizationType
	}

	// Check if the allocations matches the arguments provided and returns the index of the allocation and spend limit found
	spendLimit, allocationIdx, err := checkAllocationExists(transferAuthz.Allocations, sourcePort, sourceChannel, denom)
	if err != nil {
		return nil, err
	}

	expense := sdk.NewIntFromBigInt(amount)

	if spendLimit.Amount.LT(expense) {
		return nil, fmt.Errorf(cmn.ErrNegativeAmount)
	}

	// Checking if the amount here is negative or zero and remove the coin from the spend limit otherwise
	// subtract from the allowance like normal
	allocation := transferAuthz.Allocations[allocationIdx]
	for i, coin := range allocation.SpendLimit {
		if coin.Denom != denom {
			continue
		}
		coinDiff := coin.Amount.Sub(expense)
		// Remove if it's negative or 0
		if !coinDiff.IsPositive() {
			allocation.SpendLimit = append(
				allocation.SpendLimit[:i],
				allocation.SpendLimit[i+1:]...)
		} else {
			allocation.SpendLimit[i].Amount = coinDiff
		}
	}
	transferAuthz.Allocations[allocationIdx] = allocation
	if err = p.AuthzKeeper.SaveGrant(ctx, grantee.Bytes(), origin.Bytes(), transferAuthz, expiration); err != nil {
		return nil, err
	}

	// NOTE: Using the new more generic event emitter that was created
	if err = authorization.EmitAllowanceChangeEvent(cmn.EmitEventArgs{
		Ctx:            ctx,
		StateDB:        stateDB,
		ContractAddr:   p.Address(),
		ContractEvents: p.ABI.Events,
		EventData: authorization.EventAllowanceChange{
			Granter: origin,
			Grantee: grantee,
			Values:  []*big.Int{amount},
			Methods: []string{TransferMsg},
		},
	}); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// AcceptGrant implements the ICS20 accept grant.
func (p Precompile) AcceptGrant(
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
func (p Precompile) UpdateGrant(
	ctx sdk.Context,
	grantee, origin common.Address,
	expiration *time.Time,
	resp *authz.AcceptResponse,
) (err error) {
	if resp.Delete {
		err = p.AuthzKeeper.DeleteGrant(ctx, grantee.Bytes(), origin.Bytes(), TransferMsg)
	} else if resp.Updated != nil {
		err = p.AuthzKeeper.SaveGrant(ctx, grantee.Bytes(), origin.Bytes(), resp.Updated, expiration)
	}

	if err != nil {
		return err
	}

	return nil
}
