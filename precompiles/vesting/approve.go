// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package vesting

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v17/precompiles/common"
	vestingtypes "github.com/evmos/evmos/v17/x/vesting/types"
)

var (
	// FundVestingAccountMsgURL defines the vesting authorization type for MsgFundVestingAccount
	FundVestingAccountMsgURL = sdk.MsgTypeURL(&vestingtypes.MsgFundVestingAccount{})
	// UpdateVestingFunderMsgURL defines the vesting authorization type for MsgUpdateVestingFunder
	UpdateVestingFunderMsgURL = sdk.MsgTypeURL(&vestingtypes.MsgUpdateVestingFunder{})
	// ClawbackMsgURL defines the vesting authorization type for MsgClawback
	ClawbackMsgURL = sdk.MsgTypeURL(&vestingtypes.MsgClawback{})
)

// Approve is the precompile function for approving vesting transactions with a generic grant.
func (p Precompile) Approve(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	grantee, typeURL, err := checkApprovalArgs(args)
	if err != nil {
		return nil, err
	}

	switch typeURL {
	case FundVestingAccountMsgURL, ClawbackMsgURL, UpdateVestingFunderMsgURL:
		if err := CreateGenericAuthz(ctx, p.AuthzKeeper, grantee, origin, typeURL); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf(cmn.ErrInvalidMsgType, "vesting", typeURL)
	}

	if err := p.EmitApprovalEvent(ctx, stateDB, origin, grantee, typeURL); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// CreateGenericAuthz creates a generic authorization grant.
func CreateGenericAuthz(
	ctx sdk.Context,
	authzKeeper authzkeeper.Keeper,
	grantee, granter common.Address,
	msg string,
) error {
	genericAuthorization := authz.GenericAuthorization{Msg: msg}

	expiration := ctx.BlockTime().Add(cmn.DefaultExpirationDuration).UTC()
	return authzKeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), &genericAuthorization, &expiration)
}
