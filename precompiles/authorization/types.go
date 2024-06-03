// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package authorization

import (
	"fmt"
	"math/big"
	"slices"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
)

const (
	// ApproveMethod defines the ABI method name for the authorization Approve transaction.
	ApproveMethod = "approve"
	// RevokeMethod defines the ABI method name for the authorization Revoke transaction.
	RevokeMethod = "revoke"
	// DecreaseAllowanceMethod defines the ABI method name for the DecreaseAllowance transaction.
	DecreaseAllowanceMethod = "decreaseAllowance"
	// IncreaseAllowanceMethod defines the ABI method name for the IncreaseAllowance transaction.
	IncreaseAllowanceMethod = "increaseAllowance"
	// AllowanceMethod defines the ABI method name for the Allowance query.
	AllowanceMethod = "allowance"
	// EventTypeApproval defines the event type for the distribution Approve transaction.
	EventTypeApproval = "Approval"
	// EventTypeRevocation defines the event type for the distribution Revoke transaction.
	EventTypeRevocation = "Revocation"
	// EventTypeAllowanceChange defines the event type for the staking IncreaseAllowance or
	// DecreaseAllowance transactions.
	EventTypeAllowanceChange = "AllowanceChange"
)

// CheckApprovalArgs checks the arguments passed to the approve function as well as
// the functions to change the allowance. This is refactored into one function as
// they all take in the same arguments.
func CheckApprovalArgs(args []interface{}, denom string) (common.Address, *sdk.Coin, []string, error) {
	if len(args) != 3 {
		return common.Address{}, nil, nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 3, len(args))
	}
	// TODO: (optional) new Go 1.20 functionality would allow to check all args and then return a joint list of errors.
	// This would improve UX as everything wrong with the input would be returned at once.

	grantee, ok := args[0].(common.Address)
	if !ok || grantee == (common.Address{}) {
		return common.Address{}, nil, nil, fmt.Errorf(ErrInvalidGranter, args[0])
	}

	var coin *sdk.Coin
	amount, ok := args[1].(*big.Int)
	if ok {
		if amount.Sign() == -1 {
			return common.Address{}, nil, nil, fmt.Errorf("amount cannot be negative: %v", args[1])
		}
		// If amount is not MaxUint256, create a coin with the given amount
		if amount.Cmp(abi.MaxUint256) != 0 {
			coin = &sdk.Coin{
				Denom:  denom,
				Amount: math.NewIntFromBigInt(amount),
			}
		}
	}

	typeURLs, ok := args[2].([]string)
	if !ok {
		return common.Address{}, nil, nil, fmt.Errorf(ErrInvalidMethods, args[2])
	}
	if len(typeURLs) == 0 {
		return common.Address{}, nil, nil, fmt.Errorf(ErrEmptyMethods)
	}
	if slices.Contains(typeURLs, "") {
		return common.Address{}, nil, nil, fmt.Errorf(ErrEmptyStringInMethods, typeURLs)
	}
	// TODO: check if the typeURLs are valid? e.g. with a regex pattern?

	return grantee, coin, typeURLs, nil
}

