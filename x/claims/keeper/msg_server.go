// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	"github.com/evmos/evmos/v13/x/claims/types"
)

// UpdateParams implements the gRPC MsgServer interface. When an UpdateParams
// proposal passes, it updates the module parameters. The update can only be
// performed if the requested authority is the Cosmos SDK governance module
// account.
func (k *Keeper) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority.String() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority.String(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the requested authorized channels
	for _, channelID := range req.Params.AuthorizedChannels {
		if err := checkIfChannelOpen(ctx, k.channelKeeper, channelID); err != nil {
			return nil, errorsmod.Wrapf(err, "invalid authorized channel")
		}
	}

	// Validate the requested EVM channels
	for _, channelID := range req.Params.EVMChannels {
		if err := checkIfChannelOpen(ctx, k.channelKeeper, channelID); err != nil {
			return nil, errorsmod.Wrapf(err, "invalid evm channel")
		}
	}

	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// checkIfChannelOpen checks if an IBC channel with the given channel id is registered
// in the channel keeper and is in the OPEN state. It also requires the channel id to
// be in a valid format.
//
// NOTE: this function is looking for a channel with the default transfer port id and will fail
// if no channel with the given channel id has an open connection to this port.
func checkIfChannelOpen(ctx sdk.Context, ck types.ChannelKeeper, channelID string) error {
	channel, found := ck.GetChannel(ctx, transfertypes.PortID, channelID)
	if !found {
		return fmt.Errorf(
			"trying to add a channel to the claims module's available channels parameters, when it is not found in the app's IBCKeeper.ChannelKeeper: %s",
			channelID,
		)
	}

	if channel.State != channeltypes.OPEN {
		return fmt.Errorf(
			"trying to add a channel to the claims module's available channels parameters, when it is not in the OPEN state: %s",
			channelID,
		)
	}

	return nil
}
