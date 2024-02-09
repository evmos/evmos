// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	"github.com/evmos/evmos/v16/utils"

	"github.com/evmos/evmos/v16/precompiles/bech32"
	"github.com/evmos/evmos/v16/precompiles/werc20"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"golang.org/x/exp/maps"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v7/modules/core/04-channel/keeper"
	bankprecompile "github.com/evmos/evmos/v16/precompiles/bank"
	distprecompile "github.com/evmos/evmos/v16/precompiles/distribution"
	erc20precompile "github.com/evmos/evmos/v16/precompiles/erc20"
	ics20precompile "github.com/evmos/evmos/v16/precompiles/ics20"
	osmosisoutpost "github.com/evmos/evmos/v16/precompiles/outposts/osmosis"
	strideoutpost "github.com/evmos/evmos/v16/precompiles/outposts/stride"
	"github.com/evmos/evmos/v16/precompiles/p256"
	stakingprecompile "github.com/evmos/evmos/v16/precompiles/staking"
	vestingprecompile "github.com/evmos/evmos/v16/precompiles/vesting"
	erc20Keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	transferkeeper "github.com/evmos/evmos/v16/x/ibc/transfer/keeper"
	vestingkeeper "github.com/evmos/evmos/v16/x/vesting/keeper"
)

// AvailablePrecompiles returns the list of all available precompiled contracts.
// NOTE: this should only be used during initialization of the Keeper.
func AvailablePrecompiles(
	chainID string,
	stakingKeeper stakingkeeper.Keeper,
	distributionKeeper distributionkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
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

	bech32Precompile, err := bech32.NewPrecompile(6000)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate bech32 precompile: %w", err))
	}

	stakingPrecompile, err := stakingprecompile.NewPrecompile(stakingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate staking precompile: %w", err))
	}

	distributionPrecompile, err := distprecompile.NewPrecompile(distributionKeeper, stakingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate distribution precompile: %w", err))
	}

	ibcTransferPrecompile, err := ics20precompile.NewPrecompile(transferKeeper, channelKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate ICS20 precompile: %w", err))
	}

	vestingPrecompile, err := vestingprecompile.NewPrecompile(vestingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate vesting precompile: %w", err))
	}

	bankPrecompile, err := bankprecompile.NewPrecompile(bankKeeper, erc20Keeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate bank precompile: %w", err))
	}

	var WEVMOSAddress common.Address
	if utils.IsMainnet(chainID) {
		WEVMOSAddress = common.HexToAddress(erc20precompile.WEVMOSContractMainnet)
	} else {
		WEVMOSAddress = common.HexToAddress(erc20precompile.WEVMOSContractTestnet)
	}
	tokenPair := erc20types.NewTokenPair(WEVMOSAddress, "aevmos", erc20types.OWNER_MODULE)
	wevmosprecompile, err := werc20.NewPrecompile(
		tokenPair,
		bankKeeper,
		authzKeeper,
		transferKeeper,
	)

	if err != nil {
		panic(fmt.Errorf("failed to wevmos bank precompile: %w", err))
	}

	strideOutpost, err := strideoutpost.NewPrecompile(
		WEVMOSAddress,
		transferKeeper,
		erc20Keeper,
		authzKeeper,
		stakingKeeper,
	)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate stride outpost: %w", err))
	}

	osmosisOutpost, err := osmosisoutpost.NewPrecompile(
		WEVMOSAddress,
		authzKeeper,
		bankKeeper,
		transferKeeper,
		stakingKeeper,
		erc20Keeper,
		channelKeeper,
	)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate osmosis outpost: %w", err))
	}

	// Stateless precompiles
	precompiles[bech32Precompile.Address()] = bech32Precompile
	precompiles[p256Precompile.Address()] = p256Precompile

	// Stateful precompiles
	precompiles[stakingPrecompile.Address()] = stakingPrecompile
	precompiles[distributionPrecompile.Address()] = distributionPrecompile
	precompiles[vestingPrecompile.Address()] = vestingPrecompile
	precompiles[ibcTransferPrecompile.Address()] = ibcTransferPrecompile
	precompiles[bankPrecompile.Address()] = bankPrecompile

	// Outposts
	precompiles[strideOutpost.Address()] = strideOutpost
	precompiles[osmosisOutpost.Address()] = osmosisOutpost

	// Wevmos
	precompiles[WEVMOSAddress] = wevmosprecompile

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

// GetStaticPrecompilesInstances returns the subset of the available precompiled contracts that
// are active given the current parameters.
func (k Keeper) GetStaticPrecompilesInstances(
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

// IsAvailablePrecompile returns true if the given precompile address is contained in the
// EVM keeper's available precompiles map.
func (k Keeper) IsAvailablePrecompile(address common.Address) bool {
	_, ok := k.precompiles[address]
	return ok
}

func (k Keeper) GetIfAvailablePrecompile(address common.Address) vm.PrecompiledContract {
	precompile, ok := k.precompiles[address]
	if !ok {
		return nil
	}
	return precompile
}
