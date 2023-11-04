// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
	auth "github.com/evmos/evmos/v15/precompiles/authorization"
)

// Approve sets the given amount as the allowance of the spender address over
// the caller’s tokens. It returns a boolean value indicating whether the
// operation succeeded and emits the Approval event on success.
//
// The Approve method handles 4 cases:
//  1. no authorization, amount 0 or negative -> return error
//  2. no authorization, amount positive -> create a new authorization
//  3. authorization exists, amount 0 or negative -> delete authorization
//  4. authorization exists, amount positive -> update authorization
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
	authorization, _, err := auth.CheckAuthzExists(ctx, p.AuthzKeeper, grantee, granter, SendMsgURL)
	if err != nil {
		return nil, err
	}
	// case 1: authorization doesn't exist

	switch {
	case authorization == nil && amount != nil && amount.Cmp(common.Big0) <= 0:
		// case 1: no authorization, amount 0 or negative -> error
		// TODO: (@fedekunze) check if this is correct by comparing behavior with
		// regular ERC20
		err = errors.New("cannot approve non-positive values")
	case authorization == nil && amount != nil && amount.Cmp(common.Big0) > 0:
		// case 2: no authorization, amount positive -> create a new authorization
		err = p.createAuthorization(ctx, grantee, granter, amount)
	case authorization != nil && amount != nil && amount.Cmp(common.Big0) <= 0:
		// case 3: authorization exists, amount 0 or negative -> delete authorization
		err = p.deleteAuthorization(ctx, grantee, granter)
	case authorization != nil && amount != nil && amount.Cmp(common.Big0) > 0:
		// case 4: authorization exists, amount positive -> update authorization
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

// IncreaseAllowance increases the allowance of the spender address over
// the caller’s tokens by the given added value. It returns a boolean value
// indicating whether the operation succeeded and emits the Approval event on
// success.
//
// The IncreaseAllowance method handles 3 cases:
//  1. addedValue 0 or negative -> return error
//  2. no authorization, addedValue positive -> create a new authorization
//  3. authorization exists, addedValue positive -> update authorization
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
	authorization, _, err := auth.CheckAuthzExists(ctx, p.AuthzKeeper, grantee, granter, SendMsgURL)
	if err != nil {
		return nil, err
	}

	var amount *big.Int
	switch {
	case addedValue != nil && addedValue.Cmp(common.Big0) <= 0:
		// case 1: addedValue 0 or negative -> error
		// TODO: (@fedekunze) check if this is correct by comparing behavior with
		// regular ERC20
		err = errors.New("cannot increase allowance with non-positive values")
	case authorization == nil && addedValue != nil && addedValue.Cmp(common.Big0) > 0:
		// case 2: no authorization, amount positive -> create a new authorization
		amount = addedValue
		err = p.createAuthorization(ctx, grantee, granter, addedValue)
	case authorization != nil && addedValue != nil && addedValue.Cmp(common.Big0) > 0:
		// case 3: authorization exists, amount positive -> update authorization
		amount, err = p.increaseAllowance(ctx, grantee, granter, addedValue, authorization)
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

// DecreaseAllowance decreases the allowance of the spender address over
// the caller’s tokens by the given subtracted value. It returns a boolean value
// indicating whether the operation succeeded and emits the Approval event on
// success.
//
// The DecreaseAllowance method handles 4 cases:
//  1. subtractedValue 0 or negative -> return error
//  2. no authorization -> return error
//  3. authorization exists, subtractedValue positive and subtractedValue less than allowance -> update authorization
//  4. authorization exists, subtractedValue positive and subtractedValue equal to allowance -> delete authorization
//  5. authorization exists, subtractedValue positive than higher than allowance -> return error
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
	authorization, _, err := auth.CheckAuthzExists(ctx, p.AuthzKeeper, grantee, granter, SendMsgURL)
	if err != nil {
		return nil, err
	}

	// get allowance and ignore the error as it will be checked in the switch statement below
	allowance, _ := GetAllowance(p.AuthzKeeper, ctx, grantee, granter, p.tokenPair.Denom)

	// TODO: (@fedekunze) check if this is correct by comparing behavior with
	// regular ERC-20
	var amount *big.Int
	switch {
	case subtractedValue != nil && subtractedValue.Cmp(common.Big0) <= 0:
		// case 1. subtractedValue 0 or negative -> return error
		err = errors.New("cannot decrease allowance with non-positive values")
	case authorization == nil:
		// case 2. no authorization -> return error
		err = errors.New("allowance does not exist")
	case subtractedValue != nil && subtractedValue.Cmp(allowance) < 0:
		// case 3. subtractedValue positive and subtractedValue less than allowance -> update authorization
		amount, err = p.decreaseAllowance(ctx, grantee, granter, subtractedValue, authorization)
	case subtractedValue != nil && subtractedValue.Cmp(allowance) == 0:
		// case 4. subtractedValue positive and subtractedValue equal to allowance -> delete authorization
		amount, err = p.decreaseAllowance(ctx, grantee, granter, subtractedValue, authorization)
	case subtractedValue != nil && subtractedValue.Cmp(allowance) > 0:
		// case 5. subtractedValue positive than higher than allowance -> return error
		err = fmt.Errorf("subtracted value cannot be greater than existing allowance: %s > %s", subtractedValue, allowance)
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

func (p Precompile) createAuthorization(ctx sdk.Context, grantee, granter common.Address, amount *big.Int) error {
	coins := sdk.Coins{{Denom: p.tokenPair.Denom, Amount: sdk.NewIntFromBigInt(amount)}}
	expiration := ctx.BlockTime().Add(p.ApprovalExpiration)

	// NOTE: we leave the allowed arg empty as all recipients are allowed (per ERC20 standard)
	authorization := banktypes.NewSendAuthorization(coins, []sdk.AccAddress{})
	if err := authorization.ValidateBasic(); err != nil {
		return err
	}

	return p.AuthzKeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), authorization, &expiration)
}

func (p Precompile) updateAuthorization(ctx sdk.Context, grantee, granter common.Address, amount *big.Int, authorization *banktypes.SendAuthorization) error {
	authorization.SpendLimit = sdk.Coins{{Denom: p.tokenPair.Denom, Amount: sdk.NewIntFromBigInt(amount)}}
	if err := authorization.ValidateBasic(); err != nil {
		return err
	}

	expiration := ctx.BlockTime().Add(p.ApprovalExpiration)

	return p.AuthzKeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), authorization, &expiration)
}

func (p Precompile) deleteAuthorization(ctx sdk.Context, grantee, granter common.Address) error {
	return p.AuthzKeeper.DeleteGrant(ctx, grantee.Bytes(), granter.Bytes(), SendMsgURL)
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

	allowance := sendAuthz.SpendLimit.AmountOfNoDenomValidation(p.tokenPair.Denom)
	amount = new(big.Int).Add(allowance.BigInt(), addedValue)

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

	found, allowance := sendAuthz.SpendLimit.Find(p.tokenPair.Denom)
	if !found {
		return nil, errors.New("allowance for token does not exist")
	}

	amount = new(big.Int).Sub(allowance.Amount.BigInt(), subtractedValue)

	if err := p.updateAuthorization(ctx, grantee, granter, amount, sendAuthz); err != nil {
		return nil, err
	}

	return amount, nil
}
