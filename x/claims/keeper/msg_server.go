// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"

	"github.com/evmos/evmos/v11/x/claims/types"
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

	// Get available channels that are stored in the IBC keeper
	availableChannels := k.ibcKeeper.ChannelKeeper.GetAllChannels(ctx)

	// Validate the requested authorized channels
	authorizedChannels := req.Params.AuthorizedChannels
	for _, channelID := range authorizedChannels {
		if err := host.ChannelIdentifierValidator(channelID); err != nil {
			return nil, errorsmod.Wrapf(err,
				"invalid authorized channel contained in the request to update the claims parameters: %s",
				channelID,
			)
		}
		found := false
		for _, availableChannel := range availableChannels {
			if "channel-"+availableChannel.ChannelId == channelID {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf(
				"trying to add a channel to the claims module's available channels parameters, when it is not found in the app's IBCKeeper.ChannelKeeper: %s",
				channelID,
			)
		}
	}

	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
