// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package staking

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v16/precompiles/common"
)

// EventCreateValidator defines the event data for the staking CreateValidator transaction.
type EventCreateValidator struct {
	DelegatorAddress common.Address
	ValidatorAddress common.Address
	Value            *big.Int
}

// EventDelegate defines the event data for the staking Delegate transaction.
type EventDelegate struct {
	DelegatorAddress common.Address
	ValidatorAddress common.Address
	Amount           *big.Int
	NewShares        *big.Int
}

// EventUnbond defines the event data for the staking Undelegate transaction.
type EventUnbond struct {
	DelegatorAddress common.Address
	ValidatorAddress common.Address
	Amount           *big.Int
	CompletionTime   *big.Int
}

// EventRedelegate defines the event data for the staking Redelegate transaction.
type EventRedelegate struct {
	DelegatorAddress    common.Address
	ValidatorSrcAddress common.Address
	ValidatorDstAddress common.Address
	Amount              *big.Int
	CompletionTime      *big.Int
}

// EventCancelUnbonding defines the event data for the staking CancelUnbond transaction.
type EventCancelUnbonding struct {
	DelegatorAddress common.Address
	ValidatorAddress common.Address
	Amount           *big.Int
	CreationHeight   *big.Int
}

// Description use golang type alias defines a validator description.
type Description = struct {
	Moniker         string "json:\"moniker\""
	Identity        string "json:\"identity\""
	Website         string "json:\"website\""
	SecurityContact string "json:\"securityContact\""
	Details         string "json:\"details\""
}

// Commission use golang type alias defines a validator commission.
// since solidity does not support decimals, after passing in the big int, convert the big int into a decimal with a precision of 18
type Commission = struct {
	Rate          *big.Int "json:\"rate\""
	MaxRate       *big.Int "json:\"maxRate\""
	MaxChangeRate *big.Int "json:\"maxChangeRate\""
}

// NewMsgCreateValidator creates a new MsgCreateValidator instance and does sanity checks
// on the given arguments before populating the message.
func NewMsgCreateValidator(args []interface{}, denom string) (*stakingtypes.MsgCreateValidator, common.Address, error) {
	if len(args) != 7 {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 7, len(args))
	}

	description := stakingtypes.Description{}
	if descriptionInput, ok := args[0].(Description); ok {
		description.Moniker = descriptionInput.Moniker
		description.Identity = descriptionInput.Identity
		description.Website = descriptionInput.Website
		description.SecurityContact = descriptionInput.SecurityContact
		description.Details = descriptionInput.Details
	} else {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidDescription, args[0])
	}

	commission := stakingtypes.CommissionRates{}
	if commissionInput, ok := args[1].(Commission); ok {
		commission.Rate = math.LegacyNewDecFromBigIntWithPrec(commissionInput.Rate, math.LegacyPrecision)
		commission.MaxRate = math.LegacyNewDecFromBigIntWithPrec(commissionInput.MaxRate, math.LegacyPrecision)
		commission.MaxChangeRate = math.LegacyNewDecFromBigIntWithPrec(commissionInput.MaxChangeRate, math.LegacyPrecision)
	} else {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidCommission, args[1])
	}

	minSelfDelegation, ok := args[2].(*big.Int)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidAmount, args[2])
	}

	delegatorAddress, ok := args[3].(common.Address)
	if !ok || delegatorAddress == (common.Address{}) {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidDelegator, args[3])
	}

	validatorAddress, ok := args[4].(string)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidType, "validatorAddress", "string", args[4])
	}

	// use cli `evmosd tendermint show-validator` get pubkey
	pubkeyBase64Str, ok := args[5].(string)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidType, "pubkey", "string", args[5])
	}
	pubkeyBytes, err := base64.StdEncoding.DecodeString(pubkeyBase64Str)
	if err != nil {
		return nil, common.Address{}, err
	}

	var ed25519pk cryptotypes.PubKey = &ed25519.PubKey{Key: pubkeyBytes}
	pubkey, err := codectypes.NewAnyWithValue(ed25519pk)
	if err != nil {
		return nil, common.Address{}, err
	}

	value, ok := args[6].(*big.Int)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidAmount, args[6])
	}

	msg := &stakingtypes.MsgCreateValidator{
		Description:       description,
		Commission:        commission,
		MinSelfDelegation: math.NewIntFromBigInt(minSelfDelegation),
		DelegatorAddress:  sdk.AccAddress(delegatorAddress.Bytes()).String(),
		ValidatorAddress:  validatorAddress,
		Pubkey:            pubkey,
		Value:             sdk.Coin{Denom: denom, Amount: math.NewIntFromBigInt(value)},
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, common.Address{}, err
	}

	return msg, delegatorAddress, nil
}

