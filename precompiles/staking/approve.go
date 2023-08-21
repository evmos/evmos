// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package staking

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/precompiles/authorization"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
)

var (
	// DelegateMsg defines the authorization type for MsgDelegate
	DelegateMsg = sdk.MsgTypeURL(&stakingtypes.MsgDelegate{})
	// UndelegateMsg defines the authorization type for MsgUndelegate
	UndelegateMsg = sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{})
	// RedelegateMsg defines the authorization type for MsgRedelegate
	RedelegateMsg = sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{})
	// CancelUnbondingDelegationMsg defines the authorization type for MsgCancelUnbondingDelegation
	CancelUnbondingDelegationMsg = sdk.MsgTypeURL(&stakingtypes.MsgCancelUnbondingDelegation{})
)

// Approve sets amount as the allowance of a grantee over the caller’s tokens.
// Returns a boolean value indicating whether the operation succeeded.
func (p Precompile) Approve(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	grantee, coin, typeURLs, err := authorization.CheckApprovalArgs(args, p.stakingKeeper.BondDenom(ctx))
	if err != nil {
		return nil, err
	}

	for _, typeURL := range typeURLs {
		switch typeURL {
		case DelegateMsg, UndelegateMsg, RedelegateMsg, CancelUnbondingDelegationMsg:
			authzType, err := convertMsgToAuthz(typeURL)
			if err != nil {
				return nil, errorsmod.Wrap(err, fmt.Sprintf(cmn.ErrInvalidMsgType, "staking", typeURL))
			}
			if err = p.grantOrDeleteStakingAuthz(ctx, grantee, origin, coin, authzType); err != nil {
				return nil, err
			}
		default:
			// TODO: do we need to return an error here or just no-op?
			// Implications of returning an error could be that we approve some parts of the txs but not all
			return nil, fmt.Errorf(cmn.ErrInvalidMsgType, "staking", typeURL)
		}
	}

	// TODO: do we want to emit one approval for all typeUrls, or one approval for each typeUrl?
	// NOTE: This might have gas implications as we are emitting a slice of strings
	if err := p.EmitApprovalEvent(ctx, stateDB, grantee, origin, coin, typeURLs); err != nil {
		return nil, err
	}
	return method.Outputs.Pack(true)
}

