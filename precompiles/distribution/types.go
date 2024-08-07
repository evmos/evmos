// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package distribution

import (
	"fmt"
	"math/big"

	"github.com/evmos/evmos/v19/utils"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/cmd/config"
	cmn "github.com/evmos/evmos/v19/precompiles/common"
)

// EventSetWithdrawAddress defines the event data for the SetWithdrawAddress transaction.
type EventSetWithdrawAddress struct {
	Caller            common.Address
	WithdrawerAddress string
}

// EventWithdrawDelegatorRewards defines the event data for the WithdrawDelegatorRewards transaction.
type EventWithdrawDelegatorRewards struct {
	DelegatorAddress common.Address
	ValidatorAddress common.Address
	Amount           *big.Int
}

// EventWithdrawValidatorRewards defines the event data for the WithdrawValidatorRewards transaction.
type EventWithdrawValidatorRewards struct {
	ValidatorAddress common.Hash
	Commission       *big.Int
}

// EventClaimRewards defines the event data for the ClaimRewards transaction.
type EventClaimRewards struct {
	DelegatorAddress common.Address
	Amount           *big.Int
}

// EventFundCommunityPool defines the event data for the FundCommunityPool transaction.
type EventFundCommunityPool struct {
	Depositor common.Address
	Amount    *big.Int
}

// parseClaimRewardsArgs parses the arguments for the ClaimRewards method.
func parseClaimRewardsArgs(args []interface{}) (common.Address, uint32, error) {
	if len(args) != 2 {
		return common.Address{}, 0, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	delegatorAddress, ok := args[0].(common.Address)
	if !ok || delegatorAddress == (common.Address{}) {
		return common.Address{}, 0, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	maxRetrieve, ok := args[1].(uint32)
	if !ok {
		return common.Address{}, 0, fmt.Errorf(cmn.ErrInvalidType, "maxRetrieve", uint32(0), args[1])
	}

	return delegatorAddress, maxRetrieve, nil
}

// NewMsgSetWithdrawAddress creates a new MsgSetWithdrawAddress instance.
func NewMsgSetWithdrawAddress(args []interface{}) (*distributiontypes.MsgSetWithdrawAddress, common.Address, error) {
	if len(args) != 2 {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	delegatorAddress, ok := args[0].(common.Address)
	if !ok || delegatorAddress == (common.Address{}) {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	withdrawerAddress, _ := args[1].(string)

	// If the withdrawer address is a hex address, convert it to a bech32 address.
	if common.IsHexAddress(withdrawerAddress) {
		var err error
		withdrawerAddress, err = sdk.Bech32ifyAddressBytes(config.Bech32Prefix, common.HexToAddress(withdrawerAddress).Bytes())
		if err != nil {
			return nil, common.Address{}, err
		}
	}

	msg := &distributiontypes.MsgSetWithdrawAddress{
		DelegatorAddress: sdk.AccAddress(delegatorAddress.Bytes()).String(),
		WithdrawAddress:  withdrawerAddress,
	}

	return msg, delegatorAddress, nil
}

// NewMsgWithdrawDelegatorReward creates a new MsgWithdrawDelegatorReward instance.
func NewMsgWithdrawDelegatorReward(args []interface{}) (*distributiontypes.MsgWithdrawDelegatorReward, common.Address, error) {
	if len(args) != 2 {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	delegatorAddress, ok := args[0].(common.Address)
	if !ok || delegatorAddress == (common.Address{}) {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	validatorAddress, _ := args[1].(string)

	msg := &distributiontypes.MsgWithdrawDelegatorReward{
		DelegatorAddress: sdk.AccAddress(delegatorAddress.Bytes()).String(),
		ValidatorAddress: validatorAddress,
	}

	return msg, delegatorAddress, nil
}

// NewMsgWithdrawValidatorCommission creates a new MsgWithdrawValidatorCommission message.
func NewMsgWithdrawValidatorCommission(args []interface{}) (*distributiontypes.MsgWithdrawValidatorCommission, common.Address, error) {
	if len(args) != 1 {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	validatorAddress, _ := args[0].(string)

	msg := &distributiontypes.MsgWithdrawValidatorCommission{
		ValidatorAddress: validatorAddress,
	}

	validatorHexAddr, err := cmn.HexAddressFromBech32String(msg.ValidatorAddress)
	if err != nil {
		return nil, common.Address{}, err
	}

	return msg, validatorHexAddr, nil
}

// NewMsgFundCommunityPool creates a new NewMsgFundCommunityPool message.
func NewMsgFundCommunityPool(args []interface{}) (*distributiontypes.MsgFundCommunityPool, common.Address, error) {
	if len(args) != 2 {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	depositorAddress, ok := args[0].(common.Address)
	if !ok || depositorAddress == (common.Address{}) {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidHexAddress, args[0])
	}

	amount, ok := args[1].(*big.Int)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidAmount, args[1])
	}

	msg := &distributiontypes.MsgFundCommunityPool{
		Depositor: sdk.AccAddress(depositorAddress.Bytes()).String(),
		Amount:    sdk.Coins{sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewIntFromBigInt(amount)}},
	}

	return msg, depositorAddress, nil
}

// NewValidatorDistributionInfoRequest creates a new QueryValidatorDistributionInfoRequest  instance and does sanity
// checks on the provided arguments.
func NewValidatorDistributionInfoRequest(args []interface{}) (*distributiontypes.QueryValidatorDistributionInfoRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	validatorAddress, _ := args[0].(string)

	return &distributiontypes.QueryValidatorDistributionInfoRequest{
		ValidatorAddress: validatorAddress,
	}, nil
}

// NewValidatorOutstandingRewardsRequest creates a new QueryValidatorOutstandingRewardsRequest  instance and does sanity
// checks on the provided arguments.
func NewValidatorOutstandingRewardsRequest(args []interface{}) (*distributiontypes.QueryValidatorOutstandingRewardsRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	validatorAddress, _ := args[0].(string)

	return &distributiontypes.QueryValidatorOutstandingRewardsRequest{
		ValidatorAddress: validatorAddress,
	}, nil
}

// NewValidatorCommissionRequest creates a new QueryValidatorCommissionRequest  instance and does sanity
// checks on the provided arguments.
func NewValidatorCommissionRequest(args []interface{}) (*distributiontypes.QueryValidatorCommissionRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	validatorAddress, _ := args[0].(string)

	return &distributiontypes.QueryValidatorCommissionRequest{
		ValidatorAddress: validatorAddress,
	}, nil
}

// NewValidatorSlashesRequest creates a new QueryValidatorSlashesRequest  instance and does sanity
// checks on the provided arguments.
func NewValidatorSlashesRequest(method *abi.Method, args []interface{}) (*distributiontypes.QueryValidatorSlashesRequest, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 4, len(args))
	}

	if _, ok := args[1].(uint64); !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "startingHeight", uint64(0), args[1])
	}
	if _, ok := args[2].(uint64); !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "endingHeight", uint64(0), args[2])
	}

	var input ValidatorSlashesInput
	if err := method.Inputs.Copy(&input, args); err != nil {
		return nil, fmt.Errorf("error while unpacking args to ValidatorSlashesInput struct: %s", err)
	}

	return &distributiontypes.QueryValidatorSlashesRequest{
		ValidatorAddress: input.ValidatorAddress,
		StartingHeight:   input.StartingHeight,
		EndingHeight:     input.EndingHeight,
		Pagination:       &input.PageRequest,
	}, nil
}

