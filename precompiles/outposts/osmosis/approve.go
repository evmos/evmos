// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis

import (
	"fmt"
	"time"

	"github.com/evmos/evmos/v14/precompiles/ics20"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/precompiles/authorization"
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
	grantee, transferAuthz, err := ics20.NewTransferAuthorization(method, args)
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
