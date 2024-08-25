// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"context"

	stdmath "math"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/x/staterent/types"
)

var _ types.MsgServer = &Keeper{}

// UpdateParams defines a method for updating staterent params
func (k Keeper) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority.String() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority.String(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, errorsmod.Wrapf(err, "error setting params")
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// UpdateParams defines a method for updating staterent params
func (k Keeper) FlagContract(goCtx context.Context, req *types.MsgFlagContractParams) (*types.MsgFlagContractResponse, error) {
	if k.authority.String() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority.String(), req.Authority)
	}

	// TODO: validate the address
	addr := common.HexToAddress(req.Address)

	ctx := sdk.UnwrapSDKContext(goCtx)
	info := types.FlaggedInfo{
		Contract:              req.Address,
		TotalEntries:          req.TotalEntires,
		PaymentDeposit:        math.NewInt(0),
		IsInactive:            false,
		StartDeletionTic:      stdmath.MaxUint64,
		CurrentDeletedEntries: math.NewInt(0),
	}
	k.SetFlaggedInfo(ctx, addr, info)

	return &types.MsgFlagContractResponse{}, nil
}