// NewDelegationRewardsRequest creates a new QueryDelegationRewardsRequest  instance and does sanity
// checks on the provided arguments.
func NewDelegationRewardsRequest(args []interface{}) (*distributiontypes.QueryDelegationRewardsRequest, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	delegatorAddress, ok := args[0].(common.Address)
	if !ok || delegatorAddress == (common.Address{}) {
		return nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	validatorAddress, _ := args[1].(string)

	return &distributiontypes.QueryDelegationRewardsRequest{
		DelegatorAddress: sdk.AccAddress(delegatorAddress.Bytes()).String(),
		ValidatorAddress: validatorAddress,
	}, nil
}

// NewDelegationTotalRewardsRequest creates a new QueryDelegationTotalRewardsRequest  instance and does sanity
// checks on the provided arguments.
func NewDelegationTotalRewardsRequest(args []interface{}) (*distributiontypes.QueryDelegationTotalRewardsRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	delegatorAddress, ok := args[0].(common.Address)
	if !ok || delegatorAddress == (common.Address{}) {
		return nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	return &distributiontypes.QueryDelegationTotalRewardsRequest{
		DelegatorAddress: sdk.AccAddress(delegatorAddress.Bytes()).String(),
	}, nil
}

// NewDelegatorValidatorsRequest creates a new QueryDelegatorValidatorsRequest  instance and does sanity
// checks on the provided arguments.
func NewDelegatorValidatorsRequest(args []interface{}) (*distributiontypes.QueryDelegatorValidatorsRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	delegatorAddress, ok := args[0].(common.Address)
	if !ok || delegatorAddress == (common.Address{}) {
		return nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	return &distributiontypes.QueryDelegatorValidatorsRequest{
		DelegatorAddress: sdk.AccAddress(delegatorAddress.Bytes()).String(),
	}, nil
}

// NewDelegatorWithdrawAddressRequest creates a new QueryDelegatorWithdrawAddressRequest  instance and does sanity
// checks on the provided arguments.
func NewDelegatorWithdrawAddressRequest(args []interface{}) (*distributiontypes.QueryDelegatorWithdrawAddressRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	delegatorAddress, ok := args[0].(common.Address)
	if !ok || delegatorAddress == (common.Address{}) {
		return nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	return &distributiontypes.QueryDelegatorWithdrawAddressRequest{
		DelegatorAddress: sdk.AccAddress(delegatorAddress.Bytes()).String(),
	}, nil
}

// ValidatorDistributionInfo is a struct to represent the key information from
// a ValidatorDistributionInfoResponse.
type ValidatorDistributionInfo struct {
	OperatorAddress string        `abi:"operatorAddress"`
	SelfBondRewards []cmn.DecCoin `abi:"selfBondRewards"`
	Commission      []cmn.DecCoin `abi:"commission"`
}

// ValidatorDistributionInfoOutput is a wrapper for ValidatorDistributionInfo to return in the response.
type ValidatorDistributionInfoOutput struct {
	DistributionInfo ValidatorDistributionInfo `abi:"distributionInfo"`
}

// FromResponse converts a response to a ValidatorDistributionInfo.
func (o *ValidatorDistributionInfoOutput) FromResponse(res *distributiontypes.QueryValidatorDistributionInfoResponse) ValidatorDistributionInfoOutput {
	return ValidatorDistributionInfoOutput{
		DistributionInfo: ValidatorDistributionInfo{
			OperatorAddress: res.OperatorAddress,
			SelfBondRewards: cmn.NewDecCoinsResponse(res.SelfBondRewards),
			Commission:      cmn.NewDecCoinsResponse(res.Commission),
		},
	}
}

// ValidatorSlashEvent is a struct to represent the key information from
// a ValidatorSlashEvent response.
type ValidatorSlashEvent struct {
	ValidatorPeriod uint64  `abi:"validatorPeriod"`
	Fraction        cmn.Dec `abi:"fraction"`
}

// ValidatorSlashesInput is a struct to represent the key information
// to perform a ValidatorSlashes query.
type ValidatorSlashesInput struct {
	ValidatorAddress string
	StartingHeight   uint64
	EndingHeight     uint64
	PageRequest      query.PageRequest
}

// ValidatorSlashesOutput is a struct to represent the key information from
// a ValidatorSlashes response.
type ValidatorSlashesOutput struct {
	Slashes      []ValidatorSlashEvent
	PageResponse query.PageResponse
}

// FromResponse populates the ValidatorSlashesOutput from a QueryValidatorSlashesResponse.
func (vs *ValidatorSlashesOutput) FromResponse(res *distributiontypes.QueryValidatorSlashesResponse) *ValidatorSlashesOutput {
	vs.Slashes = make([]ValidatorSlashEvent, len(res.Slashes))
	for i, s := range res.Slashes {
		vs.Slashes[i] = ValidatorSlashEvent{
			ValidatorPeriod: s.ValidatorPeriod,
			Fraction: cmn.Dec{
				Value:     s.Fraction.BigInt(),
				Precision: math.LegacyPrecision,
			},
		}
	}

	if res.Pagination != nil {
		vs.PageResponse.Total = res.Pagination.Total
		vs.PageResponse.NextKey = res.Pagination.NextKey
	}

	return vs
}

// Pack packs a given slice of abi arguments into a byte array.
func (vs *ValidatorSlashesOutput) Pack(args abi.Arguments) ([]byte, error) {
	return args.Pack(vs.Slashes, vs.PageResponse)
}

// DelegationDelegatorReward is a struct to represent the key information from
// a query for the rewards of a delegation to a given validator.
type DelegationDelegatorReward struct {
	ValidatorAddress string
	Reward           []cmn.DecCoin
}

// DelegationTotalRewardsOutput is a struct to represent the key information from
// a DelegationTotalRewards response.
type DelegationTotalRewardsOutput struct {
	Rewards []DelegationDelegatorReward
	Total   []cmn.DecCoin
}

// FromResponse populates the DelegationTotalRewardsOutput from a QueryDelegationTotalRewardsResponse.
func (dtr *DelegationTotalRewardsOutput) FromResponse(res *distributiontypes.QueryDelegationTotalRewardsResponse) *DelegationTotalRewardsOutput {
	dtr.Rewards = make([]DelegationDelegatorReward, len(res.Rewards))
	for i, r := range res.Rewards {
		dtr.Rewards[i] = DelegationDelegatorReward{
			ValidatorAddress: r.ValidatorAddress,
			Reward:           cmn.NewDecCoinsResponse(r.Reward),
		}
	}
	dtr.Total = cmn.NewDecCoinsResponse(res.Total)
	return dtr
}

// Pack packs a given slice of abi arguments into a byte array.
func (dtr *DelegationTotalRewardsOutput) Pack(args abi.Arguments) ([]byte, error) {
	return args.Pack(dtr.Rewards, dtr.Total)
}