// NewMsgDelegate creates a new MsgDelegate instance and does sanity checks
// on the given arguments before populating the message.
func NewMsgDelegate(args []interface{}, denom string) (*stakingtypes.MsgDelegate, common.Address, error) {
	delegatorAddr, validatorAddress, amount, err := checkDelegationUndelegationArgs(args)
	if err != nil {
		return nil, common.Address{}, err
	}

	msg := &stakingtypes.MsgDelegate{
		DelegatorAddress: sdk.AccAddress(delegatorAddr.Bytes()).String(),
		ValidatorAddress: validatorAddress,
		Amount: sdk.Coin{
			Denom:  denom,
			Amount: math.NewIntFromBigInt(amount),
		},
	}

	if err = msg.ValidateBasic(); err != nil {
		return nil, common.Address{}, err
	}

	return msg, delegatorAddr, nil
}

// NewMsgUndelegate creates a new MsgUndelegate instance and does sanity checks
// on the given arguments before populating the message.
func NewMsgUndelegate(args []interface{}, denom string) (*stakingtypes.MsgUndelegate, common.Address, error) {
	delegatorAddr, validatorAddress, amount, err := checkDelegationUndelegationArgs(args)
	if err != nil {
		return nil, common.Address{}, err
	}

	msg := &stakingtypes.MsgUndelegate{
		DelegatorAddress: sdk.AccAddress(delegatorAddr.Bytes()).String(),
		ValidatorAddress: validatorAddress,
		Amount: sdk.Coin{
			Denom:  denom,
			Amount: math.NewIntFromBigInt(amount),
		},
	}

	if err = msg.ValidateBasic(); err != nil {
		return nil, common.Address{}, err
	}

	return msg, delegatorAddr, nil
}

// NewMsgRedelegate creates a new MsgRedelegate instance and does sanity checks
// on the given arguments before populating the message.
func NewMsgRedelegate(args []interface{}, denom string) (*stakingtypes.MsgBeginRedelegate, common.Address, error) {
	if len(args) != 4 {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 4, len(args))
	}

	delegatorAddr, ok := args[0].(common.Address)
	if !ok || delegatorAddr == (common.Address{}) {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	validatorSrcAddress, ok := args[1].(string)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidType, "validatorSrcAddress", "string", args[1])
	}

	validatorDstAddress, ok := args[2].(string)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidType, "validatorDstAddress", "string", args[2])
	}

	amount, ok := args[3].(*big.Int)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidAmount, args[3])
	}

	msg := &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    sdk.AccAddress(delegatorAddr.Bytes()).String(), // bech32 formatted
		ValidatorSrcAddress: validatorSrcAddress,
		ValidatorDstAddress: validatorDstAddress,
		Amount: sdk.Coin{
			Denom:  denom,
			Amount: math.NewIntFromBigInt(amount),
		},
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, common.Address{}, err
	}

	return msg, delegatorAddr, nil
}

