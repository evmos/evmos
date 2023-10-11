// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ics20

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/precompiles/authorization"
)

const (
	// TransferMethod defines the ABI method name for the ICS20 Transfer
	// transaction.
	TransferMethod = "transfer"
)

// Transfer implements the ICS20 transfer transactions.
func (p Precompile) Transfer(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, sender, err := NewMsgTransfer(method, args)
	if err != nil {
		return nil, err
	}

	// check if channel exists and is open
	if !p.channelKeeper.HasChannel(ctx, msg.SourcePort, msg.SourceChannel) {
		return nil, errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", msg.SourcePort, msg.SourceChannel)
	}

	// The provided sender address should always be equal to the origin address.
	// In case the contract caller address is the same as the sender address provided,
	// update the sender address to be equal to the origin address.
	// Otherwise, if the provided delegator address is different from the origin address,
	// return an error because is a forbidden operation
	if contract.CallerAddress == sender {
		sender = origin
	} else if origin != sender {
		return nil, fmt.Errorf(ErrDifferentOriginFromSender, origin.String(), sender.String())
	}

	// no need to have authorization when the contract caller is the same as origin (owner of funds)
	// and the sender is the origin
	var (
		expiration *time.Time
		auth       authz.Authorization
		resp       *authz.AcceptResponse
	)

	if contract.CallerAddress != origin {
		// check if authorization exists
		auth, expiration, err = authorization.CheckAuthzExists(ctx, p.AuthzKeeper, contract.CallerAddress, origin, TransferMsgURL)
		if err != nil {
			return nil, fmt.Errorf(authorization.ErrAuthzDoesNotExistOrExpired, contract.CallerAddress, origin)
		}

		// Accept the grant and return an error if the grant is not accepted
		resp, err = AcceptGrant(ctx, contract.CallerAddress, origin, msg, auth)
		if err != nil {
			return nil, err
		}
	}

	res, err := p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	// Update grant only if is needed
	if contract.CallerAddress != origin {
		// accepts and updates the grant adjusting the spending limit
		if err = UpdateGrant(ctx, p.AuthzKeeper, contract.CallerAddress, origin, expiration, resp); err != nil {
			return nil, err
		}
	}

	if err = p.EmitIBCTransferEvent(
		ctx,
		stateDB,
		sender,
		msg.Receiver,
		msg.SourcePort,
		msg.SourceChannel,
		msg.Token,
		msg.Memo,
	); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(res.Sequence)
}