// Revoke removes the authorization grants given in the typeUrls for a given granter to a given grantee.
// It only works if the origin matches the spender to avoid unauthorized revocations.
// Works only for staking messages.
func (p Precompile) Revoke(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	grantee, typeURLs, err := authorization.CheckRevokeArgs(args)
	if err != nil {
		return nil, err
	}

	for _, typeURL := range typeURLs {
		switch typeURL {
		case DelegateMsg, UndelegateMsg, RedelegateMsg, CancelUnbondingDelegationMsg:
			if err = p.AuthzKeeper.DeleteGrant(ctx, grantee.Bytes(), origin.Bytes(), typeURL); err != nil {
				return nil, err
			}
		default:
			// TODO: do we need to return an error here or just no-op?
			// Implications of returning an error could be that we approve some parts of the txs but not all
			return nil, fmt.Errorf(cmn.ErrInvalidMsgType, "staking", typeURL)
		}
	}

	// NOTE: Using the new more generic event emitter that was created
	if err = authorization.EmitRevocationEvent(cmn.EmitEventArgs{
		Ctx:            ctx,
		StateDB:        stateDB,
		ContractAddr:   p.Address(),
		ContractEvents: p.ABI.Events,
		EventData: authorization.EventRevocation{
			Granter:  origin,
			Grantee:  grantee,
			TypeUrls: typeURLs,
		},
	}); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// DecreaseAllowance decreases the allowance of grantee over the caller’s tokens by the amount.
func (p Precompile) DecreaseAllowance(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	grantee, coin, typeUrls, err := authorization.CheckApprovalArgs(args, p.stakingKeeper.BondDenom(ctx))
	if err != nil {
		return nil, err
	}

	for _, typeURL := range typeUrls {
		switch typeURL {
		case DelegateMsg, UndelegateMsg, RedelegateMsg, CancelUnbondingDelegationMsg:
			authzGrant, expiration, err := authorization.CheckAuthzExists(ctx, p.AuthzKeeper, grantee, origin, typeURL)
			if err != nil {
				return nil, err
			}

			stakeAuthz, ok := authzGrant.(*stakingtypes.StakeAuthorization)
			if !ok {
				return nil, errorsmod.Wrapf(authz.ErrUnknownAuthorizationType, "expected: *types.StakeAuthorization, received: %T", authzGrant)
			}

			if err = p.decreaseAllowance(ctx, grantee, origin, coin, stakeAuthz, expiration); err != nil {
				return nil, err
			}
		default:
			// TODO: do we need to return an error here or just no-op?
			// Implications of returning an error could be that we decrease allowance for some parts of the typeUrls but not all
			return nil, fmt.Errorf(cmn.ErrInvalidMsgType, "staking", typeURL)
		}
	}

	if err := p.EmitAllowanceChangeEvent(ctx, stateDB, grantee, origin, typeUrls); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// IncreaseAllowance increases the allowance of grantee over the caller’s tokens by the amount.
func (p Precompile) IncreaseAllowance(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	grantee, coin, typeUrls, err := authorization.CheckApprovalArgs(args, p.stakingKeeper.BondDenom(ctx))
	if err != nil {
		return nil, err
	}

	for _, typeURL := range typeUrls {
		switch typeURL {
		case DelegateMsg, UndelegateMsg, RedelegateMsg:
			if err = p.increaseAllowance(ctx, grantee, origin, coin, typeURL); err != nil {
				return nil, err
			}
		default:
			// TODO: do we need to return an error here or just no-op?
			// Implications of returning an error could be that we decrease allowance for some parts of the typeUrls but not all
			return nil, fmt.Errorf(cmn.ErrInvalidMsgType, "staking", typeURL)
		}
	}
	if err := p.EmitAllowanceChangeEvent(ctx, stateDB, grantee, origin, typeUrls); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// grantOrDeleteStakingAuthz grants staking method authorization to the precompiled contract for a spender.
// If the amount is zero, it deletes the authorization if it exists.
func (p Precompile) grantOrDeleteStakingAuthz(
	ctx sdk.Context,
	grantee, granter common.Address,
	coin *sdk.Coin,
	authzType stakingtypes.AuthorizationType,
) error {
	// Case 1: coin is nil -> set authorization with no limit
	if coin == nil || coin.IsNil() {
		p.Logger(ctx).Debug(
			"setting authorization without limit",
			"grantee", grantee.String(),
			"granter", granter.String(),
		)
		return p.createStakingAuthz(ctx, grantee, granter, coin, authzType)
	}

	// Case 2: coin amount is zero or negative -> delete the authorization
	if !coin.Amount.IsPositive() {
		p.Logger(ctx).Debug(
			"deleting authorization",
			"grantee", grantee.String(),
			"granter", granter.String(),
		)
		stakingAuthz := stakingtypes.StakeAuthorization{AuthorizationType: authzType}
		return p.AuthzKeeper.DeleteGrant(ctx, grantee.Bytes(), granter.Bytes(), stakingAuthz.MsgTypeURL())
	}

	// Case 3: coin amount is non zero -> and not coin is not nil set with custom amount
	return p.createStakingAuthz(ctx, grantee, granter, coin, authzType)
}

// createStakingAuthz creates a staking authorization for a spender.
func (p Precompile) createStakingAuthz(
	ctx sdk.Context,
	grantee, granter common.Address,
	coin *sdk.Coin,
	authzType stakingtypes.AuthorizationType,
) error {
	// Get all available validators and filter out jailed validators
	validators := make([]sdk.ValAddress, 0)
	p.stakingKeeper.IterateValidators(
		ctx, func(_ int64, validator stakingtypes.ValidatorI) (stop bool) {
			if validator.IsJailed() {
				return
			}
			validators = append(validators, validator.GetOperator())
			return
		},
	)
	stakingAuthz, err := stakingtypes.NewStakeAuthorization(validators, nil, authzType, coin)
	if err != nil {
		return err
	}

	if err := stakingAuthz.ValidateBasic(); err != nil {
		return err
	}

	expiration := ctx.BlockTime().Add(p.ApprovalExpiration).UTC()
	return p.AuthzKeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), stakingAuthz, &expiration)
}

// decreaseAllowance decreases the allowance of spender over the caller’s tokens by the amount.
func (p Precompile) decreaseAllowance(
	ctx sdk.Context,
	grantee, granter common.Address,
	coin *sdk.Coin,
	stakeAuthz *stakingtypes.StakeAuthorization,
	expiration *time.Time,
) error {
	// If the authorization has no limit, no operation is performed
	if stakeAuthz.MaxTokens == nil {
		p.Logger(ctx).Debug("decreaseAllowance called with no limit (stakeAuthz.MaxTokens == nil): no-op")
		return nil
	}

	// If the authorization limit is less than the substation amount, return error
	if stakeAuthz.MaxTokens.Amount.LT(coin.Amount) {
		return fmt.Errorf(ErrDecreaseAmountTooBig, coin.Amount, stakeAuthz.MaxTokens.Amount)
	}

	// If amount is less than or equal to the Authorization amount, subtract the amount from the limit
	if coin.Amount.LTE(stakeAuthz.MaxTokens.Amount) {
		stakeAuthz.MaxTokens.Amount = stakeAuthz.MaxTokens.Amount.Sub(coin.Amount)
	}

	return p.AuthzKeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), stakeAuthz, expiration)
}

