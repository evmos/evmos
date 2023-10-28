// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package distribution

import (
	"bytes"
	"embed"
	"fmt"

	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
)

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

// Precompile defines the precompiled contract for distribution.
type Precompile struct {
	cmn.Precompile
	distributionKeeper distributionkeeper.Keeper
	stakingKeeper      stakingkeeper.Keeper
}

// NewPrecompile creates a new distribution Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	distributionKeeper distributionkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
	authzKeeper authzkeeper.Keeper,
) (*Precompile, error) {
	abiBz, err := f.ReadFile("abi.json")
	if err != nil {
		return nil, fmt.Errorf("error loading the distribution ABI %s", err)
	}

	newAbi, err := abi.JSON(bytes.NewReader(abiBz))
	if err != nil {
		return nil, fmt.Errorf(cmn.ErrInvalidABI, err)
	}

	return &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  newAbi,
			AuthzKeeper:          authzKeeper,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ApprovalExpiration:   cmn.DefaultExpirationDuration, // should be configurable in the future.
		},
		stakingKeeper:      stakingKeeper,
		distributionKeeper: distributionKeeper,
	}, nil
}

// Address defines the address of the distribution compile contract.
// address: 0x0000000000000000000000000000000000000801
func (p Precompile) Address() common.Address {
	return common.HexToAddress("0x0000000000000000000000000000000000000801")
}

// RequiredGas calculates the precompiled contract's base gas rate.
func (p Precompile) RequiredGas(input []byte) uint64 {
	methodID := input[:4]

	method, err := p.MethodById(methodID)
	if err != nil {
		// This should never happen since this method is going to fail during Run
		return 0
	}

	return p.Precompile.RequiredGas(input, p.IsTransaction(method.Name))
}

// Run executes the precompiled contract distribution methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	ctx, stateDB, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors that may occur during the execution of a precompile tx or query.
	// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	switch method.Name {
	// Custom transactions
	case ClaimRewardsMethod:
		bz, err = p.ClaimRewards(ctx, evm.Origin, contract, stateDB, method, args)
	// Distribution transactions
	case SetWithdrawAddressMethod:
		bz, err = p.SetWithdrawAddress(ctx, evm.Origin, contract, stateDB, method, args)
	case WithdrawDelegatorRewardsMethod:
		bz, err = p.WithdrawDelegatorRewards(ctx, evm.Origin, contract, stateDB, method, args)
	case WithdrawValidatorCommissionMethod:
		bz, err = p.WithdrawValidatorCommission(ctx, evm.Origin, contract, stateDB, method, args)
	// Distribution queries
	case ValidatorDistributionInfoMethod:
		bz, err = p.ValidatorDistributionInfo(ctx, contract, method, args)
	case ValidatorOutstandingRewardsMethod:
		bz, err = p.ValidatorOutstandingRewards(ctx, contract, method, args)
	case ValidatorCommissionMethod:
		bz, err = p.ValidatorCommission(ctx, contract, method, args)
	case ValidatorSlashesMethod:
		bz, err = p.ValidatorSlashes(ctx, contract, method, args)
	case DelegationRewardsMethod:
		bz, err = p.DelegationRewards(ctx, contract, method, args)
	case DelegationTotalRewardsMethod:
		bz, err = p.DelegationTotalRewards(ctx, contract, method, args)
	case DelegatorValidatorsMethod:
		bz, err = p.DelegatorValidators(ctx, contract, method, args)
	case DelegatorWithdrawAddressMethod:
		bz, err = p.DelegatorWithdrawAddress(ctx, contract, method, args)
	}

	if err != nil {
		return nil, err
	}

	cost := ctx.GasMeter().GasConsumed() - initialGas

	if !contract.UseGas(cost) {
		return nil, vm.ErrOutOfGas
	}

	return bz, nil
}

// IsTransaction checks if the given methodID corresponds to a transaction or query.
//
// Available distribution transactions are:
//   - ClaimRewards
//   - SetWithdrawAddress
//   - WithdrawDelegatorRewards
//   - WithdrawValidatorCommission
func (Precompile) IsTransaction(methodID string) bool {
	switch methodID {
	case ClaimRewardsMethod,
		SetWithdrawAddressMethod,
		WithdrawDelegatorRewardsMethod,
		WithdrawValidatorCommissionMethod:
		return true
	default:
		return false
	}
}
