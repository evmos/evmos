package vesting

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
	vestingtypes "github.com/evmos/evmos/v14/x/vesting/types"
)

var (
	// FundVestingAccountMsg defines the authorization type for MsgDelegate
	FundVestingAccountMsg = sdk.MsgTypeURL(&vestingtypes.MsgFundVestingAccount{})
	// UpdateVestingFunderMsg defines the authorization type for MsgDelegate
	UpdateVestingFunderMsg = sdk.MsgTypeURL(&vestingtypes.MsgUpdateVestingFunder{})
	// ClawbackMsg defines the authorization type for MsgDelegate
	ClawbackMsg = sdk.MsgTypeURL(&vestingtypes.MsgClawback{})
)

// Approve is the precompile function for approving vesting transactions with a generic grant.
func (p Precompile) Approve(
	ctx sdk.Context,
	origin common.Address,
	_ vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	grantee, typeURL, err := checkApprovalArgs(args)
	if err != nil {
		return nil, err
	}

	switch typeURL {
	case FundVestingAccountMsg, ClawbackMsg, UpdateVestingFunderMsg:
		if err := CreateGenericAuthz(ctx, p.AuthzKeeper, grantee, origin, typeURL); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf(cmn.ErrInvalidMsgType, "vesting", typeURL)
	}

	// TODO: Add event emitting maybe ?

	return method.Outputs.Pack(true)
}

// CreateGenericAuthz Creates a generic authorization grant.
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