// NewMsgCancelUnbondingDelegation creates a new MsgCancelUnbondingDelegation instance and does sanity checks
// on the given arguments before populating the message.
func NewMsgCancelUnbondingDelegation(args []interface{}, denom string) (*stakingtypes.MsgCancelUnbondingDelegation, common.Address, error) {
	if len(args) != 4 {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 4, len(args))
	}

	delegatorAddr, ok := args[0].(common.Address)
	if !ok || delegatorAddr == (common.Address{}) {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	validatorAddress, ok := args[1].(string)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidType, "validatorAddress", "string", args[1])
	}

	amount, ok := args[2].(*big.Int)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidAmount, args[2])
	}

	creationHeight, ok := args[3].(*big.Int)
	if !ok {
		return nil, common.Address{}, fmt.Errorf("invalid creation height")
	}

	msg := &stakingtypes.MsgCancelUnbondingDelegation{
		DelegatorAddress: sdk.AccAddress(delegatorAddr.Bytes()).String(), // bech32 formatted
		ValidatorAddress: validatorAddress,
		Amount: sdk.Coin{
			Denom:  denom,
			Amount: math.NewIntFromBigInt(amount),
		},
		CreationHeight: creationHeight.Int64(),
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, common.Address{}, err
	}

	return msg, delegatorAddr, nil
}

// NewDelegationRequest creates a new QueryDelegationRequest instance and does sanity checks
// on the given arguments before populating the request.
func NewDelegationRequest(args []interface{}) (*stakingtypes.QueryDelegationRequest, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	delegatorAddr, ok := args[0].(common.Address)
	if !ok || delegatorAddr == (common.Address{}) {
		return nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	validatorAddress, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "validatorAddress", "string", args[1])
	}

	return &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: sdk.AccAddress(delegatorAddr.Bytes()).String(), // bech32 formatted
		ValidatorAddr: validatorAddress,
	}, nil
}

// NewValidatorRequest create a new QueryValidatorRequest instance and does sanity checks
// on the given arguments before populating the request.
func NewValidatorRequest(args []interface{}) (*stakingtypes.QueryValidatorRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	validatorHexAddr, ok := args[0].(common.Address)
	if !ok || validatorHexAddr == (common.Address{}) {
		return nil, fmt.Errorf(cmn.ErrInvalidValidator, args[0])
	}

	validatorAddress := sdk.ValAddress(validatorHexAddr.Bytes()).String()

	return &stakingtypes.QueryValidatorRequest{ValidatorAddr: validatorAddress}, nil
}

// NewValidatorsRequest create a new QueryValidatorsRequest instance and does sanity checks
// on the given arguments before populating the request.
func NewValidatorsRequest(method *abi.Method, args []interface{}) (*stakingtypes.QueryValidatorsRequest, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	var input ValidatorsInput
	if err := method.Inputs.Copy(&input, args); err != nil {
		return nil, fmt.Errorf("error while unpacking args to ValidatorsInput struct: %s", err)
	}

	if bytes.Equal(input.PageRequest.Key, []byte{0}) {
		input.PageRequest.Key = nil
	}

	return &stakingtypes.QueryValidatorsRequest{
		Status:     input.Status,
		Pagination: &input.PageRequest,
	}, nil
}

// NewRedelegationRequest create a new QueryRedelegationRequest instance and does sanity checks
// on the given arguments before populating the request.
func NewRedelegationRequest(args []interface{}) (*RedelegationRequest, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 3, len(args))
	}

	delegatorAddr, ok := args[0].(common.Address)
	if !ok || delegatorAddr == (common.Address{}) {
		return nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	validatorSrcAddress, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "validatorSrcAddress", "string", args[1])
	}

	validatorSrcAddr, err := sdk.ValAddressFromBech32(validatorSrcAddress)
	if err != nil {
		return nil, err
	}

	validatorDstAddress, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "validatorDstAddress", "string", args[2])
	}

	validatorDstAddr, err := sdk.ValAddressFromBech32(validatorDstAddress)
	if err != nil {
		return nil, err
	}

	return &RedelegationRequest{
		DelegatorAddress:    delegatorAddr.Bytes(), // bech32 formatted
		ValidatorSrcAddress: validatorSrcAddr,
		ValidatorDstAddress: validatorDstAddr,
	}, nil
}

