// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package distribution

import (
	"fmt"

	"github.com/evmos/evmos/v19/utils"

	cmn "github.com/evmos/evmos/v19/precompiles/common"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
)

const (
	// SetWithdrawAddressMethod defines the ABI method name for the distribution
	// SetWithdrawAddress transaction.
	SetWithdrawAddressMethod = "setWithdrawAddress"
	// WithdrawDelegatorRewardsMethod defines the ABI method name for the distribution
	// WithdrawDelegatorRewards transaction.
	WithdrawDelegatorRewardsMethod = "withdrawDelegatorRewards"
	// WithdrawValidatorCommissionMethod defines the ABI method name for the distribution
	// WithdrawValidatorCommission transaction.
	WithdrawValidatorCommissionMethod = "withdrawValidatorCommission"
	// FundCommunityPoolMethod defines the ABI method name for the fundCommunityPool transaction
	FundCommunityPoolMethod = "fundCommunityPool"
	// ClaimRewardsMethod defines the ABI method name for the custom ClaimRewards transaction
	ClaimRewardsMethod = "claimRewards"
)

// ClaimRewards claims the rewards accumulated by a delegator from multiple or all validators.
func (p Precompile) ClaimRewards(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	delegatorAddr, maxRetrieve, err := parseClaimRewardsArgs(args)
	if err != nil {
		return nil, err
	}

	maxVals := p.stakingKeeper.MaxValidators(ctx)
	if maxRetrieve > maxVals {
		return nil, fmt.Errorf("maxRetrieve (%d) parameter exceeds the maximum number of validators (%d)", maxRetrieve, maxVals)
	}

	// If the contract is the delegator, we don't need an origin check
	// Otherwise check if the origin matches the delegator address
	isContractDelegator := contract.CallerAddress == delegatorAddr && origin != delegatorAddr
	if !isContractDelegator && origin != delegatorAddr {
		return nil, fmt.Errorf(cmn.ErrDelegatorDifferentOrigin, origin.String(), delegatorAddr.String())
	}

	validators := p.stakingKeeper.GetDelegatorValidators(ctx, delegatorAddr.Bytes(), maxRetrieve)
	totalCoins := sdk.Coins{}
	for _, validator := range validators {
		// Convert the validator operator address into an ValAddress
		valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
		if err != nil {
			return nil, err
		}

		// Withdraw the rewards for each validator address
		coins, err := p.distributionKeeper.WithdrawDelegationRewards(ctx, delegatorAddr.Bytes(), valAddr)
		if err != nil {
			return nil, err
		}

		totalCoins = totalCoins.Add(coins...)
	}

	// NOTE: This ensures that the changes in the bank keeper are correctly mirrored to the EVM stateDB.
	// This prevents the stateDB from overwriting the changed balance in the bank keeper when committing the EVM state.
	// this happens when the precompile is called from a smart contract
	if contract.CallerAddress != origin {
		// rewards go to the withdrawer address
		withdrawerHexAddr := p.getWithdrawerHexAddr(ctx, delegatorAddr)
		p.SetBalanceChangeEntries(cmn.NewBalanceChangeEntry(withdrawerHexAddr, totalCoins.AmountOf(utils.BaseDenom).BigInt(), cmn.Add))
	}

	if err := p.EmitClaimRewardsEvent(ctx, stateDB, delegatorAddr, totalCoins); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// SetWithdrawAddress sets the withdrawal address for a delegator (or validator self-delegation).
func (p Precompile) SetWithdrawAddress(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, delegatorHexAddr, err := NewMsgSetWithdrawAddress(args)
	if err != nil {
		return nil, err
	}

	// If the contract is the delegator, we don't need an origin check
	// Otherwise check if the origin matches the delegator address
	isContractDelegator := contract.CallerAddress == delegatorHexAddr && origin != delegatorHexAddr
	if !isContractDelegator && origin != delegatorHexAddr {
		return nil, fmt.Errorf(cmn.ErrDelegatorDifferentOrigin, origin.String(), delegatorHexAddr.String())
	}

	msgSrv := distributionkeeper.NewMsgServerImpl(p.distributionKeeper)
	if _, err = msgSrv.SetWithdrawAddress(sdk.WrapSDKContext(ctx), msg); err != nil {
		return nil, err
	}

	if err = p.EmitSetWithdrawAddressEvent(ctx, stateDB, delegatorHexAddr, msg.WithdrawAddress); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// WithdrawDelegatorRewards withdraws the rewards of a delegator from a single validator.
func (p *Precompile) WithdrawDelegatorRewards(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, delegatorHexAddr, err := NewMsgWithdrawDelegatorReward(args)
	if err != nil {
		return nil, err
	}

	// If the contract is the delegator, we don't need an origin check
	// Otherwise check if the origin matches the delegator address
	isContractDelegator := contract.CallerAddress == delegatorHexAddr && origin != delegatorHexAddr
	if !isContractDelegator && origin != delegatorHexAddr {
		return nil, fmt.Errorf(cmn.ErrDelegatorDifferentOrigin, origin.String(), delegatorHexAddr.String())
	}

	msgSrv := distributionkeeper.NewMsgServerImpl(p.distributionKeeper)
	res, err := msgSrv.WithdrawDelegatorReward(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	// NOTE: This ensures that the changes in the bank keeper are correctly mirrored to the EVM stateDB
	// when calling the precompile from a smart contract
	// This prevents the stateDB from overwriting the changed balance in the bank keeper when committing the EVM state.
	if contract.CallerAddress != origin {
		// rewards go to the withdrawer address
		withdrawerHexAddr := p.getWithdrawerHexAddr(ctx, delegatorHexAddr)
		p.SetBalanceChangeEntries(cmn.NewBalanceChangeEntry(withdrawerHexAddr, res.Amount[0].Amount.BigInt(), cmn.Add))
	}

	if err = p.EmitWithdrawDelegatorRewardsEvent(ctx, stateDB, delegatorHexAddr, msg.ValidatorAddress, res.Amount); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(cmn.NewCoinsResponse(res.Amount))
}

// WithdrawValidatorCommission withdraws the rewards of a validator.
func (p *Precompile) WithdrawValidatorCommission(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, validatorHexAddr, err := NewMsgWithdrawValidatorCommission(args)
	if err != nil {
		return nil, err
	}

	// If the contract is the validator, we don't need an origin check
	// Otherwise check if the origin matches the validator address
	isContractValidator := contract.CallerAddress == validatorHexAddr && origin != validatorHexAddr
	if !isContractValidator && origin != validatorHexAddr {
		return nil, fmt.Errorf(cmn.ErrDelegatorDifferentOrigin, origin.String(), validatorHexAddr.String())
	}

	msgSrv := distributionkeeper.NewMsgServerImpl(p.distributionKeeper)
	res, err := msgSrv.WithdrawValidatorCommission(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	// NOTE: This ensures that the changes in the bank keeper are correctly mirrored to the EVM stateDB
	// when calling the precompile from a smart contract
	// This prevents the stateDB from overwriting the changed balance in the bank keeper when committing the EVM state.
	if contract.CallerAddress != origin {
		// commissions go to the withdrawer address
		withdrawerHexAddr := p.getWithdrawerHexAddr(ctx, validatorHexAddr)
		p.SetBalanceChangeEntries(cmn.NewBalanceChangeEntry(withdrawerHexAddr, res.Amount[0].Amount.BigInt(), cmn.Add))
	}

	if err = p.EmitWithdrawValidatorCommissionEvent(ctx, stateDB, msg.ValidatorAddress, res.Amount); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(cmn.NewCoinsResponse(res.Amount))
}

// FundCommunityPool directly fund the community pool
func (p *Precompile) FundCommunityPool(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, depositorHexAddr, err := NewMsgFundCommunityPool(args)
	if err != nil {
		return nil, err
	}

	// If the contract is the depositor, we don't need an origin check
	// Otherwise check if the origin matches the depositor address
	isContractDepositor := contract.CallerAddress == depositorHexAddr && origin != depositorHexAddr
	if !isContractDepositor && origin != depositorHexAddr {
		return nil, fmt.Errorf(cmn.ErrSpenderDifferentOrigin, origin.String(), depositorHexAddr.String())
	}

	msgSrv := distributionkeeper.NewMsgServerImpl(p.distributionKeeper)
	_, err = msgSrv.FundCommunityPool(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	// NOTE: This ensures that the changes in the bank keeper are correctly mirrored to the EVM stateDB
	// when calling the precompile from a smart contract
	// This prevents the stateDB from overwriting the changed balance in the bank keeper when committing the EVM state.
	if contract.CallerAddress != origin {
		p.SetBalanceChangeEntries(cmn.NewBalanceChangeEntry(depositorHexAddr, msg.Amount.AmountOf(utils.BaseDenom).BigInt(), cmn.Sub))
	}

	if err = p.EmitFundCommunityPoolEvent(ctx, stateDB, depositorHexAddr, msg.Amount); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// getWithdrawerHexAddr is a helper function to get the hex address
// of the withdrawer for the specified account address
func (p Precompile) getWithdrawerHexAddr(ctx sdk.Context, delegatorAddr common.Address) common.Address {
	withdrawerAccAddr := p.distributionKeeper.GetDelegatorWithdrawAddr(ctx, delegatorAddr.Bytes())
	return common.BytesToAddress(withdrawerAccAddr)
}
