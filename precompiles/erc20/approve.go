package erc20

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
	auth "github.com/evmos/evmos/v14/precompiles/authorization"
)

// SendMsgURL defines the authorization type for MsgSend
var SendMsgURL = sdk.MsgTypeURL(&banktypes.MsgSend{})

// Approve sets amount as the allowance of spender over the caller’s tokens.
// Returns a boolean value indicating whether the operation succeeded.
func (p Precompile) Approve(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	spender, amount, err := ParseApproveArgs(args)
	if err != nil {
		return nil, err
	}

	grantee := spender
	granter := contract.CallerAddress

	// TODO: owner should be the owner of the contract
	authorization, _, err := auth.CheckAuthzExists(ctx, p.authzKeeper, grantee, granter, SendMsgURL)
	if err != nil {
		return nil, err
	}
	// case 1: authorization doesn't exist

	// case 2.1: amount 0 or negative -> delete authorization
	// case 2.2: amount positive -> update authorization (amt and timeout)

	switch {
	case authorization == nil && amount != nil && amount.Cmp(common.Big0) <= 0:
		// case 1.1: no authorization, amount 0 or negative -> error
		// TODO: (@fedekunze) check if this is correct by comparing behaviour with
		// regular ERC20
		err = errors.New("cannot approve non-positive values")
	case authorization == nil && amount != nil && amount.Cmp(common.Big0) > 0:
		// case 1.2: no authorization, amount positive -> create a new authorization
		err = p.createAuthorization(ctx, grantee, granter, amount)
	case authorization != nil && amount != nil && amount.Cmp(common.Big0) <= 0:
		// case 2.1: authorization exists, amount 0 or negative -> delete authorization
		err = p.deleteAuthorization(ctx, grantee, granter)
	case authorization != nil && amount != nil && amount.Cmp(common.Big0) > 0:
		// case 2.2: authorization exists, amount positive -> update authorization
		sendAuthz, ok := authorization.(*banktypes.SendAuthorization)
		if !ok {
			return nil, authz.ErrUnknownAuthorizationType
		}

		err = p.updateAuthorization(ctx, grantee, granter, amount, sendAuthz)
	}

	if err != nil {
		return nil, err
	}

	// TODO: check owner?
	if err := p.EmitApprovalEvent(ctx, stateDB, p.Address(), spender, amount); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// Approve sets amount as the allowance of spender over the caller’s tokens.
// Returns a boolean value indicating whether the operation succeeded.
func (p Precompile) IncreaseAllowance(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	spender, addedValue, err := ParseApproveArgs(args)
	if err != nil {
		return nil, err
	}

	grantee := spender
	granter := contract.CallerAddress

	// TODO: owner should be the owner of the contract
	authorization, _, err := auth.CheckAuthzExists(ctx, p.authzKeeper, grantee, granter, SendMsgURL)
	if err != nil {
		return nil, err
	}

	var amount *big.Int
	switch {
	case authorization == nil && addedValue != nil && addedValue.Cmp(common.Big0) <= 0:
		// case 1.1: no authorization, amount 0 or negative -> error
		// TODO: (@fedekunze) check if this is correct by comparing behaviour with
		// regular ERC20
		err = errors.New("cannot approve non-positive values")
	case authorization == nil && addedValue != nil && addedValue.Cmp(common.Big0) > 0:
		// case 1.2: no authorization, amount positive -> create a new authorization
		amount = addedValue
		err = p.createAuthorization(ctx, grantee, granter, addedValue)
	case authorization != nil && addedValue != nil && addedValue.Cmp(common.Big0) <= 0:
		// case 2.1: authorization exists, amount 0 or negative -> delete authorization
		// TODO: check what happens when addedValue = 0
		err = p.deleteAuthorization(ctx, grantee, granter)
	case authorization != nil && addedValue != nil && addedValue.Cmp(common.Big0) > 0:
		// case 2.2: authorization exists, amount positive -> update authorization
		amount, err = p.increaseAllowance(ctx, grantee, granter, addedValue, authorization)
	}

	// TODO: check owner?
	if err := p.EmitApprovalEvent(ctx, stateDB, p.Address(), spender, amount); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// Approve sets amount as the allowance of spender over the caller’s tokens.
// Returns a boolean value indicating whether the operation succeeded.
func (p Precompile) DecreaseAllowance(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	spender, subtractedValue, err := ParseApproveArgs(args)
	if err != nil {
		return nil, err
	}

	grantee := spender
	granter := contract.CallerAddress

	// TODO: owner should be the owner of the contract
	authorization, _, err := auth.CheckAuthzExists(ctx, p.authzKeeper, grantee, granter, SendMsgURL)
	if err != nil {
		return nil, err
	}

	var amount *big.Int
	switch {
	case authorization == nil && subtractedValue != nil && subtractedValue.Cmp(common.Big0) <= 0:
		// case 1.1: no authorization, amount 0 or negative -> error
		// TODO: (@fedekunze) check if this is correct by comparing behaviour with
		// regular ERC20
		err = errors.New("cannot approve non-positive values")
	case authorization == nil && subtractedValue != nil && subtractedValue.Cmp(common.Big0) > 0:
		// case 1.2: no authorization, amount positive -> create a new authorization
		amount = subtractedValue
		err = p.createAuthorization(ctx, grantee, granter, subtractedValue)
	case authorization != nil && subtractedValue != nil && subtractedValue.Cmp(common.Big0) <= 0:
		// case 2.1: authorization exists, amount 0 or negative -> delete authorization
		// TODO: check what happens when subtractedValue = 0
		err = p.deleteAuthorization(ctx, grantee, granter)
	case authorization != nil && subtractedValue != nil && subtractedValue.Cmp(common.Big0) > 0:
		// case 2.2: authorization exists, amount positive -> update authorization
		amount, err = p.decreaseAllowance(ctx, grantee, granter, subtractedValue, authorization)
	}

	// TODO: check owner?
	if err := p.EmitApprovalEvent(ctx, stateDB, p.Address(), spender, amount); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

func (p Precompile) createAuthorization(ctx sdk.Context, grantee, granter common.Address, amount *big.Int) error {
	coins := sdk.Coins{{Denom: p.tokenPair.Denom, Amount: sdk.NewIntFromBigInt(amount)}}
	expiration := ctx.BlockTime().Add(p.approvalExpiration)

	// NOTE: we leave the allowed arg empty as all recipients are allowed (per ERC20 standard)
	authorization := banktypes.NewSendAuthorization(coins, []sdk.AccAddress{})
	if err := authorization.ValidateBasic(); err != nil {
		return err
	}

	return p.authzKeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), authorization, &expiration)
}

func (p Precompile) deleteAuthorization(ctx sdk.Context, grantee, granter common.Address) error {
	return p.authzKeeper.DeleteGrant(ctx, grantee.Bytes(), granter.Bytes(), SendMsgURL)
}

func (p Precompile) increaseAllowance(
	ctx sdk.Context,
	grantee, granter common.Address,
	addedValue *big.Int,
	authorization authz.Authorization,
) (amount *big.Int, err error) {
	sendAuthz, ok := authorization.(*banktypes.SendAuthorization)
	if !ok {
		return nil, authz.ErrUnknownAuthorizationType
	}

	found, coin := sendAuthz.SpendLimit.Find(p.tokenPair.Denom)
	if found {
		amount = new(big.Int).Add(coin.Amount.BigInt(), addedValue)
	} else {
		amount = new(big.Int).Set(addedValue)
	}

	if err := p.updateAuthorization(ctx, grantee, granter, amount, sendAuthz); err != nil {
		return nil, err
	}

	return amount, nil
}

func (p Precompile) decreaseAllowance(
	ctx sdk.Context,
	grantee, granter common.Address,
	subtractedValue *big.Int,
	authorization authz.Authorization,
) (amount *big.Int, err error) {
	sendAuthz, ok := authorization.(*banktypes.SendAuthorization)
	if !ok {
		return nil, authz.ErrUnknownAuthorizationType
	}

	found, coin := sendAuthz.SpendLimit.Find(p.tokenPair.Denom)
	if found {
		amount = new(big.Int).Sub(coin.Amount.BigInt(), subtractedValue)
	} else {
		amount = new(big.Int).Set(subtractedValue)
	}

	if err := p.updateAuthorization(ctx, grantee, granter, amount, sendAuthz); err != nil {
		return nil, err
	}

	return amount, nil
}

func (p Precompile) updateAuthorization(ctx sdk.Context, grantee, granter common.Address, amount *big.Int, authorization *banktypes.SendAuthorization) error {
	// case 2.2: amount positive -> update authorization (amt and timeout)

	authorization.SpendLimit = sdk.Coins{{Denom: p.tokenPair.Denom, Amount: sdk.NewIntFromBigInt(amount)}}
	if err := authorization.ValidateBasic(); err != nil {
		return err
	}

	expiration := ctx.BlockTime().Add(p.approvalExpiration)

	return p.authzKeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), authorization, &expiration)
}