// NewRedelegationsRequest create a new QueryRedelegationsRequest instance and does sanity checks
// on the given arguments before populating the request.
func NewRedelegationsRequest(method *abi.Method, args []interface{}) (*stakingtypes.QueryRedelegationsRequest, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 4, len(args))
	}

	// delAddr, srcValAddr & dstValAddr
	// can be empty strings. The query will return the
	// corresponding redelegations according to the addresses specified
	// however, cannot pass all as empty strings, need to provide at least
	// the delegator address or the source validator address
	var input RedelegationsInput
	if err := method.Inputs.Copy(&input, args); err != nil {
		return nil, fmt.Errorf("error while unpacking args to RedelegationsInput struct: %s", err)
	}

	var (
		// delegatorAddr is the string representation of the delegator address
		delegatorAddr = ""
		// emptyAddr is an empty address
		emptyAddr = common.Address{}.Hex()
	)
	if input.DelegatorAddress.Hex() != emptyAddr {
		delegatorAddr = sdk.AccAddress(input.DelegatorAddress.Bytes()).String() // bech32 formatted
	}

	if delegatorAddr == "" && input.SrcValidatorAddress == "" && input.DstValidatorAddress == "" ||
		delegatorAddr == "" && input.SrcValidatorAddress == "" && input.DstValidatorAddress != "" {
		return nil, errors.New("invalid query. Need to specify at least a source validator address or delegator address")
	}

	return &stakingtypes.QueryRedelegationsRequest{
		DelegatorAddr:    delegatorAddr, // bech32 formatted
		SrcValidatorAddr: input.SrcValidatorAddress,
		DstValidatorAddr: input.DstValidatorAddress,
		Pagination:       &input.PageRequest,
	}, nil
}

// RedelegationRequest is a struct that contains the information to pass into a redelegation query.
type RedelegationRequest struct {
	DelegatorAddress    sdk.AccAddress
	ValidatorSrcAddress sdk.ValAddress
	ValidatorDstAddress sdk.ValAddress
}

// RedelegationsRequest is a struct that contains the information to pass into a redelegations query.
type RedelegationsRequest struct {
	DelegatorAddress sdk.AccAddress
	MaxRetrieve      int64
}

// UnbondingDelegationEntry is a struct that contains the information about an unbonding delegation entry.
type UnbondingDelegationEntry struct {
	CreationHeight          int64
	CompletionTime          int64
	InitialBalance          *big.Int
	Balance                 *big.Int
	UnbondingId             uint64 //nolint
	UnbondingOnHoldRefCount int64
}

// UnbondingDelegationResponse is a struct that contains the information about an unbonding delegation.
type UnbondingDelegationResponse struct {
	DelegatorAddress string
	ValidatorAddress string
	Entries          []UnbondingDelegationEntry
}

// UnbondingDelegationOutput is the output response returned by the query method.
type UnbondingDelegationOutput struct {
	UnbondingDelegation UnbondingDelegationResponse
}

// FromResponse populates the DelegationOutput from a QueryDelegationResponse.
func (do *UnbondingDelegationOutput) FromResponse(res *stakingtypes.QueryUnbondingDelegationResponse) *UnbondingDelegationOutput {
	do.UnbondingDelegation.Entries = make([]UnbondingDelegationEntry, len(res.Unbond.Entries))
	do.UnbondingDelegation.ValidatorAddress = res.Unbond.ValidatorAddress
	do.UnbondingDelegation.DelegatorAddress = res.Unbond.DelegatorAddress
	for i, entry := range res.Unbond.Entries {
		do.UnbondingDelegation.Entries[i] = UnbondingDelegationEntry{
			UnbondingId:             entry.UnbondingId,
			UnbondingOnHoldRefCount: entry.UnbondingOnHoldRefCount,
			CreationHeight:          entry.CreationHeight,
			CompletionTime:          entry.CompletionTime.UTC().Unix(),
			InitialBalance:          entry.InitialBalance.BigInt(),
			Balance:                 entry.Balance.BigInt(),
		}
	}
	return do
}

// DelegationOutput is a struct to represent the key information from
// a delegation response.
type DelegationOutput struct {
	Shares  *big.Int
	Balance cmn.Coin
}

// FromResponse populates the DelegationOutput from a QueryDelegationResponse.
func (do *DelegationOutput) FromResponse(res *stakingtypes.QueryDelegationResponse) *DelegationOutput {
	do.Shares = res.DelegationResponse.Delegation.Shares.BigInt()
	do.Balance = cmn.Coin{
		Denom:  res.DelegationResponse.Balance.Denom,
		Amount: res.DelegationResponse.Balance.Amount.BigInt(),
	}
	return do
}

