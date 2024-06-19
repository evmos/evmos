// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ics20

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/evmos/evmos/v18/precompiles/authorization"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channelkeeper "github.com/cosmos/ibc-go/v7/modules/core/04-channel/keeper"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/x/evm/core/vm"
)

// TransferMsgURL is the ICS20 transfer message URL string.
var TransferMsgURL = sdk.MsgTypeURL(&transfertypes.MsgTransfer{})

// Approve implements the ICS20 Authorization approve transactions.
func Approve(
	ctx sdk.Context,
	authzKeeper authzkeeper.Keeper,
	channelKeeper channelkeeper.Keeper,
	precompileAddr, grantee, origin common.Address,
	approvalExpiration time.Duration,
	transferAuthz *transfertypes.TransferAuthorization,
	event abi.Event,
	stateDB vm.StateDB,
) error {
	// If one of the allocations contains a non-existing channel, throw and error
	for _, allocation := range transferAuthz.Allocations {
		found := channelKeeper.HasChannel(ctx, allocation.SourcePort, allocation.SourceChannel)
		if !found {
			return errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", allocation.SourcePort, allocation.SourceChannel)
		}
	}

	// Only the origin can approve a transfer to the grantee address
	expiration := ctx.BlockTime().Add(approvalExpiration).UTC()
	if err := authzKeeper.SaveGrant(ctx, grantee.Bytes(), origin.Bytes(), transferAuthz, &expiration); err != nil {
		return err
	}

	allocations := convertToAllocation(transferAuthz.Allocations)
	// Emit the IBC transfer authorization event
	return authorization.EmitIBCTransferAuthorizationEvent(
		event,
		ctx,
		stateDB,
		precompileAddr,
		grantee,
		origin,
		allocations,
	)
}

// Revoke implements the ICS20 Authorization revoke transactions.
func Revoke(
	ctx sdk.Context,
	authzKeeper authzkeeper.Keeper,
	precompileAddr, grantee, origin common.Address,
	event abi.Event,
	stateDB vm.StateDB,
) error {
	// NOTE: we do not need to check the expiration as it will return nil if both not found or expired
	msgAuthz, _, err := authorization.CheckAuthzExists(ctx, authzKeeper, grantee, origin, TransferMsgURL)
	if err != nil {
		return fmt.Errorf(authorization.ErrAuthzDoesNotExistOrExpired, grantee, origin)
	}

	// check that the stored authorization matches the transfer authorization
	if _, ok := msgAuthz.(*transfertypes.TransferAuthorization); !ok {
		return authz.ErrUnknownAuthorizationType
	}

	if err = authzKeeper.DeleteGrant(ctx, grantee.Bytes(), origin.Bytes(), TransferMsgURL); err != nil {
		return err
	}

	return authorization.EmitIBCTransferAuthorizationEvent(
		event,
		ctx,
		stateDB,
		precompileAddr,
		grantee,
		origin,
		[]cmn.ICS20Allocation{},
	)
}

// IncreaseAllowance implements the ICS20 Authorization increase allowance transactions.
func IncreaseAllowance(
	ctx sdk.Context,
	authzKeeper authzkeeper.Keeper,
	precompileAddr, grantee, granter common.Address,
	sourcePort, sourceChannel, denom string,
	amount *big.Int,
	event abi.Event,
	stateDB vm.StateDB,
) error {
	// NOTE: we do not need to check the expiration as it will return nil if both found or expired
	msgAuthz, expiration, err := authorization.CheckAuthzExists(ctx, authzKeeper, grantee, granter, TransferMsgURL)
	if err != nil {
		return fmt.Errorf(authorization.ErrAuthzDoesNotExistOrExpired, grantee, granter)
	}

	// NOTE: we do not need to check the expiration as it will return nil if both found or expired
	transferAuthz, ok := msgAuthz.(*transfertypes.TransferAuthorization)
	if !ok {
		return authz.ErrUnknownAuthorizationType
	}

	// Check if the allocations matches the arguments provided and returns the index of the allocation and coin found
	spendLimit, allocationIdx, err := checkAllocationExists(transferAuthz.Allocations, sourcePort, sourceChannel, denom)
	if err != nil {
		return err
	}

	allowance := math.NewIntFromBigInt(amount)
	if _, overflow := cmn.SafeAdd(spendLimit.Amount, allowance); overflow {
		return errors.New(cmn.ErrIntegerOverflow)
	}

	allowanceCoin := sdk.Coin{Denom: denom, Amount: allowance}

	transferAuthz.Allocations[allocationIdx].SpendLimit = transferAuthz.Allocations[allocationIdx].SpendLimit.Add(allowanceCoin)

	if err = authzKeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), transferAuthz, expiration); err != nil {
		return err
	}

	allocations := convertToAllocation(transferAuthz.Allocations)
	// Emit the IBC transfer authorization event
	return authorization.EmitIBCTransferAuthorizationEvent(
		event,
		ctx,
		stateDB,
		precompileAddr,
		grantee,
		granter,
		allocations,
	)
}

// DecreaseAllowance implements the ICS20 Authorization decrease allowance transactions.
func DecreaseAllowance(
	ctx sdk.Context,
	authzKeeper authzkeeper.Keeper,
	precompileAddr, grantee, granter common.Address,
	sourcePort, sourceChannel, denom string,
	amount *big.Int,
	event abi.Event,
	stateDB vm.StateDB,
) error {
	// NOTE: we do not need to check the expiration as it will return nil if both found or expired
	msgAuthz, expiration, err := authorization.CheckAuthzExists(ctx, authzKeeper, grantee, granter, TransferMsgURL)
	if err != nil {
		return fmt.Errorf(authorization.ErrAuthzDoesNotExistOrExpired, grantee, granter)
	}

	transferAuthz, ok := msgAuthz.(*transfertypes.TransferAuthorization)
	if !ok {
		return authz.ErrUnknownAuthorizationType
	}

	// Check if the allocations matches the arguments provided and returns the index of the allocation and spend limit found
	spendLimit, allocationIdx, err := checkAllocationExists(transferAuthz.Allocations, sourcePort, sourceChannel, denom)
	if err != nil {
		return err
	}

	expense := math.NewIntFromBigInt(amount)
	if spendLimit.Amount.LT(expense) {
		return fmt.Errorf(cmn.ErrNegativeAmount)
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
	if err = authzKeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), transferAuthz, expiration); err != nil {
		return err
	}

	allocations := convertToAllocation(transferAuthz.Allocations)
	// Emit the IBC transfer authorization event
	return authorization.EmitIBCTransferAuthorizationEvent(
		event,
		ctx,
		stateDB,
		precompileAddr,
		grantee,
		granter,
		allocations,
	)
}

// AcceptGrant implements the ICS20 accept grant.
func AcceptGrant(
	ctx sdk.Context,
	contractCaller, granter common.Address,
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
		return nil, fmt.Errorf(authorization.ErrAuthzNotAccepted, contractCaller, granter)
	}

	return &resp, nil
}

// UpdateGrant implements the ICS20 authz update grant.
func UpdateGrant(
	ctx sdk.Context,
	authzKeeper authzkeeper.Keeper,
	grantee, granter common.Address,
	expiration *time.Time,
	resp *authz.AcceptResponse,
) (err error) {
	if resp.Delete {
		err = authzKeeper.DeleteGrant(ctx, grantee.Bytes(), granter.Bytes(), TransferMsgURL)
	} else if resp.Updated != nil {
		err = authzKeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), resp.Updated, expiration)
	}

	if err != nil {
		return err
	}

	return nil
}