// CheckRevokeArgs checks the arguments passed to the revoke function.
func CheckRevokeArgs(args []interface{}) (common.Address, []string, error) {
	if len(args) != 2 {
		return common.Address{}, nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	// TODO: (optional) new Go 1.20 functionality would allow to check all args and then return a joint list of errors.
	// This would improve UX as everything wrong with the input would be returned at once.

	granteeAddr, ok := args[0].(common.Address)
	if !ok || granteeAddr == (common.Address{}) {
		return common.Address{}, nil, fmt.Errorf(ErrInvalidGranter, args[0])
	}

	typeURLs, err := validateMsgTypes(args[1])
	if err != nil {
		return common.Address{}, nil, err
	}
	// TODO: check if the typeURLs are valid? e.g. with a regex pattern?
	// Check - ENG-1632 on Linear

	return granteeAddr, typeURLs, nil
}

// CheckAllowanceArgs checks the arguments for the Allowance function.
func CheckAllowanceArgs(args []interface{}) (common.Address, common.Address, string, error) {
	if len(args) != 3 {
		return common.Address{}, common.Address{}, "", fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 3, len(args))
	}

	granteeAddr, ok := args[0].(common.Address)
	if !ok || granteeAddr == (common.Address{}) {
		return common.Address{}, common.Address{}, "", fmt.Errorf(ErrInvalidGrantee, args[0])
	}

	granterAddr, ok := args[1].(common.Address)
	if !ok || granterAddr == (common.Address{}) {
		return common.Address{}, common.Address{}, "", fmt.Errorf(ErrInvalidGranter, args[1])
	}

	typeURL, ok := args[2].(string)
	if !ok {
		return common.Address{}, common.Address{}, "", fmt.Errorf(ErrInvalidMethod, args[2])
	}
	if typeURL == "" {
		return common.Address{}, common.Address{}, "", fmt.Errorf("empty method defined; expected a valid message type url")
	}
	// TODO: check if the typeURL is valid? e.g. with a regex pattern?

	return granteeAddr, granterAddr, typeURL, nil
}

// CheckAuthzExists checks if the authorization exists for the given granter and
// returns the authorization and its expiration time.
//
// NOTE: It's not necessary to check for expiration of the authorization, because that is already handled
// by the GetAuthorization method. If a grant is expired, it will return nil.
func CheckAuthzExists(
	ctx sdk.Context,
	authzKeeper authzkeeper.Keeper,
	grantee, granter common.Address,
	msgURL string,
) (authz.Authorization, *time.Time, error) {
	msgAuthz, expiration := authzKeeper.GetAuthorization(ctx, grantee.Bytes(), granter.Bytes(), msgURL)
	if msgAuthz == nil {
		return nil, nil, fmt.Errorf(ErrAuthzDoesNotExistOrExpired, msgURL, grantee)
	}
	return msgAuthz, expiration, nil
}

// CheckAuthzAndAllowanceForGranter checks if the authorization exists and is not expired for the
// given spender and the allowance is not exceeded.
// If the authorization has a limit, checks that the provided amount does not exceed the current limit.
// Returns an error if the authorization does not exist
// is expired or the allowance is exceeded.
func CheckAuthzAndAllowanceForGranter(
	ctx sdk.Context,
	authzKeeper authzkeeper.Keeper,
	grantee, granter common.Address,
	amount *sdk.Coin,
	msgURL string,
) (*stakingtypes.StakeAuthorization, *time.Time, error) {
	msgAuthz, expiration := authzKeeper.GetAuthorization(ctx, grantee.Bytes(), granter.Bytes(), msgURL)
	if msgAuthz == nil {
		return nil, nil, fmt.Errorf(ErrAuthzDoesNotExistOrExpired, msgURL, grantee)
	}

	stakeAuthz, ok := msgAuthz.(*stakingtypes.StakeAuthorization)
	if !ok {
		return nil, nil, authz.ErrUnknownAuthorizationType
	}

	if stakeAuthz.MaxTokens != nil && amount.Amount.GT(stakeAuthz.MaxTokens.Amount) {
		return nil, nil, fmt.Errorf(ErrExceededAllowance, amount.Amount, stakeAuthz.MaxTokens.Amount)
	}

	return stakeAuthz, expiration, nil
}

// validateMsgTypes checks if the typeURLs are of the correct type,
// performs basic validation on the length and checks for any empty strings
func validateMsgTypes(arg interface{}) ([]string, error) {
	typeURLs, ok := arg.([]string)
	if !ok {
		return nil, fmt.Errorf(ErrInvalidMethods, arg)
	}
	if len(typeURLs) == 0 {
		return nil, fmt.Errorf(ErrEmptyMethods)
	}

	if slices.Contains(typeURLs, "") {
		return nil, fmt.Errorf(ErrEmptyStringInMethods, typeURLs)
	}

	return typeURLs, nil
}
