// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	"golang.org/x/exp/maps"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channelkeeper "github.com/cosmos/ibc-go/v7/modules/core/04-channel/keeper"
	distprecompile "github.com/evmos/evmos/v15/precompiles/distribution"
	ics20precompile "github.com/evmos/evmos/v15/precompiles/ics20"
	strideoutpost "github.com/evmos/evmos/v15/precompiles/outposts/stride"
	"github.com/evmos/evmos/v15/precompiles/p256"
	stakingprecompile "github.com/evmos/evmos/v15/precompiles/staking"
	vestingprecompile "github.com/evmos/evmos/v15/precompiles/vesting"
	erc20Keeper "github.com/evmos/evmos/v15/x/erc20/keeper"
	transferkeeper "github.com/evmos/evmos/v15/x/ibc/transfer/keeper"
	vestingkeeper "github.com/evmos/evmos/v15/x/vesting/keeper"
)

// AvailablePrecompiles returns the list of all available precompiled contracts.
// NOTE: this should only be used during initialization of the Keeper.
func AvailablePrecompiles(
	stakingKeeper stakingkeeper.Keeper,
	distributionKeeper distributionkeeper.Keeper,
	erc20Keeper erc20Keeper.Keeper,
	vestingKeeper vestingkeeper.Keeper,
	authzKeeper authzkeeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	channelKeeper channelkeeper.Keeper,
) map[common.Address]vm.PrecompiledContract {
	// Clone the mapping from the latest EVM fork.
	precompiles := maps.Clone(vm.PrecompiledContractsBerlin)

	// secp256r1 precompile as per EIP-7212
	p256Precompile := &p256.Precompile{}

	stakingPrecompile, err := stakingprecompile.NewPrecompile(stakingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to load staking precompile: %w", err))
	}

	distributionPrecompile, err := distprecompile.NewPrecompile(distributionKeeper, stakingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to load distribution precompile: %w", err))
	}

	ibcTransferPrecompile, err := ics20precompile.NewPrecompile(transferKeeper, channelKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to load ICS20 precompile: %w", err))
	}

	vestingPrecompile, err := vestingprecompile.NewPrecompile(vestingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to load vesting precompile: %w", err))
	}

	strideOutpost, err := strideoutpost.NewPrecompile(transfertypes.PortID, "channel-25", transferKeeper, erc20Keeper, authzKeeper, stakingKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to load stride outpost: %w", err))
	}

	precompiles[p256Precompile.Address()] = p256Precompile
	precompiles[stakingPrecompile.Address()] = stakingPrecompile
	precompiles[distributionPrecompile.Address()] = distributionPrecompile
	precompiles[vestingPrecompile.Address()] = vestingPrecompile
	precompiles[ibcTransferPrecompile.Address()] = ibcTransferPrecompile
	precompiles[strideOutpost.Address()] = strideOutpost
	return precompiles
}

// WithPrecompiles sets the available precompiled contracts.
func (k *Keeper) WithPrecompiles(precompiles map[common.Address]vm.PrecompiledContract) *Keeper {
	if k.precompiles != nil {
		panic("available precompiles map already set")
	}

	if len(precompiles) == 0 {
		panic("empty precompiled contract map")
	}

	k.precompiles = precompiles
	return k
}

// Precompiles returns the subset of the available precompiled contracts that
// are active given the current parameters.
func (k Keeper) Precompiles(
	activePrecompiles ...common.Address,
) map[common.Address]vm.PrecompiledContract {
	activePrecompileMap := make(map[common.Address]vm.PrecompiledContract)

	for _, address := range activePrecompiles {
		precompile, ok := k.precompiles[address]
		if !ok {
			panic(fmt.Sprintf("precompiled contract not initialized: %s", address))
		}

		activePrecompileMap[address] = precompile
	}

	return activePrecompileMap
}