// Pack packs a given slice of abi arguments into a byte array.
func (do *DelegationOutput) Pack(args abi.Arguments) ([]byte, error) {
	return args.Pack(do.Shares, do.Balance)
}

// ValidatorInfo is a struct to represent the key information from
// a validator response.
type ValidatorInfo struct {
	OperatorAddress   string   `abi:"operatorAddress"`
	ConsensusPubkey   string   `abi:"consensusPubkey"`
	Jailed            bool     `abi:"jailed"`
	Status            uint8    `abi:"status"`
	Tokens            *big.Int `abi:"tokens"`
	DelegatorShares   *big.Int `abi:"delegatorShares"` // TODO: Decimal
	Description       string   `abi:"description"`
	UnbondingHeight   int64    `abi:"unbondingHeight"`
	UnbondingTime     int64    `abi:"unbondingTime"`
	Commission        *big.Int `abi:"commission"`
	MinSelfDelegation *big.Int `abi:"minSelfDelegation"`
}

type ValidatorOutput struct {
	Validator ValidatorInfo
}

// DefaultValidatorOutput returns a ValidatorOutput with default values.
func DefaultValidatorOutput() ValidatorOutput {
	return ValidatorOutput{
		ValidatorInfo{
			OperatorAddress:   "",
			ConsensusPubkey:   "",
			Jailed:            false,
			Status:            uint8(0),
			Tokens:            big.NewInt(0),
			DelegatorShares:   big.NewInt(0),
			Description:       "",
			UnbondingHeight:   int64(0),
			UnbondingTime:     int64(0),
			Commission:        big.NewInt(0),
			MinSelfDelegation: big.NewInt(0),
		},
	}
}

// FromResponse populates the ValidatorOutput from a QueryValidatorResponse.
func (vo *ValidatorOutput) FromResponse(res *stakingtypes.QueryValidatorResponse) ValidatorOutput {
	operatorAddress, err := sdk.ValAddressFromBech32(res.Validator.OperatorAddress)
	if err != nil {
		return DefaultValidatorOutput()
	}

	return ValidatorOutput{
		Validator: ValidatorInfo{
			OperatorAddress: common.BytesToAddress(operatorAddress.Bytes()).String(),
			ConsensusPubkey: FormatConsensusPubkey(res.Validator.ConsensusPubkey),
			Jailed:          res.Validator.Jailed,
			Status:          uint8(stakingtypes.BondStatus_value[res.Validator.Status.String()]),
			Tokens:          res.Validator.Tokens.BigInt(),
			DelegatorShares: res.Validator.DelegatorShares.BigInt(), // TODO: Decimal
			// TODO: create description type,
			Description:       res.Validator.Description.Details,
			UnbondingHeight:   res.Validator.UnbondingHeight,
			UnbondingTime:     res.Validator.UnbondingTime.UTC().Unix(),
			Commission:        res.Validator.Commission.CommissionRates.Rate.BigInt(),
			MinSelfDelegation: res.Validator.MinSelfDelegation.BigInt(),
		},
	}
}

// ValidatorsInput is a struct to represent the input information for
// the validators query. Needed to unpack arguments into the PageRequest struct.
type ValidatorsInput struct {
	Status      string
	PageRequest query.PageRequest
}

// ValidatorsOutput is a struct to represent the key information from
// a validators response.
type ValidatorsOutput struct {
	Validators   []ValidatorInfo
	PageResponse query.PageResponse
}