// increaseAllowance increases the allowance of spender over the caller’s tokens by the amount.
func (p Precompile) increaseAllowance(
	ctx sdk.Context,
	grantee, granter common.Address,
	coin *sdk.Coin,
	msgURL string,
) error {
	// Check if the authorization exists for the given spender
	existingAuthz, expiration, err := authorization.CheckAuthzExists(ctx, p.AuthzKeeper, grantee, granter, msgURL)
	if err != nil {
		return err
	}

	// Cast the authorization to a staking authorization
	stakeAuthz, ok := existingAuthz.(*stakingtypes.StakeAuthorization)
	if !ok {
		return errorsmod.Wrapf(authz.ErrUnknownAuthorizationType, "expected: *types.StakeAuthorization, received: %T", existingAuthz)
	}

	// If the authorization has no limit, no operation is performed
	if stakeAuthz.MaxTokens == nil {
		p.Logger(ctx).Debug("increaseAllowance called with no limit (stakeAuthz.MaxTokens == nil): no-op")
		return nil
	}

	// Add the amount to the limit
	stakeAuthz.MaxTokens.Amount = stakeAuthz.MaxTokens.Amount.Add(coin.Amount)

	return p.AuthzKeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), stakeAuthz, expiration)
}

// UpdateStakingAuthorization updates the staking grant based on the authz AcceptResponse for the given granter and grantee.
func (p Precompile) UpdateStakingAuthorization(
	ctx sdk.Context,
	grantee, granter common.Address,
	stakeAuthz *stakingtypes.StakeAuthorization,
	expiration *time.Time,
	messageType string,
	msg sdk.Msg,
) error {
	updatedResponse, err := stakeAuthz.Accept(ctx, msg)
	if err != nil {
		return err
	}

	if updatedResponse.Delete {
		err = p.AuthzKeeper.DeleteGrant(ctx, grantee.Bytes(), granter.Bytes(), messageType)
	} else {
		err = p.AuthzKeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), updatedResponse.Updated, expiration)
	}

	if err != nil {
		return err
	}
	return nil
}

// convertMsgToAuthz converts a msg to an authorization type.
func convertMsgToAuthz(msg string) (stakingtypes.AuthorizationType, error) {
	switch msg {
	case DelegateMsg:
		return DelegateAuthz, nil
	case UndelegateMsg:
		return UndelegateAuthz, nil
	case RedelegateMsg:
		return RedelegateAuthz, nil
	case CancelUnbondingDelegationMsg:
		return CancelUnbondingDelegationAuthz, nil
	default:
		return stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNSPECIFIED, authz.ErrUnknownAuthorizationType.Wrapf("cannot convert msg to authorization type with %T", msg)
	}
}