// FromResponse populates the ValidatorsOutput from a QueryValidatorsResponse.
func (vo *ValidatorsOutput) FromResponse(res *stakingtypes.QueryValidatorsResponse) *ValidatorsOutput {
	vo.Validators = make([]ValidatorInfo, len(res.Validators))
	for i, v := range res.Validators {
		operatorAddress, err := sdk.ValAddressFromBech32(v.OperatorAddress)
		if err != nil {
			vo.Validators[i] = DefaultValidatorOutput().Validator
		} else {
			vo.Validators[i] = ValidatorInfo{
				OperatorAddress:   common.BytesToAddress(operatorAddress.Bytes()).String(),
				ConsensusPubkey:   FormatConsensusPubkey(v.ConsensusPubkey),
				Jailed:            v.Jailed,
				Status:            uint8(stakingtypes.BondStatus_value[v.Status.String()]),
				Tokens:            v.Tokens.BigInt(),
				DelegatorShares:   v.DelegatorShares.BigInt(),
				Description:       v.Description.Details,
				UnbondingHeight:   v.UnbondingHeight,
				UnbondingTime:     v.UnbondingTime.UTC().Unix(),
				Commission:        v.Commission.CommissionRates.Rate.BigInt(),
				MinSelfDelegation: v.MinSelfDelegation.BigInt(),
			}
		}
	}

	if res.Pagination != nil {
		vo.PageResponse.Total = res.Pagination.Total
		vo.PageResponse.NextKey = res.Pagination.NextKey
	}

	return vo
}

// Pack packs a given slice of abi arguments into a byte array.
func (vo *ValidatorsOutput) Pack(args abi.Arguments) ([]byte, error) {
	return args.Pack(vo.Validators, vo.PageResponse)
}

// RedelegationEntry is a struct to represent the key information from
// a redelegation entry response.
type RedelegationEntry struct {
	CreationHeight int64
	CompletionTime int64
	InitialBalance *big.Int
	SharesDst      *big.Int
}

// RedelegationValues is a struct to represent the key information from
// a redelegation response.
type RedelegationValues struct {
	DelegatorAddress    string
	ValidatorSrcAddress string
	ValidatorDstAddress string
	Entries             []RedelegationEntry
}

// RedelegationOutput returns the output for a redelegation query.
type RedelegationOutput struct {
	Redelegation RedelegationValues
}

// FromResponse populates the RedelegationOutput from a QueryRedelegationsResponse.
func (ro *RedelegationOutput) FromResponse(res stakingtypes.Redelegation) *RedelegationOutput {
	ro.Redelegation.Entries = make([]RedelegationEntry, len(res.Entries))
	ro.Redelegation.DelegatorAddress = res.DelegatorAddress
	ro.Redelegation.ValidatorSrcAddress = res.ValidatorSrcAddress
	ro.Redelegation.ValidatorDstAddress = res.ValidatorDstAddress
	for i, entry := range res.Entries {
		ro.Redelegation.Entries[i] = RedelegationEntry{
			CreationHeight: entry.CreationHeight,
			CompletionTime: entry.CompletionTime.UTC().Unix(),
			InitialBalance: entry.InitialBalance.BigInt(),
			SharesDst:      entry.SharesDst.BigInt(),
		}
	}
	return ro
}

// RedelegationEntryResponse is equivalent to a RedelegationEntry except that it
// contains a balance in addition to shares which is more suitable for client
// responses.
type RedelegationEntryResponse struct {
	RedelegationEntry RedelegationEntry
	Balance           *big.Int
}

// Redelegation contains the list of a particular delegator's redelegating bonds
// from a particular source validator to a particular destination validator.
type Redelegation struct {
	DelegatorAddress    string
	ValidatorSrcAddress string
	ValidatorDstAddress string
	Entries             []RedelegationEntry
}

// RedelegationResponse is equivalent to a Redelegation except that its entries
// contain a balance in addition to shares which is more suitable for client
// responses.
type RedelegationResponse struct {
	Redelegation Redelegation
	Entries      []RedelegationEntryResponse
}

// RedelegationsInput is a struct to represent the input information for
// the redelegations query. Needed to unpack arguments into the PageRequest struct.
type RedelegationsInput struct {
	DelegatorAddress    common.Address
	SrcValidatorAddress string
	DstValidatorAddress string
	PageRequest         query.PageRequest
}

// RedelegationsOutput is a struct to represent the key information from
// a redelegations response.
type RedelegationsOutput struct {
	Response     []RedelegationResponse
	PageResponse query.PageResponse
}

// FromResponse populates the RedelgationsOutput from a QueryRedelegationsResponse.
func (ro *RedelegationsOutput) FromResponse(res *stakingtypes.QueryRedelegationsResponse) *RedelegationsOutput {
	ro.Response = make([]RedelegationResponse, len(res.RedelegationResponses))
	for i, resp := range res.RedelegationResponses {
		// for each RedelegationResponse
		// there's a RedelegationEntryResponse array ('Entries' field)
		entries := make([]RedelegationEntryResponse, len(resp.Entries))
		for j, e := range resp.Entries {
			entries[j] = RedelegationEntryResponse{
				RedelegationEntry: RedelegationEntry{
					CreationHeight: e.RedelegationEntry.CreationHeight,
					CompletionTime: e.RedelegationEntry.CompletionTime.Unix(),
					InitialBalance: e.RedelegationEntry.InitialBalance.BigInt(),
					SharesDst:      e.RedelegationEntry.SharesDst.BigInt(),
				},
				Balance: e.Balance.BigInt(),
			}
		}

		// the Redelegation field has also an 'Entries' field of type RedelegationEntry
		redelEntries := make([]RedelegationEntry, len(resp.Redelegation.Entries))
		for j, e := range resp.Redelegation.Entries {
			redelEntries[j] = RedelegationEntry{
				CreationHeight: e.CreationHeight,
				CompletionTime: e.CompletionTime.Unix(),
				InitialBalance: e.InitialBalance.BigInt(),
				SharesDst:      e.SharesDst.BigInt(),
			}
		}

		ro.Response[i] = RedelegationResponse{
			Entries: entries,
			Redelegation: Redelegation{
				DelegatorAddress:    resp.Redelegation.DelegatorAddress,
				ValidatorSrcAddress: resp.Redelegation.ValidatorSrcAddress,
				ValidatorDstAddress: resp.Redelegation.ValidatorDstAddress,
				Entries:             redelEntries,
			},
		}
	}

	if res.Pagination != nil {
		ro.PageResponse.Total = res.Pagination.Total
		ro.PageResponse.NextKey = res.Pagination.NextKey
	}

	return ro
}

// Pack packs a given slice of abi arguments into a byte array.
func (ro *RedelegationsOutput) Pack(args abi.Arguments) ([]byte, error) {
	return args.Pack(ro.Response, ro.PageResponse)
}

// NewUnbondingDelegationRequest creates a new QueryUnbondingDelegationRequest instance and does sanity checks
// on the given arguments before populating the request.
func NewUnbondingDelegationRequest(args []interface{}) (*stakingtypes.QueryUnbondingDelegationRequest, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	delegatorAddr, ok := args[0].(common.Address)
	if !ok || delegatorAddr == (common.Address{}) {
		return nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	validatorAddress, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "validatorAddress", "string", args[1])
	}

	return &stakingtypes.QueryUnbondingDelegationRequest{
		DelegatorAddr: sdk.AccAddress(delegatorAddr.Bytes()).String(), // bech32 formatted
		ValidatorAddr: validatorAddress,
	}, nil
}

// checkDelegationUndelegationArgs checks the arguments for the delegation and undelegation functions.
func checkDelegationUndelegationArgs(args []interface{}) (common.Address, string, *big.Int, error) {
	if len(args) != 3 {
		return common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 3, len(args))
	}

	delegatorAddr, ok := args[0].(common.Address)
	if !ok || delegatorAddr == (common.Address{}) {
		return common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	validatorAddress, ok := args[1].(string)
	if !ok {
		return common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidType, "validatorAddress", "string", args[1])
	}

	amount, ok := args[2].(*big.Int)
	if !ok {
		return common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidAmount, args[2])
	}

	return delegatorAddr, validatorAddress, amount, nil
}

// FormatConsensusPubkey format ConsensusPubkey into a base64 string
func FormatConsensusPubkey(consensusPubkey *codectypes.Any) string {
	ed25519pk, ok := consensusPubkey.GetCachedValue().(cryptotypes.PubKey)
	if ok {
		return base64.StdEncoding.EncodeToString(ed25519pk.Bytes())
	}
	return consensusPubkey.String()
}
